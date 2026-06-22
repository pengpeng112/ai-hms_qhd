import { useState, useEffect, useCallback } from 'react'
import { Scale, Activity, Loader2, CheckCircle2, AlertTriangle } from 'lucide-react'
import { message, Button, Modal, InputNumber, Select, Radio, Checkbox, Card, Tag, Space } from 'antd'
import { SectionHeader } from '@/components/ui'
import { dryWeightApi, type DryWeightAssessment, type DwCurrentData } from '@/services/dryWeightApi'
import type { TabProps } from '../types'
import dayjs from 'dayjs'

const PHASE_LABEL: Record<string, { label: string; color: string }> = {
  induction: { label: '诱导期', color: 'orange' },
  maintenance: { label: '维持期', color: 'green' },
}

const DECISION_LABEL: Record<string, string> = {
  hold: '维持',
  lower: '下调',
  raise: '上调',
}

const MAX_PHASE_ADJUST: Record<string, number> = {
  induction: 1.0,
  maintenance: 0.5,
}

export default function DryWeightTab({ patient }: TabProps) {
  const [assessments, setAssessments] = useState<DryWeightAssessment[]>([])
  const [current, setCurrent] = useState<DwCurrentData | null>(null)
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [confirmModalOpen, setConfirmModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  const [form, setForm] = useState({
    assessType: 'daily' as string,
    phase: current?.phase || 'induction',
    sbp: undefined as number | undefined,
    dbp: undefined as number | undefined,
    heartRate: undefined as number | undefined,
    edema: false, palpitation: false, heartFailure: false, cramp: false,
    ctr: undefined as number | undefined,
    actr: undefined as number | undefined,
    postWeight: undefined as number | undefined,
    targetWeight: undefined as number | undefined,
    decision: 'hold' as string,
    adjustKg: undefined as number | undefined,
    rnaSetting: undefined as number | undefined,
  })

  const [confirmForm, setConfirmForm] = useState({
    dryWeight: current?.dryWeight ?? 0,
    phase: current?.phase || 'maintenance',
    actr: undefined as number | undefined,
    ctr: undefined as number | undefined,
  })

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [rows, cur] = await Promise.all([
        dryWeightApi.listAssessments(patient.id),
        dryWeightApi.current(patient.id),
      ])
      setAssessments(rows)
      setCurrent(cur)
      setConfirmForm((f) => ({ ...f, dryWeight: cur.dryWeight ?? 0, phase: cur.phase ?? 'maintenance' }))
    } catch { /* ignore */ }
    setLoading(false)
  }, [patient.id])

  useEffect(() => { fetchData() }, [fetchData])

  // 主判据实时预判
  const mainMetPre = (() => {
    if (!form.sbp || !form.dbp || !form.heartRate) return false
    if (form.sbp < 110 || form.dbp < 60) return false
    if (form.heartRate < 60 || form.heartRate > 100) return false
    if (form.edema || form.palpitation || form.heartFailure || form.cramp) return false
    return true
  })()

  const maxPhaseAdj = MAX_PHASE_ADJUST[form.phase] || 0.5
  const adjOverLimit = form.adjustKg != null && form.adjustKg > maxPhaseAdj

  async function handleAssess() {
    if (!form.sbp || !form.dbp || !form.heartRate) { message.error('血压心率必填'); return }
    setSaving(true)
    try {
      await dryWeightApi.assess(patient.id, form)
      message.success('评估记录成功')
      setModalOpen(false)
      fetchData()
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message || '评估失败')
    }
    setSaving(false)
  }

  async function handleConfirm() {
    if (confirmForm.dryWeight <= 0) { message.error('干体重必须>0'); return }
    setSaving(true)
    try {
      const result = await dryWeightApi.confirm(patient.id, confirmForm)
      message.success(`确诊干体重 ${result.dryWeight} kg` + (result.legacyPlanUpdated ? '，已写回方案' : '（无启用方案未写回）'))
      setConfirmModalOpen(false)
      fetchData()
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message || '确定失败')
    }
    setSaving(false)
  }

  if (loading) {
    return <div className="py-12 text-center text-slate-400"><Loader2 size={20} className="inline animate-spin" /> 加载中…</div>
  }

  return (
    <div className="space-y-4">
      <SectionHeader icon={Scale} title="干体重评估" />

      {/* Current dry weight card */}
      {current && (
        <Card size="small" bordered className="rounded-xl bg-gradient-to-r from-blue-50 to-white border-blue-200">
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <span className="font-bold text-slate-800 text-[15px]">
                  {current.dryWeight != null ? `${current.dryWeight.toFixed(1)} kg` : '尚未确定'}
                </span>
                <Tag color={PHASE_LABEL[current.phase]?.color || 'default'} className="text-[11px]">
                  {PHASE_LABEL[current.phase]?.label || current.phase}
                </Tag>
              </div>
              <div className="text-[12px] text-slate-400 space-y-0.5">
                {current.standardActr != null && <div>锚定 ACTR：{current.standardActr.toFixed(3)}</div>}
                <div>建议 RNa：{current.suggestedRNa.toFixed(3)}</div>
              </div>
            </div>
            <Button size="small" type="primary" onClick={() => setConfirmModalOpen(true)}>
              确定/修改干体重
            </Button>
          </div>
        </Card>
      )}

      <div className="flex gap-2">
        <Button type="primary" icon={<Activity size={16} />} onClick={() => setModalOpen(true)}>
          录入评估
        </Button>
      </div>

      {/* Main met pre-judgment */}
      {form.sbp != null && (
        <div className={`flex items-center gap-2 rounded-xl px-4 py-2.5 text-[13px] font-bold ${mainMetPre ? 'border border-emerald-200 bg-emerald-50 text-emerald-700' : 'border border-orange-200 bg-orange-50 text-orange-700'}`}>
          {mainMetPre ? <CheckCircle2 size={16} className="text-emerald-500" /> : <AlertTriangle size={16} className="text-orange-500" />}
          {mainMetPre ? '主判据预判：全部达标，可确认干体重' : '主判据预判：未达标（血压/心率/症状）'}
        </div>
      )}

      {/* Assessment history */}
      {assessments.length === 0 ? (
        <div className="text-[12px] text-slate-400 py-2">暂无评估记录</div>
      ) : (
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          {assessments.map((a) => (
            <Card key={a.id} size="small" bordered className={a.mainMet ? 'rounded-xl border-emerald-200' : 'rounded-xl border-orange-200'}>
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  <Tag color={a.mainMet ? 'green' : 'orange'} className="text-[11px]">
                    {a.mainMet ? '主判据达标' : '未达标'}
                  </Tag>
                  <Tag color={PHASE_LABEL[a.phase]?.color || 'default'} className="text-[11px]">
                    {PHASE_LABEL[a.phase]?.label || a.phase}
                  </Tag>
                  <span className="text-[12px] text-slate-500">{a.assessType === 'cycle' ? '周期' : '日常'}</span>
                </div>
                <span className="text-[11px] text-slate-400">{dayjs(a.createdAt).format('MM-DD HH:mm')}</span>
              </div>
              <div className="grid grid-cols-4 gap-1 text-[12px] text-slate-500">
                {a.sbp != null && <span>收缩压 {a.sbp}</span>}
                {a.dbp != null && <span>舒张压 {a.dbp}</span>}
                {a.heartRate != null && <span>心率 {a.heartRate}</span>}
                {a.ctr != null && <span>CTR {a.ctr.toFixed(3)}</span>}
                {a.actr != null && <span>ACTR {a.actr.toFixed(3)}</span>}
                {a.adjustKg != null && <span>调整 {a.adjustKg}kg</span>}
                {a.decision && <span>{DECISION_LABEL[a.decision] || a.decision}</span>}
              </div>
            </Card>
          ))}
        </Space>
      )}

      {/* Assess Modal */}
      <Modal title="录入干体重评估" open={modalOpen} onCancel={() => setModalOpen(false)}
        onOk={handleAssess} confirmLoading={saving} okText="记录" width={560} destroyOnClose>
        <Space direction="vertical" style={{ width: '100%' }} size={10}>
          <div>
            <div className="text-[13px] font-bold mb-1">评估类型</div>
            <Radio.Group value={form.assessType} onChange={(e) => setForm({ ...form, assessType: e.target.value })}>
              <Radio.Button value="daily">日常评估</Radio.Button>
              <Radio.Button value="cycle">周期评估</Radio.Button>
            </Radio.Group>
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">阶段</div>
            <Select style={{ width: '100%' }} value={form.phase}
              onChange={(v) => setForm({ ...form, phase: v })}
              options={[{ label: '诱导期（≤1kg/次）', value: 'induction' }, { label: '维持期（≤0.5kg/次）', value: 'maintenance' }]}
            />
          </div>
          <div className="grid grid-cols-3 gap-2">
            <div><div className="text-[12px] font-bold mb-1">收缩压</div><InputNumber min={60} max={250} style={{ width: '100%' }} placeholder="mmHg"
              value={form.sbp} onChange={(v) => setForm({ ...form, sbp: v ?? undefined })} /></div>
            <div><div className="text-[12px] font-bold mb-1">舒张压</div><InputNumber min={30} max={150} style={{ width: '100%' }} placeholder="mmHg"
              value={form.dbp} onChange={(v) => setForm({ ...form, dbp: v ?? undefined })} /></div>
            <div><div className="text-[12px] font-bold mb-1">心率</div><InputNumber min={30} max={200} style={{ width: '100%' }} placeholder="bpm"
              value={form.heartRate} onChange={(v) => setForm({ ...form, heartRate: v ?? undefined })} /></div>
          </div>
          <div>
            <div className="text-[12px] font-bold mb-1">症状</div>
            <div className="flex flex-wrap gap-2">
              <Checkbox checked={form.edema} onChange={(e) => setForm({ ...form, edema: e.target.checked })}>显性水肿</Checkbox>
              <Checkbox checked={form.palpitation} onChange={(e) => setForm({ ...form, palpitation: e.target.checked })}>心慌气短</Checkbox>
              <Checkbox checked={form.heartFailure} onChange={(e) => setForm({ ...form, heartFailure: e.target.checked })}>心衰</Checkbox>
              <Checkbox checked={form.cramp} onChange={(e) => setForm({ ...form, cramp: e.target.checked })}>肌肉痉挛</Checkbox>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div><div className="text-[12px] font-bold mb-1">CTR</div><InputNumber style={{ width: '100%' }} placeholder="0.52" step={0.01}
              value={form.ctr} onChange={(v) => setForm({ ...form, ctr: v ?? undefined })} /></div>
            <div><div className="text-[12px] font-bold mb-1">ACTR</div><InputNumber style={{ width: '100%' }} placeholder="0.35" step={0.01}
              value={form.actr} onChange={(v) => setForm({ ...form, actr: v ?? undefined })} /></div>
          </div>
          <div>
            <div className="text-[12px] font-bold mb-1">决策</div>
            <Radio.Group value={form.decision} onChange={(e) => setForm({ ...form, decision: e.target.value })}>
              <Radio.Button value="hold">维持</Radio.Button>
              <Radio.Button value="lower">下调</Radio.Button>
              <Radio.Button value="raise">上调</Radio.Button>
            </Radio.Group>
          </div>
          <div>
            <div className="text-[12px] font-bold mb-1">
              调整幅度 kg（{form.phase === 'induction' ? '诱导期 ≤1.0' : '维持期 ≤0.5'}）
            </div>
            <InputNumber style={{ width: '100%' }} min={0} max={5} step={0.1} placeholder="0.5"
              value={form.adjustKg} onChange={(v) => setForm({ ...form, adjustKg: v ?? undefined })}
              status={adjOverLimit ? 'error' : undefined} />
            {adjOverLimit && <div className="text-[12px] text-rose-500 mt-1">超过 {form.phase} 限幅 {maxPhaseAdj} kg</div>}
          </div>
          <div>
            <div className="text-[12px] font-bold mb-1">RNa 设置</div>
            <InputNumber style={{ width: '100%' }} min={0.8} max={1.5} step={0.005}
              value={form.rnaSetting} onChange={(v) => setForm({ ...form, rnaSetting: v ?? undefined })} />
          </div>
        </Space>
      </Modal>

      {/* Confirm Modal */}
      <Modal title="确定干体重" open={confirmModalOpen} onCancel={() => setConfirmModalOpen(false)}
        onOk={handleConfirm} confirmLoading={saving} okText="确定" destroyOnClose>
        <Space direction="vertical" style={{ width: '100%' }} size={10}>
          <div>
            <div className="text-[13px] font-bold mb-1">确定干体重（kg）</div>
            <InputNumber style={{ width: '100%' }} min={0} max={200} step={0.1}
              value={confirmForm.dryWeight} onChange={(v) => setConfirmForm({ ...confirmForm, dryWeight: v ?? 0 })} />
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">阶段</div>
            <Select style={{ width: '100%' }} value={confirmForm.phase}
              onChange={(v) => setConfirmForm({ ...confirmForm, phase: v })}
              options={[{ label: '维持期', value: 'maintenance' }, { label: '诱导期', value: 'induction' }]} />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div><div className="text-[12px] font-bold mb-1">锚定 ACTR</div>
              <InputNumber style={{ width: '100%' }} step={0.01} value={confirmForm.actr}
                onChange={(v) => setConfirmForm({ ...confirmForm, actr: v ?? undefined })} /></div>
            <div><div className="text-[12px] font-bold mb-1">锚定 CTR</div>
              <InputNumber style={{ width: '100%' }} step={0.01} value={confirmForm.ctr}
                onChange={(v) => setConfirmForm({ ...confirmForm, ctr: v ?? undefined })} /></div>
          </div>
          {current?.confirmedAt && <div className="text-[12px] text-slate-400">上次确认：{dayjs(current.confirmedAt).format('YYYY-MM-DD HH:mm')}</div>}
          <div className="text-[12px] text-slate-400">确认将同时写回老库方案表</div>
        </Space>
      </Modal>
    </div>
  )
}
