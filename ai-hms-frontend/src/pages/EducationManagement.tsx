import { useCallback, useEffect, useMemo, useState } from 'react'
import { message, Table, Button, Space, Modal, Form, Input, Switch, Popconfirm } from 'antd'
import { Plus, Edit3, Trash2, RefreshCw } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { educationManagementApi, type EducationItem, type EducationPayload } from '@/services/managementApi'

export default function EducationManagement() {
  const [loading, setLoading] = useState(false)
  const [items, setItems] = useState<EducationItem[]>([])
  const [editVisible, setEditVisible] = useState(false)
  const [editingItem, setEditingItem] = useState<EducationItem | null>(null)
  const [form] = Form.useForm()

  const loadItems = useCallback(async () => {
    setLoading(true)
    try {
      const res = await educationManagementApi.list()
      setItems(res)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void loadItems() }, [loadItems])

  const handleCreate = () => {
    setEditingItem(null)
    form.resetFields()
    setEditVisible(true)
  }

  const handleEdit = (record: EducationItem) => {
    setEditingItem(record)
    form.setFieldsValue({
      name: record.name,
      description: record.description || '',
      sort: record.sort ?? 0,
      type: record.type || '',
      classify: record.classify || '',
      note: record.note || '',
      isDisabled: record.isDisabled,
    })
    setEditVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const payload: EducationPayload = {
        name: values.name,
        description: values.description || undefined,
        sort: values.sort ?? 0,
        type: values.type || undefined,
        classify: values.classify || undefined,
        note: values.note || undefined,
        isDisabled: values.isDisabled ?? false,
      }
      if (editingItem) {
        await educationManagementApi.update(editingItem.id, payload)
        message.success('修改成功')
      } else {
        await educationManagementApi.create(payload)
        message.success('新增成功')
      }
      setEditVisible(false)
      void loadItems()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await educationManagementApi.remove(id)
      message.success('删除成功')
      void loadItems()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const columns = useMemo(() => [
    { title: '宣教名称', dataIndex: 'name', key: 'name', width: 200 },
    { title: '类型', dataIndex: 'type', key: 'type', width: 100, render: (v: string) => v || '-' },
    { title: '分类', dataIndex: 'classify', key: 'classify', width: 100, render: (v: string) => v || '-' },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
    { title: '描述', dataIndex: 'description', key: 'description', render: (v: string) => v || '-' },
    {
      title: '状态',
      dataIndex: 'isDisabled',
      key: 'isDisabled',
      width: 80,
      render: (v: boolean) => v ? '已禁用' : '已启用',
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_: unknown, record: EducationItem) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => handleEdit(record)}><Edit3 size={14} /> 编辑</Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger><Trash2 size={14} /> 删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  // eslint-disable-next-line react-hooks/exhaustive-deps
  ], [])

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="bg-white rounded-lg p-6 shadow-sm">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h2 className="text-xl font-bold text-slate-800">宣教管理</h2>
            <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
              <span>基础数据</span><span>/</span><span className="text-slate-600">宣教管理</span>
            </div>
          </div>
          <Space>
            <Button onClick={loadItems} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建宣教</Button>
          </Space>
        </div>
        <Table dataSource={items} columns={columns} rowKey="id" loading={loading} pagination={false} />
      </div>

      <Modal title={editingItem ? '编辑宣教' : '新建宣教'} open={editVisible} onCancel={() => setEditVisible(false)} onOk={handleSubmit} width={520} destroyOnClose>
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="name" label="宣教名称" rules={[{ required: true, message: '请输入宣教名称' }]}>
            <Input placeholder="请输入宣教名称" />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="type" label="类型" className="flex-1">
              <Input placeholder="如：透析知识、用药指导" />
            </Form.Item>
            <Form.Item name="classify" label="分类" className="flex-1">
              <Input placeholder="如：基础宣教、专项宣教" />
            </Form.Item>
          </Space>
          <Space className="w-full" size="middle">
            <Form.Item name="sort" label="排序" className="flex-1">
              <Input type="number" placeholder="序号" />
            </Form.Item>
          </Space>
          <Form.Item name="description" label="描述内容">
            <Input.TextArea rows={4} placeholder="请输入宣教描述内容" />
          </Form.Item>
          <Form.Item name="note" label="备注">
            <Input.TextArea rows={2} placeholder="请输入备注" />
          </Form.Item>
          <Form.Item name="isDisabled" label="是否禁用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
