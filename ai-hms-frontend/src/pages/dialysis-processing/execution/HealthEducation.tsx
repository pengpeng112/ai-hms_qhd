import { BookMarked, ClipboardList, MessageSquareMore } from 'lucide-react'
import type { RestTreatment } from '@/services'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
}

export default function HealthEducation({ patient, treatment }: Props) {
  return (
    <div className="space-y-6 pb-8">
      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-2 border-b border-slate-200 px-6 py-4">
          <BookMarked size={16} className="text-emerald-600" />
          <h3 className="text-sm font-black text-slate-800">健康宣教计划</h3>
        </div>
        <div className="grid grid-cols-1 gap-4 p-6 lg:grid-cols-3">
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">宣教对象</div>
            <input
              value={patient.name}
              readOnly
              className="h-11 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-semibold text-slate-700 outline-none"
            />
          </label>
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">当前治疗方式</div>
            <input
              value={treatment?.treatmentType || patient.treatmentPlan}
              readOnly
              className="h-11 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-semibold text-slate-700 outline-none"
            />
          </label>
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">当前状态</div>
            <input
              value={patient.status}
              readOnly
              className="h-11 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-semibold text-slate-700 outline-none"
            />
          </label>
          <label className="block lg:col-span-3">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">说明</div>
            <textarea
              rows={5}
              readOnly
              value="当前版本未接入独立的健康宣教后端数据源。为避免演示假数据误导联调，此页面暂只展示患者与治疗上下文。待后端接口或老库表口径确认后，再补录入、历史记录与保存能力。"
              className="w-full resize-none rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-slate-700 outline-none"
            />
          </label>
        </div>
      </section>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-2 border-b border-slate-200 px-6 py-4">
          <ClipboardList size={16} className="text-blue-600" />
          <h3 className="text-sm font-black text-slate-800">宣教记录</h3>
        </div>
        <div className="px-6 py-10 text-center text-sm text-slate-400">
          暂无真实宣教记录数据源，当前不展示静态假记录。
        </div>
      </section>

      <section className="rounded-3xl bg-slate-900 px-6 py-5 text-white shadow-lg">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <MessageSquareMore size={18} className="text-slate-300" />
            <span className="text-sm font-semibold">
              健康宣教页面已切换为真实上下文占位，不再展示静态样例数据。
            </span>
          </div>
          <div className="text-sm font-semibold text-slate-300">待后端接口补齐后再接录入</div>
        </div>
      </section>
    </div>
  )
}
