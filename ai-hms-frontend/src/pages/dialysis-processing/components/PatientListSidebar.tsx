import { Search, Filter } from 'lucide-react'
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
  const [filter, setFilter] = useState<'all' | 'inDept'>('all')

  const filteredPatients = useMemo(() => {
    const trimmed = keyword.trim().toLowerCase()
    let list = patients
    if (!trimmed) {
      list = patients
    } else {
      list = patients.filter((item) =>
        [item.name, item.patientId, item.bedId, item.status].some((value) =>
          value.toLowerCase().includes(trimmed)
        )
      )
    }
    if (filter === 'inDept') {
      list = list.filter((item) => item.status === '透析中' || item.status === '候诊')
    }
    return list
  }, [keyword, patients, filter])

  return (
    <aside className={`w-60 h-full bg-white flex flex-col transition-opacity duration-200 ${isVisible ? 'opacity-100' : 'opacity-0'}`}>
      <div className="p-3 border-b border-slate-200 shrink-0 space-y-2">
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={() => setFilter('all')}
            className={`flex items-center gap-1 px-3 py-1 text-xs font-medium rounded-md transition-colors ${
              filter === 'all' ? 'bg-blue-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            }`}
          >
            <Filter size={12} />
            全部
          </button>
          <span className="text-xs text-slate-400">{patients.length}人</span>
        </div>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" size={15} />
          <input
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            placeholder="搜索姓名 / 床位 / 患者ID"
            className="w-full h-9 pl-8 pr-3 rounded-md border border-slate-200 bg-slate-50 text-sm outline-none focus:border-blue-500 focus:bg-white"
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
              className={`w-full text-left px-3 py-2 border-b border-slate-100 transition-colors ${
                active ? 'bg-blue-50 border-l-[3px] border-l-blue-600' : 'hover:bg-slate-50 border-l-[3px] border-l-transparent'
              }`}
            >
              <div className="flex items-center justify-between gap-2">
                <div className="min-w-0 flex-1">
                  <div className={`text-sm font-semibold truncate ${active ? 'text-blue-700' : 'text-slate-800'}`}>{patient.name}</div>
                  <div className="mt-0.5 text-xs text-slate-500">{patient.gender} / {patient.age}岁</div>
                </div>
                <span className="text-xs font-medium text-slate-500 bg-slate-100 rounded px-1.5 py-0.5 shrink-0">
                  {patient.status}
                </span>
              </div>
            </button>
          )
        })}
      </div>
    </aside>
  )
}
