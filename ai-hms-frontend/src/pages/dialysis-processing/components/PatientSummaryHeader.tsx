import { FileOutput, Printer } from 'lucide-react'
import type { Patient } from '../types'

interface Props {
  patient: Patient
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs font-extrabold text-slate-500">{label}</div>
      <div className="mt-0.5 text-[15px] font-black text-slate-950">{value || '--'}</div>
    </div>
  )
}

export default function PatientSummaryHeader({ patient }: Props) {
  return (
    <div className="flex min-h-[72px] items-center gap-4 rounded-2xl border border-blue-100 bg-white px-4 py-2 shadow-sm">
      <div className="min-w-[140px]">
        <div className="flex items-center gap-2">
          <h2 className="m-0 text-2xl font-black text-slate-950">{patient.name}</h2>
          <span className="inline-flex items-center rounded-lg bg-blue-600 px-2.5 py-0.5 text-[13px] font-black text-white">
            {patient.bedId || '未排床'}
          </span>
        </div>
      </div>

      <div className="grid flex-1 grid-cols-2 gap-3 xl:grid-cols-4">
        <InfoItem label="患者ID" value={patient.patientId} />
        <InfoItem label="性别 / 年龄" value={`${patient.gender || '--'} / ${patient.age || '--'}岁`} />
        <InfoItem label="费用类别" value={patient.costType || '--'} />
        <InfoItem label="透龄" value={patient.dialysisAge || '待补充'} />
      </div>

      <div className="ml-auto flex items-center gap-2 shrink-0">
        <div className="flex h-[54px] w-20 flex-col items-center justify-center rounded-[13px] border border-blue-100 bg-blue-50/40">
          <span className="text-xs font-extrabold text-slate-500">干体重</span>
          <b className="mt-1 text-lg font-black text-blue-600">{patient.dryWeight}<span className="ml-0.5 text-xs text-slate-400">kg</span></b>
        </div>
        <div className="flex h-[54px] w-20 flex-col items-center justify-center rounded-[13px] border border-slate-200 bg-slate-50">
          <span className="text-xs font-extrabold text-slate-500">治疗方案</span>
          <b className="mt-1 text-lg font-black text-slate-950">{patient.treatmentPlan || '--'}</b>
        </div>
        <button type="button" className="h-10 w-10 rounded-lg border border-slate-200 bg-white text-slate-500 flex items-center justify-center hover:text-blue-600 hover:border-blue-300">
          <Printer size={16} />
        </button>
        <button type="button" className="h-10 w-10 rounded-lg border border-slate-200 bg-white text-slate-500 flex items-center justify-center hover:text-blue-600 hover:border-blue-300">
          <FileOutput size={16} />
        </button>
      </div>
    </div>
  )
}
