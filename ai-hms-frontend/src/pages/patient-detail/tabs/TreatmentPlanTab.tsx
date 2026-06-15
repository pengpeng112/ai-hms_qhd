// TreatmentPlanTab - 治疗方案 Tab
// 基于 UI 设计稿 v1.3 重构

import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { message, Modal } from 'antd'
import {
  ClipboardList, PlusCircle, RotateCcw,
  Save, Search, Sparkles, X, Check, AlertTriangle
} from 'lucide-react'
import MaterialSyncModal from '@/components/patient/modals/MaterialSyncModal'
import type { MaterialSyncResult } from '@/components/patient/modals/MaterialSyncModal'
import {
  DialysisModeSection,
  AnticoagulantSection,
  DialysisParamsSection,
  MaterialsSection,
  NumericPad,
  useNumericPad,
} from '@/components/treatment-form'
import type { DialysisModeValues, AnticoagulantValues, DialysisParamsValues } from '@/components/treatment-form'
import {
  planTemplateApi,
  drugCatalogApi,
  type PlanTemplate,
  type DrugCatalog
} from '@/services/treatmentConfigApi'
import { patientApi, type TreatmentPlan as ApiTreatmentPlan, type AdjustmentRecord } from '@/services/patientApi'
import { restApi, type VascularAccessApi } from '@/services/restClient'
import { getErrorMessage } from '@/services/restClient'
import { dictCache, DICT_TYPES } from '@/services/dictApi'

// 类型定义
interface MaterialItem {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
  note: string
}

interface TreatmentPlan {
  weeklyFrequency: number
  biweeklyFrequency: number
  duration: number
  dryWeight: number
  extraWeight: number
  vascularAccess: string
  indicators: {
    mode: string
    bloodFlow: number
    bv: string
    frequencyDesc: string
    autoConfirm: boolean
    status: string
    notes: string
    substituteMode: string
    substituteFlow: number
    substituteVolume: number
  }
  anticoagulant: {
    initialDrug: string
    initialDose: string
    totalDose: string
    maintenanceDrug: string
    infusionRate: string
    infusionTime: string
    maintenanceDose: string
  }
  parameters: {
    dialysateType: string
    dialysateGroup: string
    flowRate: number
    na: number
    ca: number
    k: number
    hco3: number
    glucose: string
    conductivity: number
    temp: number
    volume: number
  }
  materials: MaterialItem[]
  adjustmentHistory?: Array<{
    id: string
    date: string
    content: string
    operator: string
    reason: string
  }>
}

interface TreatmentPlanTabProps {
  patientId?: string
  patientName?: string
  treatmentPlan?: TreatmentPlan | null
}

// 辅助函数：从药品浓度中提取剂量单位
const extractDoseUnit = (concentration?: string) => {
  const trimmed = concentration?.trim()
  if (!trimmed) return ''
  const match = trimmed.match(/[a-zA-Zμµ]+/g)
  if (!match || match.length === 0) return ''
  return match[0].replace('µ', 'u')
}

// NewPlanModal 组件
interface NewPlanModalProps {
  isOpen: boolean
  onClose: () => void
  patientName?: string
  onSave?: (data: TreatmentPlan) => Promise<void>
}

const NewPlanModal = ({ isOpen, onClose, patientName = '', onSave }: NewPlanModalProps) => {
  const [templateSearch, setTemplateSearch] = useState('')
  const [isTemplateDropdownOpen, setIsTemplateDropdownOpen] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<PlanTemplate | null>(null)

  // 空的默认表单数据（weeklyFrequency/biweeklyFrequency/dryWeight/extraWeight
  // 在弹窗中不展示，保存时由 onSave 调用方从主方案继承）
  const defaultFormData: TreatmentPlan = {
    weeklyFrequency: 0,
    biweeklyFrequency: 0,
    duration: 4,
    dryWeight: 0,
    extraWeight: 0,
    vascularAccess: '-',
    indicators: {
      mode: 'HD',
      bloodFlow: 200,
      bv: '',
      frequencyDesc: '',
      autoConfirm: false,
      status: '启用',
      notes: '',
      substituteMode: '',
      substituteFlow: 0,
      substituteVolume: 0,
    },
    anticoagulant: {
      initialDrug: '',
      initialDose: '',
      totalDose: '',
      maintenanceDrug: '',
      infusionRate: '',
      infusionTime: '',
      maintenanceDose: ''
    },
    parameters: {
      dialysateType: '',
      dialysateGroup: '',
      flowRate: 500,
      na: 140,
      ca: 1.5,
      k: 2.0, // 标准默认钾浓度（与建档草稿种子、后端默认一致）；≠2.0 时字段标红提示
      hco3: 35,
      glucose: '',
      conductivity: 14.1,
      temp: 37,
      volume: 120
    },
    materials: []
  }

  const [formData, setFormData] = useState<TreatmentPlan>(defaultFormData)
  const templateRef = useRef<HTMLDivElement>(null)

  // API 数据状态
  const [planTemplates, setPlanTemplates] = useState<PlanTemplate[]>([])

  // 字典数据状态
  const [dictOptions, setDictOptions] = useState<Record<string, Array<{ value: string; label: string }>>>({})

  // 抗凝剂选项和规格映射
  const [anticoagulantOptions, setAnticoagulantOptions] = useState<Array<{ value: string; label: string }>>([])
  const [anticoagulantSpecMap, setAnticoagulantSpecMap] = useState<Record<string, string>>({})

  // 数字键盘
  const { padState, openNumericPad, closeNumericPad, handleKeyPress, handleClear, handleConfirm } = useNumericPad()

  // 加载 API 数据
  useEffect(() => {
    const loadConfigData = async () => {
      // 重置模板搜索状态
      setTemplateSearch('')
      setSelectedTemplate(null)

      try {
        // 并行加载方案模板和药品目录
        const [templatesRes, drugsRes] = await Promise.allSettled([
          planTemplateApi.list({ isEnabled: true }),
          drugCatalogApi.list({ isEnabled: true, pageSize: 9999 })
        ])

        if (templatesRes.status === 'fulfilled') {
          setPlanTemplates(templatesRes.value.items)
        }

        // 加载字典数据（含 DRUG_CATEGORY 用于过滤抗凝剂）
        const dictTypes = [
          DICT_TYPES.DIALYSIS_MODE,
          DICT_TYPES.DIALYSATE_TYPE,
          DICT_TYPES.DIALYSATE_GROUP,
          DICT_TYPES.DIALYSATE_FLOW,
          DICT_TYPES.GLUCOSE,
          DICT_TYPES.DRUG_CATEGORY,
        ]

        const options: Record<string, Array<{ value: string; label: string }>> = {}
        for (const type of dictTypes) {
          try {
            const items = await dictCache.getOptions(type)
            options[type] = items
          } catch (error) {
            console.error(`加载字典 ${type} 失败:`, error)
          }
        }
        setDictOptions(options)

        // 构建抗凝剂选项和规格映射
        if (drugsRes.status === 'fulfilled') {
          const allDrugs = drugsRes.value.items
          const categoryMap = new Map(
            (options[DICT_TYPES.DRUG_CATEGORY] || []).map(item => [item.value, item.label])
          )
          const filtered = allDrugs.filter((item: DrugCatalog) => {
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
        }

      } catch (error) {
        console.error('加载配置数据失败:', error)
      }
    }

    if (isOpen) {
      loadConfigData()
    }
  }, [isOpen])

  // 过滤逻辑
  const filteredTemplates = templateSearch
    ? planTemplates.filter(t => t.name.toLowerCase().includes(templateSearch.toLowerCase()))
    : planTemplates

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (templateRef.current && !templateRef.current.contains(event.target as Node)) {
        setIsTemplateDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // --- Section 适配器：桥接 formData (TreatmentPlan) ↔ 共享组件 (all-string) ---

  // 透析模式
  const dialysisModeValues: DialysisModeValues = {
    method: formData.indicators.mode,
    duration: String(formData.duration),
    bloodFlow: String(formData.indicators.bloodFlow),
    bv: formData.indicators.bv || '',
    substituteMode: formData.indicators.substituteMode || '',
    substituteFlow: formData.indicators.substituteFlow != null ? String(formData.indicators.substituteFlow) : '',
    substituteVolume: formData.indicators.substituteVolume != null ? String(formData.indicators.substituteVolume) : '',
    notes: formData.indicators.notes || '',
  }

  const handleDialysisModeChange = (updater: (prev: DialysisModeValues) => DialysisModeValues) => {
    setFormData(prev => {
      const modeValues: DialysisModeValues = {
        method: prev.indicators.mode,
        duration: String(prev.duration),
        bloodFlow: String(prev.indicators.bloodFlow),
        bv: prev.indicators.bv || '',
        substituteMode: prev.indicators.substituteMode || '',
        substituteFlow: prev.indicators.substituteFlow != null ? String(prev.indicators.substituteFlow) : '',
        substituteVolume: prev.indicators.substituteVolume != null ? String(prev.indicators.substituteVolume) : '',
        notes: prev.indicators.notes || '',
      }
      const next = updater(modeValues)
      const parsedDuration = parseFloat(next.duration)
      const parsedBloodFlow = parseInt(next.bloodFlow)
      const parsedSubstituteFlow = parseFloat(next.substituteFlow)
      const parsedSubstituteVolume = parseFloat(next.substituteVolume)
      return {
        ...prev,
        duration: Number.isFinite(parsedDuration) ? parsedDuration : prev.duration,
        indicators: {
          ...prev.indicators,
          mode: next.method,
          bloodFlow: Number.isFinite(parsedBloodFlow) ? parsedBloodFlow : prev.indicators.bloodFlow,
          bv: next.bv,
          notes: next.notes,
          substituteMode: next.substituteMode,
          substituteFlow: Number.isFinite(parsedSubstituteFlow) ? parsedSubstituteFlow : 0,
          substituteVolume: Number.isFinite(parsedSubstituteVolume) ? parsedSubstituteVolume : 0,
        }
      }
    })
  }

  // 抗凝剂
  const anticoagulantValues: AnticoagulantValues = {
    initialDrug: formData.anticoagulant.initialDrug,
    initialDose: formData.anticoagulant.initialDose,
    maintenanceDrug: formData.anticoagulant.maintenanceDrug,
    infusionRate: formData.anticoagulant.infusionRate,
    infusionTime: formData.anticoagulant.infusionTime,
    maintenanceDose: formData.anticoagulant.maintenanceDose,
    totalDose: formData.anticoagulant.totalDose,
  }

  const handleAnticoagulantChange = (updater: (prev: AnticoagulantValues) => AnticoagulantValues) => {
    setFormData(prev => {
      const acValues: AnticoagulantValues = {
        initialDrug: prev.anticoagulant.initialDrug,
        initialDose: prev.anticoagulant.initialDose,
        maintenanceDrug: prev.anticoagulant.maintenanceDrug,
        infusionRate: prev.anticoagulant.infusionRate,
        infusionTime: prev.anticoagulant.infusionTime,
        maintenanceDose: prev.anticoagulant.maintenanceDose,
        totalDose: prev.anticoagulant.totalDose,
      }
      const next = updater(acValues)
      return {
        ...prev,
        anticoagulant: {
          initialDrug: next.initialDrug,
          initialDose: next.initialDose,
          maintenanceDrug: next.maintenanceDrug,
          infusionRate: next.infusionRate,
          infusionTime: next.infusionTime,
          maintenanceDose: next.maintenanceDose,
          totalDose: next.totalDose,
        }
      }
    })
  }

  // 透析参数
  const dialysisParamsValues: DialysisParamsValues = {
    dialysateType: formData.parameters.dialysateType,
    dialysateGroup: formData.parameters.dialysateGroup,
    dialysateFlow: String(formData.parameters.flowRate),
    na: String(formData.parameters.na),
    ca: String(formData.parameters.ca),
    k: String(formData.parameters.k),
    hco3: String(formData.parameters.hco3),
    glucose: formData.parameters.glucose,
    conductivity: String(formData.parameters.conductivity),
    temp: String(formData.parameters.temp),
    dialysateVolume: String(formData.parameters.volume),
  }

  const handleDialysisParamsChange = (updater: (prev: DialysisParamsValues) => DialysisParamsValues) => {
    setFormData(prev => {
      const paramValues: DialysisParamsValues = {
        dialysateType: prev.parameters.dialysateType,
        dialysateGroup: prev.parameters.dialysateGroup,
        dialysateFlow: String(prev.parameters.flowRate),
        na: String(prev.parameters.na),
        ca: String(prev.parameters.ca),
        k: String(prev.parameters.k),
        hco3: String(prev.parameters.hco3),
        glucose: prev.parameters.glucose,
        conductivity: String(prev.parameters.conductivity),
        temp: String(prev.parameters.temp),
        dialysateVolume: String(prev.parameters.volume),
      }
      const next = updater(paramValues)
      const pf = (v: string, fallback: number) => { const n = parseFloat(v); return Number.isFinite(n) ? n : fallback }
      return {
        ...prev,
        parameters: {
          dialysateType: next.dialysateType,
          dialysateGroup: next.dialysateGroup,
          flowRate: pf(next.dialysateFlow, prev.parameters.flowRate),
          na: pf(next.na, prev.parameters.na),
          ca: pf(next.ca, prev.parameters.ca),
          k: pf(next.k, prev.parameters.k),
          hco3: pf(next.hco3, prev.parameters.hco3),
          glucose: next.glucose,
          conductivity: pf(next.conductivity, prev.parameters.conductivity),
          temp: pf(next.temp, prev.parameters.temp),
          volume: pf(next.dialysateVolume, prev.parameters.volume),
        }
      }
    })
  }

  // 材料操作
  const handleMaterialAdd = (name: string, material: Partial<MaterialItem>) => {
    const newMaterial: MaterialItem = {
      id: `mat-${Date.now()}`,
      name,
      category: material.category || '-',
      count: 1,
      code: material.code || '-',
      brand: material.brand || '-',
      spec: material.spec || '-',
      note: ''
    }
    setFormData(prev => ({
      ...prev,
      materials: [...prev.materials, newMaterial]
    }))
  }

  const handleMaterialRemove = (id: string) => {
    setFormData(prev => ({
      ...prev,
      materials: prev.materials.filter((m: MaterialItem) => m.id !== id)
    }))
  }

  const handleMaterialUpdate = (id: string, field: string, value: string | number) => {
    setFormData(prev => ({
      ...prev,
      materials: prev.materials.map((m: MaterialItem) =>
        m.id === id ? { ...m, [field]: value } : m
      )
    }))
  }

  const handleMaterialReplace = (id: string, patch: Partial<MaterialItem>) => {
    setFormData(prev => ({
      ...prev,
      materials: prev.materials.map((m: MaterialItem) =>
        m.id === id ? { ...m, ...patch } : m
      )
    }))
  }

  const handleSelectTemplate = (tpl: PlanTemplate) => {
    // 将 API 返回的方案模板转换为前端使用的 TreatmentPlan 格式
    const apiTemplate = tpl.templateContent
    const convertedPlan: TreatmentPlan = {
      weeklyFrequency: apiTemplate.weeklyFrequency ?? 3,
      biweeklyFrequency: apiTemplate.biweeklyFrequency ?? 0,
      duration: apiTemplate.duration,
      dryWeight: apiTemplate.dryWeight ?? 0,
      extraWeight: 0,
      vascularAccess: '-',
      indicators: {
        mode: apiTemplate.dialysisMode.mode,
        bloodFlow: apiTemplate.dialysisMode.bloodFlow,
        bv: apiTemplate.dialysisMode.bv,
        frequencyDesc: apiTemplate.dialysisMode.frequencyDesc,
        autoConfirm: apiTemplate.dialysisMode.autoConfirm,
        status: apiTemplate.dialysisMode.status,
        notes: apiTemplate.dialysisMode.notes,
        substituteMode: apiTemplate.dialysisMode.substituteInputMode || '',
        substituteFlow: apiTemplate.dialysisMode.substituteFlow || 0,
        substituteVolume: apiTemplate.dialysisMode.substituteVolume || 0,
      },
      anticoagulant: {
        initialDrug: apiTemplate.anticoagulant.initialDrug,
        initialDose: apiTemplate.anticoagulant.initialDose,
        totalDose: apiTemplate.anticoagulant.totalDose,
        maintenanceDrug: apiTemplate.anticoagulant.maintenanceDrug,
        infusionRate: apiTemplate.anticoagulant.infusionRate,
        infusionTime: apiTemplate.anticoagulant.infusionTime,
        maintenanceDose: apiTemplate.anticoagulant.maintenanceDose,
      },
      parameters: {
        dialysateType: apiTemplate.parameters.dialysateType,
        dialysateGroup: apiTemplate.parameters.dialysateGroup,
        flowRate: apiTemplate.parameters.flowRate,
        na: apiTemplate.parameters.na,
        ca: apiTemplate.parameters.ca,
        k: apiTemplate.parameters.k,
        hco3: apiTemplate.parameters.hco3,
        glucose: apiTemplate.parameters.glucose,
        conductivity: apiTemplate.parameters.conductivity,
        temp: apiTemplate.parameters.temp,
        volume: apiTemplate.parameters.volume,
      },
      materials: apiTemplate.materials.map(m => ({
        id: m.id,
        name: m.name,
        category: m.category,
        count: m.count,
        code: m.code,
        brand: m.brand,
        spec: m.spec,
        note: m.note,
      })),
      adjustmentHistory: [],
    }

    setFormData(convertedPlan)
    setSelectedTemplate(tpl)
    setTemplateSearch('')
    setIsTemplateDropdownOpen(false)
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in p-4">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-6xl overflow-hidden animate-scale-in flex flex-col ring-1 ring-black/5 max-h-[92vh]">
        {/* Header */}
        <div className="bg-[#f8fbff] px-10 py-6 flex items-center justify-between border-b border-blue-100 shrink-0">
          <div>
            <h3 className="text-xl font-black text-slate-800 flex items-center gap-2">
              <PlusCircle size={24} className="text-blue-600" /> 为患者 {patientName} 新建治疗方案
            </h3>
            <p className="text-xs text-slate-400 font-bold mt-1 uppercase tracking-widest leading-none">CREATE NEW TREATMENT PLAN</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/50 rounded-full transition-all text-slate-400 hover:text-slate-600">
            <X size={24} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-10 space-y-10 custom-scrollbar">
          {/* 方案完整性提示（契约02 三 · 与后端开方/上机门禁配套）：
              草稿判定 = 透析液配方（分类或分组）为空，与后端 isLegacyPlanComplete 一致 */}
          {!(formData.parameters.dialysateType?.trim() || formData.parameters.dialysateGroup) && (
            <div className="flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4">
              <AlertTriangle size={20} className="text-amber-500 shrink-0 mt-0.5" />
              <div className="text-[13px] leading-relaxed">
                <p className="font-black text-amber-700">草稿 · 未完成</p>
                <p className="text-amber-600 font-medium mt-0.5">
                  透析液配方（分类 / 分组）尚未填写。保存后仍为草稿；
                  <b className="text-amber-700">开当日处方与上机前会被系统拦截</b>，请先补全透析液配方。
                </p>
              </div>
            </div>
          )}

          {/* 模板选择区 */}
          <div className="relative" ref={templateRef}>
            <label className="text-sm font-black text-slate-700 mb-2 flex items-center gap-2">
              <Sparkles size={16} className="text-amber-500" /> 选择方案模板
            </label>
            <div className="relative">
              <input
                type="text"
                placeholder={selectedTemplate ? selectedTemplate.name : '点击选择或输入模板名称搜索...'}
                value={isTemplateDropdownOpen ? templateSearch : (selectedTemplate?.name || '')}
                onFocus={() => {
                  setTemplateSearch('')
                  setIsTemplateDropdownOpen(true)
                }}
                onChange={(e) => {
                  setTemplateSearch(e.target.value)
                  setIsTemplateDropdownOpen(true)
                }}
                className="w-full h-12 px-5 pr-12 border-2 border-slate-100 rounded-2xl outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all font-bold text-slate-700"
              />
              <Search size={20} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
            </div>
            {isTemplateDropdownOpen && (
              <div className="absolute top-full left-0 right-0 mt-3 bg-white border border-slate-100 rounded-[24px] shadow-2xl z-[160] max-h-64 overflow-y-auto no-scrollbar ring-1 ring-black/5 p-2 animate-slide-up">
                {filteredTemplates.length > 0 ? filteredTemplates.map(tpl => {
                  const isSelected = selectedTemplate?.id === tpl.id
                  return (
                  <button
                    key={tpl.id}
                    onClick={() => handleSelectTemplate(tpl)}
                    className={`w-full text-left px-5 py-4 hover:bg-blue-50 flex items-center justify-between rounded-xl transition-all mb-1 last:mb-0 ${isSelected ? 'bg-blue-50/80' : ''}`}
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-xl flex items-center justify-center font-black bg-blue-100 text-blue-600">模板</div>
                      <p className={`text-sm font-black ${isSelected ? 'text-blue-600' : 'text-slate-800'}`}>{tpl.name}</p>
                    </div>
                    {isSelected && <Check size={16} className="text-blue-600" />}
                  </button>
                  )
                }) : (
                  <div className="py-10 text-center text-slate-400 font-bold italic">未搜索到匹配模板</div>
                )}
              </div>
            )}
          </div>

          <div className="h-px bg-slate-100"></div>

          {/* 表单内容区 */}
          <div className="space-y-12">
            <DialysisModeSection
              values={dialysisModeValues}
              onChange={handleDialysisModeChange}
              dictOptions={dictOptions}
              dictTypeKey={DICT_TYPES.DIALYSIS_MODE}
              openNumericPad={openNumericPad}
              dialysateFlow={String(formData.parameters.flowRate)}
              onDialysateVolumeChange={(v) => setFormData(prev => ({...prev, parameters: {...prev.parameters, volume: parseFloat(v) || 0}}))}
              extraContent={
                <>
                  <div className="flex flex-col gap-2">
                    <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">频次描述</label>
                    <input
                      type="text"
                      value={formData.indicators.frequencyDesc}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        indicators: { ...prev.indicators, frequencyDesc: e.target.value }
                      }))}
                      className="w-full h-11 px-4 border border-slate-200 rounded-xl text-sm bg-white font-black outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                    />
                  </div>
                  <div className="flex flex-col gap-2">
                    <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">自动确认处方</label>
                    <div className="flex items-center gap-10 h-11 bg-white border border-slate-200 rounded-xl px-6">
                      <label className="flex items-center gap-2 cursor-pointer group">
                        <input
                          type="radio"
                          name="auto_new"
                          checked={formData.indicators.autoConfirm}
                          onChange={() => setFormData(prev => ({
                            ...prev,
                            indicators: { ...prev.indicators, autoConfirm: true }
                          }))}
                          className="w-4 h-4 text-blue-600 focus:ring-0 border-slate-300"
                        />
                        <span className="text-sm font-bold text-slate-600 group-hover:text-blue-600 transition-colors">是</span>
                      </label>
                      <label className="flex items-center gap-2 cursor-pointer group">
                        <input
                          type="radio"
                          name="auto_new"
                          checked={!formData.indicators.autoConfirm}
                          onChange={() => setFormData(prev => ({
                            ...prev,
                            indicators: { ...prev.indicators, autoConfirm: false }
                          }))}
                          className="w-4 h-4 text-blue-600 focus:ring-0 border-slate-300"
                        />
                        <span className="text-sm font-bold text-slate-600 group-hover:text-blue-600 transition-colors">否</span>
                      </label>
                    </div>
                  </div>
                  <div className="flex flex-col gap-2">
                    <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">方案启用状态</label>
                    <div className="flex items-center gap-10 h-11 bg-white border border-slate-200 rounded-xl px-6">
                      <label className="flex items-center gap-2 cursor-pointer group">
                        <input
                          type="radio"
                          name="status_new"
                          checked={formData.indicators.status === '启用'}
                          onChange={() => setFormData(prev => ({
                            ...prev,
                            indicators: { ...prev.indicators, status: '启用' }
                          }))}
                          className="w-4 h-4 text-green-600 focus:ring-0 border-slate-300"
                        />
                        <span className="text-sm font-bold text-slate-600 group-hover:text-green-600 transition-colors">启用</span>
                      </label>
                      <label className="flex items-center gap-2 cursor-pointer group">
                        <input
                          type="radio"
                          name="status_new"
                          checked={formData.indicators.status !== '启用'}
                          onChange={() => setFormData(prev => ({
                            ...prev,
                            indicators: { ...prev.indicators, status: '禁用' }
                          }))}
                          className="w-4 h-4 text-red-600 focus:ring-0 border-slate-300"
                        />
                        <span className="text-sm font-bold text-slate-600 group-hover:text-red-600 transition-colors">禁用</span>
                      </label>
                    </div>
                  </div>
                </>
              }
            />

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
              duration={String(formData.duration)}
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
              materials={formData.materials}
              onAdd={handleMaterialAdd}
              onRemove={handleMaterialRemove}
              onUpdate={handleMaterialUpdate}
              onReplace={handleMaterialReplace}
            />
          </div>
        </div>

        {/* Footer */}
        <div className="px-10 py-8 bg-slate-50 border-t border-slate-200 flex justify-end gap-3 shrink-0">
          <button onClick={onClose} className="px-8 py-3 rounded-2xl border border-slate-200 text-slate-500 font-bold hover:bg-slate-50 transition-all">
            取消退出
          </button>
          <button
            onClick={async () => {
              if (onSave) {
                try {
                  await onSave(formData)
                  onClose() // 只有保存成功才关闭模态框
                } catch {
                  // 保存失败，不关闭模态框
                }
              } else {
                onClose()
              }
            }}
            className="px-10 py-3 rounded-2xl bg-blue-600 text-white font-black hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all flex items-center gap-2"
          >
            <Save size={18} /> 确认并保存方案
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
  )
}

// 主组件
export default function TreatmentPlanTab({ patientId = '', patientName = '', treatmentPlan = null }: TreatmentPlanTabProps) {
  const navigate = useNavigate()
  const [isNewPlanModalOpen, setIsNewPlanModalOpen] = useState(false)
  const [isVascularModalOpen, setIsVascularModalOpen] = useState(false)
  const [syncModalKey, setSyncModalKey] = useState(0)

  // API 数据状态
  const [apiTreatmentPlans, setApiTreatmentPlans] = useState<ApiTreatmentPlan[]>([]) // 所有治疗方案
  const [selectedMode, setSelectedMode] = useState<string>('') // 当前选中的模式
  const [dictOptions, setDictOptions] = useState<Record<string, Array<{ value: string; label: string }>>>({}) // 字典选项
  const [vascularAccesses, setVascularAccesses] = useState<VascularAccessApi[]>([]) // 血管通路列表
  const [adjustmentRecords, setAdjustmentRecords] = useState<AdjustmentRecord[]>([]) // 方案调整记录
  const [showAllRecords, setShowAllRecords] = useState(false) // 是否展开显示所有调整记录

  // 顶部全局参数（不随模式 Tab 切换）
  const [sharedFields, setSharedFields] = useState({
    weeklyFrequency: 0,
    biweeklyFrequency: 0,
    duration: 4,
    dryWeight: 0,
    extraWeight: 0,
  })

  // 获取字典名称的辅助函数
  const getDictName = (typeCode: string, code: string): string => {
    const options = dictOptions[typeCode] || []
    const option = options.find(opt => opt.value === code)
    return option?.label || code
  }

  // 将 ApiTreatmentPlan 转换为 TreatmentPlan 格式
  const convertApiPlanToUiPlan = (apiPlan: ApiTreatmentPlan): TreatmentPlan => ({
    weeklyFrequency: apiPlan.weeklyFrequency,
    biweeklyFrequency: apiPlan.biweeklyFrequency,
    duration: apiPlan.duration,
    dryWeight: apiPlan.dryWeight,
    extraWeight: apiPlan.extraWeight,
    vascularAccess: '-', // API 没有这个字段，使用默认值
    indicators: {
      mode: apiPlan.dialysisMode.mode,
      bloodFlow: apiPlan.dialysisMode.bloodFlow,
      bv: apiPlan.dialysisMode.bv,
      frequencyDesc: apiPlan.dialysisMode.frequencyDesc,
      autoConfirm: apiPlan.dialysisMode.autoConfirm,
      status: apiPlan.dialysisMode.status,
      notes: apiPlan.dialysisMode.notes,
      substituteMode: apiPlan.dialysisMode.substituteInputMode || '',
      substituteFlow: apiPlan.dialysisMode.substituteFlow || 0,
      substituteVolume: apiPlan.dialysisMode.substituteVolume || 0,
    },
    anticoagulant: {
      initialDrug: apiPlan.anticoagulant.initialDrug,
      initialDose: apiPlan.anticoagulant.initialDose,
      totalDose: apiPlan.anticoagulant.totalDose,
      maintenanceDrug: apiPlan.anticoagulant.maintenanceDrug,
      infusionRate: apiPlan.anticoagulant.infusionRate,
      infusionTime: apiPlan.anticoagulant.infusionTime,
      maintenanceDose: apiPlan.anticoagulant.maintenanceDose,
    },
    parameters: {
      dialysateType: apiPlan.parameters.dialysateType,
      dialysateGroup: apiPlan.parameters.dialysateGroup,
      flowRate: apiPlan.parameters.flowRate,
      na: apiPlan.parameters.na,
      ca: apiPlan.parameters.ca,
      k: apiPlan.parameters.k,
      hco3: apiPlan.parameters.hco3,
      glucose: apiPlan.parameters.glucose,
      conductivity: apiPlan.parameters.conductivity,
      temp: apiPlan.parameters.temp,
      volume: apiPlan.parameters.volume,
    },
    materials: apiPlan.materials.map(m => ({
      id: m.id,
      name: m.name,
      category: m.category,
      count: m.count,
      code: m.code,
      brand: m.brand,
      spec: m.spec,
      note: m.note,
    })),
  })

  // 检查是否有治疗方案（同时检查 props 和 API 状态）
  const hasTreatmentPlan = (treatmentPlan !== null && treatmentPlan !== undefined) || apiTreatmentPlans.length > 0

  // 获取当前选中的治疗方案
  const currentApiPlan = apiTreatmentPlans.find(p => p.dialysisMode.mode === selectedMode)

  // 使用传入的 plan 或 API 数据（优先级：API 数据 > props）
  const plan = currentApiPlan
    ? convertApiPlanToUiPlan(currentApiPlan)
    : treatmentPlan

  // 默认治疗方案（临床常规默认值）
  const defaultPlan: TreatmentPlan = {
    weeklyFrequency: 1,
    biweeklyFrequency: 1,
    duration: 4,
    dryWeight: 0,
    extraWeight: 0,
    vascularAccess: '-',
    indicators: {
      mode: 'HD',
      bloodFlow: 200,
      bv: '',
      frequencyDesc: '',
      autoConfirm: false,
      status: '启用',
      notes: '',
      substituteMode: '',
      substituteFlow: 0,
      substituteVolume: 0,
    },
    anticoagulant: {
      initialDrug: '',
      initialDose: '',
      totalDose: '',
      maintenanceDrug: '',
      infusionRate: '',
      infusionTime: '',
      maintenanceDose: ''
    },
    parameters: {
      dialysateType: '',
      dialysateGroup: '',
      flowRate: 500,
      na: 140,
      ca: 1.5,
      k: 2.0, // 标准默认钾浓度（与建档草稿种子、后端默认一致）；≠2.0 时字段标红提示
      hco3: 35,
      glucose: '',
      conductivity: 14.1,
      temp: 37,
      volume: 120
    },
    materials: []
  }

  const safePlan = plan || defaultPlan

  // 可编辑状态：editingPlan 用于表单控件 (controlled)
  const [editingPlan, setEditingPlan] = useState<TreatmentPlan | null>(null)

  // 数字键盘（主内容区）
  const {
    padState: mainPadState,
    openNumericPad: mainOpenNumericPad,
    closeNumericPad: mainCloseNumericPad,
    handleKeyPress: mainHandleKeyPress,
    handleClear: mainHandleClear,
    handleConfirm: mainHandleConfirm,
  } = useNumericPad()

  // 抗凝剂选项和规格映射（主内容区）
  const [mainAnticoagulantOptions, setMainAnticoagulantOptions] = useState<Array<{ value: string; label: string }>>([])
  const [mainAnticoagulantSpecMap, setMainAnticoagulantSpecMap] = useState<Record<string, string>>({})

  // 初始化/同步 editingPlan：当 API 数据或模式变化时
  useEffect(() => {
    const plan = currentApiPlan ? convertApiPlanToUiPlan(currentApiPlan) : treatmentPlan || defaultPlan
    setEditingPlan(plan)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentApiPlan, treatmentPlan])

  // 加载患者治疗方案列表和字典数据
  useEffect(() => {
    const loadData = async () => {
      if (!patientId) return

      try {
        // 并行加载治疗方案、字典数据和药品目录
        const [plans, dialysisModeOpts, dialysateTypeOpts, dialysateGroupOpts, dialysateFlowOpts, glucoseOpts, vascularAccessTypeOpts, vascularSiteOpts, vascularList, drugCategoryOpts, drugsRes, adjRecords] = await Promise.allSettled([
          patientApi.getTreatmentPlans(patientId),
          dictCache.getOptions(DICT_TYPES.DIALYSIS_MODE),
          dictCache.getOptions(DICT_TYPES.DIALYSATE_TYPE),
          dictCache.getOptions(DICT_TYPES.DIALYSATE_GROUP),
          dictCache.getOptions(DICT_TYPES.DIALYSATE_FLOW),
          dictCache.getOptions(DICT_TYPES.GLUCOSE),
          dictCache.getOptions(DICT_TYPES.VASCULAR_ACCESS),
          dictCache.getOptions(DICT_TYPES.VASCULAR_SITE),
          restApi.getVascularAccesses(patientId),
          dictCache.getOptions(DICT_TYPES.DRUG_CATEGORY),
          drugCatalogApi.list({ isEnabled: true, pageSize: 9999 }),
          patientApi.getAdjustmentRecords(patientId),
        ])

        if (plans.status === 'fulfilled') {
          setApiTreatmentPlans(plans.value || [])
          // 如果有治疗方案，默认选中第一个
          if (plans.value && plans.value.length > 0 && !selectedMode) {
            setSelectedMode(plans.value[0].dialysisMode.mode)
          }
          // 初始化顶部全局参数（从第一个方案读取）
          if (plans.value && plans.value.length > 0) {
            const first = plans.value[0]
            setSharedFields({
              weeklyFrequency: first.weeklyFrequency,
              biweeklyFrequency: first.biweeklyFrequency,
              duration: first.duration,
              dryWeight: first.dryWeight,
              extraWeight: first.extraWeight,
            })
          }
        }

        // 设置字典选项
        setDictOptions({
          [DICT_TYPES.DIALYSIS_MODE]: dialysisModeOpts.status === 'fulfilled' ? dialysisModeOpts.value : [],
          [DICT_TYPES.DIALYSATE_TYPE]: dialysateTypeOpts.status === 'fulfilled' ? dialysateTypeOpts.value : [],
          [DICT_TYPES.DIALYSATE_GROUP]: dialysateGroupOpts.status === 'fulfilled' ? dialysateGroupOpts.value : [],
          [DICT_TYPES.DIALYSATE_FLOW]: dialysateFlowOpts.status === 'fulfilled' ? dialysateFlowOpts.value : [],
          [DICT_TYPES.GLUCOSE]: glucoseOpts.status === 'fulfilled' ? glucoseOpts.value : [],
          [DICT_TYPES.VASCULAR_ACCESS]: vascularAccessTypeOpts.status === 'fulfilled' ? vascularAccessTypeOpts.value : [],
          [DICT_TYPES.VASCULAR_SITE]: vascularSiteOpts.status === 'fulfilled' ? vascularSiteOpts.value : [],
        })

        // 设置血管通路列表
        if (vascularList.status === 'fulfilled') {
          setVascularAccesses(vascularList.value || [])
        }

        // 设置方案调整记录
        if (adjRecords.status === 'fulfilled') {
          setAdjustmentRecords(adjRecords.value || [])
        }

        // 构建抗凝剂选项和规格映射（主内容区）
        if (drugsRes.status === 'fulfilled') {
          const allDrugs = drugsRes.value.items
          const drugCategoryOptions = drugCategoryOpts.status === 'fulfilled' ? drugCategoryOpts.value : []
          const categoryMap = new Map(
            drugCategoryOptions.map(item => [item.value, item.label])
          )
          const filtered = allDrugs.filter((item: DrugCatalog) => {
            const label = categoryMap.get(item.category) || item.category
            return label === '抗凝剂' || label === '抗凝药' || item.category === '抗凝剂' || item.category === '抗凝药' || item.category === 'ANTICOAGULANT'
          })
          const sorted = [...filtered].sort((a, b) => {
            const aOrder = a.sortOrder && a.sortOrder > 0 ? a.sortOrder : 9999
            const bOrder = b.sortOrder && b.sortOrder > 0 ? b.sortOrder : 9999
            if (aOrder !== bOrder) return aOrder - bOrder
            return a.name.localeCompare(b.name)
          })
          setMainAnticoagulantOptions(sorted.map(item => ({ value: item.name, label: item.name })))
          setMainAnticoagulantSpecMap(
            sorted.reduce<Record<string, string>>((acc, item) => {
              const unit = extractDoseUnit(item.concentration)
              if (unit) {
                acc[item.name] = unit
              }
              return acc
            }, {})
          )
        }
      } catch (error) {
        console.error('加载数据失败:', error)
      }
    }

    loadData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [patientId])

  // --- 调整记录 diff 工具函数 ---

  // 比较两个材料列表，生成细粒度 diff 描述（逐项对比新增/删除/数量变更）
  const diffMaterials = (
    oldMats: Array<{ name: string; category: string; count: number }>,
    newMats: Array<{ name: string; category: string; count: number }>,
  ): string | null => {
    const oldMap = new Map(oldMats.map(m => [m.name, m]))
    const newMap = new Map(newMats.map(m => [m.name, m]))
    const parts: string[] = []

    // 删除的材料
    for (const [name] of oldMap) {
      if (!newMap.has(name)) parts.push(`删除【${name}】`)
    }
    // 新增的材料
    for (const [name, m] of newMap) {
      if (!oldMap.has(name)) parts.push(`新增【${name}(${m.count})】`)
    }
    // 数量变更
    for (const [name, newM] of newMap) {
      const oldM = oldMap.get(name)
      if (oldM && oldM.count !== newM.count) {
        parts.push(`${name}数量：由【${oldM.count}】调整为【${newM.count}】`)
      }
    }

    return parts.length > 0 ? parts.join('；') : null
  }

  // 材料同步保存处理
  const handleMaterialSyncSave = async (result: MaterialSyncResult) => {
    try {
      const diffParts: string[] = []

      // 1. 血管通路变更 diff
      const oldAccessId = currentDefaultAccessId
      const newAccessId = result.selectedAccessId
      if (oldAccessId && newAccessId && oldAccessId !== newAccessId) {
        const oldLabel = syncModalVascularAccesses.find(va => va.id === oldAccessId)?.label || oldAccessId
        const newLabel = syncModalVascularAccesses.find(va => va.id === newAccessId)?.label || newAccessId
        diffParts.push(`血管通路：由【${oldLabel}】调整为【${newLabel}】`)
      }

      // 2. 各方案材料变更 diff + 逐个更新
      const failedPlans: string[] = []
      for (const planData of result.plans) {
        const apiPlan = apiTreatmentPlans.find(p => p.id === planData.planId)
        if (!apiPlan) continue

        const materialDiff = diffMaterials(
          apiPlan.materials.map(m => ({ name: m.name, category: m.category, count: m.count })),
          planData.materials.map(m => ({ name: m.name, category: m.category, count: m.count })),
        )
        if (materialDiff) {
          diffParts.push(`${apiPlan.dialysisMode.mode}方案材料：${materialDiff}`)
        }

        // 更新治疗方案的材料列表（带 dialysisMode 让后端识别是哪个方案）
        try {
          await patientApi.updateTreatmentPlan(patientId, {
            dialysisMode: apiPlan.dialysisMode,
            materials: planData.materials.map(m => ({
              id: m.id,
              name: m.name,
              category: m.category,
              count: m.count,
              code: m.code,
              brand: m.brand,
              spec: m.spec,
              note: '',
            })),
          })
        } catch (err) {
          console.error(`更新${apiPlan.dialysisMode.mode}方案失败:`, err)
          failedPlans.push(apiPlan.dialysisMode.mode)
        }
      }

      // 部分失败提示
      if (failedPlans.length > 0) {
        console.warn(`以下方案更新失败: ${failedPlans.join(', ')}`)
      }

      // 3. 如果血管通路变更，更新默认通路
      if (oldAccessId && newAccessId && oldAccessId !== newAccessId) {
        try {
          // 取消旧默认
          const oldAccess = vascularAccesses.find(va => va.id === oldAccessId)
          if (oldAccess) {
            await restApi.updateVascularAccess(patientId, oldAccessId, { ...oldAccess, isDefault: false })
          }
          // 设置新默认
          const newAccess = vascularAccesses.find(va => va.id === newAccessId)
          if (newAccess) {
            await restApi.updateVascularAccess(patientId, newAccessId, { ...newAccess, isDefault: true })
          }
        } catch (err) {
          console.error('更新默认血管通路失败:', err)
        }
      }

      // 4. 如果有变更，保存调整记录
      if (diffParts.length > 0) {
        const content = diffParts.join('；') + '；'
        try {
          await patientApi.createAdjustmentRecord(patientId, { content })
        } catch (err) {
          console.error('保存调整记录失败:', err)
        }
      }

      // 5. 重新加载数据（方案 + 调整记录 + 血管通路）
      const [plansResult, recordsResult, vaResult] = await Promise.allSettled([
        patientApi.getTreatmentPlans(patientId),
        patientApi.getAdjustmentRecords(patientId),
        restApi.getVascularAccesses(patientId),
      ])
      if (plansResult.status === 'fulfilled') {
        setApiTreatmentPlans(plansResult.value || [])
        // 同步 sharedFields（从最新数据）
        if (plansResult.value && plansResult.value.length > 0) {
          const first = plansResult.value[0]
          setSharedFields({
            weeklyFrequency: first.weeklyFrequency,
            biweeklyFrequency: first.biweeklyFrequency,
            duration: first.duration,
            dryWeight: first.dryWeight,
            extraWeight: first.extraWeight,
          })
        }
      }
      if (recordsResult.status === 'fulfilled') setAdjustmentRecords(recordsResult.value || [])
      if (vaResult.status === 'fulfilled') setVascularAccesses(vaResult.value || [])
      // 同步 editingPlan
      const loadedPlans = plansResult.status === 'fulfilled' ? plansResult.value : apiTreatmentPlans
      const updatedPlan = loadedPlans.find((p: ApiTreatmentPlan) => p.dialysisMode.mode === selectedMode)
      if (updatedPlan) {
        setEditingPlan(convertApiPlanToUiPlan(updatedPlan))
      }
    } catch (error) {
      console.error('材料同步保存失败:', error)
    }
  }

  // 格式化血管通路显示名称
  const formatVascularAccessLabel = (va: VascularAccessApi): string => {
    const typeName = getDictName(DICT_TYPES.VASCULAR_ACCESS, va.accessType)
    const siteName = getDictName(DICT_TYPES.VASCULAR_SITE, va.site)
    const sideName = va.side === 'L' ? '左' : va.side === 'R' ? '右' : ''
    const parts = [typeName, sideName, siteName].filter(Boolean)
    const label = parts.join(' ')
    if (va.isDefault) return `${label}（默认）`
    if (va.isDisabled) return `${label}（已禁用）`
    return label
  }

  // 构造 MaterialSyncModal 需要的数据
  const syncModalPlans = apiTreatmentPlans.map(p => ({
    mode: p.dialysisMode.mode,
    planId: p.id,
    materials: p.materials.map(m => ({
      id: m.id,
      name: m.name,
      category: m.category,
      count: m.count,
      code: m.code,
      brand: m.brand,
      spec: m.spec,
    })),
  }))

  const syncModalVascularAccesses = vascularAccesses.map(va => ({
    id: va.id,
    label: formatVascularAccessLabel(va),
    accessType: va.accessType,
  }))

  const currentDefaultAccessId = vascularAccesses.find(va => va.isDefault)?.id || vascularAccesses[0]?.id || ''

  // --- 主内容区 Section 适配器：桥接 editingPlan (TreatmentPlan) ↔ 共享组件 (all-string) ---

  // 用于渲染的 plan：优先使用 editingPlan（编辑中的值），否则使用 safePlan
  const displayPlan = editingPlan || safePlan

  // 透析模式
  const mainDialysisModeValues: DialysisModeValues = {
    method: displayPlan.indicators.mode,
    duration: String(sharedFields.duration),
    bloodFlow: String(displayPlan.indicators.bloodFlow),
    bv: displayPlan.indicators.bv || '',
    substituteMode: displayPlan.indicators.substituteMode || '',
    substituteFlow: displayPlan.indicators.substituteFlow != null ? String(displayPlan.indicators.substituteFlow) : '',
    substituteVolume: displayPlan.indicators.substituteVolume != null ? String(displayPlan.indicators.substituteVolume) : '',
    notes: displayPlan.indicators.notes || '',
  }

  const handleMainDialysisModeChange = (updater: (prev: DialysisModeValues) => DialysisModeValues) => {
    setEditingPlan(prev => {
      if (!prev) return prev
      const modeValues: DialysisModeValues = {
        method: prev.indicators.mode,
        duration: String(sharedFields.duration),
        bloodFlow: String(prev.indicators.bloodFlow),
        bv: prev.indicators.bv || '',
        substituteMode: prev.indicators.substituteMode || '',
        substituteFlow: prev.indicators.substituteFlow != null ? String(prev.indicators.substituteFlow) : '',
        substituteVolume: prev.indicators.substituteVolume != null ? String(prev.indicators.substituteVolume) : '',
        notes: prev.indicators.notes || '',
      }
      const next = updater(modeValues)
      const parsedDuration = parseFloat(next.duration)
      const parsedBloodFlow = parseInt(next.bloodFlow)
      const parsedSubstituteFlow = parseFloat(next.substituteFlow)
      const parsedSubstituteVolume = parseFloat(next.substituteVolume)
      // duration 是全局字段，同步到 sharedFields
      if (Number.isFinite(parsedDuration) && parsedDuration !== sharedFields.duration) {
        setSharedFields(sf => ({ ...sf, duration: parsedDuration }))
      }
      return {
        ...prev,
        indicators: {
          ...prev.indicators,
          mode: next.method,
          bloodFlow: Number.isFinite(parsedBloodFlow) ? parsedBloodFlow : prev.indicators.bloodFlow,
          bv: next.bv,
          notes: next.notes,
          substituteMode: next.substituteMode,
          substituteFlow: Number.isFinite(parsedSubstituteFlow) ? parsedSubstituteFlow : 0,
          substituteVolume: Number.isFinite(parsedSubstituteVolume) ? parsedSubstituteVolume : 0,
        }
      }
    })
  }

  // 抗凝剂
  const mainAnticoagulantValues: AnticoagulantValues = {
    initialDrug: displayPlan.anticoagulant.initialDrug,
    initialDose: displayPlan.anticoagulant.initialDose,
    maintenanceDrug: displayPlan.anticoagulant.maintenanceDrug,
    infusionRate: displayPlan.anticoagulant.infusionRate,
    infusionTime: displayPlan.anticoagulant.infusionTime,
    maintenanceDose: displayPlan.anticoagulant.maintenanceDose,
    totalDose: displayPlan.anticoagulant.totalDose,
  }

  const handleMainAnticoagulantChange = (updater: (prev: AnticoagulantValues) => AnticoagulantValues) => {
    setEditingPlan(prev => {
      if (!prev) return prev
      const acValues: AnticoagulantValues = {
        initialDrug: prev.anticoagulant.initialDrug,
        initialDose: prev.anticoagulant.initialDose,
        maintenanceDrug: prev.anticoagulant.maintenanceDrug,
        infusionRate: prev.anticoagulant.infusionRate,
        infusionTime: prev.anticoagulant.infusionTime,
        maintenanceDose: prev.anticoagulant.maintenanceDose,
        totalDose: prev.anticoagulant.totalDose,
      }
      const next = updater(acValues)
      return {
        ...prev,
        anticoagulant: {
          initialDrug: next.initialDrug,
          initialDose: next.initialDose,
          maintenanceDrug: next.maintenanceDrug,
          infusionRate: next.infusionRate,
          infusionTime: next.infusionTime,
          maintenanceDose: next.maintenanceDose,
          totalDose: next.totalDose,
        }
      }
    })
  }

  // 透析参数
  const mainDialysisParamsValues: DialysisParamsValues = {
    dialysateType: displayPlan.parameters.dialysateType,
    dialysateGroup: displayPlan.parameters.dialysateGroup,
    dialysateFlow: String(displayPlan.parameters.flowRate),
    na: String(displayPlan.parameters.na),
    ca: String(displayPlan.parameters.ca),
    k: String(displayPlan.parameters.k),
    hco3: String(displayPlan.parameters.hco3),
    glucose: displayPlan.parameters.glucose,
    conductivity: String(displayPlan.parameters.conductivity),
    temp: String(displayPlan.parameters.temp),
    dialysateVolume: String(displayPlan.parameters.volume),
  }

  const handleMainDialysisParamsChange = (updater: (prev: DialysisParamsValues) => DialysisParamsValues) => {
    setEditingPlan(prev => {
      if (!prev) return prev
      const paramValues: DialysisParamsValues = {
        dialysateType: prev.parameters.dialysateType,
        dialysateGroup: prev.parameters.dialysateGroup,
        dialysateFlow: String(prev.parameters.flowRate),
        na: String(prev.parameters.na),
        ca: String(prev.parameters.ca),
        k: String(prev.parameters.k),
        hco3: String(prev.parameters.hco3),
        glucose: prev.parameters.glucose,
        conductivity: String(prev.parameters.conductivity),
        temp: String(prev.parameters.temp),
        dialysateVolume: String(prev.parameters.volume),
      }
      const next = updater(paramValues)
      const pf = (v: string, fallback: number) => { const n = parseFloat(v); return Number.isFinite(n) ? n : fallback }
      return {
        ...prev,
        parameters: {
          dialysateType: next.dialysateType,
          dialysateGroup: next.dialysateGroup,
          flowRate: pf(next.dialysateFlow, prev.parameters.flowRate),
          na: pf(next.na, prev.parameters.na),
          ca: pf(next.ca, prev.parameters.ca),
          k: pf(next.k, prev.parameters.k),
          hco3: pf(next.hco3, prev.parameters.hco3),
          glucose: next.glucose,
          conductivity: pf(next.conductivity, prev.parameters.conductivity),
          temp: pf(next.temp, prev.parameters.temp),
          volume: pf(next.dialysateVolume, prev.parameters.volume),
        }
      }
    })
  }

  // 材料操作（主内容区）
  const handleMainMaterialAdd = (name: string, material: Partial<MaterialItem>) => {
    const newMaterial: MaterialItem = {
      id: `mat-${Date.now()}`,
      name,
      category: material.category || '-',
      count: 1,
      code: material.code || '-',
      brand: material.brand || '-',
      spec: material.spec || '-',
      note: ''
    }
    setEditingPlan(prev => prev ? {
      ...prev,
      materials: [...prev.materials, newMaterial]
    } : prev)
  }

  const handleMainMaterialRemove = (id: string) => {
    setEditingPlan(prev => prev ? {
      ...prev,
      materials: prev.materials.filter((m: MaterialItem) => m.id !== id)
    } : prev)
  }

  const handleMainMaterialUpdate = (id: string, field: string, value: string | number) => {
    setEditingPlan(prev => prev ? {
      ...prev,
      materials: prev.materials.map((m: MaterialItem) =>
        m.id === id ? { ...m, [field]: value } : m
      )
    } : prev)
  }

  const handleMainMaterialReplace = (id: string, patch: Partial<MaterialItem>) => {
    setEditingPlan(prev => prev ? {
      ...prev,
      materials: prev.materials.map((m: MaterialItem) =>
        m.id === id ? { ...m, ...patch } : m
      )
    } : prev)
  }

  // 保存治疗方案（主内容区）
  const handleSavePlan = async () => {
    if (!patientId) {
      alert('缺少患者 ID')
      return
    }

    // 如果没有治疗方案，提示用户先创建
    if (!editingPlan) {
      alert('请先创建治疗方案')
      return
    }

    try {
      // 将前端数据转换为 API 格式（当前模式的完整数据 + 顶部全局参数）
      const apiData = {
        weeklyFrequency: sharedFields.weeklyFrequency,
        biweeklyFrequency: sharedFields.biweeklyFrequency,
        duration: sharedFields.duration,
        dryWeight: sharedFields.dryWeight,
        extraWeight: sharedFields.extraWeight,
        status: editingPlan.indicators?.status || '启用',
        notes: editingPlan.indicators?.notes || '',
        dialysisMode: {
          mode: editingPlan.indicators?.mode || 'HD',
          bloodFlow: editingPlan.indicators?.bloodFlow || 200,
          bv: editingPlan.indicators?.bv || '',
          frequencyDesc: editingPlan.indicators?.frequencyDesc || '',
          autoConfirm: editingPlan.indicators?.autoConfirm || false,
          status: editingPlan.indicators?.status || '启用',
          notes: editingPlan.indicators?.notes || '',
          substituteInputMode: editingPlan.indicators?.substituteMode || '',
          substituteFlow: editingPlan.indicators?.substituteFlow || 0,
          substituteVolume: editingPlan.indicators?.substituteVolume || 0,
        },
        anticoagulant: {
          initialDrug: editingPlan.anticoagulant.initialDrug,
          initialDose: editingPlan.anticoagulant.initialDose,
          totalDose: editingPlan.anticoagulant.totalDose,
          maintenanceDrug: editingPlan.anticoagulant.maintenanceDrug,
          infusionRate: editingPlan.anticoagulant.infusionRate,
          infusionTime: editingPlan.anticoagulant.infusionTime,
          maintenanceDose: editingPlan.anticoagulant.maintenanceDose,
        },
        parameters: {
          dialysateType: editingPlan.parameters.dialysateType,
          dialysateGroup: editingPlan.parameters.dialysateGroup,
          flowRate: editingPlan.parameters.flowRate,
          na: editingPlan.parameters.na,
          ca: editingPlan.parameters.ca,
          k: editingPlan.parameters.k,
          hco3: editingPlan.parameters.hco3,
          glucose: editingPlan.parameters.glucose,
          conductivity: editingPlan.parameters.conductivity,
          temp: editingPlan.parameters.temp,
          volume: editingPlan.parameters.volume,
        },
        materials: editingPlan.materials.map(m => ({
          id: m.id,
          name: m.name,
          category: m.category,
          count: m.count,
          code: m.code,
          brand: m.brand,
          spec: m.spec,
          note: m.note,
        })),
      }

      // 检查是否已存在相同模式的治疗方案
      const existingPlan = apiTreatmentPlans.find(p => p.dialysisMode.mode === apiData.dialysisMode.mode)

      // 频率变更 → 影响排班，保存后引导去「方案变更」按生效日重排（delta⑤ 轻量桥接）
      let schedulingChanged = false

      // 生成调整记录 diff
      const diffParts: string[] = []
      if (existingPlan) {
        const oldPlan = convertApiPlanToUiPlan(existingPlan)

        // 全局参数对比（与任意一个方案比即可，因为全局参数应全部一致）
        if (existingPlan.weeklyFrequency !== sharedFields.weeklyFrequency) {
          diffParts.push(`单周频次：由【${existingPlan.weeklyFrequency}】调整为【${sharedFields.weeklyFrequency}】`)
          schedulingChanged = true
        }
        if (existingPlan.biweeklyFrequency !== sharedFields.biweeklyFrequency) {
          diffParts.push(`双周频次：由【${existingPlan.biweeklyFrequency}】调整为【${sharedFields.biweeklyFrequency}】`)
          schedulingChanged = true
        }
        if (existingPlan.duration !== sharedFields.duration)
          diffParts.push(`透析时长：由【${existingPlan.duration}h】调整为【${sharedFields.duration}h】`)
        if (existingPlan.dryWeight !== sharedFields.dryWeight)
          diffParts.push(`干体重：由【${existingPlan.dryWeight}kg】调整为【${sharedFields.dryWeight}kg】`)

        // 模式特有参数对比
        const newPlan = editingPlan
        if (oldPlan.indicators.bloodFlow !== newPlan.indicators.bloodFlow)
          diffParts.push(`${selectedMode}血流量：由【${oldPlan.indicators.bloodFlow}】调整为【${newPlan.indicators.bloodFlow}】`)

        // 抗凝剂对比
        if (oldPlan.anticoagulant.initialDrug !== newPlan.anticoagulant.initialDrug)
          diffParts.push(`${selectedMode}首剂药物：由【${oldPlan.anticoagulant.initialDrug || '无'}】调整为【${newPlan.anticoagulant.initialDrug || '无'}】`)

        // 材料对比
        const materialDiff = diffMaterials(
          oldPlan.materials.map(m => ({ name: m.name, category: m.category, count: m.count })),
          newPlan.materials.map(m => ({ name: m.name, category: m.category, count: m.count })),
        )
        if (materialDiff) diffParts.push(`${selectedMode}方案${materialDiff}`)

        // 更新当前模式的方案（完整数据）
        await patientApi.updateTreatmentPlan(patientId, apiData)

        // 同步全局参数到其他模式的方案
        for (const otherPlan of apiTreatmentPlans) {
          if (otherPlan.dialysisMode.mode === selectedMode) continue
          try {
            await patientApi.updateTreatmentPlan(patientId, {
              dialysisMode: otherPlan.dialysisMode,
              weeklyFrequency: sharedFields.weeklyFrequency,
              biweeklyFrequency: sharedFields.biweeklyFrequency,
              duration: sharedFields.duration,
              dryWeight: sharedFields.dryWeight,
              extraWeight: sharedFields.extraWeight,
            })
          } catch (err) {
            console.error(`同步全局参数到${otherPlan.dialysisMode.mode}方案失败:`, err)
          }
        }

        // 更新成功后再保存调整记录
        if (diffParts.length > 0) {
          const content = diffParts.join('；') + '；'
          try {
            await patientApi.createAdjustmentRecord(patientId, { content })
          } catch (err) {
            console.error('保存调整记录失败:', err)
          }
        }
      } else {
        await patientApi.createTreatmentPlan(patientId, apiData)
      }

      // 重新加载方案列表和调整记录
      const [plansResult, recordsResult] = await Promise.allSettled([
        patientApi.getTreatmentPlans(patientId),
        patientApi.getAdjustmentRecords(patientId),
      ])
      if (plansResult.status === 'fulfilled') {
        setApiTreatmentPlans(plansResult.value || [])
        // 同步 sharedFields（从最新数据）
        if (plansResult.value && plansResult.value.length > 0) {
          const first = plansResult.value[0]
          setSharedFields({
            weeklyFrequency: first.weeklyFrequency,
            biweeklyFrequency: first.biweeklyFrequency,
            duration: first.duration,
            dryWeight: first.dryWeight,
            extraWeight: first.extraWeight,
          })
        }
      }
      if (recordsResult.status === 'fulfilled') setAdjustmentRecords(recordsResult.value || [])

      message.success('治疗方案保存成功')
      // delta⑤ 轻量桥接：频率变更影响排班，引导去「智能排班 · 方案变更」按生效日重排。
      // 临床方案(次数)与排班骨架(模式码)按设计解耦，故不自动映射，由医生在排班页显式选模式+生效日。
      if (schedulingChanged) {
        Modal.confirm({
          title: '透析频率已变更 · 是否同步排班？',
          content: '本次方案的透析频率有调整，需按「生效日」重排排班。前往「智能排班 · 方案变更」处理？（仅取消生效日之后尚未确认的排班，已二次确认的将报警人工核对）',
          okText: '去方案变更',
          cancelText: '稍后处理',
          onOk: () => navigate(`/schedule?planChangePatient=${patientId}`),
        })
      }
    } catch (error) {
      console.error('保存治疗方案失败:', error)
      message.error(getErrorMessage(error))
    }
  }

  // 保存新建方案（模态框）
  const handleNewPlanSave = async (formData: TreatmentPlan) => {
    if (!patientId) {
      alert('缺少患者 ID')
      return
    }

    try {
      // 弹窗中不展示的字段，优先用模板值，否则从全局共享字段继承
      const weeklyFrequency = formData.weeklyFrequency || sharedFields.weeklyFrequency || 0
      const biweeklyFrequency = formData.biweeklyFrequency || sharedFields.biweeklyFrequency || 0
      const dryWeight = formData.dryWeight || sharedFields.dryWeight || 0
      const extraWeight = formData.extraWeight || sharedFields.extraWeight || 0

      // 将前端数据转换为 API 格式
      const apiData = {
        weeklyFrequency,
        biweeklyFrequency,
        duration: formData.duration,
        dryWeight,
        extraWeight,
        status: formData.indicators?.status || '启用',
        notes: formData.indicators?.notes || '',
        dialysisMode: {
          mode: formData.indicators?.mode || 'HD',
          bloodFlow: formData.indicators?.bloodFlow || 200,
          bv: formData.indicators?.bv || '',
          frequencyDesc: formData.indicators?.frequencyDesc || '',
          autoConfirm: formData.indicators?.autoConfirm || false,
          status: formData.indicators?.status || '启用',
          notes: formData.indicators?.notes || '',
          substituteInputMode: formData.indicators?.substituteMode || '',
          substituteFlow: formData.indicators?.substituteFlow || 0,
          substituteVolume: formData.indicators?.substituteVolume || 0,
        },
        anticoagulant: {
          initialDrug: formData.anticoagulant.initialDrug,
          initialDose: formData.anticoagulant.initialDose,
          totalDose: formData.anticoagulant.totalDose,
          maintenanceDrug: formData.anticoagulant.maintenanceDrug,
          infusionRate: formData.anticoagulant.infusionRate,
          infusionTime: formData.anticoagulant.infusionTime,
          maintenanceDose: formData.anticoagulant.maintenanceDose,
        },
        parameters: {
          dialysateType: formData.parameters.dialysateType,
          dialysateGroup: formData.parameters.dialysateGroup,
          flowRate: formData.parameters.flowRate,
          na: formData.parameters.na,
          ca: formData.parameters.ca,
          k: formData.parameters.k,
          hco3: formData.parameters.hco3,
          glucose: formData.parameters.glucose,
          conductivity: formData.parameters.conductivity,
          temp: formData.parameters.temp,
          volume: formData.parameters.volume,
        },
        materials: formData.materials.map(m => ({
          id: m.id,
          name: m.name,
          category: m.category,
          count: m.count,
          code: m.code,
          brand: m.brand,
          spec: m.spec,
          note: m.note,
        })),
      }

      // 创建新的治疗方案
      await patientApi.createTreatmentPlan(patientId, apiData)

      // 重新加载方案列表
      const plans = await patientApi.getTreatmentPlans(patientId)
      setApiTreatmentPlans(plans || [])

      // 同步 sharedFields（从最新数据）
      if (plans && plans.length > 0) {
        const first = plans[0]
        setSharedFields({
          weeklyFrequency: first.weeklyFrequency,
          biweeklyFrequency: first.biweeklyFrequency,
          duration: first.duration,
          dryWeight: first.dryWeight,
          extraWeight: first.extraWeight,
        })
      }

      alert('治疗方案创建成功！')
    } catch (error) {
      console.error('创建治疗方案失败:', error)
      message.error(getErrorMessage(error))
      throw error // 重新抛出错误，让调用方知道保存失败
    }
  }

  return (
    <div className="space-y-8 animate-fade-in pb-10">
      {/* 顶部操作栏 */}
      <div className="flex items-center justify-between mb-2 px-1">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsNewPlanModalOpen(true)}
            className="px-6 py-2.5 bg-blue-600 text-white text-sm font-black rounded-2xl hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all flex items-center gap-2"
          >
            <PlusCircle size={18} /> 新建方案
          </button>
          <button className="px-6 py-2.5 bg-white border border-slate-200 text-slate-700 text-sm font-black rounded-2xl hover:bg-slate-50 transition-all shadow-sm flex items-center gap-2">
            <RotateCcw size={18} className="text-indigo-500" /> 同步到处方
          </button>
        </div>
        <button
          onClick={handleSavePlan}
          className="px-10 py-2.5 bg-blue-600 text-white text-sm font-black rounded-2xl hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all flex items-center gap-2"
        >
          <Save size={18} /> 保存方案
        </button>
      </div>

      {/* 顶部参数卡片 - 始终显示 */}
      <div className="bg-white rounded-3xl border border-slate-200 p-8 flex flex-wrap items-center gap-x-10 gap-y-6 shadow-sm ring-1 ring-slate-100">
          <div className="flex items-center gap-6">
            <div className="flex flex-col gap-1">
              <span className="text-[10px] font-black text-slate-400 uppercase tracking-wider">单周/双周频次</span>
              <div className="flex items-center gap-2">
                <input type="text" value={sharedFields.weeklyFrequency} onChange={(e) => { const v = parseInt(e.target.value); setSharedFields(prev => ({ ...prev, weeklyFrequency: Number.isFinite(v) ? v : prev.weeklyFrequency })) }} className="w-12 h-10 border border-slate-200 rounded-xl text-center text-sm font-black bg-slate-50/50 focus:bg-white transition-all outline-none focus:ring-1 focus:ring-blue-500" />
                <span className="text-slate-300 font-bold">/</span>
                <input type="text" value={sharedFields.biweeklyFrequency} onChange={(e) => { const v = parseInt(e.target.value); setSharedFields(prev => ({ ...prev, biweeklyFrequency: Number.isFinite(v) ? v : prev.biweeklyFrequency })) }} className="w-12 h-10 border border-slate-200 rounded-xl text-center text-sm font-black bg-slate-50/50 focus:bg-white transition-all outline-none focus:ring-1 focus:ring-blue-500" />
                <span className="text-xs text-slate-400 font-bold ml-1">次/周</span>
              </div>
            </div>
            <div className="w-px h-10 bg-slate-100"></div>
            <div className="flex flex-col gap-1">
              <span className="text-[10px] font-black text-slate-400 uppercase tracking-wider">单次透析时长</span>
              <div className="flex items-center gap-2">
                <input type="text" value={sharedFields.duration} onChange={(e) => { const v = parseFloat(e.target.value); setSharedFields(prev => ({ ...prev, duration: Number.isFinite(v) ? v : prev.duration })) }} className="w-16 h-10 border border-slate-200 rounded-xl text-center text-sm font-black bg-slate-50/50 focus:bg-white transition-all outline-none focus:ring-1 focus:ring-blue-500 text-blue-600" />
                <span className="text-xs text-slate-400 font-bold">h</span>
              </div>
            </div>
            <div className="w-px h-10 bg-slate-100"></div>
            <div className="flex flex-col gap-1">
              <span className="text-[10px] font-black text-slate-400 uppercase tracking-wider">处方干体重 / 额外</span>
              <div className="flex items-center gap-2">
                <input type="text" value={sharedFields.dryWeight} onChange={(e) => { const v = parseFloat(e.target.value); setSharedFields(prev => ({ ...prev, dryWeight: Number.isFinite(v) ? v : prev.dryWeight })) }} className="w-20 h-10 border border-slate-200 rounded-xl text-center text-sm font-black bg-slate-50/50 focus:bg-white transition-all outline-none focus:ring-1 focus:ring-blue-500 text-red-600" />
                <span className="text-slate-300 font-bold">+</span>
                <input type="text" value={sharedFields.extraWeight} onChange={(e) => { const v = parseFloat(e.target.value); setSharedFields(prev => ({ ...prev, extraWeight: Number.isFinite(v) ? v : prev.extraWeight })) }} className="w-14 h-10 border border-slate-200 rounded-xl text-center text-sm font-black bg-slate-50/50 focus:bg-white transition-all outline-none focus:ring-1 focus:ring-blue-500" />
                <span className="text-xs text-slate-400 font-bold ml-1">kg</span>
              </div>
            </div>
          </div>
          <div className="flex-1 flex flex-col gap-1 min-w-[300px]">
            <span className="text-[10px] font-black text-slate-400 uppercase tracking-wider ml-1">血管通路与材料配置</span>
            <div className="flex items-center gap-3">
              <div className="relative flex-1">
                <input
                  type="text"
                  value={vascularAccesses.find(va => va.isDefault)
                    ? formatVascularAccessLabel(vascularAccesses.find(va => va.isDefault)!)
                    : vascularAccesses.length > 0
                      ? formatVascularAccessLabel(vascularAccesses[0])
                      : '未设置...'}
                  readOnly
                  className="w-full h-10 border border-slate-200 border-dashed rounded-xl px-4 text-sm font-black bg-slate-50 text-slate-700 cursor-not-allowed outline-none"
                />
              </div>
              <button
                onClick={() => { setSyncModalKey(k => k + 1); setIsVascularModalOpen(true) }}
                className="h-10 px-6 bg-white border-2 border-blue-100 text-blue-600 text-xs font-black rounded-xl hover:bg-blue-50 hover:border-blue-200 transition-all whitespace-nowrap shadow-sm"
              >
                变更通路设置
              </button>
            </div>
          </div>
        </div>

      {/* 如果方案为空，则显示缺省页 */}
      {!hasTreatmentPlan ? (
        <div className="bg-white rounded-[40px] border border-slate-200 p-20 text-center flex flex-col items-center justify-center space-y-4 shadow-sm">
          <div className="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center text-slate-200 mb-2">
            <ClipboardList size={40} />
          </div>
          <h4 className="text-lg font-black text-slate-800">暂未建立治疗方案</h4>
          <p className="text-sm text-slate-400 max-w-xs leading-relaxed font-bold">
            该患者目前无生效的透析处方方案，请点击上方"新建方案"按钮录入。
          </p>
        </div>
      ) : (
        <>

          {/* 主内容区 */}
          <div className="bg-white rounded-[40px] border border-slate-200 overflow-hidden shadow-sm ring-1 ring-slate-100">
            {/* 模式切换 Tab（显示所有有治疗方案的模式） */}
            {apiTreatmentPlans.length > 0 && (
              <div className="flex bg-slate-50/80 border-b border-slate-100 p-3 gap-3 backdrop-blur-sm">
                {apiTreatmentPlans.map(p => (
                  <button
                    key={p.dialysisMode.mode}
                    onClick={() => setSelectedMode(p.dialysisMode.mode)}
                    className={`flex-1 min-w-[100px] py-3 text-sm font-black rounded-xl border transition-all duration-300 ${
                      selectedMode === p.dialysisMode.mode
                        ? 'bg-white border-blue-200 text-blue-600 shadow-md ring-1 ring-blue-50'
                        : 'bg-transparent border-transparent text-slate-400 hover:text-slate-600 hover:bg-white/50'
                    }`}
                  >
                    {getDictName(DICT_TYPES.DIALYSIS_MODE, p.dialysisMode.mode)}
                  </button>
                ))}
              </div>
            )}

            {/* 详细参数区 */}
            <div className="p-10 space-y-12">
              <DialysisModeSection
                values={mainDialysisModeValues}
                onChange={handleMainDialysisModeChange}
                dictOptions={dictOptions}
                dictTypeKey={DICT_TYPES.DIALYSIS_MODE}
                openNumericPad={mainOpenNumericPad}
                dialysateFlow={String(displayPlan.parameters.flowRate)}
                onDialysateVolumeChange={(v) => setEditingPlan(prev => prev ? {...prev, parameters: {...prev.parameters, volume: parseFloat(v) || prev.parameters.volume}} : prev)}
                extraContent={
                  <>
                    <div className="flex flex-col gap-2">
                      <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">频次描述</label>
                      <input
                        type="text"
                        value={displayPlan.indicators.frequencyDesc}
                        onChange={(e) => setEditingPlan(prev => prev ? ({
                          ...prev,
                          indicators: { ...prev.indicators, frequencyDesc: e.target.value }
                        }) : prev)}
                        className="w-full h-11 px-4 border border-slate-200 rounded-xl text-sm bg-white font-black outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                      />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">自动确认处方</label>
                      <div className="flex items-center gap-10 h-11 bg-white border border-slate-200 rounded-xl px-6">
                        <label className="flex items-center gap-2 cursor-pointer group">
                          <input
                            type="radio"
                            name="auto_main"
                            checked={displayPlan.indicators.autoConfirm}
                            onChange={() => setEditingPlan(prev => prev ? ({
                              ...prev,
                              indicators: { ...prev.indicators, autoConfirm: true }
                            }) : prev)}
                            className="w-4 h-4 text-blue-600 focus:ring-0 border-slate-300"
                          />
                          <span className="text-sm font-bold text-slate-600 group-hover:text-blue-600 transition-colors">是</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer group">
                          <input
                            type="radio"
                            name="auto_main"
                            checked={!displayPlan.indicators.autoConfirm}
                            onChange={() => setEditingPlan(prev => prev ? ({
                              ...prev,
                              indicators: { ...prev.indicators, autoConfirm: false }
                            }) : prev)}
                            className="w-4 h-4 text-blue-600 focus:ring-0 border-slate-300"
                          />
                          <span className="text-sm font-bold text-slate-600 group-hover:text-blue-600 transition-colors">否</span>
                        </label>
                      </div>
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">方案启用状态</label>
                      <div className="flex items-center gap-10 h-11 bg-white border border-slate-200 rounded-xl px-6">
                        <label className="flex items-center gap-2 cursor-pointer group">
                          <input
                            type="radio"
                            name="status_main"
                            checked={displayPlan.indicators.status === '启用'}
                            onChange={() => setEditingPlan(prev => prev ? ({
                              ...prev,
                              indicators: { ...prev.indicators, status: '启用' }
                            }) : prev)}
                            className="w-4 h-4 text-green-600 focus:ring-0 border-slate-300"
                          />
                          <span className="text-sm font-bold text-slate-600 group-hover:text-green-600 transition-colors">启用</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer group">
                          <input
                            type="radio"
                            name="status_main"
                            checked={displayPlan.indicators.status !== '启用'}
                            onChange={() => setEditingPlan(prev => prev ? ({
                              ...prev,
                              indicators: { ...prev.indicators, status: '禁用' }
                            }) : prev)}
                            className="w-4 h-4 text-red-600 focus:ring-0 border-slate-300"
                          />
                          <span className="text-sm font-bold text-slate-600 group-hover:text-red-600 transition-colors">禁用</span>
                        </label>
                      </div>
                    </div>
                  </>
                }
              />

              <AnticoagulantSection
                values={mainAnticoagulantValues}
                onChange={handleMainAnticoagulantChange}
                drugOptions={mainAnticoagulantOptions}
                specMap={mainAnticoagulantSpecMap}
                openNumericPad={mainOpenNumericPad}
              />

              <DialysisParamsSection
                values={mainDialysisParamsValues}
                onChange={handleMainDialysisParamsChange}
                duration={String(displayPlan.duration)}
                dictOptions={dictOptions}
                dictTypeKeys={{
                  dialysateType: DICT_TYPES.DIALYSATE_TYPE,
                  dialysateGroup: DICT_TYPES.DIALYSATE_GROUP,
                  dialysateFlow: DICT_TYPES.DIALYSATE_FLOW,
                  glucose: DICT_TYPES.GLUCOSE,
                }}
                openNumericPad={mainOpenNumericPad}
              />

              <MaterialsSection
                materials={displayPlan.materials}
                onAdd={handleMainMaterialAdd}
                onRemove={handleMainMaterialRemove}
                onUpdate={handleMainMaterialUpdate}
                onReplace={handleMainMaterialReplace}
              />

            </div>
          </div>

          {/* 方案调整记录 - 独立区块 */}
          <div className="bg-white rounded-[40px] border border-slate-200 overflow-hidden shadow-sm ring-1 ring-slate-100 p-10">
            <div className="space-y-6">
              <h4 className="text-sm font-black text-slate-800 flex items-center gap-2.5">
                <div className="w-2 h-5 bg-indigo-500 rounded-full shadow-sm"></div> 方案调整记录
              </h4>
              <div className="border border-slate-100 rounded-[32px] overflow-hidden shadow-sm ring-1 ring-slate-100">
                <table className="w-full text-left text-sm border-separate border-spacing-0">
                  <thead className="bg-[#f8fbff] text-slate-500 font-black uppercase tracking-widest text-[10px] border-b border-blue-50">
                    <tr>
                      <th className="px-6 py-5 w-20 text-center">序号</th>
                      <th className="px-6 py-5">调整内容</th>
                      <th className="px-6 py-5 w-40">调整人</th>
                      <th className="px-6 py-5 w-64">调整时间</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-50">
                    {adjustmentRecords.length > 0 ? (
                      <>
                        {(showAllRecords ? adjustmentRecords : adjustmentRecords.slice(0, 5)).map((record, idx) => (
                          <tr key={record.id} className="hover:bg-blue-50/10 transition-colors">
                            <td className="px-6 py-5 text-slate-400 font-mono text-[11px] text-center">{idx + 1}</td>
                            <td className="px-6 py-5 text-red-500 font-bold">{record.content}</td>
                            <td className="px-6 py-5 text-slate-600 font-bold">{record.operator}</td>
                            <td className="px-6 py-5 text-slate-400 font-mono text-xs">{record.createdAt}</td>
                          </tr>
                        ))}
                        {adjustmentRecords.length > 5 && (
                          <tr>
                            <td colSpan={4} className="px-6 py-4 text-center">
                              <button
                                onClick={() => setShowAllRecords(!showAllRecords)}
                                className="text-indigo-500 hover:text-indigo-600 font-medium text-sm transition-colors"
                              >
                                {showAllRecords ? '收起 ▲' : `展开更多 (${adjustmentRecords.length - 5} 条) ▼`}
                              </button>
                            </td>
                          </tr>
                        )}
                      </>
                    ) : (
                      <tr>
                        <td colSpan={4} className="py-16 text-center text-slate-300 font-bold italic bg-white">
                          暂无方案调整记录数据
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </>
      )}

      {/* 主内容区数字键盘 */}
      <NumericPad
        open={mainPadState.open}
        label={mainPadState.label}
        value={mainPadState.value}
        onKeyPress={mainHandleKeyPress}
        onConfirm={mainHandleConfirm}
        onClear={mainHandleClear}
        onClose={mainCloseNumericPad}
      />

      {/* 新建方案弹窗 */}
      <NewPlanModal
        isOpen={isNewPlanModalOpen}
        onClose={() => setIsNewPlanModalOpen(false)}
        patientName={patientName}
        onSave={handleNewPlanSave}
      />

      {/* 材料同步弹窗 */}
      <MaterialSyncModal
        key={syncModalKey}
        isOpen={isVascularModalOpen}
        onClose={() => setIsVascularModalOpen(false)}
        onSave={handleMaterialSyncSave}
        vascularAccesses={syncModalVascularAccesses}
        currentAccessId={currentDefaultAccessId}
        plans={syncModalPlans}
      />
    </div>
  )
}
