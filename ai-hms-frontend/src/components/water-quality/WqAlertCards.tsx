import { useEffect, useState } from 'react';
import { Alert, Space } from 'antd';
import { waterQualityApi } from '@/services/waterQualityApi';
import type { WaterQualityRecord } from '@/services/waterQualityApi';

export default function WqAlertCards() {
  const [exceed, setExceed] = useState<WaterQualityRecord[]>([]);
  const [due, setDue] = useState<WaterQualityRecord[]>([]);
  useEffect(() => { waterQualityApi.alerts().then(a => { setExceed(a.exceed ?? []); setDue(a.due ?? []); }).catch(() => {}); }, []);
  if (!exceed.length && !due.length) return null;
  return (
    <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }}>
      {exceed.length > 0 && <Alert type="error" showIcon message={`水质超标未处置 ${exceed.length} 项——工程师+护士长双确认`} />}
      {due.length > 0 && <Alert type="warning" showIcon message={`水质检测到期/将到期 ${due.length} 项——请安排检测`} />}
    </Space>
  );
}
