import { useState } from 'react';
import { Modal, Radio, Input, message } from 'antd';
import { waterQualityApi } from '@/services/waterQualityApi';
import type { WaterQualityRecord } from '@/services/waterQualityApi';

export default function WqHandleModal({ record, onClose }: { record: WaterQualityRecord; onClose: () => void }) {
  const [role, setRole] = useState<'engineer' | 'head_nurse'>('engineer');
  const [signerId, setSignerId] = useState('');
  const [signerName, setSignerName] = useState('');
  const [action, setAction] = useState('');
  const submit = async () => {
    if (!signerId || !signerName) { message.warning('请填签字人'); return; }
    try {
      const r = await waterQualityApi.handle(record.id, { role, signerId, signerName, action });
      message.success(r?.handledAt ? '已双确认处置' : '已记录本次签字，等待另一方确认');
      onClose();
    } catch (e: unknown) { message.error(e instanceof Error ? e.message : '处置失败'); }
  };
  return (
    <Modal title="水质超标双确认（工程师+护士长）" open onOk={submit} onCancel={onClose}>
      <div style={{ marginBottom: 8 }}>项目 {record.testType} · 测值 {record.value}{record.unit} · 阈值 {record.standardLimit}</div>
      <div style={{ marginBottom: 8 }}>我的角色：
        <Radio.Group value={role} onChange={(e) => setRole(e.target.value)}>
          <Radio value="engineer">工程师</Radio>
          <Radio value="head_nurse">护士长</Radio>
        </Radio.Group>
      </div>
      <Input style={{ marginBottom: 8 }} placeholder="签字人工号" value={signerId} onChange={(e) => setSignerId(e.target.value)} />
      <Input style={{ marginBottom: 8 }} placeholder="签字人姓名" value={signerName} onChange={(e) => setSignerName(e.target.value)} />
      <Input.TextArea placeholder="处置动作 / 恢复确认" value={action} onChange={(e) => setAction(e.target.value)} rows={2} />
    </Modal>
  );
}
