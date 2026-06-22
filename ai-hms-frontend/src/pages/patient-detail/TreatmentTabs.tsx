import { useState } from 'react'
import { Segmented } from 'antd'
import { useTranslation } from 'react-i18next'
import type { Patient } from '@/types/original'
import type { TreatmentSubTab } from './types'
import { TreatmentPlanTab, SchemeOrderTab, VascularTab, AdverseTab, MedicationTab, DryWeightTab } from './tabs'

interface TreatmentTabsProps {
  patient: Patient
  defaultSub?: TreatmentSubTab
  onSubChange?: (sub: TreatmentSubTab) => void
}

export default function TreatmentTabs({ patient, defaultSub = 'plan', onSubChange }: TreatmentTabsProps) {
  const { t } = useTranslation('patient')
  const [subTab, setSubTab] = useState<TreatmentSubTab>(defaultSub)

  const handleChange = (val: string) => {
    const next = val as TreatmentSubTab
    setSubTab(next)
    onSubChange?.(next)
  }

  const options = [
    { label: t('tab.sub.plan'), value: 'plan' },
    { label: t('tab.sub.schemeOrder'), value: 'schemeOrder' },
    { label: t('tab.sub.vascular'), value: 'vascular' },
    { label: t('tab.sub.adverse'), value: 'adverse' },
    { label: t('tab.sub.medication'), value: 'medication' },
    { label: t('tab.sub.dryWeight'), value: 'dryWeight' },
  ]

  return (
    <div className="space-y-4">
      <Segmented options={options} value={subTab} onChange={handleChange} />
      <div>
        {subTab === 'plan' && <TreatmentPlanTab patientId={patient.id} patientName={patient.name} />}
        {subTab === 'schemeOrder' && <SchemeOrderTab patient={patient} />}
        {subTab === 'vascular' && <VascularTab patient={patient} />}
        {subTab === 'adverse' && <AdverseTab patient={patient} />}
        {subTab === 'medication' && <MedicationTab patient={patient} />}
        {subTab === 'dryWeight' && <DryWeightTab patient={patient} />}
      </div>
    </div>
  )
}
