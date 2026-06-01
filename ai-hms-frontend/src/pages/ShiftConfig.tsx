import { useCallback, useEffect, useState } from 'react'
import { message, Table, Button, Space, Modal, Form, Input, Switch, Popconfirm, InputNumber, TimePicker } from 'antd'
import { Plus, Edit3, Trash2, RefreshCw } from 'lucide-react'
import dayjs from 'dayjs'
import { restApi, getErrorMessage, type RestShift } from '@/services'

export default function ShiftConfig() {
  const [loading, setLoading] = useState(false)
  const [shifts, setShifts] = useState<RestShift[]>([])
  const [editVisible, setEditVisible] = useState(false)
  const [editingShift, setEditingShift] = useState<RestShift | null>(null)
  const [form] = Form.useForm()

  const loadShifts = useCallback(async () => {
    setLoading(true)
    try {
      const res = await restApi.getShifts()
      setShifts(res.data)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void loadShifts() }, [loadShifts])

  const handleCreate = () => {
    setEditingShift(null)
    form.resetFields()
    form.setFieldsValue({ sort: 0, isDisabled: false })
    setEditVisible(true)
  }

  const handleEdit = (record: RestShift) => {
    setEditingShift(record)
    form.setFieldsValue({
      name: record.name,
      startTime: record.startTime ? dayjs(record.startTime, 'HH:mm') : null,
      endTime: record.endTime ? dayjs(record.endTime, 'HH:mm') : null,
      sort: record.sort ?? 0,
      notes: record.notes || '',
      isDisabled: record.isDisabled,
    })
    setEditVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const payload: Record<string, unknown> = {
        name: values.name,
        startTime: values.startTime ? values.startTime.format('HH:mm') : '',
        endTime: values.endTime ? values.endTime.format('HH:mm') : '',
        sort: values.sort ?? 0,
        notes: values.notes || '',
        isDisabled: values.isDisabled ?? false,
      }
      if (editingShift) {
        await restApi.updateShift(editingShift.id, payload)
        message.success('修改成功')
      } else {
        await restApi.createShift({
          name: values.name,
          startTime: values.startTime.format('HH:mm'),
          endTime: values.endTime.format('HH:mm'),
          sort: values.sort ?? 0,
          notes: values.notes || '',
        })
        message.success('新增成功')
      }
      setEditVisible(false)
      void loadShifts()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleToggle = async (record: RestShift) => {
    try {
      await restApi.updateShift(record.id, { isDisabled: !record.isDisabled })
      message.success(record.isDisabled ? '已启用' : '已禁用')
      void loadShifts()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await restApi.deleteShift(id)
      message.success('删除成功')
      void loadShifts()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const columns = [
    { title: '班次名称', dataIndex: 'name', key: 'name', width: 120 },
    { title: '开始时间', dataIndex: 'startTime', key: 'startTime', width: 100 },
    { title: '结束时间', dataIndex: 'endTime', key: 'endTime', width: 100 },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 70 },
    {
      title: '状态', dataIndex: 'isDisabled', key: 'isDisabled', width: 80,
      render: (_: unknown, record: RestShift) => (
        <Switch size="small" checked={!record.isDisabled} onChange={() => handleToggle(record)} />
      ),
    },
    {
      title: '操作', key: 'action', width: 160,
      render: (_: unknown, record: RestShift) => (
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
    <div className="max-w-[900px] mx-auto">
      <div className="bg-white rounded-lg p-6 shadow-sm">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h2 className="text-xl font-bold text-slate-800">班次配置</h2>
            <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
              <span>资源管理</span><span>/</span><span className="text-slate-600">班次配置</span>
            </div>
          </div>
          <Space>
            <Button onClick={loadShifts} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建班次</Button>
          </Space>
        </div>
        <Table dataSource={shifts} columns={columns} rowKey="id" loading={loading} pagination={false} />
      </div>

      <Modal title={editingShift ? '编辑班次' : '新建班次'} open={editVisible} onCancel={() => setEditVisible(false)} onOk={handleSubmit} width={480} destroyOnClose>
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="name" label="班次名称" rules={[{ required: true, message: '请输入班次名称' }]}>
            <Input placeholder="如：上午、下午、夜班" />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="startTime" label="开始时间" rules={[{ required: true, message: '请选择开始时间' }]}>
              <TimePicker format="HH:mm" className="w-full" />
            </Form.Item>
            <Form.Item name="endTime" label="结束时间" rules={[{ required: true, message: '请选择结束时间' }]}>
              <TimePicker format="HH:mm" className="w-full" />
            </Form.Item>
          </Space>
          <Form.Item name="sort" label="排序">
            <InputNumber className="w-full" placeholder="序号" min={0} />
          </Form.Item>
          <Form.Item name="notes" label="备注">
            <Input.TextArea rows={2} placeholder="请输入备注" />
          </Form.Item>
          {editingShift && (
            <Form.Item name="isDisabled" label="是否禁用" valuePropName="checked">
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
