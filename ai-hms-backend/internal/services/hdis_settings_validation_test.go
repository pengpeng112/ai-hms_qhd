package services

import "testing"

func TestValidateHDISSettingsInput(t *testing.T) {
	t.Run("all required fields provided", func(t *testing.T) {
		req := HdisIntegrationSettingsUpdateRequest{
			WebcmdURL:       "https://a/webcmd",
			GraphqlURL:      "https://a/pygql",
			AuthURL:         "https://a/token",
			ClientID:        "client-id",
			ServiceUsername: "svc",
			ServicePassword: "secret",
		}

		missing := validateHDISSettingsInput(req, false)
		if len(missing) != 0 {
			t.Fatalf("expected no missing fields, got %v", missing)
		}
	})

	t.Run("creating new setting requires password", func(t *testing.T) {
		req := HdisIntegrationSettingsUpdateRequest{
			WebcmdURL:       "https://a/webcmd",
			GraphqlURL:      "https://a/pygql",
			AuthURL:         "https://a/token",
			ClientID:        "client-id",
			ServiceUsername: "svc",
			ServicePassword: "",
		}

		missing := validateHDISSettingsInput(req, false)
		if len(missing) != 1 || missing[0] != "servicePassword" {
			t.Fatalf("expected [servicePassword], got %v", missing)
		}
	})

	t.Run("updating existing setting can omit password", func(t *testing.T) {
		req := HdisIntegrationSettingsUpdateRequest{
			WebcmdURL:       "https://a/webcmd",
			GraphqlURL:      "https://a/pygql",
			AuthURL:         "https://a/token",
			ClientID:        "client-id",
			ServiceUsername: "svc",
			ServicePassword: "",
		}

		missing := validateHDISSettingsInput(req, true)
		if len(missing) != 0 {
			t.Fatalf("expected no missing fields, got %v", missing)
		}
	})

	t.Run("missing multiple required fields", func(t *testing.T) {
		req := HdisIntegrationSettingsUpdateRequest{}
		missing := validateHDISSettingsInput(req, false)
		if len(missing) == 0 {
			t.Fatalf("expected missing fields, got none")
		}
	})
}
