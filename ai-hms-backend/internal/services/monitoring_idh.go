package services

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/idh"
)

// ── 包级 scorer（service 每请求新建，故 scorer 必须包级单例）──
// pkgIDHScorer 包级 IDH 评分器（atomic 以消除启动期写 vs 请求期读的竞态）。
// 存 *idh.Scorer（指针包装）以支持 nil 重置。
var pkgIDHScorer atomic.Value

func init() {
	stub := idh.Scorer(idh.StubScorer{})
	pkgIDHScorer.Store(&stub)
}

// SetIDHScorer 全局装载 IDH 评分器（main 启动期按配置注入 HTTPScorer）。插拔点。
func SetIDHScorer(s idh.Scorer) {
	if s != nil {
		pkgIDHScorer.Store(&s)
	}
}

func getIDHScorer() idh.Scorer {
	v := pkgIDHScorer.Load()
	if v == nil {
		return idh.StubScorer{}
	}
	sp := v.(*idh.Scorer)
	if sp == nil {
		return idh.StubScorer{}
	}
	return *sp
}

// ── 缓存 / in-flight / 并发 ──
const (
	idhCacheTTL     = 60 * time.Second
	idhScoreTimeout = 2 * time.Second
	idhConcurrency  = 4
)

type idhCacheKey struct {
	tenantID    int64
	treatmentID int64
}

type idhCacheEntry struct {
	result idh.RiskResult
	exp    time.Time
}

var idhCache = struct {
	mu sync.RWMutex
	m  map[idhCacheKey]idhCacheEntry
}{m: map[idhCacheKey]idhCacheEntry{}}

var idhInflight = struct {
	mu  sync.Mutex
	set map[idhCacheKey]struct{}
}{set: map[idhCacheKey]struct{}{}}

var idhSem = make(chan struct{}, idhConcurrency)

// resetIDHStateForTest 仅测试用：清空缓存/在飞/并发与 scorer。
func resetIDHStateForTest() {
	idhCache.mu.Lock()
	idhCache.m = map[idhCacheKey]idhCacheEntry{}
	idhCache.mu.Unlock()
	idhInflight.mu.Lock()
	idhInflight.set = map[idhCacheKey]struct{}{}
	idhInflight.mu.Unlock()
	stub := idh.Scorer(idh.StubScorer{})
	pkgIDHScorer.Store(&stub)
}

// idhBasic 由 GetLiveData 用已取数据组装，传给刷新（避免协程再查基本信息）。
type idhBasic struct {
	Gender         *int
	Age            *float64
	DryWeight      *float64
	DialysisMethod *string
	PreWeight      *float64
	PreSBP         *float64
	PreDBP         *float64
	SBP            *float64
	DBP            *float64
}

// dmlogWindowRow IDH 窗口查询行（仅取 legacy Device_DMLog 已确认存在的列；
// 其余训练列在 Sample 留 nil→null，待子项目B按真实 schema 扩充）。
type dmlogWindowRow struct {
	LogTime          time.Time `gorm:"column:LogTime"`
	TMP              *float64  `gorm:"column:TMP"`
	UFVolume         *float64  `gorm:"column:UFVolume"`
	VenousPressure   *float64  `gorm:"column:VenousPressure"`
	ArterialPressure *float64  `gorm:"column:ArterialPressure"`
	BF               *float64  `gorm:"column:BF"`
	Conductivity     *float64  `gorm:"column:Conductivity"`
	TreatmentTime    *float64  `gorm:"column:TreatmentTime"`
	UFSetVolume      *float64  `gorm:"column:UFSetVolume"`
}

// buildIDHInput 纯函数：DMLog 窗口行（升序）+ 基本信息 → idh.RiskInput。
func buildIDHInput(treatmentID int64, accessType string, rows []dmlogWindowRow, b idhBasic) idh.RiskInput {
	window := make([]idh.Sample, 0, len(rows))
	for i := range rows {
		r := rows[i]
		lt := float64(r.LogTime.Unix())
		window = append(window, idh.Sample{
			LogTime:          &lt,
			TMP:              r.TMP,
			UFVolume:         r.UFVolume,
			VenousPressure:   r.VenousPressure,
			ArterialPressure: r.ArterialPressure,
			BF:               r.BF,
			Conductivity:     r.Conductivity,
			TreatmentTime:    r.TreatmentTime,
			UFSetVolume:      r.UFSetVolume,
		})
	}
	return idh.RiskInput{
		TreatmentID: treatmentID,
		AccessType:  accessType,
		Window:      window,
		Basic: idh.BasicInfo{
			Gender:         b.Gender,
			Age:            b.Age,
			DialysisMethod: b.DialysisMethod,
			DryWeight:      b.DryWeight,
			PreWeight:      b.PreWeight,
			PreSBP:         b.PreSBP,
			PreDBP:         b.PreDBP,
			SBP:            b.SBP,
			DBP:            b.DBP,
		},
	}
}

// loadIDHWindow 取某治疗最近 30 行 DMLog（升序）。db 不可用返回 nil。
func loadIDHWindow(tenantID, treatmentID int64) []dmlogWindowRow {
	db := database.GetDB()
	if db == nil {
		return nil
	}
	var rows []dmlogWindowRow
	db.Table(`"Device_DMLog"`).
		Select(`"LogTime","TMP","UFVolume","VenousPressure","ArterialPressure","BF","Conductivity","TreatmentTime","UFSetVolume"`).
		Where(`"TenantId" = ? AND "TreatmentId" = ?`, tenantID, treatmentID).
		Order(`"LogTime" DESC`).
		Limit(30).
		Find(&rows)
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows
}

// refreshIDHNow 同步打分并写缓存（供 triggerIDHRefresh 协程 / 测试调用）。
func refreshIDHNow(tenantID, treatmentID int64, accessType string, basic idhBasic) {
	rows := loadIDHWindow(tenantID, treatmentID)
	in := buildIDHInput(treatmentID, accessType, rows, basic)
	ctx, cancel := context.WithTimeout(context.Background(), idhScoreTimeout)
	defer cancel()
	res := getIDHScorer().Score(ctx, in)
	idhCache.mu.Lock()
	idhCache.m[idhCacheKey{tenantID: tenantID, treatmentID: treatmentID}] = idhCacheEntry{result: res, exp: time.Now().Add(idhCacheTTL)}
	idhCache.mu.Unlock()
}

// triggerIDHRefresh 异步刷新（Stub 跳过、in-flight 去重、并发上限）。fire-and-forget。
func triggerIDHRefresh(tenantID, treatmentID int64, accessType string, basic idhBasic) {
	if _, isStub := getIDHScorer().(idh.StubScorer); isStub {
		return
	}
	key := idhCacheKey{tenantID: tenantID, treatmentID: treatmentID}
	idhInflight.mu.Lock()
	if _, busy := idhInflight.set[key]; busy {
		idhInflight.mu.Unlock()
		return
	}
	idhInflight.set[key] = struct{}{}
	idhInflight.mu.Unlock()

	go func() {
		defer func() {
			idhInflight.mu.Lock()
			delete(idhInflight.set, key)
			idhInflight.mu.Unlock()
		}()
		idhSem <- struct{}{}
		defer func() { <-idhSem }()
		refreshIDHNow(tenantID, treatmentID, accessType, basic)
	}()
}

// lookupIDHCached 读缓存（stale-while-revalidate）；缺失/过期触发异步刷新。永不阻塞。
func lookupIDHCached(tenantID, treatmentID int64, accessType string, basic idhBasic) idh.RiskResult {
	key := idhCacheKey{tenantID: tenantID, treatmentID: treatmentID}
	idhCache.mu.RLock()
	entry, ok := idhCache.m[key]
	idhCache.mu.RUnlock()

	if !ok || time.Now().After(entry.exp) {
		triggerIDHRefresh(tenantID, treatmentID, accessType, basic)
	}
	if ok {
		return entry.result
	}
	return idh.RiskResult{Available: false}
}
