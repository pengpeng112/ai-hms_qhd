// 透析液参数区域组件

import { memo } from 'react'
import { Sliders } from 'lucide-react'
import { FormSection } from './FormSection'
import { SelectField } from './SelectField'
import { NumericInputField } from './NumericInputField'
import { InputField } from './InputField'
import { calcDialysateVolume } from '@/utils/treatmentPlanCalculations'
import type { DialysisParamsValues, DictOptions, OpenNumericPad } from './types'

interface DialysisParamsSectionProps {
  values: DialysisParamsValues
  onChange: (updater: (prev: DialysisParamsValues) => DialysisParamsValues) => void
  duration: string
  dictOptions: DictOptions
  dictTypeKeys: {
    dialysateType: string
    dialysateGroup: string
    dialysateFlow: string
    glucose: string
  }
  openNumericPad: OpenNumericPad
}

const parseDialysateGroupIons = (groupValue: string) => {
  const group = (groupValue || '').trim()
  if (!group) return { k: '', ca: '' }
  const kMatch = group.match(/K\s*([0-9]+(?:\.[0-9]+)?)/i)
  const caMatch = group.match(/Ca\s*([0-9]+(?:\.[0-9]+)?)/i)
  return {
    k: kMatch ? kMatch[1] : '',
    ca: caMatch ? caMatch[1] : '',
  }
}

export const DialysisParamsSection = memo(function DialysisParamsSection({
  values,
  onChange,
  duration,
  dictOptions,
  dictTypeKeys,
  openNumericPad,
}: DialysisParamsSectionProps) {
  return (
    <FormSection title="透析液参数" icon={Sliders}>
      <SelectField
        label="透析液分类"
        options={dictOptions[dictTypeKeys.dialysateType] || []}
        required
        value={values.dialysateType}
        onChange={(e) =>
          onChange((prev) => ({ ...prev, dialysateType: e.target.value }))
        }
      />
      <SelectField
        label="透析液分组"
        options={dictOptions[dictTypeKeys.dialysateGroup] || []}
        required
        value={values.dialysateGroup}
        onChange={(e) => {
          const nextGroup = e.target.value
          const parsedIons = parseDialysateGroupIons(nextGroup)
          onChange((prev) => ({
            ...prev,
            dialysateGroup: nextGroup,
            ca: parsedIons.ca || prev.ca,
            k: parsedIons.k || prev.k,
          }))
        }}
      />
      <SelectField
        label="透析液流速"
        options={dictOptions[dictTypeKeys.dialysateFlow] || []}
        required
        value={values.dialysateFlow}
        onChange={(e) => {
          const nextFlow = e.target.value
          onChange((prev) => ({
            ...prev,
            dialysateFlow: nextFlow,
            dialysateVolume: calcDialysateVolume(duration, nextFlow),
          }))
        }}
      />
      <NumericInputField
        label="Na离子浓度"
        suffix="mmol/L"
        value={values.na}
        openNumericPad={openNumericPad}
        onConfirm={(value) =>
          onChange((prev) => ({ ...prev, na: value }))
        }
        hint="当日 RNa 钠处方起算基线"
      />
      <NumericInputField
        label="Ca离子浓度"
        suffix="mmol/L"
        value={values.ca}
        openNumericPad={openNumericPad}
        onConfirm={(value) =>
          onChange((prev) => ({ ...prev, ca: value }))
        }
      />
      <NumericInputField
        label="K离子浓度"
        suffix="mmol/L"
        value={values.k}
        openNumericPad={openNumericPad}
        onConfirm={(value) =>
          onChange((prev) => ({ ...prev, k: value }))
        }
        warn={!!values.k && Number(values.k) !== 2.0}
        warnText={!!values.k && Number(values.k) !== 2.0 ? '特殊钾浓度（标准默认 2.0）' : undefined}
      />
      <NumericInputField
        label="HCO3-浓度"
        suffix="mmol/L"
        value={values.hco3}
        openNumericPad={openNumericPad}
        onConfirm={(value) =>
          onChange((prev) => ({ ...prev, hco3: value }))
        }
      />
      <SelectField
        label="葡萄糖浓度"
        options={dictOptions[dictTypeKeys.glucose] || []}
        value={values.glucose}
        onChange={(e) =>
          onChange((prev) => ({ ...prev, glucose: e.target.value }))
        }
      />
      <NumericInputField
        label="电导度"
        suffix="mS/cm"
        value={values.conductivity}
        openNumericPad={openNumericPad}
        onConfirm={(value) =>
          onChange((prev) => ({ ...prev, conductivity: value }))
        }
      />
      <NumericInputField
        label="透析液温度"
        suffix="°C"
        value={values.temp}
        openNumericPad={openNumericPad}
        onConfirm={(value) =>
          onChange((prev) => ({ ...prev, temp: value }))
        }
      />
      <InputField label="透析液量" suffix="L" value={values.dialysateVolume} readOnly />
    </FormSection>
  )
})
