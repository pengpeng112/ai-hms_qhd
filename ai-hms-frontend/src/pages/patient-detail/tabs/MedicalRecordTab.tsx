// Medical Record Tab - 临床病史档案

import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { message } from 'antd'
import { FileHeart, ShieldCheck, Heart, History, PlusCircle, CheckCircle2, ChevronRight, Loader2 } from 'lucide-react'
import { SectionHeader, DetailCard } from '@/components/ui'
import { HistoryDetailModal, OutcomeRecordModal, OutcomeHistoryModal } from '@/components/patient/modals'
import type { Patient } from '@/types/original'
import type { OutcomeRecord } from '../types'
import {
  restApi,
  getErrorMessage,
  type OutcomeRecordApi,
  type MedicalHistoryApiResponse,
  type HistoryContent,
  type HistoryNamedContent,
} from '@/services/restClient'
import { useOutcomeDict } from '@/hooks/useOutcomeDict'

// 类型转换函数：API 类型 -> 本地类型
function toOutcomeRecord(api: OutcomeRecordApi, getTypeName: (code: string) => string, getReasonName: (code: string) => string): OutcomeRecord {
  return {
    id: api.id,
    type: api.type,
    typeName: getTypeName(api.type),
    reason: api.reason,
    reasonName: getReasonName(api.reason),
    time: api.time,
    remarks: api.remarks,
    registrar: api.registrar,
    registrationTime: api.registrationTime,
    isDoorRule: api.isDoorRule,
  }
}

// 类型转换函数：本地类型 -> API 类型
function toOutcomeRecordApi(record: OutcomeRecord): Omit<OutcomeRecordApi, 'id'> {
  return {
    type: record.type,
    reason: record.reason,
    time: record.time,
    remarks: record.remarks,
    registrar: record.registrar,
    registrationTime: record.registrationTime,
    isDoorRule: record.isDoorRule ?? false,
  }
}

interface MedicalRecordTabProps {
  patient: Patient
}

// 空的临床病史初始值
const EMPTY_MEDICAL_HISTORY: MedicalHistoryApiResponse = {
  current: { content: '' },
  past: { content: '' },
  transfusion: { content: '' },
  marital: { content: '' },
  family: { content: '' },
  diagnosis: { content: '' },
  primary: { name: '', content: '' },
  pathology: { name: '', content: '' },
  allergen: { name: '', content: '' },
  tumor: { name: '', content: '' },
  complication: { name: '', content: '' },
}

export default function MedicalRecordTab({ patient }: MedicalRecordTabProps) {
  const { t } = useTranslation('patient')
  const { getTypeName, getReasonName, loadDicts } = useOutcomeDict()

  // 加载状态
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  // 临床病史状态
  const [selectedHistoryDetail, setSelectedHistoryDetail] = useState<{ title: string; key: string; data: HistoryContent | HistoryNamedContent } | null>(null)
  const [isOutcomeRecordModalOpen, setIsOutcomeRecordModalOpen] = useState(false)
  const [isOutcomeHistoryModalOpen, setIsOutcomeHistoryModalOpen] = useState(false)
  const [editingOutcomeRecord, setEditingOutcomeRecord] = useState<OutcomeRecord | null>(null)
  const [outcomeModalKey, setOutcomeModalKey] = useState(0)

  // 数据状态（初始为空）
  const [outcomeRecords, setOutcomeRecords] = useState<OutcomeRecord[]>([])
  const [localMedicalHistory, setLocalMedicalHistory] = useState<MedicalHistoryApiResponse>(EMPTY_MEDICAL_HISTORY)

  // 加载数据
  const loadData = useCallback(async () => {
    console.log('[MedicalRecordTab] loadData called', { patientId: patient.id })
    if (!patient.id) {
      console.warn('[MedicalRecordTab] No patient.id, skipping load')
      return
    }
    setLoading(true)
    try {
      const [historyData, outcomeData] = await Promise.all([
        restApi.getMedicalHistory(patient.id),
        restApi.getOutcomeRecords(patient.id),
      ])
      console.log('[MedicalRecordTab] Data loaded', { historyData, outcomeData })
      setLocalMedicalHistory(historyData)
      setOutcomeRecords(outcomeData.map(api => toOutcomeRecord(api, getTypeName, getReasonName)))
    } catch (error) {
      console.error('[MedicalRecordTab] 加载临床病史失败:', error)
      // 保持空数据状态
    } finally {
      setLoading(false)
    }
  }, [patient.id, getTypeName, getReasonName])

  useEffect(() => {
    loadDicts()
  }, [loadDicts])

  useEffect(() => {
    console.log('[MedicalRecordTab] useEffect triggered', { patientId: patient.id })
    loadData()
  }, [patient.id, loadData])

  // 助手组件：单条病史项
  const HistoryItem = ({ label, value, onClick }: { label: string, value: string, onClick: () => void }) => (
    <div
      onClick={onClick}
      className="p-5 bg-slate-50/50 rounded-2xl border border-slate-100 flex items-center justify-between group hover:border-blue-300 hover:bg-white cursor-pointer transition-all"
    >
      <div>
        <p className="text-[10px] text-slate-400 font-black uppercase mb-1 tracking-widest">{label}</p>
        <p className="text-sm font-black text-slate-700 group-hover:text-blue-600 transition-colors">{value || t('action.viewDetails')}</p>
      </div>
      <ChevronRight size={16} className="text-slate-200 group-hover:text-blue-500 transition-colors" />
    </div>
  )

  // 保存病史（局部更新）
  const handleSaveHistory = async (newData: HistoryContent | HistoryNamedContent) => {
    if (!selectedHistoryDetail || !patient.id) return
    setSaving(true)
    try {
      const key = selectedHistoryDetail.key as keyof MedicalHistoryApiResponse
      const updatePayload: Partial<MedicalHistoryApiResponse> = {
        [key]: newData,
      }
      const result = await restApi.saveMedicalHistory(patient.id, updatePayload)
      setLocalMedicalHistory(result)
      setSelectedHistoryDetail(null)
      message.success(t('message.historyRecordSaved'))
    } catch (error) {
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  // 保存转归记录
  const handleSaveOutcome = async (newData: OutcomeRecord) => {
    console.log('[MedicalRecordTab] handleSaveOutcome called', { patientId: patient.id, newData })
    if (!patient.id) {
      console.error('[MedicalRecordTab] patient.id is empty!', patient)
      return
    }
    setSaving(true)
    try {
      console.log('[MedicalRecordTab] Calling API', { patientId: patient.id, isNew: !editingOutcomeRecord })
      if (editingOutcomeRecord) {
        // 更新
        const result = await restApi.updateOutcomeRecord(patient.id, editingOutcomeRecord.id, toOutcomeRecordApi(newData))
        console.log('[MedicalRecordTab] Update success', result)
        setOutcomeRecords(prev => prev.map(r => r.id === editingOutcomeRecord.id ? toOutcomeRecord(result, getTypeName, getReasonName) : r))
      } else {
        // 新建
        const result = await restApi.createOutcomeRecord(patient.id, toOutcomeRecordApi(newData))
        console.log('[MedicalRecordTab] Create success', result)
        setOutcomeRecords(prev => {
          const newRecords = [toOutcomeRecord(result, getTypeName, getReasonName), ...prev]
          console.log('[MedicalRecordTab] State updated', { prev: prev.length, new: newRecords.length, firstRecord: newRecords[0] })
          return newRecords
        })
      }
      setIsOutcomeRecordModalOpen(false)
      setEditingOutcomeRecord(null)
      message.success(editingOutcomeRecord ? t('message.outcomeUpdated') : t('message.outcomeCreated'))
    } catch (error) {
      console.error('[MedicalRecordTab] Save failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  // 删除转归记录
  const handleDeleteOutcome = async (id: string) => {
    if (!patient.id) return
    if (!window.confirm(t('confirm.deleteOutcomeRecord'))) return
    try {
      await restApi.deleteOutcomeRecord(patient.id, id)
      setOutcomeRecords(prev => prev.filter(r => r.id !== id))
      message.success(t('message.outcomeDeleted'))
    } catch (error) {
      message.error(getErrorMessage(error))
    }
  }

  // 调试日志
  console.log('[MedicalRecordTab] Rendering', { loading, outcomeRecordsCount: outcomeRecords.length, firstOutcome: outcomeRecords[0] })

  // 加载中状态
  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    )
  }

  return (
    <div className="grid grid-cols-12 gap-6 animate-fade-in pb-10">
      <div className="col-span-12 lg:col-span-8 space-y-6">
        <DetailCard>
          <SectionHeader icon={FileHeart} title={t('section.basicClinicalHistory')} />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
            <HistoryItem label={t('label.currentIllness')} value={localMedicalHistory.current.content ? t('action.viewDetails') : t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.currentIllness'), key: 'current', data: localMedicalHistory.current })} />
            <HistoryItem label={t('label.pastHistory')} value={localMedicalHistory.past.content ? t('action.viewDetails') : t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.pastHistory'), key: 'past', data: localMedicalHistory.past })} />
            <HistoryItem label={t('label.transfusionHistory')} value={localMedicalHistory.transfusion.content ? t('action.viewDetails') : t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.transfusionHistory'), key: 'transfusion', data: localMedicalHistory.transfusion })} />
            <HistoryItem label={t('label.maritalHistory')} value={localMedicalHistory.marital.content ? t('action.viewDetails') : t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.maritalHistory'), key: 'marital', data: localMedicalHistory.marital })} />
            <HistoryItem label={t('label.familyHistory')} value={localMedicalHistory.family.content ? t('action.viewDetails') : t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.familyHistory'), key: 'family', data: localMedicalHistory.family })} />
            <HistoryItem label={t('label.diseaseDiagnosis')} value={localMedicalHistory.diagnosis.content ? t('action.viewDetails') : t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.diseaseDiagnosis'), key: 'diagnosis', data: localMedicalHistory.diagnosis })} />
          </div>
        </DetailCard>

        <DetailCard>
          <SectionHeader icon={ShieldCheck} title={t('section.specialtyRecords')} />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
            <HistoryItem label={t('label.primaryDisease')} value={localMedicalHistory.primary.name || t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.primaryDisease'), key: 'primary', data: localMedicalHistory.primary })} />
            <HistoryItem label={t('label.pathology')} value={localMedicalHistory.pathology.name || t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.pathology'), key: 'pathology', data: localMedicalHistory.pathology })} />
            <HistoryItem label={t('label.allergenInfo')} value={localMedicalHistory.allergen.name || t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.allergenInfo'), key: 'allergen', data: localMedicalHistory.allergen })} />
            <HistoryItem label={t('label.tumorHistory')} value={localMedicalHistory.tumor.name || t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.tumorHistory'), key: 'tumor', data: localMedicalHistory.tumor })} />
            <HistoryItem label={t('label.complicationInfo')} value={localMedicalHistory.complication.name || t('label.notFilled')} onClick={() => setSelectedHistoryDetail({ title: t('label.complicationInfo'), key: 'complication', data: localMedicalHistory.complication })} />
          </div>
        </DetailCard>
      </div>

      <div className="col-span-12 lg:col-span-4 space-y-6">
        {outcomeRecords.length > 0 ? (
          <DetailCard className="border-l-8 border-l-blue-500 flex flex-col h-fit">
            <SectionHeader icon={Heart} title={t('section.outcomeAssessment')} />
            <div className="mt-4 space-y-4 flex-1">
              <div className="flex items-center gap-3">
                <div className="inline-flex items-center px-3 py-1 bg-blue-50 text-blue-700 rounded-full font-black text-xs border border-blue-200">
                  {patient.outcome?.status || t('status.maintenanceTreatment')}
                </div>
                <span className="text-xs text-slate-400">{t('label.totalRecords', { count: outcomeRecords.length })}</span>
              </div>
              <div className="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-2">
                <div className="flex items-center gap-2">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-black ${outcomeRecords[0].typeName === '在科' ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-orange-50 text-orange-700 border border-orange-200'}`}>
                    {outcomeRecords[0].typeName || '未知'}
                  </span>
                  <span className="text-sm font-bold text-slate-700">{outcomeRecords[0].reasonName || t('label.noReason')}</span>
                </div>
                <p className="text-xs text-slate-400">{outcomeRecords[0].time}</p>
                {outcomeRecords[0].remarks && <p className="text-xs text-slate-500 leading-relaxed">{outcomeRecords[0].remarks}</p>}
              </div>
            </div>
            <div className="mt-6 pt-4 border-t border-slate-100 flex gap-2">
              <button
                onClick={() => setIsOutcomeHistoryModalOpen(true)}
                className="flex-1 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-2xl text-xs font-black transition-all flex items-center justify-center gap-1.5"
              >
                <History size={14}/> {t('action.viewHistory')}
              </button>
              <button
                onClick={() => { setEditingOutcomeRecord(null); setOutcomeModalKey(k => k + 1); setIsOutcomeRecordModalOpen(true); }}
                className="flex-1 py-2.5 bg-blue-600 text-white hover:bg-blue-700 rounded-2xl text-xs font-black shadow-lg shadow-blue-100 transition-all flex items-center justify-center gap-1.5"
              >
                <PlusCircle size={14}/> {t('action.addOutcome')}
              </button>
            </div>
          </DetailCard>
        ) : (
          <DetailCard className="bg-blue-600 text-white border-none shadow-xl shadow-blue-100 flex flex-col h-fit">
            <SectionHeader icon={Heart} title={t('section.outcomeAssessment')} dark />
            <div className="mt-4 space-y-6 flex-1">
              <div>
                <p className="text-[10px] text-blue-200 font-black mb-1 uppercase tracking-widest">{t('label.managementStatus')}</p>
                <div className="inline-flex items-center px-4 py-1 bg-white/20 backdrop-blur-md rounded-full font-black text-xs border border-white/30">
                  {patient.outcome?.status || t('status.maintenanceTreatment')}
                </div>
              </div>
              <p className="text-sm text-blue-50 font-bold leading-relaxed">{t('label.noOutcomeRecord')}</p>
            </div>
            <div className="mt-8 pt-4 border-t border-white/20">
              <button
                onClick={() => { setEditingOutcomeRecord(null); setOutcomeModalKey(k => k + 1); setIsOutcomeRecordModalOpen(true); }}
                className="w-full py-2.5 bg-white text-blue-600 hover:bg-blue-50 rounded-2xl text-xs font-black shadow-lg transition-all flex items-center justify-center gap-1.5"
              >
                <PlusCircle size={14}/> {t('action.addOutcome')}
              </button>
            </div>
          </DetailCard>
        )}

        <DetailCard className="border-l-8 border-l-amber-500 bg-amber-50/20">
          <SectionHeader icon={CheckCircle2} title={t('section.followupReminder')} />
          <ul className="mt-4 space-y-3">
            {[t('followup.dryWeightAssess'), t('followup.ipthMonitor'), t('followup.eyeCheck')].map((note, i) => (
              <li key={i} className="flex items-start gap-3 group">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 mt-1.5 shrink-0 group-hover:scale-125 transition-transform"></div>
                <span className="text-xs font-bold text-slate-600 leading-relaxed">{note}</span>
              </li>
            ))}
          </ul>
        </DetailCard>
      </div>

      {/* 详细病史编辑/保存弹窗 */}
      {selectedHistoryDetail && (
        <HistoryDetailModal
          key={selectedHistoryDetail.key}
          isOpen={selectedHistoryDetail !== null}
          onClose={() => setSelectedHistoryDetail(null)}
          onSave={handleSaveHistory}
          title={selectedHistoryDetail.title}
          historyKey={selectedHistoryDetail.key}
          data={selectedHistoryDetail.data}
          saving={saving}
        />
      )}

      {/* 治疗转归历史列表 */}
      <OutcomeHistoryModal
        isOpen={isOutcomeHistoryModalOpen}
        onClose={() => setIsOutcomeHistoryModalOpen(false)}
        records={outcomeRecords}
        onEdit={(r: OutcomeRecord) => { setEditingOutcomeRecord(r); setOutcomeModalKey(k => k + 1); setIsOutcomeRecordModalOpen(true); }}
        onDelete={handleDeleteOutcome}
      />

      {/* 治疗转归填写表单 */}
      <OutcomeRecordModal
        key={outcomeModalKey}
        isOpen={isOutcomeRecordModalOpen}
        onClose={() => setIsOutcomeRecordModalOpen(false)}
        onSave={handleSaveOutcome}
        initialData={editingOutcomeRecord}
        saving={saving}
      />
    </div>
  )
}
