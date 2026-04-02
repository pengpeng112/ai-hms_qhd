// MaterialSyncModal - 更换血管通路及材料同步弹窗
// 汇总展示所有治疗方案(HD/HDF/...)的材料列表，
// 切换血管通路类型时自动同步默认材料

import { useState, useCallback, useRef, memo } from 'react'
import { X, ArrowRightLeft, Trash2, ChevronDown, Save, FileJson } from 'lucide-react'
import { MaterialSelector } from '@/components/treatment-form'

// ===== 类型定义 =====

export interface SyncMaterialItem {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
}

interface VascularAccessOption {
  id: string
  label: string
  accessType: string
}

interface PlanMaterials {
  mode: string        // HD, HDF, HD+HP, ...
  planId: string
  materials: SyncMaterialItem[]
}

export interface MaterialSyncResult {
  selectedAccessId: string
  plans: PlanMaterials[]
}

interface MaterialSyncModalProps {
  isOpen: boolean
  onClose: () => void
  onSave: (result: MaterialSyncResult) => void
  vascularAccesses: VascularAccessOption[]
  currentAccessId: string
  plans: PlanMaterials[]
}

// ===== 单个方案的材料表格 =====

const MaterialTable = memo(function MaterialTable({
  title,
  materials,
  onAdd,
  onRemove,
  onUpdateCount,
  onReplaceMaterial,
  onToggleSelect,
  onToggleSelectAll,
  selectedIds,
}: {
  title: string
  materials: SyncMaterialItem[]
  onAdd: (name: string, material: Partial<{ category: string; code: string; brand: string; spec: string }>) => void
  onRemove: () => void
  onUpdateCount: (id: string, count: number) => void
  onReplaceMaterial: (id: string, name: string, material: Partial<{ category: string; code: string; brand: string; spec: string }>) => void
  onToggleSelect: (id: string) => void
  onToggleSelectAll: () => void
  selectedIds: Set<string>
}) {
  const allSelected = materials.length > 0 && selectedIds.size === materials.length

  // 判断是否为通路特有材料（内瘘包/上机包/下机包）
  const isExclusive = (name: string) =>
    name === '内瘘包' || name === '上机包' || name === '下机包'

  return (
    <div className="bg-white border border-slate-200 rounded-[24px] overflow-hidden flex flex-col min-h-[320px] shadow-sm hover:shadow-md transition-shadow">
      {/* 标题栏 */}
      <div className="bg-slate-50 px-5 py-3 border-b border-slate-100 font-black text-slate-800 text-sm flex justify-between items-center shrink-0">
        <span>{title} 方案材料配置</span>
        <span className="text-xs text-slate-400 font-bold">{materials.length} 项</span>
      </div>

      {/* 搜索添加 + 删除 */}
      <div className="px-4 py-3 border-b border-slate-50 flex items-center gap-3 shrink-0">
        <div className="flex-1">
          <MaterialSelector
            value=""
            onChange={onAdd}
            placeholder="搜索并添加材料..."
            className="w-full text-xs"
          />
        </div>
        <button
          onClick={onRemove}
          disabled={selectedIds.size === 0}
          className={`px-3 h-9 rounded-xl text-xs flex items-center gap-1 font-bold transition-all shrink-0 ${
            selectedIds.size > 0
              ? 'bg-red-50 text-red-500 hover:bg-red-100'
              : 'bg-slate-50 text-slate-300 cursor-not-allowed'
          }`}
        >
          <Trash2 size={13} /> 删除{selectedIds.size > 0 ? `(${selectedIds.size})` : ''}
        </button>
      </div>

      {/* 材料表格 */}
      <div className="flex-1 overflow-y-auto custom-scrollbar">
        <table className="w-full text-left text-xs border-separate border-spacing-0">
          <thead className="bg-[#f0f7ff] text-blue-600 font-black uppercase tracking-widest text-[9px] border-b border-blue-100 sticky top-0 z-10">
            <tr>
              <th className="py-2.5 px-3 w-9 text-center border-b border-blue-100">
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={onToggleSelectAll}
                  className="w-3.5 h-3.5 rounded border-slate-300 accent-blue-600"
                />
              </th>
              <th className="py-2.5 px-2 w-10 text-center border-b border-blue-100">序号</th>
              <th className="py-2.5 px-3 border-b border-blue-100">材料名称</th>
              <th className="py-2.5 px-3 w-24 text-center border-b border-blue-100">分类</th>
              <th className="py-2.5 px-3 w-16 text-center border-b border-blue-100">数量</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {materials.length === 0 ? (
              <tr>
                <td colSpan={5} className="py-10 text-center text-slate-400">
                  <FileJson size={24} className="mx-auto mb-2 text-slate-200" />
                  <p className="text-xs">暂无材料，请从上方搜索添加</p>
                </td>
              </tr>
            ) : (
              materials.map((m, idx) => {
                const exclusive = isExclusive(m.name)
                return (
                  <tr
                    key={m.id}
                    className={`hover:bg-slate-50/50 group ${exclusive ? 'bg-blue-50/30' : ''}`}
                  >
                    <td className="py-2 px-3 text-center border-b border-slate-50">
                      <input
                        type="checkbox"
                        checked={selectedIds.has(m.id)}
                        onChange={() => onToggleSelect(m.id)}
                        className="w-3.5 h-3.5 rounded border-slate-300 accent-blue-600"
                      />
                    </td>
                    <td className="py-2 px-2 text-center text-slate-300 font-mono text-[10px] border-b border-slate-50">
                      {idx + 1}
                    </td>
                    <td className="py-1.5 px-3 border-b border-slate-50">
                      <MaterialSelector
                        value={m.name}
                        onChange={(name, material) => onReplaceMaterial(m.id, name, material)}
                        placeholder="选择材料"
                        className="w-full text-xs font-black rounded-lg px-2 py-1"
                      />
                    </td>
                    <td className="py-2 px-3 text-center border-b border-slate-50">
                      <span
                        className={`px-2 py-0.5 rounded-lg text-[9px] font-black uppercase ${
                          exclusive ? 'bg-blue-100 text-blue-600' : 'bg-slate-100 text-slate-400'
                        }`}
                      >
                        {m.category || '-'}
                      </span>
                    </td>
                    <td className="py-2 px-3 text-center border-b border-slate-50">
                      <input
                        type="number"
                        min={1}
                        value={m.count}
                        onChange={(e) => onUpdateCount(m.id, Math.max(1, parseInt(e.target.value) || 1))}
                        className="w-full h-7 border border-transparent hover:border-slate-200 rounded-lg text-center text-xs font-black outline-none focus:ring-2 focus:ring-blue-500/20 bg-transparent"
                      />
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
})

// ===== 主弹窗组件 =====

export default function MaterialSyncModal({
  isOpen,
  onClose,
  onSave,
  vascularAccesses,
  currentAccessId,
  plans: initialPlans,
}: MaterialSyncModalProps) {
  // 临时 ID 生成器（组件级，随 key 重置）
  const nextTempIdRef = useRef(1)
  const tempId = () => `sync-temp-${nextTempIdRef.current++}`

  // 状态均使用初始值（组件通过 key 在每次打开时重新挂载）
  const [selectedAccessId, setSelectedAccessId] = useState(currentAccessId)
  const [planMaterials, setPlanMaterials] = useState<PlanMaterials[]>(() =>
    initialPlans.map(p => ({
      ...p,
      materials: p.materials.map(m => ({ ...m })),
    }))
  )
  // 每个方案独立的选中状态
  const [selectedIds, setSelectedIds] = useState<Record<string, Set<string>>>(() => {
    const emptySelected: Record<string, Set<string>> = {}
    initialPlans.forEach(p => { emptySelected[p.planId] = new Set() })
    return emptySelected
  })

  // --- 材料操作 ---

  const handleAdd = useCallback((planId: string, name: string, material: Partial<{ category: string; code: string; brand: string; spec: string }>) => {
    setPlanMaterials(prev =>
      prev.map(p => {
        if (p.planId !== planId) return p
        // 检查重复
        if (p.materials.some(m => m.name === name)) return p
        return {
          ...p,
          materials: [
            ...p.materials,
            {
              id: tempId(),
              name,
              category: material.category || '',
              count: 1,
              code: material.code || '',
              brand: material.brand || '',
              spec: material.spec || '',
            },
          ],
        }
      })
    )
  }, [])

  const handleRemoveSelected = useCallback((planId: string) => {
    const selected = selectedIds[planId]
    if (!selected || selected.size === 0) return
    setPlanMaterials(prev =>
      prev.map(p => {
        if (p.planId !== planId) return p
        return {
          ...p,
          materials: p.materials.filter(m => !selected.has(m.id)),
        }
      })
    )
    setSelectedIds(prev => ({ ...prev, [planId]: new Set() }))
  }, [selectedIds])

  const handleUpdateCount = useCallback((planId: string, materialId: string, count: number) => {
    setPlanMaterials(prev =>
      prev.map(p => {
        if (p.planId !== planId) return p
        return {
          ...p,
          materials: p.materials.map(m =>
            m.id === materialId ? { ...m, count } : m
          ),
        }
      })
    )
  }, [])

  const handleReplaceMaterial = useCallback((planId: string, materialId: string, name: string, material: Partial<{ category: string; code: string; brand: string; spec: string }>) => {
    setPlanMaterials(prev =>
      prev.map(p => {
        if (p.planId !== planId) return p
        return {
          ...p,
          materials: p.materials.map(m =>
            m.id === materialId
              ? {
                  ...m,
                  name,
                  category: material.category || m.category,
                  code: material.code || m.code,
                  brand: material.brand || m.brand,
                  spec: material.spec || m.spec,
                }
              : m
          ),
        }
      })
    )
  }, [])

  const handleToggleSelect = useCallback((planId: string, materialId: string) => {
    setSelectedIds(prev => {
      const set = new Set(prev[planId] || [])
      if (set.has(materialId)) set.delete(materialId)
      else set.add(materialId)
      return { ...prev, [planId]: set }
    })
  }, [])

  const handleToggleSelectAll = useCallback((planId: string) => {
    const plan = planMaterials.find(p => p.planId === planId)
    if (!plan) return
    const current = selectedIds[planId] || new Set()
    const allSelected = plan.materials.length > 0 && current.size === plan.materials.length
    if (allSelected) {
      setSelectedIds(prev => ({ ...prev, [planId]: new Set() }))
    } else {
      setSelectedIds(prev => ({
        ...prev,
        [planId]: new Set(plan.materials.map(m => m.id)),
      }))
    }
  }, [planMaterials, selectedIds])

  const handleSave = () => {
    onSave({
      selectedAccessId,
      plans: planMaterials,
    })
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/60 backdrop-blur-md animate-fade-in p-6">
      <div className="bg-white rounded-[40px] shadow-2xl w-full max-w-6xl overflow-hidden animate-scale-in flex flex-col max-h-[95vh] ring-1 ring-black/5">
        {/* Header */}
        <div className="bg-[#eef6ff] px-10 py-6 flex items-center justify-between border-b border-blue-100 shrink-0">
          <h3 className="text-xl font-black text-slate-800 flex items-center gap-3">
            <ArrowRightLeft className="text-blue-600" size={24} /> 更换血管通路及材料同步
          </h3>
          <button
            onClick={onClose}
            className="p-2 hover:bg-white/50 rounded-2xl transition-all text-slate-400 hover:text-slate-600"
          >
            <X size={22} />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto px-10 py-6 bg-[#f8fbff] flex flex-col gap-6 min-h-0">
          {/* 血管通路选择 */}
          <div className="w-80 shrink-0">
            <p className="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-2 ml-1">
              选择血管通路类型
            </p>
            <div className="relative">
              <select
                value={selectedAccessId}
                onChange={(e) => setSelectedAccessId(e.target.value)}
                className="w-full h-12 border border-slate-200 rounded-[18px] px-5 text-sm appearance-none outline-none focus:ring-4 focus:ring-blue-500/10 bg-white font-black text-blue-600 shadow-sm"
              >
                {vascularAccesses.map(va => (
                  <option key={va.id} value={va.id}>
                    {va.label}
                  </option>
                ))}
              </select>
              <ChevronDown
                size={18}
                className="absolute right-5 top-1/2 -translate-y-1/2 text-blue-400 pointer-events-none"
              />
            </div>
          </div>

          {/* 方案材料列表（每行最多2个，可滚动） */}
          {planMaterials.length === 0 ? (
            <div className="flex-1 flex items-center justify-center text-slate-400">
              <div className="text-center">
                <FileJson size={40} className="mx-auto mb-3 text-slate-200" />
                <p className="text-sm font-bold">暂无治疗方案</p>
                <p className="text-xs mt-1">请先创建治疗方案后再进行材料同步</p>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-6">
              {planMaterials.map(plan => (
                <MaterialTable
                  key={plan.planId}
                  title={plan.mode}
                  materials={plan.materials}
                  onAdd={(name, material) => handleAdd(plan.planId, name, material)}
                  onRemove={() => handleRemoveSelected(plan.planId)}
                  onUpdateCount={(id, count) => handleUpdateCount(plan.planId, id, count)}
                  onReplaceMaterial={(id, name, material) => handleReplaceMaterial(plan.planId, id, name, material)}
                  onToggleSelect={(id) => handleToggleSelect(plan.planId, id)}
                  onToggleSelectAll={() => handleToggleSelectAll(plan.planId)}
                  selectedIds={selectedIds[plan.planId] || new Set()}
                />
              ))}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-10 py-6 bg-white border-t border-slate-100 flex justify-end gap-4 shrink-0">
          <button
            onClick={onClose}
            className="px-8 py-3 rounded-2xl border border-slate-200 text-slate-500 text-sm font-black hover:bg-slate-50 transition-all"
          >
            取消
          </button>
          <button
            onClick={handleSave}
            className="px-12 py-3 rounded-2xl bg-blue-600 text-white text-sm font-black shadow-xl shadow-blue-100 hover:bg-blue-700 transition-all flex items-center gap-3"
          >
            <Save size={18} /> 确认并保存同步
          </button>
        </div>
      </div>
    </div>
  )
}
