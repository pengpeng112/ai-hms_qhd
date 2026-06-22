// Nursing Doc Tab - 护理文书（C1）：量表评估 / 护理记录 / 护理计划

import { useState, useEffect, useCallback, useMemo } from 'react'
import { message } from 'antd'
import { ClipboardList, Plus, Loader2, ShieldX, AlertTriangle, ClipboardCheck, NotebookPen, Loader } from 'lucide-react'
import { SectionHeader, DetailCard } from '@/components/ui'
import { getUserInfo } from '@/utils/token'
import {
  getNursingScales,
  recordNursingScale,
  recordNursingDoc,
  getNursingDocs,
  type NursingScale,
  type NursingDoc,
  type NursingRiskLevel,
} from '@/services/nursingApi'
import type { Patient } from '@/types/original'

interface NursingTabProps {
  patient: Patient
}

const RISK_META: Record<NursingRiskLevel, { label: string; badge: string; dot: string }> = {
  high: { label: '高风险', badge: 'bg-red-50 text-red-600 border-red-100', dot: 'bg-red-500' },
  moderate: { label: '中风险', badge: 'bg-amber-50 text-amber-600 border-amber-100', dot: 'bg-amber-500' },
  low: { label: '低风险', badge: 'bg-green-50 text-green-600 border-green-100', dot: 'bg-green-500' },
  none: { label: '无风险', badge: 'bg-slate-100 text-slate-500 border-slate-200', dot: 'bg-slate-400' },
}

const DOC_TYPE_LABEL: Record<string, string> = { scale: '量表评估', record: '护理记录', plan: '护理计划' }

function fmt(ts?: string): string {
  if (!ts) return '-'
  const d = new Date(ts)
  if (isNaN(d.getTime())) return ts
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export default function NursingTab({ patient }: NursingTabProps) {
  const [scales, setScales] = useState<NursingScale[]>([])
  const [docs, setDocs] = useState<NursingDoc[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  // 量表录入
  const [activeScale, setActiveScale] = useState<string>('')
  const [scaleItems, setScaleItems] = useState<Record<string, number>>({})

  // 护理记录 / 计划
  const [record, setRecord] = useState({ observation: '', operation: '', education: '', handover: '' })
  const [plan, setPlan] = useState({ problem: '', measure: '', evaluation: '' })

  const load = useCallback(async () => {
    if (!patient.id) return
    setLoading(true)
    try {
      const data = await getNursingDocs({ patientId: String(patient.id) })
      setDocs(data)
    } catch (err) {
      console.error('加载护理文书失败:', err)
      message.error('加载护理文书失败')
    } finally {
      setLoading(false)
    }
  }, [patient.id])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    getNursingScales().then((s) => {
      setScales(s)
      if (s.length > 0) setActiveScale(s[0].scaleType)
    }).catch(() => {})
  }, [])

  const currentScale = useMemo(() => scales.find((s) => s.scaleType === activeScale), [scales, activeScale])

  // 切换量表时重置选择
  useEffect(() => {
    setScaleItems({})
  }, [activeScale])

  // 实时算分 + 风险预览
  const liveScore = useMemo(() => {
    if (!currentScale) return null
    let total = 0
    for (const item of currentScale.items) {
      const v = scaleItems[item.key]
      if (v === undefined) return null // 未填全
      total += v
    }
    const band = currentScale.bands.find((b) => total >= b.min && total <= b.max)
    return { total, band }
  }, [currentScale, scaleItems])

  const nurse = getUserInfo()

  const handleSubmitScale = async () => {
    if (!currentScale) return
    if (liveScore === null) {
      message.warning('请完成所有量表条目')
      return
    }
    setSaving(true)
    try {
      await recordNursingScale({
        patientId: String(patient.id),
        scaleType: currentScale.scaleType,
        items: scaleItems,
        nurseId: nurse?.id,
        nurseName: nurse?.name,
      })
      message.success('量表已保存')
      if (liveScore.band?.level === 'high') {
        message.warning(`${currentScale.name} 评估为高风险，请制定护理计划`)
      }
      setScaleItems({})
      await load()
    } catch (err) {
      console.error('保存量表失败:', err)
      message.error('保存量表失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSubmitRecord = async () => {
    if (!record.observation && !record.operation && !record.education && !record.handover) {
      message.warning('请至少填写一项护理记录')
      return
    }
    setSaving(true)
    try {
      await recordNursingDoc({
        patientId: String(patient.id),
        docType: 'record',
        content: JSON.stringify(record),
        nurseId: nurse?.id,
        nurseName: nurse?.name,
      })
      message.success('护理记录已保存')
      setRecord({ observation: '', operation: '', education: '', handover: '' })
      await load()
    } catch (err) {
      console.error('保存护理记录失败:', err)
      message.error('保存护理记录失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSubmitPlan = async () => {
    if (!plan.problem) {
      message.warning('请填写护理问题')
      return
    }
    setSaving(true)
    try {
      await recordNursingDoc({
        patientId: String(patient.id),
        docType: 'plan',
        content: JSON.stringify(plan),
        nurseId: nurse?.id,
        nurseName: nurse?.name,
      })
      message.success('护理计划已保存')
      setPlan({ problem: '', measure: '', evaluation: '' })
      await load()
    } catch (err) {
      console.error('保存护理计划失败:', err)
      message.error('保存护理计划失败')
    } finally {
      setSaving(false)
    }
  }

  // 最新各量表高风险提醒
  const highRiskScales = useMemo(() => {
    const seen = new Set<string>()
    const out: NursingDoc[] = []
    for (const d of docs) {
      if (d.docType !== 'scale' || !d.scaleType) continue
      if (seen.has(d.scaleType)) continue
      seen.add(d.scaleType)
      if (d.riskLevel === 'high') out.push(d)
    }
    return out
  }, [docs])

  const scaleName = (t?: string) => scales.find((s) => s.scaleType === t)?.name || t || '量表'

  function renderDocContent(d: NursingDoc) {
    if (!d.content) return null
    let parsed: Record<string, unknown> = {}
    try { parsed = JSON.parse(d.content) } catch { return <span className="text-slate-500">{d.content}</span> }
    if (d.docType === 'record') {
      const map: Record<string, string> = { observation: '观察', operation: '操作', education: '宣教', handover: '交班' }
      return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs">
          {Object.entries(map).filter(([k]) => parsed[k]).map(([k, label]) => (
            <div key={k}><span className="text-slate-400 font-black">{label}：</span><span className="text-slate-600">{String(parsed[k])}</span></div>
          ))}
        </div>
      )
    }
    if (d.docType === 'plan') {
      const map: Record<string, string> = { problem: '问题', measure: '措施', evaluation: '评价' }
      return (
        <div className="space-y-1 text-xs">
          {Object.entries(map).filter(([k]) => parsed[k]).map(([k, label]) => (
            <div key={k}><span className="text-slate-400 font-black">{label}：</span><span className="text-slate-600">{String(parsed[k])}</span></div>
          ))}
        </div>
      )
    }
    return null
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="animate-spin text-blue-500" size={32} />
        <span className="ml-3 text-slate-500">加载中...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in pb-10">
      {/* 高风险提醒 */}
      {highRiskScales.length > 0 && (
        <div className="flex items-center gap-3 px-5 py-3 bg-red-50 border border-red-100 rounded-2xl text-red-600">
          <AlertTriangle size={18} className="shrink-0" />
          <span className="text-sm font-black">
            高风险：{highRiskScales.map((d) => scaleName(d.scaleType)).join('、')} —— 请制定并执行对应护理计划
          </span>
        </div>
      )}

      {/* 量表评估 */}
      <DetailCard>
        <SectionHeader icon={ClipboardCheck} title="护理量表评估" />
        {scales.length === 0 ? (
          <div className="py-8 text-center text-slate-300 font-bold">暂无可用量表</div>
        ) : (
          <div className="mt-4">
            {/* 量表切换 */}
            <div className="flex gap-2 flex-wrap mb-4">
              {scales.map((s) => (
                <button
                  key={s.scaleType}
                  onClick={() => setActiveScale(s.scaleType)}
                  className={`px-4 py-1.5 rounded-lg text-xs font-black border transition-all ${activeScale === s.scaleType ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-slate-500 border-slate-200 hover:border-blue-300'}`}
                >
                  {s.name}
                </button>
              ))}
            </div>

            {currentScale && (
              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {currentScale.items.map((item) => (
                    <div key={item.key}>
                      <label className="block text-xs font-black text-slate-500 mb-1.5">{item.label}</label>
                      <select
                        value={scaleItems[item.key] ?? ''}
                        onChange={(e) => setScaleItems((cur) => ({ ...cur, [item.key]: Number(e.target.value) }))}
                        className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm"
                      >
                        <option value="">请选择</option>
                        {item.options.map((o, i) => (
                          <option key={i} value={o.value}>{o.label}（{o.value}）</option>
                        ))}
                      </select>
                    </div>
                  ))}
                </div>

                {/* 实时算分 */}
                <div className="flex items-center justify-between flex-wrap gap-3 border-t border-slate-100 pt-4">
                  <div className="flex items-center gap-3">
                    <span className="text-xs font-black text-slate-400">实时评分</span>
                    {liveScore === null ? (
                      <span className="text-sm text-slate-400">填全后自动计算</span>
                    ) : (
                      <>
                        <span className="text-2xl font-black text-slate-800">{liveScore.total}</span>
                        {liveScore.band && (
                          <span className={`px-2.5 py-1 rounded-lg text-xs font-black border ${RISK_META[liveScore.band.level]?.badge}`}>
                            {liveScore.band.label}
                          </span>
                        )}
                      </>
                    )}
                  </div>
                  <button
                    onClick={handleSubmitScale}
                    disabled={saving || liveScore === null}
                    className="px-5 py-2 bg-blue-600 text-white rounded-lg text-xs font-black hover:bg-blue-700 transition-all flex items-center gap-1.5 shadow-sm disabled:opacity-50"
                  >
                    {saving ? <Loader size={14} className="animate-spin" /> : <Plus size={14} />} 保存评估
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </DetailCard>

      {/* 护理记录 + 护理计划 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <DetailCard>
          <SectionHeader icon={NotebookPen} title="护理记录" />
          <div className="mt-4 space-y-3">
            {([['observation', '护理观察'], ['operation', '护理操作'], ['education', '健康宣教'], ['handover', '交班记录']] as const).map(([k, label]) => (
              <div key={k}>
                <label className="block text-xs font-black text-slate-500 mb-1.5">{label}</label>
                <textarea
                  rows={2}
                  value={record[k]}
                  onChange={(e) => setRecord((cur) => ({ ...cur, [k]: e.target.value }))}
                  className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm resize-none"
                />
              </div>
            ))}
            <button onClick={handleSubmitRecord} disabled={saving} className="px-5 py-2 bg-blue-600 text-white rounded-lg text-xs font-black hover:bg-blue-700 transition-all flex items-center gap-1.5 shadow-sm disabled:opacity-50">
              <Plus size={14} /> 保存记录
            </button>
          </div>
        </DetailCard>

        <DetailCard>
          <SectionHeader icon={ClipboardList} title="护理计划（问题-措施-评价）" />
          <div className="mt-4 space-y-3">
            <div>
              <label className="block text-xs font-black text-slate-500 mb-1.5">护理问题 <span className="text-red-500">*</span></label>
              <textarea rows={2} value={plan.problem} onChange={(e) => setPlan((c) => ({ ...c, problem: e.target.value }))} className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm resize-none" placeholder="如：有跌倒的危险" />
            </div>
            <div>
              <label className="block text-xs font-black text-slate-500 mb-1.5">护理措施</label>
              <textarea rows={2} value={plan.measure} onChange={(e) => setPlan((c) => ({ ...c, measure: e.target.value }))} className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm resize-none" />
            </div>
            <div>
              <label className="block text-xs font-black text-slate-500 mb-1.5">评价 / 转归</label>
              <textarea rows={2} value={plan.evaluation} onChange={(e) => setPlan((c) => ({ ...c, evaluation: e.target.value }))} className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm resize-none" />
            </div>
            <button onClick={handleSubmitPlan} disabled={saving} className="px-5 py-2 bg-blue-600 text-white rounded-lg text-xs font-black hover:bg-blue-700 transition-all flex items-center gap-1.5 shadow-sm disabled:opacity-50">
              <Plus size={14} /> 保存计划
            </button>
          </div>
        </DetailCard>
      </div>

      {/* 文书历史 */}
      <DetailCard>
        <SectionHeader icon={ClipboardList} title="护理文书历史" />
        <div className="mt-4 space-y-3">
          {docs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-slate-300">
              <ShieldX size={40} className="mb-3 opacity-20" />
              <p className="font-bold">暂无护理文书</p>
            </div>
          ) : (
            docs.map((d) => {
              const risk = d.riskLevel ? RISK_META[d.riskLevel] : undefined
              return (
                <div key={d.id} className={`rounded-2xl border p-4 ${d.riskLevel === 'high' ? 'border-red-200 bg-red-50/40' : 'border-slate-100 bg-white'}`}>
                  <div className="flex items-center justify-between flex-wrap gap-2 mb-2">
                    <div className="flex items-center gap-2">
                      <span className="px-2 py-0.5 rounded-lg text-[10px] font-black border bg-blue-50 text-blue-600 border-blue-100">{DOC_TYPE_LABEL[d.docType]}</span>
                      {d.docType === 'scale' && <span className="text-sm font-black text-slate-800">{scaleName(d.scaleType)}</span>}
                      {d.docType === 'scale' && d.score !== undefined && <span className="text-sm font-black text-slate-600">{d.score} 分</span>}
                      {risk && <span className={`px-2 py-0.5 rounded-lg text-[10px] font-black border ${risk.badge}`}>{risk.label}</span>}
                    </div>
                    <span className="text-[11px] text-slate-400">{d.nurseName || ''} · {fmt(d.recordedAt)}</span>
                  </div>
                  {renderDocContent(d)}
                </div>
              )
            })
          )}
        </div>
      </DetailCard>
    </div>
  )
}
