import { useCallback, useEffect, useMemo, useState } from 'react'
import { message } from 'antd'
import { Calendar, ChevronDown, X } from 'lucide-react'

import { drugCatalogApi, orderTemplateApi } from '@/services/treatmentConfigApi'
import type { DrugCatalog, OrderTemplate, OrderTemplateItem } from '@/services/treatmentConfigApi'
import { orderApi } from '@/services/orderApi'
import type { CreateFromTemplateItemRequest, Order } from '@/services/orderApi'

function FormField({
  label,
  value,
  onChange,
  suffix,
  placeholder,
  required,
  readOnly = false,
}: {
  label: string
  value?: string
  onChange?: (value: string) => void
  suffix?: string
  placeholder?: string
  required?: boolean
  readOnly?: boolean
}) {
  return (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-bold text-slate-500">
        {required && <span className="text-red-500 mr-0.5">*</span>}
        {label}
      </label>
      <div className="relative">
        <input
          type="text"
          value={value || ''}
          onChange={e => onChange?.(e.target.value)}
          readOnly={readOnly}
          placeholder={placeholder}
          className={`w-full h-10 px-3 border rounded-lg text-sm outline-none transition-all ${
            readOnly
              ? 'bg-slate-50 text-slate-400 border-slate-200 cursor-not-allowed'
              : 'bg-white border-slate-300 focus:ring-1 focus:ring-blue-500 focus:border-blue-500'
          }`}
        />
        {suffix && <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] text-slate-400">{suffix}</span>}
      </div>
    </div>
  )
}

function SelectField({
  label,
  options,
  value,
  onChange,
}: {
  label: string
  options: string[]
  value?: string
  onChange?: (value: string) => void
}) {
  return (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-bold text-slate-500">{label}</label>
      <div className="relative">
        <select
          value={value || options[0] || ''}
          onChange={e => onChange?.(e.target.value)}
          className="w-full h-10 px-3 border border-slate-300 rounded-lg text-sm font-black text-slate-800 outline-none appearance-none focus:ring-1 focus:ring-blue-500 bg-white transition-all"
        >
          {options.map(option => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
        <ChevronDown size={18} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
      </div>
    </div>
  )
}

type TemplateDraftItem = {
  id: string
  selected: boolean
  name: string
  content: string
  dose: string
  unit: string
  route: string
  frequency: string
  timing: string
  execTiming: string
  spec: string
  groupId?: string
}

interface OrderModalProps {
  isOpen: boolean
  onClose: () => void
  type?: '长期' | '临时'
  patientId: string
  editOrder?: Order
  onSave?: () => void | Promise<void>
}

const ROUTE_OPTIONS = ['静脉推注', '体外循环', '管路注入', '皮下注射', '口服', '外用']
const FREQUENCY_OPTIONS = ['每次透析', '每周三次', '每周两次', 'ST', 'qd', 'bid', 'tid']
const TIMING_OPTIONS = ['透析开始', '随嘱', '开机', '即刻', '透析结束']
const EXEC_TIMING_OPTIONS = ['立即执行', '普通']

export default function OrderModal({
  isOpen,
  onClose,
  type = '长期',
  patientId,
  editOrder,
  onSave,
}: OrderModalProps) {
  const [activeTab, setActiveTab] = useState<'NEW' | 'TEMPLATE_GROUP'>('NEW')
  const [saving, setSaving] = useState(false)

  const [drugItems, setDrugItems] = useState<DrugCatalog[]>([])
  const [selectedDrugId, setSelectedDrugId] = useState<number | undefined>(undefined)
  const [templates, setTemplates] = useState<OrderTemplate[]>([])
  const [selectedTemplateId, setSelectedTemplateId] = useState('')
  const [templateDrafts, setTemplateDrafts] = useState<TemplateDraftItem[]>([])

  const [formName, setFormName] = useState('')
  const [formContent, setFormContent] = useState('')
  const [formDose, setFormDose] = useState('')
  const [formUnit, setFormUnit] = useState('')
  const [formRoute, setFormRoute] = useState('静脉推注')
  const [formFrequency, setFormFrequency] = useState('每次透析')
  const [formTiming, setFormTiming] = useState('透析开始')
  const [formExecTiming, setFormExecTiming] = useState('立即执行')
  const [formSpec, setFormSpec] = useState('')
  const [formStartDate, setFormStartDate] = useState('')
  const [formStopDate, setFormStopDate] = useState('')
  const [formNotes, setFormNotes] = useState('')
  const [formPriority, setFormPriority] = useState('普通')

  const isLongTerm = type === '长期'
  const isEditing = !!editOrder

  const currentDrug = useMemo(
    () => drugItems.find(item => item.id === selectedDrugId),
    [drugItems, selectedDrugId],
  )
  const templateGroupSeqMap = useMemo(() => {
    const seqMap = new Map<string, number>()
    let seqCounter = 1

    templateDrafts.forEach(item => {
      const groupId = item.groupId?.trim()
      if (!groupId || seqMap.has(groupId)) return
      seqMap.set(groupId, seqCounter)
      seqCounter += 1
    })

    return seqMap
  }, [templateDrafts])
  const isDrugType = currentDrug ? currentDrug.category !== 'METHOD' : true

  const resetForm = useCallback(() => {
    setFormName('')
    setFormContent('')
    setFormDose('')
    setFormUnit('')
    setFormRoute('静脉推注')
    setFormFrequency('每次透析')
    setFormTiming('透析开始')
    setFormExecTiming('立即执行')
    setFormSpec('')
    setFormStartDate(new Date().toISOString().slice(0, 10))
    setFormStopDate('')
    setFormNotes('')
    setFormPriority('普通')
    setSelectedDrugId(undefined)
    setSelectedTemplateId('')
    setTemplateDrafts([])
    setActiveTab('NEW')
  }, [])

  const fillFormFromOrder = useCallback((order: Order) => {
    setFormName(order.name || '')
    setFormContent(order.content || '')
    setFormDose(order.dose || '')
    setFormUnit(order.unit || '')
    setFormRoute(order.route || '静脉推注')
    setFormFrequency(order.frequency || '每次透析')
    setFormTiming(order.timing || '透析开始')
    setFormExecTiming(order.execTiming || '立即执行')
    setFormSpec(order.spec || '')
    setFormStartDate(order.startTime ? order.startTime.slice(0, 10) : new Date().toISOString().slice(0, 10))
    setFormStopDate(order.endTime ? order.endTime.slice(0, 10) : '')
    setFormNotes(order.notes || '')
    setFormPriority(order.priority || '普通')
    setSelectedDrugId(order.drugId)
    setSelectedTemplateId('')
    setTemplateDrafts([])
    setActiveTab('NEW')
  }, [])

  const loadDrugs = useCallback(async () => {
    try {
      const data = await drugCatalogApi.list({ pageSize: 200, isEnabled: true })
      setDrugItems(data.items || [])
    } catch (error) {
      console.warn('加载药品目录失败:', error)
      message.warning('药品目录加载失败，可手动录入')
    }
  }, [])

  const loadTemplates = useCallback(async () => {
    try {
      const data = await orderTemplateApi.list({ pageSize: 50, isEnabled: true })
      setTemplates(data.items || [])
    } catch (error) {
      console.warn('加载医嘱模板失败:', error)
      message.warning('医嘱模板加载失败')
    }
  }, [])

  useEffect(() => {
    if (!isOpen) return
    loadDrugs()
    loadTemplates()
    if (editOrder) {
      fillFormFromOrder(editOrder)
    } else {
      resetForm()
    }
  }, [editOrder, fillFormFromOrder, isOpen, loadDrugs, loadTemplates, resetForm])

  const handleDrugSelect = (drugId: number) => {
    setSelectedDrugId(drugId)
    const drug = drugItems.find(item => item.id === drugId)
    if (!drug) return

    setFormName(drug.name)
    setFormContent(drug.name)
    setFormSpec(drug.spec || '')
    setFormUnit(drug.baseUnit || drug.specUnit || '')
    if (drug.minUnitDose) {
      setFormDose(drug.minUnitDose)
    }
    if (isLongTerm && drug.timing) {
      setFormTiming(drug.timing)
    }
  }

  const toTemplateDraft = (item: OrderTemplateItem): TemplateDraftItem => ({
    id: item.id,
    selected: true,
    name: item.drugName || '',
    content: item.drugName || '',
    dose: item.dosage || '',
    unit: item.unit || '',
    route: item.route || '静脉推注',
    frequency: item.frequency || '每次透析',
    timing: item.timing || '透析开始',
    execTiming: '立即执行',
    spec: item.spec || '',
    groupId: item.groupId,
  })

  const handleTemplateSelect = async (templateId: string) => {
    setSelectedTemplateId(templateId)
    if (!templateId) {
      setTemplateDrafts([])
      return
    }

    try {
      const template = await orderTemplateApi.get(templateId)
      setTemplateDrafts((template.items || []).map(toTemplateDraft))
    } catch (error) {
      message.error('加载模板详情失败')
      console.warn('加载模板详情失败:', error)
    }
  }

  const updateTemplateDraft = (id: string, patch: Partial<TemplateDraftItem>) => {
    setTemplateDrafts(prev => prev.map(item => (item.id === id ? { ...item, ...patch } : item)))
  }

  const validateMainForm = () => {
    if (!formName.trim() && !formContent.trim()) {
      message.warning('医嘱项目和医嘱内容不能同时为空')
      return false
    }
    return true
  }

  const buildOrderPayload = () => ({
    type,
    category: isDrugType ? '药品' : '治疗',
    name: formName.trim() || undefined,
    content: formContent.trim() || undefined,
    dose: formDose.trim() || undefined,
    unit: formUnit.trim() || undefined,
    route: formRoute || undefined,
    timing: isLongTerm ? formTiming : undefined,
    execTiming: isLongTerm ? undefined : formExecTiming,
    drugId: selectedDrugId,
    spec: formSpec.trim() || undefined,
    frequency: isLongTerm ? formFrequency : undefined,
    priority: formPriority || undefined,
    startTime: formStartDate || undefined,
    endTime: isLongTerm ? (formStopDate || undefined) : undefined,
    notes: formNotes.trim() || undefined,
  })

  const handleSave = async () => {
    if (!validateMainForm()) return

    setSaving(true)
    try {
      if (isEditing && editOrder) {
        await orderApi.revise(patientId, editOrder.id, {
          category: isDrugType ? '药品' : '治疗',
          name: formName.trim() || undefined,
          content: formContent.trim() || undefined,
          dose: formDose.trim() || undefined,
          unit: formUnit.trim() || undefined,
          route: formRoute || undefined,
          timing: isLongTerm ? formTiming : undefined,
          execTiming: isLongTerm ? undefined : formExecTiming,
          drugId: selectedDrugId,
          spec: formSpec.trim() || undefined,
          frequency: isLongTerm ? formFrequency : undefined,
          priority: formPriority || undefined,
          startTime: formStartDate || undefined,
          stopDate: isLongTerm ? (formStopDate || undefined) : undefined,
          notes: formNotes.trim() || undefined,
        })
        message.success('医嘱已修订')
      } else {
        await orderApi.create(patientId, buildOrderPayload())
        message.success('医嘱已创建')
      }
      onSave?.()
    } catch (error) {
      message.error((isEditing ? '修订' : '创建') + '失败: ' + (error instanceof Error ? error.message : '未知错误'))
    } finally {
      setSaving(false)
    }
  }

  const handleCreateFromTemplate = async () => {
    if (!selectedTemplateId) {
      message.warning('请先选择模板')
      return
    }

    const selectedItems = templateDrafts.filter(item => item.selected)
    if (selectedItems.length === 0) {
      message.warning('请至少选择一个模板条目')
      return
    }

    const items: CreateFromTemplateItemRequest[] = selectedItems.map(item => ({
      templateItemId: item.id,
      name: item.name.trim() || undefined,
      content: item.content.trim() || undefined,
      dose: item.dose.trim() || undefined,
      unit: item.unit.trim() || undefined,
      route: item.route.trim() || undefined,
      frequency: isLongTerm ? (item.frequency.trim() || undefined) : undefined,
      timing: isLongTerm ? (item.timing.trim() || undefined) : undefined,
      execTiming: isLongTerm ? undefined : (item.execTiming.trim() || undefined),
      spec: item.spec.trim() || undefined,
    }))

    setSaving(true)
    try {
      await orderApi.createFromTemplate(patientId, {
        templateId: selectedTemplateId,
        type,
        items,
      })
      message.success('从模板创建成功')
      onSave?.()
    } catch (error) {
      message.error('从模板创建失败: ' + (error instanceof Error ? error.message : '未知错误'))
    } finally {
      setSaving(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-[1240px] max-h-[90vh] overflow-hidden flex flex-col ring-1 ring-black/5">
        <div className="bg-[#eef6ff] px-10 flex items-end h-[72px] relative border-b border-blue-100 shrink-0">
          <h3 className="text-lg font-black text-slate-800 mb-4 mr-12">{isEditing ? '修订' : '新增'}患者{type}医嘱</h3>
          {!isEditing && (
            <div className="flex items-end gap-2 h-full">
              <button
                onClick={() => setActiveTab('NEW')}
                className={`px-10 py-3.5 text-sm font-black rounded-t-2xl transition-all ${
                  activeTab === 'NEW' ? 'bg-white text-blue-600 shadow-[0_-4px_10px_rgba(0,0,0,0.03)] z-10 -mb-[1px]' : 'text-slate-400 hover:text-slate-600'
                }`}
              >
                新增
              </button>
              <button
                onClick={() => setActiveTab('TEMPLATE_GROUP')}
                className={`px-10 py-3.5 text-sm font-black rounded-t-2xl transition-all ${
                  activeTab === 'TEMPLATE_GROUP' ? 'bg-white text-blue-600 shadow-[0_-4px_10px_rgba(0,0,0,0.03)] z-10 -mb-[1px]' : 'text-slate-400 hover:text-slate-600'
                }`}
              >
                选择模版组
              </button>
            </div>
          )}
          <button onClick={onClose} className="absolute right-6 top-1/2 -translate-y-1/2 p-2 hover:bg-white/50 rounded-full transition-all text-slate-400 hover:text-slate-600">
            <X size={20} />
          </button>
        </div>

        <div className="p-10 bg-white flex-1 overflow-y-auto no-scrollbar">
          {activeTab === 'NEW' ? (
            <div className="space-y-8">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-x-10 gap-y-6">
                <div className="flex flex-col gap-2">
                  <label className="text-[11px] font-bold text-slate-500">医嘱项目</label>
                  <div className="relative">
                    <select
                      value={selectedDrugId || ''}
                      onChange={e => {
                        const nextId = Number(e.target.value)
                        if (nextId) handleDrugSelect(nextId)
                        else setSelectedDrugId(undefined)
                      }}
                      className="w-full h-10 px-3 border border-slate-300 rounded-lg text-sm font-black text-slate-800 outline-none appearance-none focus:ring-1 focus:ring-blue-500 bg-white transition-all"
                    >
                      <option value="">请选择药品/方法</option>
                      {drugItems.map(item => (
                        <option key={item.id} value={item.id}>
                          {item.name}{item.spec ? ` (${item.spec})` : ''}
                        </option>
                      ))}
                    </select>
                    <ChevronDown size={18} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                <FormField label="医嘱内容" value={formContent} onChange={setFormContent} placeholder="未填写时默认取医嘱项目" />
                <FormField label="备注" value={formNotes} onChange={setFormNotes} placeholder="如有特殊说明请填写" />

                {isDrugType && currentDrug && (
                  <>
                    <FormField label="规格" value={formSpec} readOnly />
                    <FormField label="最小单位剂量" value={currentDrug.minUnitDose} suffix={currentDrug.baseUnit} readOnly />
                  </>
                )}

                <FormField label="医嘱项目名称" value={formName} onChange={setFormName} placeholder="可留空" />
                <FormField label="剂量" value={formDose} onChange={setFormDose} suffix={formUnit} placeholder="可留空" />
                <FormField label="单位" value={formUnit} onChange={setFormUnit} />
                <SelectField label="用法" options={ROUTE_OPTIONS} value={formRoute} onChange={setFormRoute} />
                {isLongTerm ? (
                  <>
                    <SelectField label="频次" options={FREQUENCY_OPTIONS} value={formFrequency} onChange={setFormFrequency} />
                    <SelectField label="使用时机" options={TIMING_OPTIONS} value={formTiming} onChange={setFormTiming} />
                  </>
                ) : (
                  <SelectField label="执行时机" options={EXEC_TIMING_OPTIONS} value={formExecTiming} onChange={setFormExecTiming} />
                )}

                <div className="flex flex-col gap-2">
                  <label className="text-[11px] font-bold text-slate-500">{isLongTerm ? '开始日期' : '治疗日期'}</label>
                  <div className="relative">
                    <input
                      type="date"
                      value={formStartDate}
                      onChange={e => setFormStartDate(e.target.value)}
                      className="w-full h-10 px-3 border border-slate-300 rounded-lg text-sm font-black focus:ring-1 focus:ring-blue-500 outline-none"
                    />
                    <Calendar size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                {isLongTerm && (
                  <div className="flex flex-col gap-2">
                    <label className="text-[11px] font-bold text-slate-500">停用日期</label>
                    <div className="relative">
                      <input
                        type="date"
                        value={formStopDate}
                        onChange={e => setFormStopDate(e.target.value)}
                        className="w-full h-10 px-3 border border-slate-300 rounded-lg text-sm font-black focus:ring-1 focus:ring-blue-500 outline-none"
                      />
                      <Calendar size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                    </div>
                  </div>
                )}
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              <div className="grid grid-cols-2 gap-x-12 border-2 border-blue-500 p-8 rounded-xl shadow-sm bg-white">
                <div className="flex flex-col gap-2">
                  <label className="text-[11px] font-bold text-slate-500">选择医嘱模板</label>
                  <div className="relative">
                    <select
                      value={selectedTemplateId}
                      onChange={e => handleTemplateSelect(e.target.value)}
                      className="w-full h-10 px-3 border border-slate-300 rounded-lg text-sm font-black text-slate-800 outline-none appearance-none focus:ring-1 focus:ring-blue-500 bg-white transition-all"
                    >
                      <option value="">请选择模板</option>
                      {templates.map(template => (
                        <option key={template.id} value={template.id}>
                          {template.name} ({template.category})
                        </option>
                      ))}
                    </select>
                    <ChevronDown size={18} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
                  </div>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-[11px] font-bold text-slate-500">业务分类</label>
                  <input
                    type="text"
                    value={templates.find(template => template.id === selectedTemplateId)?.category || ''}
                    readOnly
                    className="w-full h-10 px-3 border border-slate-300 rounded-lg text-sm font-bold outline-none bg-slate-50 text-slate-400 cursor-not-allowed"
                  />
                </div>

                <div className="col-span-2 mt-8">
                  <div className="border border-slate-100 rounded-2xl overflow-hidden shadow-sm bg-slate-50/30">
                    <table className="w-full text-left text-xs border-collapse">
                      <thead className="bg-[#f8faff] text-slate-400 font-bold uppercase tracking-widest text-[9px] border-b border-slate-100">
                        <tr>
                          <th className="py-4 px-4 w-[60px] text-center">选择</th>
                          <th className="py-4 px-4 w-[60px] text-center">序号</th>
                          <th className="py-4 px-4 min-w-[160px]">医嘱项目</th>
                          <th className="py-4 px-4 min-w-[160px]">医嘱内容</th>
                          <th className="py-4 px-4 w-[90px] text-center">剂量</th>
                          <th className="py-4 px-4 w-[80px] text-center">单位</th>
                          <th className="py-4 px-4 w-[100px] text-center">用法</th>
                          {isLongTerm ? (
                            <>
                              <th className="py-4 px-4 w-[100px] text-center">频次</th>
                              <th className="py-4 px-4 w-[110px] text-center">使用时机</th>
                            </>
                          ) : (
                            <th className="py-4 px-4 w-[110px] text-center">执行时机</th>
                          )}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100 bg-white">
                        {templateDrafts.length === 0 ? (
                          <tr>
                            <td colSpan={isLongTerm ? 9 : 8} className="py-8 text-center text-slate-300 font-bold italic">
                              请先选择模板
                            </td>
                          </tr>
                        ) : (
                          templateDrafts.map((item, index) => {
                            const groupId = item.groupId?.trim()
                            const groupSeq = groupId ? templateGroupSeqMap.get(groupId) : undefined
                            const groupedRowClass = groupSeq ? 'bg-blue-50/40 hover:bg-blue-50/60' : 'hover:bg-slate-50/80'
                            const groupedCellClass = groupSeq ? 'border-l-4 border-l-blue-400 bg-blue-50/40' : ''

                            return (
                            <tr key={item.id} className={`${groupedRowClass} transition-colors`}>
                              <td className={`py-4 px-4 text-center ${groupedCellClass}`}>
                                <input
                                  type="checkbox"
                                  checked={item.selected}
                                  onChange={e => updateTemplateDraft(item.id, { selected: e.target.checked })}
                                  className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-0 cursor-pointer"
                                />
                              </td>
                              <td className="py-4 px-4 text-center font-bold text-slate-400">
                                <div className="flex flex-col items-center gap-1">
                                  <span>{index + 1}</span>
                                  {groupSeq && (
                                    <span className="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-[10px] font-black text-blue-700">
                                      组合{groupSeq}
                                    </span>
                                  )}
                                </div>
                              </td>
                              <td className="py-3 px-4">
                                <input
                                  value={item.name}
                                  onChange={e => updateTemplateDraft(item.id, { name: e.target.value })}
                                  className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm font-bold outline-none focus:border-blue-400"
                                />
                              </td>
                              <td className="py-3 px-4">
                                <input
                                  value={item.content}
                                  onChange={e => updateTemplateDraft(item.id, { content: e.target.value })}
                                  className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-blue-400"
                                />
                              </td>
                              <td className="py-3 px-4">
                                <input
                                  value={item.dose}
                                  onChange={e => updateTemplateDraft(item.id, { dose: e.target.value })}
                                  className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm font-mono outline-none focus:border-blue-400"
                                />
                              </td>
                              <td className="py-3 px-4">
                                <input
                                  value={item.unit}
                                  onChange={e => updateTemplateDraft(item.id, { unit: e.target.value })}
                                  className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-blue-400"
                                />
                              </td>
                              <td className="py-3 px-4">
                                <input
                                  value={item.route}
                                  onChange={e => updateTemplateDraft(item.id, { route: e.target.value })}
                                  className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-blue-400"
                                />
                              </td>
                              {isLongTerm ? (
                                <>
                                  <td className="py-3 px-4">
                                    <input
                                      value={item.frequency}
                                      onChange={e => updateTemplateDraft(item.id, { frequency: e.target.value })}
                                      className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-blue-400"
                                    />
                                  </td>
                                  <td className="py-3 px-4">
                                    <input
                                      value={item.timing}
                                      onChange={e => updateTemplateDraft(item.id, { timing: e.target.value })}
                                      className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-blue-400"
                                    />
                                  </td>
                                </>
                              ) : (
                                <td className="py-3 px-4">
                                  <input
                                    value={item.execTiming}
                                    onChange={e => updateTemplateDraft(item.id, { execTiming: e.target.value })}
                                    className="w-full h-9 px-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-blue-400"
                                  />
                                </td>
                              )}
                            </tr>
                          )})
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="px-10 py-8 bg-slate-50 border-t border-slate-200 flex justify-end gap-3 shrink-0">
          <button onClick={onClose} className="px-8 py-3 rounded-2xl border border-slate-200 text-slate-500 font-bold hover:bg-white transition-all">
            取消
          </button>
          {activeTab === 'TEMPLATE_GROUP' ? (
            <button
              onClick={handleCreateFromTemplate}
              disabled={saving || !selectedTemplateId}
              className="px-12 py-3.5 rounded-2xl bg-blue-600 text-white font-black shadow-xl shadow-blue-100 hover:bg-blue-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? '保存中...' : '从模板创建'}
            </button>
          ) : (
            <button
              onClick={handleSave}
              disabled={saving}
              className="px-12 py-3.5 rounded-2xl bg-blue-600 text-white font-black shadow-xl shadow-blue-100 hover:bg-blue-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? '保存中...' : '确认并保存'}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
