import { useCallback, useEffect, useState } from 'react'
import { message, Table, Button, Space, Modal, Form, Input, Select, Switch, Popconfirm, InputNumber } from 'antd'
import { Plus, Edit3, Trash2, RefreshCw } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { wardManagementApi, type WardItem, type WardPayload } from '@/services/managementApi'
import { userApi, type RestUser } from '@/services/userApi'

export default function WardManagement() {
  const [loading, setLoading] = useState(false)
  const [wards, setWards] = useState<WardItem[]>([])
  const [userOptions, setUserOptions] = useState<Array<{ label: string; value: string }>>([])
  const [editVisible, setEditVisible] = useState(false)
  const [editingWard, setEditingWard] = useState<WardItem | null>(null)
  const [form] = Form.useForm()

  const loadWards = useCallback(async () => {
    setLoading(true)
    try {
      const res = await wardManagementApi.list()
      setWards(res)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void loadWards() }, [loadWards])

  useEffect(() => {
    const loadUsers = async () => {
      try {
        const res = await userApi.getList({ page: 1, pageSize: 200 })
        setUserOptions((res.items || []).map((u: RestUser) => ({
          label: u.realName || u.username,
          value: String(u.id),
        })))
      } catch { /* ignore */ }
    }
    void loadUsers()
  }, [])

  const handleCreate = () => {
    setEditingWard(null)
    form.resetFields()
    setEditVisible(true)
  }

  const handleEdit = (record: WardItem) => {
    setEditingWard(record)
    form.setFieldsValue({
      name: record.name,
      sort: record.sort,
      patientType: record.patientType || '',
      infectionType: record.infectionType || '',
      responsibleUsers: record.responsibleUsers
        ? record.responsibleUsers.split(',').map((s) => s.trim()).filter(Boolean)
        : [],
      note: record.note || '',
      isDisabled: record.isDisabled,
    })
    setEditVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const payload: WardPayload = {
        name: values.name,
        sort: values.sort ?? 0,
        patientType: values.patientType || undefined,
        infectionType: values.infectionType || undefined,
        responsibleUsers: Array.isArray(values.responsibleUsers)
          ? values.responsibleUsers.join(',')
          : (values.responsibleUsers || undefined),
        note: values.note || undefined,
        isDisabled: values.isDisabled ?? false,
      }
      if (editingWard) {
        await wardManagementApi.update(editingWard.id, payload)
        message.success('修改成功')
      } else {
        await wardManagementApi.create(payload)
        message.success('新增成功')
      }
      setEditVisible(false)
      void loadWards()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await wardManagementApi.remove(id)
      message.success('删除成功')
      void loadWards()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const columns = [
    { title: '病区名称', dataIndex: 'name', key: 'name', width: 160 },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
    { title: '床位数量', dataIndex: 'bedCount', key: 'bedCount', width: 90 },
    { title: '患者类型', key: 'patientType', width: 120, render: (_: unknown, r: WardItem) => r.patientTypeLabel || r.patientType || '-' },
    { title: '是否为传染病区', key: 'infectionType', width: 110, render: (_: unknown, r: WardItem) => r.infectionType || '-' },
    { title: '状态', dataIndex: 'isDisabled', key: 'isDisabled', width: 80, render: (v: boolean) => <Switch size="small" checked={!v} disabled /> },
    {
      title: '操作', key: 'action', width: 160,
      render: (_: unknown, record: WardItem) => (
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
            <h2 className="text-xl font-bold text-slate-800">病区管理</h2>
            <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
              <span>基础数据</span><span>/</span><span className="text-slate-600">病区管理</span>
            </div>
          </div>
          <Space>
            <Button onClick={loadWards} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建病区</Button>
          </Space>
        </div>
        <Table dataSource={wards} columns={columns} rowKey="id" loading={loading} pagination={false} />
      </div>

      <Modal title={editingWard ? '编辑病区' : '新建病区'} open={editVisible} onCancel={() => setEditVisible(false)} onOk={handleSubmit} width={520} destroyOnClose>
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="name" label="病区名称" rules={[{ required: true, message: '请输入病区名称' }]}>
            <Input placeholder="请输入病区名称" />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="sort" label="排序" className="flex-1">
              <InputNumber className="w-full" placeholder="序号" min={0} />
            </Form.Item>
            <Form.Item name="patientType" label="患者类型" className="flex-1">
              <Select placeholder="请选择" allowClear options={[
                { label: '普通患者', value: '普通患者' },
                { label: '隔离患者', value: '隔离患者' },
              ]} />
            </Form.Item>
          </Space>
          <Space className="w-full" size="middle">
            <Form.Item name="infectionType" label="感染类型" className="flex-1">
              <Select placeholder="请选择" allowClear options={[
                { label: '乙肝', value: '乙肝' },
                { label: '丙肝', value: '丙肝' },
                { label: '梅毒', value: '梅毒' },
                { label: 'HIV', value: 'HIV' },
              ]} />
            </Form.Item>
            <Form.Item name="responsibleUsers" label="负责医护" className="flex-1">
              <Select mode="multiple" placeholder="请选择负责医护" options={userOptions} allowClear />
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
