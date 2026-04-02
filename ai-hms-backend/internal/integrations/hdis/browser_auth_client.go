package hdis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	defaultBrowserTimeout = 120 * time.Second
	loginFormSelector     = `input[name='Input.UserName'], input[name='Mobile'], input[name='Username'], input[name='UserName'], input[type='password']`
)

// BrowserAuthClient 使用浏览器登录流程获取 token。
type BrowserAuthClient struct {
	headless bool
	timeout  time.Duration
}

// BrowserTokenRefreshRequest 浏览器登录刷新请求。
type BrowserTokenRefreshRequest struct {
	AuthURL       string
	ClientID      string
	Username      string
	Password      string
	GraphqlURL    string
	WebcmdURL     string
	TargetOrganID string // 期望的科室 ID，仅用于告警校验
}

type browserContexts struct {
	ctx     context.Context
	taskCtx context.Context
	cancel  func()
}

func NewBrowserAuthClient(headless bool, timeout time.Duration) *BrowserAuthClient {
	if timeout <= 0 {
		timeout = defaultBrowserTimeout
	}
	return &BrowserAuthClient{
		headless: headless,
		timeout:  timeout,
	}
}

func (c *BrowserAuthClient) RefreshToken(ctx context.Context, req BrowserTokenRefreshRequest) (*TokenRefreshResponse, error) {
	if strings.TrimSpace(req.AuthURL) == "" {
		return nil, errors.New("hdis auth endpoint is empty")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return nil, errors.New("hdis client id is required")
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("hdis service credentials are required")
	}

	redirectURI, err := deriveRedirectURI(req.GraphqlURL, req.WebcmdURL)
	if err != nil {
		return nil, err
	}
	authorizeURL, err := buildAuthorizeURL(req.AuthURL, req.ClientID, redirectURI, randomToken(8), randomToken(8))
	if err != nil {
		return nil, err
	}

	bc := c.newBrowserContexts(ctx)
	defer bc.cancel()

	slog.Info("browser auth: navigating to authorize URL", "url", authorizeURL)

	navigateAndWait := func(runCtx context.Context) error {
		return chromedp.Run(runCtx,
			chromedp.Navigate(authorizeURL),
			chromedp.WaitVisible(loginFormSelector, chromedp.ByQuery),
		)
	}

	err = navigateAndWait(bc.ctx)
	if err != nil && isRetryableNavigationError(err) {
		slog.Warn("browser auth: retrying authorize navigation after transient error", "error", err)
		select {
		case <-bc.ctx.Done():
		case <-time.After(2 * time.Second):
		}
		if bc.ctx.Err() == nil {
			err = navigateAndWait(bc.ctx)
		}
	}
	if err != nil {
		currentURL, pageTitle, bodyText := c.captureDiagnostics(bc.taskCtx)
		slog.Error("browser auth: login form not found",
			"authorizeURL", authorizeURL,
			"currentURL", currentURL,
			"pageTitle", pageTitle,
			"bodySnippet", bodyText,
			"error", err,
		)
		return nil, fmt.Errorf("login form not found (landed on %s, title=%s): %w", currentURL, pageTitle, err)
	}

	if err := c.fillAndSubmitLogin(bc.ctx, req.Username, req.Password); err != nil {
		currentURL, _, _ := c.captureDiagnostics(bc.taskCtx)
		slog.Error("browser auth: login submit failed", "currentURL", currentURL, "error", err)
		return nil, err
	}

	token, err := c.waitForAccessToken(bc.ctx)
	if err != nil {
		return nil, err
	}

	actualOrganID := parseJWTStringClaim(token, "organ_id")
	targetOrganID := strings.TrimSpace(req.TargetOrganID)
	if targetOrganID != "" && actualOrganID != "" && actualOrganID != targetOrganID {
		slog.Warn("browser auth: token organ_id mismatch",
			"expectedOrganID", targetOrganID,
			"actualOrganID", actualOrganID,
			"hint", "auth server selected default organ; verify account default organ or HDIS_TARGET_ORGAN_ID",
		)
	}

	expiresAt := time.Now().Add(12 * time.Hour)
	if jwtExp := parseJWTExp(token); jwtExp != nil {
		expiresAt = *jwtExp
	}

	return &TokenRefreshResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		OrganID:     actualOrganID,
	}, nil
}

func (c *BrowserAuthClient) newBrowserContexts(parent context.Context) browserContexts {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", c.headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1280, 800),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, opts...)
	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	timeoutCtx, timeoutCancel := context.WithTimeout(taskCtx, c.timeout)

	return browserContexts{
		ctx:     timeoutCtx,
		taskCtx: taskCtx,
		cancel: func() {
			timeoutCancel()
			taskCancel()
			allocCancel()
		},
	}
}

func (c *BrowserAuthClient) captureDiagnostics(taskCtx context.Context) (string, string, string) {
	if taskCtx == nil {
		return "", "", ""
	}

	diagCtx, cancel := context.WithTimeout(taskCtx, 3*time.Second)
	defer cancel()

	var currentURL, pageTitle, bodyText string
	_ = chromedp.Run(diagCtx,
		chromedp.Location(&currentURL),
		chromedp.Title(&pageTitle),
		chromedp.Evaluate(`document.body ? document.body.innerText.substring(0, 500) : ""`, &bodyText),
	)
	return currentURL, pageTitle, bodyText
}

func isRetryableNavigationError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "err_timed_out") || strings.Contains(lower, "err_incomplete_chunked_encoding")
}

func (c *BrowserAuthClient) fillAndSubmitLogin(ctx context.Context, username, password string) error {
	script := fmt.Sprintf(`(() => {
		const findFirst = (selectors) => {
			for (const s of selectors) {
				const el = document.querySelector(s);
				if (el) return el;
			}
			return null;
		};

		const user = findFirst([
			"input[name='Input.UserName']",
			"input[name='Mobile']",
			"input[name='Username']",
			"input[name='UserName']",
			"input[name='Account']",
			"input[type='text']"
		]);
		const pass = findFirst([
			"input[name='Input.Password']",
			"input[name='Password']",
			"input[type='password']"
		]);
		if (!user || !pass) return "LOGIN_FIELDS_NOT_FOUND";

		const triggerInput = (el, value) => {
			el.focus();
			el.value = value;
			el.dispatchEvent(new Event("input", { bubbles: true }));
			el.dispatchEvent(new Event("change", { bubbles: true }));
		};

		triggerInput(user, %s);
		triggerInput(pass, %s);

		const submitSelectors = [
			"button[type='submit']",
			"input[type='submit']",
			".btn-primary",
			".login-btn"
		];
		let submit = findFirst(submitSelectors);
		if (!submit) {
			const allButtons = Array.from(document.querySelectorAll("button,input[type='submit']"));
			submit = allButtons.find((el) => /登录|登\\s*录|Login|提交|确认|下一步/i.test((el.innerText || el.value || "").trim()));
		}
		if (!submit) return "SUBMIT_NOT_FOUND";
		submit.click();
		return "OK";
	})()`, strconv.Quote(username), strconv.Quote(password))

	var result string
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &result)); err != nil {
		return fmt.Errorf("submit login form failed: %w", err)
	}
	if result != "OK" {
		return fmt.Errorf("submit login form failed: %s", result)
	}
	return nil
}

func (c *BrowserAuthClient) waitForAccessToken(ctx context.Context) (string, error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	lastHref := ""
	for {
		select {
		case <-ctx.Done():
			if strings.TrimSpace(lastHref) != "" {
				return "", fmt.Errorf("hdis browser login timeout: access_token not found (last_url=%s)", lastHref)
			}
			return "", errors.New("hdis browser login timeout: access_token not found")
		case <-ticker.C:
			var href string
			if err := chromedp.Run(ctx, chromedp.Location(&href)); err != nil {
				continue
			}

			href = strings.TrimSpace(href)
			if href != "" {
				lastHref = href
			}

			if token, err := extractAccessTokenFromURL(href); err == nil && token != "" {
				return token, nil
			}

			if isLikelySPARoute(href) {
				token, err := c.readAccessTokenFromStorage(ctx)
				if err == nil && token != "" {
					return token, nil
				}
			}
		}
	}
}

func isLikelySPARoute(href string) bool {
	lower := strings.ToLower(strings.TrimSpace(href))
	if lower == "" {
		return false
	}
	return strings.Contains(lower, "/#/") || strings.Contains(lower, "#/")
}

func (c *BrowserAuthClient) readAccessTokenFromStorage(ctx context.Context) (string, error) {
	script := `(() => {
		const read = (store, key) => {
			try {
				if (!store) return "";
				return (store.getItem(key) || "").trim();
			} catch (_) {
				return "";
			}
		};

		const candidates = [
			read(window.localStorage, "access_token"),
			read(window.sessionStorage, "access_token"),
			read(window.localStorage, "token"),
			read(window.sessionStorage, "token")
		];

		for (const v of candidates) {
			if (v) return v;
		}
		return "";
	})()`

	var token string
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &token)); err != nil {
		return "", err
	}
	return strings.TrimSpace(token), nil
}

func normalizeAuthBaseURL(rawAuthURL string) (string, error) {
	raw := strings.TrimSpace(rawAuthURL)
	if raw == "" {
		return "", errors.New("auth url is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid auth url")
	}

	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func deriveRedirectURI(graphqlURL, webcmdURL string) (string, error) {
	if origin, err := originFromURL(graphqlURL); err == nil {
		return origin, nil
	}
	if origin, err := originFromURL(webcmdURL); err == nil {
		return origin, nil
	}
	return "", errors.New("unable to derive redirect uri from hdis urls")
}

func originFromURL(raw string) (string, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return "", errors.New("empty url")
	}

	parsed, err := url.Parse(clean)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid url")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func buildAuthorizeURL(authURL, clientID, redirectURI, state, nonce string) (string, error) {
	base, err := normalizeAuthBaseURL(authURL)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(clientID) == "" {
		return "", errors.New("client id is required")
	}
	if strings.TrimSpace(redirectURI) == "" {
		return "", errors.New("redirect uri is required")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	parsed.Path = "/connect/authorize"

	q := parsed.Query()
	q.Set("client_id", strings.TrimSpace(clientID))
	q.Set("response_type", "id_token token")
	q.Set("scope", "openid profile api1")
	q.Set("state", strings.TrimSpace(state))
	q.Set("nonce", strings.TrimSpace(nonce))
	q.Set("redirect_uri", strings.TrimSpace(redirectURI))
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

func extractAccessTokenFromURL(href string) (string, error) {
	clean := strings.TrimSpace(href)
	if clean == "" {
		return "", errors.New("empty callback url")
	}

	parsed, err := url.Parse(clean)
	if err != nil {
		return "", err
	}

	if parsed.Fragment != "" {
		if values, parseErr := url.ParseQuery(parsed.Fragment); parseErr == nil {
			if token := strings.TrimSpace(values.Get("access_token")); token != "" {
				return token, nil
			}
		}
	}

	if token := strings.TrimSpace(parsed.Query().Get("access_token")); token != "" {
		return token, nil
	}

	return "", errors.New("access_token not found in callback url")
}

func randomToken(size int) string {
	if size <= 0 {
		size = 8
	}

	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(buf)
}
