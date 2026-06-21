import { useEffect, useState } from 'react'
import { Alert, Space } from 'antd'
import { disinfectionApi } from '@/services/disinfectionApi'
import type { MachineDisinfStatus } from '@/services/disinfectionApi'

export default function DisinfectAlertCards({ deviceIds }: { deviceIds: number[] }) {
  const [blocked, setBlocked] = useState<MachineDisinfStatus[]>([])
  const [warn, setWarn] = useState<MachineDisinfStatus[]>([])
  useEffect(() => {
    if (!deviceIds.length) return
    disinfectionApi.alerts(deviceIds).then(a => { setBlocked(a.blocked); setWarn(a.warn) }).catch(() => {})
  }, [deviceIds])
  if (!blocked.length && !warn.length) return null
  return (
    <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }}>
      {blocked.length > 0 && <Alert type="error" showIcon message={`${blocked.length} 台机残留检测不合格已停用——须合格消毒后解除`} />}
      {warn.length > 0 && <Alert type="warning" showIcon message={`${warn.length} 台机消毒待办（终末/除钙/热消毒）`} />}
    </Space>
  )
}
