import { useEffect, useState } from 'react'
import { Table, Button, message } from 'antd'
import { Plus, RefreshCw } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { getErrorMessage } from '@/services/restClient'
import type { ScheduleTemplate } from '@/services/scheduleTemplate'

export default function ScheduleTemplateList() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [templates, setTemplates] = useState<ScheduleTemplate[]>([])

  const loadTemplates = async () => {
    setLoading(true)
    try {
      const token = localStorage.getItem('auth_token') || ''
      const res = await fetch('/api/v1/schedule/template', {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error('加载模板失败')
      const data = await res.json()
      setTemplates(data.data || data || [])
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadTemplates() }, [])

  const columns = [
    { title: '模板名称', dataIndex: 'name', key: 'name' },
    { title: '周期(周)', dataIndex: 'cycleWeeks', key: 'cycleWeeks' },
    { title: '默认', dataIndex: 'isDefault', key: 'isDefault', render: (v: boolean) => v ? '是' : '否' },
    { title: '状态', dataIndex: 'isEnabled', key: 'isEnabled', render: (v: boolean) => v ? '启用' : '禁用' },
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-h2 font-bold text-foreground">排班模板</h2>
        <div className="flex gap-2">
          <Button onClick={() => navigate('/schedule-templates/edit')} icon={<Plus size={16} />}>编辑模板</Button>
          <Button onClick={loadTemplates} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
        </div>
      </div>
      <Table dataSource={templates} columns={columns} rowKey="id" loading={loading} pagination={false} />
    </div>
  )
}
