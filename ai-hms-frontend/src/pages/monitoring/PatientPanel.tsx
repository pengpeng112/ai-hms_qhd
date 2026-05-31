import {
  ComprehensiveMonitorModal,
  PrescriptionEditModal,
  OrderListModal,
  SummaryModal,
} from '../Monitoring'
import type { MonitorDevice } from '@/types/original'
import type { ModalType } from './types'

interface PatientPanelProps {
  activeModal: ModalType
  selectedDevice: MonitorDevice | null
  onClose: () => void
}

export default function PatientPanel({ activeModal, selectedDevice, onClose }: PatientPanelProps) {
  if (!activeModal || !selectedDevice) return null

  switch (activeModal) {
    case 'COMPREHENSIVE':
      return <ComprehensiveMonitorModal device={selectedDevice} onClose={onClose} />
    case 'PRESCRIPTION':
      return <PrescriptionEditModal device={selectedDevice} onClose={onClose} />
    case 'ORDERS':
      return <OrderListModal device={selectedDevice} onClose={onClose} />
    case 'SUMMARY':
      return <SummaryModal device={selectedDevice} onClose={onClose} />
    default:
      return null
  }
}
