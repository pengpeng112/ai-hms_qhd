import { Modal, Button, Select, message, Radio, DatePicker } from 'antd'
import { useNavigate } from 'react-router-dom'
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
  const navigate = useNavigate()
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
    setLoading(true)
    try {
      const req: { startDate: string; weeks: number; wardId?: number; templateId?: number } = {
        startDate: startDate.format('YYYY-MM-DD'),
        weeks,
        wardId: wardId ?? undefined,
      }
      if (templateId) req.templateId = templateId
      const result = await restApi.generateSchedule(req)
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

  const hasTemplates = templates.length > 0

  return (
    <Modal
      title="生成排班草稿"
      open={open}
      onCancel={onClose}
      width={480}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button
          key="generate"
          type="primary"
          loading={loading}
          onClick={handleGenerate}
        >生成草稿</Button>,
      ]}
    >
      <div className="space-y-4">
        {/* 模板选择 */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">选择模板 <span className="text-slate-400">（可选）</span></label>
          {hasTemplates ? (
            <Select
              value={templateId}
              onChange={(v) => setTemplateId(v)}
              placeholder="选择已有模板（留空则从所有待排患者生成）..."
              className="w-full"
              allowClear
              options={templates.map((t) => ({
                value: t.template.id,
                label: `${t.template.name} (v${t.template.version}, ${t.itemCount}项)`,
              }))}
            />
          ) : (
            <div className="rounded-lg border border-amber-200 bg-amber-50 p-3">
              <p className="text-sm text-amber-700 mb-2">
                还没有创建排班模板，可在下方快捷创建。
              </p>
              <Button
                size="small"
                type="primary"
                onClick={() => {
                  onClose()
                  navigate('/schedule-templates/edit')
                }}
              >
                创建新模板
              </Button>
              <p className="text-xs text-slate-400 mt-2">
                留空则直接从所有待排班患者生成草稿（需患者有治疗计划）。
              </p>
            </div>
          )}
          {hasTemplates && (
            <div className="mt-1 flex items-center gap-2">
              <span className="text-xs text-slate-400">没有合适模板？</span>
              <Button
                type="link"
                size="small"
                className="p-0 h-auto text-xs"
                onClick={() => {
                  onClose()
                  navigate('/schedule-templates/edit')
                }}
              >
                新建模板
              </Button>
            </div>
          )}
        </div>

        {/* 周期 */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">周期</label>
          <Radio.Group value={weeks} onChange={(e) => setWeeks(e.target.value)}>
            <Radio.Button value={2}>2周</Radio.Button>
            <Radio.Button value={4}>4周</Radio.Button>
          </Radio.Group>
        </div>

        {/* 起始日期 */}
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

        {/* 说明 */}
        <div className="rounded-lg bg-slate-50 border border-slate-200 p-3">
          <p className="text-xs text-slate-500 leading-relaxed">
            {templateId
              ? '将根据所选模板中患者骨架自动展开透析日，按两轮分配规则生成草稿排班。'
              : '将从所有待排班患者中读取治疗骨架，按频率展开透析日，自动分配机位生成草稿排班。'}
            排不下的进入冲突队列待人工处理。
          </p>
        </div>
      </div>
    </Modal>
  )
}
