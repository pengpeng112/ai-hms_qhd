// OutcomeHistoryModal - 治疗转归记录历史列表弹窗

import { ListFilter, X, Edit3, Trash2 } from 'lucide-react'
import type { OutcomeRecord } from './types'

interface OutcomeHistoryModalProps {
  isOpen: boolean
  onClose: () => void
  records: OutcomeRecord[]
  onEdit: (record: OutcomeRecord) => void
  onDelete: (id: string) => void
}

export default function OutcomeHistoryModal({ isOpen, onClose, records, onEdit, onDelete }: OutcomeHistoryModalProps) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-4xl overflow-hidden animate-scale-in border border-slate-100 flex flex-col max-h-[85vh]">
        <div className="bg-[#eef6ff] px-10 py-6 flex items-center justify-between border-b border-blue-100 shrink-0">
          <h3 className="text-lg font-black text-slate-800 flex items-center gap-2">
            <ListFilter size={20} className="text-blue-600" /> 转归记录历史列表
          </h3>
          <button onClick={onClose} className="p-1.5 hover:bg-white/50 rounded-full text-slate-400 hover:text-slate-600 transition-all">
            <X size={20} />
          </button>
        </div>
        <div className="p-8 overflow-y-auto flex-1 custom-scrollbar">
          <table className="w-full text-left text-sm border-collapse">
            <thead className="bg-slate-50 text-slate-400 font-black text-[10px] border-b border-slate-100 uppercase tracking-widest sticky top-0 z-10">
              <tr>
                <th className="py-4 px-4 w-24">类型</th>
                <th className="py-4 px-4 w-32">时间</th>
                <th className="py-4 px-4">原因</th>
                <th className="py-4 px-4 w-20">门规</th>
                <th className="py-4 px-4">登记人</th>
                <th className="py-4 px-4 text-right">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50">
              {records.length > 0 ? (
                records.map((record) => (
                  <tr key={record.id} className="hover:bg-slate-50 transition-colors">
                    <td className="py-4 px-4">
                      <span
                        className={`px-2.5 py-0.5 rounded-lg text-[10px] font-black ${
                          record.typeName === '在科'
                            ? 'bg-green-50 text-green-600 border border-green-100'
                            : 'bg-orange-50 text-orange-600 border border-orange-100'
                        }`}
                      >
                        {record.typeName || '未知'}
                      </span>
                    </td>
                    <td className="py-4 px-4 font-mono text-xs text-slate-500">{record.time}</td>
                    <td className="py-4 px-4 font-bold text-slate-700">{record.reasonName || '未知'}</td>
                    <td className="py-4 px-4">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-black ${
                          record.isDoorRule
                            ? 'bg-purple-50 text-purple-600 border border-purple-200'
                            : 'bg-slate-100 text-slate-400 border border-slate-200'
                        }`}
                      >
                        {record.isDoorRule ? '是' : '否'}
                      </span>
                    </td>
                    <td className="py-4 px-4 text-slate-600">{record.registrar}</td>
                    <td className="py-4 px-4 text-right">
                      <div className="flex justify-end gap-2">
                        <button onClick={() => onEdit(record)} className="p-2 text-blue-600 hover:bg-blue-50 rounded-xl transition-all" title="编辑">
                          <Edit3 size={16} />
                        </button>
                        <button onClick={() => onDelete(record.id)} className="p-2 text-red-400 hover:bg-red-50 hover:text-red-600 rounded-xl transition-all" title="删除">
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={6} className="py-20 text-center text-slate-300 font-bold italic">
                    暂无转归记录数据
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="p-6 bg-slate-50 border-t border-slate-100 flex justify-end shrink-0">
          <button onClick={onClose} className="px-10 py-2.5 bg-slate-900 text-white rounded-2xl text-sm font-black hover:bg-slate-800 transition-all">
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}
