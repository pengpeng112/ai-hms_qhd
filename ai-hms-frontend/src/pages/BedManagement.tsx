import { useCallback, useEffect, useState } from 'react'
import { message, Table, Button, Space, Modal, Form, Input, Switch, Popconfirm, InputNumber, Select } from 'antd'
import { Plus, Edit3, Trash2, RefreshCw } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { bedManagementApi, type BedItem, type BedPayload, wardManagementApi, type WardItem } from '@/services/managementApi'

export default function BedManagement() {
  const [loading, setLoading] = useState(false)
  const [beds, setBeds] = useState<BedItem[]>([])
  const [wards, setWards] = useState<WardItem[]>([])
  const [editVisible, setEditVisible] = useState(false)
  const [editingBed, setEditingBed] = useState<BedItem | null>(null)
  const [form] = Form.useForm()

  const loadBeds = useCallback(async () => {
    setLoading(true)
    try {
      const res = await bedManagementApi.list()
      setBeds(res)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [])

  const loadWards = useCallback(async () => {
    try {
      const res = await wardManagementApi.list()
      setWards(res)
    } catch { /* ignore */ }
  }, [])

  useEffect(() => { void loadBeds() }, [loadBeds])
  useEffect(() => { void loadWards() }, [loadWards])

  const handleCreate = () => {
    setEditingBed(null)
    form.resetFields()
    setEditVisible(true)
  }

  const handleEdit = (record: BedItem) => {
    setEditingBed(record)
    form.setFieldsValue({
      name: record.name,
      wardId: record.wardId ? Number(record.wardId) : undefined,
      sort: record.sort,
      fepId: record.fepId ?? undefined,
      acquisiteConnectId: record.acquisiteConnectId ?? undefined,
      note: record.note || '',
      isDisabled: record.isDisabled,
    })
    setEditVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const payload: BedPayload = {
        name: values.name,
        wardId: Number(values.wardId),
        sort: values.sort ?? 0,
        fepId: values.fepId ? Number(values.fepId) : undefined,
        acquisiteConnectId: values.acquisiteConnectId ? Number(values.acquisiteConnectId) : undefined,
        note: values.note || undefined,
        isDisabled: values.isDisabled ?? false,
      }
      if (editingBed) {
        await bedManagementApi.update(editingBed.id, payload)
        message.success('修改成功')
      } else {
        await bedManagementApi.create(payload)
        message.success('新增成功')
      }
      setEditVisible(false)
      void loadBeds()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await bedManagementApi.remove(id)
      message.success('删除成功')
      void loadBeds()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const wardOptions = wards.map(w => ({ label: w.name, value: String(w.id) }))

  const columns = [
    { title: '床位名称', dataIndex: 'name', key: 'name', width: 150 },
    { title: '所属病区', dataIndex: 'wardName', key: 'wardName', width: 150 },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
    { title: '默认设备', dataIndex: 'defaultEquipmentName', key: 'defaultEquipmentName', width: 130, render: (v: string) => v || '-' },
    { title: '设备数量', dataIndex: 'equipmentCount', key: 'equipmentCount', width: 80 },
    { title: '状态', dataIndex: 'isDisabled', key: 'isDisabled', width: 80, render: (v: boolean) => <Switch size="small" checked={!v} disabled /> },
    {
      title: '操作', key: 'action', width: 160,
      render: (_: unknown, record: BedItem) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => handleEdit(record)}><Edit3 size={14} /> 编辑</Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger><Trash2 size={14} /> 删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="bg-white rounded-lg p-6 shadow-sm">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h2 className="text-xl font-bold text-slate-800">床位管理</h2>
            <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
              <span>基础数据</span><span>/</span><span className="text-slate-600">床位管理</span>
            </div>
          </div>
          <Space>
            <Button onClick={loadBeds} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建床位</Button>
          </Space>
        </div>
        <Table dataSource={beds} columns={columns} rowKey="id" loading={loading} pagination={false} />
      </div>

      <Modal title={editingBed ? '编辑床位' : '新建床位'} open={editVisible} onCancel={() => setEditVisible(false)} onOk={handleSubmit} width={560} destroyOnClose>
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="name" label="床位名称" rules={[{ required: true, message: '请输入床位名称' }]}>
            <Input placeholder="请输入床位名称（如 床位01）" />
          </Form.Item>
          <Form.Item name="wardId" label="所属病区" rules={[{ required: true, message: '请选择所属病区' }]}>
            <Select placeholder="请选择病区" options={wardOptions} />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="sort" label="排序" className="flex-1">
              <InputNumber className="w-full" placeholder="序号" min={0} />
            </Form.Item>
          </Space>
          <Space className="w-full" size="middle">
            <Form.Item name="fepId" label="FEP 设备" className="flex-1">
              <InputNumber className="w-full" placeholder="FEP设备 ID" />
            </Form.Item>
            <Form.Item name="acquisiteConnectId" label="采集连接" className="flex-1">
              <InputNumber className="w-full" placeholder="采集连接 ID" />
            </Form.Item>
          </Space>
          <Form.Item name="note" label="备注">
            <Input.TextArea rows={3} placeholder="请输入备注" />
          </Form.Item>
          <Form.Item name="isDisabled" label="是否禁用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
