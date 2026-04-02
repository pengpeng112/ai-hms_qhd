// MaterialSelector - 材料选择器组件
// 支持搜索、下拉选择、从材料目录中选择材料

import { useState, useCallback, memo, useEffect, useRef, type KeyboardEvent } from 'react'
import { Search, ChevronDown, Check, X, FileJson } from 'lucide-react'
import { materialCatalogApi } from '@/services/treatmentConfigApi'

interface Material {
  id: string
  name: string
  category: string
  code: string
  brand: string
  spec: string
  isEnabled: boolean
}

interface MaterialSelectorProps {
  value: string
  onChange: (name: string, material: Partial<Material>) => void
  placeholder?: string
  className?: string
}

// 材料列表项组件
const MaterialItem = memo(function MaterialItem({
  material,
  searchQuery,
  onSelect,
  isSelected
}: {
  material: Material
  searchQuery: string
  onSelect: (material: Material) => void
  isSelected: boolean
}) {
  const isHighlighted = searchQuery && material.name.toLowerCase().includes(searchQuery.toLowerCase())

  return (
    <div
      onClick={() => onSelect(material)}
      className={`px-4 py-3 flex items-center justify-between cursor-pointer transition-all border-b border-slate-50 last:border-0 hover:bg-blue-50 ${
        isHighlighted ? 'bg-blue-50' : ''
      }`}
    >
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <FileJson size={16} className="text-slate-400 flex-shrink-0" />
        <div className="flex-1 min-w-0">
          <div className="text-sm font-bold text-slate-700 truncate">{material.name}</div>
          <div className="text-xs text-slate-400 flex items-center gap-2">
            <span>{material.category}</span>
            {material.code && <span className="text-slate-300">| {material.code}</span>}
            {material.brand && <span className="text-slate-300">| {material.brand}</span>}
          </div>
        </div>
      </div>
      {isSelected && (
        <button className="p-1 rounded-full hover:bg-blue-100 text-blue-500 flex-shrink-0">
          <Check size={14} />
        </button>
      )}
    </div>
  )
})

export const MaterialSelector = memo(function MaterialSelector({
  value,
  onChange,
  placeholder = '请选择材料',
  className = ''
}: MaterialSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [materials, setMaterials] = useState<Material[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const containerRef = useRef<HTMLDivElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)

  // 加载启用的材料目录
  const loadMaterials = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await materialCatalogApi.list({ pageSize: 9999 })
      setMaterials(
        response.items
          .filter(m => m.isEnabled)
          .map(m => ({
            id: m.id.toString(),
            name: m.name,
            category: m.category,
            code: m.code || '',
            brand: m.brand || '',
            spec: m.spec || '',
            isEnabled: m.isEnabled
          }))
      )
    } catch (error) {
      console.error('加载材料目录失败:', error)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 打开下拉列表时加载材料
  const handleOpen = useCallback(() => {
    if (!isOpen && materials.length === 0) {
      loadMaterials()
    }
    setIsOpen(true)
    setSearchQuery('')
    setSelectedIndex(-1)
    // 自动聚焦搜索框
    setTimeout(() => searchInputRef.current?.focus(), 0)
  }, [isOpen, materials.length, loadMaterials])

  // 选择材料
  const handleSelect = useCallback((material: Material) => {
    onChange(material.name, {
      category: material.category,
      code: material.code,
      brand: material.brand,
      spec: material.spec
    })
    setIsOpen(false)
  }, [onChange])

  // 过滤后的材料列表
  const filteredMaterials = materials.filter(m => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    return (
      m.name.toLowerCase().includes(query) ||
      m.code.toLowerCase().includes(query) ||
      m.category.toLowerCase().includes(query) ||
      m.brand.toLowerCase().includes(query) ||
      m.spec.toLowerCase().includes(query)
    )
  })

  // 键盘导航
  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      if (!isOpen || filteredMaterials.length === 0) return

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex(prev => {
            const next = prev < filteredMaterials.length - 1 ? prev + 1 : prev
            handleSelect(filteredMaterials[next])
            return next
          })
          break
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex(prevIndex => {
            const newIndex = Math.max(-1, prevIndex - 1)
            if (newIndex >= 0) {
              handleSelect(filteredMaterials[newIndex])
            }
            return newIndex
          })
          break
        case 'Enter':
          e.preventDefault()
          if (selectedIndex >= 0 && filteredMaterials[selectedIndex]) {
            handleSelect(filteredMaterials[selectedIndex])
          }
          setIsOpen(false)
          break
        case 'Escape':
          e.preventDefault()
          setIsOpen(false)
          break
      }
    },
    [isOpen, filteredMaterials, selectedIndex, handleSelect]
  )

  // 点击外部关闭
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isOpen])

  return (
    <div ref={containerRef} className={`relative ${isOpen ? 'z-50' : 'z-10'}`}>
      <div
        onClick={handleOpen}
        className={`flex items-center justify-between px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 cursor-pointer hover:bg-slate-50 focus:ring-2 focus:ring-blue-500 outline-none transition-all ${className}`}
      >
        <span className={value ? 'text-slate-800' : 'text-slate-400'}>
          {value || placeholder}
        </span>
        <ChevronDown size={14} className="text-slate-400 flex-shrink-0" />
      </div>

      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-slate-200 rounded-xl shadow-lg max-h-60 overflow-hidden flex flex-col z-50">
          {/* 搜索框 */}
          <div className="p-3 border-b border-slate-100 bg-slate-50">
            <div className="relative">
              <Search
                size={14}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                ref={searchInputRef}
                type="text"
                placeholder="搜索材料名称、编码、品牌..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full pl-9 pr-8 py-2 bg-white border border-slate-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                autoFocus
              />
              {searchQuery && (
                <button
                  onClick={() => {
                    setSearchQuery('')
                    searchInputRef.current?.focus()
                  }}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
                >
                  <X size={14} />
                </button>
              )}
            </div>
          </div>

          {/* 材料列表 */}
          <div className="flex-1 overflow-y-auto">
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
                <span className="ml-2 text-xs text-slate-500">加载中...</span>
              </div>
            ) : filteredMaterials.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8">
                <FileJson size={32} className="text-slate-300 mb-2" />
                <p className="text-sm text-slate-400">未找到匹配的材料</p>
              </div>
            ) : (
              filteredMaterials.map((material) => (
                <MaterialItem
                  key={material.id}
                  material={material}
                  searchQuery={searchQuery}
                  isSelected={value === material.name}
                  onSelect={handleSelect}
                />
              ))
            )}
          </div>

          {/* 列表底部 */}
          <div className="p-2 border-t border-slate-100 bg-slate-50 text-xs text-slate-400 flex justify-between">
            <span>共 {filteredMaterials.length} 个材料</span>
            {value && <span className="text-blue-600">已选择</span>}
          </div>
        </div>
      )}
    </div>
  )
})
