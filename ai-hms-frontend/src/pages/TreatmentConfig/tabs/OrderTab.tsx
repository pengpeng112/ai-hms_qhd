// OrderTab - 医嘱模板管理
// 自包含组件，包含所有医嘱模板的状态管理和 CRUD 操作

import { useState, useEffect, useRef, useMemo, useCallback, memo, Fragment } from 'react'
import {
  Syringe, Search, Plus, Edit3, Trash2, Trash,
  X, Layers, Link, Unlink, Hash, Loader2, ChevronDown
} from 'lucide-react'
import {
  orderTemplateApi,
  drugCatalogApi,
  type OrderTemplate,
  type DrugCatalog,
  type OrderTemplateItemRequest
} from '@/services/treatmentConfigApi'
import { DICT_TYPES } from '@/services/dictApi'
import { InputField, SelectField } from '../components'

// --- Props 接口 ---
interface OrderTabProps {
  dictOptions: Record<string, Array<{ value: string; label: string }>>
  onRefreshDict: () => Promise<void>
}

// --- 医嘱条目类型 ---
interface OrderEntry {
  id: string
  drugId?: number
  drugName: string
  spec: string
  minUnitDose: string
  dosage: string
  unit: string
  route: string
  frequency: string
  timing: string
  selected?: boolean
  groupId?: string
  category?: string // 药品分类：METHOD=方法，其他=药品
}

// --- 内联输入组件 ---
const InlineInput = ({ value, onChange, placeholder, center }: {
  value: string; onChange: (v: string) => void; placeholder?: string; center?: boolean
}) => (
  <input
    type="text"
    value={value}
    onChange={(e) => onChange(e.target.value)}
    placeholder={placeholder}
    className={`w-full bg-transparent border-none focus:ring-0 text-[11px] font-black text-slate-800 placeholder:text-slate-300 placeholder:font-normal ${center ? 'text-center' : ''}`}
  />
)

// --- 内联下拉组件（可编辑 combobox + 可禁用）---
const InlineComboSelect = ({ value, options, onChange, disabled }: {
  value: string; options: Array<{ value: string; label: string }>; onChange: (v: string) => void; disabled?: boolean
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [inputText, setInputText] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const [dropdownPos, setDropdownPos] = useState({ top: 0, left: 0, width: 0 })

  // value → label 显示映射
  const displayText = useMemo(() => {
    if (!value) return ''
    const opt = options.find(o => o.value === value)
    return opt ? opt.label : value
  }, [value, options])

  // 根据输入过滤选项
  const filteredOptions = useMemo(() => {
    if (!inputText) return options
    return options.filter(o => o.label.includes(inputText))
  }, [options, inputText])

  const handleFocus = () => {
    if (disabled) return
    if (inputRef.current) {
      const rect = inputRef.current.getBoundingClientRect()
      setDropdownPos({ top: rect.bottom + 2, left: rect.left, width: Math.max(rect.width, 140) })
    }
    setInputText(displayText)
    setIsOpen(true)
  }

  const handleBlur = () => {
    setTimeout(() => {
      setIsOpen(false)
      const text = inputText
      // 匹配已有选项则存 code，否则存自定义文本
      const matched = options.find(o => o.label === text)
      if (matched) onChange(matched.value)
      else if (text !== displayText) onChange(text)
    }, 150)
  }

  const handleSelect = (opt: { value: string; label: string }) => {
    onChange(opt.value)
    setIsOpen(false)
  }

  return (
    <div className="relative group/combo">
      <input
        ref={inputRef}
        type="text"
        value={isOpen ? inputText : displayText}
        onChange={(e) => { setInputText(e.target.value); if (!isOpen) setIsOpen(true) }}
        onFocus={handleFocus}
        onBlur={handleBlur}
        disabled={disabled}
        placeholder="--"
        className={`w-full bg-transparent border-none focus:ring-0 text-[11px] font-black text-slate-800 placeholder:text-slate-300 placeholder:font-normal pr-4 ${disabled ? 'opacity-40 cursor-not-allowed' : ''}`}
      />
      <ChevronDown size={10} className={`absolute right-0 top-1/2 -translate-y-1/2 pointer-events-none transition-colors ${disabled ? 'text-slate-200' : 'text-slate-300 group-hover/combo:text-indigo-500'}`} />
      {isOpen && filteredOptions.length > 0 && (
        <div
          className="fixed bg-white border border-slate-200 rounded-xl shadow-2xl z-[9999] p-1 max-h-40 overflow-y-auto no-scrollbar ring-4 ring-black/5"
          style={{ top: dropdownPos.top, left: dropdownPos.left, width: Math.max(dropdownPos.width, 140), maxWidth: 220 }}
        >
          {filteredOptions.map(opt => (
            <button
              key={opt.value}
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => handleSelect(opt)}
              className="w-full text-left px-3 py-1.5 hover:bg-indigo-50 text-[10px] font-bold text-slate-600 rounded-lg transition-colors truncate"
              title={opt.label}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

// --- 药品搜索下拉组件 ---
function DrugSearchInput({ value, onSelect, placeholder }: {
  value: string
  onSelect: (drug: DrugCatalog | null, text: string) => void
  placeholder?: string
}) {
  const [isOpen, setIsOpen] = useState(false)
  const [searchText, setSearchText] = useState(value)
  const [results, setResults] = useState<DrugCatalog[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const [dropdownPos, setDropdownPos] = useState({ top: 0, left: 0 })

  // 同步外部 value 变化
  useEffect(() => { setSearchText(value) }, [value])

  // 卸载时清理定时器
  useEffect(() => {
    return () => { if (timerRef.current) clearTimeout(timerRef.current) }
  }, [])

  // 计算下拉框位置（fixed 定位，不受 overflow 裁剪）
  const updateDropdownPos = useCallback(() => {
    if (!inputRef.current) return
    const rect = inputRef.current.getBoundingClientRect()
    setDropdownPos({ top: rect.bottom + 4, left: rect.left })
  }, [])

  const doSearch = useCallback(async (keyword: string) => {
    setIsSearching(true)
    try {
      const params: Record<string, string | number | boolean | undefined> = { pageSize: 20, isEnabled: true }
      if (keyword.trim()) params.search = keyword
      const res = await drugCatalogApi.list(params)
      setResults(res.items)
    } catch { setResults([]) }
    finally { setIsSearching(false) }
  }, [])

  const handleChange = (text: string) => {
    setSearchText(text)
    onSelect(null, text) // 清除 drugId，保留文本
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => doSearch(text), 300)
  }

  const handleSelect = (drug: DrugCatalog) => {
    setSearchText(drug.name)
    onSelect(drug, drug.name)
    setIsOpen(false)
  }

  const handleFocus = () => {
    updateDropdownPos()
    setIsOpen(true)
    doSearch(searchText)
  }

  return (
    <div className="relative w-full group/search">
      <input
        ref={inputRef}
        type="text"
        value={searchText}
        onChange={(e) => handleChange(e.target.value)}
        onFocus={handleFocus}
        onBlur={() => setTimeout(() => setIsOpen(false), 200)}
        placeholder={placeholder}
        className="w-full bg-transparent border-none focus:ring-0 text-[11px] font-black text-slate-800 placeholder:text-slate-300 placeholder:font-normal pr-4"
      />
      <div className="absolute right-0 top-1/2 -translate-y-1/2 opacity-0 group-hover/search:opacity-100 transition-opacity">
        {isSearching ? <Loader2 size={10} className="text-slate-300 animate-spin" /> : <Search size={10} className="text-slate-300" />}
      </div>
      {isOpen && results.length > 0 && (
        <div
          className="fixed w-72 bg-white border border-slate-200 rounded-xl shadow-2xl z-[9999] p-1 max-h-56 overflow-y-auto no-scrollbar ring-4 ring-black/5"
          style={{ top: dropdownPos.top, left: dropdownPos.left }}
        >
          {results.map(drug => (
            <button
              key={drug.id}
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => handleSelect(drug)}
              className="w-full text-left px-3 py-2 hover:bg-blue-50 text-[10px] font-black text-slate-600 rounded-lg transition-colors border-b border-slate-50 last:border-0"
            >
              <span>{drug.name}</span>
              {drug.spec && <span className="ml-2 text-slate-400 font-normal">{drug.spec}</span>}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

// --- 主组件 ---
function OrderTabComponent({ dictOptions, onRefreshDict }: OrderTabProps) {
  const [orderTemplates, setOrderTemplates] = useState<OrderTemplate[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [showModal, setShowModal] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [searchKeyword, setSearchKeyword] = useState('')

  // 医嘱模板表单状态
  const [orderGroupForm, setOrderGroupForm] = useState({ name: '', type: '基础类' })

  // 模态框中的医嘱列表
  const [modalOrders, setModalOrders] = useState<OrderEntry[]>([])

  // 字典选项
  const routeOptions = dictOptions[DICT_TYPES.ORDER_ROUTE] || []
  const frequencyOptions = dictOptions[DICT_TYPES.ORDER_FREQUENCY] || []
  const timingOptions = dictOptions[DICT_TYPES.ORDER_TIMING] || []

  // 计算分组的医嘱
  const groupedOrders = useMemo(() => modalOrders.reduce((acc, order) => {
    if (order.groupId) {
      const existingGroup = acc.find(g => g.groupId === order.groupId)
      if (existingGroup) { existingGroup.items.push(order) }
      else { acc.push({ type: 'group' as const, groupId: order.groupId, items: [order] }) }
    } else {
      acc.push({ type: 'single' as const, items: [order] })
    }
    return acc
  }, [] as Array<{ type: 'group' | 'single'; groupId?: string; items: OrderEntry[] }>), [modalOrders])

  // 加载数据
  const loadData = async () => {
    setIsLoading(true)
    try {
      const response = await orderTemplateApi.list({ pageSize: 9999 })
      setOrderTemplates(response.items)
    } catch (error) {
      console.error('加载医嘱模板失败:', error)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => { loadData() }, [])

  // 打开新增弹窗
  const handleOpenAddModal = async () => {
    await onRefreshDict()
    setIsEditing(false)
    setEditingId(null)
    setOrderGroupForm({ name: '', type: '基础类' })
    setModalOrders([])
    setShowModal(true)
  }

  // 打开编辑弹窗
  const handleOpenEditModal = async (template: OrderTemplate) => {
    await onRefreshDict()
    setIsEditing(true)
    setEditingId(template.id)
    setOrderGroupForm({ name: template.name, type: template.category })
    // 从详情 API 获取完整数据（含 items）
    try {
      const detail = await orderTemplateApi.get(template.id)
      if (detail.items && detail.items.length > 0) {
        // 批量查询药品分类信息
        const drugIds = detail.items.filter(i => i.drugId).map(i => i.drugId!)
        const categoryMap = new Map<number, string>()
        if (drugIds.length > 0) {
          try {
            const drugsRes = await drugCatalogApi.list({ pageSize: 9999 })
            drugsRes.items.forEach(d => categoryMap.set(d.id, d.category))
          } catch { /* 查询失败不阻塞加载 */ }
        }
        setModalOrders(detail.items.map(item => ({
          id: item.id || `O_${crypto.randomUUID()}`,
          drugId: item.drugId,
          drugName: item.drugName,
          spec: item.spec || '',
          minUnitDose: item.minUnitDose || '',
          dosage: item.dosage || '',
          unit: item.unit || '',
          route: item.route || '',
          frequency: item.frequency || '',
          timing: item.timing || '',
          groupId: item.groupId,
          category: item.drugId ? categoryMap.get(item.drugId) : undefined,
        })))
      } else {
        setModalOrders([])
      }
    } catch (error) {
      console.error('加载医嘱模板详情失败:', error)
      setModalOrders([])
    }
    setShowModal(true)
  }

  // 删除模板
  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除该医嘱模板吗？')) return
    try {
      await orderTemplateApi.delete(id)
      await loadData()
    } catch (error) {
      console.error('删除医嘱模板失败:', error)
      alert('删除失败，请稍后重试')
    }
  }

  // 保存医嘱
  const handleSave = async () => {
    if (!orderGroupForm.name) { alert('医嘱组名称为必填项'); return }
    if (modalOrders.length === 0) { alert('请至少添加一条医嘱'); return }

    setIsLoading(true)
    try {
      const items: OrderTemplateItemRequest[] = modalOrders.map((o, idx) => ({
        drugId: o.drugId,
        drugName: o.drugName,
        spec: o.spec,
        minUnitDose: o.minUnitDose,
        dosage: o.dosage,
        unit: o.unit,
        route: o.route,
        frequency: o.frequency,
        timing: o.timing,
        groupId: o.groupId,
        sortOrder: idx,
      }))

      // 兼容旧 content 字段
      const content = modalOrders
        .map(o => `${o.drugName} ${o.dosage}${o.unit} ${o.route} ${o.frequency} ${o.timing}`)
        .filter(c => c.trim()).join('\n')

      const orderData = {
        name: orderGroupForm.name,
        type: '长期' as const,
        category: orderGroupForm.type,
        content,
        priority: '普通',
        isEnabled: true,
        items,
      }

      if (isEditing && editingId) {
        await orderTemplateApi.update(editingId, orderData)
      } else {
        await orderTemplateApi.create(orderData)
      }
      await loadData()
      setShowModal(false)
      alert(isEditing ? '医嘱模板更新成功！' : '医嘱模板保存成功！')
    } catch (error) {
      console.error('保存医嘱失败:', error)
      alert('保存失败，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }

  // 医嘱条目操作
  const handleAddOrderEntry = () => {
    setModalOrders([...modalOrders, {
      id: `O${Date.now()}`, drugName: '', spec: '', minUnitDose: '',
      dosage: '', unit: '', route: '', frequency: '', timing: '',
    }])
  }

  const handleRemoveOrderEntry = (id: string) => {
    setModalOrders(modalOrders.filter(o => o.id !== id))
  }

  const handleUpdateOrderEntry = (id: string, field: string, value: string | boolean | number | undefined) => {
    setModalOrders(modalOrders.map(o => o.id === id ? { ...o, [field]: value } : o))
  }

  // 药品搜索选中回调
  const handleDrugSelect = (orderId: string, drug: DrugCatalog | null, text: string) => {
    setModalOrders(modalOrders.map(o => {
      if (o.id !== orderId) return o
      if (drug) {
        const isMethod = drug.category === 'METHOD'
        return {
          ...o,
          drugId: drug.id,
          drugName: drug.name,
          spec: drug.spec || '',
          minUnitDose: drug.minUnitDose || '',
          unit: drug.baseUnit || '',
          category: drug.category,
          // 方法类清空用法/频次/使用时机
          route: isMethod ? '' : o.route,
          frequency: isMethod ? '' : o.frequency,
          timing: isMethod ? '' : o.timing,
        }
      }
      return { ...o, drugId: undefined, drugName: text, category: undefined }
    }))
  }

  const handleToggleSelectAllOrders = () => {
    const allSelected = modalOrders.every(o => o.selected)
    setModalOrders(modalOrders.map(o => ({ ...o, selected: !allSelected })))
  }

  const handleCombineOrders = () => {
    const selectedOrders = modalOrders.filter(o => o.selected)
    if (selectedOrders.length < 2) { alert('请至少选择2条医嘱进行合并'); return }
    const groupId = `G${Date.now()}`
    setModalOrders(modalOrders.map(o => o.selected ? { ...o, groupId } : o))
  }

  const handleUncombineOrders = () => {
    const selectedOrders = modalOrders.filter(o => o.selected && o.groupId)
    if (selectedOrders.length === 0) { alert('请选择已合并的医嘱进行拆分'); return }
    const groupIds = [...new Set(selectedOrders.map(o => o.groupId))]
    setModalOrders(modalOrders.map(o => (o.groupId && groupIds.includes(o.groupId)) ? { ...o, groupId: undefined, selected: false } : o))
  }

  // 过滤数据
  const filteredTemplates = orderTemplates.filter(order =>
    !searchKeyword || order.name.toLowerCase().includes(searchKeyword.toLowerCase())
  )

  if (isLoading && orderTemplates.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-slate-400 animate-spin" />
        <span className="ml-3 text-slate-500">加载中...</span>
      </div>
    )
  }

  // --- 渲染单行医嘱的表格单元格 ---
  // 列宽由 <colgroup> 统一控制，td 不再重复设置 w-* 类
  const renderOrderCells = (o: OrderEntry, grouped = false) => {
    const isMethod = o.category === 'METHOD'
    return (
    <>
      {!grouped && (
        <td className="px-4 py-3 text-center">
          <input type="checkbox" checked={o.selected} onChange={(e) => handleUpdateOrderEntry(o.id, 'selected', e.target.checked)} className="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500" />
        </td>
      )}
      {grouped && <td className="px-4 py-3" />}
      <td className="px-4 py-3 relative z-[60]">
        <DrugSearchInput value={o.drugName} onSelect={(drug, text) => handleDrugSelect(o.id, drug, text)} placeholder="搜索或输入药品/方法" />
      </td>
      <td className="px-4 py-3 text-center">
        <InlineInput value={o.spec} onChange={(v) => handleUpdateOrderEntry(o.id, 'spec', v)} placeholder="规格" center />
      </td>
      <td className="px-4 py-3 text-center">
        <InlineInput value={o.minUnitDose} onChange={(v) => handleUpdateOrderEntry(o.id, 'minUnitDose', v)} placeholder="数值" center />
      </td>
      <td className="px-4 py-3 text-center text-blue-600">
        <InlineInput value={o.dosage} onChange={(v) => handleUpdateOrderEntry(o.id, 'dosage', v)} placeholder="剂量" center />
      </td>
      <td className="px-4 py-3">
        <InlineInput value={o.unit} onChange={(v) => handleUpdateOrderEntry(o.id, 'unit', v)} placeholder="单位" />
      </td>
      <td className="px-4 py-3">
        <InlineComboSelect value={o.route} options={routeOptions} onChange={(v) => handleUpdateOrderEntry(o.id, 'route', v)} disabled={isMethod} />
      </td>
      <td className="px-4 py-3">
        <InlineComboSelect value={o.frequency} options={frequencyOptions} onChange={(v) => handleUpdateOrderEntry(o.id, 'frequency', v)} disabled={isMethod} />
      </td>
      <td className="px-4 py-3">
        <InlineComboSelect value={o.timing} options={timingOptions} onChange={(v) => handleUpdateOrderEntry(o.id, 'timing', v)} disabled={isMethod} />
      </td>
      <td className="px-4 py-3 text-right">
        <button onClick={() => handleRemoveOrderEntry(o.id)} className="p-2 text-indigo-300 hover:text-red-500 hover:bg-red-50 rounded-xl transition-all">
          <Trash size={14}/>
        </button>
      </td>
    </>
    )
  }

  return (
    <div className="flex-1 flex flex-col animate-fade-in">
      {/* 顶部工具栏 */}
      <div className="bg-white p-4 border-b border-slate-100 flex justify-between items-center shrink-0">
        <div className="relative w-80">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"/>
          <input type="text" placeholder="搜索医嘱模板..." value={searchKeyword} onChange={(e) => setSearchKeyword(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all" />
        </div>
        <button onClick={handleOpenAddModal} className="flex items-center gap-2 px-5 py-2 bg-indigo-600 text-white rounded-xl text-xs font-black hover:bg-indigo-700 shadow-lg shadow-indigo-100 transition-all">
          <Plus size={14}/> 新增医嘱
        </button>
      </div>

      {/* 列表区域 */}
      <div className="flex-1 overflow-auto p-6">
        {filteredTemplates.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-slate-400">
            <Syringe size={48} className="text-slate-300 mb-4" />
            <p className="text-sm font-bold">暂无医嘱模板</p>
            <p className="text-xs mt-1">点击"新增医嘱"创建第一个模板</p>
          </div>
        ) : (
          <div className="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-[10px] font-black uppercase text-slate-400">
                <tr>
                  <th className="px-8 py-5">模板名称</th>
                  <th className="px-8 py-5">医嘱类型</th>
                  <th className="px-8 py-5">医嘱分类</th>
                  <th className="px-8 py-5">优先级</th>
                  <th className="px-8 py-5">状态</th>
                  <th className="px-8 py-5 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50 font-bold text-slate-700">
                {filteredTemplates.map(order => (
                  <tr key={order.id} className="hover:bg-indigo-50/30 transition-colors">
                    <td className="px-8 py-6 flex items-center gap-3">
                      <div className="p-2 bg-indigo-50 text-indigo-600 rounded-xl"><Syringe size={16}/></div>
                      {order.name}
                    </td>
                    <td className="px-8 py-6">
                      <span className="px-2 py-0.5 bg-slate-100 text-slate-500 rounded text-[10px] font-black">{order.type}</span>
                    </td>
                    <td className="px-8 py-6">{order.category}</td>
                    <td className="px-8 py-6">{order.priority}</td>
                    <td className="px-8 py-6">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-black ${order.isEnabled ? 'bg-emerald-100 text-emerald-600' : 'bg-slate-100 text-slate-400'}`}>
                        {order.isEnabled ? '启用' : '停用'}
                      </span>
                    </td>
                    <td className="px-8 py-6 text-right">
                      <button onClick={() => handleOpenEditModal(order)} className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-white rounded-lg transition-all">
                        <Edit3 size={16}/>
                      </button>
                      <button onClick={() => handleDelete(order.id)} className="p-2 text-slate-400 hover:text-red-500 hover:bg-white rounded-lg transition-all">
                        <Trash2 size={16}/>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* 医嘱模版弹窗 */}
      {showModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[48px] shadow-2xl w-full max-w-[1500px] max-h-[95vh] overflow-hidden flex flex-col animate-scale-in">
            <div className="px-10 py-6 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className={`w-12 h-12 ${isEditing ? 'bg-indigo-500' : 'bg-indigo-600'} rounded-2xl flex items-center justify-center text-white shadow-lg shadow-indigo-200`}>
                  {isEditing ? <Edit3 size={24} strokeWidth={3} /> : <Plus size={24} strokeWidth={3} />}
                </div>
                <div>
                  <h3 className="text-xl font-black text-slate-900 tracking-tight">{isEditing ? '编辑医嘱模版' : '创建新医嘱模版'}</h3>
                  <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">Define New Medical Order Set</p>
                </div>
              </div>
              <button onClick={() => setShowModal(false)} className="p-3 hover:bg-slate-100 rounded-2xl text-slate-400 hover:text-slate-900 transition-all"><X size={24} /></button>
            </div>
            <div className="flex-1 overflow-y-auto p-8 space-y-8 no-scrollbar bg-white">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8 bg-slate-50/50 p-6 rounded-3xl border border-slate-100">
                <InputField label="医嘱组名称" placeholder="如：常规开机基础医嘱组" required value={orderGroupForm.name} onChange={(e) => setOrderGroupForm({...orderGroupForm, name: e.target.value})}/>
                <SelectField label="业务分类" options={dictOptions[DICT_TYPES.ORDER_CATEGORY] || []} value={orderGroupForm.type} onChange={(e) => setOrderGroupForm({...orderGroupForm, type: e.target.value})}/>
              </div>
              <div className="space-y-4">
                <div className="flex justify-between items-center px-2">
                  <h4 className="text-sm font-black text-indigo-600 uppercase tracking-widest flex items-center gap-2"><Layers size={16}/> 医嘱明细列表</h4>
                  <div className="flex items-center gap-3">
                    <div className="flex bg-slate-100 p-1 rounded-2xl shadow-inner">
                      <button onClick={handleCombineOrders} className="flex items-center gap-2 px-5 py-2 text-indigo-600 hover:bg-white rounded-xl text-[11px] font-black transition-all hover:shadow-sm"><Link size={14}/> 组合选中</button>
                      <button onClick={handleUncombineOrders} className="flex items-center gap-2 px-5 py-2 text-slate-500 hover:bg-white rounded-xl text-[11px] font-black transition-all hover:shadow-sm"><Unlink size={14}/> 取消组合</button>
                    </div>
                    <button onClick={handleAddOrderEntry} className="flex items-center gap-1.5 text-[11px] font-black text-indigo-600 hover:bg-indigo-50 px-4 py-2 rounded-xl transition-all border border-indigo-100 shadow-sm"><Plus size={14}/> 添加医嘱项</button>
                  </div>
                </div>
                {/* 表格 */}
                <div className="border border-slate-100 rounded-[32px] shadow-sm overflow-x-auto">
                  <table className="w-full table-fixed text-left text-[11px] border-collapse min-w-[1200px]">
                    <colgroup>
                      <col style={{ width: 48 }} />
                      <col style={{ width: 256 }} />
                      <col style={{ width: 128 }} />
                      <col style={{ width: 128 }} />
                      <col style={{ width: 112 }} />
                      <col style={{ width: 64 }} />
                      <col style={{ width: 128 }} />
                      <col style={{ width: 128 }} />
                      <col style={{ width: 128 }} />
                      <col />
                    </colgroup>
                    <thead className="bg-slate-50 text-slate-400 font-black uppercase tracking-tighter">
                      <tr>
                        <th className="px-4 py-4 text-center">
                          <input type="checkbox" onChange={handleToggleSelectAllOrders} className="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500" />
                        </th>
                        <th className="px-4 py-4">药品/方法</th>
                        <th className="px-4 py-4 text-center">规格</th>
                        <th className="px-4 py-4 text-center">最小单位剂量</th>
                        <th className="px-4 py-4 text-center">剂量</th>
                        <th className="px-4 py-4">单位</th>
                        <th className="px-4 py-4">用法</th>
                        <th className="px-4 py-4">频次</th>
                        <th className="px-4 py-4">使用时机</th>
                        <th className="px-4 py-4 text-right">操作</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-50 font-black text-slate-700">
                      {modalOrders.length === 0 ? (
                        <tr>
                          <td colSpan={10} className="px-6 py-12 text-center text-slate-400">
                            <p className="text-sm">暂无医嘱</p>
                            <p className="text-xs mt-1">点击"添加医嘱项"添加医嘱</p>
                          </td>
                        </tr>
                      ) : (
                        groupedOrders.map((group, groupIdx) => (
                          group.type === 'group' ? (
                            <Fragment key={group.groupId || groupIdx}>
                              {/* 组头 */}
                              <tr className="bg-indigo-50/30 border-t-2 border-indigo-200/50">
                                <td colSpan={10} className="px-4 py-2">
                                  <div className="flex items-center gap-2">
                                    <input
                                      type="checkbox"
                                      checked={group.items.every(o => o.selected)}
                                      onChange={(e) => {
                                        const checked = e.target.checked
                                        setModalOrders(modalOrders.map(o => o.groupId === group.groupId ? { ...o, selected: checked } : o))
                                      }}
                                      className="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
                                    />
                                    <div className="w-5 h-5 bg-indigo-600 text-white rounded flex items-center justify-center text-[9px] font-black shadow-sm"><Hash size={10}/></div>
                                    <span className="text-[10px] font-black text-indigo-600 uppercase tracking-widest">组合医嘱单元</span>
                                  </div>
                                </td>
                              </tr>
                              {/* 组内行 */}
                              {group.items.map((o, idx) => (
                                <tr key={o.id} className={`bg-indigo-50/10 hover:bg-indigo-50/30 transition-colors ${idx === group.items.length - 1 ? 'border-b-2 border-indigo-200/50' : ''}`}>
                                  {renderOrderCells(o, true)}
                                </tr>
                              ))}
                            </Fragment>
                          ) : (
                            <tr key={group.items[0].id} className="hover:bg-slate-50/50 transition-colors group font-black text-slate-700">
                              {renderOrderCells(group.items[0])}
                            </tr>
                          )
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
            <div className="px-10 py-6 bg-slate-50 border-t border-slate-100 flex justify-end gap-3 shrink-0">
              <button onClick={() => setShowModal(false)} className="px-8 py-2.5 bg-white border border-slate-200 text-slate-600 rounded-2xl text-xs font-black hover:bg-slate-100 transition-all shadow-sm">取消</button>
              <button onClick={handleSave} disabled={isLoading} className="px-12 py-2.5 bg-indigo-600 text-white rounded-2xl text-xs font-black shadow-xl shadow-indigo-100 hover:bg-indigo-700 transition-all disabled:opacity-50">
                {isLoading ? <Loader2 size={16} className="animate-spin mr-2 inline" /> : ''}确认保存
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export const OrderTab = memo(OrderTabComponent)
