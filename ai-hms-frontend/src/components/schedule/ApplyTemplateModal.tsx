import { Modal, Button, message } from 'antd'
import { useState } from 'react'
import { getErrorMessage } from '@/services/restClient'

interface ApplyTemplateModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function ApplyTemplateModal({ open, onClose, onSuccess }: ApplyTemplateModalProps) {
  const [loading, setLoading] = useState(false)

  const handleApply = async () => {
    setLoading(true)
    try {
      const token = localStorage.getItem('auth_token') || ''
      const res = await fetch('/api/v1/schedule/template/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({}),
      })
      if (!res.ok) throw new Error(`模板应用失败 (${res.status})`)
      const data = await res.json()
      message.success(`模板应用成功: 创建 ${data.createdCount || 0} 条，跳过 ${data.skippedCount || 0} 条`)
      onSuccess()
      onClose()
    } catch (error) {
      message.error(getErrorMessage(error))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="应用排班模板"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="apply" type="primary" loading={loading} onClick={handleApply}>应用模板</Button>,
      ]}
    >
      <p className="text-sm text-gray-600">将已保存的排班模板应用到当前周。</p>
      <p className="text-meta text-gray-400 mt-2">TODO: 模板选择 + 预览（待后端接口完善）</p>
    </Modal>
  )
}
