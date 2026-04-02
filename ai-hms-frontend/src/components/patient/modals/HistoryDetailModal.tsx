// HistoryDetailModal - 临床病史详细弹窗

import { useState, useEffect } from 'react'
import { FileHeart, X, Save, Loader2, CalendarDays, FileEdit, Trash2 } from 'lucide-react'
import { DatePicker } from 'antd'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import { dictCache, DICT_TYPES } from '@/services/dictApi'
import type { CascaderOption } from '@/services/dictApi'
import type { HistoryContent, HistoryNamedContent } from '@/services/restClient'

dayjs.locale('zh-cn')

type HistoryData = HistoryContent | HistoryNamedContent

// 输血记录条目
interface TransfusionRecord {
  id: number
  time: string
  desc: string
}

// 解析输血史 content 字段为结构化记录
function parseTransfusionRecords(content: string): TransfusionRecord[] {
  if (!content) return []
  try {
    const parsed = JSON.parse(content)
    if (Array.isArray(parsed)) return parsed
  } catch {
    // 兼容旧的纯文本格式：转为单条记录
    if (content.trim()) {
      return [{ id: 1, time: '', desc: content.trim() }]
    }
  }
  return []
}

// 专科 key → 字典类型映射
const HISTORY_KEY_TO_DICT: Record<string, string> = {
  primary: DICT_TYPES.PRIMARY_DISEASE,
  pathology: DICT_TYPES.PATHOLOGY,
  allergen: DICT_TYPES.ALLERGEN,
  tumor: DICT_TYPES.TUMOR,
  complication: DICT_TYPES.COMPLICATION,
}

interface HistoryDetailModalProps {
  isOpen: boolean
  onClose: () => void
  onSave: (newData: HistoryData) => void | Promise<void>
  title: string
  historyKey: string
  data: HistoryData
  saving?: boolean
}

export default function HistoryDetailModal({ isOpen, onClose, onSave, title, historyKey, data, saving = false }: HistoryDetailModalProps) {
  const [formData, setFormData] = useState<HistoryData>(data)
  const [cascaderOptions, setCascaderOptions] = useState<CascaderOption[]>([])
  const [selectedParent, setSelectedParent] = useState<string>('')

  // 输血史专用状态
  const isTransfusion = historyKey === 'transfusion'
  const [transfusionRecords, setTransfusionRecords] = useState<TransfusionRecord[]>(() =>
    isTransfusion ? parseTransfusionRecords((data as HistoryContent).content) : []
  )
  const [inputTime, setInputTime] = useState('')
  const [inputDesc, setInputDesc] = useState('')

  // 加载字典级联选项
  useEffect(() => {
    if (!isOpen) return
    const dictType = HISTORY_KEY_TO_DICT[historyKey]
    if (!dictType) return

    dictCache.getCascaderOptions(dictType).then(options => {
      setCascaderOptions(options)
      // 如果已有 type 值，自动定位到对应的父级
      const named = data as HistoryNamedContent
      if (named.type) {
        // 找到匹配的父级：type 存的是父级 name
        const parent = options.find(opt => opt.label === named.type)
        if (parent) {
          setSelectedParent(parent.value)
        }
      }
    }).catch(() => {
      setCascaderOptions([])
    })
  }, [isOpen, historyKey, data])

  if (!isOpen) return null

  const handleChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }) as HistoryData)
  }

  const isNamedContent = 'name' in formData

  // 判断当前字典是否有二级结构
  const hasChildren = cascaderOptions.some(opt => opt.children && opt.children.length > 0)

  // 获取当前选中父级的子项列表
  const childOptions = selectedParent
    ? cascaderOptions.find(opt => opt.value === selectedParent)?.children || []
    : []

  // 处理父级选择变更
  const handleParentChange = (parentValue: string) => {
    setSelectedParent(parentValue)
    const parent = cascaderOptions.find(opt => opt.value === parentValue)
    if (parent) {
      if (parent.children && parent.children.length > 0) {
        // 有子项时：type 设为父级名称，清空 name 等用户重新选
        handleChange('type', parent.label)
        handleChange('name', '')
      } else {
        // 独立项（无子项）：type 设为该项名称，name 保留让用户自定义输入
        handleChange('type', parent.label)
        handleChange('name', '')
      }
    } else {
      handleChange('type', '')
      handleChange('name', '')
    }
  }

  // 处理子项选择变更
  const handleChildChange = (childValue: string) => {
    const child = childOptions.find(opt => opt.value === childValue)
    if (child) {
      handleChange('name', child.label)
    }
  }

  // 处理扁平字典（无二级结构）的选择
  const handleFlatSelect = (value: string) => {
    const opt = cascaderOptions.find(o => o.value === value)
    if (opt) {
      handleChange('type', '')
      handleChange('name', opt.label)
    }
  }

  const renderContent = () => {
    // 输血史：结构化多条记录模式
    if (isTransfusion) {
      const handleAddRecord = () => {
        if (!inputTime || !inputDesc) return
        const newRecord: TransfusionRecord = {
          id: transfusionRecords.length > 0 ? Math.max(...transfusionRecords.map(r => r.id)) + 1 : 1,
          time: inputTime,
          desc: inputDesc,
        }
        setTransfusionRecords(prev => [newRecord, ...prev])
        setInputTime('')
        setInputDesc('')
      }

      const handleDeleteRecord = (id: number) => {
        setTransfusionRecords(prev => prev.filter(r => r.id !== id))
      }

      return (
        <div className="space-y-8">
          {/* 表单区 */}
          <div className="space-y-4">
            <label className="text-xs font-black text-slate-400 uppercase tracking-widest px-1">内容描述</label>
            <div className="bg-slate-50 border border-slate-100 rounded-2xl p-5 space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div className="space-y-1.5">
                  <p className="text-xs font-bold text-blue-500/70 flex items-center gap-1.5 px-1">
                    <CalendarDays size={14} />
                    输血时间
                  </p>
                  <DatePicker
                    showTime
                    value={inputTime ? dayjs(inputTime) : null}
                    onChange={(date) => setInputTime(date ? date.format('YYYY-MM-DD HH:mm') : '')}
                    format="YYYY年MM月DD日 HH:mm"
                    placeholder="请选择输血时间"
                    style={{ width: '100%', height: 40 }}
                    size="middle"
                  />
                </div>
                <div className="space-y-1.5">
                  <p className="text-xs font-bold text-blue-500/70 flex items-center gap-1.5 px-1">
                    <FileEdit size={14} />
                    输血详情
                  </p>
                  <input
                    type="text"
                    placeholder="请输入输血成分、剂量等描述信息"
                    value={inputDesc}
                    onChange={(e) => setInputDesc(e.target.value)}
                    onKeyDown={(e) => { if (e.key === 'Enter') handleAddRecord() }}
                    className="w-full h-10 bg-white border border-slate-200 rounded-xl px-4 text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <button
                  onClick={handleAddRecord}
                  disabled={!inputTime || !inputDesc}
                  className="px-5 py-2 bg-blue-600 text-white rounded-xl text-xs font-black hover:bg-blue-700 shadow-sm transition-all disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  添加记录
                </button>
              </div>
            </div>
          </div>

          {/* 记录列表 */}
          <div className="space-y-4">
            <div className="flex items-center justify-between px-1">
              <h4 className="text-xs font-black text-slate-400 uppercase tracking-widest">输血记录列表</h4>
              <span className="text-xs bg-blue-50 text-blue-600 px-3 py-1 rounded-full font-bold">
                已保存 {transfusionRecords.length} 条
              </span>
            </div>
            <div className="rounded-2xl border border-slate-100 overflow-hidden shadow-sm">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-slate-50/50">
                    <th className="px-5 py-3.5 text-[10px] font-black text-slate-400 uppercase tracking-wider w-12 text-center">#</th>
                    <th className="px-5 py-3.5 text-[10px] font-black text-slate-400 uppercase tracking-wider w-44">时间</th>
                    <th className="px-5 py-3.5 text-[10px] font-black text-slate-400 uppercase tracking-wider">描述</th>
                    <th className="px-5 py-3.5 text-[10px] font-black text-slate-400 uppercase tracking-wider w-16 text-center">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {transfusionRecords.map((record, index) => (
                    <tr key={record.id} className="hover:bg-blue-50/20 transition-colors group">
                      <td className="px-5 py-4 text-sm text-slate-400 font-medium text-center">{transfusionRecords.length - index}</td>
                      <td className="px-5 py-4 text-sm text-slate-800 font-bold">{record.time}</td>
                      <td className="px-5 py-4 text-sm text-slate-600 leading-relaxed font-medium">{record.desc}</td>
                      <td className="px-5 py-4 text-center">
                        <button
                          onClick={() => handleDeleteRecord(record.id)}
                          className="p-1.5 text-slate-300 hover:text-red-500 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                        >
                          <Trash2 size={14} />
                        </button>
                      </td>
                    </tr>
                  ))}
                  {transfusionRecords.length === 0 && (
                    <tr>
                      <td colSpan={4} className="px-5 py-14 text-center text-slate-300 text-sm">暂无输血病史记录</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )
    }

    // 基础病史类：现病史、既往史、婚育史、家族史、疾病诊断 (仅包含内容描述)
    if (!isNamedContent) {
      return (
        <div className="space-y-4">
          <div className="flex flex-col gap-2">
            <label className="text-xs font-black text-slate-400 uppercase tracking-widest">内容描述</label>
            <textarea
              value={(formData as HistoryContent).content || ''}
              onChange={(e) => handleChange('content', e.target.value)}
              className="w-full h-48 p-4 bg-slate-50 rounded-2xl text-sm font-bold text-slate-700 leading-relaxed border border-slate-200 focus:bg-white focus:ring-2 focus:ring-blue-500/20 outline-none transition-all resize-none"
              placeholder="请输入详细描述内容..."
            />
          </div>
        </div>
      )
    }

    // 专科信息类：原发病、病理、过敏原、肿瘤、并发症 (结构化字段)
    const named = formData as HistoryNamedContent
    return (
      <div className="grid grid-cols-2 gap-x-12 gap-y-6">
        {hasChildren ? (
          <>
            {/* 二级字典：先选父级分类，再选子项 */}
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase">分类<span className="text-red-500 ml-0.5">*</span></label>
              <select
                value={selectedParent}
                onChange={(e) => handleParentChange(e.target.value)}
                className="h-10 px-3 border border-slate-200 rounded-xl text-sm font-bold text-blue-600 outline-none focus:ring-2 focus:ring-blue-500/20 bg-white"
              >
                <option value="">请选择分类</option>
                {cascaderOptions.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase">名称{childOptions.length > 0 && <span className="text-red-500 ml-0.5">*</span>}</label>
              {childOptions.length > 0 ? (
                <select
                  value={childOptions.find(c => c.label === named.name)?.value || ''}
                  onChange={(e) => handleChildChange(e.target.value)}
                  className="h-10 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-800 outline-none focus:ring-2 focus:ring-blue-500/20 bg-white"
                >
                  <option value="">请选择</option>
                  {childOptions.map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              ) : (
                <input
                  type="text"
                  value={named.name || ''}
                  onChange={(e) => handleChange('name', e.target.value)}
                  className="h-10 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-800 outline-none focus:ring-2 focus:ring-blue-500/20 bg-white"
                  placeholder={selectedParent ? '请输入名称（选填）' : '请先选择分类'}
                />
              )}
            </div>
          </>
        ) : (
          <>
            {/* 扁平字典（如 TUMOR）：单级选择 */}
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase">分类<span className="text-red-500 ml-0.5">*</span></label>
              <select
                value={cascaderOptions.find(o => o.label === named.name)?.value || ''}
                onChange={(e) => handleFlatSelect(e.target.value)}
                className="h-10 px-3 border border-slate-200 rounded-xl text-sm font-bold text-blue-600 outline-none focus:ring-2 focus:ring-blue-500/20 bg-white"
              >
                <option value="">请选择分类</option>
                {cascaderOptions.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase">名称</label>
              <input
                type="text"
                value={named.name || ''}
                readOnly
                className="h-10 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-800 outline-none bg-slate-50"
                placeholder="请先选择分类"
              />
            </div>
          </>
        )}

        <div className="col-span-2 flex flex-col gap-1.5">
          <label className="text-[11px] font-black text-slate-400 uppercase">补充说明</label>
          <textarea
            value={(named as HistoryNamedContent & { content?: string }).content || ''}
            onChange={(e) => handleChange('content', e.target.value)}
            className="p-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-600 outline-none focus:ring-2 focus:ring-blue-500/20 h-24 resize-none"
            placeholder="请输入补充说明..."
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-black text-slate-400 uppercase">检查时间</label>
          <DatePicker
            value={named.checkTime ? dayjs(named.checkTime) : null}
            onChange={(date) => handleChange('checkTime', date ? date.format('YYYY-MM-DD') : '')}
            format="YYYY年MM月DD日"
            placeholder="请选择检查时间"
            style={{ width: '100%', height: 40 }}
            size="middle"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-black text-slate-400 uppercase">检查医生</label>
          <input
            type="text"
            value={named.checkDoctor || ''}
            onChange={(e) => handleChange('checkDoctor', e.target.value)}
            className="h-10 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-600 outline-none focus:ring-2 focus:ring-blue-500/20"
            placeholder="请输入检查医生..."
          />
        </div>
      </div>
    )
  }

  // 处理保存
  const handleSave = () => {
    if (isTransfusion) {
      onSave({ content: JSON.stringify(transfusionRecords) })
    } else {
      onSave(formData)
    }
  }

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className={`bg-white rounded-[32px] shadow-2xl w-full overflow-hidden animate-scale-in border border-slate-100 flex flex-col max-h-[90vh] ${isTransfusion ? 'max-w-3xl' : 'max-w-2xl'}`}>
        <div className="bg-[#eef6ff] px-10 py-6 flex items-center justify-between border-b border-blue-100 shrink-0">
          <h3 className="text-lg font-black text-slate-800 flex items-center gap-2">
            <FileHeart size={20} className="text-blue-600" /> {title}
          </h3>
          <button onClick={onClose} className="p-1.5 hover:bg-white/50 rounded-full transition-all text-slate-400 hover:text-slate-600">
            <X size={20} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-10">
          {renderContent()}
        </div>
        <div className="px-10 py-6 flex justify-end gap-3 shrink-0 border-t border-slate-100">
          <button onClick={onClose} disabled={saving} className="px-8 py-2.5 border border-slate-200 text-slate-500 rounded-2xl text-sm font-black hover:bg-slate-50 transition-all disabled:opacity-50">
            取消
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-10 py-2.5 bg-blue-600 text-white rounded-2xl text-sm font-black hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all flex items-center gap-2 disabled:opacity-50"
          >
            {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            {saving ? '保存中...' : '确定'}
          </button>
        </div>
      </div>
    </div>
  )
}
