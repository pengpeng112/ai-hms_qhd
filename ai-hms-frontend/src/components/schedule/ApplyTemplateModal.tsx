import { Modal, Button, Select, DatePicker, message } from 'antd'
import { useState, useEffect } from 'react'
import { restApi, getErrorMessage } from '@/services/restClient'
import type { ScheduleTemplateResponse } from '@/services/restClient'
import dayjs from 'dayjs'

interface ApplyTemplateModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
  wardId?: number
}

export default function ApplyTemplateModal({ open, onClose, onSuccess, wardId }: ApplyTemplateModalProps) {
  const [loading, setLoading] = useState(false)
  const [templates, setTemplates] = useState<ScheduleTemplateResponse[]>([])
  const [templateId, setTemplateId] = useState<number | undefined>(undefined)
  const [targetDate, setTargetDate] = useState(dayjs())

  useEffect(() => {
    if (open) {
      restApi.listScheduleTemplates(wardId)
        .then(setTemplates)
        .catch((error) => {
          message.error(getErrorMessage(error))
          setTemplates([])
        })
    }
  }, [open, wardId])

  const handleApply = async () => {
    if (!templateId) {
      message.warning('请选择模板')
      return
    }
    setLoading(true)
    try {
      const req: { templateId: number; targetDate: string; wardId?: number } = {
        templateId,
        targetDate: targetDate.format('YYYY-MM-DD'),
      }
      if (wardId) req.wardId = wardId
      const data = await restApi.applyScheduleTemplate(req)
      message.success(`模板应用成功: 创建 ${data.count} 条排班`)
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
      title="应用排班模板"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="apply" type="primary" loading={loading} onClick={handleApply}>应用模板</Button>,
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
          <label className="block text-sm text-gray-600 mb-1">目标日期</label>
          <DatePicker value={targetDate} onChange={(d) => d && setTargetDate(d)} className="w-full" allowClear={false} />
        </div>
      </div>
    </Modal>
  )
}
