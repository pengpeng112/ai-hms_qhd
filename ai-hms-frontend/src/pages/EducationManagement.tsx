import { useEffect, useState } from 'react'
import { message, Table, Button } from 'antd'
import { RefreshCw } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'

export default function EducationManagement() {
  const [loading, setLoading] = useState(false)
  const [items, setItems] = useState<unknown[]>([])

  const loadItems = async () => {
    setLoading(true)
    try {
      const token = localStorage.getItem('auth_token') || ''
      const res = await fetch('/api/v1/health-educations', {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) { setItems([]); return }
      const data = await res.json()
      setItems(data.items || data.data || [])
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadItems() }, [])

  const columns = [
    { title: '标题', dataIndex: 'title', key: 'title' },
    { title: '类型', dataIndex: 'type', key: 'type' },
    { title: '状态', dataIndex: 'status', key: 'status' },
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-h2 font-bold text-foreground">宣教管理</h2>
        <Button onClick={loadItems} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
      </div>
      <Table dataSource={items} columns={columns} loading={loading} pagination={false} />
    </div>
  )
}
