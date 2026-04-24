import { Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import type { Patient } from '../types'

interface Props {
  patients: Patient[]
  selectedId: string
  onSelect: (patient: Patient) => void
  isVisible: boolean
}

export default function PatientListSidebar({ patients, selectedId, onSelect, isVisible }: Props) {
  const [keyword, setKeyword] = useState('')

  const filteredPatients = useMemo(() => {
    const trimmed = keyword.trim().toLowerCase()
    if (!trimmed) return patients
    return patients.filter((item) =>
      [item.name, item.patientId, item.bedId, item.status].some((value) =>
        value.toLowerCase().includes(trimmed)
      )
    )
  }, [keyword, patients])

  return (
    <aside className={`w-72 h-full bg-white flex flex-col transition-opacity duration-200 ${isVisible ? 'opacity-100' : 'opacity-0'}`}>
      <div className="p-4 border-b border-slate-200 shrink-0">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={16} />
          <input
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            placeholder="搜索姓名 / 床位 / 患者ID"
            className="w-full h-10 pl-9 pr-3 rounded-xl border border-slate-200 bg-slate-50 text-sm outline-none focus:border-blue-500 focus:bg-white"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto">
        {filteredPatients.map((patient) => {
          const active = patient.id === selectedId
          return (
            <button
              key={patient.id}
              type="button"
              onClick={() => onSelect(patient)}
              className={`w-full text-left p-4 border-b border-slate-100 transition-colors ${
                active ? 'bg-blue-50 border-l-4 border-l-blue-600' : 'hover:bg-slate-50'
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className={`text-sm font-bold truncate ${active ? 'text-blue-700' : 'text-slate-800'}`}>{patient.name}</div>
                  <div className="mt-1 text-xs text-slate-500 truncate">{patient.patientId}</div>
                </div>
                <span className="text-[11px] font-bold text-slate-600 bg-white border border-slate-200 rounded-md px-2 py-1 shrink-0">
                  {patient.bedId}
                </span>
              </div>
              <div className="mt-3 flex items-center justify-between text-xs">
                <span className="text-slate-500">{patient.gender} / {patient.age}岁</span>
                <span className={`font-semibold ${patient.status === '透析中' ? 'text-blue-600' : 'text-slate-400'}`}>{patient.status}</span>
              </div>
            </button>
          )
        })}
      </div>
    </aside>
  )
}
