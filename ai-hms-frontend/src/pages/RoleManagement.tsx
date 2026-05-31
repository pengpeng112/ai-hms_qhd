import { useEffect, useState } from 'react'
import { message, Table, Button, Space, Modal } from 'antd'
import { Plus, Edit3, RefreshCw } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { roleManagementApi } from '@/services/roleManagementApi'

interface RoleItem {
  id: string
  name: string
  code: string
  description?: string
  isEnabled: boolean
}

export default function RoleManagement() {
  const [loading, setLoading] = useState(false)
  const [roles, setRoles] = useState<RoleItem[]>([])

  const loadRoles = async () => {
    setLoading(true)
    try {
      const res = await roleManagementApi.getRoleList()
      setRoles(res as RoleItem[] || [])
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadRoles() }, [])

  const handleCreate = () => {
    Modal.info({ title: '新建角色', content: '创建角色功能待实现' })
  }

  const columns = [
    { title: '角色名称', dataIndex: 'name', key: 'name' },
    { title: '角色编码', dataIndex: 'code', key: 'code' },
    { title: '描述', dataIndex: 'description', key: 'description' },
    { title: '状态', dataIndex: 'isEnabled', key: 'isEnabled', render: (v: boolean) => v ? '已启用' : '已禁用' },
    { title: '操作', key: 'action', render: (_: unknown, record: RoleItem) => (
      <Button type="link" size="small" onClick={() => Modal.info({ title: '编辑', content: `编辑 ${record.name}` })}>
        <Edit3 size={14} /> 编辑
      </Button>
    )},
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-h2 font-bold text-foreground">角色管理</h2>
        <Space>
          <Button onClick={loadRoles} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
          <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建角色</Button>
        </Space>
      </div>
      <Table dataSource={roles} columns={columns} rowKey="id" loading={loading} pagination={false} />
    </div>
  )
}
