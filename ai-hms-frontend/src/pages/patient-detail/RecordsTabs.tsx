import { useState } from 'react'
import { Segmented } from 'antd'
import { useTranslation } from 'react-i18next'
import type { Patient } from '@/types/original'
import type { RecordsSubTab } from './types'
import { BasicInfoTab, LabsExamsTab, MonthlySummaryTab, MedicalRecordTab } from './tabs'

interface RecordsTabsProps {
  patient: Patient
  defaultSub?: RecordsSubTab
  onSubChange?: (sub: RecordsSubTab) => void
}

export default function RecordsTabs({ patient, defaultSub = 'basicInfo', onSubChange }: RecordsTabsProps) {
  const { t } = useTranslation('patient')
  const [subTab, setSubTab] = useState<RecordsSubTab>(defaultSub)

  const handleChange = (val: string) => {
    const next = val as RecordsSubTab
    setSubTab(next)
    onSubChange?.(next)
  }

  const options = [
    { label: t('tab.sub.basicInfo'), value: 'basicInfo' },
    { label: t('tab.sub.labs'), value: 'labs' },
    { label: t('tab.sub.monthly'), value: 'monthly' },
    { label: t('tab.sub.medicalRecord'), value: 'medicalRecord' },
  ]

  return (
    <div className="space-y-4">
      <Segmented options={options} value={subTab} onChange={handleChange} />
      <div>
        {subTab === 'basicInfo' && <BasicInfoTab patient={patient} />}
        {subTab === 'labs' && <LabsExamsTab patient={patient} />}
        {subTab === 'monthly' && <MonthlySummaryTab patient={patient} />}
        {subTab === 'medicalRecord' && <MedicalRecordTab patient={patient} />}
      </div>
    </div>
  )
}
