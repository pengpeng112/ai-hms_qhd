import { useState, useEffect, useCallback } from 'react'
import { Button, Table, Tag, message, Modal, Input, Space, Card, Empty } from 'antd'
import { Plus, CheckCircle, FileText, XCircle, RefreshCw } from 'lucide-react'
import type { TabProps } from '../types'
import type { ChargeRecord, ChargeLine } from '@services/billingApi'
import {
  getCharge, listCharges, addChargeLine,
  deleteChargeLine, confirmCharge, checkCharge, markExported, cancelCharge,
} from '@services/billingApi'
import { exportChargeToExcel } from '@/lib/billingExcel'

const STATUS_META: Record<string, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' },
  confirmed: { label: '已确认', color: 'blue' },
  checked: { label: '已核对', color: 'green' },
  cancelled: { label: '已取消', color: 'red' },
}

const CATEGORY_LABELS: Record<string, string> = {
  treatment: 'A 治疗费', material: 'B 耗材费', nursing: 'C 护理费',
  injection: 'D 注射费', drug: 'E 药品',
}

export default function BillingTab({ patient }: TabProps) {
  const [active, setActive] = useState<ChargeRecord | null>(null)
  const [loading, setLoading] = useState(false)
  const [busy, setBusy] = useState(false)
  const [cancelModalOpen, setCancelModalOpen] = useState(false)
  const [cancelReason, setCancelReason] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await listCharges({ patientId: patient.id ? Number(patient.id) : undefined, pageSize: 50 })
      if (res.items?.length > 0) {
        const full = await getCharge(res.items[0].id)
        setActive(full)
      }
    } catch {
      message.error('加载收费清单失败')
    } finally {
      setLoading(false)
    }
  }, [patient.id])

  useEffect(() => { load() }, [load])

  const reloadActive = useCallback(async (id: string) => {
    try { setActive(await getCharge(id)) } catch { message.error('刷新清单失败') }
  }, [])

  const handleConfirm = async () => {
    if (!active) return
    setBusy(true)
    try { const r = await confirmCharge(active.id); message.success('清单已确认'); reloadActive(r.id); load() }
    catch { message.error('确认失败') } finally { setBusy(false) }
  }

  const handleCheck = async () => {
    if (!active) return
    setBusy(true)
    try { const r = await checkCharge(active.id); message.success('双人核对完成'); reloadActive(r.id); load() }
    catch { message.error('核对失败') } finally { setBusy(false) }
  }

  const handleExport = async () => {
    if (!active) return
    try { exportChargeToExcel(active); await markExported(active.id); message.success('已导出 Excel'); reloadActive(active.id) }
    catch { message.error('导出失败') }
  }

  const handleCancel = async () => {
    if (!active) return
    setBusy(true)
    try { const r = await cancelCharge(active.id, cancelReason || '用户取消'); message.success('清单已取消'); reloadActive(r.id); load(); setCancelModalOpen(false); setCancelReason('') }
    catch { message.error('取消失败') } finally { setBusy(false) }
  }

  const handleAddManual = async () => {
    if (!active) return
    setBusy(true)
    try { await addChargeLine(active.id, { itemName: '手工添加项目', category: 'material', billable: true, quantity: 1, source: 'manual' }); message.success('已添加'); reloadActive(active.id) }
    catch { message.error('添加失败') } finally { setBusy(false) }
  }

  const handleDeleteLine = async (line: ChargeLine) => {
    Modal.confirm({
      title: '确认删除此行？', onOk: async () => {
        try { await deleteChargeLine(line.id); message.success('已删除'); if (active) reloadActive(active.id) }
        catch { message.error('删除失败') }
      },
    })
  }

  const columns = [
    { title: '类别', dataIndex: 'category', width: 90, render: (v: string) => CATEGORY_LABELS[v] ?? v },
    { title: '项目名称', dataIndex: 'itemName', width: 200 },
    { title: '规格', dataIndex: 'spec', width: 100, render: (v: string) => v ?? '-' },
    { title: '数量', dataIndex: 'quantity', width: 60 },
    { title: '参考单价', dataIndex: 'unitPrice', width: 80, render: (v: number) => v != null ? v.toFixed(2) : '-' },
    { title: '参考金额', dataIndex: 'amount', width: 80, render: (v: number, r: ChargeLine) => r.billable && v != null ? v.toFixed(2) : '-' },
    { title: '可计费', dataIndex: 'billable', width: 60, render: (v: boolean) => v ? <Tag color="green">是</Tag> : <Tag color="default">否</Tag> },
    {
      title: '操作', width: 60, render: (_: unknown, r: ChargeLine) =>
        active && active.status !== 'checked' && active.status !== 'cancelled'
          ? <Button size="small" danger type="link" onClick={() => handleDeleteLine(r)}>删除</Button>
          : null,
    },
  ]

  const editable = active && active.status !== 'checked' && active.status !== 'cancelled'

  return (
    <div style={{ padding: 16 }}>
      <div style={{ marginBottom: 16, display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
        {active && active.status === 'draft' && (
          <Button type="primary" icon={<CheckCircle size={14} />} onClick={handleConfirm} loading={busy}>确认</Button>
        )}
        {active && active.status === 'confirmed' && (
          <Button type="primary" icon={<CheckCircle size={14} />} onClick={handleCheck} loading={busy}>双人核对</Button>
        )}
        {active && (active.status === 'confirmed' || active.status === 'checked') && (
          <Button icon={<FileText size={14} />} onClick={handleExport}>导出 Excel</Button>
        )}
        {active && active.status !== 'cancelled' && (
          <Button danger icon={<XCircle size={14} />} onClick={() => setCancelModalOpen(true)}>取消</Button>
        )}
        {editable && <Button icon={<Plus size={14} />} onClick={handleAddManual}>添加明细</Button>}
        <Button icon={<RefreshCw size={14} />} onClick={load}>刷新</Button>
      </div>

      {active && (
        <Card size="small" style={{ marginBottom: 12 }}>
          <Space>
            <Tag color={STATUS_META[active.status]?.color as 'blue' | 'green' | 'red' | 'default'}>{STATUS_META[active.status]?.label ?? active.status}</Tag>
            <span>治疗ID: {active.treatmentId}</span>
            {active.recordedName && <span>记账人: {active.recordedName}</span>}
            {active.checkedName && <span>核对人: {active.checkedName}</span>}
            {active.totalAmount != null && <span>参考总价: ¥{active.totalAmount.toFixed(2)}</span>}
          </Space>
        </Card>
      )}

      {active ? (
        <Table rowKey="id" columns={columns} dataSource={active.lines ?? []} loading={loading} size="small" pagination={false}
          locale={{ emptyText: <Empty description="暂无明细，点击「添加明细」手工添加" /> }} />
      ) : (
        <Empty description="暂无收费清单，请在治疗执行页发起归集" />
      )}

      <Modal open={cancelModalOpen} title="取消清单" onOk={handleCancel} onCancel={() => setCancelModalOpen(false)} confirmLoading={busy}>
        <Input placeholder="取消原因" value={cancelReason} onChange={e => setCancelReason(e.target.value)} />
      </Modal>
    </div>
  )
}
