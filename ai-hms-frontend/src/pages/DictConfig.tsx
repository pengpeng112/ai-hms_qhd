import { useEffect, useMemo, useState } from 'react'
import { message } from 'antd'
import {
  BookOpen,
  ChevronDown,
  ChevronRight,
  Edit3,
  Loader2,
  Plus,
  RefreshCw,
  ToggleLeft,
  ToggleRight,
  Trash2,
} from 'lucide-react'
import { dictApi, dictCache, type DictItem, type DictType } from '@/services/dictApi'
import { getErrorMessage } from '@/services/restClient'

type CategoryKey =
  | 'dialysis'
  | 'vascular'
  | 'order'
  | 'staff'
  | 'clinical'
  | 'outcome'
  | 'other'

type FlatDictItem = DictItem & {
  level: number
}

function isLegacySource(source?: string) {
  return source === 'legacy'
}

function SourceTag({ source }: { source?: string }) {
  if (isLegacySource(source)) {
    return (
      <span
        title="老库来源，当前仅支持只读查看"
        className="rounded-full bg-amber-50 px-2 py-0.5 text-[11px] font-bold text-amber-700"
      >
        老库来源
      </span>
    )
  }

  return (
    <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-bold text-emerald-700">本地维护</span>
  )
}

type ItemFormState = {
  id: string
  typeCode: string
  code: string
  name: string
  description: string
  sortOrder: number
  isEnabled: boolean
  parentCode: string
}

const CATEGORY_META: Record<CategoryKey, { label: string; description: string }> = {
  dialysis: { label: '透析治疗', description: '直接维护透析治疗相关字典值' },
  vascular: { label: '血管通路', description: '直接维护血管通路相关字典值' },
  order: { label: '医嘱处方', description: '直接维护医嘱、处方、药品材料相关字典值' },
  staff: { label: '人员信息', description: '直接维护人员及基础主数据字典值' },
  clinical: { label: '临床诊疗', description: '直接维护临床诊疗相关字典值' },
  outcome: { label: '转归记录', description: '直接维护转归记录相关字典值' },
  other: { label: '其他字典', description: '先选择具体字典类型，再维护字典值' },
}

const TYPE_CATEGORY_MAP: Record<string, CategoryKey> = {
  DIALYSIS_MODE: 'dialysis',
  ANTICOAGULANT: 'dialysis',
  DIALYSATE_TYPE: 'dialysis',
  DIALYSATE_GROUP: 'dialysis',
  DIALYSATE_FLOW: 'dialysis',
  GLUCOSE: 'dialysis',

  VASCULAR_ACCESS: 'vascular',
  VASCULAR_SITE: 'vascular',
  VEIN_TYPE: 'vascular',
  ARTERY_TYPE: 'vascular',
  SURGERY_TYPE: 'vascular',

  ORDER_TYPE: 'order',
  ORDER_CATEGORY: 'order',
  ORDER_ROUTE: 'order',
  ORDER_FREQUENCY: 'order',
  ORDER_TIMING: 'order',
  MATERIAL_CATEGORY: 'order',
  DRUG_CATEGORY: 'order',

  DOCTOR: 'staff',
  DOCTOR_LIST: 'staff',
  DOCTOR_INFO: 'staff',
  Doctor: 'staff',
  HOSPITAL: 'staff',
  Hospital: 'staff',

  PRIMARY_DISEASE: 'clinical',
  COMPLICATION: 'clinical',
  PATHOLOGY: 'clinical',
  TUMOR: 'clinical',
  ALLERGEN: 'clinical',
  INSURANCE_TYPE: 'clinical',
  PATIENT_TYPE: 'clinical',
  ID_TYPE: 'clinical',
  VISIT_CATEGORY: 'clinical',
  BLOOD_TYPE_ABO: 'clinical',
  BLOOD_TYPE_RH: 'clinical',
  EDUCATION_LEVEL: 'clinical',
  MARITAL_STATUS: 'clinical',

  OUTCOME: 'outcome',
}

const EMPTY_ITEM_FORM: ItemFormState = {
  id: '',
  typeCode: '',
  code: '',
  name: '',
  description: '',
  sortOrder: 1,
  isEnabled: true,
  parentCode: '',
}

function resolveCategory(typeCode: string): CategoryKey {
  return TYPE_CATEGORY_MAP[typeCode] ?? 'other'
}

function flattenItems(items: DictItem[], level = 0): FlatDictItem[] {
  return items.flatMap((item) => {
    const current: FlatDictItem = { ...item, level }
    const children = Array.isArray(item.children) ? flattenItems(item.children, level + 1) : []
    return [current, ...children]
  })
}

function getMaxSortOrder(items: DictItem[]): number {
  const flat = flattenItems(items)
  return flat.length > 0 ? Math.max(...flat.map((item) => item.sortOrder || 0)) : 0
}

function getAllParentOptions(items: DictItem[]): Array<{ value: string; label: string }> {
  return flattenItems(items).map((item) => ({
    value: item.code,
    label: `${'　'.repeat(item.level)}${item.name} (${item.code})`,
  }))
}

function getItemById(items: DictItem[], id: string): DictItem | null {
  const flat = flattenItems(items)
  return flat.find((item) => item.id === id) ?? null
}

function getDescendantIds(items: DictItem[], itemId: string): Set<string> {
  const target = getItemById(items, itemId)
  if (!target) return new Set()

  const targetRoot = findTreeNode(items, itemId)
  if (!targetRoot) return new Set()

  const ids = new Set<string>()
  const walk = (node: DictItem) => {
    ids.add(node.id)
    for (const child of node.children || []) {
      walk(child)
    }
  }
  walk(targetRoot)
  return ids
}

function findTreeNode(items: DictItem[], id: string): DictItem | null {
  for (const item of items) {
    if (item.id === id) return item
    if (item.children?.length) {
      const child = findTreeNode(item.children, id)
      if (child) return child
    }
  }
  return null
}

function TypeValueTable({
  type,
  items,
  loading,
  onAdd,
  onEdit,
  onToggle,
  onDelete,
}: {
  type: DictType
  items: DictItem[]
  loading: boolean
  onAdd: (typeCode: string) => void
  onEdit: (typeCode: string, itemId: string) => void
  onToggle: (typeCode: string, itemId: string) => void
  onDelete: (typeCode: string, itemId: string) => void
}) {
  const flatItems = useMemo(() => flattenItems(items), [items])

  return (
    <section className="rounded-3xl border border-slate-200 bg-white shadow-sm">
      <div className="flex flex-col gap-3 border-b border-slate-100 px-5 py-4 md:flex-row md:items-center md:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h3 className="text-base font-black text-slate-900">{type.name}</h3>
            <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-bold text-slate-500">
              {type.code}
            </span>
            <SourceTag source={type.source} />
            <span className="rounded-full bg-blue-50 px-2 py-0.5 text-[11px] font-bold text-blue-600">
              {flatItems.length} 项
            </span>
          </div>
          <p className="mt-1 text-sm text-slate-500">{type.description || '维护该字典类型下的编码与名称。'}</p>
        </div>
        <button
          type="button"
          onClick={() => onAdd(type.code)}
          className="inline-flex items-center justify-center gap-2 rounded-2xl bg-blue-600 px-4 py-2 text-sm font-bold text-white transition hover:bg-blue-700"
        >
          <Plus size={16} />
          新增字典值
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead className="bg-slate-50">
            <tr className="text-left text-xs uppercase tracking-wide text-slate-500">
              <th className="px-5 py-3 font-bold">编码</th>
              <th className="px-5 py-3 font-bold">名称</th>
              <th className="px-5 py-3 font-bold">描述</th>
              <th className="px-5 py-3 font-bold">上级编码</th>
              <th className="px-5 py-3 font-bold">排序</th>
              <th className="px-5 py-3 font-bold">来源</th>
              <th className="px-5 py-3 font-bold">状态</th>
              <th className="px-5 py-3 font-bold text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={8} className="px-5 py-10 text-center text-slate-500">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 size={16} className="animate-spin" />
                    加载中
                  </span>
                </td>
              </tr>
            ) : flatItems.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-5 py-10 text-center text-slate-400">
                  暂无字典值
                </td>
              </tr>
            ) : (
              flatItems.map((item) => (
                <tr key={item.id} className={`border-t border-slate-100 ${isLegacySource(item.source) ? 'bg-amber-50/30' : ''}`}>
                  <td className="px-5 py-3 font-mono text-xs font-bold text-slate-700">
                    <span style={{ paddingLeft: `${item.level * 20}px` }} className="inline-flex items-center gap-2">
                      {item.level > 0 ? <span className="text-slate-300">└</span> : null}
                      {item.code}
                    </span>
                  </td>
                  <td className="px-5 py-3 font-semibold text-slate-900">{item.name}</td>
                  <td className="px-5 py-3 text-slate-500">{item.description || '-'}</td>
                  <td className="px-5 py-3 text-slate-500">{item.parentCode || '-'}</td>
                  <td className="px-5 py-3 text-slate-500">{item.sortOrder ?? '-'}</td>
                  <td className="px-5 py-3">
                    <SourceTag source={item.source} />
                  </td>
                  <td className="px-5 py-3">
                    <span
                      className={`rounded-full px-2 py-1 text-xs font-bold ${
                        item.isEnabled ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500'
                      }`}
                    >
                      {item.isEnabled ? '启用' : '停用'}
                    </span>
                  </td>
                  <td className="px-5 py-3">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => onEdit(type.code, item.id)}
                        disabled={isLegacySource(item.source)}
                        className="rounded-xl p-2 text-slate-500 transition hover:bg-blue-50 hover:text-blue-600 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
                        title={isLegacySource(item.source) ? '老库字典暂不支持直接维护' : '编辑'}
                      >
                        <Edit3 size={15} />
                      </button>
                      <button
                        type="button"
                        onClick={() => onToggle(type.code, item.id)}
                        disabled={isLegacySource(item.source)}
                        className="rounded-xl p-2 text-slate-500 transition hover:bg-amber-50 hover:text-amber-600 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
                        title={isLegacySource(item.source) ? '老库字典暂不支持直接维护' : '启用/停用'}
                      >
                        {item.isEnabled ? <ToggleRight size={17} /> : <ToggleLeft size={17} />}
                      </button>
                      <button
                        type="button"
                        onClick={() => onDelete(type.code, item.id)}
                        disabled={isLegacySource(item.source)}
                        className="rounded-xl p-2 text-slate-500 transition hover:bg-red-50 hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
                        title={isLegacySource(item.source) ? '老库字典暂不支持直接维护' : '删除'}
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </section>
  )
}

export default function DictConfig() {
  const [dictTypes, setDictTypes] = useState<DictType[]>([])
  const [itemsByType, setItemsByType] = useState<Record<string, DictItem[]>>({})
  const [loadingTypes, setLoadingTypes] = useState(false)
  const [loadingTypeCodes, setLoadingTypeCodes] = useState<Set<string>>(new Set())
  const [selectedCategory, setSelectedCategory] = useState<CategoryKey>('dialysis')
  const [selectedOtherTypeCode, setSelectedOtherTypeCode] = useState<string>('')
  const [otherExpanded, setOtherExpanded] = useState(true)

  const [showItemModal, setShowItemModal] = useState(false)
  const [itemSaving, setItemSaving] = useState(false)
  const [itemForm, setItemForm] = useState<ItemFormState>(EMPTY_ITEM_FORM)

  const sortedTypes = useMemo(
    () =>
      [...dictTypes].sort((a, b) => {
        if (a.sortOrder === b.sortOrder) return a.code.localeCompare(b.code)
        return a.sortOrder - b.sortOrder
      }),
    [dictTypes]
  )

  const typesByCategory = useMemo(() => {
    const grouped: Record<CategoryKey, DictType[]> = {
      dialysis: [],
      vascular: [],
      order: [],
      staff: [],
      clinical: [],
      outcome: [],
      other: [],
    }

    for (const type of sortedTypes) {
      grouped[resolveCategory(type.code)].push(type)
    }

    return grouped
  }, [sortedTypes])

  const selectedOtherType = useMemo(
    () => typesByCategory.other.find((item) => item.code === selectedOtherTypeCode) ?? null,
    [selectedOtherTypeCode, typesByCategory.other]
  )

  const visibleTypes = useMemo(() => {
    if (selectedCategory === 'other') {
      return selectedOtherType ? [selectedOtherType] : []
    }
    return typesByCategory[selectedCategory]
  }, [selectedCategory, selectedOtherType, typesByCategory])

  const loadTypes = async () => {
    setLoadingTypes(true)
    try {
      const types = await dictApi.listTypes()
      setDictTypes(Array.isArray(types) ? types : [])
    } catch (error) {
      console.error('加载字典类型失败:', error)
      message.error(getErrorMessage(error))
    } finally {
      setLoadingTypes(false)
    }
  }

  const loadItems = async (typeCode: string, forceRefresh = false) => {
    setLoadingTypeCodes((prev) => new Set(prev).add(typeCode))
    try {
      if (forceRefresh) {
        dictCache.clear(typeCode)
      }
      const tree = await dictApi.getItemsTree(typeCode, false)
      setItemsByType((prev) => ({ ...prev, [typeCode]: Array.isArray(tree) ? tree : [] }))
    } catch (error) {
      console.error(`加载字典项失败: ${typeCode}`, error)
      setItemsByType((prev) => ({ ...prev, [typeCode]: [] }))
      message.error(getErrorMessage(error))
    } finally {
      setLoadingTypeCodes((prev) => {
        const next = new Set(prev)
        next.delete(typeCode)
        return next
      })
    }
  }

  useEffect(() => {
    loadTypes()
  }, [])

  useEffect(() => {
    const ordinaryCategories: CategoryKey[] = ['dialysis', 'vascular', 'order', 'staff', 'clinical', 'outcome']
    if (selectedCategory !== 'other' && typesByCategory[selectedCategory].length === 0) {
      const fallback = ordinaryCategories.find((key) => typesByCategory[key].length > 0)
      if (fallback) {
        setSelectedCategory(fallback)
        return
      }
      if (typesByCategory.other.length > 0) {
        setSelectedCategory('other')
      }
    }
  }, [selectedCategory, typesByCategory])

  useEffect(() => {
    if (selectedCategory === 'other') {
      if (!selectedOtherTypeCode && typesByCategory.other.length > 0) {
        setSelectedOtherTypeCode(typesByCategory.other[0].code)
      }
      return
    }

    for (const type of typesByCategory[selectedCategory]) {
      if (!itemsByType[type.code]) {
        void loadItems(type.code)
      }
    }
  }, [itemsByType, selectedCategory, selectedOtherTypeCode, typesByCategory])

  useEffect(() => {
    if (selectedCategory === 'other' && selectedOtherTypeCode && !itemsByType[selectedOtherTypeCode]) {
      void loadItems(selectedOtherTypeCode)
    }
  }, [itemsByType, selectedCategory, selectedOtherTypeCode])

  const openAddModal = (typeCode: string) => {
    const items = itemsByType[typeCode] || []
    setItemForm({
      ...EMPTY_ITEM_FORM,
      typeCode,
      sortOrder: getMaxSortOrder(items) + 1,
    })
    setShowItemModal(true)
  }

  const openEditModal = (typeCode: string, itemId: string) => {
    const item = getItemById(itemsByType[typeCode] || [], itemId)
    if (!item) return

    setItemForm({
      id: item.id,
      typeCode,
      code: item.code,
      name: item.name,
      description: item.description || '',
      sortOrder: item.sortOrder || 0,
      isEnabled: item.isEnabled,
      parentCode: item.parentCode || '',
    })
    setShowItemModal(true)
  }

  const handleSaveItem = async () => {
    if (!itemForm.typeCode || !itemForm.code.trim() || !itemForm.name.trim()) {
      message.error('请填写字典编码和名称')
      return
    }

    setItemSaving(true)
    try {
      if (itemForm.id) {
        await dictApi.updateItem(itemForm.id, {
          code: itemForm.code.trim(),
          name: itemForm.name.trim(),
          description: itemForm.description.trim(),
          sortOrder: Number(itemForm.sortOrder) || 0,
          isEnabled: itemForm.isEnabled,
          parent_code: itemForm.parentCode || null,
        })
      } else {
        await dictApi.createItem({
          typeCode: itemForm.typeCode,
          code: itemForm.code.trim(),
          name: itemForm.name.trim(),
          description: itemForm.description.trim(),
          sortOrder: Number(itemForm.sortOrder) || 0,
          isEnabled: itemForm.isEnabled,
          parentCode: itemForm.parentCode || undefined,
        })
      }

      dictCache.clear(itemForm.typeCode)
      await loadItems(itemForm.typeCode, true)
      setShowItemModal(false)
      setItemForm(EMPTY_ITEM_FORM)
    } catch (error) {
      console.error('保存字典项失败:', error)
      message.error(getErrorMessage(error))
    } finally {
      setItemSaving(false)
    }
  }

  const handleDeleteItem = async (typeCode: string, itemId: string) => {
    const items = itemsByType[typeCode] || []
    const item = getItemById(items, itemId)
    if (!item) return

    const descendants = getDescendantIds(items, itemId)
    const descendantCount = Math.max(descendants.size - 1, 0)
    const confirmMessage =
      descendantCount > 0
        ? `确定删除“${item.name}”吗？该项下还有 ${descendantCount} 个子项会一起删除。`
        : `确定删除“${item.name}”吗？`

    if (!window.confirm(confirmMessage)) return

    try {
      await dictApi.deleteItem(itemId)
      dictCache.clear(typeCode)
      await loadItems(typeCode, true)
    } catch (error) {
      console.error('删除字典项失败:', error)
      message.error(getErrorMessage(error))
    }
  }

  const handleToggleItem = async (typeCode: string, itemId: string) => {
    try {
      await dictApi.toggleItemEnabled(itemId)
      dictCache.clear(typeCode)
      await loadItems(typeCode, true)
    } catch (error) {
      console.error('切换字典项状态失败:', error)
      message.error(getErrorMessage(error))
    }
  }

  const handleRefreshVisible = async () => {
    if (selectedCategory === 'other') {
      if (selectedOtherTypeCode) {
        await loadItems(selectedOtherTypeCode, true)
      }
      return
    }

    await Promise.all(visibleTypes.map((type) => loadItems(type.code, true)))
  }

  const parentOptions = useMemo(() => {
    const typeCode = itemForm.typeCode
    const items = itemsByType[typeCode] || []
    const options = getAllParentOptions(items)
    if (!itemForm.id) return options

    const blockedIds = getDescendantIds(items, itemForm.id)
    return options.filter((option) => {
      const node = flattenItems(items).find((item) => item.code === option.value)
      return node ? !blockedIds.has(node.id) : true
    })
  }, [itemForm.id, itemForm.typeCode, itemsByType])

  return (
    <div className="space-y-6">
      <section className="rounded-[32px] border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className="text-xs font-black uppercase tracking-[0.28em] text-blue-600">Dictionary Config</p>
            <h1 className="mt-2 text-3xl font-black tracking-tight text-slate-900">字典配置</h1>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-500">
              左侧按业务分类展示字典入口。普通业务字典在右侧直接维护字典值；“其他字典”先选具体类型，再维护编码和名称。
            </p>
          </div>
          <button
            type="button"
            onClick={() => {
              void loadTypes()
              void handleRefreshVisible()
            }}
            className="inline-flex items-center justify-center gap-2 rounded-2xl border border-slate-200 px-4 py-2 text-sm font-bold text-slate-700 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700"
          >
            <RefreshCw size={16} />
            刷新
          </button>
        </div>
      </section>

      <section className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <aside className="rounded-[32px] border border-slate-200 bg-white p-4 shadow-sm">
          <div className="mb-4 flex items-center gap-3 px-2">
            <div className="rounded-2xl bg-blue-50 p-3 text-blue-600">
              <BookOpen size={20} />
            </div>
            <div>
              <h2 className="text-lg font-black text-slate-900">字典分类</h2>
              <p className="text-xs text-slate-500">按业务结构组织后端字典类型</p>
            </div>
          </div>

          <div className="space-y-2">
            {(['dialysis', 'vascular', 'order', 'staff', 'clinical', 'outcome'] as CategoryKey[]).map((key) => {
              const types = typesByCategory[key]
              const active = selectedCategory === key
              return (
                <button
                  key={key}
                  type="button"
                  onClick={() => setSelectedCategory(key)}
                  className={`w-full rounded-2xl border px-4 py-3 text-left transition ${
                    active
                      ? 'border-blue-200 bg-blue-50'
                      : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50'
                  }`}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <div className="text-sm font-black text-slate-900">{CATEGORY_META[key].label}</div>
                      <div className="mt-1 text-xs text-slate-500">{CATEGORY_META[key].description}</div>
                    </div>
                    <span className="rounded-full bg-white px-2 py-1 text-xs font-bold text-slate-500">
                      {types.length}
                    </span>
                  </div>
                </button>
              )
            })}

            <div className="rounded-2xl border border-slate-200 bg-white">
              <button
                type="button"
                onClick={() => {
                  setSelectedCategory('other')
                  setOtherExpanded((prev) => !prev)
                }}
                className={`flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition ${
                  selectedCategory === 'other' ? 'bg-blue-50' : 'hover:bg-slate-50'
                }`}
              >
                <div>
                  <div className="text-sm font-black text-slate-900">{CATEGORY_META.other.label}</div>
                  <div className="mt-1 text-xs text-slate-500">{CATEGORY_META.other.description}</div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="rounded-full bg-slate-100 px-2 py-1 text-xs font-bold text-slate-500">
                    {typesByCategory.other.length}
                  </span>
                  {otherExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                </div>
              </button>

              {otherExpanded ? (
                <div className="border-t border-slate-100 px-2 py-2">
                  {loadingTypes ? (
                    <div className="flex items-center gap-2 px-3 py-4 text-sm text-slate-500">
                      <Loader2 size={14} className="animate-spin" />
                      加载类型中
                    </div>
                  ) : typesByCategory.other.length === 0 ? (
                    <div className="px-3 py-4 text-sm text-slate-400">暂无其他字典类型</div>
                  ) : (
                    <div className="space-y-1">
                      {typesByCategory.other.map((type) => (
                        <button
                          key={type.code}
                          type="button"
                          onClick={() => {
                            setSelectedCategory('other')
                            setSelectedOtherTypeCode(type.code)
                          }}
                          className={`w-full rounded-xl px-3 py-2 text-left text-sm transition ${
                            selectedCategory === 'other' && selectedOtherTypeCode === type.code
                              ? 'bg-slate-900 text-white'
                              : 'text-slate-700 hover:bg-slate-50'
                          }`}
                        >
                          <div className="font-bold">{type.name}</div>
                          <div
                            className={`mt-0.5 text-xs ${
                              selectedCategory === 'other' && selectedOtherTypeCode === type.code
                                ? 'text-slate-300'
                                : 'text-slate-400'
                            }`}
                          >
                            {type.code}
                          </div>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              ) : null}
            </div>
          </div>
        </aside>

        <main className="space-y-4">
          <section className="rounded-[32px] border border-slate-200 bg-white px-6 py-5 shadow-sm">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div>
                <p className="text-xs font-black uppercase tracking-[0.24em] text-slate-400">当前区域</p>
                <h2 className="mt-1 text-2xl font-black text-slate-900">
                  {selectedCategory === 'other'
                    ? selectedOtherType?.name || '其他字典'
                    : CATEGORY_META[selectedCategory].label}
                </h2>
                <p className="mt-1 text-sm text-slate-500">
                  {selectedCategory === 'other'
                    ? selectedOtherType
                      ? `当前正在维护 ${selectedOtherType.code} 的字典值。`
                      : '请先从左侧选择一个具体字典类型。'
                    : CATEGORY_META[selectedCategory].description}
                </p>
              </div>
              <div className="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-500">
                {selectedCategory === 'other'
                  ? `类型数 ${typesByCategory.other.length}`
                  : `包含 ${visibleTypes.length} 个字典类型`}
              </div>
            </div>
          </section>

          {selectedCategory === 'other' && !selectedOtherType ? (
            <section className="rounded-[32px] border border-dashed border-slate-300 bg-white px-6 py-16 text-center text-slate-400">
              请先从左侧“其他字典”中选择具体字典类型。
            </section>
          ) : visibleTypes.length === 0 ? (
            <section className="rounded-[32px] border border-dashed border-slate-300 bg-white px-6 py-16 text-center text-slate-400">
              当前分类下暂无字典类型。
            </section>
          ) : (
            visibleTypes.map((type) => (
              <TypeValueTable
                key={type.code}
                type={type}
                items={itemsByType[type.code] || []}
                loading={loadingTypeCodes.has(type.code)}
                onAdd={openAddModal}
                onEdit={openEditModal}
                onToggle={handleToggleItem}
                onDelete={handleDeleteItem}
              />
            ))
          )}
        </main>
      </section>

      {showItemModal ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/45 px-4">
          <div className="w-full max-w-2xl rounded-[28px] bg-white shadow-2xl">
            <div className="border-b border-slate-100 px-6 py-5">
              <h3 className="text-xl font-black text-slate-900">{itemForm.id ? '编辑字典值' : '新增字典值'}</h3>
              <p className="mt-1 text-sm text-slate-500">维护编码、名称及层级信息，保存后会立即刷新当前字典。</p>
            </div>

            <div className="grid gap-4 px-6 py-5 md:grid-cols-2">
              <label className="space-y-2">
                <span className="text-xs font-black uppercase tracking-wide text-slate-500">字典类型</span>
                <input
                  value={itemForm.typeCode}
                  disabled
                  className="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-slate-600 outline-none"
                />
              </label>
              <label className="space-y-2">
                <span className="text-xs font-black uppercase tracking-wide text-slate-500">上级编码</span>
                <select
                  value={itemForm.parentCode}
                  onChange={(event) => setItemForm((prev) => ({ ...prev, parentCode: event.target.value }))}
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 outline-none transition focus:border-blue-400"
                >
                  <option value="">无</option>
                  {parentOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              <label className="space-y-2">
                <span className="text-xs font-black uppercase tracking-wide text-slate-500">编码</span>
                <input
                  value={itemForm.code}
                  onChange={(event) => setItemForm((prev) => ({ ...prev, code: event.target.value }))}
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>
              <label className="space-y-2">
                <span className="text-xs font-black uppercase tracking-wide text-slate-500">名称</span>
                <input
                  value={itemForm.name}
                  onChange={(event) => setItemForm((prev) => ({ ...prev, name: event.target.value }))}
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>

              <label className="space-y-2 md:col-span-2">
                <span className="text-xs font-black uppercase tracking-wide text-slate-500">描述</span>
                <textarea
                  value={itemForm.description}
                  onChange={(event) => setItemForm((prev) => ({ ...prev, description: event.target.value }))}
                  rows={3}
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>

              <label className="space-y-2">
                <span className="text-xs font-black uppercase tracking-wide text-slate-500">排序</span>
                <input
                  type="number"
                  value={itemForm.sortOrder}
                  onChange={(event) =>
                    setItemForm((prev) => ({ ...prev, sortOrder: Number(event.target.value || 0) }))
                  }
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>
              <label className="flex items-center gap-3 pt-8">
                <input
                  type="checkbox"
                  checked={itemForm.isEnabled}
                  onChange={(event) => setItemForm((prev) => ({ ...prev, isEnabled: event.target.checked }))}
                  className="h-4 w-4 rounded border-slate-300 text-blue-600"
                />
                <span className="text-sm font-semibold text-slate-700">启用状态</span>
              </label>
            </div>

            <div className="flex items-center justify-end gap-3 border-t border-slate-100 px-6 py-5">
              <button
                type="button"
                onClick={() => {
                  setShowItemModal(false)
                  setItemForm(EMPTY_ITEM_FORM)
                }}
                className="rounded-2xl border border-slate-200 px-4 py-2 text-sm font-bold text-slate-600 transition hover:bg-slate-50"
              >
                取消
              </button>
              <button
                type="button"
                onClick={() => void handleSaveItem()}
                disabled={itemSaving}
                className="inline-flex items-center gap-2 rounded-2xl bg-blue-600 px-4 py-2 text-sm font-bold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {itemSaving ? <Loader2 size={16} className="animate-spin" /> : null}
                保存
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
