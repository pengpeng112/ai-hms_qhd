import { useCallback, useEffect, useState } from 'react'
import { Upload, Table, InputNumber, Tag, Space, Button, message } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { getErrorMessage } from '@/services'
import { actrApi, type PatientACTR, type ActrStatus } from '@/services/actrApi'

interface Props {
  open: boolean
  onClose: () => void
  patientId: string | number
  prescriptionId?: string
}

export default function ActrPanel({ open, onClose, patientId, prescriptionId }: Props) {
  const [rows, setRows] = useState<PatientACTR[]>([])
  const [loading, setLoading] = useState(false)
  const [latest, setLatest] = useState<PatientACTR | null>(null)
  const [dryWeight, setDryWeight] = useState<number | null>(null)
  const [uf, setUf] = useState<number | null>(null)
  const [status, setStatus] = useState<ActrStatus & { loaded: boolean }>({ enabled: false, configured: false, reachable: false, loaded: false })

  const loadStatus = useCallback(async () => {
    try {
      const s = await actrApi.status()
      setStatus({ ...s, loaded: true })
    } catch {
      setStatus({ enabled: false, configured: false, reachable: false, loaded: true })
    }
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await actrApi.history(patientId)
      setRows(data)
      setLatest(data[0] ?? null)
    } catch {
      message.error('影像历史加载失败')
    } finally {
      setLoading(false)
    }
  }, [patientId])

  useEffect(() => {
    if (open) {
      loadStatus()
      load()
    }
  }, [open, loadStatus, load])

  const onUpload = async (file: File) => {
    setLoading(true)
    try {
      const rec = await actrApi.analyze(patientId, file)
      setLatest(rec)
      if (rec.qcPass === 0) message.warning('影像质控不合格，CTR 仅供参考，不计入质控')
      await load()
    } catch (e: unknown) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
    return false
  }

  const onAdopt = async () => {
    if (!prescriptionId || !latest) return
    try {
      await actrApi.adopt(patientId, {
        prescriptionId,
        actrRecordId: latest.id,
        dryWeight: dryWeight ?? undefined,
        ufQuantity: uf ?? undefined,
      })
      message.success('已带入当日处方草稿，请在处方页确认并签发')
      onClose()
    } catch (e: unknown) {
      message.error(getErrorMessage(e))
    }
  }

  const qcTag = (qcPass: number) =>
    qcPass === 1 ? <Tag color="green">QC 合格</Tag> : <Tag color="orange">QC 不合格·仅供参考</Tag>

  if (!open) return null

  return (
    <div className="fixed inset-0 z-[200]">
      <div className="absolute inset-0 bg-black/40 backdrop-blur-sm" onClick={onClose} />
      <div className="absolute right-0 top-0 h-full w-[520px] bg-white shadow-2xl flex flex-col"
        style={{ animation: 'slide-in-right 0.2s ease-out' }}>
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 shrink-0">
          <h3 className="text-lg font-bold text-slate-800">影像 / ACTR 辅助</h3>
          <button type="button" onClick={onClose}
            className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>

        <div className="flex-1 overflow-auto px-6 py-4 space-y-5">
          {status.loaded && !status.enabled && (
            <div className="bg-amber-50 border border-amber-200 rounded-lg px-4 py-3 text-sm text-amber-700">
              {!status.configured ? '影像服务未配置，请联系管理员' : '影像服务未启用，仅可查看历史'}
            </div>
          )}

          {status.loaded && status.enabled && !status.reachable && (
            <div className="bg-rose-50 border border-rose-200 rounded-lg px-4 py-3 text-sm text-rose-700">
              影像服务不可达，请联系管理员检查 ACTRS 服务状态
            </div>
          )}

          {status.loaded && status.enabled && status.reachable && (
            <>
              <Upload beforeUpload={(f) => { onUpload(f as File); return Upload.LIST_IGNORE }} showUploadList={false} accept="image/*,.dcm,.dicom,.ima">
                <Button icon={<UploadOutlined />} loading={loading}>上传胸片即时分析</Button>
              </Upload>

              {latest && (
                <div className="space-y-2">
                  <Space wrap>
                    <Tag color="blue">CTR {latest.ctr?.toFixed(3) ?? '-'}</Tag>
                    <Tag color="geekblue">ACTR {latest.actr?.toFixed(3) ?? '-'}</Tag>
                    {qcTag(latest.qcPass)}
                  </Space>
                  {latest.qcWarnings && <p className="text-xs text-slate-500">提示：{latest.qcWarnings}</p>}
                </div>
              )}
            </>
          )}

          {!status.enabled && rows.length > 0 && latest && (
            <div className="space-y-2">
              <Space wrap>
                <Tag color="blue">CTR {latest.ctr?.toFixed(3) ?? '-'}</Tag>
                <Tag color="geekblue">ACTR {latest.actr?.toFixed(3) ?? '-'}</Tag>
                {qcTag(latest.qcPass)}
              </Space>
            </div>
          )}

          <Table
            size="small"
            loading={loading}
            rowKey="id"
            dataSource={rows}
            pagination={{ pageSize: 5 }}
            columns={[
              { title: '日期', dataIndex: 'analysisDate', render: (v: string) => v?.slice(0, 10) ?? '-' },
              { title: 'CTR', dataIndex: 'ctr', render: (v: number) => v?.toFixed(3) ?? '-' },
              { title: 'ACTR', dataIndex: 'actr', render: (v: number) => v?.toFixed(3) ?? '-' },
              { title: 'QC', dataIndex: 'qcPass', render: (v: number) => (v === 1 ? '合格' : '不合格') },
            ]}
          />

          {status.enabled && prescriptionId && latest && (
            <div className="bg-slate-50 rounded-lg p-4 space-y-3">
              <p className="text-sm font-bold text-slate-700">采纳到当日处方草稿</p>
              <Space>
                <span className="text-sm text-slate-500">干体重</span>
                <InputNumber value={dryWeight ?? undefined} onChange={(v) => setDryWeight(v as number)} step={0.1} addonAfter="kg" size="small" />
                <span className="text-sm text-slate-500">超滤量</span>
                <InputNumber value={uf ?? undefined} onChange={(v) => setUf(v as number)} step={0.1} addonAfter="L" size="small" />
              </Space>
              <Button type="primary" onClick={onAdopt} block>一键带入当日处方</Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
