/**
 * InfectiousPanel - 传染病筛查面板（②带 patient-detail）
 * 展示筛查历史、分区门控状态、录入筛查、阳性双人处置
 */

import { useEffect, useState, useCallback } from 'react'
import { Card, Button, Table, Tag, Modal, Form, Select, DatePicker, Space, message } from 'antd'
import dayjs from 'dayjs'
import { infectiousApi } from '@/services/infectiousApi'
import type { InfectiousRecord, GateResult } from '@/services/infectiousApi'
import DispositionModal from '@/components/infectious/DispositionModal'

const SCREEN_ITEMS = ['HBsAg', '抗-HBs', 'HBeAg', 'HBcAb', '抗-HCV', 'HIV 抗体', 'TPPA', 'RPR']

const SCREEN_ITEM_OPTIONS = SCREEN_ITEMS.map((i) => ({ value: i, label: i }))

const RESULT_OPTIONS = [
  { value: 'negative', label: '阴性' },
  { value: 'positive', label: '阳性' },
  { value: 'indeterminate', label: '可疑' },
]

export default function InfectiousPanel({ patientId }: { patientId: string | number }) {
  const [rows, setRows] = useState<InfectiousRecord[]>([])
  const [gate, setGate] = useState<GateResult | null>(null)
  const [open, setOpen] = useState(false)
  const [disposeRec, setDisposeRec] = useState<InfectiousRecord | null>(null)
  const [form] = Form.useForm()

  const load = useCallback(async () => {
    try {
      const d = await infectiousApi.history(String(patientId))
      setRows(d.records)
      setGate(d.gate)
    } catch {
      message.error('传染病记录加载失败')
    }
  }, [patientId])

  useEffect(() => {
    void load()
  }, [load])

  const submit = async () => {
    const v = await form.validateFields()
    await infectiousApi.screen(String(patientId), {
      screenDate: (v.screenDate as ReturnType<typeof dayjs>).format('YYYY-MM-DD'),
      source: 'manual',
      items: ((v.items as Array<{ item: string; result: string }>) ?? []).map((it) => ({
        item: it.item,
        result: it.result,
      })),
    })
    setOpen(false)
    form.resetFields()
    void load()
    message.success('已录入')
  }

  const gateColor =
    gate?.state === 'FROZEN'
      ? 'red'
      : gate?.state === 'C_ZONE_CRRT'
        ? 'volcano'
        : 'orange'

  const columns = [
    {
      title: '日期',
      dataIndex: 'screenDate',
      render: (v?: string) => v?.slice(0, 10) ?? '-',
    },
    {
      title: '结果',
      dataIndex: 'resultOverall',
      render: (v: string) =>
        v === 'positive' ? (
          <Tag color="red">阳性</Tag>
        ) : v === 'pending' ? (
          <Tag color="orange">待定</Tag>
        ) : (
          <Tag color="green">阴性</Tag>
        ),
    },
    {
      title: '阳性项',
      dataIndex: 'positiveMarkers',
    },
    {
      title: '到期',
      dataIndex: 'nextDueDate',
      render: (v?: string) => v?.slice(0, 10) ?? '-',
    },
    {
      title: '处置',
      dataIndex: 'disposition',
      render: (v: string | undefined, r: InfectiousRecord) =>
        r.resultOverall === 'positive' && !r.handledAt ? (
          <Button danger size="small" onClick={() => setDisposeRec(r)}>
            双人处置
          </Button>
        ) : (
          v || '-'
        ),
    },
  ]

  return (
    <Card
      title="传染病筛查"
      size="small"
      extra={
        <Button type="primary" onClick={() => setOpen(true)}>
          录入筛查
        </Button>
      }
    >
      {gate && gate.state !== 'ALLOW_NORMAL' && (
        <Tag color={gateColor} style={{ marginBottom: 8 }}>
          {gate.state}：{gate.reason}
        </Tag>
      )}

      <Table
        size="small"
        rowKey="id"
        dataSource={rows}
        pagination={{ pageSize: 5 }}
        columns={columns}
      />

      <Modal title="录入传染病筛查" open={open} onOk={submit} onCancel={() => setOpen(false)}>
        <Form form={form} layout="vertical" initialValues={{ screenDate: dayjs() }}>
          <Form.Item name="screenDate" label="筛查日期" rules={[{ required: true }]}>
            <DatePicker />
          </Form.Item>
          <Form.List name="items">
            {(fields, { add, remove }) => (
              <>
                {fields.map((f) => (
                  <Space key={f.key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item name={[f.name, 'item']} rules={[{ required: true }]} noStyle>
                      <Select
                        style={{ width: 140 }}
                        options={SCREEN_ITEM_OPTIONS}
                        placeholder="项目"
                      />
                    </Form.Item>
                    <Form.Item name={[f.name, 'result']} rules={[{ required: true }]} noStyle>
                      <Select
                        style={{ width: 110 }}
                        options={RESULT_OPTIONS}
                        placeholder="结果"
                      />
                    </Form.Item>
                    <Button onClick={() => remove(f.name)}>删</Button>
                  </Space>
                ))}
                <Button type="dashed" onClick={() => add()} block>
                  + 加一项
                </Button>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>

      {disposeRec && (
        <DispositionModal
          patientId={String(patientId)}
          record={disposeRec}
          onClose={() => {
            setDisposeRec(null)
            void load()
          }}
        />
      )}
    </Card>
  )
}
