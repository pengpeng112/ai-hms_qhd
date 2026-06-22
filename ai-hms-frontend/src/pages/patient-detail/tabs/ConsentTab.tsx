// Consent Tab - 知情同意（C2）：开具 / 患者家属签署 / 到期复签

import { useState, useEffect, useCallback, useMemo } from 'react'
import { message, Modal } from 'antd'
import { FileSignature, Plus, Loader2, ShieldX, AlertTriangle, PenLine, Ban, CheckCircle2 } from 'lucide-react'
import { SectionHeader, DetailCard } from '@/components/ui'
import { getUserInfo } from '@/utils/token'
import {
  getConsentTemplates,
  issueConsent,
  getConsents,
  signConsent,
  revokeConsent,
  type ConsentTemplate,
  type ConsentRecord,
  type ConsentStatus,
} from '@/services/consentApi'
import type { Patient } from '@/types/original'

interface ConsentTabProps {
  patient: Patient
}

const STATUS_META: Record<ConsentStatus, { label: string; badge: string }> = {
  pending: { label: '待签', badge: 'bg-orange-50 text-orange-600 border-orange-100' },
  signed: { label: '已签', badge: 'bg-green-50 text-green-600 border-green-100' },
  expired: { label: '已过期', badge: 'bg-red-50 text-red-600 border-red-100' },
  revoked: { label: '已撤销', badge: 'bg-slate-100 text-slate-500 border-slate-200' },
}

function fmt(ts?: string): string {
  if (!ts) return '-'
  const d = new Date(ts)
  if (isNaN(d.getTime())) return ts
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export default function ConsentTab({ patient }: ConsentTabProps) {
  const [templates, setTemplates] = useState<ConsentTemplate[]>([])
  const [records, setRecords] = useState<ConsentRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [issuing, setIssuing] = useState(false)
  const [issueType, setIssueType] = useState('')

  const load = useCallback(async () => {
    if (!patient.id) return
    setLoading(true)
    try {
      const data = await getConsents({ patientId: String(patient.id) })
      setRecords(data)
    } catch (err) {
      console.error('加载知情同意失败:', err)
      message.error('加载知情同意失败')
    } finally {
      setLoading(false)
    }
  }, [patient.id])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    getConsentTemplates().then((t) => {
      setTemplates(t)
      if (t.length > 0) setIssueType(t[0].consentType)
    }).catch(() => {})
  }, [])

  const templateName = (t: string) => templates.find((x) => x.consentType === t)?.name || t
  const user = getUserInfo()

  const pendingExpired = useMemo(
    () => records.filter((r) => r.status === 'pending' || r.status === 'expired'),
    [records]
  )

  const handleIssue = async () => {
    if (!issueType) {
      message.warning('请选择同意书类型')
      return
    }
    setIssuing(true)
    try {
      await issueConsent({ patientId: String(patient.id), consentType: issueType, issuedBy: user?.id })
      message.success('已开具，待患者/家属签署')
      await load()
    } catch (err) {
      console.error('开具失败:', err)
      message.error('开具失败')
    } finally {
      setIssuing(false)
    }
  }

  const handleSign = (rec: ConsentRecord) => {
    let signedBy = '患者本人'
    Modal.confirm({
      title: `签署：${templateName(rec.consentType)}`,
      icon: <PenLine size={18} />,
      content: (
        <div className="pt-2">
          <p className="text-xs text-slate-500 mb-2">签署人（患者本人 / 家属-关系-姓名）：</p>
          <input
            defaultValue={signedBy}
            onChange={(e) => (signedBy = e.target.value)}
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm"
          />
          <p className="text-[11px] text-slate-400 mt-2">签署将写入统一电子签留痕(sign_record)并归档。</p>
        </div>
      ),
      okText: '确认签署',
      cancelText: '取消',
      onOk: async () => {
        try {
          await signConsent(rec.id, { signedBy: signedBy.trim() || '患者本人' })
          message.success('已签署')
          await load()
        } catch (err) {
          console.error('签署失败:', err)
          message.error('签署失败')
        }
      },
    })
  }

  const handleRevoke = (rec: ConsentRecord) => {
    Modal.confirm({
      title: '撤销同意书',
      content: `确定撤销「${templateName(rec.consentType)}」吗？`,
      okText: '撤销',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await revokeConsent(rec.id)
          message.success('已撤销')
          await load()
        } catch (err) {
          console.error('撤销失败:', err)
          message.error('撤销失败')
        }
      },
    })
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="animate-spin text-blue-500" size={32} />
        <span className="ml-3 text-slate-500">加载中...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in pb-10">
      {/* 待签/过期提醒 */}
      {pendingExpired.length > 0 && (
        <div className="flex items-center gap-3 px-5 py-3 bg-orange-50 border border-orange-100 rounded-2xl text-orange-600">
          <AlertTriangle size={18} className="shrink-0" />
          <span className="text-sm font-black">
            {pendingExpired.length} 份同意书待签 / 已过期需复签：{pendingExpired.map((r) => templateName(r.consentType)).join('、')}
          </span>
        </div>
      )}

      {/* 开具 */}
      <DetailCard>
        <SectionHeader icon={FileSignature} title="开具知情同意书" />
        <div className="mt-4 flex items-end gap-3 flex-wrap">
          <div className="flex-1 min-w-[240px]">
            <label className="block text-xs font-black text-slate-500 mb-1.5">同意书类型</label>
            <select
              value={issueType}
              onChange={(e) => setIssueType(e.target.value)}
              className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm"
            >
              {templates.map((t) => (
                <option key={t.consentType} value={t.consentType}>{t.name}（{t.timing}）</option>
              ))}
            </select>
          </div>
          <button
            onClick={handleIssue}
            disabled={issuing || !issueType}
            className="px-5 py-2 bg-blue-600 text-white rounded-lg text-xs font-black hover:bg-blue-700 transition-all flex items-center gap-1.5 shadow-sm disabled:opacity-50"
          >
            {issuing ? <Loader2 size={14} className="animate-spin" /> : <Plus size={14} />} 开具
          </button>
        </div>
      </DetailCard>

      {/* 同意书列表 */}
      <DetailCard>
        <SectionHeader icon={FileSignature} title="知情同意记录" />
        <div className="mt-4 space-y-3">
          {records.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-slate-300">
              <ShieldX size={40} className="mb-3 opacity-20" />
              <p className="font-bold">暂无知情同意记录</p>
            </div>
          ) : (
            records.map((rec) => {
              const st = STATUS_META[rec.status]
              const needSign = rec.status === 'pending' || rec.status === 'expired'
              return (
                <div key={rec.id} className={`rounded-2xl border p-4 ${needSign ? 'border-orange-200 bg-orange-50/40' : 'border-slate-100 bg-white'}`}>
                  <div className="flex items-center justify-between flex-wrap gap-2">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-black text-slate-800">{templateName(rec.consentType)}</span>
                      <span className={`px-2 py-0.5 rounded-lg text-[10px] font-black border ${st.badge}`}>{st.label}</span>
                      {rec.templateVersion && <span className="text-[10px] text-slate-400">{rec.templateVersion}</span>}
                    </div>
                    <div className="flex items-center gap-1.5">
                      {needSign && (
                        <button onClick={() => handleSign(rec)} className="flex items-center gap-1 px-3 py-1.5 bg-white border border-blue-100 text-blue-600 rounded-xl text-[10px] font-black hover:bg-blue-600 hover:text-white transition-all shadow-sm">
                          <PenLine size={12} /> 签署
                        </button>
                      )}
                      {rec.status !== 'revoked' && (
                        <button onClick={() => handleRevoke(rec)} className="flex items-center gap-1 px-3 py-1.5 bg-white border border-red-100 text-red-400 rounded-xl text-[10px] font-black hover:bg-red-600 hover:text-white transition-all shadow-sm">
                          <Ban size={12} /> 撤销
                        </button>
                      )}
                    </div>
                  </div>
                  <div className="mt-3 grid grid-cols-2 md:grid-cols-4 gap-2 text-[11px] border-t border-slate-100 pt-3">
                    <div><span className="text-slate-400">开具：</span><span className="text-slate-600">{fmt(rec.createdAt)}</span></div>
                    <div><span className="text-slate-400">签署人：</span><span className="text-slate-600">{rec.signedBy || '-'}</span></div>
                    <div><span className="text-slate-400">签署：</span><span className="text-slate-600">{fmt(rec.signedAt)}</span></div>
                    <div>
                      <span className="text-slate-400">到期：</span>
                      <span className="text-slate-600">{rec.expiresAt ? fmt(rec.expiresAt) : '长期'}</span>
                    </div>
                  </div>
                  {rec.status === 'signed' && (
                    <div className="mt-2 flex items-center gap-1.5 text-[11px] text-green-600">
                      <CheckCircle2 size={13} /> 已归档电子签留痕
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>
      </DetailCard>
    </div>
  )
}
