import { Modal, Button, Select, DatePicker, Input, message } from 'antd'
import { useState, useEffect, useMemo } from 'react'
import { restApi, getErrorMessage } from '@/services/restClient'
import type { RestScheduleBed, RestSchedulePendingPatient, RestScheduleWard, RestShift } from '@/services/restClient'
import dayjs from 'dayjs'

const MODE_OPTIONS = ['HD', 'HDF', 'HF', 'HP', 'PE'].map((m) => ({ value: m, label: m }))

interface CreateScheduleModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
  initialDate?: string
  initialWardId?: number
  initialBedId?: number
  initialShiftId?: number
  initialPatientId?: number
  wards: RestScheduleWard[]
  beds: RestScheduleBed[]
  shifts: RestShift[]
  pendingPatients: RestSchedulePendingPatient[]
}

function pad(n: number) {
  return n < 10 ? `0${n}` : `${n}`
}

function toDateString(d: Date) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export default function CreateScheduleModal({
  open,
  onClose,
  onSuccess,
  initialDate,
  initialWardId,
  initialBedId,
  initialShiftId,
  initialPatientId,
  wards,
  beds,
  shifts,
  pendingPatients,
}: CreateScheduleModalProps) {
  const [loading, setLoading] = useState(false)
  const [patientId, setPatientId] = useState<number | undefined>(undefined)
  const [scheduleDate, setScheduleDate] = useState(dayjs())
  const [wardId, setWardId] = useState<number | undefined>(undefined)
  const [bedId, setBedId] = useState<number | undefined>(undefined)
  const [shiftId, setShiftId] = useState<number | undefined>(undefined)
  const [dialysisMode, setDialysisMode] = useState('HD')
  const [notes, setNotes] = useState('')

  // Reset form when modal opens
  useEffect(() => {
    if (open) {
      setPatientId(initialPatientId ?? undefined)
      setScheduleDate(initialDate ? dayjs(initialDate) : dayjs())
      setWardId(initialWardId ?? undefined)
      setBedId(initialBedId ?? undefined)
      setShiftId(initialShiftId ?? undefined)
      setDialysisMode('HD')
      setNotes('')
    }
  }, [open, initialDate, initialPatientId, initialWardId, initialBedId, initialShiftId])

  // Auto-fill from selected patient
  useEffect(() => {
    if (patientId) {
      const p = pendingPatients.find((x) => x.id === patientId)
      if (p) {
        setDialysisMode(p.dialysisMode || 'HD')
      }
    }
  }, [patientId, pendingPatients])

  // Auto-fill wardId from bedId
  useEffect(() => {
    if (bedId) {
      const bed = beds.find((b) => Number(b.id) === bedId)
      if (bed) {
        setWardId(Number(bed.wardId))
      }
    }
  }, [bedId, beds])

  // Filter beds by ward
  const filteredBeds = useMemo(() => {
    if (wardId) {
      return beds.filter((b) => Number(b.wardId) === wardId)
    }
    return beds
  }, [beds, wardId])

  // Clear bed if it doesn't belong to current ward
  useEffect(() => {
    if (bedId && wardId) {
      const bed = beds.find((b) => Number(b.id) === bedId)
      if (bed && Number(bed.wardId) !== wardId) {
        setBedId(undefined)
      }
    }
  }, [wardId, bedId, beds])

  const selectedPatient = pendingPatients.find((p) => p.id === patientId)

  const handleCreate = async () => {
    if (!patientId) {
      message.warning('请选择患者')
      return
    }
    if (!shiftId) {
      message.warning('请选择班次')
      return
    }

    setLoading(true)
    try {
      await restApi.createPatientShift({
        patientId,
        scheduleDate: scheduleDate.format('YYYY-MM-DD'),
        shiftId,
        bedId,
        wardId,
        dialysisMode: dialysisMode || 'HD',
        patientPlanId: selectedPatient?.patientPlanId ?? 0,
        shiftTiming: 20,
        status: 1,
        notes: notes || undefined,
      })
      message.success('排班已创建')
      onSuccess()
      onClose()
    } catch (error) {
      message.error(getErrorMessage(error))
    } finally {
      setLoading(false)
    }
  }

  const today = toDateString(new Date())
  const isPastDate = scheduleDate.format('YYYY-MM-DD') < today

  return (
    <Modal
      title="新建排班"
      open={open}
      onCancel={onClose}
      width={520}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="create" type="primary" loading={loading} disabled={isPastDate} onClick={handleCreate}>创建排班</Button>,
      ]}
    >
      <div className="space-y-4">
        {/* 患者选择 */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">患者 *</label>
          <Select
            value={patientId}
            onChange={(v) => setPatientId(v)}
            placeholder="选择待排班患者..."
            className="w-full"
            showSearch
            filterOption={(input, option) =>
              (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
            }
            options={pendingPatients.map((p) => ({
              value: p.id,
              label: `${p.name} (${p.dialysisMode || '--'} · ${p.oddWeekFrequency || 0}/${p.evenWeekFrequency || 0}次)`,
            }))}
          />
          {selectedPatient && (
            <div className="mt-1 flex gap-1.5 text-meta font-bold text-slate-400">
              <span>模式:{selectedPatient.dialysisMode || '--'}</span>
              <span>剩余:{selectedPatient.remainingTimes ?? '--'}次</span>
            </div>
          )}
        </div>

        {/* 日期 */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">排班日期 *</label>
          <DatePicker
            value={scheduleDate}
            onChange={(d) => d && setScheduleDate(d)}
            className="w-full"
            allowClear={false}
            disabledDate={(d) => d && d.isBefore(dayjs(), 'day')}
          />
          {isPastDate && (
            <div className="mt-1 text-meta font-bold text-red-500">历史日期不可创建排班</div>
          )}
        </div>

        {/* 病区 + 床位 */}
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-sm text-gray-600 mb-1">病区</label>
            <Select
              value={wardId}
              onChange={(v) => setWardId(v)}
              placeholder="选择病区"
              className="w-full"
              allowClear
              options={wards.map((w) => ({
                value: Number(w.id),
                label: `${w.name} (${w.bedCount ?? 0}床)`,
              }))}
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">床位</label>
            <Select
              value={bedId}
              onChange={(v) => setBedId(v)}
              placeholder="选择床位"
              className="w-full"
              allowClear
              showSearch
              filterOption={(input, option) =>
                (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
              }
              options={filteredBeds.map((b) => ({
                value: Number(b.id),
                label: `${b.name} (${b.wardName})`,
              }))}
            />
          </div>
        </div>

        {/* 班次 + 透析模式 */}
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-sm text-gray-600 mb-1">班次 *</label>
            <Select
              value={shiftId}
              onChange={(v) => setShiftId(v)}
              placeholder="选择班次"
              className="w-full"
              options={shifts.map((s) => ({
                value: s.id,
                label: `${s.name} (${s.startTime}-${s.endTime})`,
              }))}
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">透析模式</label>
            <Select
              value={dialysisMode}
              onChange={(v) => setDialysisMode(v)}
              className="w-full"
              options={MODE_OPTIONS}
            />
          </div>
        </div>

        {/* 备注 */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">备注</label>
          <Input
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="排班备注（可选）"
            maxLength={200}
          />
        </div>
      </div>
    </Modal>
  )
}
