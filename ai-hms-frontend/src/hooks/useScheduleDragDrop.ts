import { useState, type Dispatch, type SetStateAction } from 'react'
import { message } from 'antd'
import {
  restApi,
  getErrorMessage,
  type RestScheduleBed,
  type RestScheduleWeekShift,
  type RestShift,
} from '@/services'

function isDateLocked(dt: string) {
  return dt < toDateString(new Date())
}

function pad(n: number) {
  return n < 10 ? `0${n}` : `${n}`
}

function toDateString(d: Date) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export interface UseScheduleDragDropReturn {
  dragItem: RestScheduleWeekShift | null
  setDragItem: Dispatch<SetStateAction<RestScheduleWeekShift | null>>
  dragOverKey: string | null
  setDragOverKey: Dispatch<SetStateAction<string | null>>
  onDragStart: (e: React.DragEvent, item: RestScheduleWeekShift) => void
  onDragOver: (e: React.DragEvent, key: string) => void
  onDropOnEmpty: (
    e: React.DragEvent,
    bed: RestScheduleBed,
    dt: string,
    shift: RestShift,
    loadWeek: () => Promise<void>
  ) => Promise<void>
  onDropOnOccupied: (
    e: React.DragEvent,
    targetItem: RestScheduleWeekShift,
    loadWeek: () => Promise<void>
  ) => Promise<void>
  onDragEnd: () => void
  onDragLeave: () => void
}

export function useScheduleDragDrop(): UseScheduleDragDropReturn {
  const [dragItem, setDragItem] = useState<RestScheduleWeekShift | null>(null)
  const [dragOverKey, setDragOverKey] = useState<string | null>(null)

  const onDragStart = (e: React.DragEvent, item: RestScheduleWeekShift) => {
    setDragItem(item)
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(item.id))
  }

  const onDragEnd = () => {
    setDragItem(null)
    setDragOverKey(null)
  }

  const onDragOver = (e: React.DragEvent, cellKey: string) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverKey(cellKey)
  }

  const onDragLeave = () => setDragOverKey(null)

  const onDropOnEmpty = async (
    e: React.DragEvent,
    bed: RestScheduleBed,
    _dt: string,
    _shift: RestShift,
    loadWeek: () => Promise<void>
  ) => {
    e.preventDefault()
    setDragOverKey(null)
    if (!dragItem) return
    if (isDateLocked(_dt)) {
      message.warning('历史排班不可修改')
      setDragItem(null)
      return
    }
    if (isDateLocked(dragItem.treatmentTime.split('T')[0])) {
      message.warning('历史排班不可修改')
      setDragItem(null)
      return
    }
    const targetBedId = Number(bed.id)
    if (targetBedId === dragItem.bedId) return
    try {
      await restApi.movePatientShift(dragItem.id, {
        bedId: targetBedId,
        wardId: Number(bed.wardId),
      })
      message.success(`${dragItem.patientName} 已换床至 ${bed.name}`)
      await loadWeek()
    } catch (err) {
      message.error(getErrorMessage(err))
    } finally {
      setDragItem(null)
    }
  }

  const onDropOnOccupied = async (
    e: React.DragEvent,
    targetItem: RestScheduleWeekShift,
    loadWeek: () => Promise<void>
  ) => {
    e.preventDefault()
    setDragOverKey(null)
    if (!dragItem || dragItem.id === targetItem.id) {
      setDragItem(null)
      return
    }
    if (isDateLocked(targetItem.treatmentTime.split('T')[0])) {
      message.warning('历史排班不可修改')
      setDragItem(null)
      return
    }
    if (isDateLocked(dragItem.treatmentTime.split('T')[0])) {
      message.warning('历史排班不可修改')
      setDragItem(null)
      return
    }
    try {
      await restApi.swapPatientShifts(dragItem.id, targetItem.id)
      message.success(`${dragItem.patientName} 与 ${targetItem.patientName} 已对调`)
      await loadWeek()
    } catch (err) {
      message.error(getErrorMessage(err))
    } finally {
      setDragItem(null)
    }
  }

  return {
    dragItem,
    setDragItem,
    dragOverKey,
    setDragOverKey,
    onDragStart,
    onDragOver,
    onDropOnEmpty,
    onDropOnOccupied,
    onDragEnd,
    onDragLeave,
  }
}