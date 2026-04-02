// VascularAccessModal - 血管通路弹窗
// 根据通路类型动态显示不同字段：AVF/AVG（内瘘）vs TCC/NCC（导管）
// 使用字典数据，静脉和动脉支持多选

import { useState, useEffect, useRef } from 'react'
import { X, ChevronDown, Upload, XCircle } from 'lucide-react'
import { Select, DatePicker } from 'antd'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import { dictCache, DICT_TYPES } from '@/services/dictApi'

// 设置 dayjs 中文语言包
dayjs.locale('zh-cn')

interface VascularAccessModalProps {
  isOpen: boolean
  onClose: () => void
  onSave?: (data: VascularFormData) => Promise<void>
  initialData?: Partial<VascularFormData>
  mode?: 'create' | 'edit'
  saving?: boolean
}

export interface VascularFormData {
  accessType: string
  site: string
  artery: string[]  // 改为数组，支持多选
  vein: string[]    // 改为数组，支持多选
  side: string
  hospital: string
  surgeon: string
  surgeryDate: string
  firstUseDate: string
  accessNumber: number
  interventionCount: number
  interventionDate: string
  catheterMethod: string
  catheterDepth: string
  vPuncturePosition: string[]  // 改为数组，支持多项输入
  aPuncturePosition: string[]  // 改为数组，支持多项输入
  remark: string
  images: string[]  // 图片URLs
  isDefault: boolean
  isDisabled: boolean
}

// 中心静脉置管方法（暂不做字典）
const CATHETER_METHOD_OPTIONS = [
  { value: '超声介入', label: '超声介入' },
  { value: 'X线介入', label: 'X线介入' },
  { value: '盲穿', label: '盲穿' },
]

// 左右选项
const SIDE_OPTIONS = [
  { value: 'L', label: '左' },
  { value: 'R', label: '右' },
]

const initialFormData: VascularFormData = {
  accessType: '',
  site: '',
  artery: [],
  vein: [],
  side: 'L',
  hospital: '',
  surgeon: '',
  surgeryDate: '',
  firstUseDate: '',
  accessNumber: 1,
  interventionCount: 0,
  interventionDate: '',
  catheterMethod: '超声介入',
  catheterDepth: '',
  vPuncturePosition: [],
  aPuncturePosition: [],
  remark: '',
  images: [],
  isDefault: true,
  isDisabled: false,
}

interface DictOption {
  value: string
  label: string
}

export default function VascularAccessModal({ isOpen, onClose, onSave, initialData, mode = 'create', saving = false }: VascularAccessModalProps) {
  const [formData, setFormData] = useState<VascularFormData>({
    ...initialFormData,
    ...initialData,
  })

  // 当 initialData 变化时重置表单
  useEffect(() => {
    if (isOpen) {
      setFormData({
        ...initialFormData,
        ...initialData,
      })
    }
  }, [isOpen, initialData])

  // 字典选项
  const [accessTypeOptions, setAccessTypeOptions] = useState<DictOption[]>([])
  const [siteOptions, setSiteOptions] = useState<DictOption[]>([])
  const [veinOptions, setVeinOptions] = useState<DictOption[]>([])
  const [arteryOptions, setArteryOptions] = useState<DictOption[]>([])
  const [hospitalOptions, setHospitalOptions] = useState<DictOption[]>([])
  const [surgeonOptions, setSurgeonOptions] = useState<DictOption[]>([])
  const [loading, setLoading] = useState(false)

  // 图片上传相关
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  // 触发文件选择
  const handleUploadClick = () => {
    fileInputRef.current?.click()
  }

  // 处理文件选择
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files || files.length === 0) return

    setUploading(true)
    try {
      const newImages: string[] = []

      for (let i = 0; i < files.length; i++) {
        const file = files[i]

        // 验证文件类型
        if (!file.type.startsWith('image/')) {
          console.error('只能上传图片文件')
          continue
        }

        // 验证文件大小（限制5MB）
        if (file.size > 5 * 1024 * 1024) {
          console.error('图片大小不能超过5MB')
          continue
        }

        // 将文件转换为 Base64 URL（实际项目中应该上传到服务器）
        const reader = new FileReader()
        const promise = new Promise<string>((resolve) => {
          reader.onload = (e) => resolve(e.target?.result as string)
          reader.readAsDataURL(file)
        })
        const dataUrl = await promise
        newImages.push(dataUrl)
      }

      // 更新表单数据
      setFormData(prev => ({
        ...prev,
        images: [...prev.images, ...newImages],
      }))
    } catch (error) {
      console.error('图片上传失败:', error)
    } finally {
      setUploading(false)
      // 重置文件输入
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  // 删除图片
  const handleRemoveImage = (index: number) => {
    setFormData(prev => ({
      ...prev,
      images: prev.images.filter((_, i) => i !== index),
    }))
  }

  // 加载字典数据
  useEffect(() => {
    if (!isOpen) return

    const loadDictData = async () => {
      setLoading(true)
      try {
        const [accessTypes, sites, veins, arteries, hospitals, surgeons] = await Promise.all([
          dictCache.getOptions(DICT_TYPES.VASCULAR_ACCESS),
          dictCache.getOptions(DICT_TYPES.VASCULAR_SITE),
          dictCache.getOptions(DICT_TYPES.VEIN_TYPE),
          dictCache.getOptions(DICT_TYPES.ARTERY_TYPE),
          dictCache.getOptions(DICT_TYPES.HOSPITAL),
          dictCache.getOptions(DICT_TYPES.DOCTOR),
        ])
        setAccessTypeOptions(accessTypes)
        setSiteOptions(sites)
        setVeinOptions(veins)
        setArteryOptions(arteries)
        setHospitalOptions(hospitals)
        setSurgeonOptions(surgeons)
      } catch (error) {
        console.error('加载字典数据失败:', error)
        // 使用默认值
        setAccessTypeOptions([
          { value: 'AVF', label: '内瘘 - AVF' },
          { value: 'AVG', label: '内瘘 - AVG' },
          { value: 'TCC', label: '导管 - TCC' },
          { value: 'NCC', label: '导管 - NCC' },
        ])
        setSiteOptions([
          { value: '下肢', label: '下肢' },
          { value: '上臂', label: '上臂' },
          { value: '肘部', label: '肘部' },
          { value: '锁骨下', label: '锁骨下' },
          { value: '腕部', label: '腕部' },
          { value: '腹股沟', label: '腹股沟' },
          { value: '颈部', label: '颈部' },
          { value: '前臂中段', label: '前臂中段' },
        ])
        setVeinOptions([
          { value: '头静脉', label: '头静脉' },
          { value: '贵要静脉', label: '贵要静脉' },
          { value: '肘正中静脉', label: '肘正中静脉' },
          { value: '颈内静脉', label: '颈内静脉' },
          { value: '锁骨下静脉', label: '锁骨下静脉' },
          { value: '股静脉', label: '股静脉' },
        ])
        setArteryOptions([
          { value: '桡动脉', label: '桡动脉' },
          { value: '肱动脉', label: '肱动脉' },
          { value: '尺动脉', label: '尺动脉' },
          { value: '股动脉', label: '股动脉' },
        ])
        setHospitalOptions([
          { value: 'H_LOCAL', label: '本院' },
          { value: 'H_PEK_UNION', label: '北京协和医院' },
          { value: 'H_OTHER', label: '其他医院' },
        ])
        setSurgeonOptions([
          { value: 'DR_WANG', label: '王医生' },
          { value: 'DR_LI', label: '李医生' },
          { value: 'DR_ZHANG', label: '张医生' },
        ])
      } finally {
        setLoading(false)
      }
    }

    loadDictData()
  }, [isOpen])

  const handleChange = (field: keyof VascularFormData, value: string | string[] | number | boolean) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const handleSave = async () => {
    if (saving) return
    try {
      await onSave?.(formData)
    } catch {
      // 错误已在父组件处理
    }
  }

  // 判断是否为导管类型
  const isCatheter = formData.accessType === 'TCC' || formData.accessType === 'NCC'

  // 计算使用天数
  const calculateDays = () => {
    if (!formData.firstUseDate) return 0
    const firstDate = new Date(formData.firstUseDate)
    const today = new Date()
    const diff = Math.floor((today.getTime() - firstDate.getTime()) / (1000 * 60 * 60 * 24))
    return diff > 0 ? diff : 0
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-[12px] shadow-2xl w-full max-w-5xl overflow-hidden animate-scale-in flex flex-col ring-1 ring-black/5">
        {/* Header */}
        <div className="bg-[#eef6ff] px-6 py-4 flex items-center justify-between border-b border-blue-100">
          <h3 className="text-base font-bold text-slate-800 flex items-center">
            {mode === 'edit' ? '编辑血管通路' : '新增血管通路'}
          </h3>
          <button
            onClick={onClose}
            disabled={saving}
            className="p-1.5 hover:bg-white/50 rounded-full transition-all text-slate-400 hover:text-slate-600 disabled:opacity-50"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="p-8 bg-white overflow-y-auto max-h-[85vh]">
          {loading ? (
            <div className="flex items-center justify-center py-20">
              <div className="text-slate-400">加载中...</div>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-x-12 gap-y-6">
              {/* 左侧列 */}
              <div className="space-y-6">
                {/* 通路类型 - 使用字典 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">
                    <span className="text-red-500 mr-1">*</span>通路类型:
                  </label>
                  <div className="relative flex-1">
                    <select
                      value={formData.accessType}
                      onChange={(e) => handleChange('accessType', e.target.value)}
                      className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm appearance-none outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                    >
                      <option value="">请选择通路类型</option>
                      {accessTypeOptions.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                    <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                {/* 通路部位（仅内瘘显示）- 使用字典 */}
                {!isCatheter && (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">通路部位:</label>
                    <div className="relative flex-1">
                      <select
                        value={formData.site}
                        onChange={(e) => handleChange('site', e.target.value)}
                        className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm appearance-none outline-none focus:ring-1 focus:ring-blue-500 bg-white text-slate-600"
                      >
                        <option value="">请选择通路部位</option>
                        {siteOptions.map(opt => (
                          <option key={opt.value} value={opt.value}>{opt.label}</option>
                        ))}
                      </select>
                      <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                    </div>
                  </div>
                )}

                {/* 静脉（导管）- 使用字典，多选 */}
                {isCatheter ? (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">静脉:</label>
                    <div className="flex-1">
                      <Select
                        mode="multiple"
                        allowClear
                        placeholder="请选择静脉（可多选）"
                        value={formData.vein}
                        onChange={(value) => handleChange('vein', value)}
                        options={veinOptions.filter(opt =>
                          ['颈内静脉', '股静脉', '锁骨下静脉'].includes(opt.value) || opt.label.includes('静脉')
                        )}
                        style={{ width: '100%' }}
                        size="middle"
                      />
                    </div>
                  </div>
                ) : (
                  /* 动脉（内瘘）- 使用字典，多选 */
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">动脉:</label>
                    <div className="flex-1">
                      <Select
                        mode="multiple"
                        allowClear
                        placeholder="请选择动脉（可多选）"
                        value={formData.artery}
                        onChange={(value) => handleChange('artery', value)}
                        options={arteryOptions}
                        style={{ width: '100%' }}
                        size="middle"
                      />
                    </div>
                  </div>
                )}

                {/* 左右 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">左右:</label>
                  <div className="relative flex-1">
                    <select
                      value={formData.side}
                      onChange={(e) => handleChange('side', e.target.value)}
                      className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm appearance-none outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                    >
                      {SIDE_OPTIONS.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                    <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                {/* 手术医院 - 使用字典 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">手术医院:</label>
                  <div className="relative flex-1">
                    <select
                      value={formData.hospital}
                      onChange={(e) => handleChange('hospital', e.target.value)}
                      className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm appearance-none outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                    >
                      <option value="">请选择手术医院</option>
                      {hospitalOptions.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                    <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                {/* 中心静脉置管方法（仅导管显示） */}
                {isCatheter && (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">中心静脉置管方法:</label>
                    <div className="relative flex-1">
                      <select
                        value={formData.catheterMethod}
                        onChange={(e) => handleChange('catheterMethod', e.target.value)}
                        className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm appearance-none outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                      >
                        {CATHETER_METHOD_OPTIONS.map(opt => (
                          <option key={opt.value} value={opt.value}>{opt.label}</option>
                        ))}
                      </select>
                      <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                    </div>
                  </div>
                )}

                {/* 手术医生 - 使用字典 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">手术医生:</label>
                  <div className="relative flex-1">
                    <select
                      value={formData.surgeon}
                      onChange={(e) => handleChange('surgeon', e.target.value)}
                      className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm appearance-none outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                    >
                      <option value="">请选择手术医生</option>
                      {surgeonOptions.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                    <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                {/* V侧穿刺点位置（仅内瘘显示）- 多项输入 */}
                {!isCatheter && (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">V侧穿刺点位置:</label>
                    <div className="flex-1">
                      <Select
                        mode="tags"
                        allowClear
                        placeholder="输入位置后按回车"
                        value={formData.vPuncturePosition}
                        onChange={(value) => handleChange('vPuncturePosition', value)}
                        style={{ width: '100%' }}
                        size="middle"
                        tokenSeparators={[',']}
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* 右侧列 */}
              <div className="space-y-6">
                {/* 静脉（仅内瘘显示）- 使用字典，多选 */}
                {!isCatheter && (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">静脉:</label>
                    <div className="flex-1">
                      <Select
                        mode="multiple"
                        allowClear
                        placeholder="请选择静脉（可多选）"
                        value={formData.vein}
                        onChange={(value) => handleChange('vein', value)}
                        options={veinOptions}
                        style={{ width: '100%' }}
                        size="middle"
                      />
                    </div>
                  </div>
                )}

                {/* 手术时间 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">
                    <span className="text-red-500 mr-1">*</span>手术时间:
                  </label>
                  <div className="relative flex-1">
                    <DatePicker
                      value={formData.surgeryDate ? dayjs(formData.surgeryDate) : null}
                      onChange={(date) => handleChange('surgeryDate', date ? date.format('YYYY-MM-DD') : '')}
                      format="YYYY年MM月DD日"
                      placeholder="请选择手术时间"
                      style={{ width: '100%', height: 40 }}
                      size="middle"
                    />
                  </div>
                </div>

                {/* 导管深度（仅导管显示） */}
                {isCatheter && (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">导管深度:</label>
                    <div className="flex-1 relative">
                      <input
                        type="number"
                        placeholder="输入深度"
                        value={formData.catheterDepth}
                        onChange={(e) => handleChange('catheterDepth', e.target.value)}
                        className="w-full h-10 border border-slate-200 rounded-md px-3 text-sm outline-none focus:ring-1 focus:ring-blue-500 pr-10"
                      />
                      <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-400">cm</span>
                    </div>
                  </div>
                )}

                {/* 首次使用时间 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">首次使用时间:</label>
                  <div className="flex-1 flex items-center gap-2">
                    <div className="relative flex-1">
                      <DatePicker
                        value={formData.firstUseDate ? dayjs(formData.firstUseDate) : null}
                        onChange={(date) => handleChange('firstUseDate', date ? date.format('YYYY-MM-DD') : '')}
                        format="YYYY年MM月DD日"
                        placeholder="请选择首次使用时间"
                        style={{ width: '100%', height: 40 }}
                        size="middle"
                      />
                    </div>
                    <span className="text-xs text-slate-400 whitespace-nowrap">（{calculateDays()}天）</span>
                  </div>
                </div>

                {/* 第几次血管通路 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">第几次血管通路:</label>
                  <input
                    type="number"
                    value={formData.accessNumber}
                    onChange={(e) => handleChange('accessNumber', parseInt(e.target.value) || 1)}
                    className="flex-1 h-10 border border-slate-200 rounded-md px-3 text-sm outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>

                {/* 本通路干预次数 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">本通路干预次数:</label>
                  <input
                    type="number"
                    value={formData.interventionCount}
                    onChange={(e) => handleChange('interventionCount', parseInt(e.target.value) || 0)}
                    className="flex-1 h-10 border border-slate-200 rounded-md px-3 text-sm outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>

                {/* 通路干预日期 */}
                <div className="flex items-center">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">通路干预日期:</label>
                  <div className="relative flex-1">
                    <DatePicker
                      value={formData.interventionDate ? dayjs(formData.interventionDate) : null}
                      onChange={(date) => handleChange('interventionDate', date ? date.format('YYYY-MM-DD') : '')}
                      format="YYYY年MM月DD日"
                      placeholder="请选择通路干预日期"
                      style={{ width: '100%', height: 40 }}
                      size="middle"
                    />
                  </div>
                </div>

                {/* A侧穿刺点位置（仅内瘘显示）- 多项输入 */}
                {!isCatheter && (
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">A侧穿刺点位置:</label>
                    <div className="flex-1">
                      <Select
                        mode="tags"
                        allowClear
                        placeholder="输入位置后按回车"
                        value={formData.aPuncturePosition}
                        onChange={(value) => handleChange('aPuncturePosition', value)}
                        style={{ width: '100%' }}
                        size="middle"
                        tokenSeparators={[',']}
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* 底部备注区域 */}
              <div className="col-span-2">
                <div className="flex items-start">
                  <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4 mt-2">备注:</label>
                  <textarea
                    value={formData.remark}
                    onChange={(e) => handleChange('remark', e.target.value)}
                    className="flex-1 h-20 border border-slate-200 rounded-md p-3 text-sm outline-none focus:ring-1 focus:ring-blue-500 resize-none"
                    placeholder="输入备注信息"
                  ></textarea>
                </div>
              </div>

              {/* 开关及图片 */}
              <div className="col-span-2 grid grid-cols-3 items-start">
                <div className="flex flex-col gap-3">
                  <div className="flex items-center">
                    <label className="w-32 text-right text-sm font-medium text-slate-600 mr-4">图片:</label>
                    <button
                      onClick={handleUploadClick}
                      disabled={uploading}
                      className="flex items-center gap-2 px-4 py-1.5 border border-slate-200 rounded-md text-slate-600 text-xs font-bold hover:bg-slate-50 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <Upload size={14} /> {uploading ? '上传中...' : '上传图片'}
                    </button>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      multiple
                      onChange={handleFileChange}
                      className="hidden"
                    />
                  </div>
                  {/* 图片预览 */}
                  {formData.images.length > 0 && (
                    <div className="flex items-center gap-2 ml-36">
                      {formData.images.map((img, idx) => (
                        <div key={idx} className="relative group">
                          <img
                            src={img}
                            alt={`图片${idx + 1}`}
                            className="w-16 h-16 object-cover rounded-md border border-slate-200"
                          />
                          <button
                            onClick={() => handleRemoveImage(idx)}
                            className="absolute -top-2 -right-2 w-5 h-5 bg-red-500 text-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center hover:bg-red-600"
                          >
                            <XCircle size={12} />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
                <div className="flex items-center justify-center">
                  <label className="text-sm font-medium text-slate-600 mr-4">是否默认:</label>
                  <input
                    type="checkbox"
                    checked={formData.isDefault}
                    onChange={(e) => handleChange('isDefault', e.target.checked)}
                    className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-0"
                  />
                </div>
                <div className="flex items-center justify-end pr-12">
                  <label className="text-sm font-medium text-slate-600 mr-4">是否禁用:</label>
                  <input
                    type="checkbox"
                    checked={formData.isDisabled}
                    onChange={(e) => handleChange('isDisabled', e.target.checked)}
                    className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-0"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Footer */}
          <div className="mt-12 flex justify-end gap-3 border-t border-slate-100 pt-8">
            <button
              onClick={onClose}
              disabled={saving}
              className="px-10 py-2 rounded-md border border-slate-200 text-slate-600 text-sm font-bold hover:bg-slate-50 transition-all disabled:opacity-50"
            >
              取消
            </button>
            <button
              onClick={handleSave}
              disabled={saving || !formData.accessType}
              className="px-10 py-2 rounded-md bg-blue-600 text-white text-sm font-bold shadow-lg hover:bg-blue-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {saving && <span className="animate-spin inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full"></span>}
              {saving ? '保存中...' : '保存设置'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
