import { useState, useEffect, useCallback } from 'react'
import { Button, message, Tag, Card } from 'antd'
import { Syringe, ShieldCheck, CheckCircle2, Clock } from 'lucide-react'
import { medicationApi, type MedicationAdmin } from '@/services/medicationApi'
import dayjs from 'dayjs'

interface Props {
  treatmentId: number
  patientId: number
}

export default function MedicationAdminPanel({ treatmentId, patientId }: Props) {
  const [records, setRecords] = useState<MedicationAdmin[]>([])
  const [loading, setLoading] = useState(false)

  const fetchRecords = useCallback(async () => {
    setLoading(true)
    try {
      const rows = await medicationApi.list({ treatmentId, patientId })
      setRecords(rows)
    } catch { /* ignore */ }
    setLoading(false)
  }, [treatmentId, patientId])

  useEffect(() => { fetchRecords() }, [fetchRecords])

  async function handleSecondCheck(maId: string) {
    try {
      await medicationApi.secondCheck(maId)
      message.success('双人核对完成')
      fetchRecords()
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message || '核对失败')
    }
  }

  // Build available orders from records that need verification
  const unverified = records.filter((r) => r.status === 'recorded')
  const verified = records.filter((r) => r.status === 'verified')

  return (
    <div className="space-y-3 mt-4">
      <div className="flex items-center gap-2 text-[13px] font-bold text-slate-700">
        <Syringe size={16} className="text-blue-500" />
        长嘱给药执行
        {unverified.length > 0 && (
          <Tag color="orange" className="text-[11px]">{unverified.length} 待核对</Tag>
        )}
        {verified.length > 0 && (
          <Tag color="green" className="text-[11px]">{verified.length} 已核</Tag>
        )}
      </div>

      {unverified.map((r) => (
        <Card key={r.id} size="small" bordered className="rounded-xl border-orange-200 bg-orange-50">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Clock size={14} className="text-orange-500" />
              <span className="font-bold text-[13px]">{r.drugName}</span>
              {r.dose && <span className="text-[12px] text-slate-500">{r.dose} {r.route}</span>}
              <span className="text-[11px] text-slate-400">{dayjs(r.administeredAt).format('HH:mm')}</span>
            </div>
            <Button size="small" type="primary" icon={<ShieldCheck size={14} />}
              onClick={() => handleSecondCheck(r.id)}>
              双人核对
            </Button>
          </div>
          <div className="text-[11px] text-slate-400 mt-1">执行人：{r.administeredName || r.administeredBy}</div>
        </Card>
      ))}

      {verified.map((r) => (
        <Card key={r.id} size="small" bordered className="rounded-xl border-emerald-200 bg-emerald-50">
          <div className="flex items-center gap-2">
            <CheckCircle2 size={14} className="text-emerald-500" />
            <span className="font-bold text-[13px]">{r.drugName}</span>
            {r.dose && <span className="text-[12px] text-slate-500">{r.dose}</span>}
            <span className="text-[11px] text-slate-400 ml-auto">
              {r.administeredName || r.administeredBy} · {r.secondCheckName || r.secondCheckBy} 核对
            </span>
          </div>
        </Card>
      ))}

      {!loading && records.length === 0 && (
        <div className="text-[12px] text-slate-400 py-2">本次治疗暂无给药记录</div>
      )}
    </div>
  )
}
