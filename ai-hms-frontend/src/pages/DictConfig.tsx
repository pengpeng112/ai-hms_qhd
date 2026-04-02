// DictConfig - 字典配置管理页面

import React, { useState, useEffect, useRef, useMemo } from 'react'
import {
  BookOpen, Plus, Edit3, Trash2, ToggleLeft, ToggleRight,
  ChevronRight, ChevronDown, X, Loader2, CheckCircle2, Clock,
  LayoutGrid, List, Download, Upload
} from 'lucide-react'
import { dictApi, dictCache, type DictType, type DictItem } from '@/services/dictApi'

// 可选图标列表
const ICON_OPTIONS = [
  '📚', '🩺', '💉', '🧪', '⚗️', '💧', '🍬', '📦', '💊', '📋',
  '📁', '🩸', '📍', '🔵', '🔴', '👨‍⚕️', '🏥', '🔪', '⚙️', '🔧',
  '📊', '📈', '🗂️', '📝', '🏷️', '🔬', '💡', '🎯', '⭐', '🌡️',
  '❤️', '🫀', '🫁', '🦴', '🧬', '🩻', '💪', '🧠', '👁️', '👂',
]

// 字典类型配置（显示图标和描述）
const DICT_TYPE_CONFIG: Record<string, { icon: string; color: string; description: string }> = {
  DIALYSIS_MODE: { icon: '🩺', color: 'blue', description: '血液透析治疗方式' },
  ANTICOAGULANT: { icon: '💉', color: 'red', description: '透析抗凝剂类型' },
  DIALYSATE_TYPE: { icon: '🧪', color: 'purple', description: '透析液配方类型' },
  DIALYSATE_GROUP: { icon: '⚗️', color: 'cyan', description: '透析液离子组分' },
  DIALYSATE_FLOW: { icon: '💧', color: 'blue', description: '透析液流速' },
  GLUCOSE: { icon: '🍬', color: 'amber', description: '透析液葡萄糖含量' },
  MATERIAL_CATEGORY: { icon: '📦', color: 'orange', description: '透析耗材分类' },
  DRUG_CATEGORY: { icon: '💊', color: 'green', description: '透析药品分类' },
  ORDER_TYPE: { icon: '📋', color: 'indigo', description: '医嘱类型' },
  ORDER_CATEGORY: { icon: '📁', color: 'violet', description: '医嘱分类' },
  VASCULAR_ACCESS: { icon: '🩸', color: 'red', description: '血管通路类型' },
  VASCULAR_SITE: { icon: '📍', color: 'pink', description: '血管通路部位' },
  VEIN_TYPE: { icon: '🔵', color: 'blue', description: '静脉类型' },
  ARTERY_TYPE: { icon: '🔴', color: 'red', description: '动脉类型' },
  DOCTOR: { icon: '👨‍⚕️', color: 'teal', description: '医生列表' },
  HOSPITAL: { icon: '🏥', color: 'rose', description: '手术医院' },
  SURGERY_TYPE: { icon: '🔪', color: 'slate', description: '血管通路干预手术类型' },
  // 临床诊疗分类
  PRIMARY_DISEASE: { icon: '🦠', color: 'rose', description: '原发疾病分类' },
  COMPLICATION: { icon: '⚠️', color: 'orange', description: '透析并发症类型' },
  PATHOLOGY: { icon: '🔬', color: 'purple', description: '病理诊断分类' },
  TUMOR: { icon: '🎗️', color: 'red', description: '肿瘤类型分类' },
  ALLERGEN: { icon: '🌿', color: 'green', description: '过敏原分类' },
  // 转归字典
  OUTCOME: { icon: '📋', color: 'blue', description: '患者转归（在科/转出及原因）' },
  // 医嘱用药扩展
  ORDER_ROUTE: { icon: '💉', color: 'blue', description: '医嘱用法' },
  ORDER_FREQUENCY: { icon: '🔄', color: 'purple', description: '医嘱频次' },
  ORDER_TIMING: { icon: '⏰', color: 'orange', description: '医嘱使用时机' },
}

// 字典分组定义（前端虚拟分组，不影响后端）
const DICT_GROUPS: { key: string; label: string }[] = [
  { key: 'dialysis', label: '透析治疗' },
  { key: 'vascular', label: '血管通路' },
  { key: 'order', label: '医嘱处方' },
  { key: 'staff', label: '人员信息' },
  { key: 'clinical', label: '临床诊疗' },
  { key: 'outcome', label: '转归记录' },
]

// 默认分组映射（code → group key）
const DEFAULT_GROUP_MAPPING: Record<string, string> = {
  DIALYSIS_MODE: 'dialysis', ANTICOAGULANT: 'dialysis', DIALYSATE_TYPE: 'dialysis',
  DIALYSATE_GROUP: 'dialysis', DIALYSATE_FLOW: 'dialysis', GLUCOSE: 'dialysis',
  VASCULAR_ACCESS: 'vascular', VASCULAR_SITE: 'vascular', VEIN_TYPE: 'vascular',
  ARTERY_TYPE: 'vascular', SURGERY_TYPE: 'vascular', HOSPITAL: 'vascular',
  ORDER_TYPE: 'order', ORDER_CATEGORY: 'order', DRUG_CATEGORY: 'order', MATERIAL_CATEGORY: 'order',
  ORDER_ROUTE: 'order', ORDER_FREQUENCY: 'order', ORDER_TIMING: 'order',
  DOCTOR: 'staff',
  // 临床诊疗分组
  PRIMARY_DISEASE: 'clinical', COMPLICATION: 'clinical', PATHOLOGY: 'clinical', TUMOR: 'clinical', ALLERGEN: 'clinical',
  // 转归分组
  OUTCOME: 'outcome',
}
// localStorage 读写分组映射
const STORAGE_KEY = 'dict_group_mapping'
function loadGroupMapping(): Record<string, string> {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) return { ...DEFAULT_GROUP_MAPPING, ...JSON.parse(saved) }
  } catch { /* ignore */ }
  return { ...DEFAULT_GROUP_MAPPING }
}
function persistGroupMapping(mapping: Record<string, string>) {
  // 只持久化与默认不同的条目
  const custom: Record<string, string> = {}
  for (const [k, v] of Object.entries(mapping)) {
    if (DEFAULT_GROUP_MAPPING[k] !== v) custom[k] = v
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(custom))
}

// 辅助组件
const FormSection = ({ title, icon: Icon, children }: { title: string; icon: React.ElementType; children?: React.ReactNode }) => (
  <div className="bg-slate-50/50 rounded-2xl border border-slate-100 p-4 space-y-3">
    <div className="flex items-center gap-2 mb-1">
      <div className="p-1.5 bg-white rounded-lg shadow-sm text-blue-600">
        <Icon size={16} />
      </div>
      <h4 className="text-[11px] font-black text-slate-800 uppercase tracking-widest">{title}</h4>
    </div>
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {children}
    </div>
  </div>
)

interface InputFieldProps {
  label: string
  placeholder?: string
  required?: boolean
  type?: string
  value?: string | number
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void
}

const InputField = ({ label, placeholder, required, value, onChange }: InputFieldProps) => (
  <div className="space-y-1">
    <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
      {required && <span className="text-red-500 mr-1">*</span>}{label}
    </label>
    <input
      type="text"
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      className="w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all placeholder:text-slate-300 placeholder:font-medium"
    />
  </div>
)

export default function DictConfig() {
  // 视图状态
  const [view, setView] = useState<'types' | 'items'>('types')
  const [selectedType, setSelectedType] = useState<DictType | null>(null)
  const [layoutMode, setLayoutMode] = useState<'table' | 'card'>('table') // 布局模式
  const [selectedGroup, setSelectedGroup] = useState<string>('all') // 分组筛选
  const [groupMapping, setGroupMapping] = useState<Record<string, string>>(loadGroupMapping) // 分组映射

  // 数据状态
  const [dictTypes, setDictTypes] = useState<DictType[]>([])
  const [dictItems, setDictItems] = useState<DictItem[]>([])
  const [loading, setLoading] = useState(false)
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set())

  // 导入相关
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [importing, setImporting] = useState(false)
  const [importPreview, setImportPreview] = useState<{
    data: { version: string; types: DictType[]; items: DictItem[] }
    selectedCodes: Set<string>
  } | null>(null)

  // 批量选择状态
  const [selectedTypeIds, setSelectedTypeIds] = useState<Set<string>>(new Set())

  // 字典类型表单状态
  const [showTypeModal, setShowTypeModal] = useState(false)
  const [isEditingType, setIsEditingType] = useState(false)
  const [typeForm, setTypeForm] = useState<{
    id: string
    code: string
    name: string
    description: string
    icon: string
    sortOrder: number
    isEnabled: boolean
    group: string
  }>({
    id: '',
    code: '',
    name: '',
    description: '',
    icon: '📚',
    sortOrder: 0,
    isEnabled: true,
    group: '',
  })

  // 字典项表单状态
  const [showItemModal, setShowItemModal] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [itemForm, setItemForm] = useState<{
    id: string
    code: string
    name: string
    description: string
    sortOrder: number
    isEnabled: boolean
    parentCode: string
  }>({
    id: '',
    code: '',
    name: '',
    description: '',
    sortOrder: 0,
    isEnabled: true,
    parentCode: '',
  })

  // 加载字典类型列表
  const loadDictTypes = async () => {
    setLoading(true)
    try {
      const types = await dictApi.listTypes()
      setDictTypes(types)
    } catch (error) {
      console.error('加载字典类型失败:', error)
    } finally {
      setLoading(false)
    }
  }

  // 加载字典项列表（树形结构）
  const loadDictItems = async (typeCode: string) => {
    setLoading(true)
    try {
      const tree = await dictApi.getItemsTree(typeCode, false)
      // 验证 API 返回值是否为数组，如果不是则设置为空数组
      if (Array.isArray(tree)) {
        setDictItems(tree)
        console.log(`[DictConfig] 加载字典项成功: ${typeCode}, 共 ${tree.length} 条`)
      } else {
        console.warn(`[DictConfig] API 返回非数组数据:`, tree)
        setDictItems([])
      }
    } catch (error) {
      console.error('[DictConfig] 加载字典项失败:', error)
      // 确保 catch 块中也设置为空数组
      setDictItems([])
    } finally {
      setLoading(false)
    }
  }

  // 初始化加载
  useEffect(() => {
    loadDictTypes()
  }, [])

  // 根据分组筛选字典类型
  const filteredDictTypes = useMemo(() => {
    if (selectedGroup === 'all') return dictTypes
    if (selectedGroup === 'other') return dictTypes.filter(t => !groupMapping[t.code])
    return dictTypes.filter(t => groupMapping[t.code] === selectedGroup)
  }, [dictTypes, selectedGroup, groupMapping])

  // 打开新增字典类型弹窗
  const handleOpenAddTypeModal = () => {
    setIsEditingType(false)
    setTypeForm({
      id: '',
      code: '',
      name: '',
      description: '',
      icon: '📚',
      sortOrder: dictTypes.length > 0 ? Math.max(...dictTypes.map(t => t.sortOrder)) + 1 : 1,
      isEnabled: true,
      group: selectedGroup !== 'all' && selectedGroup !== 'other' ? selectedGroup : '',
    })
    setShowTypeModal(true)
  }

  // 打开编辑字典类型弹窗
  const handleOpenEditTypeModal = (type: DictType) => {
    setIsEditingType(true)
    setTypeForm({
      id: type.id,
      code: type.code,
      name: type.name,
      description: type.description,
      icon: DICT_TYPE_CONFIG[type.code]?.icon || '📚',
      sortOrder: type.sortOrder,
      isEnabled: type.isEnabled,
      group: groupMapping[type.code] || '',
    })
    setShowTypeModal(true)
  }

  // 保存字典类型
  const handleSaveType = async () => {
    if (!typeForm.code || !typeForm.name) {
      alert('代码和名称为必填项')
      return
    }

    setLoading(true)
    try {
      if (isEditingType && typeForm.id) {
        await dictApi.updateType(typeForm.id, {
          code: typeForm.code,
          name: typeForm.name,
          description: typeForm.description,
          sortOrder: typeForm.sortOrder,
          isEnabled: typeForm.isEnabled,
        })
      } else {
        await dictApi.createType({
          code: typeForm.code,
          name: typeForm.name,
          description: typeForm.description,
          sortOrder: typeForm.sortOrder,
          isEnabled: typeForm.isEnabled,
        })
      }

      // 更新分组映射
      const newMapping = { ...groupMapping }
      if (typeForm.group) {
        newMapping[typeForm.code] = typeForm.group
      } else {
        delete newMapping[typeForm.code]
      }
      setGroupMapping(newMapping)
      persistGroupMapping(newMapping)

      await loadDictTypes()
      setShowTypeModal(false)
    } catch (error) {
      console.error('保存字典类型失败:', error)
      alert('保存失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 删除单个字典类型（后端已级联删除所有字典项）
  const handleDeleteType = async (id: string, code: string) => {
    if (!window.confirm(`确定要删除该字典类型吗？\n注意：该类型下的所有字典项也将被删除！`)) {
      return
    }

    setLoading(true)
    try {
      await dictApi.deleteType(id)
      dictCache.clear(code)
      await loadDictTypes()
    } catch (error) {
      console.error('删除字典类型失败:', error)
      alert('删除失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 批量删除字典类型（后端已级联删除所有字典项）
  const handleBatchDeleteTypes = async () => {
    if (selectedTypeIds.size === 0) {
      alert('请先选择要删除的字典类型')
      return
    }

    if (!window.confirm(`确定要删除选中的 ${selectedTypeIds.size} 个字典类型吗？\n注意：这些类型下的所有字典项也将被删除！`)) {
      return
    }

    setLoading(true)
    try {
      for (const typeId of selectedTypeIds) {
        const type = dictTypes.find(t => t.id === typeId)
        if (type) {
          await dictApi.deleteType(typeId)
          dictCache.clear(type.code)
        }
      }
      setSelectedTypeIds(new Set())
      await loadDictTypes()
    } catch (error) {
      console.error('批量删除字典类型失败:', error)
      alert('删除失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 切换选中状态
  const handleToggleTypeSelection = (id: string) => {
    const newSelected = new Set(selectedTypeIds)
    if (newSelected.has(id)) {
      newSelected.delete(id)
    } else {
      newSelected.add(id)
    }
    setSelectedTypeIds(newSelected)
  }

  // 全选/取消全选（基于当前筛选结果）
  const handleToggleSelectAll = () => {
    if (selectedTypeIds.size === filteredDictTypes.length) {
      setSelectedTypeIds(new Set())
    } else {
      setSelectedTypeIds(new Set(filteredDictTypes.map(t => t.id)))
    }
  }

  // 选择字典类型
  const handleSelectType = (type: DictType) => {
    setSelectedType(type)
    setView('items')
    loadDictItems(type.code)
  }

  // 返回类型列表
  const handleBackToTypes = () => {
    setView('types')
    setSelectedType(null)
    setDictItems([])
    setExpandedItems(new Set())
  }

  // 切换展开/收起
  const toggleExpand = (code: string) => {
    setExpandedItems(prev => {
      const next = new Set(prev)
      if (next.has(code)) {
        next.delete(code)
      } else {
        next.add(code)
      }
      return next
    })
  }

  // 计算所有字典项总数（包含子项）
  const totalItemCount = useMemo(() => {
    // 防御性检查：确保 dictItems 是数组
    const items = Array.isArray(dictItems) ? dictItems : []
    let count = 0
    for (const item of items) {
      count += 1
      if (item.children) count += item.children.length
    }
    return count
  }, [dictItems])

  // 获取顶级项列表（用于父级分类下拉）
  const topLevelItems = useMemo(() => {
    // 防御性检查：确保 dictItems 是数组
    const items = Array.isArray(dictItems) ? dictItems : []
    return items.filter(item => !item.parentCode)
  }, [dictItems])

  // 打开新增字典项弹窗
  const handleOpenAddItemModal = (parentCode?: string) => {
    setIsEditing(false)
    if (parentCode) {
      // 快捷添加子项：自动预填
      const parent = dictItems.find(i => i.code === parentCode)
      const children = parent?.children || []
      // 取现有子项编码后缀的最大值 +1，避免删除中间项后编码冲突
      const maxSuffix = children.reduce((max, child) => {
        const parts = child.code.split('-')
        const num = parseInt(parts[parts.length - 1], 10)
        return isNaN(num) ? max : Math.max(max, num)
      }, 0)
      setItemForm({
        id: '',
        code: `${parentCode}-${String(maxSuffix + 1).padStart(2, '0')}`,
        name: '',
        description: '',
        sortOrder: children.length + 1,
        isEnabled: true,
        parentCode,
      })
    } else {
      // 计算所有项的最大 sortOrder
      let maxSort = 0
      for (const item of dictItems) {
        if (item.sortOrder > maxSort) maxSort = item.sortOrder
        if (item.children) {
          for (const child of item.children) {
            if (child.sortOrder > maxSort) maxSort = child.sortOrder
          }
        }
      }
      setItemForm({
        id: '',
        code: '',
        name: '',
        description: '',
        sortOrder: maxSort + 1,
        isEnabled: true,
        parentCode: '',
      })
    }
    setShowItemModal(true)
  }

  // 打开编辑字典项弹窗
  const handleOpenEditItemModal = (item: DictItem) => {
    setIsEditing(true)
    setItemForm({
      id: item.id,
      code: item.code,
      name: item.name,
      description: item.description,
      sortOrder: item.sortOrder,
      isEnabled: item.isEnabled,
      parentCode: item.parentCode || '',
    })
    setShowItemModal(true)
  }

  // 保存字典项
  const handleSaveItem = async () => {
    if (!itemForm.code || !itemForm.name) {
      alert('代码和名称为必填项')
      return
    }

    setLoading(true)
    try {
      if (isEditing && itemForm.id) {
        await dictApi.updateItem(itemForm.id, {
          code: itemForm.code,
          name: itemForm.name,
          description: itemForm.description,
          sortOrder: itemForm.sortOrder,
          isEnabled: itemForm.isEnabled,
          parent_code: itemForm.parentCode || null,
        })
      } else {
        if (!selectedType) {
          alert('未选择字典类型')
          return
        }
        await dictApi.createItem({
          typeCode: selectedType.code,
          code: itemForm.code,
          name: itemForm.name,
          description: itemForm.description,
          sortOrder: itemForm.sortOrder,
          isEnabled: itemForm.isEnabled,
          parentCode: itemForm.parentCode || undefined,
        })
      }

      // 清除缓存并重新加载
      dictCache.clear(selectedType?.code)
      if (selectedType) {
        await loadDictItems(selectedType.code)
      }
      setShowItemModal(false)
    } catch (error) {
      console.error('保存字典项失败:', error)
      alert('保存失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 删除字典项
  const handleDeleteItem = async (id: string) => {
    // 查找该项（可能是顶级项或子项）
    const item = dictItems.find(i => i.id === id) ||
      dictItems.flatMap(i => i.children || []).find(c => c.id === id)
    // 递归统计全部后代数量
    const countDescendants = (node: DictItem | undefined): number => {
      if (!node?.children?.length) return 0
      return node.children.reduce((sum, child) => sum + 1 + countDescendants(child), 0)
    }
    const descendantCount = countDescendants(item)
    const msg = descendantCount > 0
      ? `确定要删除该字典项吗？\n注意：该项下的 ${descendantCount} 个子项也将被删除！`
      : '确定要删除该字典项吗？'
    if (!window.confirm(msg)) {
      return
    }

    setLoading(true)
    try {
      await dictApi.deleteItem(id)
      dictCache.clear(selectedType?.code)
      if (selectedType) {
        await loadDictItems(selectedType.code)
      }
    } catch (error) {
      console.error('删除字典项失败:', error)
      alert('删除失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 切换启用状态
  const handleToggleEnabled = async (id: string) => {
    setLoading(true)
    try {
      await dictApi.toggleItemEnabled(id)
      dictCache.clear(selectedType?.code)
      if (selectedType) {
        await loadDictItems(selectedType.code)
      }
    } catch (error) {
      console.error('切换状态失败:', error)
      alert('操作失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 导出字典数据
  // 优先级：
  // 1) 字典项视图：导出当前选中类型
  // 2) 字典类型视图且有勾选：导出勾选类型
  // 3) 其他情况：导出全部
  const handleExport = async () => {
    setLoading(true)
    try {
      const isSingleType = view === 'items' && selectedType
      const selectedTypesInList = view === 'types'
        ? dictTypes.filter((type) => selectedTypeIds.has(type.id))
        : []

      const typesToExport = isSingleType
        ? [selectedType]
        : selectedTypesInList.length > 0
          ? selectedTypesInList
          : dictTypes

      const allItems: DictItem[] = []
      for (const type of typesToExport) {
        try {
          const items = await dictApi.getItems(type.code, false)
          allItems.push(...items.map(item => ({ ...item, typeCode: type.code })))
        } catch (err) {
          console.error(`获取 ${type.code} 字典项失败:`, err)
        }
      }

      const exportData = {
        exportTime: new Date().toISOString(),
        version: '1.0',
        types: typesToExport,
        items: allItems
      }

      const suffix = isSingleType
        ? selectedType.code
        : selectedTypesInList.length > 0
          ? `selected_${selectedTypesInList.length}`
          : 'all'
      const filename = `dict_${suffix}_${new Date().toISOString().slice(0, 10)}.json`

      const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('导出失败:', error)
      alert('导出失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 导入字典数据
  const handleImport = () => {
    fileInputRef.current?.click()
  }

  // 处理文件导入 — 解析文件后打开预览弹窗
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    try {
      const text = await file.text()
      const data = JSON.parse(text)

      if (!data.version || !Array.isArray(data.types)) {
        alert('无效的导入文件格式')
        return
      }

      const normalizedTypes: DictType[] = (data.types as Array<Partial<DictType>>)
        .filter((type) => Boolean(type?.code))
        .map((type) => ({
          id: type.id || '',
          code: type.code!,
          name: type.name || type.code!,
          description: type.description || '',
          icon: type.icon,
          sortOrder: type.sortOrder ?? 0,
          isEnabled: type.isEnabled ?? true,
          createdAt: type.createdAt || '',
          updatedAt: type.updatedAt || '',
        }))

      const normalizedItems: DictItem[] = (Array.isArray(data.items) ? data.items : [])
        .map((item: DictItem & { type_code?: string }) => ({
          ...item,
          typeCode: item.typeCode || item.type_code || '',
        }))
        .filter((item: DictItem) => Boolean(item.typeCode))

      if (normalizedTypes.length === 0) {
        alert('导入文件中没有有效的字典类型')
        return
      }

      // 打开预览弹窗，默认全选
      const allCodes = new Set<string>(normalizedTypes.map((t: DictType) => t.code))
      setImportPreview({
        data: {
          version: data.version,
          types: normalizedTypes,
          items: normalizedItems,
        },
        selectedCodes: allCodes,
      })
    } catch {
      alert('导入失败，请检查文件格式')
    } finally {
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  // 确认导入选中的字典类型
  const handleConfirmImport = async () => {
    if (!importPreview) return
    const { data, selectedCodes } = importPreview

    if (selectedCodes.size === 0) {
      alert('请至少选择一个字典类型')
      return
    }

    setImporting(true)
    try {
      const filteredTypes = data.types.filter((t: DictType) => selectedCodes.has(t.code))
      const filteredItems = (data.items || []).filter((i: DictItem) => selectedCodes.has(i.typeCode))

      const result = await dictApi.importData({
        types: filteredTypes,
        items: filteredItems
      })

      // 清除缓存并重新加载
      dictCache.clear()
      if (view === 'items' && selectedType) {
        await loadDictItems(selectedType.code)
      } else {
        await loadDictTypes()
      }

      setImportPreview(null)
      alert(`导入完成！\n\n字典类型:\n  - 新增: ${result.typesCreated} 个\n  - 更新: ${result.typesUpdated} 个\n\n字典项:\n  - 新增: ${result.itemsCreated} 条\n  - 更新: ${result.itemsUpdated} 条`)
    } catch (error) {
      console.error('导入失败:', error)
      alert('导入失败，请稍后重试')
    } finally {
      setImporting(false)
    }
  }

  // 渲染字典类型列表
  const renderTypesView = () => {
    // 检查是否存在未分组的字典类型
    const hasUngrouped = dictTypes.some(t => !groupMapping[t.code])

    return (
    <div className="flex-1 flex flex-col">
      {/* 头部 */}
      <div className="bg-white p-4 border-b border-slate-100 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-blue-50 rounded-xl">
            <BookOpen size={20} className="text-blue-600" />
          </div>
          <div>
            <h2 className="text-lg font-black text-slate-900">字典类型</h2>
            <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest">Dictionary Types · {dictTypes.length} 个类型</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {layoutMode === 'table' && selectedTypeIds.size > 0 && (
            <button
              onClick={handleBatchDeleteTypes}
              disabled={loading}
              className="flex items-center gap-2 px-4 py-2 bg-red-500 text-white rounded-xl text-xs font-black hover:bg-red-600 transition-all disabled:opacity-50"
            >
              <Trash2 size={14} /> 删除选中 ({selectedTypeIds.size})
            </button>
          )}
          <button
            onClick={handleOpenAddTypeModal}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-xl text-xs font-black hover:bg-blue-700 shadow-lg shadow-blue-100 transition-all"
          >
            <Plus size={14} /> 新增字典
          </button>
        </div>
      </div>

      {/* 分组筛选栏 */}
      <div className="bg-white px-6 py-2 border-b border-slate-100 flex items-center gap-1 shrink-0 overflow-x-auto">
        <button
          onClick={() => setSelectedGroup('all')}
          className={`px-3 py-1.5 rounded-lg text-xs font-bold whitespace-nowrap transition-all ${
            selectedGroup === 'all'
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-slate-500 hover:bg-slate-100'
          }`}
        >
          全部 ({dictTypes.length})
        </button>
        {DICT_GROUPS.map(group => {
          const count = dictTypes.filter(t => groupMapping[t.code] === group.key).length
          return (
            <button
              key={group.key}
              onClick={() => setSelectedGroup(group.key)}
              className={`px-3 py-1.5 rounded-lg text-xs font-bold whitespace-nowrap transition-all ${
                selectedGroup === group.key
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-500 hover:bg-slate-100'
              }`}
            >
              {group.label} ({count})
            </button>
          )
        })}
        {hasUngrouped && (
          <button
            onClick={() => setSelectedGroup('other')}
            className={`px-3 py-1.5 rounded-lg text-xs font-bold whitespace-nowrap transition-all ${
              selectedGroup === 'other'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-slate-500 hover:bg-slate-100'
            }`}
          >
            其他 ({dictTypes.filter(t => !groupMapping[t.code]).length})
          </button>
        )}
      </div>

      {/* 列表 */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="w-8 h-8 text-slate-400 animate-spin" />
          </div>
        ) : layoutMode === 'table' ? (
          // 表格视图
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
            <table className="w-full text-left">
              <thead className="bg-slate-50 text-[10px] font-black uppercase text-slate-400">
                <tr>
                  <th className="px-4 py-4 w-12">
                    <input
                      type="checkbox"
                      checked={selectedTypeIds.size === filteredDictTypes.length && filteredDictTypes.length > 0}
                      onChange={handleToggleSelectAll}
                      className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                    />
                  </th>
                  <th className="px-4 py-4 w-12">#</th>
                  <th className="px-6 py-4">图标</th>
                  <th className="px-6 py-4">代码</th>
                  <th className="px-6 py-4">名称</th>
                  <th className="px-6 py-4">描述</th>
                  <th className="px-6 py-4 w-20 text-center">排序</th>
                  <th className="px-6 py-4 w-24 text-center">状态</th>
                  <th className="px-6 py-4 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50 font-bold text-slate-700">
                {filteredDictTypes.map((type, index) => {
                  const config = DICT_TYPE_CONFIG[type.code] || { icon: '📚', color: 'gray', description: type.description }
                  return (
                    <tr key={type.id} className={`hover:bg-slate-50/50 transition-colors ${!type.isEnabled ? 'opacity-50 grayscale' : ''}`}>
                      <td className="px-4 py-4">
                        <input
                          type="checkbox"
                          checked={selectedTypeIds.has(type.id)}
                          onChange={() => handleToggleTypeSelection(type.id)}
                          className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                      </td>
                      <td className="px-4 py-4 text-slate-400 font-mono text-xs">{index + 1}</td>
                      <td className="px-6 py-4 text-xl">{config.icon}</td>
                      <td className="px-6 py-4">
                        <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-[10px] font-mono font-black">
                          {type.code}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-slate-800">{type.name}</td>
                      <td className="px-6 py-4 text-slate-500 text-sm">{config.description || type.description || '-'}</td>
                      <td className="px-6 py-4 text-center text-slate-400 text-xs">{type.sortOrder}</td>
                      <td className="px-6 py-4 text-center">
                        <span className={`px-2 py-0.5 rounded text-[10px] font-black ${
                          type.isEnabled ? 'bg-green-50 text-green-600' : 'bg-slate-100 text-slate-400'
                        }`}>
                          {type.isEnabled ? '启用' : '停用'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right space-x-1">
                        <button
                          onClick={() => handleOpenEditTypeModal(type)}
                          className="p-2 text-slate-400 hover:text-amber-600 hover:bg-white rounded-lg transition-all shadow-sm"
                          title="编辑"
                        >
                          <Edit3 size={14} />
                        </button>
                        <button
                          onClick={() => handleSelectType(type)}
                          className="p-2 text-slate-400 hover:text-blue-600 hover:bg-white rounded-lg transition-all shadow-sm"
                          title="查看字典项"
                        >
                          <ChevronRight size={14} />
                        </button>
                        <button
                          onClick={() => handleDeleteType(type.id, type.code)}
                          className="p-2 text-slate-400 hover:text-red-500 hover:bg-white rounded-lg transition-all shadow-sm"
                          title="删除"
                        >
                          <Trash2 size={14} />
                        </button>
                      </td>
                    </tr>
                  )
                })}
                {filteredDictTypes.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-6 py-12 text-center text-slate-400">
                      <p className="text-sm">{selectedGroup === 'all' ? '暂无字典类型' : '该分组下暂无字典类型'}</p>
                      <p className="text-xs mt-1">点击"新增字典"添加数据</p>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        ) : (
          // 卡片视图
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredDictTypes.map(type => {
              const config = DICT_TYPE_CONFIG[type.code] || { icon: '📚', color: 'gray', description: type.description }

              return (
                <div
                  key={type.id}
                  className="bg-white rounded-2xl border border-slate-100 p-5 shadow-sm hover:shadow-md hover:border-blue-200 transition-all group"
                >
                  <div className="flex items-start justify-between mb-3 cursor-pointer" onClick={() => handleSelectType(type)}>
                    <div className="flex items-center gap-3">
                      <div className={`text-2xl`}>{config.icon}</div>
                      <div>
                        <h3 className="text-sm font-black text-slate-800 group-hover:text-blue-600 transition-colors">
                          {type.name}
                        </h3>
                        <p className="text-[10px] text-slate-400 font-mono">{type.code}</p>
                      </div>
                    </div>
                    <ChevronRight size={16} className="text-slate-300 group-hover:text-blue-500 transition-colors" />
                  </div>
                  <p className="text-xs text-slate-500 mb-3">{config.description}</p>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-black ${
                        type.isEnabled ? 'bg-green-50 text-green-600' : 'bg-slate-100 text-slate-400'
                      }`}>
                        {type.isEnabled ? '启用' : '停用'}
                      </span>
                      <span className="text-[10px] text-slate-300">
                        排序: {type.sortOrder}
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => handleOpenEditTypeModal(type)}
                        className="p-1.5 text-slate-400 hover:text-amber-600 hover:bg-amber-50 rounded-lg transition-all"
                        title="编辑"
                      >
                        <Edit3 size={13} />
                      </button>
                      <button
                        onClick={() => handleDeleteType(type.id, type.code)}
                        className="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all"
                        title="删除"
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
  }

  // 渲染字典项列表（树形结构）
  const renderItemsView = () => (
    <div className="flex-1 flex flex-col">
      {/* 头部 */}
      <div className="bg-white p-4 border-b border-slate-100 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-3">
          <button onClick={handleBackToTypes} className="p-2 hover:bg-slate-100 rounded-xl transition-colors">
            <ChevronRight size={20} className="text-slate-400 rotate-180" />
          </button>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xl">{DICT_TYPE_CONFIG[selectedType?.code || '']?.icon || '📚'}</span>
              <h2 className="text-lg font-black text-slate-900">{selectedType?.name}</h2>
            </div>
            <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest">
              {selectedType?.code} · {totalItemCount} 个字典项
            </p>
          </div>
        </div>
        <button onClick={() => handleOpenAddItemModal()} className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-xl text-xs font-black hover:bg-blue-700 shadow-lg shadow-blue-100 transition-all">
          <Plus size={14} /> 新增字典项
        </button>
      </div>

      {/* 树形列表 */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="w-8 h-8 text-slate-400 animate-spin" />
          </div>
        ) : (
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
            <table className="w-full text-left">
              <thead className="bg-slate-50 text-[10px] font-black uppercase text-slate-400">
                <tr>
                  <th className="px-6 py-4 w-16 text-center">#</th>
                  <th className="px-6 py-4">代码</th>
                  <th className="px-6 py-4">名称</th>
                  <th className="px-6 py-4">描述</th>
                  <th className="px-6 py-4 w-20 text-center">排序</th>
                  <th className="px-6 py-4 w-24 text-center">状态</th>
                  <th className="px-6 py-4 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50 font-bold text-slate-700">
                {dictItems.map((item, index) => {
                  const hasChildren = item.children && item.children.length > 0
                  const isExpanded = expandedItems.has(item.code)

                  return (
                    <React.Fragment key={item.id}>
                      {/* 父级行 / 独立项行 */}
                      <tr className={`hover:bg-slate-50/50 transition-colors ${!item.isEnabled ? 'opacity-50 grayscale' : ''}`}>
                        <td className="px-6 py-4 text-center text-slate-400 font-mono text-xs">{index + 1}</td>
                        <td className="px-6 py-4">
                          <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-[10px] font-mono font-black">
                            {item.code}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2">
                            {hasChildren ? (
                              <button
                                onClick={() => toggleExpand(item.code)}
                                className="p-0.5 hover:bg-slate-100 rounded transition-all"
                              >
                                {isExpanded
                                  ? <ChevronDown size={16} className="text-slate-400" />
                                  : <ChevronRight size={16} className="text-slate-400" />
                                }
                              </button>
                            ) : (
                              <span className="w-5" />
                            )}
                            <span className={`text-slate-800 ${hasChildren ? 'font-black' : ''}`}>
                              {item.name}
                            </span>
                            {hasChildren && (
                              <span className="px-1.5 py-0.5 bg-blue-50 text-blue-500 rounded-full text-[10px] font-black">
                                {item.children!.length}
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-6 py-4 text-slate-500 text-sm">{item.description || '-'}</td>
                        <td className="px-6 py-4 text-center text-slate-400 text-xs">{item.sortOrder}</td>
                        <td className="px-6 py-4 text-center">
                          <button
                            onClick={() => handleToggleEnabled(item.id)}
                            className={`p-0.5 rounded-lg transition-all ${item.isEnabled ? 'text-emerald-500 bg-emerald-50' : 'text-red-500 bg-red-50'}`}
                          >
                            {item.isEnabled ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                          </button>
                        </td>
                        <td className="px-6 py-4 text-right space-x-1">
                          {hasChildren && (
                            <button
                              onClick={() => handleOpenAddItemModal(item.code)}
                              className="p-2 text-slate-400 hover:text-green-600 hover:bg-white rounded-lg transition-all shadow-sm"
                              title="添加子项"
                            >
                              <Plus size={14} />
                            </button>
                          )}
                          <button onClick={() => handleOpenEditItemModal(item)} className="p-2 text-slate-400 hover:text-blue-600 hover:bg-white rounded-lg transition-all shadow-sm">
                            <Edit3 size={14} />
                          </button>
                          <button onClick={() => handleDeleteItem(item.id)} className="p-2 text-slate-400 hover:text-red-500 hover:bg-white rounded-lg transition-all shadow-sm">
                            <Trash2 size={14} />
                          </button>
                        </td>
                      </tr>

                      {/* 子项行 */}
                      {hasChildren && isExpanded && (
                        <>
                          {item.children!.map((child) => (
                            <tr key={child.id} className={`hover:bg-blue-50/30 transition-colors bg-slate-50/30 ${!child.isEnabled ? 'opacity-50 grayscale' : ''}`}>
                              <td className="px-6 py-3 text-center text-slate-300 font-mono text-xs">
                                <span className="text-slate-300">-</span>
                              </td>
                              <td className="px-6 py-3">
                                <span className="px-2 py-0.5 bg-blue-50 text-blue-500 rounded text-[10px] font-mono font-black">
                                  {child.code}
                                </span>
                              </td>
                              <td className="px-6 py-3">
                                <div className="flex items-center gap-2 pl-7">
                                  <span className="text-slate-300 select-none">└</span>
                                  <span className="text-slate-700">{child.name}</span>
                                </div>
                              </td>
                              <td className="px-6 py-3 text-slate-400 text-sm">{child.description || '-'}</td>
                              <td className="px-6 py-3 text-center text-slate-300 text-xs">{child.sortOrder}</td>
                              <td className="px-6 py-3 text-center">
                                <button
                                  onClick={() => handleToggleEnabled(child.id)}
                                  className={`p-0.5 rounded-lg transition-all ${child.isEnabled ? 'text-emerald-500 bg-emerald-50' : 'text-red-500 bg-red-50'}`}
                                >
                                  {child.isEnabled ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                                </button>
                              </td>
                              <td className="px-6 py-3 text-right space-x-1">
                                <button onClick={() => handleOpenEditItemModal(child)} className="p-2 text-slate-400 hover:text-blue-600 hover:bg-white rounded-lg transition-all shadow-sm">
                                  <Edit3 size={14} />
                                </button>
                                <button onClick={() => handleDeleteItem(child.id)} className="p-2 text-slate-400 hover:text-red-500 hover:bg-white rounded-lg transition-all shadow-sm">
                                  <Trash2 size={14} />
                                </button>
                              </td>
                            </tr>
                          ))}
                          {/* 添加子项快捷按钮行 */}
                          <tr className="bg-slate-50/20">
                            <td colSpan={7} className="px-6 py-2">
                              <button
                                onClick={() => handleOpenAddItemModal(item.code)}
                                className="flex items-center gap-1.5 pl-9 text-xs text-blue-500 hover:text-blue-700 font-bold transition-colors"
                              >
                                <Plus size={12} /> 添加子项
                              </button>
                            </td>
                          </tr>
                        </>
                      )}
                    </React.Fragment>
                  )
                })}
                {dictItems.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-6 py-12 text-center text-slate-400">
                      <p className="text-sm">暂无字典项</p>
                      <p className="text-xs mt-1">点击"新增字典项"添加数据</p>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )

  return (
    <div className="h-full flex flex-col bg-slate-50/50">
      {/* 隐藏的文件输入 */}
      <input
        ref={fileInputRef}
        type="file"
        accept=".json"
        onChange={handleFileChange}
        className="hidden"
      />

      {/* 顶部状态栏 */}
      <div className="px-8 py-3 bg-white border-b border-slate-100 flex items-center justify-between shrink-0 shadow-sm z-10">
        <div className="flex items-center gap-4 text-[11px] font-bold text-slate-400">
          <span className="flex items-center gap-1.5">
            <CheckCircle2 size={14} className="text-emerald-500" /> 系统正常
          </span>
          <span className="flex items-center gap-1.5">
            <Clock size={14} /> {new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })}
          </span>
        </div>
        <div className="flex items-center gap-3">
          {/* 布局切换按钮 */}
          <div className="flex items-center gap-1 bg-slate-100 rounded-xl p-1">
            <button
              onClick={() => setLayoutMode('table')}
              className={`p-2 rounded-lg transition-all ${layoutMode === 'table' ? 'bg-white shadow-sm text-blue-600' : 'text-slate-400 hover:text-slate-600'}`}
              title="表格视图"
            >
              <List size={16} />
            </button>
            <button
              onClick={() => setLayoutMode('card')}
              className={`p-2 rounded-lg transition-all ${layoutMode === 'card' ? 'bg-white shadow-sm text-blue-600' : 'text-slate-400 hover:text-slate-600'}`}
              title="卡片视图"
            >
              <LayoutGrid size={16} />
            </button>
          </div>
          {/* 分隔线 */}
          <div className="h-6 w-px bg-slate-200" />
          {/* 导入导出按钮 */}
          <button
            onClick={handleExport}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-2 bg-slate-100 text-slate-600 rounded-xl text-xs font-bold hover:bg-slate-200 transition-all disabled:opacity-50"
            title={
              view === 'items' && selectedType
                ? `导出「${selectedType.name}」`
                : view === 'types' && selectedTypeIds.size > 0
                  ? `导出已勾选 ${selectedTypeIds.size} 个字典`
                  : '导出所有字典数据'
            }
          >
            {loading ? <Loader2 size={14} className="animate-spin" /> : <Download size={14} />}
            导出{
              view === 'items' && selectedType
                ? `「${selectedType.name}」`
                : view === 'types' && selectedTypeIds.size > 0
                  ? `已选(${selectedTypeIds.size})`
                  : ''
            }
          </button>
          <button
            onClick={handleImport}
            disabled={importing}
            className="flex items-center gap-1.5 px-3 py-2 bg-slate-100 text-slate-600 rounded-xl text-xs font-bold hover:bg-slate-200 transition-all disabled:opacity-50"
            title="导入字典数据"
          >
            {importing ? <Loader2 size={14} className="animate-spin" /> : <Upload size={14} />}
            导入
          </button>
        </div>
      </div>

      {/* 主内容 */}
      {view === 'types' ? renderTypesView() : renderItemsView()}

      {/* 字典项编辑弹窗 */}
      {showItemModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col animate-scale-in">
            {/* 头部 */}
            <div className="px-8 py-4 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className={`w-10 h-10 ${isEditing ? 'bg-amber-500' : 'bg-blue-600'} rounded-xl flex items-center justify-center text-white shadow-lg`}>
                  {isEditing ? <Edit3 size={20} strokeWidth={3} /> : <Plus size={20} strokeWidth={3} />}
                </div>
                <div>
                  <h3 className="text-lg font-black text-slate-900 tracking-tight">
                    {isEditing ? '编辑字典项' : '新增字典项'}
                  </h3>
                  <p className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">
                    {selectedType?.name} - Dictionary Item
                  </p>
                </div>
              </div>
              <button onClick={() => setShowItemModal(false)} className="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-all">
                <X size={20} />
              </button>
            </div>

            {/* 表单 */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 no-scrollbar bg-white max-h-[70vh]">
              {/* 父级分类选择 */}
              <div className="bg-slate-50/50 rounded-2xl border border-slate-100 p-4 space-y-3">
                <div className="flex items-center gap-2 mb-1">
                  <div className="p-1.5 bg-white rounded-lg shadow-sm text-blue-600">
                    <LayoutGrid size={16} />
                  </div>
                  <h4 className="text-[11px] font-black text-slate-800 uppercase tracking-widest">父级分类</h4>
                </div>
                <div className="space-y-1">
                  <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
                    所属父级
                  </label>
                  <select
                    value={itemForm.parentCode}
                    onChange={(e) => setItemForm({ ...itemForm, parentCode: e.target.value })}
                    className="w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all"
                  >
                    <option value="">无（顶级项）</option>
                    {topLevelItems
                      .filter(item => item.id !== itemForm.id)
                      .map(item => (
                        <option key={item.code} value={item.code}>
                          {item.name}（{item.code}）
                        </option>
                      ))
                    }
                  </select>
                  <p className="text-[10px] text-slate-400 ml-1">
                    选择父级后，该项将作为子项显示在父级下方
                  </p>
                </div>
              </div>

              <FormSection title="基本信息" icon={BookOpen}>
                <InputField
                  label="代码"
                  required
                  placeholder="如: HD, HDF, HP..."
                  value={itemForm.code}
                  onChange={(e) => setItemForm({ ...itemForm, code: e.target.value.toUpperCase() })}
                />
                <InputField
                  label="名称"
                  required
                  placeholder="如: 血液透析"
                  value={itemForm.name}
                  onChange={(e) => setItemForm({ ...itemForm, name: e.target.value })}
                />
                <InputField
                  label="描述"
                  placeholder="如: Hemodialysis"
                  value={itemForm.description}
                  onChange={(e) => setItemForm({ ...itemForm, description: e.target.value })}
                />
                <InputField
                  label="排序"
                  type="number"
                  value={itemForm.sortOrder}
                  onChange={(e) => setItemForm({ ...itemForm, sortOrder: parseInt(e.target.value) || 0 })}
                />
              </FormSection>

              <div className="flex items-center justify-between p-4 bg-slate-50/50 rounded-2xl border border-slate-100">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${itemForm.isEnabled ? 'bg-emerald-100 text-emerald-600' : 'bg-red-100 text-red-600'}`}>
                    {itemForm.isEnabled ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                  </div>
                  <div>
                    <p className="text-xs font-black text-slate-800">启用状态</p>
                    <p className="text-[10px] text-slate-400 font-bold uppercase">
                      {itemForm.isEnabled ? '已启用' : '已停用'}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => setItemForm({ ...itemForm, isEnabled: !itemForm.isEnabled })}
                  className={`px-4 py-2 rounded-xl text-xs font-black transition-all ${
                    itemForm.isEnabled
                      ? 'bg-emerald-50 text-emerald-600 hover:bg-emerald-100'
                      : 'bg-red-50 text-red-600 hover:bg-red-100'
                  }`}
                >
                  {itemForm.isEnabled ? '已启用' : '已停用'}
                </button>
              </div>
            </div>

            {/* 底部按钮 */}
            <div className="px-8 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3 shrink-0">
              <button onClick={() => setShowItemModal(false)} className="px-6 py-2 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-100 transition-all">
                取消
              </button>
              <button onClick={handleSaveItem} disabled={loading} className={`px-10 py-2 ${isEditing ? 'bg-amber-500 hover:bg-amber-600 shadow-amber-100' : 'bg-blue-600 hover:bg-blue-700 shadow-blue-100'} text-white rounded-xl text-xs font-black shadow-xl transition-all disabled:opacity-50`}>
                {loading ? <Loader2 size={16} className="animate-spin mr-2 inline" /> : ''}保存
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 字典类型编辑弹窗 */}
      {showTypeModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col animate-scale-in">
            {/* 头部 */}
            <div className="px-8 py-4 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className={`w-10 h-10 ${isEditingType ? 'bg-amber-500' : 'bg-blue-600'} rounded-xl flex items-center justify-center text-white shadow-lg`}>
                  {isEditingType ? <Edit3 size={20} strokeWidth={3} /> : <Plus size={20} strokeWidth={3} />}
                </div>
                <div>
                  <h3 className="text-lg font-black text-slate-900 tracking-tight">
                    {isEditingType ? '编辑字典类型' : '新增字典类型'}
                  </h3>
                  <p className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">
                    Dictionary Type
                  </p>
                </div>
              </div>
              <button onClick={() => setShowTypeModal(false)} className="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-all">
                <X size={20} />
              </button>
            </div>

            {/* 表单 */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 no-scrollbar bg-white max-h-[70vh]">
              {/* 图标选择 */}
              <div className="bg-slate-50/50 rounded-2xl border border-slate-100 p-4 space-y-3">
                <div className="flex items-center gap-2 mb-1">
                  <div className="p-1.5 bg-white rounded-lg shadow-sm text-blue-600">
                    <BookOpen size={16} />
                  </div>
                  <h4 className="text-[11px] font-black text-slate-800 uppercase tracking-widest">选择图标</h4>
                </div>
                <div className="flex items-center gap-4">
                  <div className="w-16 h-16 bg-white rounded-2xl border-2 border-blue-200 flex items-center justify-center text-4xl shadow-sm">
                    {typeForm.icon}
                  </div>
                  <div className="flex-1 grid grid-cols-10 gap-1">
                    {ICON_OPTIONS.map((icon) => (
                      <button
                        key={icon}
                        type="button"
                        onClick={() => setTypeForm({ ...typeForm, icon })}
                        className={`w-8 h-8 rounded-lg flex items-center justify-center text-lg transition-all ${
                          typeForm.icon === icon
                            ? 'bg-blue-100 ring-2 ring-blue-500'
                            : 'bg-white hover:bg-slate-100'
                        }`}
                      >
                        {icon}
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              <FormSection title="基本信息" icon={BookOpen}>
                <InputField
                  label="代码"
                  required
                  placeholder="如: DIALYSIS_MODE, ANTICOAGULANT..."
                  value={typeForm.code}
                  onChange={(e) => setTypeForm({ ...typeForm, code: e.target.value.toUpperCase() })}
                />
                <InputField
                  label="名称"
                  required
                  placeholder="如: 透析方式"
                  value={typeForm.name}
                  onChange={(e) => setTypeForm({ ...typeForm, name: e.target.value })}
                />
                <InputField
                  label="描述"
                  placeholder="如: 血液透析治疗方式"
                  value={typeForm.description}
                  onChange={(e) => setTypeForm({ ...typeForm, description: e.target.value })}
                />
                <InputField
                  label="排序"
                  type="number"
                  value={typeForm.sortOrder}
                  onChange={(e) => setTypeForm({ ...typeForm, sortOrder: parseInt(e.target.value) || 0 })}
                />
              </FormSection>

              {/* 所属分组 */}
              <div className="bg-slate-50/50 rounded-2xl border border-slate-100 p-4 space-y-3">
                <div className="flex items-center gap-2 mb-1">
                  <div className="p-1.5 bg-white rounded-lg shadow-sm text-blue-600">
                    <LayoutGrid size={16} />
                  </div>
                  <h4 className="text-[11px] font-black text-slate-800 uppercase tracking-widest">所属分组</h4>
                </div>
                <div className="flex flex-wrap gap-2">
                  {DICT_GROUPS.map(g => (
                    <button
                      key={g.key}
                      type="button"
                      onClick={() => setTypeForm({ ...typeForm, group: g.key })}
                      className={`px-3 py-1.5 rounded-lg text-xs font-bold transition-all ${
                        typeForm.group === g.key
                          ? 'bg-blue-600 text-white shadow-sm'
                          : 'bg-white text-slate-500 border border-slate-200 hover:bg-slate-50'
                      }`}
                    >
                      {g.label}
                    </button>
                  ))}
                  <button
                    type="button"
                    onClick={() => setTypeForm({ ...typeForm, group: '' })}
                    className={`px-3 py-1.5 rounded-lg text-xs font-bold transition-all ${
                      !typeForm.group
                        ? 'bg-slate-600 text-white shadow-sm'
                        : 'bg-white text-slate-500 border border-slate-200 hover:bg-slate-50'
                    }`}
                  >
                    其他
                  </button>
                </div>
              </div>

              <div className="flex items-center justify-between p-4 bg-slate-50/50 rounded-2xl border border-slate-100">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${typeForm.isEnabled ? 'bg-emerald-100 text-emerald-600' : 'bg-red-100 text-red-600'}`}>
                    {typeForm.isEnabled ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                  </div>
                  <div>
                    <p className="text-xs font-black text-slate-800">启用状态</p>
                    <p className="text-[10px] text-slate-400 font-bold uppercase">
                      {typeForm.isEnabled ? '已启用' : '已停用'}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => setTypeForm({ ...typeForm, isEnabled: !typeForm.isEnabled })}
                  className={`px-4 py-2 rounded-xl text-xs font-black transition-all ${
                    typeForm.isEnabled
                      ? 'bg-emerald-50 text-emerald-600 hover:bg-emerald-100'
                      : 'bg-red-50 text-red-600 hover:bg-red-100'
                  }`}
                >
                  {typeForm.isEnabled ? '已启用' : '已停用'}
                </button>
              </div>
            </div>

            {/* 底部按钮 */}
            <div className="px-8 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3 shrink-0">
              <button onClick={() => setShowTypeModal(false)} className="px-6 py-2 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-100 transition-all">
                取消
              </button>
              <button onClick={handleSaveType} disabled={loading} className={`px-10 py-2 ${isEditingType ? 'bg-amber-500 hover:bg-amber-600 shadow-amber-100' : 'bg-blue-600 hover:bg-blue-700 shadow-blue-100'} text-white rounded-xl text-xs font-black shadow-xl transition-all disabled:opacity-50`}>
                {loading ? <Loader2 size={16} className="animate-spin mr-2 inline" /> : ''}保存
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 导入预览弹窗 */}
      {importPreview && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md p-6 animate-fade-in">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-lg overflow-hidden flex flex-col animate-scale-in">
            <div className="px-8 py-4 border-b border-slate-100 flex justify-between items-center bg-white shrink-0">
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center text-white shadow-lg">
                  <Upload size={20} strokeWidth={3} />
                </div>
                <div>
                  <h3 className="text-lg font-black text-slate-900 tracking-tight">导入字典数据</h3>
                  <p className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">
                    选择要导入的字典类型
                  </p>
                </div>
              </div>
              <button onClick={() => setImportPreview(null)} className="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-all">
                <X size={20} />
              </button>
            </div>

            <div className="p-6 space-y-4 max-h-[60vh] overflow-y-auto">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-500">
                  文件中包含 {importPreview.data.types.length} 个字典类型，{(importPreview.data.items || []).length} 条字典项
                </span>
                <button
                  onClick={() => {
                    const allSelected =
                      importPreview.data.types.length > 0 &&
                      importPreview.selectedCodes.size === importPreview.data.types.length
                    setImportPreview({
                      ...importPreview,
                      selectedCodes: allSelected ? new Set() : new Set(importPreview.data.types.map(t => t.code))
                    })
                  }}
                  className="text-xs font-bold text-blue-600 hover:text-blue-700"
                >
                  {importPreview.data.types.length > 0 && importPreview.selectedCodes.size === importPreview.data.types.length ? '取消全选' : '全选'}
                </button>
              </div>

              <div className="space-y-1.5">
                {importPreview.data.types.map((type: DictType) => {
                  const itemCount = (importPreview.data.items || []).filter((i: DictItem) => i.typeCode === type.code).length
                  const isSelected = importPreview.selectedCodes.has(type.code)
                  return (
                    <label
                      key={type.code}
                      className={`flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer transition-all ${
                        isSelected ? 'bg-blue-50 border border-blue-200' : 'bg-slate-50 border border-slate-100 hover:bg-slate-100'
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => {
                          const next = new Set(importPreview.selectedCodes)
                          if (isSelected) {
                            next.delete(type.code)
                          } else {
                            next.add(type.code)
                          }
                          setImportPreview({ ...importPreview, selectedCodes: next })
                        }}
                        className="w-4 h-4 rounded accent-blue-600"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-bold text-slate-800 truncate">{type.name}</div>
                        <div className="text-[10px] text-slate-400 font-medium">{type.code}</div>
                      </div>
                      <span className="text-[10px] font-bold text-slate-400 bg-white px-2 py-0.5 rounded-full">
                        {itemCount} 项
                      </span>
                    </label>
                  )
                })}
              </div>

              <p className="text-[10px] text-slate-400 px-1">
                相同代码的数据将被更新，不同代码的数据将新增
              </p>
            </div>

            <div className="px-8 py-4 bg-slate-50 border-t border-slate-100 flex justify-between items-center shrink-0">
              <span className="text-xs font-bold text-slate-500">
                已选 {importPreview.selectedCodes.size} / {importPreview.data.types.length} 个类型
              </span>
              <div className="flex gap-3">
                <button onClick={() => setImportPreview(null)} className="px-6 py-2 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-100 transition-all">
                  取消
                </button>
                <button
                  onClick={handleConfirmImport}
                  disabled={importing || importPreview.selectedCodes.size === 0}
                  className="px-10 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-black shadow-xl shadow-blue-100 transition-all disabled:opacity-50 flex items-center gap-2"
                >
                  {importing && <Loader2 size={14} className="animate-spin" />}
                  确认导入
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
