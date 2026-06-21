import { useState, useEffect } from 'react'
import { Modal, Form, Select, DatePicker, Input, InputNumber, Checkbox, message } from 'antd'
import type { Dayjs } from 'dayjs'
import { vascularEventApi } from '@/services/vascularEventApi'

const EVENT_TYPES: Record<string, string> = {
  establish: '建立',
  maturation: '成熟评估',
  first_use: '首次使用',
  physical_check: '物理检查',
  complication: '并发症',
  intervention: '介入',
  failure: '失功',
  replacement: '更换',
}

const COMP_TYPES = [
  { label: '狭窄', value: 'stenosis' },
  { label: '血栓形成', value: 'thrombosis' },
  { label: '感染', value: 'infection' },
  { label: '出血', value: 'bleeding' },
  { label: '假性动脉瘤', value: 'aneurysm' },
  { label: '缺血综合征', value: 'ischemia' },
]

interface VascularEventModalProps {
  patientId: number
  accessOptions: { value: number; label: string }[]
  onClose: () => void
}

export default function VascularEventModal({ patientId, accessOptions, onClose }: VascularEventModalProps) {
  const [form] = Form.useForm()
  const [saving, setSaving] = useState(false)
  const [eventType, setEventType] = useState<string>('maturation')

  useEffect(() => {
    form.setFieldsValue({ accessId: accessOptions[0]?.value })
  }, [accessOptions, form])

  const buildDetail = (): string => {
    const vals = form.getFieldsValue()
    const t = vals.eventType
    if (t === 'physical_check') {
      return JSON.stringify({ abnormal: vals.physicalAbnormal || false })
    }
    if (t === 'maturation') {
      return JSON.stringify({ bloodFlow: vals.bloodFlow || 0, usable: vals.usable || false })
    }
    if (t === 'complication') {
      return JSON.stringify({ compType: vals.compType || '', severity: vals.severity || '', desc: vals.complicationDesc || '' })
    }
    if (t === 'intervention') {
      return JSON.stringify({ procedure: vals.procedure || '', result: vals.interventionResult || '' })
    }
    return JSON.stringify({ note: vals.detailNote || '' })
  }

  const handleSubmit = async () => {
    try {
      const vals = await form.validateFields()
      setSaving(true)
      const eventDate = (vals.eventDate as Dayjs).format('YYYY-MM-DD')
      const detail = buildDetail()
      await vascularEventApi.recordEvent(patientId, {
        accessId: vals.accessId,
        eventType: vals.eventType,
        eventDate,
        detail,
        operatorId: vals.operatorId || '',
        note: vals.note || '',
      })
      message.success('事件已记录')
      onClose()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error('记录事件失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      title="录血管通路事件"
      open
      onCancel={onClose}
      onOk={handleSubmit}
      confirmLoading={saving}
      destroyOnClose
    >
      <Form form={form} layout="vertical" initialValues={{ eventType: 'maturation' }}>
        <Form.Item name="accessId" label="通路" rules={[{ required: true }]}>
          <Select options={accessOptions.map((o) => ({ label: o.label, value: o.value }))} />
        </Form.Item>
        <Form.Item name="eventType" label="事件类型" rules={[{ required: true }]}>
          <Select options={Object.entries(EVENT_TYPES).map(([k, v]) => ({ label: v, value: k }))} onChange={setEventType} />
        </Form.Item>
        <Form.Item name="eventDate" label="事件日期" rules={[{ required: true }]}>
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>

        {eventType === 'maturation' && (
          <>
            <Form.Item name="bloodFlow"><InputNumber placeholder="血流量 (ml/min)" style={{ width: '100%' }} /></Form.Item>
            <Form.Item name="usable" valuePropName="checked"><Checkbox>通路可用</Checkbox></Form.Item>
          </>
        )}
        {eventType === 'physical_check' && (
          <Form.Item name="physicalAbnormal" valuePropName="checked"><Checkbox>检查异常（震颤减弱/杂音异常等）</Checkbox></Form.Item>
        )}
        {eventType === 'complication' && (
          <>
            <Form.Item name="compType"><Select placeholder="并发症类型" options={COMP_TYPES} allowClear /></Form.Item>
            <Form.Item name="severity"><Input placeholder="严重程度" /></Form.Item>
            <Form.Item name="complicationDesc"><Input.TextArea placeholder="并发症描述" rows={2} /></Form.Item>
          </>
        )}
        {eventType === 'intervention' && (
          <>
            <Form.Item name="procedure"><Input placeholder="介入操作 (如 PTA/thrombectomy)" /></Form.Item>
            <Form.Item name="interventionResult"><Input placeholder="介入结果" /></Form.Item>
          </>
        )}

        <Form.Item name="operatorId"><Input placeholder="操作人" /></Form.Item>
        <Form.Item name="note"><Input.TextArea placeholder="备注" rows={2} /></Form.Item>
      </Form>
    </Modal>
  )
}
