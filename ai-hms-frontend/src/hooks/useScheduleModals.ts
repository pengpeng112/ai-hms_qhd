import { useState, type Dispatch, type SetStateAction } from 'react'
import {
  type RestScheduleBed,
  type RestScheduleWeekShift,
  type RestShift,
  type RestTreatment,
  type RestPatientShift,
} from '@/services'

interface ModalState {
  open: boolean
  bed: RestScheduleBed | null
  date: string
  shift: RestShift | null
  existing: RestScheduleWeekShift | null
}

interface ActionMenuState {
  visible: boolean
  x: number
  y: number
  item: RestScheduleWeekShift | null
}

interface MoveModalState {
  open: boolean
  item: RestScheduleWeekShift | null
  targetBedId: number | undefined
}

interface TreatModalState {
  open: boolean
  patientId: number | undefined
  patientName: string
}

interface HistoryModalState {
  open: boolean
  patientId: number | undefined
  patientName: string
}

export interface UseScheduleModalsReturn {
  // 排班弹窗
  modal: ModalState
  setModal: Dispatch<SetStateAction<ModalState>>
  selPatient: number | undefined
  setSelPatient: Dispatch<SetStateAction<number | undefined>>

  // 右键菜单
  actionMenu: ActionMenuState
  setActionMenu: Dispatch<SetStateAction<ActionMenuState>>

  // 应用模板
  applyTemplateOpen: boolean
  setApplyTemplateOpen: Dispatch<SetStateAction<boolean>>

  // 换床弹窗
  moveModal: MoveModalState
  setMoveModal: Dispatch<SetStateAction<MoveModalState>>
  moveLoading: boolean
  setMoveLoading: Dispatch<SetStateAction<boolean>>

  // 治疗记录弹窗
  treatModal: TreatModalState
  setTreatModal: Dispatch<SetStateAction<TreatModalState>>
  treatments: RestTreatment[]
  setTreatments: Dispatch<SetStateAction<RestTreatment[]>>
  treatLoading: boolean
  setTreatLoading: Dispatch<SetStateAction<boolean>>

  // 换床历史弹窗
  historyModal: HistoryModalState
  setHistoryModal: Dispatch<SetStateAction<HistoryModalState>>
  shiftHistory: RestPatientShift[]
  setShiftHistory: Dispatch<SetStateAction<RestPatientShift[]>>
  historyLoading: boolean
  setHistoryLoading: Dispatch<SetStateAction<boolean>>
}

export function useScheduleModals(): UseScheduleModalsReturn {
  // 排班弹窗
  const [modal, setModal] = useState<ModalState>({
    open: false,
    bed: null,
    date: '',
    shift: null,
    existing: null,
  })
  const [selPatient, setSelPatient] = useState<number | undefined>()

  // 右键菜单
  const [actionMenu, setActionMenu] = useState<ActionMenuState>({
    visible: false,
    x: 0,
    y: 0,
    item: null,
  })

  // 应用模板弹窗
  const [applyTemplateOpen, setApplyTemplateOpen] = useState(false)

  // 换床弹窗
  const [moveModal, setMoveModal] = useState<MoveModalState>({
    open: false,
    item: null,
    targetBedId: undefined,
  })
  const [moveLoading, setMoveLoading] = useState(false)

  // 治疗记录弹窗
  const [treatModal, setTreatModal] = useState<TreatModalState>({
    open: false,
    patientId: undefined,
    patientName: '',
  })
  const [treatments, setTreatments] = useState<RestTreatment[]>([])
  const [treatLoading, setTreatLoading] = useState(false)

  // 换床历史弹窗
  const [historyModal, setHistoryModal] = useState<HistoryModalState>({
    open: false,
    patientId: undefined,
    patientName: '',
  })
  const [shiftHistory, setShiftHistory] = useState<RestPatientShift[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)

  return {
    modal,
    setModal,
    selPatient,
    setSelPatient,
    actionMenu,
    setActionMenu,
    applyTemplateOpen,
    setApplyTemplateOpen,
    moveModal,
    setMoveModal,
    moveLoading,
    setMoveLoading,
    treatModal,
    setTreatModal,
    treatments,
    setTreatments,
    treatLoading,
    setTreatLoading,
    historyModal,
    setHistoryModal,
    shiftHistory,
    setShiftHistory,
    historyLoading,
    setHistoryLoading,
  }
}