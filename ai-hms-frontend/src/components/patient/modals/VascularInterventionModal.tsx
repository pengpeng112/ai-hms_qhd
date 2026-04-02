// VascularInterventionModal - 血管通路干预弹窗
// 使用字典数据：通路类型、手术类型、干预医生

import { useState, useEffect } from 'react'
import { Loader2 } from 'lucide-react'
import { Select, DatePicker, message } from 'antd'
import dayjs from 'dayjs'
import { dictCache, DICT_TYPES } from '@/services/dictApi'

interface VascularInterventionModalProps {
  isOpen: boolean
  onClose: () => void
  onSave?: (data: InterventionFormData) => Promise<void>
  initialData?: Partial<InterventionFormData>
  vascularAccessType?: string  // 当前血管通路类型
}

export interface InterventionFormData {
  accessType: string       // 通路类型（字典）
  avgBloodFlow: number     // 平均血流量
  surgeryType: string      // 手术类型（字典，必填）
  interventionReason: string  // 干预原因（必填）
  usageDays: number        // 使用时间（天）
  doctor: string           // 干预医生（字典）
  interventionDate: string // 干预时间
  description: string      // 干预描述
}

interface DictOption {
  value: string
  label: string
}

const initialFormData: InterventionFormData = {
  accessType: '',
  avgBloodFlow: 0,
  surgeryType: '',
  interventionReason: '',
  usageDays: 0,
  doctor: '',
  interventionDate: dayjs().format('YYYY-MM-DD'),
  description: '',
}

export default function VascularInterventionModal({
  isOpen,
  onClose,
  onSave,
  initialData,
  vascularAccessType,
}: VascularInterventionModalProps) {
  const [formData, setFormData] = useState<InterventionFormData>({
    ...initialFormData,
    ...initialData,
    accessType: vascularAccessType || initialData?.accessType || '',
  })
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(false)

  // 字典选项
  const [accessTypeOptions, setAccessTypeOptions] = useState<DictOption[]>([])
  const [surgeryTypeOptions, setSurgeryTypeOptions] = useState<DictOption[]>([])
  const [doctorOptions, setDoctorOptions] = useState<DictOption[]>([])

  // 当弹窗打开时重置表单并加载字典数据
  useEffect(() => {
    if (isOpen) {
      setFormData({
        ...initialFormData,
        ...initialData,
        accessType: vascularAccessType || initialData?.accessType || '',
      })
      loadDictData()
    }
  }, [isOpen, initialData, vascularAccessType])

  // 加载字典数据
  const loadDictData = async () => {
    setLoading(true)
    try {
      const [accessTypes, surgeryTypes, doctors] = await Promise.all([
        dictCache.getOptions(DICT_TYPES.VASCULAR_ACCESS),
        dictCache.getOptions(DICT_TYPES.SURGERY_TYPE),
        dictCache.getOptions(DICT_TYPES.DOCTOR),
      ])
      setAccessTypeOptions(accessTypes)
      setSurgeryTypeOptions(surgeryTypes)
      setDoctorOptions(doctors)
    } catch (error) {
      console.error('加载字典数据失败:', error)
      // 使用默认值
      setAccessTypeOptions([
        { value: 'AVF', label: '内瘘 - AVF' },
        { value: 'AVG', label: '内瘘 - AVG' },
        { value: 'TCC', label: '导管 - TCC' },
        { value: 'NCC', label: '导管 - NCC' },
      ])
      setSurgeryTypeOptions([
        { value: 'PTA', label: 'PTA（经皮腔内血管成形术）' },
        { value: 'THROMBECTOMY', label: '血栓清除术' },
        { value: 'REVISION', label: '修复术' },
        { value: 'LIGATION', label: '结扎术' },
        { value: 'CATHETER_EXCHANGE', label: '导管更换' },
        { value: 'CATHETER_REPOSITIONING', label: '导管重新定位' },
        { value: 'OTHER', label: '其他' },
      ])
      setDoctorOptions([
        { value: 'DR_WANG', label: '王医生' },
        { value: 'DR_LI', label: '李医生' },
        { value: 'DR_ZHANG', label: '张医生' },
        { value: 'DR_CHEN', label: '陈医生' },
      ])
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (field: keyof InterventionFormData, value: string | number) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const handleSave = async () => {
    // 验证必填项
    if (!formData.surgeryType) {
      message.error('请选择手术类型')
      return
    }
    if (!formData.interventionReason) {
      message.error('请填写干预原因')
      return
    }

    if (saving) return
    setSaving(true)

    try {
      await onSave?.(formData)
      message.success('保存成功')
      onClose()
    } catch (error) {
      console.error('保存干预记录失败:', error)
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const isFormValid = formData.surgeryType && formData.interventionReason

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-[#f0f7ff] rounded-lg shadow-2xl w-full max-w-5xl overflow-hidden animate-scale-in border border-slate-200">
        {/* 标题 */}
        <div className="px-10 py-4 bg-white border-b border-slate-100">
          <h3 className="text-base font-bold text-slate-800">血管通路干预记录</h3>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="animate-spin text-blue-500" size={32} />
            <span className="ml-3 text-slate-500">加载中...</span>
          </div>
        ) : (
          <div className="p-10 space-y-8">
            <div className="grid grid-cols-2 gap-x-16 gap-y-6">
              {/* Left Column */}
              <div className="space-y-6">
                {/* 通路类型 - 使用字典 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">通路类型:</label>
                  <div className="relative flex-1">
                    <Select
                      value={formData.accessType || undefined}
                      onChange={(value) => handleChange('accessType', value)}
                      options={accessTypeOptions}
                      placeholder="请选择通路类型"
                      style={{ width: '100%' }}
                      size="large"
                      allowClear
                    />
                  </div>
                </div>

                {/* 平均血流量 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">平均血流量:</label>
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      value={formData.avgBloodFlow}
                      onChange={(e) => handleChange('avgBloodFlow', parseInt(e.target.value) || 0)}
                      className="w-32 h-10 border border-slate-300 rounded px-3 text-sm focus:ring-1 focus:ring-blue-500 outline-none"
                    />
                    <span className="text-sm text-slate-500">ml/min</span>
                  </div>
                </div>

                {/* 手术类型 - 使用字典，必填 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">
                    <span className="text-red-500 mr-1">*</span>手术类型:
                  </label>
                  <div className="relative flex-1">
                    <Select
                      value={formData.surgeryType || undefined}
                      onChange={(value) => handleChange('surgeryType', value)}
                      options={surgeryTypeOptions}
                      placeholder="请选择手术类型"
                      style={{ width: '100%' }}
                      size="large"
                      allowClear
                      status={!formData.surgeryType ? 'error' : undefined}
                    />
                  </div>
                </div>

                {/* 干预原因 - 必填 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">
                    <span className="text-red-500 mr-1">*</span>干预原因:
                  </label>
                  <input
                    type="text"
                    value={formData.interventionReason}
                    onChange={(e) => handleChange('interventionReason', e.target.value)}
                    placeholder="请输入干预原因"
                    className={`flex-1 h-10 border rounded px-3 text-sm focus:ring-1 focus:ring-blue-500 outline-none ${
                      !formData.interventionReason ? 'border-red-300' : 'border-slate-300'
                    }`}
                  />
                </div>
              </div>

              {/* Right Column */}
              <div className="space-y-6">
                {/* 使用时间 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">使用时间:</label>
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      value={formData.usageDays}
                      onChange={(e) => handleChange('usageDays', parseInt(e.target.value) || 0)}
                      className="w-32 h-10 border border-slate-300 rounded px-3 text-sm focus:ring-1 focus:ring-blue-500 outline-none"
                    />
                    <span className="text-sm text-slate-500">天</span>
                  </div>
                </div>

                {/* 干预医生 - 使用字典 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">干预医生:</label>
                  <div className="relative flex-1">
                    <Select
                      value={formData.doctor || undefined}
                      onChange={(value) => handleChange('doctor', value)}
                      options={doctorOptions}
                      placeholder="请选择干预医生"
                      style={{ width: '100%' }}
                      size="large"
                      allowClear
                      showSearch
                      filterOption={(input, option) =>
                        (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                      }
                    />
                  </div>
                </div>

                {/* 干预时间 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">干预时间:</label>
                  <div className="relative flex-1">
                    <DatePicker
                      value={formData.interventionDate ? dayjs(formData.interventionDate) : null}
                      onChange={(date) => handleChange('interventionDate', date ? date.format('YYYY-MM-DD') : '')}
                      style={{ width: '100%' }}
                      size="large"
                      placeholder="选择日期"
                    />
                  </div>
                </div>

                {/* 干预描述 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-bold text-slate-600 mr-4">干预描述:</label>
                  <input
                    type="text"
                    value={formData.description}
                    onChange={(e) => handleChange('description', e.target.value)}
                    placeholder="请输入干预描述"
                    className="flex-1 h-10 border border-slate-300 rounded px-3 text-sm focus:ring-1 focus:ring-blue-500 outline-none"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Footer */}
        <div className="px-10 py-6 bg-slate-50 border-t border-slate-200 flex justify-end gap-3 shrink-0">
          <button
            onClick={onClose}
            disabled={saving}
            className="px-8 py-2 bg-white border border-slate-300 text-slate-700 rounded text-sm font-bold hover:bg-slate-50 transition-all disabled:opacity-50"
          >
            取消
          </button>
          <button
            onClick={handleSave}
            disabled={saving || !isFormValid}
            className={`px-8 py-2 rounded text-sm font-bold transition-all flex items-center gap-2 ${
              isFormValid
                ? 'bg-blue-600 text-white hover:bg-blue-700 shadow-lg'
                : 'bg-[#f4f4f4] border border-slate-200 text-slate-400 cursor-not-allowed'
            }`}
          >
            {saving && <Loader2 className="animate-spin" size={16} />}
            {saving ? '保存中...' : '确定'}
          </button>
        </div>
      </div>
    </div>
  )
}
