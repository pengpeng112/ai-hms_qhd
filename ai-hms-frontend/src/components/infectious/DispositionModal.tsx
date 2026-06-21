import { useState } from 'react'
import { Modal, Radio, Input, Descriptions, Alert, message } from 'antd'
import { infectiousApi } from '@/services/infectiousApi'
import type { InfectiousRecord } from '@/services/infectiousApi'

interface Props {
  patientId: string
  record: InfectiousRecord
  onClose: () => void
}

export default function DispositionModal({ patientId, record, onClose }: Props) {
  const [disposition, setDisposition] = useState<'c_zone_crrt' | 'transfer_out'>('c_zone_crrt')
  const [role, setRole] = useState<'doctor' | 'head_nurse'>('doctor')
  const [signerId, setSignerId] = useState('')
  const [signerName, setSignerName] = useState('')

  const submit = async () => {
    if (!signerId || !signerName) {
      message.warning('请填签字人')
      return
    }
    try {
      const r = await infectiousApi.dispose(patientId, record.id, {
        disposition,
        role,
        signerId,
        signerName,
      })
      message.success(r.handledAt ? '双人处置完成，已生效' : '已记录本次签字，等待另一方确认')
      onClose()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '处置失败'
      message.error(msg)
    }
  }

  return (
    <Modal title="阳性双人处置（医生+护士长）" open onOk={submit} onCancel={onClose}>
      <Alert
        type="warning"
        message="阳性患者须医生+护士长双签同意后才能安排。C 区 CRRT＝全警戒继续治疗；转外院＝转出退册。"
        style={{ marginBottom: 12 }}
      />
      <Descriptions size="small" column={1}>
        <Descriptions.Item label="阳性项">{record.positiveMarkers ?? '-'}</Descriptions.Item>
      </Descriptions>
      <div style={{ marginTop: 8 }}>
        处置：
        <Radio.Group value={disposition} onChange={(e) => setDisposition(e.target.value)}>
          <Radio value="c_zone_crrt">C 区全警戒 + CRRT</Radio>
          <Radio value="transfer_out">转外院（退册）</Radio>
        </Radio.Group>
      </div>
      <div style={{ marginTop: 8 }}>
        我的角色：
        <Radio.Group value={role} onChange={(e) => setRole(e.target.value)}>
          <Radio value="doctor">医生</Radio>
          <Radio value="head_nurse">护士长</Radio>
        </Radio.Group>
      </div>
      <Input
        style={{ marginTop: 8 }}
        placeholder="签字人工号"
        value={signerId}
        onChange={(e) => setSignerId(e.target.value)}
      />
      <Input
        style={{ marginTop: 8 }}
        placeholder="签字人姓名"
        value={signerName}
        onChange={(e) => setSignerName(e.target.value)}
      />
    </Modal>
  )
}
