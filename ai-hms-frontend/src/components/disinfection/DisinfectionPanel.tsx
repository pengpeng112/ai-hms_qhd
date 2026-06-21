import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Table, Tag } from 'antd'
import type { TableProps } from 'antd'

import { disinfectionApi } from '@/services/disinfectionApi'
import type { MachineDisinfStatus } from '@/services/disinfectionApi'
import DisinfectRecordModal from '@/components/disinfection/DisinfectRecordModal'

interface DisinfectionPanelProps {
  deviceIds: number[]
  deviceLabel?: (id: number) => string
}

const stateTagMap: Record<MachineDisinfStatus['state'], { color: string; text: string }> = {
  OK: { color: 'green', text: '正常' },
  WARN: { color: 'orange', text: '待办' },
  BLOCKED_RESIDUAL: { color: 'red', text: '停用' },
}

export default function DisinfectionPanel({ deviceIds, deviceLabel }: DisinfectionPanelProps) {
  const [machines, setMachines] = useState<MachineDisinfStatus[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)

  const reload = useCallback(async () => {
    if (deviceIds.length === 0) return
    setLoading(true)
    try {
      const data = await disinfectionApi.machines(deviceIds)
      setMachines(data)
    } finally {
      setLoading(false)
    }
  }, [deviceIds])

  useEffect(() => {
    void reload()
  }, [reload])

  const handleModalClose = useCallback(() => {
    setModalOpen(false)
    void reload()
  }, [reload])

  const columns: TableProps<MachineDisinfStatus>['columns'] = [
    {
      title: '机号',
      dataIndex: 'deviceId',
      key: 'deviceId',
      render: (_: unknown, r: MachineDisinfStatus) => deviceLabel?.(r.deviceId) ?? r.deviceId,
    },
    {
      title: '状态',
      dataIndex: 'state',
      key: 'state',
      render: (_: unknown, r: MachineDisinfStatus) => {
        const tag = stateTagMap[r.state]
        return <Tag color={tag.color}>{tag.text}</Tag>
      },
    },
    {
      title: '说明',
      dataIndex: 'reasons',
      key: 'reasons',
      render: (_: unknown, r: MachineDisinfStatus) => {
        if (r.state === 'BLOCKED_RESIDUAL') {
          return (
            <span style={{ color: '#ff4d4f' }}>
              {r.reasons.length > 0 ? r.reasons.join('；') : '该机已停用，须合格消毒后解除'}
            </span>
          )
        }
        return r.reasons.join('；') || '—'
      },
    },
    {
      title: '最近终末',
      dataIndex: 'lastTerminal',
      key: 'lastTerminal',
      render: (_: unknown, r: MachineDisinfStatus) => r.lastTerminal?.slice(0, 10) ?? '—',
    },
    {
      title: '最近除钙',
      dataIndex: 'lastDecalc',
      key: 'lastDecalc',
      render: (_: unknown, r: MachineDisinfStatus) => r.lastDecalc?.slice(0, 10) ?? '—',
    },
    {
      title: '最近热消毒',
      dataIndex: 'lastHeat',
      key: 'lastHeat',
      render: (_: unknown, r: MachineDisinfStatus) =>
        r.lastHeat?.slice(0, 16)?.replace('T', ' ') ?? '—',
    },
  ]

  const deviceOptions = deviceIds.map((id) => ({
    value: id,
    label: deviceLabel?.(id) ?? String(id),
  }))

  return (
    <>
      <Card
        title="消毒监管"
        extra={
          <Button type="primary" onClick={() => setModalOpen(true)}>
            登记消毒
          </Button>
        }
      >
        <Table<MachineDisinfStatus>
          rowKey="deviceId"
          columns={columns}
          dataSource={machines}
          loading={loading}
          pagination={false}
          size="middle"
        />
      </Card>

      {modalOpen && (
        <DisinfectRecordModal deviceOptions={deviceOptions} onClose={handleModalClose} />
      )}
    </>
  )
}
