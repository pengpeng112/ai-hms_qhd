import { useEffect, useState } from 'react'
import { Table, Button, message, Tag } from 'antd'
import { RefreshCw, Edit3, Plus } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { restApi, getErrorMessage } from '@/services/restClient'
import type { ScheduleTemplateResponse } from '@/services/restClient'

const SCOPE_LABELS: Record<string, string> = { ALL: '全局', A: 'A区', B: 'B区', C: 'C区' }

export default function ScheduleTemplateList() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [templates, setTemplates] = useState<ScheduleTemplateResponse[]>([])

  const loadTemplates = async () => {
    setLoading(true)
    try {
      const data = await restApi.listScheduleTemplates()
      setTemplates(data)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadTemplates() }, [])

  const columns = [
    { title: '模板名称', dataIndex: ['template', 'name'], key: 'name' },
    { title: '范围', key: 'scope', render: (_: unknown, r: ScheduleTemplateResponse) => SCOPE_LABELS[r.template.scope] || r.template.scope || '--' },
    { title: '病区', key: 'wardId', render: (_: unknown, r: ScheduleTemplateResponse) => r.template.wardId ?? (r.template.scope === 'ALL' ? '全局' : '--') },
    { title: '项目数', dataIndex: 'itemCount', key: 'itemCount' },
    { title: '版本', dataIndex: ['template', 'version'], key: 'version' },
    { title: '状态', key: 'isActive', render: (_: unknown, r: ScheduleTemplateResponse) => <Tag color={r.template.isActive ? 'green' : 'default'}>{r.template.isActive ? '启用' : '禁用'}</Tag> },
    { title: '操作', key: 'action', render: (_: unknown, r: ScheduleTemplateResponse) => <Button size="small" icon={<Edit3 size={14} />} onClick={() => navigate(`/schedule-templates/edit?id=${r.template.id}`)}>编辑</Button> },
  ]

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-h2 font-bold text-foreground">排班模板</h2>
        <div className="flex gap-2">
          <Button onClick={() => navigate('/schedule-templates/edit')} icon={<Plus size={16} />} type="primary">新建模板</Button>
          <Button onClick={loadTemplates} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
        </div>
      </div>
      <Table dataSource={templates} columns={columns} rowKey={(r) => r.template.id} loading={loading} pagination={false} />
    </div>
  )
}
