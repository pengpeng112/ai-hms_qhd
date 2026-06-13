import { message } from 'antd'
import { AlertCircle, BookMarked, CheckCircle2, Clock, FileText, MessageSquareMore, RotateCcw } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { restApi } from '@/services'
import type { HealthEducationContentApi, PatientHealthEducationApi, RestTreatment } from '@/services'
import { getErrorMessage } from '@/services/restClient'
import type { CreatePatientHealthEducationRequest } from '@/services/restClient'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  treatmentLoading?: boolean
}

interface EducationFormState {
  contentId: string
  educationType: string
  educationTime: string
  finishTime: string
  educationResult: string
  nurseSign: string
  patientSign: string
  note: string
}

const today = new Date().toISOString().slice(0, 10)

const EMPTY_FORM: EducationFormState = {
  contentId: '',
  educationType: '口头宣教',
  educationTime: today,
  finishTime: today,
  educationResult: '',
  nurseSign: '',
  patientSign: '',
  note: '',
}

const QUICK_TEMPLATES = [
  { label: '已掌握', text: '患者已掌握本次宣教内容。' },
  { label: '需继续宣教', text: '需在下次透析时继续宣教。' },
  { label: '家属已知晓', text: '家属已知晓本次宣教内容。' },
]

function formatDate(value?: string | null) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toISOString().slice(0, 10)
}

export default function HealthEducation({ patient, treatment, treatmentLoading = false }: Props) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [contents, setContents] = useState<HealthEducationContentApi[]>([])
  const [records, setRecords] = useState<PatientHealthEducationApi[]>([])
  const [form, setForm] = useState<EducationFormState>(EMPTY_FORM)

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        const [contentList, recordList] = await Promise.all([
          restApi.getHealthEducationContents(),
          restApi.getPatientHealthEducations(patient.id),
        ])
        setContents(contentList)
        setRecords(recordList)
        setForm((current) => ({ ...current, contentId: current.contentId || contentList[0]?.id || '' }))
      } catch (error) {
        console.error('[HealthEducation] load failed', error)
        message.error(getErrorMessage(error))
      } finally {
        setLoading(false)
      }
    }
    void loadData()
  }, [patient.id])

  const selectedContent = useMemo(
    () => contents.find((item) => item.id === form.contentId) || null,
    [contents, form.contentId]
  )

  const missingRequiredFields = useMemo(() => [
    !form.contentId ? '宣教内容' : '',
    !form.educationTime ? '宣教日期' : '',
    !form.nurseSign.trim() ? '宣教人' : '',
  ].filter(Boolean), [form.contentId, form.educationTime, form.nurseSign])

  const updateField = (key: keyof EducationFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const appendNoteTemplate = (text: string) => {
    setForm((current) => ({
      ...current,
      note: current.note ? `${current.note}\n${text}` : text,
    }))
  }

  const resetForm = () => {
    setForm({ ...EMPTY_FORM, contentId: contents[0]?.id || '' })
  }

  const handleSave = async () => {
    if (!form.contentId) {
      message.warning('请选择宣教内容题目')
      return
    }
    try {
      setSaving(true)
      const payload: CreatePatientHealthEducationRequest = {
        healthEducationId: form.contentId,
        educationTime: form.educationTime,
        educationType: form.educationType.trim() || undefined,
        educationResult: form.educationResult.trim() || undefined,
        nurseSign: form.nurseSign.trim() || undefined,
        patientSign: form.patientSign.trim() || undefined,
        finishTime: form.finishTime || undefined,
        note: form.note.trim() || undefined,
      }
      const created = await restApi.createPatientHealthEducation(patient.id, payload)
      setRecords((items) => [created, ...items])
      message.success('健康宣教记录已保存')
    } catch (error) {
      console.error('[HealthEducation] save failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-4 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm font-semibold text-blue-700">
          正在加载新患者治疗上下文，宣教页面暂不展示上一位患者的治疗方式。
        </section>
      ) : null}

      <div className="flex flex-wrap items-center gap-3 text-xs text-slate-500">
        <span className="text-lg font-black text-slate-900">{patient.name}</span>
        <span>ID: {patient.id}</span>
        <span>{patient.gender} / {patient.age}岁</span>
        <span>方案: {patient.treatmentPlan || '--'}</span>
        <span className="rounded-md bg-slate-100 px-2 py-0.5 font-semibold">干体重 {patient.dryWeight || 0}kg</span>
      </div>

      <div className="flex items-center justify-between rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-2 text-xs">
        <span className="font-semibold text-emerald-700">健康宣教内容来源：Auxiliary_HealthEducation</span>
        <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-bold ${missingRequiredFields.length > 0 ? 'bg-amber-100 text-amber-700' : 'bg-emerald-200 text-emerald-800'}`}>
          {missingRequiredFields.length > 0 ? `缺 ${missingRequiredFields.length} 项` : '可保存'}
        </span>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1fr_1.2fr_1fr]">
        <section className="overflow-hidden rounded-xl border border-blue-100 bg-white shadow-sm">
          <div className="flex items-center gap-2 border-b border-blue-50 bg-blue-50/40 px-4 py-3">
            <BookMarked size={15} className="text-blue-600" />
            <h3 className="text-sm font-bold text-slate-800">宣教内容选择</h3>
          </div>
          <div className="space-y-3 p-4">
            <label className="block">
              <span className="mb-1.5 block text-[11px] font-semibold text-rose-500">* 宣教内容题目</span>
              <select value={form.contentId} onChange={(e) => updateField('contentId', e.target.value)} disabled={loading} className="h-9 w-full rounded-lg border border-slate-200 bg-white px-3 text-sm font-semibold text-slate-800 outline-none">
                {contents.length === 0 ? <option value="">暂无宣教内容</option> : null}
                {contents.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>

            <div className="flex flex-wrap gap-1.5">
              {contents.slice(0, 6).map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => updateField('contentId', item.id)}
                  className={`rounded-lg border px-2.5 py-1.5 text-left text-xs font-semibold transition ${
                    form.contentId === item.id
                      ? 'border-blue-300 bg-blue-50 text-blue-700'
                      : 'border-slate-100 bg-white text-slate-500 hover:border-slate-200 hover:bg-slate-50'
                  }`}
                >
                  {item.name}
                </button>
              ))}
            </div>

            {selectedContent?.description && (
              <div className="rounded-lg bg-slate-50 px-3 py-2.5 text-xs leading-5 text-slate-500">
                {selectedContent.description}
              </div>
            )}

            <div className="flex items-center gap-2 rounded-lg border border-slate-100 bg-slate-50 px-3 py-2.5">
              <span className={`text-xs font-semibold ${form.educationType === '口头宣教' ? 'text-blue-600' : 'text-slate-400'}`}>宣教方式:</span>
              <select value={form.educationType} onChange={(e) => updateField('educationType', e.target.value)} className="h-7 flex-1 rounded border border-slate-200 bg-white px-2 text-xs font-semibold outline-none">
                <option value="口头宣教">口头宣教</option>
                <option value="书面宣教">书面宣教</option>
                <option value="视频宣教">视频宣教</option>
              </select>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <label className="block">
                <span className="mb-1 block text-[11px] font-semibold text-rose-500">* 宣教日期</span>
                <input type="date" value={form.educationTime} onChange={(e) => updateField('educationTime', e.target.value)} className="h-9 w-full rounded-lg border border-slate-200 px-3 text-xs font-semibold outline-none" />
              </label>
              <label className="block">
                <span className="mb-1 block text-[11px] font-semibold text-slate-400">完成日期</span>
                <input type="date" value={form.finishTime} onChange={(e) => updateField('finishTime', e.target.value)} className="h-9 w-full rounded-lg border border-slate-200 px-3 text-xs font-semibold outline-none" />
              </label>
            </div>
          </div>
        </section>

        <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center gap-2 border-b border-slate-100 px-4 py-3">
            <FileText size={15} className="text-blue-600" />
            <h3 className="text-sm font-bold text-slate-800">本次宣教记录</h3>
          </div>
          <div className="space-y-3 p-4">
            <label className="block">
              <span className="mb-1.5 block text-[11px] font-semibold text-slate-400">详情记录描述</span>
              <textarea
                value={form.note}
                onChange={(e) => updateField('note', e.target.value)}
                placeholder="请输入本次宣教的具体内容、患者的反馈或重点强调的注意事项..."
                rows={6}
                className="w-full resize-none rounded-lg border border-slate-200 px-3 py-2.5 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
              />
            </label>

            <div className="flex flex-wrap items-center gap-1.5">
              <span className="text-[10px] text-slate-400">快捷模板：</span>
              {QUICK_TEMPLATES.map((tpl) => (
                <button
                  key={tpl.label}
                  type="button"
                  onClick={() => appendNoteTemplate(tpl.text)}
                  className="rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-[11px] font-medium text-slate-500 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700"
                >
                  {tpl.label}
                </button>
              ))}
            </div>

            <div className="grid grid-cols-3 gap-3">
              <label className="block">
                <span className="mb-1 block text-[11px] font-semibold text-slate-400">效果评价</span>
                <input value={form.educationResult} onChange={(e) => updateField('educationResult', e.target.value)} placeholder="已掌握/需继续" className="h-9 w-full rounded-lg border border-slate-200 px-2.5 text-xs font-semibold outline-none" />
              </label>
              <label className="block">
                <span className="mb-1 block text-[11px] font-semibold text-slate-400">病患/家属签字</span>
                <input value={form.patientSign} onChange={(e) => updateField('patientSign', e.target.value)} placeholder="签字" className="h-9 w-full rounded-lg border border-slate-200 px-2.5 text-xs font-semibold outline-none" />
              </label>
              <label className="block">
                <span className="mb-1 block text-[11px] font-semibold text-rose-500">* 宣教人</span>
                <input value={form.nurseSign} onChange={(e) => updateField('nurseSign', e.target.value)} placeholder="护士签名/工号" className="h-9 w-full rounded-lg border border-slate-200 px-2.5 text-xs font-semibold outline-none" />
              </label>
            </div>

            <div className="rounded-lg bg-slate-50 px-3 py-2 text-[11px] text-slate-400">
              填写路径：写记录 → 评价 → 签字 → 宣教人 → 保存
            </div>
          </div>
        </section>

        <section className="space-y-3">
          <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
            <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
              <div className="flex items-center gap-2">
                <Clock size={14} className="text-orange-500" />
                <h3 className="text-sm font-bold text-slate-800">宣教历史</h3>
              </div>
              <span className="text-[11px] font-bold text-slate-400">{records.length} 条</span>
            </div>
            <div className="max-h-48 overflow-y-auto">
              {records.length > 0 ? (
                <div className="divide-y divide-slate-50">
                  {records.slice(0, 5).map((item) => (
                    <div key={item.id} className="px-4 py-2.5">
                      <div className="text-xs font-bold text-slate-800 truncate">{item.healthEducationName}</div>
                      <div className="mt-0.5 flex items-center gap-2 text-[10px] text-slate-400">
                        <span>{item.operatorName || item.nurseSign || '--'}</span>
                        <span>{item.educationType || '--'}</span>
                        <span>{item.educationResult || '--'}</span>
                      </div>
                      <div className="mt-0.5 text-[10px] text-slate-300">{formatDate(item.educationTime)}</div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="py-6 text-center text-xs text-slate-400">
                  暂无宣教历史
                  <div className="mt-1 text-[10px] text-slate-300">保存后会在这里形成记录。</div>
                </div>
              )}
            </div>
          </div>

          <div className="overflow-hidden rounded-xl border border-amber-100 bg-amber-50/30">
            <div className="flex items-center gap-2 border-b border-amber-100 px-4 py-2.5">
              <AlertCircle size={13} className="text-amber-500" />
              <h3 className="text-xs font-bold text-amber-800">宣教完成前核对</h3>
            </div>
            <div className="space-y-1.5 p-3 text-[11px] text-amber-700">
              <div className="flex items-start gap-1.5">
                <CheckCircle2 size={11} className="mt-0.5 shrink-0 text-amber-400" />
                <span>患者是否理解本次宣教重点</span>
              </div>
              <div className="flex items-start gap-1.5">
                <CheckCircle2 size={11} className="mt-0.5 shrink-0 text-amber-400" />
                <span>是否需要家属共同知情</span>
              </div>
              <div className="flex items-start gap-1.5">
                <CheckCircle2 size={11} className="mt-0.5 shrink-0 text-amber-400" />
                <span>是否需要后续继续宣教</span>
              </div>
              <div className="flex items-start gap-1.5">
                <CheckCircle2 size={11} className="mt-0.5 shrink-0 text-amber-400" />
                <span>签字和日期是否完整</span>
              </div>
            </div>
          </div>
        </section>
      </div>

      <div className="sticky bottom-0 z-10 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-200 bg-white/95 px-4 py-3.5 shadow-lg backdrop-blur">
        <div className="flex flex-wrap items-center gap-x-5 gap-y-1 text-xs text-slate-500">
          <span>保存状态：<span className={`font-bold ${missingRequiredFields.length > 0 ? 'text-amber-600' : 'text-emerald-600'}`}>{missingRequiredFields.length > 0 ? '缺项待补' : '可保存'}</span></span>
          <span>所属患者：{patient.name}</span>
          <span>治疗方式：{treatment?.treatmentType || patient.treatmentPlan || '--'}</span>
          {missingRequiredFields.length > 0 && (
            <span className="font-bold text-rose-500">缺项：{missingRequiredFields.join('、')}</span>
          )}
        </div>
        <div className="flex gap-3">
          <button type="button" onClick={resetForm} className="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-5 py-2 text-sm font-semibold text-slate-500 transition hover:bg-slate-50">
            <RotateCcw size={14} />重置
          </button>
          <button type="button" onClick={() => void handleSave()} disabled={saving || loading || treatmentLoading} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white shadow-sm shadow-blue-900/20 transition hover:bg-blue-700 disabled:opacity-60">
            <MessageSquareMore size={15} />{saving ? '保存中...' : '保存宣教记录'}
          </button>
        </div>
      </div>
    </div>
  )
}
