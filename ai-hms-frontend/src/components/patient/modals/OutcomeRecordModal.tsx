// OutcomeRecordModal - 治疗转归记录填写弹窗

import { useState, useEffect } from 'react'
import { Heart, X, Save, Loader2, ChevronDown } from 'lucide-react'
import { DatePicker } from 'antd'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import type { OutcomeRecord } from './types'
import { useOutcomeDict } from '@/hooks/useOutcomeDict'

dayjs.locale('zh-cn')

interface OutcomeRecordModalProps {
  isOpen: boolean
  onClose: () => void
  onSave: (data: OutcomeRecord) => void
  initialData?: OutcomeRecord | null
  saving?: boolean
}

const formatDateTime = (date: Date): string => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const createDefaultFormData = (): Partial<OutcomeRecord> => ({
  type: '',
  reason: '',
  time: formatDateTime(new Date()),
  remarks: '',
  registrar: '',
  registrationTime: formatDateTime(new Date()),
  isDoorRule: false,
})

const selectClassName = "w-full h-11 border border-slate-200 rounded-xl px-3 text-sm font-bold text-slate-700 appearance-none outline-none focus:ring-2 focus:ring-blue-500/20 bg-white"

export default function OutcomeRecordModal({ isOpen, onClose, onSave, initialData, saving = false }: OutcomeRecordModalProps) {
  const { typeOptions, getReasonOptions, loading: dictLoading, loadDicts } = useOutcomeDict()

  // 弹窗打开时加载字典数据
  useEffect(() => {
    if (isOpen) {
      loadDicts()
    }
  }, [isOpen, loadDicts])

  // 表单状态：父组件通过 key 变化保证每次打开弹窗都重新挂载，useState 初始化器自然生效
  const [formData, setFormData] = useState<Partial<OutcomeRecord>>(() => {
    if (initialData) {
      return {
        ...initialData,
        type: initialData.type != null ? String(initialData.type) : '',
        reason: initialData.reason != null ? String(initialData.reason) : '',
      }
    }
    return createDefaultFormData()
  })

  // 联动逻辑：当转归类型变化时，清空已选的转归原因
  const handleTypeChange = (newTypeCode: string) => {
    setFormData(prev => ({
      ...prev,
      type: newTypeCode,
      reason: '',  // 清空原因
    }))
  }

  // 获取当前类型对应的原因选项
  const selectedTypeCode = formData.type != null ? String(formData.type) : ''
  const availableReasons = getReasonOptions(selectedTypeCode)

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[250] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-xl overflow-hidden animate-scale-in border border-slate-100">
        <div className="bg-[#eef6ff] px-10 py-6 flex items-center justify-between border-b border-blue-100">
          <h3 className="text-lg font-black text-slate-800 flex items-center gap-2">
            <Heart size={20} className="text-blue-600" /> {initialData ? '编辑转归记录' : '新增转归记录'}
          </h3>
          <button onClick={onClose} className="p-1.5 hover:bg-white/50 rounded-full text-slate-400 hover:text-slate-600 transition-all">
            <X size={20} />
          </button>
        </div>
        <div className="p-10 space-y-6">
          <div className="grid grid-cols-2 gap-6">
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                <span className="text-red-500 mr-1">*</span>转归类型
              </label>
              <div className="relative">
                <select
                  value={selectedTypeCode}
                  onChange={(e) => handleTypeChange(e.target.value)}
                  disabled={dictLoading}
                  className={selectClassName}
                >
                  <option value="">请选择转归类型</option>
                  {typeOptions.map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
                <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
              </div>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                <span className="text-red-500 mr-1">*</span>转归时间
              </label>
              <DatePicker
                showTime
                value={formData.time ? dayjs(formData.time) : null}
                onChange={(date) => setFormData({ ...formData, time: date ? date.format('YYYY-MM-DD HH:mm') : '' })}
                format="YYYY年MM月DD日 HH:mm"
                placeholder="请选择转归时间"
                getPopupContainer={(trigger) => trigger.closest('.fixed') as HTMLElement || document.body}
                style={{ width: '100%', height: 44 }}
                size="middle"
              />
            </div>
            <div className="col-span-2 flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                <span className="text-red-500 mr-1">*</span>转归原因
              </label>
              <div className="relative">
                <select
                  value={formData.reason != null ? String(formData.reason) : ''}
                  onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
                  disabled={dictLoading || !selectedTypeCode}
                  className={selectClassName}
                >
                  <option value="">{selectedTypeCode ? '请选择转归原因' : '请先选择转归类型'}</option>
                  {availableReasons.map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
                <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
              </div>
            </div>
            <div className="col-span-2 flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">备注</label>
              <textarea
                value={formData.remarks}
                onChange={(e) => setFormData({ ...formData, remarks: e.target.value })}
                className="p-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20 h-24 resize-none"
                placeholder="如有其他说明请在此输入"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">登记人</label>
              <input
                type="text"
                value={formData.registrar}
                onChange={(e) => setFormData({ ...formData, registrar: e.target.value })}
                placeholder="请输入登记人姓名"
                className="h-11 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">登记时间</label>
              <input
                type="text"
                value={formData.registrationTime}
                readOnly
                className="h-11 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-400 bg-slate-50 outline-none"
              />
            </div>
            <div className="col-span-2 flex items-center gap-3 px-3 py-3 bg-slate-50 rounded-xl border border-slate-200">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={formData.isDoorRule ?? false}
                  onChange={(e) => setFormData({ ...formData, isDoorRule: e.target.checked })}
                  className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm font-bold text-slate-700">是否门规</span>
              </label>
            </div>
          </div>
          <div className="mt-8 flex justify-end gap-3">
            <button onClick={onClose} disabled={saving} className="px-8 py-2.5 border border-slate-200 text-slate-500 rounded-2xl text-sm font-black hover:bg-slate-50 transition-all disabled:opacity-50">
              取消
            </button>
            <button
              onClick={() => onSave(formData as OutcomeRecord)}
              disabled={saving}
              className="px-10 py-2.5 bg-blue-600 text-white rounded-2xl text-sm font-black hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all flex items-center gap-2 disabled:opacity-50"
            >
              {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
              {saving ? '保存中...' : '确定'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
