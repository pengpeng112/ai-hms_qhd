/**
 * ROCenterPanel -- RO 水处理远程运维中心
 *
 * Static visual shell ported from the ai-hms-demo command-center panel.
 * Process-flow, gauges, remote-control remain blueprint placeholders.
 * Conductivity trend is wired to real API data (recharts).
 * 检测记录 card (WqRecords) renders below the command-center panel.
 */
import { useState, useEffect } from 'react';
import {
  ResponsiveContainer, LineChart, Line, XAxis, YAxis,
  Tooltip, ReferenceLine, CartesianGrid,
} from 'recharts';
import { waterQualityApi } from '@/services/waterQualityApi';
import type { ConductivityPoint } from '@/services/waterQualityApi';
import WqRecords from './WqRecords';
import '@/styles/ro-center.css';

export default function ROCenterPanel() {
  const [cond, setCond] = useState<ConductivityPoint[]>([]);
  useEffect(() => { waterQualityApi.conductivity(7).then(setCond).catch(() => {}); }, []);

  return (
    <div className="ro-center">
      {/* ---------- title ---------- */}
      <h2 style={{ fontSize: 22, fontWeight: 700, letterSpacing: '0.02em' }}>
        水处理远程运维中心
      </h2>
      <div className="ro-sub">
        — 站点/机组接 RO 设备后统计 —
      </div>

      <div className="roview">
        <div className="window">
          {/* window chrome */}
          <div className="win-bar">
            <span className="wd" style={{ background: '#ff5f57' }} />
            <span className="wd" style={{ background: '#febc2e' }} />
            <span className="wd" style={{ background: '#28c840' }} />
            <span className="u">RO-CENTER · 水处理远程运维中心</span>
          </div>

          <div className="ro-grid">
            {/* ======== LEFT: unit list + alarms (static sample) ======== */}
            <div className="ro-side">
              <div className="sh">
                <span>机组 · UNITS</span>
                <span>实时</span>
              </div>

              <div className="ro-unit sel">
                <span className="ud run" />
                <span className="un">RO-A1<small>中心机房</small></span>
                <span className="ust st-run">运行</span>
              </div>
              <div className="ro-unit">
                <span className="ud run" />
                <span className="un">RO-A2<small>中心机房</small></span>
                <span className="ust st-run">运行</span>
              </div>
              <div className="ro-unit">
                <span className="ud flush" />
                <span className="un">RO-B1<small>三楼分区</small></span>
                <span className="ust st-flush">冲洗中</span>
              </div>
              <div className="ro-unit">
                <span className="ud alarm" />
                <span className="un">RO-C1<small>新院区</small></span>
                <span className="ust st-alarm">预警</span>
              </div>

              <div className="ro-alarms">
                <div className="ro-al w">
                  <b>膜污染趋势</b>RO-A1 ΔP↑
                  <time>实时</time>
                </div>
                <div className="ro-al c">
                  <b>余氯偏高</b>RO-C1 0.12
                  <time>2min</time>
                </div>
                <div className="ro-al">
                  <b>反洗完成</b>RO-B1 已恢复
                  <time>09:14</time>
                </div>
              </div>
            </div>

            {/* ======== RIGHT: main telemetry (all blueprint) ======== */}
            <div className="ro-main">
              {/* --- process flow (blueprint) --- */}
              <div className="ro-blueprint">
                <span className="ro-bp-badge">待 RO 直采接入</span>
                <div className="ro-flow">
                  <div className="fl-h">
                    <span className="live"><i />LIVE</span>
                    水机工作状况 · 工艺流程逐秒回传
                    <span className="rg">RO-A1 / 中心机房</span>
                  </div>
                  <div className="fl-track">
                    <div className="fnode">
                      <div className="fic">🛢️</div>
                      <div className="fnm">原水箱</div>
                      <div className="fvl">浊度 0.4</div>
                    </div>
                    <div className="fpipe" />
                    <div className="fnode">
                      <div className="fic">▦</div>
                      <div className="fnm">预处理</div>
                      <div className="fvl">砂·炭·软化</div>
                    </div>
                    <div className="fpipe" />
                    <div className="fnode">
                      <div className="fic">⛁</div>
                      <div className="fnm">保安过滤</div>
                      <div className="fvl">5 µm</div>
                    </div>
                    <div className="fpipe" />
                    <div className="fnode">
                      <div className="fic">
                        <span className="spin" style={{ display: 'inline-block' }}>⚙️</span>
                      </div>
                      <div className="fnm">高压泵</div>
                      <div className="fvl g">14.8 bar</div>
                    </div>
                    <div className="fpipe" />
                    <div className="fnode ro">
                      <div className="fic">▤</div>
                      <div className="fnm">RO 膜组</div>
                      <div className="fvl w">ΔP 1.6↑</div>
                    </div>
                    <div className="fpipe" />
                    <div className="fnode out">
                      <div className="fic">💧</div>
                      <div className="fnm">产水箱</div>
                      <div className="fvl g">8.6 µS/cm</div>
                    </div>
                  </div>
                  <div className="conc">
                    在线电导 · 压力 · 流量 · 温度 · 余氯 · 浊度 — 全程逐秒回传中心
                  </div>
                </div>
              </div>

              {/* --- gauges (blueprint) --- */}
              <div className="ro-blueprint" style={{ marginTop: 13 }}>
                <span className="ro-bp-badge">待 RO 直采接入</span>
                <div className="ro-gaug">
                  <div className="gau ok">
                    <div className="gl">产水电导率</div>
                    <div className="gv">8.6<small>µS/cm</small></div>
                    <div className="gbar"><i style={{ width: '16%' }} /></div>
                  </div>
                  <div className="gau ok">
                    <div className="gl">脱盐率</div>
                    <div className="gv">98.7<small>%</small></div>
                    <div className="gbar"><i style={{ width: '99%' }} /></div>
                  </div>
                  <div className="gau warn">
                    <div className="gl">膜压差 ΔP</div>
                    <div className="gv">1.6<small>bar</small></div>
                    <div className="gbar"><i style={{ width: '64%' }} /></div>
                  </div>
                  <div className="gau ok">
                    <div className="gl">系统回收率</div>
                    <div className="gv">75<small>%</small></div>
                    <div className="gbar"><i style={{ width: '75%' }} /></div>
                  </div>
                </div>
              </div>

              {/* --- lower row: conductivity trend (REAL) + remote controls (blueprint) --- */}
              <div className="ro-low" style={{ marginTop: 13 }}>
                {/* Conductivity trend -- real data via API */}
                <div className="ro-trend">
                  <div className="tt">
                    <span>透析液电导 · 近7日（引用监控）</span>
                    <span>正常 13–15 mS/cm</span>
                  </div>
                  <ResponsiveContainer width="100%" height={100}>
                    <LineChart data={cond} margin={{ top: 8, right: 12, bottom: 0, left: -16 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                      <XAxis dataKey="day" tick={{ fontSize: 11, fill: '#8ca0b3' }} />
                      <YAxis domain={[10, 18]} tick={{ fontSize: 11, fill: '#8ca0b3' }} />
                      <Tooltip />
                      <ReferenceLine y={13} stroke="rgba(40,200,64,0.5)" strokeDasharray="4 4" label={{ value: '13', position: 'left', fontSize: 10 }} />
                      <ReferenceLine y={15} stroke="rgba(255,91,110,0.5)" strokeDasharray="4 4" label={{ value: '15', position: 'left', fontSize: 10 }} />
                      <Line
                        type="monotone" dataKey="value" stroke="#19E3C8" strokeWidth={2}
                        dot={(props) => {
                          const { cx, cy, payload, index } = props as { cx?: number; cy?: number; payload?: ConductivityPoint; index?: number };
                          if (cx == null || cy == null) return <circle key={index} />;
                          const color = payload?.inRange === false ? '#ff5b6e' : '#19E3C8';
                          return <circle key={index} cx={cx} cy={cy} r={4} fill={color} stroke={color} />;
                        }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                {/* remote-control buttons (disabled) -- blueprint */}
                <div className="ro-blueprint" style={{ flex: 'none' }}>
                <span className="ro-bp-badge">待 RO 直采接入</span>
                  <div className="ro-ctrl">
                    <div className="ch">远程控制 · REMOTE</div>
                    <button
                      className="ro-btn"
                      disabled
                      onClick={(e) => e.preventDefault()}
                    >
                      <span className="bi">🚿</span>一键正冲洗
                      <span className="bk">FLUSH</span>
                    </button>
                    <button
                      className="ro-btn"
                      disabled
                      onClick={(e) => e.preventDefault()}
                    >
                      <span className="bi">🧪</span>化学清洗 CIP
                      <span className="bk">CIP</span>
                    </button>
                    <button
                      className="ro-btn"
                      disabled
                      onClick={(e) => e.preventDefault()}
                    >
                      <span className="bi">🎚️</span>参数远程下发
                      <span className="bk">SET</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* ======== 检测记录 (real data) ======== */}
      <WqRecords />

      <div className="ro-foot">
        * 水处理远程监测与控制为统一愿景蓝图,按实施路线图分阶段落地。
      </div>
    </div>
  );
}
