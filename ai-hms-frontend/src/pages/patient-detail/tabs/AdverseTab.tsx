import { useState, useEffect, useCallback } from 'react'
import { AlertTriangle, Plus, ShieldAlert, Send, Loader2 } from 'lucide-react'
import { message, Modal, Button, Select, DatePicker, Input, Radio, Tag, Space, Card } from 'antd'
import { SectionHeader, DetailCard, LabelValue } from '@/components/ui'
import { adverseEventApi, type AdverseEvent, type AeRegisterBody, type AeStatusBody } from '@/services/adverseEventApi'
import { dictCache, DICT_TYPES } from '@/services/dictApi'
import type { TabProps } from '../types'
import dayjs from 'dayjs'

const SEVERITY_META: Record<string, { label: string; color: string; dot: string }> = {
  mild: { label: '轻', color: 'bg-sky-50 text-sky-600 border-sky-200', dot: 'bg-sky-500' },
  moderate: { label: '中', color: 'bg-orange-50 text-orange-600 border-orange-200', dot: 'bg-orange-500' },
  severe: { label: '重', color: 'bg-rose-50 text-rose-600 border-rose-200', dot: 'bg-rose-500' },
}

const STATUS_META: Record<string, { label: string; color: string }> = {
  registered: { label: '已登记', color: 'default' as const },
  reported: { label: '已上报', color: 'blue' as const },
  acknowledged: { label: '已受理', color: 'cyan' as const },
  processing: { label: '处理中', color: 'processing' as const },
  closed: { label: '已结案', color: 'success' as const },
}

const NEXT_STATUS: Record<string, { label: string; value: string }[]> = {
  reported: [{ label: '受理', value: 'acknowledged' }],
  acknowledged: [
    { label: '处理中', value: 'processing' },
    { label: '结案', value: 'closed' },
  ],
  processing: [{ label: '结案', value: 'closed' }],
}

export default function AdverseTab({ patient }: TabProps) {
  const [events, setEvents] = useState<AdverseEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [reportModalOpen, setReportModalOpen] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [eventTypes, setEventTypes] = useState<{ label: string; value: string }[]>([])

  const [form, setForm] = useState({
    eventType: '',
    severity: 'mild' as string,
    occurredAt: dayjs(),
    description: '',
    handling: '',
    outcome: '',
  })

  const [reportRoles, setReportRoles] = useState([{ role: '护士长', userId: '' }, { role: '科主任', userId: '' }])

  const fetchEvents = useCallback(async () => {
    setLoading(true)
    try {
      const rows = await adverseEventApi.list({ patientId: String(patient.id) })
      setEvents(rows)
    } catch { /* ignore */ }
    setLoading(false)
  }, [patient.id])

  useEffect(() => {
    fetchEvents()
    dictCache.getItems(DICT_TYPES.COMPLICATION).then((items) => {
      setEventTypes(items.map((i) => ({ label: i.name, value: i.name })))
    }).catch(() => {})
  }, [fetchEvents])

  const severeUnreported = events.filter((e) => e.severity === 'severe' && e.status === 'registered')

  async function handleRegister() {
    if (!form.eventType) { message.error('请选择事件分类'); return }
    setSaving(true)
    try {
      const body: AeRegisterBody = {
        patientId: Number(patient.id),
        eventType: form.eventType,
        severity: form.severity,
        occurredAt: form.occurredAt.toISOString(),
        description: form.description,
        handling: form.handling,
        outcome: form.outcome,
      }
      await adverseEventApi.register(body)
      message.success('不良事件登记成功')
      setModalOpen(false)
      resetForm()
      fetchEvents()
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message || '登记失败')
    }
    setSaving(false)
  }

  async function handleReport(id: string) {
    if (reportRoles.some((r) => !r.userId)) { message.error('请填写上报对象的用户ID'); return }
    setSaving(true)
    try {
      await adverseEventApi.report(id, { reportedTo: reportRoles })
      message.success('上报成功')
      setReportModalOpen(null)
      setReportRoles([{ role: '护士长', userId: '' }, { role: '科主任', userId: '' }])
      fetchEvents()
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message || '上报失败')
    }
    setSaving(false)
  }

  async function handleStatus(id: string, body: AeStatusBody) {
    try {
      await adverseEventApi.updateStatus(id, body)
      message.success('状态更新成功')
      fetchEvents()
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message || '状态更新失败')
    }
  }

  function resetForm() {
    setForm({ eventType: '', severity: 'mild', occurredAt: dayjs(), description: '', handling: '', outcome: '' })
  }

  return (
    <div className="space-y-4">
      <SectionHeader icon={AlertTriangle} title="不良事件" />

      {severeUnreported.length > 0 && (
        <div className="flex items-center justify-between rounded-xl border border-rose-200 bg-rose-50 px-4 py-3">
          <div className="flex items-center gap-2 text-[13px] font-bold text-rose-700">
            <ShieldAlert size={18} className="text-rose-500 shrink-0" />
            本患者 {severeUnreported.length} 条严重事件未上报——须 6 小时内上报护士长+科主任
          </div>
        </div>
      )}

      <div className="flex gap-2">
        <Button type="primary" icon={<Plus size={16} />} onClick={() => setModalOpen(true)}>
          登记不良事件
        </Button>
      </div>

      {loading ? (
        <div className="py-12 text-center text-slate-400"><Loader2 size={20} className="inline animate-spin" /> 加载中…</div>
      ) : events.length === 0 ? (
        <div className="py-8 text-center text-slate-400 text-[13px]">暂无不良事件记录</div>
      ) : (
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          {events.map((e) => {
            const sm = SEVERITY_META[e.severity] || { label: e.severity, color: '', dot: 'bg-slate-400' }
            const st = STATUS_META[e.status] || { label: e.status, color: 'default' as const }
            let reportedTo: { role: string; userId: string }[] = []
            try { reportedTo = JSON.parse(e.reportedTo || '[]') } catch { /* ignore */ }
            const nextActions = NEXT_STATUS[e.status] || []

            return (
              <Card key={e.id} size="small" bordered className="rounded-xl">
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <div className={`w-2.5 h-2.5 rounded-full ${sm.dot}`} />
                    <span className="font-bold text-slate-800">{e.eventType}</span>
                    <span className={`px-1.5 py-0.5 rounded text-[11px] font-bold border ${sm.color}`}>{sm.label}</span>
                    <Tag color={st.color} className="text-[11px]">{st.label}</Tag>
                    {e.cqiLinked && <Tag color="purple" className="text-[11px]">CQI</Tag>}
                  </div>
                  <span className="text-[11px] text-slate-400">{dayjs(e.occurredAt).format('YYYY-MM-DD HH:mm')}</span>
                </div>

                {e.description && <DetailCard className="mb-2"><LabelValue label="经过" value={e.description} /></DetailCard>}
                {e.handling && <DetailCard className="mb-2"><LabelValue label="处置" value={e.handling} /></DetailCard>}
                {e.outcome && <DetailCard className="mb-2"><LabelValue label="转归" value={e.outcome} /></DetailCard>}

                {e.reportedAt && (
                  <div className="flex items-center gap-4 text-[12px] text-slate-400 mt-2">
                    <span className="flex items-center gap-1"><Send size={12} /> 已上报：{dayjs(e.reportedAt).format('MM-DD HH:mm')}</span>
                    {reportedTo.map((r, i) => <span key={i}>● {r.role}({r.userId})</span>)}
                    {e.within6h !== undefined && (
                      <span className={e.within6h ? 'text-emerald-600 font-bold' : 'text-rose-600 font-bold'}>
                        {e.within6h ? '6h内达标' : '超时'}
                      </span>
                    )}
                  </div>
                )}

                <div className="flex gap-2 mt-3 flex-wrap">
                  {e.status === 'registered' && (
                    <Button size="small" icon={<Send size={12} />} onClick={() => setReportModalOpen(e.id)}>
                      上报
                    </Button>
                  )}
                  {nextActions.map((a) => (
                    <Button key={a.value} size="small" onClick={() => handleStatus(e.id, { status: a.value })}>
                      {a.label}
                    </Button>
                  ))}
                  {!e.cqiLinked && e.status !== 'registered' && (
                    <Button size="small" onClick={() => handleStatus(e.id, { status: e.status, cqiLinked: true })}>
                      纳入CQI
                    </Button>
                  )}
                  {e.cqiLinked && (
                    <Button size="small" danger onClick={() => handleStatus(e.id, { status: e.status, cqiLinked: false })}>
                      移出CQI
                    </Button>
                  )}
                </div>
              </Card>
            )
          })}
        </Space>
      )}

      <Modal
        title="登记不良事件"
        open={modalOpen}
        onCancel={() => { setModalOpen(false); resetForm() }}
        onOk={handleRegister}
        confirmLoading={saving}
        okText="登记"
        width={520}
        destroyOnClose
      >
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <div>
            <div className="text-[13px] font-bold mb-1">事件分类</div>
            <Select
              showSearch
              style={{ width: '100%' }}
              placeholder="选择并发症类型"
              value={form.eventType || undefined}
              onChange={(v) => setForm({ ...form, eventType: v })}
              options={eventTypes}
            />
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">严重程度</div>
            <Radio.Group value={form.severity} onChange={(e) => setForm({ ...form, severity: e.target.value })}>
              <Radio.Button value="mild">轻</Radio.Button>
              <Radio.Button value="moderate">中</Radio.Button>
              <Radio.Button value="severe">重</Radio.Button>
            </Radio.Group>
            {form.severity === 'severe' && (
              <div className="mt-1 text-[12px] text-rose-500 font-bold">严重事件须 6 小时内上报护士长+科主任</div>
            )}
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">发生时间</div>
            <DatePicker
              showTime
              style={{ width: '100%' }}
              value={form.occurredAt}
              onChange={(v) => setForm({ ...form, occurredAt: v || dayjs() })}
            />
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">发生经过</div>
            <Input.TextArea rows={3} placeholder="描述事件经过…" value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">处理措施</div>
            <Input.TextArea rows={2} placeholder="处理措施…" value={form.handling}
              onChange={(e) => setForm({ ...form, handling: e.target.value })} />
          </div>
          <div>
            <div className="text-[13px] font-bold mb-1">转归结果</div>
            <Input.TextArea rows={2} placeholder="转归…" value={form.outcome}
              onChange={(e) => setForm({ ...form, outcome: e.target.value })} />
          </div>
        </Space>
      </Modal>

      <Modal
        title="上报不良事件"
        open={!!reportModalOpen}
        onCancel={() => setReportModalOpen(null)}
        onOk={() => reportModalOpen && handleReport(reportModalOpen)}
        confirmLoading={saving}
        okText="上报"
        destroyOnClose
      >
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          {reportRoles.map((r, i) => (
            <div key={i}>
              <div className="text-[13px] font-bold mb-1">{r.role} 用户ID</div>
              <Input
                placeholder={`${r.role}的用户ID`}
                value={r.userId}
                onChange={(e) => {
                  const copy = [...reportRoles]
                  copy[i] = { ...r, userId: e.target.value }
                  setReportRoles(copy)
                }}
              />
            </div>
          ))}
          <div className="text-[12px] text-slate-400">默认上报护士长、科主任，可修改；严重事件自动判定 6h 达标</div>
        </Space>
      </Modal>
    </div>
  )
}
