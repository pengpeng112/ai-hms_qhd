import { useState } from 'react'
import { Table, Button, InputNumber, Switch, Select, Popconfirm, Space, Card, Spin } from 'antd'
import { useThresholds } from './hooks/useThresholds'
import type { FixedThreshold, VPStratum, ThresholdPayload } from '@/services/monitoringThresholdApi'

type VPRow = VPStratum & { _rk: number }
type DraftState = Omit<ThresholdPayload, 'vpReference'> & { vpReference: VPRow[] }

let vpRowSeq = 0
const nextRk = () => ++vpRowSeq

const ACCESS_OPTIONS = ['AVF', 'AVG', 'TCC', 'NCC'].map((a) => ({ label: a, value: a }))

export default function MonitoringThresholds() {
  const { data, loading, saving, save, reset } = useThresholds()
  const [draft, setDraft] = useState<DraftState | null>(null)
  const [prevData, setPrevData] = useState<typeof data>(null)

  if (data !== prevData) {
    setPrevData(data)
    if (data) {
      setDraft({
        fixed: structuredClone(data.fixed),
        naFactor: data.naFactor,
        vpReference: data.vpReference.map((row) => ({ ...row, _rk: nextRk() })),
      })
    }
  }

  if (loading || !draft) return <div className="p-8"><Spin /> 正在加载阈值表...</div>

  const patchFixed = (i: number, patch: Partial<FixedThreshold>) => {
    const fixed = draft.fixed.map((r, idx) => (idx === i ? { ...r, ...patch } : r))
    setDraft({ ...draft, fixed })
  }
  const patchVP = (i: number, patch: Partial<VPRow>) => {
    const vpReference = draft.vpReference.map((r, idx) => (idx === i ? { ...r, ...patch } : r))
    setDraft({ ...draft, vpReference })
  }
  const addVP = () =>
    setDraft({
      ...draft,
      vpReference: [...draft.vpReference, { access: 'AVF', bfMin: 0, bfMax: 150, normalLow: 0, warnHigh: 0, dangerHigh: 0, basis: '', enabled: true, _rk: nextRk() }],
    })
  const removeVP = (i: number) => setDraft({ ...draft, vpReference: draft.vpReference.filter((_, idx) => idx !== i) })

  const numCell = (val: number | null, on: (v: number | null) => void) => (
    <InputNumber value={val ?? undefined} onChange={(v) => on(v ?? null)} style={{ width: 90 }} />
  )

  const fixedColumns = [
    { title: '指标', dataIndex: 'label', render: (_v: string, r: FixedThreshold) => r.label || r.metricKey },
    { title: '危险低', render: (_v: unknown, r: FixedThreshold, i: number) => numCell(r.dangerLow, (v) => patchFixed(i, { dangerLow: v })) },
    { title: '警戒低', render: (_v: unknown, r: FixedThreshold, i: number) => numCell(r.warnLow, (v) => patchFixed(i, { warnLow: v })) },
    { title: '警戒高', render: (_v: unknown, r: FixedThreshold, i: number) => numCell(r.warnHigh, (v) => patchFixed(i, { warnHigh: v })) },
    { title: '危险高', render: (_v: unknown, r: FixedThreshold, i: number) => numCell(r.dangerHigh, (v) => patchFixed(i, { dangerHigh: v })) },
    { title: '单位', dataIndex: 'unit' },
    { title: '依据', dataIndex: 'basis', render: (_v: string, r: FixedThreshold) => <span className="text-xs text-gray-500">{r.basis}</span> },
    { title: '启用', render: (_v: unknown, r: FixedThreshold, i: number) => <Switch checked={r.enabled} onChange={(v) => patchFixed(i, { enabled: v })} /> },
  ]

  const vpColumns = [
    { title: '通路', render: (_v: unknown, r: VPStratum, i: number) => <Select value={r.access} options={ACCESS_OPTIONS} style={{ width: 80 }} onChange={(v: string) => patchVP(i, { access: v as VPStratum['access'] })} /> },
    { title: 'BF下限', render: (_v: unknown, r: VPStratum, i: number) => <InputNumber value={r.bfMin} onChange={(v) => patchVP(i, { bfMin: v ?? 0 })} style={{ width: 90 }} /> },
    { title: 'BF上限', render: (_v: unknown, r: VPStratum, i: number) => <InputNumber value={r.bfMax} onChange={(v) => patchVP(i, { bfMax: v ?? 0 })} style={{ width: 90 }} /> },
    { title: '正常下限', render: (_v: unknown, r: VPStratum, i: number) => <InputNumber value={r.normalLow} onChange={(v) => patchVP(i, { normalLow: v ?? 0 })} style={{ width: 90 }} /> },
    { title: '警戒高起', render: (_v: unknown, r: VPStratum, i: number) => <InputNumber value={r.warnHigh} onChange={(v) => patchVP(i, { warnHigh: v ?? 0 })} style={{ width: 90 }} /> },
    { title: '危险高起', render: (_v: unknown, r: VPStratum, i: number) => <InputNumber value={r.dangerHigh} onChange={(v) => patchVP(i, { dangerHigh: v ?? 0 })} style={{ width: 90 }} /> },
    { title: '启用', render: (_v: unknown, r: VPStratum, i: number) => <Switch checked={r.enabled} onChange={(v) => patchVP(i, { enabled: v })} /> },
    { title: '操作', render: (_v: unknown, _r: VPStratum, i: number) => <Popconfirm title="删除该行?" onConfirm={() => removeVP(i)}><Button danger size="small">删除</Button></Popconfirm> },
  ]

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">实时监控报警阈值表</h1>
        <Space>
          <Popconfirm title="用系统默认覆盖当前阈值?" onConfirm={() => void reset()}>
            <Button danger loading={saving}>恢复默认</Button>
          </Popconfirm>
          <Button type="primary" loading={saving} onClick={() => {
            const payload: ThresholdPayload = {
              fixed: draft.fixed,
              naFactor: draft.naFactor,
              vpReference: draft.vpReference.map((r) => {
                const { _rk, ...rest } = r
                void _rk
                return rest
              }),
            }
            void save(payload)
          }}>保存</Button>
        </Space>
      </div>

      <Card title="固定阈值(五档)" size="small">
        <Table rowKey="metricKey" dataSource={draft.fixed} columns={fixedColumns} pagination={false} size="small" />
      </Card>

      <Card title="静脉压 VP 分层(通路 x 血流量 BF)" size="small"
            extra={<Button size="small" onClick={addVP}>新增行</Button>}>
        <Table rowKey="_rk" dataSource={draft.vpReference} columns={vpColumns} pagination={false} size="small" />
      </Card>

      <Card title="透析液钠电导率系数" size="small">
        <Space>
          <span>系数(Na = 电导率 x 系数, 默认 9.9):</span>
          <InputNumber value={draft.naFactor} step={0.1} min={0.1} onChange={(v) => setDraft({ ...draft, naFactor: v ?? 9.9 })} />
        </Space>
      </Card>
    </div>
  )
}
