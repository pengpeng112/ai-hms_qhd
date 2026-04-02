// 抗凝剂区域组件

import { memo, useMemo } from 'react'
import { Syringe } from 'lucide-react'
import { FormSection } from './FormSection'
import { SelectField } from './SelectField'
import { NumericInputField } from './NumericInputField'
import { InputField } from './InputField'
import { calcInjectionVolume, calcTotalDose } from '@/utils/treatmentPlanCalculations'
import type { AnticoagulantValues, OpenNumericPad } from './types'

interface AnticoagulantSectionProps {
  values: AnticoagulantValues
  onChange: (updater: (prev: AnticoagulantValues) => AnticoagulantValues) => void
  drugOptions: Array<{ value: string; label: string }>
  specMap: Record<string, string>
  openNumericPad: OpenNumericPad
}

const isNoHeparinOption = (value: string) => value === '相对无肝素' || value === '绝对无肝素'

const extractDoseUnit = (concentration?: string) => {
  const trimmed = concentration?.trim()
  if (!trimmed) return ''
  const match = trimmed.match(/[a-zA-Zμµ]+/g)
  if (!match || match.length === 0) return ''
  return match[0].replace('µ', 'u')
}

export const AnticoagulantSection = memo(function AnticoagulantSection({
  values,
  onChange,
  drugOptions,
  specMap,
  openNumericPad,
}: AnticoagulantSectionProps) {
  // 互斥过滤
  const initialOptions = useMemo(
    () =>
      isNoHeparinOption(values.maintenanceDrug)
        ? drugOptions.filter((opt) => opt.value !== values.maintenanceDrug)
        : drugOptions,
    [drugOptions, values.maintenanceDrug]
  )

  const maintenanceOptions = useMemo(
    () =>
      isNoHeparinOption(values.initialDrug)
        ? drugOptions.filter((opt) => opt.value !== values.initialDrug)
        : drugOptions,
    [drugOptions, values.initialDrug]
  )

  // 动态单位
  const initialDoseUnit = isNoHeparinOption(values.initialDrug)
    ? ''
    : extractDoseUnit(specMap[values.initialDrug] || '')
  const infusionRateUnit = isNoHeparinOption(values.maintenanceDrug)
    ? ''
    : extractDoseUnit(specMap[values.maintenanceDrug] || '')
  const infusionRateSuffix = infusionRateUnit ? `${infusionRateUnit}/h` : ''
  const maintenanceDoseUnit = isNoHeparinOption(values.maintenanceDrug)
    ? ''
    : extractDoseUnit(specMap[values.maintenanceDrug] || '')
  const totalDoseUnit = isNoHeparinOption(values.initialDrug)
    ? ''
    : extractDoseUnit(specMap[values.initialDrug] || '')

  return (
    <FormSection title="抗凝剂" icon={Syringe}>
      {/* 首剂名称 */}
      <SelectField
        label="首剂名称"
        options={initialOptions}
        required
        value={values.initialDrug}
        onChange={(e) => {
          const nextValue = e.target.value
          onChange((prev) => {
            const isInitialNoHeparin = isNoHeparinOption(nextValue)
            const isMaintenanceNoHeparin = isNoHeparinOption(prev.maintenanceDrug)
            const nextInitialDose = isInitialNoHeparin ? '' : prev.initialDose
            const shouldClearMaintenance = isInitialNoHeparin && prev.maintenanceDrug === nextValue
            const nextMaintenanceDrug = shouldClearMaintenance ? '' : prev.maintenanceDrug
            const nextInfusionRate = shouldClearMaintenance ? '' : prev.infusionRate
            const nextInfusionTime = shouldClearMaintenance ? '' : prev.infusionTime
            const nextMaintenanceDose =
              shouldClearMaintenance || isMaintenanceNoHeparin
                ? ''
                : calcInjectionVolume(nextInfusionRate, nextInfusionTime)
            const nextTotalDose =
              shouldClearMaintenance || isMaintenanceNoHeparin
                ? ''
                : calcTotalDose(nextMaintenanceDose, nextInitialDose)
            return {
              ...prev,
              initialDrug: nextValue,
              initialDose: nextInitialDose,
              maintenanceDrug: nextMaintenanceDrug,
              infusionRate: nextInfusionRate,
              infusionTime: nextInfusionTime,
              maintenanceDose: nextMaintenanceDose,
              totalDose: nextTotalDose,
            }
          })
        }}
      />

      {/* 首剂量 */}
      <NumericInputField
        label="首剂量"
        suffix={initialDoseUnit}
        placeholder="2500"
        required={!isNoHeparinOption(values.initialDrug)}
        disabled={isNoHeparinOption(values.initialDrug)}
        value={values.initialDose}
        openNumericPad={openNumericPad}
        onConfirm={(value) => {
          onChange((prev) => {
            const isMaintenanceNoHeparin = isNoHeparinOption(prev.maintenanceDrug)
            const nextTotalDose = isMaintenanceNoHeparin
              ? ''
              : calcTotalDose(prev.maintenanceDose, value)
            return {
              ...prev,
              initialDose: value,
              totalDose: nextTotalDose,
            }
          })
        }}
      />

      {/* 维持剂名称 */}
      <SelectField
        label="维持剂名称"
        options={maintenanceOptions}
        value={values.maintenanceDrug}
        onChange={(e) => {
          const nextValue = e.target.value
          onChange((prev) => {
            if (isNoHeparinOption(nextValue) && prev.initialDrug === nextValue) {
              return {
                ...prev,
                initialDrug: '',
                initialDose: '',
                maintenanceDrug: nextValue,
                infusionRate: '',
                infusionTime: '',
                maintenanceDose: '',
                totalDose: '',
              }
            }
            if (isNoHeparinOption(nextValue)) {
              return {
                ...prev,
                maintenanceDrug: nextValue,
                infusionRate: '',
                infusionTime: '',
                maintenanceDose: '',
                totalDose: '',
              }
            }
            const nextMaintenanceDose = calcInjectionVolume(prev.infusionRate, prev.infusionTime)
            const nextTotalDose = calcTotalDose(nextMaintenanceDose, prev.initialDose)
            return {
              ...prev,
              maintenanceDrug: nextValue,
              maintenanceDose: nextMaintenanceDose,
              totalDose: nextTotalDose,
            }
          })
        }}
      />

      {/* 注入速率 */}
      <NumericInputField
        label="注入速率"
        suffix={infusionRateSuffix}
        placeholder="5.0"
        disabled={isNoHeparinOption(values.maintenanceDrug)}
        value={values.infusionRate}
        openNumericPad={openNumericPad}
        onConfirm={(value) => {
          onChange((prev) => {
            if (isNoHeparinOption(prev.maintenanceDrug)) {
              return { ...prev, infusionRate: '', maintenanceDose: '', totalDose: '' }
            }
            const nextMaintenanceDose = calcInjectionVolume(value, prev.infusionTime)
            const nextTotalDose = calcTotalDose(nextMaintenanceDose, prev.initialDose)
            return {
              ...prev,
              infusionRate: value,
              maintenanceDose: nextMaintenanceDose,
              totalDose: nextTotalDose,
            }
          })
        }}
      />

      {/* 注入时间 */}
      <NumericInputField
        label="注入时间"
        suffix="h"
        placeholder="3.5"
        disabled={isNoHeparinOption(values.maintenanceDrug)}
        value={values.infusionTime}
        openNumericPad={openNumericPad}
        onConfirm={(value) => {
          onChange((prev) => {
            if (isNoHeparinOption(prev.maintenanceDrug)) {
              return { ...prev, infusionTime: '', maintenanceDose: '', totalDose: '' }
            }
            const nextMaintenanceDose = calcInjectionVolume(prev.infusionRate, value)
            const nextTotalDose = calcTotalDose(nextMaintenanceDose, prev.initialDose)
            return {
              ...prev,
              infusionTime: value,
              maintenanceDose: nextMaintenanceDose,
              totalDose: nextTotalDose,
            }
          })
        }}
      />

      {/* 维持量（只读） */}
      <InputField
        label="维持量"
        suffix={maintenanceDoseUnit}
        placeholder="500"
        readOnly
        value={values.maintenanceDose}
      />

      {/* 总剂量（只读） */}
      <InputField
        label="总剂量"
        suffix={totalDoseUnit}
        placeholder="3000"
        readOnly
        value={values.totalDose}
      />
    </FormSection>
  )
})
