// ElectronicDocumentModal - 新增电子文书弹窗

import { Files, X, ChevronDown, PenTool, Upload, Save } from 'lucide-react'

interface ElectronicDocumentModalProps {
  isOpen: boolean
  onClose: () => void
}

export default function ElectronicDocumentModal({ isOpen, onClose }: ElectronicDocumentModalProps) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[250] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-2xl overflow-hidden animate-scale-in border border-slate-100 flex flex-col ring-1 ring-black/5">
        <div className="bg-[#eef6ff] px-10 py-6 flex items-center justify-between border-b border-blue-100">
          <h3 className="text-lg font-black text-slate-800 flex items-center gap-2">
            <Files size={20} className="text-blue-600" /> 新增电子文书
          </h3>
          <button
            onClick={onClose}
            className="p-1.5 hover:bg-white/50 rounded-full text-slate-400 hover:text-slate-600 transition-all"
          >
            <X size={20} />
          </button>
        </div>
        <div className="p-10 space-y-6">
          <div className="grid grid-cols-2 gap-6">
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                <span className="text-red-500 mr-1">*</span>文书类别
              </label>
              <div className="relative">
                <select className="w-full h-11 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20 bg-white appearance-none transition-all">
                  <option value="">请选择文书类别</option>
                  <option value="知情同意书">知情同意书</option>
                  <option value="入院评估单">入院评估单</option>
                  <option value="治疗告知单">治疗告知单</option>
                  <option value="风险告知书">风险告知书</option>
                  <option value="检查报告附件">检查报告附件</option>
                  <option value="其他">其他</option>
                </select>
                <ChevronDown size={14} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
              </div>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                <span className="text-red-500 mr-1">*</span>签字时间
              </label>
              <input
                type="datetime-local"
                defaultValue={new Date().toISOString().slice(0, 16)}
                className="w-full h-11 px-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20 transition-all"
              />
            </div>
            <div className="col-span-2 flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                <span className="text-red-500 mr-1">*</span>患者/家属签字
              </label>
              <div className="relative">
                <input
                  type="text"
                  placeholder="请输入签字姓名或在此区域手写模拟"
                  className="w-full h-11 px-11 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20 transition-all"
                />
                <PenTool size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
              </div>
            </div>
            <div className="col-span-2 flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">备注信息</label>
              <textarea
                className="w-full h-24 p-3 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/20 resize-none transition-all"
                placeholder="请输入相关的备注说明内容"
              />
            </div>
            <div className="col-span-2 flex flex-col gap-1.5">
              <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest">附件上传</label>
              <div className="border-2 border-dashed border-slate-100 rounded-2xl p-8 flex flex-col items-center justify-center gap-3 bg-slate-50/50 hover:bg-slate-50 hover:border-blue-200 transition-all cursor-pointer group">
                <div className="w-12 h-12 rounded-full bg-white flex items-center justify-center text-slate-300 shadow-sm group-hover:text-blue-500 group-hover:scale-110 transition-all">
                  <Upload size={24} />
                </div>
                <div className="text-center">
                  <p className="text-sm font-bold text-slate-500 group-hover:text-slate-700">点击或将文件拖拽至此上传</p>
                  <p className="text-[10px] text-slate-400 mt-1 uppercase">支持 JPG, PDF, PNG 格式 (最大 20MB)</p>
                </div>
              </div>
            </div>
          </div>
          <div className="mt-8 flex justify-end gap-3 pt-4 border-t border-slate-50">
            <button
              onClick={onClose}
              className="px-8 py-2.5 border border-slate-200 text-slate-500 rounded-2xl text-sm font-black hover:bg-slate-50 transition-all"
            >
              取消
            </button>
            <button
              onClick={() => {
                alert('保存成功')
                onClose()
              }}
              className="px-10 py-2.5 bg-blue-600 text-white rounded-2xl text-sm font-black hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all flex items-center gap-2"
            >
              <Save size={16} /> 确认提交
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
