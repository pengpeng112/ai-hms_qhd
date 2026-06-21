import { useState } from 'react'
import { ScanLine } from 'lucide-react'
import ActrPanel from './ActrPanel'

interface Props {
  patientId: string
  prescriptionId?: string
}

export default function ActrButton({ patientId, prescriptionId }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <button type="button" onClick={(e) => { e.stopPropagation(); setOpen(true) }}
        className="flex items-center gap-1 shrink-0 px-2.5 py-1 rounded-lg bg-slate-600 text-white font-bold hover:bg-slate-700 transition-all">
        <ScanLine size={12} />
        影像
      </button>
      <ActrPanel open={open} onClose={() => setOpen(false)} patientId={patientId} prescriptionId={prescriptionId} />
    </>
  )
}
