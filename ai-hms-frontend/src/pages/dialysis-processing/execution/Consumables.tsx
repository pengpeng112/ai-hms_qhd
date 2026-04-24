import type { Patient } from '../types'

export default function Consumables({ patient }: { patient: Patient }) {
  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="text-lg font-black text-slate-800">耗材核对</div>
      <div className="mt-3 text-sm text-slate-600">患者：{patient.name}</div>
    </div>
  )
}
