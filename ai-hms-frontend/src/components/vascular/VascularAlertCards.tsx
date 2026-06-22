import { useEffect, useState } from 'react'
import { Alert, Space } from 'antd'
import { vascularEventApi } from '@/services/vascularEventApi'
import type { VascReminder } from '@/services/vascularEventApi'

export default function VascularAlertCards() {
  const [alerts, setAlerts] = useState<VascReminder[]>([])

  useEffect(() => {
    vascularEventApi
      .alerts()
      .then((a) => setAlerts(a ?? []))
      .catch(() => {})
  }, [])

  const errorAlerts = alerts.filter((a) => a.kind === 'cvc_over_limit' || a.kind === 'failure_no_replace')
  const warnAlerts = alerts.filter((a) => a.kind === 'maturation_due' || a.kind === 'periodic_due' || a.kind === 'physical_abnormal' || a.kind === 'type_unrecognized')

  if (!alerts.length) return null

  return (
    <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }}>
      {errorAlerts.length > 0 && (
        <Alert
          type="error"
          showIcon
          message={`通路超时限/失功 ${errorAlerts.length} 例——请立即处理`}
        />
      )}
      {warnAlerts.length > 0 && (
        <Alert
          type="warning"
          showIcon
          message={`通路评估待办/异常 ${warnAlerts.length} 例——请安排检查`}
        />
      )}
    </Space>
  )
}
