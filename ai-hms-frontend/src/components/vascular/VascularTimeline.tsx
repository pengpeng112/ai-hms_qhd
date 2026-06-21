import { useEffect, useState, useCallback } from 'react'
import { Timeline, Alert, Button, Space } from 'antd'
import { Clock, Plus } from 'lucide-react'
import { vascularEventApi } from '@/services/vascularEventApi'
import type { VascTimelineEntry, VascReminder } from '@/services/vascularEventApi'
import VascularEventModal from './VascularEventModal'

const EVENT_LABELS: Record<string, string> = {
  establish: '建立',
  maturation: '成熟评估',
  first_use: '首次使用',
  physical_check: '物理检查',
  complication: '并发症',
  intervention: '介入',
  failure: '失功',
  replacement: '更换',
}

const REMINDER_SEVERITY: Record<string, 'error' | 'warning'> = {
  maturation_due: 'warning',
  cvc_over_limit: 'error',
  periodic_due: 'warning',
  physical_abnormal: 'error',
  failure_no_replace: 'error',
  type_unrecognized: 'warning',
}

interface VascularTimelineProps {
  patientId: number
  accessOptions: { value: number; label: string }[]
}

export default function VascularTimeline({ patientId, accessOptions }: VascularTimelineProps) {
  const [timeline, setTimeline] = useState<VascTimelineEntry[]>([])
  const [reminders, setReminders] = useState<VascReminder[]>([])
  const [showModal, setShowModal] = useState(false)

  const load = useCallback(async () => {
    try {
      const [tl, rm] = await Promise.all([
        vascularEventApi.timeline(patientId),
        vascularEventApi.reminders(patientId),
      ])
      setTimeline(tl)
      setReminders(rm)
    } catch {
      // silent
    }
  }, [patientId])

  useEffect(() => { load() }, [load])

  const eventColor = (type: string) => {
    switch (type) {
      case 'establish': return '#3b82f6'
      case 'maturation': return '#22c55e'
      case 'first_use': return '#8b5cf6'
      case 'physical_check': return '#f59e0b'
      case 'complication': return '#ef4444'
      case 'intervention': return '#f97316'
      case 'failure': return '#dc2626'
      case 'replacement': return '#06b6d4'
      default: return '#94a3b8'
    }
  }

  const fmtDate = (d?: string) => {
    if (!d) return '--'
    return new Date(d).toLocaleDateString('zh-CN')
  }

  const accessLabel = (id: number) => {
    const opt = accessOptions.find((o) => o.value === id)
    return opt?.label || `通路 #${id}`
  }

  return (
    <div className="mt-6 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-black uppercase tracking-wider flex items-center text-slate-800">
          <Clock size={18} className="mr-2 text-purple-600" /> 全程时间线
        </h3>
        <Button
          type="primary"
          size="small"
          icon={<Plus size={14} />}
          onClick={() => setShowModal(true)}
          disabled={accessOptions.length === 0}
        >
          录事件
        </Button>
      </div>

      {reminders.length > 0 && (
        <Space direction="vertical" style={{ width: '100%' }}>
          {reminders.map((r, i) => (
            <Alert
              key={`${r.kind}-${i}`}
              type={REMINDER_SEVERITY[r.kind] || 'warning'}
              showIcon
              message={r.message}
            />
          ))}
        </Space>
      )}

      {timeline.length > 0 ? (
        <Timeline
          items={timeline.map((e) => ({
            color: eventColor(e.eventType),
            children: (
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold text-slate-500">{fmtDate(e.eventDate)}</span>
                  <span className="text-[10px] font-black px-1.5 py-0.5 rounded" style={{ backgroundColor: eventColor(e.eventType) + '18', color: eventColor(e.eventType) }}>
                    {EVENT_LABELS[e.eventType] || e.eventType}
                  </span>
                  {e.eventType !== 'establish' && <span className="text-[10px] text-slate-400">{accessLabel(e.accessId)}</span>}
                </div>
                {e.note && <div className="text-xs text-slate-600 mt-1">{e.note}</div>}
                {e.detail && e.eventType !== 'establish' && (
                  <div className="text-[10px] text-slate-400 mt-0.5">{e.detail}</div>
                )}
              </div>
            ),
          }))}
        />
      ) : (
        <div className="text-center py-8 text-slate-300 text-xs font-bold">暂无时间线数据</div>
      )}

      {showModal && (
        <VascularEventModal
          patientId={patientId}
          accessOptions={accessOptions}
          onClose={() => { setShowModal(false); load() }}
        />
      )}
    </div>
  )
}
