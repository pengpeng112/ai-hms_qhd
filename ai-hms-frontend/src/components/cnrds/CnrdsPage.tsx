import { useEffect, useState, useCallback } from 'react'
import {
  Table, Button, Tag, Modal, Select, InputNumber,
  DatePicker, Space, message, Descriptions,
} from 'antd'
import { FileSpreadsheet, Download, Send, Plus, Eye } from 'lucide-react'
import type { ColumnsType } from 'antd/es/table'
import { cnrdsApi, type CnrdsReport, type CnrdsContentRow } from '@/services/cnrdsApi'

const statusColors: Record<string, string> = {
  draft: 'default',
  exported: 'blue',
  submitted: 'green',
}

const statusLabels: Record<string, string> = {
  draft: '草稿',
  exported: '已导出',
  submitted: '已提交',
}

const eventTypeOptions = [
  { label: '死亡', value: 'death' },
  { label: '肾移植', value: 'transplant' },
  { label: '转出', value: 'transfer_out' },
]

export default function CnrdsPage() {
  const [reports, setReports] = useState<CnrdsReport[]>([])
  const [loading, setLoading] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [period, setPeriod] = useState<string>('')
  const [eventVisible, setEventVisible] = useState(false)
  const [eventPatientId, setEventPatientId] = useState<number | null>(null)
  const [eventType, setEventType] = useState<string>('death')
  const [previewVisible, setPreviewVisible] = useState(false)
  const [previewReport, setPreviewReport] = useState<CnrdsReport | null>(null)
  const [submitVisible, setSubmitVisible] = useState(false)
  const [submitTargetId, setSubmitTargetId] = useState<string>('')
  const [reviewedBy, setReviewedBy] = useState('')

  const fetchReports = useCallback(async () => {
    setLoading(true)
    try {
      const data = await cnrdsApi.list()
      setReports(data)
    } catch {
      message.error('加载报告列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
	queueMicrotask(() => { void fetchReports() })
  }, [fetchReports])

  const handleGenerateMonthly = async () => {
    if (!period) {
      message.warning('请选择月份')
      return
    }
    setGenerating(true)
    try {
      await cnrdsApi.monthly(period)
      message.success('月度上报包已生成')
      void fetchReports()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: { message?: string } } } }
      message.error(err?.response?.data?.error?.message || '生成月报失败')
    } finally {
      setGenerating(false)
    }
  }

  const handleGenerateEvent = async () => {
    if (!eventPatientId) {
      message.warning('请输入患者ID')
      return
    }
    setGenerating(true)
    try {
      await cnrdsApi.event(String(eventPatientId), eventType)
      message.success('事件上报已生成')
      setEventVisible(false)
      void fetchReports()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: { message?: string } } } }
      message.error(err?.response?.data?.error?.message || '生成事件报失败')
    } finally {
      setGenerating(false)
    }
  }

  const handleExport = async (report: CnrdsReport) => {
    try {
      const { blob, filename, contentType } = await cnrdsApi.exportCsv(report.id)
      if (contentType && !contentType.includes('csv') && !contentType.includes('text')) {
        const text = await blob.text()
        try {
          const err = JSON.parse(text) as { error?: { message?: string } }
          message.error(err.error?.message || '导出失败')
        } catch {
          message.error('导出失败：服务器返回异常')
        }
        return
      }
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
      message.success('导出成功')
      void fetchReports()
    } catch (e: unknown) {
      const err = e as { response?: { data?: Blob | { error?: { message?: string } } } }
      if (err.response?.data instanceof Blob) {
        try {
          const text = await err.response.data.text()
          const parsed = JSON.parse(text) as { error?: { message?: string } }
          message.error(parsed.error?.message || '导出失败')
          return
        } catch {
          message.error('导出失败：服务器返回异常')
          return
        }
      }
      message.error(err?.response?.data && 'error' in err.response.data ? err.response.data.error?.message || '导出失败' : '导出失败')
    }
  }

  const handleSubmit = async () => {
    if (!reviewedBy.trim()) {
      message.warning('请输入核对人')
      return
    }
    try {
      await cnrdsApi.submit(submitTargetId, reviewedBy.trim())
      message.success('已提交')
      setSubmitVisible(false)
      setReviewedBy('')
      void fetchReports()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: { message?: string } } } }
      message.error(err?.response?.data?.error?.message || '提交失败')
    }
  }

  const openPreview = async (report: CnrdsReport) => {
    try {
      const full = await cnrdsApi.get(report.id)
      setPreviewReport(full)
      setPreviewVisible(true)
    } catch {
      message.error('加载报告详情失败')
    }
  }

  const parseContentRows = (content: string): CnrdsContentRow[] => {
    try {
      const obj = JSON.parse(content) as { rows?: CnrdsContentRow[] }
      return obj.rows ?? []
    } catch {
      return []
    }
  }

  const columns: ColumnsType<CnrdsReport> = [
    { title: '周期', dataIndex: 'period', key: 'period', width: 100 },
    {
      title: '类型', dataIndex: 'reportType', key: 'reportType', width: 80,
      render: (t: string, r: CnrdsReport) => t === 'monthly' ? '月报' : `事件(${r.eventType})`,
    },
    { title: '患者数', dataIndex: 'patientCount', key: 'patientCount', width: 80 },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (s: string) => <Tag color={statusColors[s] || 'default'}>{statusLabels[s] || s}</Tag>,
    },
    {
      title: '核对人', dataIndex: 'reviewedBy', key: 'reviewedBy', width: 100,
      render: (v: string) => v || '-',
    },
    {
      title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 170,
      render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作', key: 'actions', width: 260,
      render: (_: unknown, record: CnrdsReport) => (
        <Space size="small">
          <Button size="small" icon={<Eye size={14} />} onClick={() => { void openPreview(record) }}>预览</Button>
          {record.status === 'draft' || record.status === 'exported' ? (
            <Button size="small" icon={<Download size={14} />} onClick={() => { void handleExport(record) }}>导出</Button>
          ) : null}
          {record.status === 'exported' ? (
            <Button size="small" type="primary" icon={<Send size={14} />}
              onClick={() => { setSubmitTargetId(record.id); setSubmitVisible(true); setReviewedBy('') }}
            >提交</Button>
          ) : null}
        </Space>
      ),
    },
  ]

  const previewRows = previewReport ? parseContentRows(previewReport.content) : []

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
        <FileSpreadsheet size={24} />
        CNRDS 上报
      </h2>

      <div className="mb-4 flex flex-wrap items-center gap-3">
        <DatePicker picker="month" onChange={(d) => { setPeriod(d ? d.format('YYYY-MM') : '') }} placeholder="选择月份" />
        <Button type="primary" icon={<Plus size={14} />} loading={generating} onClick={() => { void handleGenerateMonthly() }}>
          生成月度上报包
        </Button>
        <Button icon={<Plus size={14} />} onClick={() => { setEventVisible(true) }}>
          生成转归事件上报
        </Button>
        <Button onClick={() => { void fetchReports() }}>刷新</Button>
      </div>

      <Table
        dataSource={reports}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20 }}
        size="middle"
      />

      <Modal title="生成转归事件上报" open={eventVisible} onCancel={() => setEventVisible(false)}
        onOk={() => { void handleGenerateEvent() }} confirmLoading={generating}>
        <Space direction="vertical" className="w-full">
          <div>
            <label className="block mb-1 text-sm">患者ID</label>
            <InputNumber className="w-full" min={1} value={eventPatientId}
              onChange={(v) => { setEventPatientId(v) }} placeholder="输入患者ID" />
          </div>
          <div>
            <label className="block mb-1 text-sm">事件类型</label>
            <Select className="w-full" value={eventType} onChange={(v) => { setEventType(v) }}
              options={eventTypeOptions} />
          </div>
        </Space>
      </Modal>

      <Modal title="报告预览" open={previewVisible} onCancel={() => setPreviewVisible(false)}
        footer={null} width={900}>
        {previewReport && (
          <>
            <Descriptions size="small" column={3} bordered className="mb-4">
              <Descriptions.Item label="周期">{previewReport.period}</Descriptions.Item>
              <Descriptions.Item label="类型">{previewReport.reportType === 'monthly' ? '月报' : `事件(${previewReport.eventType})`}</Descriptions.Item>
              <Descriptions.Item label="患者数">{previewReport.patientCount}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={statusColors[previewReport.status]}>{statusLabels[previewReport.status]}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="核对人">{previewReport.reviewedBy || '-'}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{new Date(previewReport.createdAt).toLocaleString('zh-CN')}</Descriptions.Item>
            </Descriptions>
            {previewRows.length > 0 && (
              <Table
                dataSource={previewRows.map((r, i) => ({ ...r, key: i }))}
                columns={[
                  { title: '患者ID', dataIndex: 'patientId', width: 80 },
                  { title: '姓名', dataIndex: 'name', width: 80 },
                  { title: '性别', dataIndex: 'gender', width: 50 },
                  { title: 'Hb', dataIndex: 'hb', width: 60, render: (v: number | null) => v != null ? v : '-' },
                  { title: 'Ca', dataIndex: 'ca', width: 60, render: (v: number | null) => v != null ? v : '-' },
                  { title: 'P', dataIndex: 'p', width: 60, render: (v: number | null) => v != null ? v : '-' },
                  { title: 'PTH', dataIndex: 'pth', width: 70, render: (v: number | null) => v != null ? v : '-' },
                  { title: 'Albumin', dataIndex: 'albumin', width: 70, render: (v: number | null) => v != null ? v : '-' },
                  { title: 'Kt/V', dataIndex: 'ktv', width: 60, render: (v: number | null) => v != null ? v : '-' },
                  { title: '转归', dataIndex: 'outcomeType', width: 80 },
                  { title: '转归日期', dataIndex: 'outcomeDate', width: 120 },
                  { title: '死亡原因', dataIndex: 'deathReason', width: 100, render: (v: string) => v || '-' },
                ]}
                size="small"
                scroll={{ x: 1000 }}
                pagination={{ pageSize: 50 }}
              />
            )}
          </>
        )}
      </Modal>

      <Modal title="提交报告" open={submitVisible} onCancel={() => setSubmitVisible(false)}
        onOk={() => { void handleSubmit() }}>
        <div>
          <label className="block mb-1 text-sm">核对人</label>
          <input
            type="text"
            className="border rounded px-3 py-2 w-full"
            value={reviewedBy}
            onChange={(e) => { setReviewedBy(e.target.value) }}
            placeholder="请输入核对人姓名"
          />
        </div>
      </Modal>
    </div>
  )
}
