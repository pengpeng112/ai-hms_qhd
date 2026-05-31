import { message } from 'antd'
import { BookMarked, ClipboardList, FileText, MessageSquareMore, RotateCcw } from 'lucide-react'
import { type ReactNode, useEffect, useMemo, useState } from 'react'
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

function formatDate(value?: string | null) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toISOString().slice(0, 10)
}

function Field({ label, children, required }: { label: string; children: ReactNode; required?: boolean }) {
  return (
    <label className="block min-w-0">
      <span className={`mb-2 block text-xs font-bold ${required ? 'text-rose-500' : 'text-slate-500'}`}>{required ? `* ${label}` : label}</span>
      {children}
    </label>
  )
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

  const updateField = (key: keyof EducationFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
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
    <div className="space-y-5 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">
          正在加载新患者治疗上下文，宣教页面暂不展示上一位患者的治疗方式。
        </section>
      ) : null}

      <section className="rounded-full border border-emerald-200 bg-emerald-50 px-6 py-4 text-sm font-black text-emerald-800">
        健康宣教内容来源于 Auxiliary_HealthEducation，保存记录写入 Auxiliary_PatientHealthEducation。
      </section>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[0.7fr_1fr]">
        <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center gap-2 border-b border-slate-100 px-6 py-4">
            <span className="h-6 w-1 rounded-full bg-blue-600" />
            <BookMarked size={16} className="text-blue-600" />
            <h3 className="text-sm font-black text-slate-800">核心宣教参数</h3>
          </div>
          <div className="space-y-5 p-6">
            <Field label="宣教内容题目" required>
              <select value={form.contentId} onChange={(e) => updateField('contentId', e.target.value)} disabled={loading} className="h-12 w-full rounded-lg border border-slate-200 bg-white px-4 text-sm font-bold text-slate-800 outline-none">
                {contents.length === 0 ? <option value="">暂无宣教内容</option> : null}
                {contents.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </Field>
            <div className="max-h-56 overflow-y-auto rounded-lg bg-slate-50 px-4 py-4 text-sm leading-7 text-slate-600">
              {selectedContent?.description || '暂无宣教内容描述'}
            </div>
            <Field label="宣教方式">
              <select value={form.educationType} onChange={(e) => updateField('educationType', e.target.value)} className="h-12 w-full rounded-lg border border-slate-200 bg-white px-4 text-sm font-bold text-slate-800 outline-none">
                <option value="口头宣教">口头宣教</option>
                <option value="书面宣教">书面宣教</option>
                <option value="视频宣教">视频宣教</option>
              </select>
            </Field>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field label="宣教日期" required><input type="date" value={form.educationTime} onChange={(e) => updateField('educationTime', e.target.value)} className="h-12 w-full rounded-lg border border-slate-200 px-4 text-sm font-bold outline-none" /></Field>
              <Field label="宣教人" required><input value={form.nurseSign} onChange={(e) => updateField('nurseSign', e.target.value)} placeholder="护士签名/工号" className="h-12 w-full rounded-lg border border-slate-200 px-4 text-sm font-bold outline-none" /></Field>
            </div>
            <Field label="效果评价"><input value={form.educationResult} onChange={(e) => updateField('educationResult', e.target.value)} placeholder="如：已掌握、需继续宣教..." className="h-12 w-full rounded-lg border border-slate-200 px-4 text-sm font-bold outline-none" /></Field>
            <Field label="病患/家属签字"><input value={form.patientSign} onChange={(e) => updateField('patientSign', e.target.value)} placeholder="病患或家属签字" className="h-12 w-full rounded-lg border border-slate-200 px-4 text-sm font-bold outline-none" /></Field>
            <Field label="完成日期"><input type="date" value={form.finishTime} onChange={(e) => updateField('finishTime', e.target.value)} className="h-12 w-full rounded-lg border border-slate-200 px-4 text-sm font-bold outline-none" /></Field>
          </div>
        </section>

        <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4">
            <div className="flex items-center gap-2"><FileText size={16} className="text-blue-600" /><h3 className="text-sm font-black text-slate-800">宣教详情记录描述</h3></div>
            <span className="text-xs font-black tracking-widest text-slate-300">DETAIL NARRATIVE</span>
          </div>
          <textarea value={form.note} onChange={(e) => updateField('note', e.target.value)} placeholder="请输入本次宣教的具体内容、患者的反馈或重点强调的注意事项..." className="h-[520px] w-full resize-none border-0 px-6 py-6 text-sm font-semibold text-slate-700 outline-none" />
          <div className="px-6 py-4 text-right text-xs font-semibold text-slate-400">保存后写入患者健康宣教记录</div>
        </section>
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4">
          <div className="flex items-center gap-2"><ClipboardList size={16} className="text-blue-600" /><h3 className="text-sm font-black text-slate-800">患者宣教历史</h3></div>
          <span className="text-xs font-bold text-slate-400">{records.length} 条</span>
        </div>
        <div className="space-y-3 p-6">
          {records.length > 0 ? records.map((item) => (
            <div key={item.id} className="rounded-lg border border-slate-100 px-5 py-4">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <div className="font-black text-slate-900">{item.healthEducationName}</div>
                  <div className="mt-2 text-xs font-semibold text-slate-500">宣教人：{item.operatorName || item.nurseSign || '--'} / 方式：{item.educationType || '--'} / 评价：{item.educationResult || '--'}</div>
                </div>
                <div className="text-sm font-bold text-slate-400">{formatDate(item.educationTime)}</div>
              </div>
            </div>
          )) : <div className="py-10 text-center text-sm text-slate-400">暂无患者宣教历史</div>}
        </div>
      </section>

      <div className="flex items-center justify-between border-t border-slate-200 pt-5">
        <div className="flex items-center gap-8 text-sm font-semibold text-slate-500"><span>保存状态：可保存</span><span>所属患者：{patient.name}</span><span>治疗方式：{treatment?.treatmentType || patient.treatmentPlan || '--'}</span></div>
        <div className="flex gap-3"><button type="button" onClick={resetForm} className="inline-flex items-center gap-2 rounded-lg border border-slate-200 px-8 py-3 text-sm font-semibold text-slate-500"><RotateCcw size={15} />重置当前表单</button><button type="button" onClick={() => void handleSave()} disabled={saving || loading || treatmentLoading} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-8 py-3 text-sm font-bold text-white disabled:opacity-60"><MessageSquareMore size={16} />{saving ? '保存中...' : '保存宣教记录'}</button></div>
      </div>
    </div>
  )
}
