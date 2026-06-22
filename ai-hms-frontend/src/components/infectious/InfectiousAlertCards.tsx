import { useEffect, useState } from 'react'
import { Alert, Space } from 'antd'
import { infectiousApi } from '@/services/infectiousApi'
import type { InfectiousRecord } from '@/services/infectiousApi'

export default function InfectiousAlertCards() {
  const [pos, setPos] = useState<InfectiousRecord[]>([])
  const [due, setDue] = useState<InfectiousRecord[]>([])
  useEffect(() => {
    infectiousApi
      .alerts()
      .then((a) => {
        setPos(a.positives ?? [])
        setDue(a.due ?? [])
      })
      .catch(() => {})
  }, [])
  if (!pos.length && !due.length) return null
  return (
    <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }}>
      {pos.length > 0 && (
        <Alert
          type="error"
          showIcon
          message={`传染病阳性未处置 ${pos.length} 人——须医生+护士长双处理（排班已冻结）`}
        />
      )}
      {due.length > 0 && (
        <Alert
          type="warning"
          showIcon
          message={`传染病筛查到期/将到期 ${due.length} 人——请安排复查`}
        />
      )}
    </Space>
  )
}
