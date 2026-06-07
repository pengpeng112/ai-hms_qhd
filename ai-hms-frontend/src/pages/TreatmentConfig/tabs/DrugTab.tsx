// DrugTab - 药品目录 Tab 组件

import { useState, useCallback, useEffect, useMemo, memo } from 'react'
import {
  Search, Plus, Edit3, Trash2, Download, X,
  Pill, ToggleLeft, ToggleRight, Info, Sliders, Activity, AlertCircle
} from 'lucide-react'
import * as XLSX from 'xlsx'
import {
  drugCatalogApi,
  type DrugCatalog
} from '@/services/treatmentConfigApi'
import { dictApi, DICT_TYPES } from '@/services/dictApi'
import { getToken } from '@/utils/token'
import { usePagination } from '../hooks/usePagination'
import { useSelection } from '../hooks/useSelection'
import { useSearch } from '../hooks/useSearch'
import { PaginationBar } from '../components/PaginationBar'
import { Checkbox } from '../components/Checkbox'
import { BatchImportModal } from '../components/BatchImportModal'
import { FormSection } from '../components/FormSection'
import { InputField } from '../components/InputField'
import { SelectField } from '../components/SelectField'
import { INITIAL_DRUG_FORM, LOAD_ALL_PAGE_SIZE, PAGE_SIZE } from '../constants'

// 字典常量 - 从 TemplateCenter 复制（标准分类和使用时机）
const DICT_STD_TYPES = ['类一', '类二', '类三', '处方药', '精神类', '抢救类']
const DICT_TIMINGS = ['开机', '治疗中', '下机', '透析前', '透析后']

// "方法"类别的 dict code
const METHOD_CATEGORY_CODE = 'METHOD'

// Excel 列定义
const DRUG_EXCEL_COLUMNS = [
  { key: 'code', label: '药品编码', required: false, example: 'D001' },
  { key: 'name', label: '药品名称', required: true, example: '肝素钠注射液' },
  { key: 'genericName', label: '通用名', required: false, example: '肝素钠' },
  { key: 'category', label: '药品分类', required: true, example: '抗凝药' },
  { key: 'spec', label: '规格', required: false, example: '2ml:12500IU' },
  { key: 'concentration', label: '浓度', required: false, example: '1250IU/ml' },
  { key: 'unit', label: '单位', required: false, example: '支' },
  { key: 'manufacturer', label: '生产厂家', required: false, example: '常州千红' },
  { key: 'isEnabled', label: '启用状态', required: false, example: 'true' },
  { key: 'notes', label: '备注', required: false, example: '' }
]

interface DrugTabProps {
  dictOptions: Record<string, Array<{ value: string; label: string }>>
  onRefreshDict: () => Promise<void>
}

function DrugTabComponent({ dictOptions, onRefreshDict }: DrugTabProps) {
  // 构建分类查找 Map
  const categoryMap = useMemo(() => {
    const map = new Map<string, string>()
    const options = dictOptions[DICT_TYPES.DRUG_CATEGORY] || []
    for (const opt of options) {
      map.set(opt.value, opt.label)
    }
    return map
  }, [dictOptions])

  // 判断某条目是否为"方法"类别
  const isMethodItem = useCallback((item: DrugCatalog) => {
    return item.category === METHOD_CATEGORY_CODE
  }, [])

  // 获取分类显示名称的辅助函数
  const getCategoryLabel = useCallback((categoryValue: string) => {
    if (!categoryValue) return '-'
    return categoryMap.get(categoryValue) || categoryValue
  }, [categoryMap])

  // 数据状态
  const [drugCatalog, setDrugCatalog] = useState<DrugCatalog[]>([])
  const [showModal, setShowModal] = useState(false)
  const [showBatchImportModal, setShowBatchImportModal] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [currentDrug, setCurrentDrug] = useState<Partial<DrugCatalog>>(INITIAL_DRUG_FORM)
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all')

  // 判断当前编辑的条目是否为"方法"类别
  const isMethodCategory = currentDrug.category === METHOD_CATEGORY_CODE

  // 为搜索添加类别显示名，让用户可以按中文名搜索类别
  const searchData = useMemo(() =>
    drugCatalog.map(item => ({ ...item, categoryLabel: getCategoryLabel(item.category) })),
    [drugCatalog, getCategoryLabel]
  )

  // 搜索、分页、选择
  const search = useSearch({
    data: searchData,
    searchFields: ['name', 'code', 'categoryLabel', 'genericName', 'manufacturer'] as const
  })

  // 状态过滤
  const filteredByStatus = useMemo(() => {
    if (statusFilter === 'all') return search.filteredData
    return search.filteredData.filter(item => {
      return statusFilter === 'enabled' ? item.isEnabled : !item.isEnabled
    })
  }, [search.filteredData, statusFilter])

  const pagination = usePagination({ data: filteredByStatus, pageSize: PAGE_SIZE })
  const selection = useSelection({ data: pagination.currentPageData, keyField: 'id' })

  // 加载数据
  const loadData = useCallback(async () => {
    try {
      const response = await drugCatalogApi.list({ pageSize: LOAD_ALL_PAGE_SIZE })
      const sorted = [...response.items].sort((a, b) => {
        const aOrder = a.sortOrder ?? 0
        const bOrder = b.sortOrder ?? 0
        if (aOrder !== bOrder) return aOrder - bOrder
        return a.id - b.id
      })
      setDrugCatalog(sorted)
    } catch (error) {
      console.error('加载药品目录失败:', error)
    }
  }, [])

  // 初始加载
  useEffect(() => {
    loadData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 打开新增模态框
  const handleOpenAddModal = useCallback(async () => {
    await onRefreshDict()
    setIsEditing(false)
    setCurrentDrug(INITIAL_DRUG_FORM)
    setShowModal(true)
  }, [onRefreshDict])

  // 打开编辑模态框
  const handleOpenEditModal = useCallback((item: DrugCatalog) => {
    setIsEditing(true)
    setCurrentDrug({ ...item })
    setShowModal(true)
  }, [])

  // 保存药品
  const handleSaveDrug = useCallback(async () => {
    if (!currentDrug.name || !currentDrug.category) {
      alert('类别、名称为必填项')
      return
    }

    // 方法类别：清空药品专属字段
    const isMethod = currentDrug.category === METHOD_CATEGORY_CODE
    const drugData = isMethod ? {
      ...currentDrug,
      genericName: '',
      shortName: '',
      mnemonic: '',
      spec: '',
      concentration: '',
      specUnit: '',
      minUnitDose: '',
      brand: '',
      packaging: '',
      manufacturer: '',
      standardType: '',
    } : currentDrug

    try {
      if (isEditing && drugData.id) {
        await drugCatalogApi.update(drugData.id, {
          name: drugData.name,
          shortName: drugData.shortName,
          mnemonic: drugData.mnemonic,
          genericName: drugData.genericName,
          category: drugData.category,
          spec: drugData.spec,
          concentration: drugData.concentration,
          specUnit: drugData.specUnit,
          minUnitDose: drugData.minUnitDose,
          baseUnit: drugData.baseUnit,
          brand: drugData.brand,
          packaging: drugData.packaging,
          manufacturer: drugData.manufacturer,
          standardType: drugData.standardType,
          timing: drugData.timing,
          tips: drugData.tips,
          sortOrder: drugData.sortOrder,
          isEnabled: drugData.isEnabled ?? true,
          note: drugData.note
        })
      } else {
        await drugCatalogApi.create({
          code: drugData.code || '',
          name: drugData.name || '',
          shortName: drugData.shortName,
          mnemonic: drugData.mnemonic,
          genericName: drugData.genericName,
          category: drugData.category || '',
          spec: drugData.spec,
          concentration: drugData.concentration,
          specUnit: drugData.specUnit,
          minUnitDose: drugData.minUnitDose,
          baseUnit: drugData.baseUnit,
          brand: drugData.brand,
          packaging: drugData.packaging,
          manufacturer: drugData.manufacturer,
          standardType: drugData.standardType,
          timing: drugData.timing,
          tips: drugData.tips,
          sortOrder: drugData.sortOrder,
          isEnabled: drugData.isEnabled ?? true,
          note: drugData.note
        })
      }
      await loadData()
      setShowModal(false)
    } catch (error) {
      console.error('保存药品失败:', error)
      alert('保存失败，请稍后重试')
    }
  }, [currentDrug, isEditing, loadData])

  // 删除
  const handleDeleteDrug = useCallback(async (id: number) => {
    if (window.confirm('确定要删除该药品条目吗？')) {
      try {
        await drugCatalogApi.delete(id)
        await loadData()
      } catch (error) {
        console.error('删除药品失败:', error)
        alert('删除失败，请稍后重试')
      }
    }
  }, [loadData])

  // 切换状态
  const handleToggleDrugStatus = useCallback(async (id: number) => {
    try {
      await drugCatalogApi.toggleEnabled(id)
      await loadData()
    } catch (error) {
      console.error('切换药品状态失败:', error)
      alert('操作失败，请稍后重试')
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

    try {
      await Promise.all(
        Array.from(selection.selectedIds).map(id => drugCatalogApi.delete(Number(id)))
      )
      await loadData()
      selection.clearSelection()
      alert('删除成功')
    } catch (error) {
      console.error('批量删除失败:', error)
      alert('删除失败')
    }
  }, [selection, loadData])

  // 导出药品数据到 Excel
  const handleExport = useCallback(async () => {
    try {
      const response = await drugCatalogApi.list({ pageSize: LOAD_ALL_PAGE_SIZE })
      const sortedData = [...response.items].sort((a, b) => a.id - b.id)

      const excelData = sortedData.map(item => ({
        '药品编码': item.code,
        '药品名称': item.name,
        '通用名': item.genericName || '',
        '药品分类': item.category,
        '规格': item.spec || '',
        '浓度': item.concentration || '',
        '单位': item.baseUnit || '',
        '生产厂家': item.manufacturer || '',
        '启用状态': item.isEnabled ? '是' : '否',
        '备注': item.note || ''
      }))

      const ws = XLSX.utils.json_to_sheet(excelData)
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, '药品目录')

      ws['!cols'] = [
        { wch: 12 }, { wch: 20 }, { wch: 15 }, { wch: 12 },
        { wch: 15 }, { wch: 15 }, { wch: 8 }, { wch: 20 }, { wch: 8 }, { wch: 30 }
      ]

      XLSX.writeFile(wb, `药品目录_${new Date().toLocaleDateString('zh-CN')}.xlsx`)
    } catch (error) {
      console.error('导出失败:', error)
      alert('导出失败，请稍后重试')
    }
  }, [])

  // 批量导入
  const handleBatchImport = useCallback(async (
    data: unknown[],
    autoAddCategories = false,
    onProgress?: (percent: number) => void
  ) => {
    const token = getToken()
    if (!token) {
      return {
        success: 0,
        failed: data.length,
        errors: ['用户未登录或登录已过期，请先登录后再试']
      }
    }

    // 第一步：收集Excel中的所有药品分类
    const excelCategories = new Set<string>()
    for (const item of data) {
      const row = item as { category?: string }
      if (row.category) {
        excelCategories.add(String(row.category).trim())
      }
    }

    // 第二步：获取系统中现有的药品分类
    const existingDictItems = await dictApi.getItems(DICT_TYPES.DRUG_CATEGORY, true)
    const existingCategories = new Set(existingDictItems.map(item => item.name))

    // 第三步：找出缺失的分类
    const missingCategories = Array.from(excelCategories).filter(cat => !existingCategories.has(cat))

    // 如果有缺失的分类且用户未确认自动添加，返回特殊状态
    if (missingCategories.length > 0 && !autoAddCategories) {
      return {
        success: 0,
        failed: 0,
        errors: [],
        needConfirmCategories: true,
        missingCategories,
        data // 保存原始数据以便后续使用
      }
    }

    // 如果用户确认自动添加，先添加缺失的分类
    if (missingCategories.length > 0 && autoAddCategories) {
      for (const categoryName of missingCategories) {
        try {
          // 生成分类代码：使用拼音首字母或简化处理
          const code = categoryName
            .toLowerCase()
            .replace(/[^a-z0-9\u4e00-\u9fa5]/g, '')
            .substring(0, 20) || 'custom'

          await dictApi.createItem({
            typeCode: DICT_TYPES.DRUG_CATEGORY,
            code,
            name: categoryName,
            isEnabled: true
          })
        } catch (err) {
          return {
            success: 0,
            failed: data.length,
            errors: [`添加药品分类 "${categoryName}" 失败：${err instanceof Error ? err.message : '未知错误'}`]
          }
        }
      }
      // 刷新字典缓存
      await onRefreshDict()
    }

    // 第四步：获取现有的药品 code，避免冲突
    const existingDrugs = await drugCatalogApi.list({ pageSize: LOAD_ALL_PAGE_SIZE })
    const existingCodes = new Set(existingDrugs.items.map(d => d.code))

    // 第五步：执行批量导入
    let successCount = 0
    let failCount = 0
    const errors: string[] = []

    for (let i = 0; i < data.length; i++) {
      try {
        const item = data[i] as {
          code: string
          name: string
          genericName?: string
          category: string
          spec?: string
          concentration?: string
          unit?: string
          manufacturer?: string
          isEnabled?: boolean | string
          notes?: string
        }
        if (!item.name || !item.category) {
          errors.push(`第 ${i + 1} 行：缺少必填字段（名称、分类）`)
          failCount++
          onProgress?.(Math.round(((i + 1) / data.length) * 100))
          continue
        }

        // 方法类别：只需 name + category，其他字段可选
        const isMethod = String(item.category).trim() === METHOD_CATEGORY_CODE

        let isEnabled = true
        if (item.isEnabled !== undefined && item.isEnabled !== null && item.isEnabled !== '') {
          if (typeof item.isEnabled === 'boolean') {
            isEnabled = item.isEnabled
          } else if (typeof item.isEnabled === 'string') {
            const str = item.isEnabled.toString().trim().toLowerCase()
            isEnabled = (str === '是' || str === 'true' || str === '1' || str === 'yes')
          }
        }

        // 确保所有字段都有有效值，避免 undefined/null 导致后端验证失败
        // 处理 code 字段：如果为空或已存在，生成唯一编码
        let codeValue = String(item.code || '').trim()
        if (!codeValue || existingCodes.has(codeValue)) {
          // 生成唯一 code
          const nameForCode = String(item.name)
            .toLowerCase()
            .replace(/[^a-z0-9\u4e00-\u9fa5]/g, '')
            .substring(0, 15)
          let suffix = 1
          codeValue = `${nameForCode}_${suffix}`

          // 确保生成的 code 不与现有 code 冲突
          while (existingCodes.has(codeValue)) {
            suffix++
            codeValue = `${nameForCode}_${suffix}`
          }
        }

        const safeStr = (v: unknown) => (v === undefined || v === null) ? '' : String(v)

        const createData = {
          code: codeValue,
          name: String(item.name),
          genericName: isMethod ? '' : safeStr(item.genericName),
          category: String(item.category),
          spec: isMethod ? '' : safeStr(item.spec),
          concentration: isMethod ? '' : safeStr(item.concentration),
          baseUnit: safeStr(item.unit),
          manufacturer: isMethod ? '' : safeStr(item.manufacturer),
          isEnabled,
          note: safeStr(item.notes)
        }

        await drugCatalogApi.create(createData)
        // 将新创建的 code 加入已存在集合，避免后续冲突
        existingCodes.add(codeValue)
        successCount++
        onProgress?.(Math.round(((i + 1) / data.length) * 100))
      } catch (err) {
        failCount++
        errors.push(`第 ${i + 1} 行：${err instanceof Error ? err.message : '导入失败'}`)
        onProgress?.(Math.round(((i + 1) / data.length) * 100))
      }
    }
    onProgress?.(100)

    await loadData()
    if (failCount === 0) {
      setShowBatchImportModal(false)
    }
    return { success: successCount, failed: failCount, errors }
  }, [loadData, onRefreshDict])

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
              placeholder="搜索药品目录..."
              value={search.keyword}
              onChange={(e) => search.setKeyword(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all"
            />
          </div>
          {/* 状态过滤 */}
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as 'all' | 'enabled' | 'disabled')}
            className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-600 focus:ring-2 focus:ring-blue-500 outline-none transition-all cursor-pointer"
          >
            <option value="all">全部状态</option>
            <option value="enabled">已启用</option>
            <option value="disabled">已停用</option>
          </select>
          {(search.keyword || statusFilter !== 'all') && (
            <span className="text-xs text-slate-500">
              找到 <span className="font-bold text-emerald-600">{filteredByStatus.length}</span> 条结果
            </span>
          )}
          {selection.selectedCount > 0 && (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-emerald-50 text-emerald-700 rounded-lg text-xs font-bold">
              <span>已选 {selection.selectedCount} 项</span>
              <button onClick={() => selection.clearSelection()} className="ml-1 text-emerald-600 hover:text-emerald-800">取消</button>
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
          <button onClick={handleOpenAddModal} className="flex items-center gap-2 px-5 py-2 bg-emerald-600 shadow-emerald-100 text-white rounded-xl text-xs font-black hover:opacity-90 shadow-lg transition-all">
            <Plus size={14} /> 新增药品条目
          </button>
        </div>
      </div>

      {/* 表格 - 匹配 TemplateCenter 设计 */}
      <div className="flex-1 overflow-auto p-6">
        <div className="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden overflow-x-auto">
          <table className="w-full text-left text-sm min-w-[1400px]">
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
                <th className="px-6 py-5">类别</th>
                <th className="px-6 py-5">药品名称</th>
                <th className="px-6 py-5">简称</th>
                <th className="px-6 py-5">编码</th>
                <th className="px-6 py-5">品牌</th>
                <th className="px-6 py-5">规格</th>
                <th className="px-6 py-5">基本单位</th>
                <th className="px-6 py-5 text-center">排序</th>
                <th className="px-6 py-5 text-center">禁用状态</th>
                <th className="px-6 py-5 text-right">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50 font-bold text-slate-700">
              {pagination.currentPageData.map((item: DrugCatalog, index: number) => (
                <tr
                  key={item.id}
                  className={`hover:bg-slate-50 transition-colors ${!item.isEnabled ? 'opacity-50 grayscale' : ''} ${selection.selectedIds.has(String(item.id)) ? 'bg-emerald-50/50' : ''}`}
                >
                  <td className="px-6 py-6 text-center">
                    <Checkbox
                      checked={selection.selectedIds.has(String(item.id))}
                      onChange={() => selection.toggleSelection(item.id)}
                    />
                  </td>
                  <td className="px-6 py-6 text-center text-slate-400 font-mono text-xs">{pagination.displayInfo.startIndex + index}</td>
                  <td className="px-6 py-6 text-slate-500">{getCategoryLabel(item.category)}</td>
                  <td className="px-6 py-6 flex items-center gap-3">
                    <div className={`p-2 rounded-xl ${!item.isEnabled ? 'bg-slate-100 text-slate-400' : 'bg-emerald-50 text-emerald-600'}`}>
                      <Pill size={18} />
                    </div>
                    <div className="flex flex-col">
                      <span className="text-slate-800 font-bold">{item.name}</span>
                      {!isMethodItem(item) && item.genericName && <span className="text-[9px] text-slate-400">({item.genericName})</span>}
                      {!item.isEnabled && <span className="text-[9px] font-black text-red-500 uppercase tracking-tighter">已禁用</span>}
                    </div>
                  </td>
                  <td className="px-6 py-6 text-slate-600 text-xs">{isMethodItem(item) ? '-' : (item.shortName || '-')}</td>
                  <td className="px-6 py-6 font-mono text-xs text-blue-600">{item.code}</td>
                  <td className="px-6 py-6 text-slate-600">{isMethodItem(item) ? '-' : (item.brand || '-')}</td>
                  <td className="px-6 py-6 text-slate-500 text-xs">{isMethodItem(item) ? '-' : (item.spec || '-')}</td>
                  <td className="px-6 py-6">{item.baseUnit || '-'}</td>
                  <td className="px-6 py-6 text-center text-slate-500 text-xs">{item.sortOrder ?? 0}</td>
                  <td className="px-6 py-6 text-center">
                    <button
                      onClick={() => handleToggleDrugStatus(item.id)}
                      className={`p-1 rounded-lg transition-all ${!item.isEnabled ? 'text-red-500 bg-red-50' : 'text-emerald-500 bg-emerald-50'}`}
                    >
                      {!item.isEnabled ? <ToggleLeft size={24} /> : <ToggleRight size={24} />}
                    </button>
                  </td>
                  <td className="px-6 py-6 text-right space-x-1">
                    <button onClick={() => handleOpenEditModal(item)} className="p-2 text-slate-400 hover:text-blue-600 hover:bg-white rounded-lg transition-all shadow-sm">
                      <Edit3 size={16} />
                    </button>
                    <button onClick={() => handleDeleteDrug(item.id)} className="p-2 text-slate-400 hover:text-red-500 hover:bg-white rounded-lg transition-all shadow-sm">
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

      {/* 药品新增/编辑弹窗 - 采用分段式紧凑布局 */}
      {showModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-5xl overflow-hidden flex flex-col animate-scale-in">
            {/* 标题栏 */}
            <div className="px-8 py-4 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className={`w-10 h-10 ${isEditing ? 'bg-blue-500' : 'bg-emerald-600'} rounded-xl flex items-center justify-center text-white shadow-lg shadow-emerald-100`}>
                  {isEditing ? <Edit3 size={20} strokeWidth={3} /> : <Plus size={20} strokeWidth={3} />}
                </div>
                <div>
                  <h3 className="text-lg font-black text-slate-900 tracking-tight">{isEditing ? '编辑药品条目' : '新增药品条目'}</h3>
                  <p className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">Drug Dictionary Configuration</p>
                </div>
              </div>
              <button onClick={() => setShowModal(false)} className="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-all">
                <X size={20} />
              </button>
            </div>

            {/* 表单内容 */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 no-scrollbar bg-white">
              {/* 核心信息 */}
              <FormSection title="核心信息" icon={Info}>
                <SelectField
                  label="类别"
                  required
                  options={dictOptions[DICT_TYPES.DRUG_CATEGORY] || []}
                  value={currentDrug.category}
                  onChange={(e) => setCurrentDrug({ ...currentDrug, category: e.target.value })}
                />
                <InputField
                  label={isMethodCategory ? '方法名称' : '药品名称'}
                  required
                  placeholder={isMethodCategory ? '方法全称' : '完整药品全称'}
                  value={currentDrug.name}
                  onChange={(e) => setCurrentDrug({ ...currentDrug, name: e.target.value })}
                />
                {!isMethodCategory && (
                  <>
                    <InputField
                      label="通用名"
                      placeholder="药品通用名"
                      value={currentDrug.genericName}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, genericName: e.target.value })}
                    />
                    <InputField
                      label="简称"
                      placeholder="临床简称"
                      value={currentDrug.shortName}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, shortName: e.target.value })}
                    />
                    <InputField
                      label="助记码"
                      placeholder="简拼搜索码"
                      value={currentDrug.mnemonic}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, mnemonic: e.target.value })}
                    />
                  </>
                )}
              </FormSection>

              {/* 规格与属性 */}
              <FormSection title="规格与属性" icon={Sliders}>
                {!isMethodCategory && (
                  <>
                    <InputField
                      label="规格"
                      placeholder="如：0.4ml/2500iu"
                      value={currentDrug.spec}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, spec: e.target.value })}
                    />
                    <InputField
                      label="浓度"
                      placeholder="如：1250IU/ml"
                      value={currentDrug.concentration}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, concentration: e.target.value })}
                    />
                    <InputField
                      label="规格单位"
                      placeholder="如：iu / mg / ml"
                      value={currentDrug.specUnit}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, specUnit: e.target.value })}
                    />
                    <InputField
                      label="最小单位剂量"
                      placeholder="换算录入数值"
                      value={currentDrug.minUnitDose}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, minUnitDose: e.target.value })}
                    />
                  </>
                )}
                <InputField
                  label={isMethodCategory ? '单位' : '基本单位'}
                  placeholder={isMethodCategory ? '如：次 / 分钟' : '如：支 / 瓶 / 盒'}
                  value={currentDrug.baseUnit}
                  onChange={(e) => setCurrentDrug({ ...currentDrug, baseUnit: e.target.value })}
                />
                <InputField
                  label="编码"
                  placeholder="HIS/系统编码（可选）"
                  value={currentDrug.code}
                  onChange={(e) => setCurrentDrug({ ...currentDrug, code: e.target.value })}
                />
                {!isMethodCategory && (
                  <>
                    <InputField
                      label="品牌"
                      placeholder="制剂品牌"
                      value={currentDrug.brand}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, brand: e.target.value })}
                    />
                    <InputField
                      label="包装"
                      placeholder="如：10支/盒"
                      value={currentDrug.packaging}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, packaging: e.target.value })}
                    />
                    <InputField
                      label="生产厂家"
                      placeholder="厂家全名"
                      value={currentDrug.manufacturer}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, manufacturer: e.target.value })}
                    />
                  </>
                )}
              </FormSection>

              {/* 临床使用 */}
              <FormSection title="临床使用" icon={Activity}>
                {!isMethodCategory && (
                  <SelectField
                    label="标准分类"
                    options={DICT_STD_TYPES}
                    value={currentDrug.standardType}
                    onChange={(e) => setCurrentDrug({ ...currentDrug, standardType: e.target.value })}
                  />
                )}
                <SelectField
                  label="使用时机"
                  options={DICT_TIMINGS}
                  value={currentDrug.timing}
                  onChange={(e) => setCurrentDrug({ ...currentDrug, timing: e.target.value })}
                />
                <div className="md:col-span-2 lg:col-span-2">
                  <InputField
                    label="使用贴士"
                    placeholder="用药注意事项、推注速度建议等临床提醒"
                    value={currentDrug.tips}
                    onChange={(e) => setCurrentDrug({ ...currentDrug, tips: e.target.value })}
                  />
                </div>
              </FormSection>

              {/* 备注和状态控制 */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 p-5 bg-slate-50/50 rounded-2xl border border-slate-100">
                <div className="md:col-span-2 space-y-4">
                  <div className="space-y-1">
                    <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">备注信息</label>
                    <textarea
                      value={currentDrug.note}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, note: e.target.value })}
                      className="w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-emerald-500/10 focus:border-emerald-500 transition-all resize-none h-16"
                      placeholder="输入药品补充说明..."
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <InputField
                      label="药品排序"
                      type="number"
                      placeholder="列表显示顺序"
                      value={currentDrug.sortOrder}
                      onChange={(e) => setCurrentDrug({ ...currentDrug, sortOrder: parseInt(e.target.value) || 0 })}
                    />
                  </div>
                </div>
                <div className="flex flex-col justify-center gap-3">
                  <p className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">状态控制</p>
                  <div className="flex items-center gap-4 bg-white p-3 rounded-xl border border-slate-100 shadow-sm">
                    <span className="text-xs font-bold text-slate-600 flex-1">是否禁用该药品</span>
                    <button
                      onClick={() => setCurrentDrug({ ...currentDrug, isEnabled: !(currentDrug.isEnabled ?? true) })}
                      className={`p-0.5 rounded-lg transition-all ${!(currentDrug.isEnabled ?? true) ? 'text-red-500 bg-red-50' : 'text-emerald-500 bg-emerald-50'}`}
                    >
                      {!(currentDrug.isEnabled ?? true) ? <ToggleLeft size={28} /> : <ToggleRight size={28} />}
                    </button>
                  </div>
                  <div className="flex items-start gap-1.5 text-[9px] text-slate-400 leading-relaxed px-1">
                    <AlertCircle size={10} className="shrink-0 mt-0.5" />
                    <span>禁用后，该药品在处方开立和医嘱执行的检索列表中将不再出现。</span>
                  </div>
                </div>
              </div>
            </div>

            {/* 底部按钮 */}
            <div className="px-8 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3 shrink-0">
              <button onClick={() => setShowModal(false)} className="px-6 py-2 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-100 transition-all">取消</button>
              <button onClick={handleSaveDrug} className="px-10 py-2 bg-emerald-600 text-white rounded-xl text-xs font-black shadow-xl shadow-emerald-100 hover:bg-emerald-700 transition-all">确认保存</button>
            </div>
          </div>
        </div>
      )}

      {/* 批量导入模态框 */}
      {showBatchImportModal && (
        <BatchImportModal
          title="批量导入药品"
          columns={DRUG_EXCEL_COLUMNS}
          onImport={handleBatchImport}
          onExport={handleExport}
          onClose={() => setShowBatchImportModal(false)}
          maxSize={5}
        />
      )}
    </div>
  )
}

export const DrugTab = memo(DrugTabComponent)
