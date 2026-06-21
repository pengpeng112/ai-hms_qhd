import { useState, useEffect, useCallback } from 'react';
import {
  Card, Table, Tag, Button, Alert, Modal, Form,
  DatePicker, Select, InputNumber, Input, message,
} from 'antd';
import dayjs from 'dayjs';
import { waterQualityApi } from '@/services/waterQualityApi';
import type { WaterQualityRecord, WaterQualityAlerts } from '@/services/waterQualityApi';
import WqHandleModal from './WqHandleModal';

const TYPE_LABEL: Record<string, string> = {
  bacteria: '细菌菌落数',
  endotoxin: '内毒素',
  conductivity: '电导率',
};

export default function WqRecords() {
  const [records, setRecords] = useState<WaterQualityRecord[]>([]);
  const [alerts, setAlerts] = useState<WaterQualityAlerts>({ exceed: [], due: [] });
  const [loading, setLoading] = useState(false);
  const [showAdd, setShowAdd] = useState(false);
  const [handleRec, setHandleRec] = useState<WaterQualityRecord | null>(null);
  const [form] = Form.useForm();

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const [list, al] = await Promise.all([
        waterQualityApi.list(),
        waterQualityApi.alerts(),
      ]);
      setRecords(list);
      setAlerts(al);
    } catch { /* swallow */ } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void reload(); }, [reload]);

  const onSubmitRecord = async () => {
    try {
      const vals = await form.validateFields();
      await waterQualityApi.record({
        testDate: (vals.testDate as dayjs.Dayjs).format('YYYY-MM-DD'),
        testType: vals.testType as string,
        samplePoint: vals.samplePoint as string,
        value: vals.value as number,
        unit: (vals.unit as string) || '',
      });
      message.success('检测记录已提交');
      setShowAdd(false);
      form.resetFields();
      void reload();
    } catch { /* validation or API error, antd form shows inline */ }
  };

  const columns = [
    { title: '日期', dataIndex: 'testDate', key: 'testDate', width: 110, render: (v: string) => v?.slice(0, 10) ?? '-' },
    { title: '项目', dataIndex: 'testType', key: 'testType', width: 110, render: (v: string) => TYPE_LABEL[v] ?? v },
    { title: '取样点', dataIndex: 'samplePoint', key: 'samplePoint', width: 110 },
    { title: '测值', key: 'val', width: 100, render: (_: unknown, r: WaterQualityRecord) => `${r.value} ${r.unit}` },
    { title: '阈值', dataIndex: 'standardLimit', key: 'standardLimit', width: 100 },
    {
      title: '结果', dataIndex: 'result', key: 'result', width: 80,
      render: (v: string) =>
        v === 'pass' ? <Tag color="green">合格</Tag> :
        v === 'fail' ? <Tag color="red">超标</Tag> :
        <Tag>待判</Tag>,
    },
    { title: '到期', dataIndex: 'nextDueDate', key: 'nextDueDate', width: 110, render: (v: string) => v?.slice(0, 10) ?? '-' },
    {
      title: '处置', key: 'handle', width: 140,
      render: (_: unknown, r: WaterQualityRecord) => {
        if (r.result === 'fail' && !r.handledAt) {
          return <Button danger size="small" onClick={() => setHandleRec(r)}>双确认</Button>;
        }
        return r.action || r.handledAt?.slice(0, 10) || '-';
      },
    },
  ];

  return (
    <Card title="检测记录" style={{ marginTop: 16 }}
      extra={<Button type="primary" onClick={() => setShowAdd(true)}>录入检测</Button>}
    >
      {alerts.exceed.length > 0 && (
        <Alert type="error" showIcon style={{ marginBottom: 8 }}
          message={`水质超标未处置 ${alerts.exceed.length} 项`} />
      )}
      {alerts.due.length > 0 && (
        <Alert type="warning" showIcon style={{ marginBottom: 8 }}
          message={`检测到期/将到期 ${alerts.due.length} 项`} />
      )}

      <Table<WaterQualityRecord>
        rowKey="id"
        dataSource={records}
        columns={columns}
        loading={loading}
        size="small"
        pagination={{ pageSize: 10 }}
      />

      {/* 录入弹窗 */}
      <Modal title="录入水质检测" open={showAdd}
        onOk={onSubmitRecord} onCancel={() => { setShowAdd(false); form.resetFields(); }}>
        <Form form={form} layout="vertical">
          <Form.Item name="testDate" label="检测日期" rules={[{ required: true, message: '请选择日期' }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="testType" label="检测项目" rules={[{ required: true, message: '请选择项目' }]}>
            <Select options={[
              { value: 'bacteria', label: '细菌菌落数' },
              { value: 'endotoxin', label: '内毒素' },
            ]} placeholder="请选择" />
          </Form.Item>
          <Form.Item name="samplePoint" label="取样点" rules={[{ required: true, message: '请选择取样点' }]}>
            <Select options={[
              { value: 'ro_outlet', label: '反渗水出口' },
              { value: 'dialysate', label: '透析液' },
            ]} placeholder="请选择" />
          </Form.Item>
          <Form.Item name="value" label="测值" rules={[{ required: true, message: '请填测值' }]}>
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="unit" label="单位">
            <Input placeholder="如 CFU/mL、EU/mL" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 双确认弹窗 */}
      {handleRec && (
        <WqHandleModal record={handleRec} onClose={() => { setHandleRec(null); void reload(); }} />
      )}
    </Card>
  );
}
