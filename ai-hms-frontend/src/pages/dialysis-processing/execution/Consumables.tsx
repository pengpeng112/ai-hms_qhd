import { useState, useEffect } from 'react'
import { Table, Button, InputNumber, Input, Popconfirm, message } from 'antd'
import { Plus, Trash2 } from 'lucide-react'
import type { Patient } from '../types'
import { listConsumables, createConsumable, deleteConsumable, type ConsumableRecord } from '@/services/consumableApi'

export default function Consumables({ patient, treatmentId }: { patient: Patient; treatmentId: number }) {
  const [items, setItems] = useState<ConsumableRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({ materialId: 0, num: 1, unit: '', batch: '', serialNo: '', note: '' })

  const loadItems = async () => {
    if (!treatmentId) return
    setLoading(true)
    try { setItems(await listConsumables(treatmentId)) } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  useEffect(() => { void loadItems() }, [treatmentId])

  const handleAdd = async () => {
    if (!treatmentId || !form.materialId) { message.warning('请填写材料ID'); return }
    try {
      await createConsumable(treatmentId, form)
      message.success('已添加')
      setForm({ materialId: 0, num: 1, unit: '', batch: '', serialNo: '', note: '' })
      void loadItems()
    } catch { message.error('添加失败') }
  }

  const handleDelete = async (id: number) => {
    try { await deleteConsumable(treatmentId, id); message.success('已删除'); void loadItems() }
    catch { message.error('删除失败') }
  }

  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="text-lg font-black text-slate-800 mb-4">耗材核对</div>
      <div className="text-sm text-slate-600 mb-4">患者：{patient.name}</div>

      {!treatmentId ? (
        <div className="rounded-xl bg-amber-50 border border-amber-200 p-4 text-sm text-amber-700">
          <p className="font-semibold">请先选择当前治疗记录</p>
          <p className="mt-1">耗材记录需要关联到当前治疗记录。请从透析执行主页面选择患者后，确保传入 treatmentId。</p>
        </div>
      ) : (
        <>
          <div className="flex flex-wrap gap-2 mb-4">
            <InputNumber placeholder="材料ID" value={form.materialId || undefined} onChange={v => setForm(p => ({ ...p, materialId: v || 0 }))} className="w-24" />
            <InputNumber placeholder="数量" value={form.num} onChange={v => setForm(p => ({ ...p, num: v || 0 }))} className="w-20" min={0} />
            <Input placeholder="单位" value={form.unit} onChange={e => setForm(p => ({ ...p, unit: e.target.value }))} className="w-20" />
            <Input placeholder="批号" value={form.batch} onChange={e => setForm(p => ({ ...p, batch: e.target.value }))} className="w-32" />
            <Input placeholder="序列号" value={form.serialNo} onChange={e => setForm(p => ({ ...p, serialNo: e.target.value }))} className="w-32" />
            <Input placeholder="备注" value={form.note} onChange={e => setForm(p => ({ ...p, note: e.target.value }))} className="w-40" />
            <Button type="primary" onClick={handleAdd} icon={<Plus size={14} />}>添加</Button>
          </div>
          <Table dataSource={items} rowKey="id" loading={loading} pagination={false} size="small"
            columns={[
              { title: '材料ID', dataIndex: 'materialId', width: 80 },
              { title: '数量', dataIndex: 'num', width: 60 },
              { title: '单位', dataIndex: 'unit', width: 60 },
              { title: '批号', dataIndex: 'batch', width: 100 },
              { title: '序列号', dataIndex: 'serialNo', width: 120 },
              { title: '备注', dataIndex: 'note' },
              { title: '操作', width: 60, render: (_: unknown, r: ConsumableRecord) => (
                <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                  <Button type="link" size="small" danger><Trash2 size={14} /></Button>
                </Popconfirm>
              )},
            ]}
          />
        </>
      )}
    </div>
  )
}
