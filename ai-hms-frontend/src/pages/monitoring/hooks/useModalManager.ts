import { useState } from 'react'
import type { MonitorDevice } from '@/types/original'
import type { ModalType } from '../types'

export function useModalManager() {
  const [activeModal, setActiveModal] = useState<ModalType>(null)
  const [selectedDevice, setSelectedDevice] = useState<MonitorDevice | null>(null)

  const openModal = (device: MonitorDevice, type: ModalType) => {
    setSelectedDevice(device)
    setActiveModal(type)
  }

  const closeModal = () => {
    setActiveModal(null)
    setSelectedDevice(null)
  }

  return { activeModal, selectedDevice, openModal, closeModal }
}
