/**
 * CreatePatientModal - 新增患者建档弹窗
 */

import { useState, useEffect } from 'react'
import { X, User } from 'lucide-react'
import { message, DatePicker, Cascader } from 'antd'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import { restApi } from '@/services'
import { dictCache, DICT_TYPES } from '@/services/dictApi'

// 设置 dayjs 中文语言包
dayjs.locale('zh-cn')

// ===== 输入验证函数 =====

// 验证身份证号码（中国大陆 18 位身份证）
function validateIDNumber(idNumber: string, idType: string): boolean {
  if (!idNumber?.trim()) return false

  const trimmed = idNumber.trim()

  if (idType === '身份证' || idType === 'ID_CARD') {
    // 18 位身份证号码正则
    const idCardRegex = /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/
    if (!idCardRegex.test(trimmed)) return false

    // 验证校验位
    const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
    const checkCodes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']
    let sum = 0
    for (let i = 0; i < 17; i++) {
      sum += parseInt(trimmed[i]) * weights[i]
    }
    const checkCode = checkCodes[sum % 11]
    return trimmed[17].toUpperCase() === checkCode
  }

  // 其他证件类型，只验证非空和长度
  return trimmed.length >= 4 && trimmed.length <= 50
}

// 验证手机号码（中国大陆手机号）
function validatePhoneNumber(phone: string): boolean {
  if (!phone?.trim()) return false
  // 中国大陆手机号正则：1 开头，第二位 3-9，共 11 位数字
  const phoneRegex = /^1[3-9]\d{9}$/
  return phoneRegex.test(phone.trim())
}

// 验证身高范围（cm）
function validateHeight(height: string): boolean {
  const h = parseFloat(height)
  return !isNaN(h) && h >= 30 && h <= 250
}

// 验证年龄范围
function validateAge(age: string): boolean {
  const a = parseInt(age)
  return !isNaN(a) && a >= 0 && a <= 150
}

interface CreatePatientModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

interface CreatePatientForm {
  // 基本信息
  name: string
  age: string
  gender: 'M' | 'F'
  bedNumber?: string
  diagnosis?: string
  patientType: string
  insuranceType: string
  insuranceTypeArray?: string[]  // 用于 Cascader 的级联选择
  dryWeight: string
  defaultMode: string
  doctorName: string
  // 基本信息档案
  birthday?: string
  height?: string
  idType?: string
  idNumber?: string
  visitCategory?: string
  visitNo?: string
  insuranceNo?: string
  phone?: string
  address?: string
}

// 表单初始值常量（避免重复代码）
const INITIAL_FORM: CreatePatientForm = {
  name: '',
  age: '',
  gender: 'M',
  bedNumber: '',
  diagnosis: '',
  patientType: '门诊',
  insuranceType: '自费',
  insuranceTypeArray: ['自费'],
  dryWeight: '65',
  defaultMode: 'HD',
  doctorName: '',
  birthday: '',
  height: '',
  idType: '身份证',
  idNumber: '',
  visitCategory: '门诊',
  visitNo: '',
  insuranceNo: '',
  phone: '',
  address: '',
}

// 医保类型级联选项（默认数据，将被字典数据覆盖）
const DEFAULT_INSURANCE_OPTIONS = [
  { value: '省医保普通', label: '省医保普通' },
  { value: '省医保大病', label: '省医保大病' },
  { value: '济南市职工普通', label: '济南市职工普通' },
  { value: '济南市居民普通', label: '济南市居民普通' },
  { value: '新农合', label: '新农合' },
  { value: '异地职工医保', label: '异地职工医保' },
  { value: '异地居民医保', label: '异地居民医保' },
  { value: '商业保险', label: '商业保险' },
  {
    value: '其他',
    label: '其他',
    children: [
      { value: '市职工工伤', label: '市职工工伤' },
      { value: '企业离休', label: '企业离休' },
      { value: '市医保补充', label: '市医保补充' },
      { value: '公费', label: '公费' },
      { value: '自费', label: '自费' },
      { value: '城镇', label: '城镇' },
      { value: '军免', label: '军免' },
    ]
  }
]

export default function CreatePatientModal({ open, onClose, onSuccess }: CreatePatientModalProps) {
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState<CreatePatientForm>(INITIAL_FORM)
  const [dictOptions, setDictOptions] = useState<Record<string, Array<{ value: string; label: string }>>>({})
  // 医保类型选项（树形结构用于 Cascader）
  const [insuranceOptions, setInsuranceOptions] = useState<Array<{ value: string; label: string; children?: Array<{ value: string; label: string }> }>>(DEFAULT_INSURANCE_OPTIONS)

  // 加载字典数据
  useEffect(() => {
    const loadDictData = async () => {
      try {
        console.log('开始加载字典数据...')

        // 单独加载透析模式字典（用于调试）
        const dialysisModes = await dictCache.getOptions(DICT_TYPES.DIALYSIS_MODE)
        console.log('透析模式字典数据:', dialysisModes)

        // 并行加载所有需要的字典数据
        const [patientTypes, idTypes, visitCategories, insuranceTree] = await Promise.all([
          dictCache.getOptions(DICT_TYPES.PATIENT_TYPE),
          dictCache.getOptions(DICT_TYPES.ID_TYPE),
          dictCache.getOptions(DICT_TYPES.VISIT_CATEGORY),
          dictCache.getCascaderOptions(DICT_TYPES.INSURANCE_TYPE),
        ])

        const newDictOptions = {
          [DICT_TYPES.PATIENT_TYPE]: patientTypes,
          [DICT_TYPES.ID_TYPE]: idTypes,
          [DICT_TYPES.VISIT_CATEGORY]: visitCategories,
          [DICT_TYPES.DIALYSIS_MODE]: dialysisModes,
        }

        console.log('设置字典选项:', newDictOptions)

        setDictOptions(newDictOptions)

        // 更新医保类型选项
        if (insuranceTree && insuranceTree.length > 0) {
          setInsuranceOptions(insuranceTree)
        }

        // 设置表单默认值为字典第一个值
        setForm(prev => ({
          ...prev,
          patientType: patientTypes[0]?.label || '门诊',
          idType: idTypes[0]?.label || '身份证',
          visitCategory: visitCategories[0]?.label || '门诊',
          defaultMode: dialysisModes[0]?.value || 'HD',
        }))
      } catch (error) {
        console.error('加载字典数据失败:', error)
        // 打印更详细的错误信息
        if (error instanceof Error) {
          console.error('错误消息:', error.message)
          console.error('错误堆栈:', error.stack)
        }
      }
    }

    loadDictData()
  }, [])

  const handleSubmit = async () => {
    // 验证必填字段
    if (!form.name.trim()) {
      message.error('请输入患者姓名')
      return
    }
    if (!validateAge(form.age)) {
      message.error('请输入有效的年龄（0-150）')
      return
    }
    // 验证基本信息档案必填字段
    if (!validateHeight(form.height || '')) {
      message.error('请输入有效的身高（30-250cm）')
      return
    }
    // 验证身份证号
    if (!form.idNumber?.trim()) {
      message.error('请输入证件号码')
      return
    }
    const idTypeCode = dictOptions[DICT_TYPES.ID_TYPE]?.find(
      opt => opt.label === form.idType
    )?.value || 'ID_CARD'
    console.log('证件验证 - form.idType:', form.idType)
    console.log('证件验证 - idTypeCode:', idTypeCode)
    console.log('证件验证 - idNumber:', form.idNumber)
    const validationResult = validateIDNumber(form.idNumber, idTypeCode)
    console.log('证件验证 - 结果:', validationResult)
    if (!validationResult) {
      message.error('请输入有效的证件号码')
      return
    }
    // 验证就诊类别
    if (!form.visitCategory?.trim()) {
      message.error('请选择就诊类别')
      return
    }
    if (!form.visitNo?.trim()) {
      message.error('请输入就诊号')
      return
    }
    if (!form.insuranceNo?.trim()) {
      message.error('请输入医保号')
      return
    }
    // 验证手机号
    if (!form.phone?.trim()) {
      message.error('请输入联系电话')
      return
    }
    if (!validatePhoneNumber(form.phone)) {
      message.error('请输入有效的手机号码')
      return
    }
    if (!form.address?.trim()) {
      message.error('请输入联系地址')
      return
    }

    setLoading(true)
    try {
      await restApi.createPatient({
        name: form.name.trim(),
        age: parseInt(form.age),
        gender: form.gender,
        bedNumber: form.bedNumber || undefined,
        diagnosis: form.diagnosis || undefined,
        patientType: form.patientType,
        insuranceType: form.insuranceType,
        dryWeight: parseFloat(form.dryWeight) || 65,
        defaultMode: form.defaultMode,
        doctorName: form.doctorName || undefined,
        // 基本信息档案
        birthday: form.birthday || undefined,
        height: form.height || undefined,
        idType: form.idType || undefined,
        idNumber: form.idNumber || undefined,
        visitCategory: form.visitCategory || undefined,
        visitNo: form.visitNo || undefined,
        insuranceNo: form.insuranceNo || undefined,
        phone: form.phone || undefined,
        address: form.address || undefined,
      })

      message.success('建档成功')
      setForm(INITIAL_FORM)  // 重置表单，下次打开时为空
      onSuccess()  // 通知父组件刷新列表并关闭弹窗
    } catch (error) {
      console.error('创建患者失败:', error)
      message.error('建档失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setForm(INITIAL_FORM)
    onClose()
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <div className="flex items-center">
            <div className="w-10 h-10 bg-blue-50 rounded-xl flex items-center justify-center mr-3">
              <User className="text-blue-600" size={20} />
            </div>
            <div>
              <h2 className="text-lg font-bold text-gray-800">新增患者建档</h2>
              <p className="text-xs text-gray-500">填写患者基本信息</p>
            </div>
          </div>
          <button
            onClick={handleClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X size={20} className="text-gray-400" />
          </button>
        </div>

        {/* Form Content */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
          <div className="grid grid-cols-2 gap-6">
            {/* 必填字段 */}
            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                姓名 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="请输入患者姓名"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                年龄 <span className="text-red-500">*</span>
              </label>
              <input
                type="number"
                value={form.age}
                onChange={(e) => setForm({ ...form, age: e.target.value })}
                placeholder="请输入年龄"
                min={0}
                max={150}
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                性别 <span className="text-red-500">*</span>
              </label>
              <div className="flex gap-4">
                <button
                  type="button"
                  onClick={() => setForm({ ...form, gender: 'M' })}
                  className={`flex-1 py-2.5 px-4 rounded-lg border-2 font-medium transition-all ${
                    form.gender === 'M'
                      ? 'border-blue-500 bg-blue-50 text-blue-700'
                      : 'border-gray-200 text-gray-600 hover:border-gray-300'
                  }`}
                >
                  男
                </button>
                <button
                  type="button"
                  onClick={() => setForm({ ...form, gender: 'F' })}
                  className={`flex-1 py-2.5 px-4 rounded-lg border-2 font-medium transition-all ${
                    form.gender === 'F'
                      ? 'border-pink-500 bg-pink-50 text-pink-700'
                      : 'border-gray-200 text-gray-600 hover:border-gray-300'
                  }`}
                >
                  女
                </button>
              </div>
            </div>

            {/* 可选字段 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                床号
              </label>
              <input
                type="text"
                value={form.bedNumber}
                onChange={(e) => setForm({ ...form, bedNumber: e.target.value })}
                placeholder="例如：A01"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                患者类型
              </label>
              <select
                value={form.patientType}
                onChange={(e) => setForm({ ...form, patientType: e.target.value })}
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              >
                {(dictOptions[DICT_TYPES.PATIENT_TYPE] || [{ value: 'OUTPATIENT', label: '门诊' }, { value: 'INPATIENT', label: '住院' }]).map(opt => (
                  <option key={opt.value} value={opt.label}>{opt.label}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                医保类型
              </label>
              <Cascader
                value={form.insuranceTypeArray || []}
                onChange={(value) => {
                  // 将数组转换为字符串保存
                  const insuranceType = value.length > 1 ? value[1] : value[0]
                  setForm({ ...form, insuranceType, insuranceTypeArray: value })
                }}
                options={insuranceOptions}
                placeholder="请选择医保类型"
                style={{ width: '100%' }}
                size="large"
                expandTrigger="click"
                changeOnSelect
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                干体重 (kg)
              </label>
              <input
                type="number"
                value={form.dryWeight}
                onChange={(e) => setForm({ ...form, dryWeight: e.target.value })}
                placeholder="默认 65"
                step="0.1"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                默认透析模式
              </label>
              <select
                value={form.defaultMode}
                onChange={(e) => setForm({ ...form, defaultMode: e.target.value })}
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              >
                {(dictOptions[DICT_TYPES.DIALYSIS_MODE] || [{ value: 'HD', label: 'HD' }, { value: 'HFD', label: 'HFD' }, { value: 'HDF', label: 'HDF' }, { value: 'HF', label: 'HF' }]).map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                主治医生
              </label>
              <input
                type="text"
                value={form.doctorName}
                onChange={(e) => setForm({ ...form, doctorName: e.target.value })}
                placeholder="请输入医生姓名"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                诊断
              </label>
              <input
                type="text"
                value={form.diagnosis}
                onChange={(e) => setForm({ ...form, diagnosis: e.target.value })}
                placeholder="例如：慢性肾脏病5期"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            {/* 分隔线 */}
            <div className="col-span-2 border-t border-gray-200 my-4"></div>

            {/* 基本信息档案 */}
            <div className="col-span-2">
              <h3 className="text-base font-bold text-gray-800 mb-4 flex items-center">
                <span className="w-1 h-5 bg-purple-500 rounded mr-2"></span>
                基本信息
              </h3>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                出生日期
              </label>
              <DatePicker
                value={form.birthday ? dayjs(form.birthday) : null}
                onChange={(date) => setForm({ ...form, birthday: date ? date.format('YYYY-MM-DD') : '' })}
                placeholder="请选择出生日期"
                style={{ width: '100%' }}
                format="YYYY年MM月DD日"
                className="w-full"
                getPopupContainer={(trigger) => trigger.parentElement!}
                size="large"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                身高 (cm) <span className="text-red-500">*</span>
              </label>
              <input
                type="number"
                value={form.height}
                onChange={(e) => setForm({ ...form, height: e.target.value })}
                placeholder="请输入身高"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                证件类型
              </label>
              <select
                value={form.idType}
                onChange={(e) => setForm({ ...form, idType: e.target.value })}
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              >
                {(dictOptions[DICT_TYPES.ID_TYPE] || [{ value: 'ID_CARD', label: '身份证' }, { value: 'PASSPORT', label: '护照' }, { value: 'OTHER', label: '其他' }]).map(opt => (
                  <option key={opt.value} value={opt.label}>{opt.label}</option>
                ))}
              </select>
            </div>

            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                证件号码 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={form.idNumber}
                onChange={(e) => setForm({ ...form, idNumber: e.target.value })}
                placeholder="请输入证件号码"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                就诊类别 <span className="text-red-500">*</span>
              </label>
              <select
                value={form.visitCategory}
                onChange={(e) => setForm({ ...form, visitCategory: e.target.value })}
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              >
                {(dictOptions[DICT_TYPES.VISIT_CATEGORY] || [{ value: 'OUTPATIENT', label: '门诊' }, { value: 'INPATIENT', label: '住院' }]).map(opt => (
                  <option key={opt.value} value={opt.label}>{opt.label}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                就诊号 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={form.visitNo}
                onChange={(e) => setForm({ ...form, visitNo: e.target.value })}
                placeholder="请输入就诊号"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                医保号 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={form.insuranceNo}
                onChange={(e) => setForm({ ...form, insuranceNo: e.target.value })}
                placeholder="请输入医保号"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                联系电话 <span className="text-red-500">*</span>
              </label>
              <input
                type="tel"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                placeholder="请输入联系电话"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                联系地址 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={form.address}
                onChange={(e) => setForm({ ...form, address: e.target.value })}
                placeholder="请输入联系地址"
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              />
            </div>

            {/* 分隔线 */}
            <div className="col-span-2 border-t border-gray-200 my-4"></div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100 bg-gray-50">
          <button
            onClick={handleClose}
            disabled={loading}
            className="px-6 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-100 transition-colors disabled:opacity-50"
          >
            取消
          </button>
          <button
            onClick={handleSubmit}
            disabled={loading}
            className="px-6 py-2.5 bg-blue-600 rounded-lg text-sm font-medium text-white hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center"
          >
            {loading ? (
              <>
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full mr-2 animate-spin"></div>
                提交中...
              </>
            ) : (
              <>
                <User size={16} className="mr-2" />
                确认建档
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
