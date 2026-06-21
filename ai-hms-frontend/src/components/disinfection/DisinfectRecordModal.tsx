import { useState } from 'react'
import { Modal, Form, Select, Input, InputNumber, DatePicker, message } from 'antd'
import type { Dayjs } from 'dayjs'

import { disinfectionApi } from '@/services/disinfectionApi'

interface DisinfectRecordModalProps {
  deviceOptions: { value: number; label: string }[]
  onClose: () => void
}

const typeOptions = [
  { value: 'heat', label: '热消毒' },
  { value: 'terminal', label: '终末消毒' },
  { value: 'decalc', label: '除钙' },
  { value: 'enhanced', label: '强化消毒' },
]

const residualOptions = [
  { value: 'pass', label: '合格' },
  { value: 'fail', label: '不合格' },
  { value: '', label: '未检' },
]

const resultOptions = [
  { value: 'pass', label: '合格' },
  { value: 'fail', label: '不合格' },
]

export default function DisinfectRecordModal({
  deviceOptions,
  onClose,
}: DisinfectRecordModalProps) {
  const [form] = Form.useForm()
  const [submitting, setSubmitting] = useState(false)

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)

      const startTime = (values.startTime as Dayjs).toISOString()
      const endTime = (values.endTime as Dayjs).toISOString()

      await disinfectionApi.record({
        deviceId: values.deviceId as number,
        disinfectType: values.disinfectType as string,
        disinfectant: values.disinfectant as string | undefined,
        concentration: values.concentration as string | undefined,
        operatorId: values.operatorId as number,
        startTime,
        endTime,
        residualCheck: values.residualCheck as string | undefined,
        result: values.result as string | undefined,
        docRef: values.docRef as string | undefined,
        source: 'manual',
      })

      message.success('已登记')
      onClose()
    } catch {
      // validation failed or API error — antd shows field errors automatically
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      open
      title="登记消毒"
      onOk={handleOk}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
      width={520}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item name="deviceId" label="设备" rules={[{ required: true, message: '请选择设备' }]}>
          <Select options={deviceOptions} placeholder="选择设备" />
        </Form.Item>

        <Form.Item
          name="disinfectType"
          label="消毒类型"
          rules={[{ required: true, message: '请选择消毒类型' }]}
        >
          <Select options={typeOptions} placeholder="选择类型" />
        </Form.Item>

        <Form.Item name="disinfectant" label="消毒剂">
          <Input placeholder="如：过氧乙酸" />
        </Form.Item>

        <Form.Item name="concentration" label="浓度">
          <Input placeholder="如：0.5%" />
        </Form.Item>

        <Form.Item
          name="operatorId"
          label="操作人工号"
          rules={[{ required: true, message: '请输入操作人工号' }]}
        >
          <InputNumber style={{ width: '100%' }} placeholder="工号" />
        </Form.Item>

        <Form.Item
          name="startTime"
          label="起始时间"
          rules={[{ required: true, message: '请选择起始时间' }]}
        >
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item
          name="endTime"
          label="结束时间"
          rules={[{ required: true, message: '请选择结束时间' }]}
        >
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="residualCheck" label="残留检测">
          <Select options={residualOptions} placeholder="选择结果" allowClear />
        </Form.Item>

        <Form.Item name="result" label="结果">
          <Select options={resultOptions} placeholder="选择结果" allowClear />
        </Form.Item>

        <Form.Item name="docRef" label="文档引用">
          <Input placeholder="相关文档编号" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
