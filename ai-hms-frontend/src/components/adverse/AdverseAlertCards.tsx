import { useEffect, useState } from 'react'
import { Alert, Space } from 'antd'
import { adverseEventApi } from '@/services/adverseEventApi'
import type { AdverseEvent } from '@/services/adverseEventApi'

export default function AdverseAlertCards() {
  const [severeUnreported, setSevereUnreported] = useState<AdverseEvent[]>([])
  const [severeOverdue, setSevereOverdue] = useState<AdverseEvent[]>([])
  const [pending, setPending] = useState<AdverseEvent[]>([])

  useEffect(() => {
    adverseEventApi
      .alerts()
      .then((a) => {
        setSevereUnreported(a.severeUnreported ?? [])
        setSevereOverdue(a.severeOverdue ?? [])
        setPending(a.pending ?? [])
      })
      .catch(() => {})
  }, [])

  if (!severeUnreported.length && !severeOverdue.length && !pending.length) return null

  return (
    <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }}>
      {severeOverdue.length > 0 && (
        <Alert
          type="error"
          showIcon
          message={`严重不良事件超时未上报 ${severeOverdue.length} 例——请立即上报护士长+科主任`}
        />
      )}
      {severeUnreported.length > 0 && (
        <Alert
          type="error"
          showIcon
          message={`严重不良事件待上报 ${severeUnreported.length} 例——须6小时内上报`}
        />
      )}
      {pending.length > 0 && (
        <Alert
          type="warning"
          showIcon
          message={`不良事件待处理 ${pending.length} 例——请跟进受理/结案`}
        />
      )}
    </Space>
  )
}
