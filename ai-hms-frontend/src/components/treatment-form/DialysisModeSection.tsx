// 透析模式区域组件

import { memo } from 'react'
import { Activity } from 'lucide-react'
import { FormSection } from './FormSection'
import { SelectField } from './SelectField'
import { NumericInputField } from './NumericInputField'
import { InputField } from './InputField'
import { calcSubstituteVolume, calcDialysateVolume } from '@/utils/treatmentPlanCalculations'
import type { DialysisModeValues, DictOptions, OpenNumericPad } from './types'

interface DialysisModeSectionProps {
  values: DialysisModeValues
  onChange: (updater: (prev: DialysisModeValues) => DialysisModeValues) => void
  dictOptions: DictOptions
  dictTypeKey: string
  openNumericPad: OpenNumericPad
  /** 透析液流速值，用于时间变化时联动计算透析液量 */
  dialysateFlow?: string
  /** 透析液量变化回调 */
  onDialysateVolumeChange?: (volume: string) => void
  extraContent?: React.ReactNode
}

const isSubstituteMode = (method: string) => method === 'HF' || method === 'HDF'

export const DialysisModeSection = memo(function DialysisModeSection({
  values,
  onChange,
  dictOptions,
  dictTypeKey,
  openNumericPad,
  dialysateFlow,
  onDialysateVolumeChange,
  extraContent,
}: DialysisModeSectionProps) {
  return (
    <FormSection title="透析模式" icon={Activity}>
      <SelectField
        label="透析方法"
        options={dictOptions[dictTypeKey] || []}
        required
        value={values.method}
        onChange={(e) => {
          const nextMethod = e.target.value
          onChange((prev) => {
            const next = { ...prev, method: nextMethod }
            if (isSubstituteMode(nextMethod)) {
              return {
                ...next,
                substituteVolume: calcSubstituteVolume(next.duration, next.substituteFlow),
              }
            }
            return next
          })
        }}
      />
      <NumericInputField
        label="透析时间"
        suffix="h"
        placeholder="4.0"
        required
        value={values.duration}
        openNumericPad={openNumericPad}
        onConfirm={(value) => {
          onChange((prev) => {
            const next = { ...prev, duration: value }
            const nextDialysateVolume = calcDialysateVolume(value, dialysateFlow || '')
            if (onDialysateVolumeChange) {
              onDialysateVolumeChange(nextDialysateVolume)
            }
            if (isSubstituteMode(next.method)) {
              return {
                ...next,
                substituteVolume: calcSubstituteVolume(value, next.substituteFlow),
              }
            }
            return next
          })
        }}
      />
      <NumericInputField
        label="标准血流量"
        suffix="ml/min"
        placeholder="200"
        required
        value={values.bloodFlow}
        openNumericPad={openNumericPad}
        onConfirm={(value) => {
          onChange((prev) => ({ ...prev, bloodFlow: value }))
        }}
      />
      {isSubstituteMode(values.method) && (
        <>
          <SelectField
            label="置换液输入方式"
            options={['前置换', '后置换']}
            value={values.substituteMode}
            onChange={(e) =>
              onChange((prev) => ({ ...prev, substituteMode: e.target.value }))
            }
          />
          <NumericInputField
            label="置换液流速"
            suffix="ml/min"
            placeholder="0"
            value={values.substituteFlow}
            openNumericPad={openNumericPad}
            onConfirm={(value) => {
              onChange((prev) => ({
                ...prev,
                substituteFlow: value,
                substituteVolume: calcSubstituteVolume(prev.duration, value),
              }))
            }}
          />
          <NumericInputField
            label="置换液总量"
            suffix="L"
            placeholder="0"
            value={values.substituteVolume}
            openNumericPad={openNumericPad}
            onConfirm={(value) => {
              onChange((prev) => ({ ...prev, substituteVolume: value }))
            }}
          />
        </>
      )}
      <div className="lg:col-span-1">
        <InputField
          label="备注"
          placeholder="请输入备注"
          value={values.notes}
          onChange={(e) =>
            onChange((prev) => ({ ...prev, notes: e.target.value }))
          }
        />
      </div>
      {extraContent}
    </FormSection>
  )
})
