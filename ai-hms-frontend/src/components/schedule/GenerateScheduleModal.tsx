import { Modal, Button, Select, message, Radio, DatePicker } from 'antd'
import { useState, useEffect } from 'react'
import { restApi, getErrorMessage } from '@/services/restClient'
import type { ScheduleTemplateResponse } from '@/services/restClient'
import dayjs from 'dayjs'

interface GenerateScheduleModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
  wardId?: number
}

export default function GenerateScheduleModal({ open, onClose, onSuccess, wardId }: GenerateScheduleModalProps) {
  const [loading, setLoading] = useState(false)
  const [templates, setTemplates] = useState<ScheduleTemplateResponse[]>([])
  const [templateId, setTemplateId] = useState<number | undefined>(undefined)
  const [weeks, setWeeks] = useState(2)
  const [startDate, setStartDate] = useState(dayjs())

  useEffect(() => {
    if (open) {
      setTemplateId(undefined)
      setWeeks(2)
      setStartDate(dayjs())
      restApi.listScheduleTemplates(wardId)
        .then(setTemplates)
        .catch(() => setTemplates([]))
    }
  }, [open, wardId])

  const handleGenerate = async () => {
    if (!templateId) {
      message.warning('请选择模板')
      return
    }
    setLoading(true)
    try {
      const result = await restApi.generateSchedule({
        templateId,
        startDate: startDate.format('YYYY-MM-DD'),
        weeks,
        wardId: wardId ?? undefined,
      })
      const msg = `生成完毕: ${result.drafts} 条草稿`
      if (result.conflicts > 0) {
        message.warning(`${msg}, ${result.conflicts} 条冲突, 请查看冲突队列`)
      } else {
        message.success(msg)
      }
      onSuccess()
      onClose()
    } catch (error) {
      message.error(getErrorMessage(error))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="生成排班草稿"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="generate" type="primary" loading={loading} onClick={handleGenerate}>生成草稿</Button>,
      ]}
    >
      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-600 mb-1">选择模板</label>
          <Select
            value={templateId}
            onChange={(v) => setTemplateId(v)}
            placeholder="选择排班模板..."
            className="w-full"
            options={templates.map((t) => ({
              value: t.template.id,
              label: `${t.template.name} (v${t.template.version}, ${t.itemCount}项)`,
            }))}
          />
        </div>
        <div>
          <label className="block text-sm text-gray-600 mb-1">周期</label>
          <Radio.Group value={weeks} onChange={(e) => setWeeks(e.target.value)}>
            <Radio.Button value={2}>2周</Radio.Button>
            <Radio.Button value={4}>4周</Radio.Button>
          </Radio.Group>
        </div>
        <div>
          <label className="block text-sm text-gray-600 mb-1">起始日期</label>
          <DatePicker
            value={startDate}
            onChange={(d) => d && setStartDate(d)}
            className="w-full"
            allowClear={false}
            disabledDate={(d) => d && d.isBefore(dayjs(), 'day')}
          />
        </div>
        <div className="text-xs text-slate-400">
          将根据模板中患者骨架自动展开透析日，按两轮分配规则生成草稿排班。有冲突的进入冲突队列待人工处理。
        </div>
      </div>
    </Modal>
  )
}
