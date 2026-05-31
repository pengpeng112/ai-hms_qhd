import { useEffect, useState } from 'react'
import { message, Modal, Table, Button, Space, Popconfirm } from 'antd'
import { Plus, Edit3, Trash2, RefreshCw } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { bedManagementApi, type BedItem } from '@/services/managementApi'

export default function BedManagement() {
  const [loading, setLoading] = useState(false)
  const [beds, setBeds] = useState<BedItem[]>([])

  const loadBeds = async () => {
    setLoading(true)
    try {
      const res = await bedManagementApi.list()
      setBeds(res)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadBeds() }, [])

  const handleCreate = () => {
    Modal.info({ title: '新建床位', content: '创建床位功能待实现' })
  }

  const columns = [
    { title: '床位名称', dataIndex: 'name', key: 'name' },
    { title: '所属病区', dataIndex: 'wardName', key: 'wardName' },
    { title: '默认设备', dataIndex: 'defaultEquipmentName', key: 'defaultEquipmentName' },
    { title: '设备数量', dataIndex: 'equipmentCount', key: 'equipmentCount' },
    { title: '状态', dataIndex: 'isDisabled', key: 'isDisabled', render: (v: boolean) => v ? '已禁用' : '正常' },
    { title: '操作', key: 'action', render: (_: unknown, record: BedItem) => (
      <Space>
        <Button type="link" size="small" onClick={() => Modal.info({ title: '编辑', content: `编辑 ${record.name}` })}>
          <Edit3 size={14} /> 编辑
        </Button>
        <Popconfirm title="确定删除？" onConfirm={() => message.info('删除功能待实现')}>
          <Button type="link" size="small" danger><Trash2 size={14} /> 删除</Button>
        </Popconfirm>
      </Space>
    )},
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-h2 font-bold text-foreground">床位管理</h2>
        <Space>
          <Button onClick={loadBeds} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
          <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建床位</Button>
        </Space>
      </div>
      <Table dataSource={beds} columns={columns} rowKey="id" loading={loading} pagination={false} />
    </div>
  )
}
