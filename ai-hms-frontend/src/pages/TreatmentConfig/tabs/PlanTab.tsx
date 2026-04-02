// PlanTab - 方案模板管理
// 自包含组件，包含所有方案模板的状态管理和 CRUD 操作

import { useState, useEffect, useCallback, memo } from 'react'
import {
  FileText, Search, Plus, Edit3, Trash2, ChevronRight,
  Clock, X, Info, Loader2
} from 'lucide-react'
import {
  planTemplateApi,
  drugCatalogApi,
  type PlanTemplate
} from '@/services/treatmentConfigApi'
import { DICT_TYPES } from '@/services/dictApi'
import { InputField } from '../components'
import {
  NumericPad,
  useNumericPad,
  DialysisModeSection,
  AnticoagulantSection,
  DialysisParamsSection,
  MaterialsSection,
  type DialysisModeValues,
  type AnticoagulantValues,
  type DialysisParamsValues,
  type MaterialItem,
} from '@/components/treatment-form'

// --- Props 接口 ---
interface PlanTabProps {
  dictOptions: Record<string, Array<{ value: string; label: string }>>
  onRefreshDict: () => Promise<void>
}

// --- 主组件 ---
function PlanTabComponent({ dictOptions, onRefreshDict }: PlanTabProps) {
  const [planTemplates, setPlanTemplates] = useState<PlanTemplate[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [anticoagulantOptions, setAnticoagulantOptions] = useState<Array<{ value: string; label: string }>>([])
  const [anticoagulantSpecMap, setAnticoagulantSpecMap] = useState<Record<string, string>>({})
  const [showModal, setShowModal] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [searchKeyword, setSearchKeyword] = useState('')

  // 使用共享的数字键盘 hook
  const { padState, openNumericPad, closeNumericPad, handleKeyPress, handleClear, handleConfirm } = useNumericPad()

  // 方案模板表单状态
  const [planForm, setPlanForm] = useState<{
    id: string
    name: string
    method: string
    time: string
    bloodFlow: string
    substituteInputMode: string
    substituteFlow: string
    substituteVolume: string
    note: string
    initialAnticoag: string
    initialDose: string
    maintenanceAnticoag: string
    infusionRate: string
    infusionTime: string
    maintenanceDose: string
    totalDose: string
    dialysateType: string
    dialysateGroup: string
    dialysateFlow: string
    na: string
    ca: string
    k: string
    hco3: string
    glucose: string
    conductivity: string
    temp: string
    volume: string
    category: string
    description: string
  }>({
    id: '',
    name: '',
    method: '',
    time: '4.0',
    bloodFlow: '250',
    substituteInputMode: '',
    substituteFlow: '',
    substituteVolume: '',
    note: '',
    initialAnticoag: '',
    initialDose: '',
    maintenanceAnticoag: '',
    infusionRate: '',
    infusionTime: '',
    maintenanceDose: '',
    totalDose: '',
    dialysateType: '',
    dialysateGroup: '',
    dialysateFlow: '',
    na: '140',
    ca: '1.5',
    k: '2.0',
    hco3: '35',
    glucose: '',
    conductivity: '',
    temp: '37.0',
    volume: '',
    category: '常规',
    description: ''
  })

  // 模态框中的材料列表
  const [modalMaterials, setModalMaterials] = useState<MaterialItem[]>([])

  // 初始表单状态
  const getInitialPlanForm = () => ({
    id: '',
    name: '',
    method: '',
    time: '4.0',
    bloodFlow: '250',
    substituteInputMode: '',
    substituteFlow: '',
    substituteVolume: '',
    note: '',
    initialAnticoag: '',
    initialDose: '',
    maintenanceAnticoag: '',
    infusionRate: '',
    infusionTime: '',
    maintenanceDose: '',
    totalDose: '',
    dialysateType: '',
    dialysateGroup: '',
    dialysateFlow: '',
    na: '140',
    ca: '1.5',
    k: '2.0',
    hco3: '35',
    glucose: '',
    conductivity: '',
    temp: '37.0',
    volume: '',
    category: '常规',
    description: ''
  })

  // 加载数据
  const loadData = async () => {
    setIsLoading(true)
    try {
      const response = await planTemplateApi.list({ pageSize: 9999 })
      setPlanTemplates(response.items)
    } catch (error) {
      console.error('加载方案模板失败:', error)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const isNoHeparinOption = (value: string) => value === '相对无肝素' || value === '绝对无肝素'
  const sanitizeDecimalInput = (value: string) => {
    if (value === '') return ''
    const normalized = value
      .replace(/[０-９]/g, (d) => String.fromCharCode(d.charCodeAt(0) - 0xff10 + 0x30))
      .replace(/[．。]/g, '.')
    let cleaned = normalized.replace(/[^0-9.]/g, '')
    const firstDot = cleaned.indexOf('.')
    if (firstDot !== -1) {
      cleaned = cleaned.slice(0, firstDot + 1) + cleaned.slice(firstDot + 1).replace(/\./g, '')
    }
    return cleaned
  }
  const parseDialysateGroupIons = (groupValue: string) => {
    const group = (groupValue || '').trim()
    if (!group) {
      return { k: '', ca: '' }
    }

    const kMatch = group.match(/K\s*([0-9]+(?:\.[0-9]+)?)/i)
    const caMatch = group.match(/Ca\s*([0-9]+(?:\.[0-9]+)?)/i)

    return {
      k: kMatch ? sanitizeDecimalInput(kMatch[1]) : '',
      ca: caMatch ? sanitizeDecimalInput(caMatch[1]) : ''
    }
  }
  const extractDoseUnit = (concentration?: string) => {
    const trimmed = concentration?.trim()
    if (!trimmed) return ''
    const match = trimmed.match(/[a-zA-Zμµ]+/g)
    if (!match || match.length === 0) return ''
    return match[0].replace('µ', 'u')
  }

  // --- Section 适配器 ---
  // DialysisModeSection 值桥接：planForm ↔ DialysisModeValues
  const dialysisModeValues: DialysisModeValues = {
    method: planForm.method,
    duration: planForm.time,
    bloodFlow: planForm.bloodFlow,
    bv: '',
    substituteMode: planForm.substituteInputMode,
    substituteFlow: planForm.substituteFlow,
    substituteVolume: planForm.substituteVolume,
    notes: planForm.note,
  }

  const handleDialysisModeChange = useCallback(
    (updater: (prev: DialysisModeValues) => DialysisModeValues) => {
      setPlanForm((prev) => {
        const modeValues: DialysisModeValues = {
          method: prev.method,
          duration: prev.time,
          bloodFlow: prev.bloodFlow,
          bv: '',
          substituteMode: prev.substituteInputMode,
          substituteFlow: prev.substituteFlow,
          substituteVolume: prev.substituteVolume,
          notes: prev.note,
        }
        const next = updater(modeValues)
        return {
          ...prev,
          method: next.method,
          time: next.duration,
          bloodFlow: next.bloodFlow,
          substituteInputMode: next.substituteMode,
          substituteFlow: next.substituteFlow,
          substituteVolume: next.substituteVolume,
          note: next.notes,
        }
      })
    },
    []
  )

  const handleDialysateVolumeChange = useCallback((volume: string) => {
    setPlanForm((prev) => ({ ...prev, volume }))
  }, [])

  // AnticoagulantSection 值桥接：planForm ↔ AnticoagulantValues
  const anticoagulantValues: AnticoagulantValues = {
    initialDrug: planForm.initialAnticoag,
    initialDose: planForm.initialDose,
    maintenanceDrug: planForm.maintenanceAnticoag,
    infusionRate: planForm.infusionRate,
    infusionTime: planForm.infusionTime,
    maintenanceDose: planForm.maintenanceDose,
    totalDose: planForm.totalDose,
  }

  const handleAnticoagulantChange = useCallback(
    (updater: (prev: AnticoagulantValues) => AnticoagulantValues) => {
      setPlanForm((prev) => {
        const acValues: AnticoagulantValues = {
          initialDrug: prev.initialAnticoag,
          initialDose: prev.initialDose,
          maintenanceDrug: prev.maintenanceAnticoag,
          infusionRate: prev.infusionRate,
          infusionTime: prev.infusionTime,
          maintenanceDose: prev.maintenanceDose,
          totalDose: prev.totalDose,
        }
        const next = updater(acValues)
        return {
          ...prev,
          initialAnticoag: next.initialDrug,
          initialDose: next.initialDose,
          maintenanceAnticoag: next.maintenanceDrug,
          infusionRate: next.infusionRate,
          infusionTime: next.infusionTime,
          maintenanceDose: next.maintenanceDose,
          totalDose: next.totalDose,
        }
      })
    },
    []
  )

  // DialysisParamsSection 值桥接：planForm ↔ DialysisParamsValues
  const dialysisParamsValues: DialysisParamsValues = {
    dialysateType: planForm.dialysateType,
    dialysateGroup: planForm.dialysateGroup,
    dialysateFlow: planForm.dialysateFlow,
    na: planForm.na,
    ca: planForm.ca,
    k: planForm.k,
    hco3: planForm.hco3,
    glucose: planForm.glucose,
    conductivity: planForm.conductivity,
    temp: planForm.temp,
    dialysateVolume: planForm.volume,
  }

  const handleDialysisParamsChange = useCallback(
    (updater: (prev: DialysisParamsValues) => DialysisParamsValues) => {
      setPlanForm((prev) => {
        const paramsValues: DialysisParamsValues = {
          dialysateType: prev.dialysateType,
          dialysateGroup: prev.dialysateGroup,
          dialysateFlow: prev.dialysateFlow,
          na: prev.na,
          ca: prev.ca,
          k: prev.k,
          hco3: prev.hco3,
          glucose: prev.glucose,
          conductivity: prev.conductivity,
          temp: prev.temp,
          dialysateVolume: prev.volume,
        }
        const next = updater(paramsValues)
        return {
          ...prev,
          dialysateType: next.dialysateType,
          dialysateGroup: next.dialysateGroup,
          dialysateFlow: next.dialysateFlow,
          na: next.na,
          ca: next.ca,
          k: next.k,
          hco3: next.hco3,
          glucose: next.glucose,
          conductivity: next.conductivity,
          temp: next.temp,
          volume: next.dialysateVolume,
        }
      })
    },
    []
  )

  // MaterialsSection 回调
  const handleMaterialAdd = useCallback(
    (name: string, material: Partial<MaterialItem>) => {
      setModalMaterials((prev) => {
        const exists = prev.some((m) => m.name === name)
        if (exists) return prev
        return [
          ...prev,
          {
            id: `M${Date.now()}`,
            name,
            category: material.category || '',
            count: 1,
            code: material.code || '',
            brand: material.brand || '',
            spec: material.spec || '',
            note: '',
          },
        ]
      })
    },
    []
  )

  const handleMaterialRemove = useCallback((id: string) => {
    setModalMaterials((prev) => prev.filter((m) => m.id !== id))
  }, [])

  const handleMaterialUpdate = useCallback(
    (id: string, field: string, value: string | number) => {
      setModalMaterials((prev) =>
        prev.map((m) => (m.id === id ? { ...m, [field]: value } : m))
      )
    },
    []
  )

  const handleMaterialReplace = useCallback(
    (id: string, patch: Partial<MaterialItem>) => {
      setModalMaterials((prev) =>
        prev.map((m) => (m.id === id ? { ...m, ...patch } : m))
      )
    },
    []
  )

  const loadAnticoagulantOptions = async () => {
    try {
      const response = await drugCatalogApi.list({ pageSize: 9999, isEnabled: true })
      const categoryMap = new Map(
        (dictOptions[DICT_TYPES.DRUG_CATEGORY] || []).map(item => [item.value, item.label])
      )
      const filtered = response.items.filter((item) => {
        const label = categoryMap.get(item.category) || item.category
        return label === '抗凝剂' || label === '抗凝药' || item.category === '抗凝剂' || item.category === '抗凝药' || item.category === 'ANTICOAGULANT'
      })
      const sorted = [...filtered].sort((a, b) => {
        const aOrder = a.sortOrder && a.sortOrder > 0 ? a.sortOrder : 9999
        const bOrder = b.sortOrder && b.sortOrder > 0 ? b.sortOrder : 9999
        if (aOrder !== bOrder) return aOrder - bOrder
        return a.name.localeCompare(b.name)
      })
      setAnticoagulantOptions(sorted.map(item => ({ value: item.name, label: item.name })))
      setAnticoagulantSpecMap(
        sorted.reduce<Record<string, string>>((acc, item) => {
          const unit = extractDoseUnit(item.concentration)
          if (unit) {
            acc[item.name] = unit
          }
          return acc
        }, {})
      )
    } catch (error) {
      console.error('加载抗凝剂药品失败:', error)
    }
  }

  // 打开新增弹窗
  const handleOpenAddModal = async () => {
    await onRefreshDict()
    await loadAnticoagulantOptions()
    closeNumericPad()
    setIsEditing(false)
    setPlanForm(getInitialPlanForm())
    setModalMaterials([])
    setShowModal(true)
  }

  // 打开编辑弹窗
  const handleOpenEditModal = async (plan: PlanTemplate) => {
    setIsEditing(true)
    await onRefreshDict()
    await loadAnticoagulantOptions()
    closeNumericPad()

    try {
      const planDetail = await planTemplateApi.get(plan.id)

      const initialAnticoag = planDetail.templateContent.anticoagulant.initialDrug || '肝素钠'
      const rawMaintenanceAnticoag = planDetail.templateContent.anticoagulant.maintenanceDrug || ''
      const isInitialNoHeparin = isNoHeparinOption(initialAnticoag)
      const isSameNoHeparin = isInitialNoHeparin && rawMaintenanceAnticoag === initialAnticoag
      const maintenanceAnticoag = isSameNoHeparin ? '' : rawMaintenanceAnticoag
      const shouldClearMaintenance = isSameNoHeparin || isNoHeparinOption(maintenanceAnticoag)
      const dialysateGroup = planDetail.templateContent.parameters.dialysateGroup || ''
      const parsedIons = parseDialysateGroupIons(dialysateGroup)
      setPlanForm({
        id: planDetail.id,
        name: planDetail.name || '',
        method: planDetail.templateContent.dialysisMode.mode || 'HD',
        time: planDetail.templateContent.duration?.toString() || '4.0',
        bloodFlow: planDetail.templateContent.dialysisMode.bloodFlow?.toString() || '250',
        substituteInputMode: planDetail.templateContent.dialysisMode.substituteInputMode || '',
        substituteFlow: planDetail.templateContent.dialysisMode.substituteFlow?.toString() || '',
        substituteVolume: planDetail.templateContent.dialysisMode.substituteVolume?.toString() || '',
        note: planDetail.templateContent.dialysisMode.notes || '',
        initialAnticoag,
        initialDose: isInitialNoHeparin ? '' : (planDetail.templateContent.anticoagulant.initialDose || ''),
        maintenanceAnticoag,
        infusionRate: shouldClearMaintenance ? '' : (planDetail.templateContent.anticoagulant.infusionRate || ''),
        infusionTime: shouldClearMaintenance ? '' : (planDetail.templateContent.anticoagulant.infusionTime || ''),
        maintenanceDose: shouldClearMaintenance ? '' : (planDetail.templateContent.anticoagulant.maintenanceDose || ''),
        totalDose: shouldClearMaintenance ? '' : (planDetail.templateContent.anticoagulant.totalDose || ''),
        dialysateType: planDetail.templateContent.parameters.dialysateType || '',
        dialysateGroup,
        dialysateFlow: planDetail.templateContent.parameters.flowRate?.toString() || '',
        na: planDetail.templateContent.parameters.na?.toString() || '140',
        ca: parsedIons.ca || planDetail.templateContent.parameters.ca?.toString() || '1.5',
        k: parsedIons.k || planDetail.templateContent.parameters.k?.toString() || '2.0',
        hco3: planDetail.templateContent.parameters.hco3?.toString() || '35',
        glucose: planDetail.templateContent.parameters.glucose || '无糖',
        conductivity: planDetail.templateContent.parameters.conductivity?.toString() || '',
        temp: planDetail.templateContent.parameters.temp?.toString() || '37.0',
        volume: planDetail.templateContent.parameters.volume?.toString() || '',
        category: planDetail.category || '常规',
        description: planDetail.description || ''
      })

      setModalMaterials(planDetail.templateContent.materials.map(m => ({
        id: m.id || `M${Date.now()}_${Math.random()}`,
        name: m.name,
        category: m.category,
        count: m.count,
        code: m.code,
        brand: m.brand,
        spec: m.spec,
        note: m.note
      })))

      setShowModal(true)
    } catch (error) {
      console.error('加载方案详情失败:', error)
      alert('加载方案详情失败，请稍后重试')
    }
  }

  // 保存方案
  const handleSave = async () => {
    const missingFields = []
    const isInitialNoHeparin = isNoHeparinOption(planForm.initialAnticoag)
    const isMaintenanceNoHeparin = isNoHeparinOption(planForm.maintenanceAnticoag)
    if (!planForm.name?.trim()) missingFields.push('模版名称')
    if (!planForm.method?.trim()) missingFields.push('透析方法')
    if (!planForm.time?.trim()) missingFields.push('透析时间')
    if (!planForm.bloodFlow?.trim()) missingFields.push('标准血流量')
    if (!planForm.initialAnticoag?.trim()) missingFields.push('首剂名称')
    if (!isInitialNoHeparin && !planForm.initialDose?.trim()) missingFields.push('首剂量')
    if (!planForm.dialysateType?.trim()) missingFields.push('透析液分类')
    if (!planForm.dialysateGroup?.trim()) missingFields.push('透析液分组')
    if (!planForm.dialysateFlow?.trim()) missingFields.push('透析液流速')

    if (missingFields.length > 0) {
      alert(`缺少必填字段：${missingFields.join('、')}`)
      return
    }

    setIsLoading(true)
    try {
      const templateContent = {
        weeklyFrequency: 3,
        biweeklyFrequency: 0,
        duration: parseFloat(planForm.time) || 4,
        dryWeight: 65,
        dialysisMode: {
          mode: planForm.method,
          bloodFlow: parseInt(planForm.bloodFlow) || 250,
          substituteInputMode: planForm.substituteInputMode,
          substituteFlow: parseFloat(planForm.substituteFlow) || 0,
          substituteVolume: parseFloat(planForm.substituteVolume) || 0,
          bv: '',
          frequencyDesc: '一周三次',
          autoConfirm: false,
          status: 'active',
          notes: planForm.note
        },
        anticoagulant: {
          initialDrug: planForm.initialAnticoag,
          initialDose: isInitialNoHeparin ? '' : planForm.initialDose,
          totalDose: isMaintenanceNoHeparin ? '' : planForm.totalDose,
          maintenanceDrug: planForm.maintenanceAnticoag,
          infusionRate: isMaintenanceNoHeparin ? '' : planForm.infusionRate,
          infusionTime: isMaintenanceNoHeparin ? '' : planForm.infusionTime,
          maintenanceDose: isMaintenanceNoHeparin ? '' : planForm.maintenanceDose
        },
        parameters: {
          dialysateType: planForm.dialysateType,
          dialysateGroup: planForm.dialysateGroup,
          flowRate: parseInt(planForm.dialysateFlow) || 0,
          na: parseFloat(planForm.na) || 140,
          ca: parseFloat(planForm.ca) || 1.5,
          k: parseFloat(planForm.k) || 2.0,
          hco3: parseFloat(planForm.hco3) || 35,
          glucose: planForm.glucose,
          conductivity: parseFloat(planForm.conductivity) || 14,
          temp: parseFloat(planForm.temp) || 37,
          volume: parseFloat(planForm.volume) || 0
        },
        materials: modalMaterials.filter(m => m.name).map(m => ({
          id: m.id,
          name: m.name,
          category: m.category,
          count: m.count,
          code: m.code,
          brand: m.brand,
          spec: m.spec,
          note: m.note
        }))
      }

      if (isEditing && planForm.id) {
        await planTemplateApi.update(planForm.id, {
          name: planForm.name,
          description: planForm.description || planForm.note,
          mode: planForm.method as 'HD' | 'HFD' | 'HP' | 'HF' | 'HDF',
          category: planForm.category,
          templateContent
        })
      } else {
        await planTemplateApi.create({
          name: planForm.name,
          description: planForm.description || planForm.note,
          mode: planForm.method as 'HD' | 'HFD' | 'HP' | 'HF' | 'HDF',
          category: planForm.category,
          isEnabled: true,
          templateContent
        })
      }

      await loadData()
      closeNumericPad()
      setShowModal(false)
      alert(isEditing ? '方案模板更新成功！' : '方案模板保存成功！')
    } catch (error) {
      console.error('保存方案失败:', error)
      alert('保存失败，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }

  // 删除方案
  const handleDelete = async (plan: PlanTemplate) => {
    if (plan.isDefault) {
      alert('默认方案不能删除')
      return
    }
    if (window.confirm(`确定要删除方案"${plan.name}"吗？`)) {
      setIsLoading(true)
      try {
        await planTemplateApi.delete(plan.id)
        await loadData()
        alert('方案删除成功！')
      } catch (error) {
        console.error('删除方案失败:', error)
        alert('删除失败，请稍后重试')
      } finally {
        setIsLoading(false)
      }
    }
  }

  // 过滤数据
  const filteredTemplates = planTemplates.filter(plan =>
    !searchKeyword || plan.name.toLowerCase().includes(searchKeyword.toLowerCase())
  )

  if (isLoading && planTemplates.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-slate-400 animate-spin" />
        <span className="ml-3 text-slate-500">加载中...</span>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col animate-fade-in">
      {/* 顶部工具栏 */}
      <div className="bg-white p-4 border-b border-slate-100 flex justify-between items-center shrink-0">
        <div className="relative w-80">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"/>
          <input
            type="text"
            placeholder="搜索方案模板..."
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all"
          />
        </div>
        <button onClick={handleOpenAddModal} className="flex items-center gap-2 px-5 py-2 bg-blue-600 text-white rounded-xl text-xs font-black hover:bg-blue-700 shadow-lg shadow-blue-100 transition-all">
          <Plus size={14}/> 新增方案
        </button>
      </div>

      {/* 列表区域 */}
      <div className="flex-1 overflow-auto p-6">
        {filteredTemplates.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-slate-400">
            <FileText size={48} className="text-slate-300 mb-4" />
            <p className="text-sm font-bold">暂无方案模板</p>
            <p className="text-xs mt-1">点击"新增方案"创建第一个模板</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredTemplates.map(plan => (
              <div key={plan.id} className="bg-white rounded-[32px] border border-slate-100 p-5 shadow-sm hover:shadow-md transition-all group hover:border-blue-200 relative overflow-hidden flex flex-col h-[210px] justify-between">
                <div>
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-blue-50 text-blue-600 rounded-xl group-hover:bg-blue-600 group-hover:text-white transition-colors duration-300">
                        <FileText size={18}/>
                      </div>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-tighter ${
                        plan.mode === 'HD' ? 'bg-blue-50 text-blue-600' :
                        plan.mode === 'HDF' ? 'bg-purple-50 text-purple-600' :
                        plan.mode === 'HF' ? 'bg-green-50 text-green-600' :
                        'bg-orange-50 text-orange-600'
                      }`}>{plan.mode}</span>
                    </div>
                    <span className="text-[10px] text-slate-300 font-bold font-mono uppercase tracking-widest">{plan.id.slice(0, 8)}</span>
                  </div>
                  <h4 className="text-[15px] font-black text-slate-800 mb-1 leading-tight group-hover:text-blue-700 transition-colors">{plan.name}</h4>
                  <p className="text-[11px] text-slate-400 font-bold mb-4">{plan.description || '暂无描述'}</p>
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] px-2 py-0.5 bg-slate-100 text-slate-500 rounded">{plan.category}</span>
                    {plan.isDefault && <span className="text-[10px] px-2 py-0.5 bg-blue-100 text-blue-600 rounded">默认</span>}
                  </div>
                </div>
                <div className="pt-4 border-t border-slate-50 flex items-center justify-between">
                  <div className="flex items-center gap-1.5 text-[10px] text-slate-300 font-bold ml-1">
                    <Clock size={11} className="shrink-0" />
                    {new Date(plan.createdAt).toLocaleDateString()}
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleDelete(plan)}
                      disabled={plan.isDefault}
                      className={`px-3 py-1.5 rounded-xl text-[11px] font-black transition-all flex items-center gap-1 ${
                        plan.isDefault
                          ? 'bg-slate-100 text-slate-300 cursor-not-allowed'
                          : 'bg-red-50 text-red-600 hover:bg-red-600 hover:text-white'
                      }`}
                      title={plan.isDefault ? '默认方案不能删除' : '删除方案'}
                    >
                      <Trash2 size={14} />
                    </button>
                    <button onClick={() => handleOpenEditModal(plan)} className="px-4 py-1.5 bg-blue-50 text-blue-600 rounded-xl text-[11px] font-black hover:bg-blue-600 hover:text-white transition-all flex items-center gap-1">
                      编辑 <ChevronRight size={14} className="group-hover:translate-x-0.5 transition-transform"/>
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 方案模版弹窗 */}
      {showModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[48px] shadow-2xl w-full max-w-7xl max-h-[95vh] overflow-hidden flex flex-col animate-scale-in">
            <div className="px-10 py-6 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className={`w-12 h-12 ${isEditing ? 'bg-amber-500' : 'bg-blue-600'} rounded-2xl flex items-center justify-center text-white shadow-lg shadow-blue-200`}>
                  {isEditing ? <Edit3 size={24} strokeWidth={3} /> : <Plus size={24} strokeWidth={3} />}
                </div>
                <div>
                  <h3 className="text-xl font-black text-slate-900 tracking-tight">{isEditing ? '编辑方案' : '新增方案'}</h3>
                  <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">Define New Dialysis Treatment Template</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={() => {
                    closeNumericPad()
                    setShowModal(false)
                  }}
                  className="p-3 hover:bg-slate-100 rounded-2xl text-slate-400 hover:text-slate-900 transition-all"
                >
                  <X size={24} />
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-10 space-y-10 no-scrollbar bg-white">
              <div className="space-y-6">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-black text-blue-600 uppercase tracking-widest flex items-center gap-2">
                    <Info size={16}/> 01. 基本配置
                  </h4>
                  <div className="w-1/3">
                    <InputField label="模版名称" placeholder="请输入方案名称" required value={planForm.name} onChange={(e) => setPlanForm({...planForm, name: e.target.value})}/>
                  </div>
                </div>
                <DialysisModeSection
                  values={dialysisModeValues}
                  onChange={handleDialysisModeChange}
                  dictOptions={dictOptions}
                  dictTypeKey={DICT_TYPES.DIALYSIS_MODE}
                  openNumericPad={openNumericPad}
                  dialysateFlow={planForm.dialysateFlow}
                  onDialysateVolumeChange={handleDialysateVolumeChange}
                />
              </div>
              <AnticoagulantSection
                values={anticoagulantValues}
                onChange={handleAnticoagulantChange}
                drugOptions={anticoagulantOptions}
                specMap={anticoagulantSpecMap}
                openNumericPad={openNumericPad}
              />
              <DialysisParamsSection
                values={dialysisParamsValues}
                onChange={handleDialysisParamsChange}
                duration={planForm.time}
                dictOptions={dictOptions}
                dictTypeKeys={{
                  dialysateType: DICT_TYPES.DIALYSATE_TYPE,
                  dialysateGroup: DICT_TYPES.DIALYSATE_GROUP,
                  dialysateFlow: DICT_TYPES.DIALYSATE_FLOW,
                  glucose: DICT_TYPES.GLUCOSE,
                }}
                openNumericPad={openNumericPad}
              />
              <MaterialsSection
                materials={modalMaterials}
                onAdd={handleMaterialAdd}
                onRemove={handleMaterialRemove}
                onUpdate={handleMaterialUpdate}
                onReplace={handleMaterialReplace}
              />
            </div>
            <div className="px-10 py-6 bg-slate-50 border-t border-slate-100 flex justify-end gap-3 shrink-0">
              <button
                onClick={() => {
                  closeNumericPad()
                  setShowModal(false)
                }}
                className="px-8 py-2.5 bg-white border border-slate-200 text-slate-600 rounded-2xl text-xs font-black hover:bg-slate-100 transition-all"
              >
                取消
              </button>
              <button onClick={handleSave} disabled={isLoading} className={`px-12 py-2.5 ${isEditing ? 'bg-amber-500 hover:bg-amber-600 shadow-amber-100' : 'bg-blue-600 hover:bg-blue-700 shadow-blue-100'} text-white rounded-2xl text-xs font-black shadow-xl transition-all disabled:opacity-50`}>
                {isLoading ? <Loader2 size={16} className="animate-spin mr-2 inline" /> : ''}保存方案
              </button>
            </div>
          </div>
          <NumericPad
            open={padState.open}
            label={padState.label}
            value={padState.value}
            onKeyPress={handleKeyPress}
            onConfirm={handleConfirm}
            onClear={handleClear}
            onClose={closeNumericPad}
          />
        </div>
      )}
    </div>
  )
}

export const PlanTab = memo(PlanTabComponent)
