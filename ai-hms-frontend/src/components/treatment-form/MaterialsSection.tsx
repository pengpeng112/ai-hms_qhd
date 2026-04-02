// 透析材料区域组件

import { memo } from 'react'
import { Search, ClipboardList, Trash } from 'lucide-react'
import { MaterialSelector } from './MaterialSelector'
import type { MaterialItem } from './types'

interface MaterialsSectionProps {
  materials: MaterialItem[]
  onAdd: (name: string, material: Partial<MaterialItem>) => void
  onRemove: (id: string) => void
  onUpdate: (id: string, field: string, value: string | number) => void
  onReplace?: (id: string, patch: Partial<MaterialItem>) => void
}

export const MaterialsSection = memo(function MaterialsSection({
  materials,
  onAdd,
  onRemove,
  onUpdate,
  onReplace,
}: MaterialsSectionProps) {
  return (
    <div className="space-y-4">
      <h4 className="text-sm font-black text-blue-600 uppercase tracking-widest flex items-center gap-2">
        <ClipboardList size={16} /> 02. 材料清单
      </h4>

      {/* 搜索添加材料 */}
      <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-2xl p-4 border border-blue-100">
        <label className="block text-xs font-black text-slate-600 mb-2">
          <Search size={12} className="inline mr-1" /> 搜索并添加材料
        </label>
        <MaterialSelector
          value=""
          onChange={onAdd}
          placeholder="输入材料名称、编码或品牌搜索..."
          className="w-full"
        />
        <p className="text-[10px] text-slate-400 mt-2">
          搜索并选择材料后，材料将自动添加到下方列表中
        </p>
      </div>

      {/* 已选材料列表 */}
      <div className="border border-slate-100 rounded-[32px] overflow-visible shadow-sm">
        <table className="w-full text-left text-[11px]">
          <thead className="bg-slate-50 text-slate-400 font-black uppercase tracking-tighter">
            <tr>
              <th className="px-6 py-4 w-16">#</th>
              <th className="px-6 py-4 min-w-[360px] w-[380px]">材料名称</th>
              <th className="px-6 py-4 w-28 whitespace-nowrap">材料分类</th>
              <th className="px-6 py-4 w-24 text-center whitespace-nowrap">数量</th>
              <th className="px-6 py-4 w-28 whitespace-nowrap">品牌</th>
              <th className="px-6 py-4 w-28 whitespace-nowrap">规格</th>
              <th className="px-6 py-4">备注</th>
              <th className="px-6 py-4 text-right w-24 whitespace-nowrap">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50 font-bold text-slate-700">
            {materials.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-6 py-12 text-center text-slate-400">
                  <ClipboardList size={32} className="mx-auto mb-2 text-slate-300" />
                  <p className="text-sm">暂无材料</p>
                  <p className="text-xs mt-1">在上方搜索框中搜索并添加材料</p>
                </td>
              </tr>
            ) : (
              materials.map((m, index) => (
                <tr key={m.id} className="hover:bg-slate-50/50 group">
                  <td className="px-6 py-4 text-slate-400">{index + 1}</td>
                  <td className="px-6 py-4 min-w-[360px] w-[380px]">
                    <MaterialSelector
                      value={m.name}
                      onChange={(name, material) => {
                        if (onReplace) {
                          onReplace(m.id, {
                            name,
                            category: material.category || '',
                            code: material.code || '',
                            brand: material.brand || '',
                            spec: material.spec || '',
                          })
                        } else {
                          onUpdate(m.id, 'name', name)
                          if (material.category) onUpdate(m.id, 'category', material.category)
                          if (material.code) onUpdate(m.id, 'code', material.code)
                          if (material.brand) onUpdate(m.id, 'brand', material.brand)
                          if (material.spec) onUpdate(m.id, 'spec', material.spec)
                        }
                      }}
                      placeholder="请选择材料"
                      className="w-full text-xs font-black rounded-lg px-3 py-2"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-2 py-1 bg-slate-100 text-slate-600 rounded-lg text-[10px] font-bold">
                      {m.category || '-'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center whitespace-nowrap">
                    <input
                      type="number"
                      min="1"
                      value={m.count}
                      onChange={(e) =>
                        onUpdate(m.id, 'count', Math.max(1, parseInt(e.target.value) || 1))
                      }
                      className="w-16 px-2 py-1 border border-slate-200 rounded text-center font-bold focus:ring-2 focus:ring-blue-500 outline-none"
                    />
                  </td>
                  <td className="px-6 py-4 text-slate-500 whitespace-nowrap">{m.brand || '-'}</td>
                  <td className="px-6 py-4 text-slate-500 whitespace-nowrap">{m.spec || '-'}</td>
                  <td className="px-6 py-4">
                    <input
                      type="text"
                      value={m.note}
                      onChange={(e) => onUpdate(m.id, 'note', e.target.value)}
                      className="w-full px-2 py-1 border border-slate-200 rounded text-slate-300 focus:ring-2 focus:ring-blue-500 outline-none"
                      placeholder="备注"
                    />
                  </td>
                  <td className="px-6 py-4 text-right whitespace-nowrap">
                    <button
                      onClick={() => onRemove(m.id)}
                      className="p-2 text-slate-300 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all"
                    >
                      <Trash size={14} />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
})
