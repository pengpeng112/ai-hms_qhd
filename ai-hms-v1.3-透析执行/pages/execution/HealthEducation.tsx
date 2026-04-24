
import React, { useState } from 'react';
import { Patient } from '../../types';
import { 
  Calendar, 
  ChevronDown, 
  FileText, 
  ClipboardCheck, 
  Save, 
  BookOpen,
  User,
  Zap,
  Clock,
  MessageCircle
} from 'lucide-react';

interface Props {
  patient: Patient;
}

const HealthEducation: React.FC<Props> = ({ patient }) => {
  const [formData, setFormData] = useState({
    content: '为什么有些透析患者容易出现高钾',
    method: '口头宣教',
    date: '2026-02-09',
    educator: '李聪',
    description: ''
  });

  return (
    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500 pb-10">
      


      {/* 2. 主录入区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-10">
        
        {/* 左侧：核心参数 (4 cols) */}
        <div className="lg:col-span-5 space-y-8 bg-white border border-slate-100 p-8 rounded-[32px] shadow-sm">
          <div className="flex items-center space-x-2 border-l-4 border-blue-500 pl-3 mb-2">
            <span className="text-[12px] font-black text-slate-800 uppercase tracking-widest">核心宣教参数</span>
          </div>

          <div className="space-y-6">
            <div className="space-y-2">
              <label className="text-[11px] font-bold text-slate-400 uppercase flex items-center">
                <span className="text-red-500 mr-1">*</span>宣教内容题目
              </label>
              <div className="relative group">
                <select 
                  value={formData.content}
                  onChange={(e) => setFormData({...formData, content: e.target.value})}
                  className="w-full h-12 px-5 border border-slate-200 rounded-2xl text-[13px] font-black text-slate-700 outline-none bg-slate-50/50 appearance-none focus:bg-white focus:border-blue-400 focus:ring-4 focus:ring-blue-50 transition-all"
                >
                  <option>为什么有些透析患者容易出现高钾</option>
                  <option>血液透析饮食指导</option>
                  <option>内瘘自我护理要点</option>
                  <option>透析期间体重管理</option>
                  <option>血管通路的日常维护</option>
                </select>
                <ChevronDown size={16} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none group-hover:text-blue-400 transition-colors" />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-[11px] font-bold text-slate-400 uppercase">宣教方式</label>
              <div className="relative">
                <input 
                  type="text" 
                  value={formData.method}
                  onChange={(e) => setFormData({...formData, method: e.target.value})}
                  className="w-full h-12 px-5 border border-slate-200 rounded-2xl text-[13px] font-black text-slate-700 outline-none bg-slate-50/50 focus:bg-white focus:border-blue-400 focus:ring-4 focus:ring-blue-50 transition-all"
                  placeholder="如：口头宣教、发放手册..."
                />
                <MessageCircle size={16} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300" />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-[11px] font-bold text-slate-400 uppercase flex items-center">
                  <span className="text-red-500 mr-1">*</span>宣教日期
                </label>
                <div className="relative">
                  <input 
                    type="text" 
                    value={formData.date}
                    readOnly
                    className="w-full h-12 px-5 border border-slate-200 rounded-2xl text-[13px] font-black text-slate-400 outline-none bg-slate-50/30"
                  />
                  <Calendar size={16} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-200" />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-[11px] font-bold text-slate-400 uppercase flex items-center">
                  <span className="text-red-500 mr-1">*</span>宣教人
                </label>
                <div className="relative">
                  <select 
                    value={formData.educator}
                    onChange={(e) => setFormData({...formData, educator: e.target.value})}
                    className="w-full h-12 px-5 border border-slate-200 rounded-2xl text-[13px] font-black text-slate-700 outline-none bg-slate-50/50 appearance-none focus:bg-white"
                  >
                    <option>李聪</option>
                    <option>刘护士长</option>
                    <option>郑九盈</option>
                  </select>
                  <User size={16} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300" />
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 右侧：宣教内容描述 (8 cols) */}
        <div className="lg:col-span-7 flex flex-col bg-white border border-slate-100 rounded-[32px] shadow-sm overflow-hidden">
          <div className="px-8 py-5 border-b border-slate-50 flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <FileText size={18} className="text-blue-500" />
              <span className="text-[13px] font-black text-slate-800">宣教详情记录描述</span>
            </div>
            <span className="text-[10px] text-slate-300 font-bold uppercase tracking-widest">Detail Narrative</span>
          </div>
          <div className="flex-1 p-8">
            <textarea 
              className="w-full h-full min-h-[300px] text-[14px] font-medium text-slate-600 leading-relaxed outline-none resize-none placeholder:text-slate-200" 
              placeholder="请输入本次宣教的具体内容、患者的反馈或重点强调的注意事项..."
              value={formData.description}
              onChange={(e) => setFormData({...formData, description: e.target.value})}
            ></textarea>
          </div>
          <div className="px-8 py-4 bg-slate-50/50 text-right">
            <span className="text-[10px] text-slate-400 font-black uppercase tracking-tighter italic">Auto-save to draft is active</span>
          </div>
        </div>
      </div>

      {/* 3. 底部操作栏 */}
      <div className="mt-12 pt-8 border-t border-slate-100 flex items-center justify-between">
        <div className="flex items-center space-x-6">
          <div className="flex items-center space-x-3 text-slate-400">
            <Clock size={16} />
            <span className="text-[11px] font-bold">最后更新: 刚刚</span>
          </div>
          <div className="h-4 w-px bg-slate-200"></div>
          <div className="flex items-center space-x-3 text-slate-400">
            <ClipboardCheck size={16} />
            <span className="text-[11px] font-bold">所属病区: 肾内透析中心一病区</span>
          </div>
        </div>

        <div className="flex space-x-4">
          <button className="px-10 py-4 border border-slate-200 text-slate-400 text-[12px] font-black rounded-[20px] hover:bg-slate-50 transition-all">
            重置当前表单
          </button>
          <button className="px-16 py-4 bg-blue-600 text-white text-[12px] font-black rounded-[20px] hover:bg-blue-700 shadow-2xl shadow-blue-100 transition-all transform active:scale-[0.98] flex items-center space-x-3">
            <Save size={18} />
            <span>保存并下发宣教记录</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default HealthEducation;
