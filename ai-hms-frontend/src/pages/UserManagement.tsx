import { useEffect, useState } from 'react'
import { message, Table, Button, Space, Input } from 'antd'
import { RefreshCw, Search } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { userApi, type RestUser } from '@/services/userApi'

export default function UserManagement() {
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<RestUser[]>([])
  const [keyword, setKeyword] = useState('')

  const loadUsers = async () => {
    setLoading(true)
    try {
      const res = await userApi.getList({ keyword: keyword || undefined })
      setUsers(res.items || [])
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadUsers() }, [])

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: '姓名', dataIndex: 'realName', key: 'realName' },
    { title: '角色', dataIndex: 'role', key: 'role' },
    { title: '状态', dataIndex: 'status', key: 'status' },
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-h2 font-bold text-foreground">用户管理</h2>
        <Space>
          <Input placeholder="搜索姓名/用户名" value={keyword} onChange={e => setKeyword(e.target.value)}
            onPressEnter={() => loadUsers()} prefix={<Search size={16} className="text-gray-400" />}
            className="w-48" />
          <Button onClick={loadUsers} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
        </Space>
      </div>
      <Table dataSource={users} columns={columns} rowKey="id" loading={loading} pagination={false} />
    </div>
  )
}
