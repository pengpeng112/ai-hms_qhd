// MaterialTab - 材料目录 Tab 组件

import { useState, useCallback, useEffect, useMemo, memo } from 'react'
import {
  Search, Plus, Edit3, Trash2, Download,
  Container, ToggleLeft, ToggleRight, X,
  Info, Sliders, ChevronDown, AlertCircle
} from 'lucide-react'
import * as XLSX from 'xlsx'
import {
  materialCatalogApi,
  type MaterialCatalog
} from '@/services/treatmentConfigApi'
import { DICT_TYPES } from '@/services/dictApi'
import { getToken } from '@/utils/token'
import { usePagination } from '../hooks/usePagination'
import { useSelection } from '../hooks/useSelection'
import { useSearch } from '../hooks/useSearch'
import { PaginationBar } from '../components/PaginationBar'
import { Checkbox } from '../components/Checkbox'
import { BatchImportModal } from '../components/BatchImportModal'
import { FormSection } from '../components/FormSection'
import { INITIAL_MATERIAL_FORM, LOAD_ALL_PAGE_SIZE, PAGE_SIZE } from '../constants'

// Excel 列定义
const MATERIAL_EXCEL_COLUMNS = [
  { key: 'code', label: '材料编码', required: false, example: 'M001' },
  { key: 'name', label: '材料名称', required: true, example: '透析器' },
  { key: 'shortName', label: '简称', required: true, example: '' },
  { key: 'mnemonic', label: '助记码', required: false, example: 'TXQ' },
  { key: 'category', label: '材料分类', required: true, example: '透析器' },
  { key: 'spec', label: '规格', required: false, example: '1.5m²' },
  { key: 'standardType', label: '标准分类', required: false, example: '类一' },
  { key: 'brand', label: '品牌', required: false, example: '费森尤斯' },
  { key: 'unit', label: '单位', required: true, example: '个' },
  { key: 'packaging', label: '包装', required: false, example: '24个/箱' },
  { key: 'manufacturer', label: '生产厂家', required: false, example: '费森尤斯医疗' },
  { key: 'sortOrder', label: '排序', required: false, example: '0' },
  { key: 'isEnabled', label: '启用状态', required: false, example: 'true' },
  { key: 'notes', label: '备注', required: false, example: '' }
]

const MATERIAL_REQUIRED_FIELDS = ['category', 'name', 'shortName', 'unit'] as const

type MaterialRequiredField = typeof MATERIAL_REQUIRED_FIELDS[number]

type MaterialValidationInput = {
  category?: string
  name?: string
  shortName?: string
  unit?: string
  code?: string
}

const MATERIAL_REQUIRED_LABELS: Record<MaterialRequiredField, string> = {
  category: '材料分类',
  name: '材料名称',
  shortName: '简称',
  unit: '单位'
}

const isBlank = (value?: string) => !value || value.trim() === ''

const getMissingMaterialFields = (input: MaterialValidationInput): MaterialRequiredField[] => {
  const missing: MaterialRequiredField[] = []
  for (const field of MATERIAL_REQUIRED_FIELDS) {
    if (isBlank(input[field])) {
      missing.push(field)
    }
  }
  return missing
}

const formatMissingMaterialFields = (missing: MaterialRequiredField[]) =>
  missing.map((field) => MATERIAL_REQUIRED_LABELS[field]).join('、')

// 输入字段组件类型定义
interface InputFieldProps {
  label: string
  placeholder?: string
  suffix?: string
  required?: boolean
  type?: 'text' | 'number'
  value: string | number
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  readOnly?: boolean
}

interface SelectFieldProps {
  label: string
  required?: boolean
  options: string[] | Array<{ value: string; label: string }>
  value: string
  onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void
}

// 输入字段组件
const InputField = ({ label, placeholder, suffix, required, type = 'text', value, onChange, readOnly }: InputFieldProps) => (
  <div className="space-y-1">
    <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
      {required && <span className="text-red-500 mr-1">*</span>}{label}
    </label>
    <div className="relative group">
      <input
        type={type}
        value={value}
        onChange={onChange}
        readOnly={readOnly}
        placeholder={placeholder}
        className={`w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all placeholder:text-slate-300 placeholder:font-medium ${readOnly ? 'bg-slate-50 cursor-not-allowed' : ''}`}
      />
      {suffix && (
        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-slate-300 uppercase">
          {suffix}
        </span>
      )}
    </div>
  </div>
)

// 下拉字段组件
const SelectField = ({ label, options, required, value, onChange }: SelectFieldProps) => (
  <div className="space-y-1">
    <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
      {required && <span className="text-red-500 mr-1">*</span>}{label}
    </label>
    <div className="relative">
      <select
        value={value}
        onChange={onChange}
        className="w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none appearance-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all cursor-pointer"
      >
        <option value="">请选择</option>
        {options.map((opt, i) => {
          const optValue = typeof opt === 'string' ? opt : opt.value
          const optLabel = typeof opt === 'string' ? opt : opt.label
          return <option key={optValue || i} value={optValue}>{optLabel}</option>
        })}
      </select>
      <ChevronDown size={12} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
    </div>
  </div>
)

interface MaterialTabProps {
  dictOptions: Record<string, Array<{ value: string; label: string }>>
  onRefreshDict: () => Promise<void>
}

function MaterialTabComponent({ dictOptions, onRefreshDict }: MaterialTabProps) {
  // 构建分类查找 Map (js-index-maps, js-cache-function-results)
  const categoryMap = useMemo(() => {
    const map = new Map<string, string>()
    const options = dictOptions[DICT_TYPES.MATERIAL_CATEGORY] || []
    for (const opt of options) {
      map.set(opt.value, opt.label)
    }
    return map
  }, [dictOptions])
  const categoryLabelToValueMap = useMemo(() => {
    const map = new Map<string, string>()
    const options = dictOptions[DICT_TYPES.MATERIAL_CATEGORY] || []
    for (const opt of options) {
      map.set(opt.label, opt.value)
    }
    return map
  }, [dictOptions])
  const categoryValueToLabelLowerMap = useMemo(() => {
    const map = new Map<string, string>()
    const options = dictOptions[DICT_TYPES.MATERIAL_CATEGORY] || []
    for (const opt of options) {
      map.set(opt.value.trim().toLowerCase(), opt.label.trim().toLowerCase())
      map.set(opt.label.trim().toLowerCase(), opt.label.trim().toLowerCase())
    }
    return map
  }, [dictOptions])

  // 检查用户是否已登录 (js-cache-storage)
  const checkAuthStatus = useCallback(() => {
    const token = getToken()
    if (!token) {
      return false
    }
    // 检查 token 是否过期
    const expiry = localStorage.getItem('hdis_token_expiry')
    if (expiry && new Date(expiry) < new Date()) {
      localStorage.removeItem('hdis_access_token')
      localStorage.removeItem('hdis_user_info')
      localStorage.removeItem('hdis_token_expiry')
      return false
    }
    return true
  }, [])

  // 获取分类显示名称的辅助函数 (js-index-maps)
  const getCategoryLabel = useCallback((categoryValue: string) => {
    if (!categoryValue) return '-'
    return categoryMap.get(categoryValue) || categoryValue
  }, [categoryMap])
  const getCategorySelectValue = useCallback((storedCategory?: string) => {
    if (!storedCategory) return ''
    if (categoryMap.has(storedCategory)) return storedCategory
    return categoryLabelToValueMap.get(storedCategory) || storedCategory
  }, [categoryMap, categoryLabelToValueMap])

  // 数据状态
  const [materialCatalog, setMaterialCatalog] = useState<MaterialCatalog[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [showModal, setShowModal] = useState(false)
  const [showBatchImportModal, setShowBatchImportModal] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [currentMaterial, setCurrentMaterial] = useState<Partial<MaterialCatalog>>(INITIAL_MATERIAL_FORM)

  // 搜索、分页、选择
  const search = useSearch({
    data: materialCatalog,
    searchFields: ['name', 'code', 'category', 'brand', 'spec'] as const,
    statusField: 'isEnabled'
  })
  const pagination = usePagination({ data: search.filteredData, pageSize: PAGE_SIZE })
  const selection = useSelection({ data: pagination.currentPageData, keyField: 'id' })

  // 加载数据
  const loadData = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await materialCatalogApi.list({ pageSize: LOAD_ALL_PAGE_SIZE })
      const sorted = [...response.items].sort((a, b) => a.id - b.id)
      setMaterialCatalog(sorted)
    } catch (error) {
      console.error('加载材料目录失败:', error)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 初始加载
  useEffect(() => {
    loadData()
  }, [loadData])

  // 打开新增模态框
  const handleOpenAddModal = useCallback(async () => {
    await onRefreshDict()
    setIsEditing(false)
    setCurrentMaterial(INITIAL_MATERIAL_FORM)
    setShowModal(true)
  }, [onRefreshDict])

  // 打开编辑模态框
  const handleOpenEditModal = useCallback((item: MaterialCatalog) => {
    setIsEditing(true)
    setCurrentMaterial({
      ...item,
      category: getCategorySelectValue(item.category)
    })
    setShowModal(true)
  }, [getCategorySelectValue])

  // 保存
  const handleSave = useCallback(async () => {
    const missingFields = getMissingMaterialFields(currentMaterial)
    if (missingFields.length > 0) {
      alert(`缺少必填字段：${formatMissingMaterialFields(missingFields)}`)
      return
    }

    const normalizedName = (currentMaterial.name ?? '').trim()
    const normalizedShortName = (currentMaterial.shortName ?? '').trim()
    const normalizedCategory = (currentMaterial.category ?? '').trim()
    const normalizedUnit = (currentMaterial.unit ?? '').trim()
    const normalizedCode = currentMaterial.code?.trim() || undefined
    const normalizedMnemonic = currentMaterial.mnemonic?.trim() || undefined
    const normalizedSpec = currentMaterial.spec?.trim() || undefined
    const normalizedStandardType = currentMaterial.standardType?.trim() || undefined
    const normalizedBrand = currentMaterial.brand?.trim() || undefined
    const normalizedPackaging = currentMaterial.packaging?.trim() || undefined
    const normalizedManufacturer = currentMaterial.manufacturer?.trim() || undefined
    const normalizedNotes = currentMaterial.notes?.trim() || undefined

    setIsLoading(true)
    try {
      const materialData = {
        name: normalizedName,
        shortName: normalizedShortName,
        mnemonic: normalizedMnemonic,
        category: normalizedCategory,
        code: normalizedCode,
        spec: normalizedSpec,
        standardType: normalizedStandardType,
        brand: normalizedBrand,
        unit: normalizedUnit,
        packaging: normalizedPackaging,
        manufacturer: normalizedManufacturer,
        sortOrder: currentMaterial.sortOrder ?? 0,
        isEnabled: currentMaterial.isEnabled ?? true,
        notes: normalizedNotes
      }

      if (isEditing && currentMaterial.id) {
        await materialCatalogApi.update(currentMaterial.id, materialData)
      } else {
        await materialCatalogApi.create(materialData)
      }
      await loadData()
      setShowModal(false)
    } catch (error) {
      console.error('保存材料失败:', error)
      alert('保存失败，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }, [currentMaterial, isEditing, loadData])

  // 删除
  const handleDelete = useCallback(async (id: number) => {
    if (window.confirm('确定要删除该材料条目吗？')) {
      setIsLoading(true)
      try {
        await materialCatalogApi.delete(id)
        await loadData()
      } catch (error) {
        console.error('删除材料失败:', error)
        alert('删除失败，请稍后重试')
      } finally {
        setIsLoading(false)
      }
    }
  }, [loadData])

  // 切换状态
  const handleToggleStatus = useCallback(async (id: number) => {
    setIsLoading(true)
    try {
      await materialCatalogApi.toggleEnabled(id)
      await loadData()
    } catch (error) {
      console.error('切换材料状态失败:', error)
      alert('操作失败，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }, [loadData])

  // 批量删除
  const handleBatchDelete = useCallback(async () => {
    if (selection.selectedIds.size === 0) {
      alert('请先选择要删除的项目')
      return
    }

    if (!confirm(`确定要删除选中的 ${selection.selectedIds.size} 项吗？`)) {
      return
    }

    setIsLoading(true)
    try {
      await Promise.all(
        Array.from(selection.selectedIds).map(id => materialCatalogApi.delete(Number(id)))
      )
      await loadData()
      selection.clearSelection()
      alert('删除成功')
    } catch (error) {
      console.error('批量删除失败:', error)
      alert('删除失败')
    } finally {
      setIsLoading(false)
    }
  }, [selection, loadData])

  // 导出材料数据到 Excel
  const handleExport = useCallback(async () => {
    try {
      // 获取所有材料数据
      const response = await materialCatalogApi.list({ pageSize: LOAD_ALL_PAGE_SIZE })
      
      // 按 ID 升序排序
      const sortedData = [...response.items].sort((a, b) => a.id - b.id)
      
      // 转换为 Excel 格式
      const excelData = sortedData.map(item => ({
        '材料编码': item.code,
        '材料名称': item.name,
        '简称': item.shortName || '',
        '助记码': item.mnemonic || '',
        '材料分类': item.category,
        '规格': item.spec || '',
        '标准分类': item.standardType || '',
        '品牌': item.brand || '',
        '单位': item.unit || '',
        '包装': item.packaging || '',
        '生产厂家': item.manufacturer || '',
        '排序': item.sortOrder,
        '启用状态': item.isEnabled ? '是' : '否',
        '备注': item.notes || ''
      }))

      // 创建工作簿
      const ws = XLSX.utils.json_to_sheet(excelData)
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, '材料目录')

      // 设置列宽
      ws['!cols'] = [
        { wch: 12 }, // 材料编码
        { wch: 20 }, // 材料名称
        { wch: 12 }, // 简称
        { wch: 10 }, // 助记码
        { wch: 15 }, // 材料分类
        { wch: 12 }, // 规格
        { wch: 10 }, // 标准分类
        { wch: 15 }, // 品牌
        { wch: 8 },  // 单位
        { wch: 12 }, // 包装
        { wch: 20 }, // 生产厂家
        { wch: 6 },  // 排序
        { wch: 8 },  // 启用状态
        { wch: 30 }  // 备注
      ]

      // 下载文件
      XLSX.writeFile(wb, `材料目录_${new Date().toLocaleDateString('zh-CN')}.xlsx`)
    } catch (error) {
      console.error('导出失败:', error)
      alert('导出失败，请稍后重试')
    }
  }, [])

  // 批量导入 - 返回结果而非使用 alert
  // 注意：autoAddCategories 参数仅用于 DrugTab，MaterialTab 不需要
  const handleBatchImport = useCallback(async (
    data: unknown[],
    _autoAddCategories?: boolean,
    onProgress?: (percent: number) => void
  ) => {
    // _autoAddCategories 参数用于匹配 BatchImportModal 的签名，MaterialTab 不需要此功能
    void _autoAddCategories // 显式标记为未使用
    // 检查登录状态
    if (!checkAuthStatus()) {
      return {
        success: 0,
        failed: data.length,
        errors: ['用户未登录或登录已过期，请先登录后再试']
      }
    }

    let successCount = 0
    let failCount = 0
    let skippedCount = 0
    const errors: string[] = []
    const normalizeText = (value: unknown) => String(value ?? '').trim()
    const normalizeCodeKey = (value: unknown) => normalizeText(value).toLowerCase()
    const normalizeCategoryKey = (value: unknown) => {
      const raw = normalizeText(value).toLowerCase()
      if (!raw) return ''
      return categoryValueToLabelLowerMap.get(raw) || raw
    }
    const buildSemanticKey = (item: {
      category?: string
      name?: string
      spec?: string
      brand?: string
      unit?: string
    }) => {
      return [
        normalizeCategoryKey(item.category),
        normalizeText(item.name).toLowerCase(),
        normalizeText(item.spec).toLowerCase(),
        normalizeText(item.brand).toLowerCase(),
        normalizeText(item.unit).toLowerCase()
      ].join('|')
    }

    const existingData = await materialCatalogApi.list({ pageSize: LOAD_ALL_PAGE_SIZE })
    const existingCodeSet = new Set<string>()
    const existingSemanticMap = new Map<string, { code: string; name: string }>()
    for (const existing of existingData.items) {
      const codeKey = normalizeCodeKey(existing.code)
      if (codeKey) {
        existingCodeSet.add(codeKey)
      }
      const semanticKey = buildSemanticKey(existing)
      if (semanticKey && !existingSemanticMap.has(semanticKey)) {
        existingSemanticMap.set(semanticKey, { code: existing.code, name: existing.name })
      }
    }

    const importedCodeSet = new Set<string>()
    const importedSemanticSet = new Set<string>()

    for (let i = 0; i < data.length; i++) {
      try {
        const item = data[i] as {
          code: string
          name: string
          shortName?: string
          mnemonic?: string
          category: string
          spec?: string
          standardType?: string
          brand?: string
          unit?: string
          packaging?: string
          manufacturer?: string
          sortOrder?: number
          isEnabled?: boolean | string
          notes?: string
        }
        const missingFields = getMissingMaterialFields(item)
        if (missingFields.length > 0) {
          errors.push(`第 ${i + 1} 行：缺少必填字段（${formatMissingMaterialFields(missingFields)}）`)
          failCount++
          onProgress?.(Math.round(((i + 1) / data.length) * 100))
          continue
        }

        const codeKey = normalizeCodeKey(item.code)
        if (codeKey && (existingCodeSet.has(codeKey) || importedCodeSet.has(codeKey))) {
          errors.push(`第 ${i + 1} 行：材料编码 "${normalizeText(item.code)}" 已存在，已跳过`)
          skippedCount++
          onProgress?.(Math.round(((i + 1) / data.length) * 100))
          continue
        }

        const semanticKey = buildSemanticKey(item)
        if (existingSemanticMap.has(semanticKey)) {
          const duplicated = existingSemanticMap.get(semanticKey)
          errors.push(`第 ${i + 1} 行：与现有材料重复（编码 ${duplicated?.code}，名称 ${duplicated?.name}），已跳过`)
          skippedCount++
          onProgress?.(Math.round(((i + 1) / data.length) * 100))
          continue
        }
        if (importedSemanticSet.has(semanticKey)) {
          errors.push(`第 ${i + 1} 行：与本次导入中的其他行重复，已跳过`)
          skippedCount++
          onProgress?.(Math.round(((i + 1) / data.length) * 100))
          continue
        }

        // 将 isEnabled 转换为布尔值
        let isEnabled = true
        if (item.isEnabled !== undefined && item.isEnabled !== null && item.isEnabled !== '') {
          if (typeof item.isEnabled === 'boolean') {
            isEnabled = item.isEnabled
          } else if (typeof item.isEnabled === 'string') {
            // 支持 "是"/"否"、"true"/"false"、1/0 等格式
            const str = item.isEnabled.toString().trim().toLowerCase()
            isEnabled = str === '是' || str === 'true' || str === '1' || str === 'yes'
          } else if (typeof item.isEnabled === 'number') {
            isEnabled = item.isEnabled === 1
          }
        }

        await materialCatalogApi.create({
          code: item.code?.trim() || undefined,
          name: item.name?.trim() ?? '',
          shortName: item.shortName?.trim() ?? '',
          mnemonic: item.mnemonic?.trim() || undefined,
          category: item.category?.trim() ?? '',
          spec: item.spec?.trim() || undefined,
          standardType: item.standardType?.trim() || undefined,
          brand: item.brand?.trim() || undefined,
          unit: item.unit?.trim() ?? '',
          packaging: item.packaging?.trim() || undefined,
          manufacturer: item.manufacturer?.trim() || undefined,
          sortOrder: item.sortOrder ?? 0,
          isEnabled,
          notes: item.notes?.trim() || undefined
        })
        if (codeKey) {
          existingCodeSet.add(codeKey)
          importedCodeSet.add(codeKey)
        }
        importedSemanticSet.add(semanticKey)
        successCount++
        onProgress?.(Math.round(((i + 1) / data.length) * 100))
      } catch (err) {
        failCount++
        errors.push(`第 ${i + 1} 行：${err instanceof Error ? err.message : '导入失败'}`)
        onProgress?.(Math.round(((i + 1) / data.length) * 100))
      }
    }
    onProgress?.(100)

    // 先刷新数据
    await loadData()
    // 导入成功后自动关闭模态框
    if (failCount === 0 && skippedCount === 0) {
      setShowBatchImportModal(false)
    }
    return { success: successCount, failed: failCount, skipped: skippedCount, errors }
  }, [loadData, checkAuthStatus, categoryValueToLabelLowerMap])

  return (
    <div className="flex-1 flex flex-col animate-fade-in">
      {/* 工具栏 */}
      <div className="bg-white p-4 border-b border-slate-100 flex justify-between items-center shrink-0">
        <div className="flex items-center gap-4">
          {/* 搜索框 */}
          <div className="relative w-72">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              placeholder="搜索材料名称、编码、分类、品牌..."
              value={search.keyword}
              onChange={(e) => search.setKeyword(e.target.value)}
              className="w-full pl-10 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all"
            />
            {search.keyword && (
              <button
                onClick={search.clearSearch}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
              >
                <X size={14} />
              </button>
            )}
          </div>
          {/* 状态过滤 */}
          <select
            value={search.statusFilter}
            onChange={(e) => search.setStatusFilter(e.target.value as 'all' | 'enabled' | 'disabled')}
            className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-600 focus:ring-2 focus:ring-blue-500 outline-none transition-all cursor-pointer"
          >
            <option value="all">全部状态</option>
            <option value="enabled">已启用</option>
            <option value="disabled">已停用</option>
          </select>
          {search.hasFilter && (
            <span className="text-xs text-slate-500">
              找到 <span className="font-bold text-orange-600">{search.resultCount}</span> 条结果
            </span>
          )}
          {selection.selectedCount > 0 && (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-blue-50 text-blue-700 rounded-lg text-xs font-bold">
              <span>已选 {selection.selectedCount} 项</span>
              <button onClick={() => selection.clearSelection()} className="ml-1 text-blue-600 hover:text-blue-800">取消</button>
            </div>
          )}
        </div>
        <div className="flex gap-2">
          {selection.selectedCount > 0 && (
            <button onClick={handleBatchDelete} className="flex items-center gap-2 px-5 py-2 bg-red-600 text-white rounded-xl text-xs font-black hover:bg-red-700 shadow-lg transition-all">
              <Trash2 size={14} /> 批量删除
            </button>
          )}
          <button onClick={() => setShowBatchImportModal(true)} className="flex items-center gap-2 px-5 py-2 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-50 transition-all">
            <Download size={14} /> 批量导入
          </button>
          <button onClick={handleOpenAddModal} className="flex items-center gap-2 px-5 py-2 bg-orange-600 shadow-orange-100 text-white rounded-xl text-xs font-black hover:opacity-90 shadow-lg transition-all">
            <Plus size={14} /> 新增材料
          </button>
        </div>
      </div>

      {/* 表格 */}
      <div className="flex-1 overflow-auto p-6">
        <div className="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden overflow-x-auto">
          <table className="w-full text-left text-sm min-w-[1300px]">
            <thead className="bg-slate-50 text-[10px] font-black uppercase text-slate-400">
              <tr>
                <th className="px-6 py-5 w-12 text-center">
                  <Checkbox
                    checked={selection.allSelected}
                    indeterminate={selection.someSelected && !selection.allSelected}
                    onChange={selection.toggleSelectAll}
                  />
                </th>
                <th className="px-6 py-5 w-16 text-center">序号</th>
                <th className="px-6 py-5 w-32 whitespace-nowrap">材料分类</th>
                <th className="px-6 py-5 min-w-[200px]">材料名称</th>
                <th className="px-6 py-5 w-24 whitespace-nowrap">简称</th>
                <th className="px-6 py-5 w-28 whitespace-nowrap">品牌</th>
                <th className="px-6 py-5 w-28 whitespace-nowrap">规格</th>
                <th className="px-6 py-5 w-20 text-center whitespace-nowrap">单位</th>
                <th className="px-6 py-5 w-28 text-center whitespace-nowrap">禁用状态</th>
                <th className="px-6 py-5 w-28 text-right whitespace-nowrap">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50 font-bold text-slate-700">
              {pagination.currentPageData.map((item: MaterialCatalog, index: number) => (
                <tr
                  key={item.id}
                  className={`hover:bg-slate-50 transition-colors ${!item.isEnabled ? 'opacity-50 grayscale' : ''} ${selection.selectedIds.has(String(item.id)) ? 'bg-blue-50/50' : ''}`}
                >
                  <td className="px-6 py-6 text-center">
                    <Checkbox
                      checked={selection.selectedIds.has(String(item.id))}
                      onChange={() => selection.toggleSelection(item.id)}
                    />
                  </td>
                  <td className="px-6 py-6 text-center text-slate-400 font-mono text-xs">{pagination.displayInfo.startIndex + index}</td>
                  <td className="px-6 py-6 text-slate-500 whitespace-nowrap">{getCategoryLabel(item.category)}</td>
                  <td className="px-6 py-6 flex items-center gap-3">
                    <div className={`p-2 rounded-xl ${!item.isEnabled ? 'bg-slate-100 text-slate-400' : 'bg-orange-50 text-orange-600'}`}>
                      <Container size={18} />
                    </div>
                    <span className="text-slate-800">{item.name}</span>
                  </td>
                  <td className="px-6 py-6 text-slate-600 text-xs whitespace-nowrap">{item.shortName || '-'}</td>
                  <td className="px-6 py-6 text-slate-600 whitespace-nowrap">{item.brand || '-'}</td>
                  <td className="px-6 py-6 text-slate-500 text-xs whitespace-nowrap">{item.spec || '-'}</td>
                  <td className="px-6 py-6 text-center whitespace-nowrap">{item.unit || '-'}</td>
                  <td className="px-6 py-6 text-center whitespace-nowrap">
                    <button
                      onClick={() => handleToggleStatus(item.id)}
                      className={`p-1 rounded-lg transition-all ${!item.isEnabled ? 'text-red-500 bg-red-50' : 'text-emerald-500 bg-emerald-50'}`}
                    >
                      {item.isEnabled ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                    </button>
                  </td>
                  <td className="px-6 py-6 text-right space-x-1 whitespace-nowrap">
                    <button onClick={() => handleOpenEditModal(item)} className="p-2 text-slate-400 hover:text-blue-600 hover:bg-white rounded-lg transition-all shadow-sm">
                      <Edit3 size={16} />
                    </button>
                    <button onClick={() => handleDelete(item.id)} className="p-2 text-slate-400 hover:text-red-500 hover:bg-white rounded-lg transition-all shadow-sm">
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <PaginationBar
            currentPage={pagination.currentPage}
            totalPages={pagination.totalPages}
            displayInfo={pagination.displayInfo}
            goToPage={pagination.goToPage}
            firstPage={pagination.firstPage}
            lastPage={pagination.lastPage}
            prevPage={pagination.prevPage}
            nextPage={pagination.nextPage}
          />
        </div>
      </div>

      {/* 编辑/新增模态框 */}
      {showModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-4xl overflow-hidden flex flex-col animate-scale-in">
            {/* 标题区域 */}
            <div className="px-8 py-4 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className={`w-10 h-10 ${isEditing ? 'bg-blue-500' : 'bg-orange-600'} rounded-xl flex items-center justify-center text-white shadow-lg ${isEditing ? 'shadow-blue-100' : 'shadow-orange-100'}`}>
                  {isEditing ? <Edit3 size={20} strokeWidth={3} /> : <Plus size={20} strokeWidth={3} />}
                </div>
                <div>
                  <h3 className="text-lg font-black text-slate-900 tracking-tight">{isEditing ? '编辑材料档案' : '新增材料档案'}</h3>
                  <p className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">Material Catalog Management</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={() => setShowModal(false)}
                  className="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-all"
                >
                  <X size={20} />
                </button>
              </div>
            </div>

            {/* 表单内容 */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 no-scrollbar bg-white">
              {/* 核心信息 */}
              <FormSection title="核心信息" icon={Info}>
                <SelectField
                  label="材料分类"
                  required
                  options={dictOptions[DICT_TYPES.MATERIAL_CATEGORY] || []}
                  value={currentMaterial.category || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, category: e.target.value })}
                />
                <InputField
                  label="材料名称"
                  required
                  placeholder="完整名称"
                  value={currentMaterial.name || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, name: e.target.value })}
                />
                <InputField
                  label="简称"
                  required
                  placeholder="简称"
                  value={currentMaterial.shortName || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, shortName: e.target.value })}
                />
                <InputField
                  label="助记码"
                  placeholder="拼音简码"
                  value={currentMaterial.mnemonic || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, mnemonic: e.target.value })}
                />
              </FormSection>

              {/* 规格与属性 */}
              <FormSection title="规格与属性" icon={Sliders}>
                <InputField
                  label="材料编码"
                  placeholder="内部编码（可选）"
                  value={currentMaterial.code || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, code: e.target.value })}
                />
                <SelectField
                  label="标准分类"
                  options={['类一', '类二', '类三', '非医疗']}
                  value={currentMaterial.standardType || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, standardType: e.target.value })}
                />
                <InputField
                  label="单位"
                  required
                  placeholder="支/个/套"
                  value={currentMaterial.unit || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, unit: e.target.value })}
                />
                <InputField
                  label="规格"
                  placeholder="如: FX60"
                  value={currentMaterial.spec || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, spec: e.target.value })}
                />
                <InputField
                  label="品牌"
                  placeholder="品牌名称"
                  value={currentMaterial.brand || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, brand: e.target.value })}
                />
                <InputField
                  label="包装"
                  placeholder="如: 24个/箱"
                  value={currentMaterial.packaging || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, packaging: e.target.value })}
                />
                <InputField
                  label="生产厂家"
                  placeholder="厂家全称"
                  value={currentMaterial.manufacturer || ''}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, manufacturer: e.target.value })}
                />
                <InputField
                  label="材料排序"
                  type="number"
                  placeholder="显示顺序"
                  value={currentMaterial.sortOrder ?? 0}
                  onChange={(e) => setCurrentMaterial({ ...currentMaterial, sortOrder: parseInt(e.target.value) || 0 })}
                />
              </FormSection>

              {/* 备注与状态 - 独立区域，不用 FormSection */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 p-5 bg-slate-50/50 rounded-2xl border border-slate-100">
                <div className="md:col-span-2 space-y-1">
                  <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">备注说明</label>
                  <textarea
                    value={currentMaterial.notes || ''}
                    onChange={(e) => setCurrentMaterial({ ...currentMaterial, notes: e.target.value })}
                    className="w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all resize-none h-20"
                    placeholder="请输入补充说明..."
                  />
                </div>
                <div className="flex flex-col justify-center gap-3">
                  <p className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">状态控制</p>
                  <div className="flex items-center gap-4 bg-white p-3 rounded-xl border border-slate-100 shadow-sm">
                    <span className="text-xs font-bold text-slate-600 flex-1">是否禁用材料</span>
                    <button
                      onClick={() => setCurrentMaterial({ ...currentMaterial, isEnabled: !(currentMaterial.isEnabled ?? true) })}
                      className={`p-0.5 rounded-lg transition-all ${!(currentMaterial.isEnabled ?? true) ? 'text-red-500 bg-red-50' : 'text-emerald-500 bg-emerald-50'}`}
                    >
                      {!(currentMaterial.isEnabled ?? true) ? <ToggleLeft size={28} /> : <ToggleRight size={28} />}
                    </button>
                  </div>
                  <div className="flex items-start gap-1.5 text-[9px] text-slate-400 leading-relaxed px-1">
                    <AlertCircle size={10} className="shrink-0 mt-0.5" />
                    <span>禁用后，该材料在排班及处方配置中将不可被检索。</span>
                  </div>
                </div>
              </div>
            </div>

            {/* 底部按钮 */}
            <div className="px-8 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3 shrink-0">
              <button
                onClick={() => setShowModal(false)}
                className="px-6 py-2 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-100 transition-all"
              >
                取消
              </button>
              <button
                onClick={handleSave}
                disabled={isLoading}
                className="px-10 py-2 bg-blue-600 text-white rounded-xl text-xs font-black shadow-xl shadow-blue-100 hover:bg-blue-700 transition-all disabled:opacity-50"
              >
                保存更改
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 批量导入模态框 - Excel 版本 */}
      {showBatchImportModal && (
        <BatchImportModal
          title="批量导入材料"
          columns={MATERIAL_EXCEL_COLUMNS}
          onImport={handleBatchImport}
          onExport={handleExport}
          onClose={() => setShowBatchImportModal(false)}
          maxSize={5}
        />
      )}
    </div>
  )
}

export const MaterialTab = memo(MaterialTabComponent)
