import { useCallback, useEffect, useMemo, useState } from 'react'
import { message } from 'antd'
import {
  Activity,
  ClipboardList,
  FileText,
  Loader2,
  MoreHorizontal,
  Plus,
  RefreshCw,
  Search,
  Stethoscope,
  ToggleLeft,
  ToggleRight,
  Trash2,
  Users,
  Edit3,
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
        className="rounded-full bg-amber-50 px-1.5 py-px text-[10px] font-bold text-amber-700"
      >
        老库
      </span>
    )
  }
  return (
    <span className="rounded-full bg-emerald-50 px-1.5 py-px text-[10px] font-bold text-emerald-700">
      本地
    </span>
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

const CATEGORY_META: Record<CategoryKey, { label: string }> = {
  dialysis: { label: '透析治疗' },
  vascular: { label: '血管通路' },
  order: { label: '医嘱处方' },
  staff: { label: '人员信息' },
  clinical: { label: '临床诊疗' },
  outcome: { label: '转归记录' },
  other: { label: '通用/未归类' },
}

const CATEGORY_ICONS: Record<CategoryKey, React.ComponentType<{ size: number; className?: string }>> = {
  dialysis: Activity,
  vascular: Stethoscope,
  order: ClipboardList,
  staff: Users,
  clinical: Stethoscope,
  outcome: FileText,
  other: MoreHorizontal,
}

const CATEGORY_ORDER: CategoryKey[] = [
  'dialysis', 'vascular', 'order', 'staff', 'clinical', 'outcome', 'other',
]

const TYPE_CATEGORY_MAP: Record<string, CategoryKey> = {
  DialysisMethod: 'dialysis',
  DIALYSIS_MODE: 'dialysis',
  HeparinType: 'dialysis',
  ANTICOAGULANT: 'dialysis',
  Dialysate: 'dialysis',
  DIALYSATE_TYPE: 'dialysis',
  DIALYSATE_GROUP: 'dialysis',
  DialysateFlow: 'dialysis',
  DIALYSATE_FLOW: 'dialysis',
  GlucoseConOptions: 'dialysis',
  GLUCOSE: 'dialysis',
  FluxType: 'dialysis',
  DilutionMnt: 'dialysis',
  SealType: 'dialysis',
  DuringSymptomType: 'dialysis',
  DisinfectionType: 'dialysis',
  Disinfection10Disinfectant: 'dialysis',
  Disinfection10Way: 'dialysis',
  Disinfection20Disinfectant: 'dialysis',
  Disinfection20Way: 'dialysis',
  DealOpportunityType: 'dialysis',
  KType: 'dialysis',
  CaType: 'dialysis',
  HealthEducationType: 'dialysis',
  Treatment_TreatmentStatus: 'dialysis',
  TreatmentStatus: 'dialysis',
  PatientPrescriptionStatus: 'dialysis',
  PatientPlanPrescriptionAdjustmentStatus: 'dialysis',
  PatientDayOrderStatus: 'dialysis',

  AccessType: 'vascular',
  VASCULAR_ACCESS: 'vascular',
  AccessPosition: 'vascular',
  VASCULAR_SITE: 'vascular',
  VenousType: 'vascular',
  VEIN_TYPE: 'vascular',
  ArteryType: 'vascular',
  ARTERY_TYPE: 'vascular',
  VascularAccessChange_Type: 'vascular',
  SURGERY_TYPE: 'vascular',
  CatheterizeMethodType: 'vascular',

  UseMethodType: 'order',
  ORDER_TYPE: 'order',
  CatalogType: 'order',
  ORDER_CATEGORY: 'order',
  UseWayType: 'order',
  ORDER_ROUTE: 'order',
  FrequencyType: 'order',
  ORDER_FREQUENCY: 'order',
  UseOpportunityType: 'order',
  ORDER_TIMING: 'order',
  MaterialType: 'order',
  MATERIAL_CATEGORY: 'order',
  DrugType: 'order',
  DRUG_CATEGORY: 'order',
  BasicUnitOptions: 'order',
  SpecificationUnitOptions: 'order',
  DrugUseOpportunity: 'order',
  BillType: 'order',
  BillTypePosition: 'order',
  ElectronicDocumentType: 'order',

  DOCTOR: 'staff',
  DOCTOR_LIST: 'staff',
  DOCTOR_INFO: 'staff',
  Doctor: 'staff',
  HOSPITAL: 'staff',
  Hospital: 'staff',
  PatientShiftStatus: 'staff',
  '人员类型': 'staff',

  ABOType: 'clinical',
  BLOOD_TYPE_ABO: 'clinical',
  RHType: 'clinical',
  BLOOD_TYPE_RH: 'clinical',
  ExpenseType: 'clinical',
  INSURANCE_TYPE: 'clinical',
  IDType: 'clinical',
  ID_TYPE: 'clinical',
  HospPatientType: 'clinical',
  VISIT_CATEGORY: 'clinical',
  PatientType: 'clinical',
  PATIENT_TYPE: 'clinical',
  EducationLevel: 'clinical',
  EDUCATION_LEVEL: 'clinical',
  MaritalStatus: 'clinical',
  MARITAL_STATUS: 'clinical',
  RelationshipOptions: 'clinical',
  InfectionType: 'clinical',
  PRIMARY_DISEASE: 'clinical',
  COMPLICATION: 'clinical',
  PATHOLOGY: 'clinical',
  TUMOR: 'clinical',
  ALLERGEN: 'clinical',
  DiseaseCourseType: 'clinical',
  InfectionDay: 'clinical',
  InfectionIntervalDay: 'clinical',
  FamilyType: 'clinical',
  PressurePointOptions: 'clinical',
  CaseStatus: 'clinical',
  HealthClassify: 'clinical',
  MessageStatus: 'clinical',
  Classification: 'clinical',

  OutComeType: 'outcome',
  OutComeReason: 'outcome',
  OutComeStatus: 'outcome',
  OUTCOME: 'outcome',

  EquipmentInfomationType: 'other',
  EquipmentInfomationMaintenance: 'other',
  archiveLocation: 'other',
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
    label: `${'\u3000'.repeat(item.level)}${item.name} (${item.code})`,
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
  itemKeyword,
  onItemKeywordChange,
  itemStatusFilter,
  onItemStatusFilterChange,
}: {
  type: DictType
  items: DictItem[]
  loading: boolean
  onAdd: (typeCode: string) => void
  onEdit: (typeCode: string, itemId: string) => void
  onToggle: (typeCode: string, itemId: string) => void
  onDelete: (typeCode: string, itemId: string) => void
  itemKeyword?: string
  onItemKeywordChange?: (value: string) => void
  itemStatusFilter?: 'all' | 'enabled' | 'disabled'
  onItemStatusFilterChange?: (value: 'all' | 'enabled' | 'disabled') => void
}) {
  const flatItems = useMemo(() => flattenItems(items), [items])

  const filteredItems = useMemo(() => {
    let list = flatItems
    if (itemKeyword) {
      const kw = itemKeyword.toLowerCase()
      list = list.filter(
        (item) =>
          item.code.toLowerCase().includes(kw) ||
          item.name.toLowerCase().includes(kw) ||
          (item.description || '').toLowerCase().includes(kw)
      )
    }
    if (itemStatusFilter === 'enabled') {
      list = list.filter((item) => item.isEnabled)
    } else if (itemStatusFilter === 'disabled') {
      list = list.filter((item) => !item.isEnabled)
    }
    return list
  }, [flatItems, itemKeyword, itemStatusFilter])

  return (
    <section className="flex h-full flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white">
      <div className="shrink-0 flex flex-col gap-2 border-b border-slate-100 px-3 py-2 md:flex-row md:items-center md:justify-between">
        <div className="flex items-center gap-1.5">
          <h3 className="text-sm font-black text-slate-900">{type.name}</h3>
          <SourceTag source={type.source} />
          <span className="rounded-full bg-blue-50 px-1.5 py-px text-[10px] font-bold text-blue-600">
            {flatItems.length} 项
          </span>
        </div>
        <button
          type="button"
          onClick={() => onAdd(type.code)}
          disabled={isLegacySource(type.source)}
          title={isLegacySource(type.source) ? '老库字典暂不支持新增' : undefined}
          className={`inline-flex items-center justify-center gap-1 rounded-xl px-3 py-1.5 text-xs font-bold text-white transition ${
            isLegacySource(type.source)
              ? 'cursor-not-allowed bg-slate-300 opacity-60'
              : 'bg-blue-600 hover:bg-blue-700'
          }`}
        >
          <Plus size={14} />
          新增字典值
        </button>
      </div>

      {(onItemKeywordChange || onItemStatusFilterChange) && (
        <div className="shrink-0 flex items-center gap-2 border-b border-slate-100 px-3 py-1.5">
          {onItemKeywordChange && (
            <div className="relative flex-1 max-w-xs">
              <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-400" size={12} />
              <input
                value={itemKeyword || ''}
                onChange={(e) => onItemKeywordChange(e.target.value)}
                placeholder="搜索编码、名称"
                className="w-full h-7 pl-7 pr-2 rounded-lg border border-slate-200 bg-slate-50 text-xs outline-none focus:border-blue-400"
              />
            </div>
          )}
          {onItemStatusFilterChange && (
            <div className="flex gap-0.5 text-[11px]">
              {(['all', 'enabled', 'disabled'] as const).map((f) => (
                <button
                  key={f}
                  type="button"
                  onClick={() => onItemStatusFilterChange(f)}
                  className={`rounded-md px-2 py-0.5 font-semibold transition ${
                    (itemStatusFilter || 'all') === f
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-100 text-slate-500 hover:bg-slate-200'
                  }`}
                >
                  {f === 'all' ? '全部' : f === 'enabled' ? '启用' : '停用'}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-auto">
        <table className="min-w-full text-xs">
          <thead className="sticky top-0 z-10 bg-slate-50">
            <tr className="text-left text-[11px] uppercase tracking-wide text-slate-500">
              <th className="px-3 py-2 font-bold">编码</th>
              <th className="px-3 py-2 font-bold">名称</th>
              <th className="px-3 py-2 font-bold">描述</th>
              <th className="px-3 py-2 font-bold">上级</th>
              <th className="px-3 py-2 font-bold">排序</th>
              <th className="px-3 py-2 font-bold">来源</th>
              <th className="px-3 py-2 font-bold">状态</th>
              <th className="px-3 py-2 font-bold text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={8} className="px-3 py-8 text-center text-slate-500">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 size={14} className="animate-spin" />
                    加载中
                  </span>
                </td>
              </tr>
            ) : filteredItems.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-3 py-10 text-center text-slate-400">
                  <div className="text-xs">{itemKeyword ? '无匹配的字典值' : '暂无字典值'}</div>
                  {!itemKeyword && (
                    <div className="mt-1 text-[11px] text-slate-400">点击"新增字典值"添加第一条数据</div>
                  )}
                </td>
              </tr>
            ) : (
              filteredItems.map((item) => (
                <tr key={item.id} className={`border-t border-slate-100 ${isLegacySource(item.source) ? 'bg-amber-50/30' : ''}`}>
                  <td className="px-3 py-1.5 font-mono text-[11px] font-bold text-slate-700 max-w-[160px] truncate" title={item.code}>
                    <span style={{ paddingLeft: `${item.level * 16}px` }} className="inline-flex items-center gap-1">
                      {item.level > 0 ? <span className="text-slate-300">\u2514</span> : null}
                      {item.code}
                    </span>
                  </td>
                  <td className="px-3 py-1.5 font-semibold text-slate-900 max-w-[140px] truncate" title={item.name}>{item.name}</td>
                  <td className="px-3 py-1.5 text-slate-500 max-w-[140px] truncate" title={item.description || ''}>{item.description || '-'}</td>
                  <td className="px-3 py-1.5 text-slate-500">{item.parentCode || '-'}</td>
                  <td className="px-3 py-1.5 text-slate-500">{item.sortOrder ?? '-'}</td>
                  <td className="px-3 py-1.5">
                    <SourceTag source={item.source} />
                  </td>
                  <td className="px-3 py-1.5">
                    <span
                      className={`rounded-full px-1.5 py-px text-[10px] font-bold ${
                        item.isEnabled ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500'
                      }`}
                    >
                      {item.isEnabled ? '启用' : '停用'}
                    </span>
                  </td>
                  <td className="px-3 py-1.5">
                    <div className="flex items-center justify-end gap-1">
                      <button
                        type="button"
                        onClick={() => onEdit(type.code, item.id)}
                        disabled={isLegacySource(item.source)}
                        className="rounded-lg p-1 text-slate-500 transition hover:bg-blue-50 hover:text-blue-600 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
                        title={isLegacySource(item.source) ? '老库字典暂不支持直接维护' : '编辑'}
                      >
                        <Edit3 size={13} />
                      </button>
                      <button
                        type="button"
                        onClick={() => onToggle(type.code, item.id)}
                        disabled={isLegacySource(item.source)}
                        className="rounded-lg p-1 text-slate-500 transition hover:bg-amber-50 hover:text-amber-600 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
                        title={isLegacySource(item.source) ? '老库字典暂不支持直接维护' : '启用/停用'}
                      >
                        {item.isEnabled ? <ToggleRight size={15} /> : <ToggleLeft size={15} />}
                      </button>
                      <button
                        type="button"
                        onClick={() => onDelete(type.code, item.id)}
                        disabled={isLegacySource(item.source)}
                        className="rounded-lg p-1 text-slate-500 transition hover:bg-red-50 hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
                        title={isLegacySource(item.source) ? '老库字典暂不支持直接维护' : '删除'}
                      >
                        <Trash2 size={13} />
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
  const [initialized, setInitialized] = useState(false)

  const [selectedTypeCode, setSelectedTypeCode] = useState<string>('')
  const [typeKeyword, setTypeKeyword] = useState('')
  const [typeSourceFilter, setTypeSourceFilter] = useState<'all' | 'local' | 'legacy'>('all')
  const [itemKeyword, setItemKeyword] = useState('')
  const [itemStatusFilter, setItemStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all')

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

  const selectedType = useMemo(
    () => sortedTypes.find((type) => type.code === selectedTypeCode) ?? null,
    [selectedTypeCode, sortedTypes]
  )

  const filteredTypes = useMemo(() => {
    let list = typesByCategory[selectedCategory]
    if (typeKeyword) {
      const kw = typeKeyword.toLowerCase()
      list = list.filter(
        (t) => t.name.toLowerCase().includes(kw) || t.code.toLowerCase().includes(kw)
      )
    }
    if (typeSourceFilter === 'local') {
      list = list.filter((t) => !t.source || t.source !== 'legacy')
    } else if (typeSourceFilter === 'legacy') {
      list = list.filter((t) => t.source === 'legacy')
    }
    return list
  }, [typesByCategory, selectedCategory, typeKeyword, typeSourceFilter])

  const totalLocalTypes = useMemo(
    () => sortedTypes.filter((t) => !t.source || t.source !== 'legacy').length,
    [sortedTypes]
  )
  const totalLegacyTypes = useMemo(
    () => sortedTypes.filter((t) => t.source === 'legacy').length,
    [sortedTypes]
  )

  const loadTypes = useCallback(async () => {
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
  }, [])

  const loadItems = useCallback(async (typeCode: string, forceRefresh = false) => {
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
  }, [])

  useEffect(() => {
    void loadTypes()
  }, [loadTypes])

  useEffect(() => {
    if (dictTypes.length === 0) return
    if (initialized) return

    for (const cat of CATEGORY_ORDER) {
      const list = typesByCategory[cat]
      if (list.length > 0) {
        setSelectedCategory(cat)
        setSelectedTypeCode(list[0].code)
        break
      }
    }
    setInitialized(true)
  }, [dictTypes, initialized, typesByCategory])

  useEffect(() => {
    if (!initialized) return
    if (selectedCategory === 'other' && selectedOtherTypeCode && typesByCategory.other.some((t) => t.code === selectedOtherTypeCode)) {
      setSelectedTypeCode(selectedOtherTypeCode)
      return
    }
    const list = typesByCategory[selectedCategory]
    if (list.some((t) => t.code === selectedTypeCode)) return
    if (list.length > 0) {
      setSelectedTypeCode(list[0].code)
    } else {
      setSelectedTypeCode('')
    }
  }, [selectedCategory, typesByCategory, selectedOtherTypeCode, initialized, selectedTypeCode])

  useEffect(() => {
    if (selectedCategory === 'other' && selectedTypeCode) {
      setSelectedOtherTypeCode(selectedTypeCode)
    }
  }, [selectedCategory, selectedTypeCode])

  useEffect(() => {
    if (selectedTypeCode && !itemsByType[selectedTypeCode]) {
      void loadItems(selectedTypeCode)
    }
  }, [selectedTypeCode, itemsByType, loadItems])

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
        ? `确定删除"${item.name}"吗？该项下还有 ${descendantCount} 个子项会一起删除。`
        : `确定删除"${item.name}"吗？`

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

  const handleRefresh = async () => {
    await loadTypes()
    if (selectedTypeCode) {
      await loadItems(selectedTypeCode, true)
    }
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

  const typeItems = itemsByType[selectedTypeCode] || []
  const currentItems = flattenItems(typeItems)

  return (
    <div className="flex flex-col h-[calc(100vh-112px)] min-h-[560px] overflow-hidden">
      <div className="shrink-0 flex items-center gap-1 text-xs text-slate-400 px-1 mb-2">
        <span>系统设置</span><span>/</span><span className="text-slate-600">字典配置</span>
      </div>

      <section className="shrink-0 rounded-2xl border border-slate-200 bg-white px-4 py-2.5 mb-2 shadow-sm">
        <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-black tracking-tight text-slate-900">字典配置</h1>
            <div className="flex items-center gap-3 text-xs text-slate-400">
              <span>共 <b className="text-slate-600">{sortedTypes.length}</b> 类型</span>
              <span>老库只读 <b className="text-amber-600">{totalLegacyTypes}</b></span>
              <span>本地维护 <b className="text-emerald-600">{totalLocalTypes}</b></span>
            </div>
          </div>
          <button
            type="button"
            onClick={() => void handleRefresh()}
            className="inline-flex items-center justify-center gap-1.5 rounded-xl border border-slate-200 px-3 py-1.5 text-xs font-bold text-slate-600 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700"
          >
            <RefreshCw size={14} />
            刷新
          </button>
        </div>
      </section>

      <section className="min-h-0 flex-1 grid gap-3 xl:grid-cols-[180px_240px_minmax(0,1fr)]">
        {/* 左侧：业务域导航 */}
        <aside className="flex flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="shrink-0 px-3 py-2 border-b border-slate-100">
            <h2 className="text-sm font-black text-slate-900">业务域</h2>
            <p className="text-[10px] text-slate-400">{sortedTypes.length} 个字典类型</p>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto p-1.5">
            <div className="space-y-0.5">
              {CATEGORY_ORDER.map((key) => {
                const CatIcon = CATEGORY_ICONS[key]
                const types = typesByCategory[key]
                const active = selectedCategory === key
                return (
                  <button
                    key={key}
                    type="button"
                    onClick={() => {
                      setSelectedCategory(key)
                      setTypeKeyword('')
                      setTypeSourceFilter('all')
                      setItemKeyword('')
                    }}
                    className={`w-full rounded-lg px-2.5 py-1.5 text-left transition ${
                      active
                        ? 'bg-slate-900 text-white shadow-md'
                        : 'text-slate-700 hover:bg-slate-50'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-1.5">
                      <div className="flex items-center gap-2 min-w-0">
                        <CatIcon size={14} className={active ? 'text-blue-400' : 'text-slate-400'} />
                        <span className="text-xs font-bold truncate">{CATEGORY_META[key].label}</span>
                      </div>
                      <span
                        className={`shrink-0 rounded-full px-1.5 py-px text-[10px] font-bold ${
                          active ? 'bg-slate-700 text-slate-300' : 'bg-slate-100 text-slate-500'
                        }`}
                      >
                        {types.length}
                      </span>
                    </div>
                  </button>
                )
              })}
            </div>
          </div>
        </aside>

        {/* 中间：字典类型列表 */}
        <aside className="flex flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="shrink-0 border-b border-slate-100 px-3 py-2">
            <h3 className="text-xs font-black text-slate-900">字典类型</h3>
            <p className="text-[10px] text-slate-400">
              {CATEGORY_META[selectedCategory].label} · {typesByCategory[selectedCategory].length} 个类型
            </p>
          </div>

          <div className="shrink-0 border-b border-slate-100 px-2 py-1.5 space-y-1.5">
            <div className="relative">
              <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-400" size={12} />
              <input
                value={typeKeyword}
                onChange={(e) => setTypeKeyword(e.target.value)}
                placeholder="搜索字典类型"
                className="w-full h-7 pl-7 pr-2 rounded-lg border border-slate-200 bg-slate-50 text-xs outline-none focus:border-blue-400"
              />
            </div>
            <div className="flex gap-0.5 text-[10px]">
              {(['all', 'local', 'legacy'] as const).map((f) => (
                <button
                  key={f}
                  type="button"
                  onClick={() => setTypeSourceFilter(f)}
                  className={`rounded-md px-2 py-0.5 font-semibold transition ${
                    typeSourceFilter === f
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-100 text-slate-500 hover:bg-slate-200'
                  }`}
                >
                  {f === 'all' ? '全部' : f === 'local' ? '本地' : '老库'}
                </button>
              ))}
            </div>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto">
            {loadingTypes ? (
              <div className="flex items-center gap-2 px-3 py-6 text-xs text-slate-500">
                <Loader2 size={12} className="animate-spin" />
                加载类型中
              </div>
            ) : filteredTypes.length === 0 ? (
              <div className="px-3 py-10 text-center text-[11px] text-slate-400">
                {typesByCategory[selectedCategory].length === 0
                  ? '当前业务域暂无字典类型'
                  : '无匹配的字典类型'}
              </div>
            ) : (
              <div className="space-y-px p-1">
                {filteredTypes.map((type) => {
                  const active = selectedTypeCode === type.code
                  return (
                    <button
                      key={type.code}
                      type="button"
                      title={type.code}
                      onClick={() => {
                        setSelectedTypeCode(type.code)
                        if (selectedCategory === 'other') {
                          setSelectedOtherTypeCode(type.code)
                        }
                        setItemKeyword('')
                      }}
                      className={`w-full rounded-lg px-2.5 py-1.5 text-left transition ${
                        active
                          ? 'bg-blue-50 border border-blue-200'
                          : 'border border-transparent hover:bg-slate-50'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-1.5">
                        <div className="min-w-0 flex-1">
                          <div className="text-xs font-bold text-slate-900 truncate">{type.name}</div>
                          <div className="mt-0.5">
                            <SourceTag source={type.source} />
                          </div>
                        </div>
                        <span className="shrink-0 text-[10px] font-bold text-slate-400">
                          {itemsByType[type.code] ? flattenItems(itemsByType[type.code]).length : '-'}
                        </span>
                      </div>
                    </button>
                  )
                })}
              </div>
            )}
          </div>
        </aside>

        {/* 右侧：当前字典值维护 */}
        <main className="flex h-full flex-col overflow-hidden">
          {selectedType ? (
            <>
              <section className="shrink-0 rounded-2xl border border-slate-200 bg-white px-3 py-2 shadow-sm mb-2">
                <div className="flex flex-col gap-1.5 md:flex-row md:items-center md:justify-between">
                  <div className="min-w-0" title={selectedType.code}>
                    <h2 className="text-sm font-black text-slate-900 truncate">{selectedType.name}</h2>
                    <p className="flex items-center gap-1.5 text-[11px] text-slate-400">
                      <span>{CATEGORY_META[resolveCategory(selectedType.code)].label}</span>
                      <span className="rounded-full bg-amber-50 px-1.5 py-px text-[10px] font-bold text-amber-700">老库只读</span>
                      <span>{currentItems.length} 项</span>
                    </p>
                  </div>
                  {!isLegacySource(selectedType.source) && (
                    <button
                      type="button"
                      onClick={() => openAddModal(selectedType.code)}
                      className="inline-flex items-center gap-1 rounded-lg bg-blue-600 px-3 py-1.5 text-[11px] font-bold text-white transition hover:bg-blue-700"
                    >
                      <Plus size={13} />
                      新增字典值
                    </button>
                  )}
                </div>
              </section>

              <div className="min-h-0 flex-1">
                <TypeValueTable
                  type={selectedType}
                  items={typeItems}
                  loading={loadingTypeCodes.has(selectedTypeCode)}
                  onAdd={openAddModal}
                  onEdit={openEditModal}
                  onToggle={handleToggleItem}
                  onDelete={handleDeleteItem}
                  itemKeyword={itemKeyword}
                  onItemKeywordChange={setItemKeyword}
                  itemStatusFilter={itemStatusFilter}
                  onItemStatusFilterChange={setItemStatusFilter}
                />
              </div>
            </>
          ) : selectedTypeCode ? (
            <section className="rounded-2xl border border-dashed border-amber-300 bg-amber-50/40 px-4 py-12 text-center">
              <p className="text-sm font-bold text-amber-800">字典类型编码未匹配</p>
              <p className="mt-1 text-xs text-amber-600">
                编码 <span className="font-mono font-bold">{selectedTypeCode}</span> 在类型列表中不存在。
              </p>
            </section>
          ) : (
            <section className="rounded-2xl border border-dashed border-slate-300 bg-white px-4 py-12 text-center text-slate-400">
              请从左侧选择字典类型以查看字典值。
            </section>
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
                <span className="text-sm font-semibold text-slate-700">启用</span>
              </label>
            </div>

            <div className="flex justify-end gap-3 border-t border-slate-100 px-6 py-4">
              <button
                type="button"
                onClick={() => {
                  setShowItemModal(false)
                  setItemForm(EMPTY_ITEM_FORM)
                }}
                className="rounded-2xl border border-slate-200 px-5 py-2.5 text-sm font-bold text-slate-600 transition hover:bg-slate-50"
              >
                取消
              </button>
              <button
                type="button"
                onClick={handleSaveItem}
                disabled={itemSaving}
                className="rounded-2xl bg-blue-600 px-6 py-2.5 text-sm font-bold text-white transition hover:bg-blue-700 disabled:opacity-60"
              >
                {itemSaving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
