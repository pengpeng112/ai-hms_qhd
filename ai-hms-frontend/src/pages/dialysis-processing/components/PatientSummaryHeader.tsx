import { FileOutput, Printer } from 'lucide-react'
import type { Patient } from '../types'

interface Props {
  patient: Patient
}

function SummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-[10px] font-semibold tracking-wide text-slate-400 uppercase">{label}</div>
      <div className="mt-1 text-sm font-bold text-slate-700">{value}</div>
    </div>
  )
}

export default function PatientSummaryHeader({ patient }: Props) {
  return (
    <div className="flex items-start justify-between gap-6">
      <div className="flex items-start gap-8 min-w-0">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-black text-slate-800">{patient.name}</h1>
            <span className="inline-flex items-center rounded-md bg-blue-600 text-white px-3 py-1 text-sm font-bold">
              {patient.bedId}
            </span>
          </div>
          <div className="mt-4 flex flex-wrap gap-8">
            <SummaryItem label="患者ID" value={patient.patientId} />
            <SummaryItem label="性别 / 年龄" value={`${patient.gender} / ${patient.age}岁`} />
            <SummaryItem label="费用类别" value={patient.costType} />
            <SummaryItem label="透龄" value={patient.dialysisAge} />
          </div>
        </div>
      </div>

      <div className="flex items-center gap-3 shrink-0">
        <div className="grid grid-cols-2 gap-6 border border-slate-200 rounded-2xl bg-white px-6 py-4 shadow-sm">
          <div className="text-center">
            <div className="text-[10px] font-semibold tracking-wide text-slate-400 uppercase">干体重</div>
            <div className="mt-2 text-2xl font-black text-blue-600">{patient.dryWeight}<span className="ml-1 text-xs text-slate-400">kg</span></div>
          </div>
          <div className="text-center">
            <div className="text-[10px] font-semibold tracking-wide text-slate-400 uppercase">治疗方案</div>
            <div className="mt-2 text-2xl font-black text-slate-800">{patient.treatmentPlan}</div>
          </div>
        </div>
        <button type="button" className="h-12 w-12 rounded-xl border border-slate-200 bg-white text-slate-500 flex items-center justify-center hover:text-blue-600 hover:border-blue-300">
          <Printer size={18} />
        </button>
        <button type="button" className="h-12 w-12 rounded-xl border border-slate-200 bg-white text-slate-500 flex items-center justify-center hover:text-blue-600 hover:border-blue-300">
          <FileOutput size={18} />
        </button>
      </div>
    </div>
  )
}
