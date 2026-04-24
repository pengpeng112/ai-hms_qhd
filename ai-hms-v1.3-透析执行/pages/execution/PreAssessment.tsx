
import React from 'react';
import { Patient } from '../../types';
import { 
  ChevronDown, 
  Clock, 
  Info,
  Calendar,
  X,
  Camera,
  MapPin,
  Scale,
  Thermometer,
  Heart,
  Activity,
  Brain,
  Stethoscope
} from 'lucide-react';

interface Props {
  patient: Patient;
}

const PreAssessment: React.FC<Props> = ({ patient }) => {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 pb-10">
      
      {/* 核心评估网格 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        
        {/* 体重组 */}
        <div className="lg:col-span-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 bg-white p-6 rounded-[32px] border border-slate-100 shadow-sm">
          <div className="lg:col-span-4 flex items-center space-x-2 mb-2">
            <Scale size={18} className="text-blue-500" />
            <h3 className="text-sm font-black text-slate-800 uppercase tracking-widest">体重与容量评估</h3>
          </div>
          
          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">
              <span className="text-red-500 mr-1">*</span>透前体重
            </label>
            <div className="relative group">
              <div className="flex items-center w-full h-12 border border-slate-200 rounded-2xl bg-slate-50/30 group-focus-within:border-blue-400 group-focus-within:ring-4 group-focus-within:ring-blue-50 transition-all overflow-hidden">
                <input type="text" defaultValue="96.6" className="w-full px-4 text-sm font-black text-slate-800 outline-none bg-transparent" />
                <span className="px-4 text-[10px] text-slate-400 font-black bg-slate-100/50 h-full flex items-center border-l border-slate-200 shrink-0">KG</span>
              </div>
              <label className="absolute -bottom-5 right-1 flex items-center space-x-1.5 text-[10px] text-slate-400 cursor-pointer hover:text-blue-500 transition-colors">
                <input type="checkbox" className="w-3.5 h-3.5 rounded-md border-slate-300 text-blue-600 focus:ring-blue-500" />
                <span className="font-bold">患者拒测</span>
              </label>
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">干体重</label>
            <div className="flex items-center w-full h-12 border border-slate-100 rounded-2xl bg-slate-100/30 overflow-hidden">
              <input type="text" defaultValue="70.5" disabled className="w-full px-4 text-sm font-black text-slate-300 outline-none bg-transparent cursor-not-allowed" />
              <span className="px-4 text-[10px] text-slate-300 font-black border-l border-slate-100 h-full flex items-center">KG</span>
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">体重增量</label>
            <div className="flex items-center w-full h-12 border border-red-100 rounded-2xl bg-red-50/30 overflow-hidden">
              <input type="text" defaultValue="2.1" className="w-full px-4 text-sm font-black text-red-600 outline-none bg-transparent" />
              <span className="px-4 text-[10px] text-red-300 font-black border-l border-red-100 h-full flex items-center shrink-0">KG</span>
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">
              <span className="text-red-500 mr-1">*</span>超滤量
            </label>
            <div className="flex items-center w-full h-12 border border-slate-200 rounded-2xl bg-white focus-within:border-blue-400 focus-within:ring-4 focus-within:ring-blue-50 transition-all overflow-hidden">
              <input type="text" defaultValue="2.3" className="w-full px-4 text-sm font-black text-slate-800 outline-none" />
              <span className="px-4 text-[10px] text-slate-400 font-black bg-slate-50 h-full flex items-center border-l border-slate-200 shrink-0">L</span>
            </div>
          </div>
        </div>

        {/* 生命体征组 */}
        <div className="lg:col-span-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 bg-white p-6 rounded-[32px] border border-slate-100 shadow-sm">
          <div className="lg:col-span-4 flex items-center space-x-2 mb-2">
            <Activity size={18} className="text-rose-500" />
            <h3 className="text-sm font-black text-slate-800 uppercase tracking-widest">生命体征监测</h3>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">
              <span className="text-red-500 mr-1">*</span>透前血压
            </label>
            <div className="flex items-center space-x-2">
              <div className="flex-1 flex items-center h-12 border border-slate-200 rounded-2xl bg-white focus-within:border-rose-400 focus-within:ring-4 focus-within:ring-rose-50 transition-all overflow-hidden">
                <input type="text" defaultValue="176" className="w-full text-center text-sm font-black text-slate-800 outline-none" />
                <span className="text-slate-300 font-light">/</span>
                <input type="text" defaultValue="71" className="w-full text-center text-sm font-black text-slate-800 outline-none" />
              </div>
              <span className="text-[10px] text-slate-400 font-black shrink-0">mmHg</span>
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">测压部位</label>
            <div className="relative">
              <select className="w-full h-12 px-4 border border-slate-200 rounded-2xl text-sm font-black text-slate-800 outline-none bg-slate-50/30 appearance-none focus:bg-white focus:border-blue-400 transition-all">
                <option>右上肢</option>
                <option>左上肢</option>
              </select>
              <ChevronDown size={16} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">
              <span className="text-red-500 mr-1">*</span>透前心率
            </label>
            <div className="flex items-center w-full h-12 border border-slate-200 rounded-2xl bg-white focus-within:border-blue-400 focus-within:ring-4 focus-within:ring-blue-50 transition-all overflow-hidden">
              <div className="pl-4 text-rose-400"><Heart size={14} /></div>
              <input type="text" defaultValue="82" className="w-full px-3 text-sm font-black text-slate-800 outline-none" />
              <span className="px-4 text-[10px] text-slate-400 font-black bg-slate-50 h-full flex items-center border-l border-slate-200 shrink-0">次/分</span>
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">
              <span className="text-red-500 mr-1">*</span>透前体温
            </label>
            <div className="flex items-center w-full h-12 border border-slate-200 rounded-2xl bg-white focus-within:border-blue-400 focus-within:ring-4 focus-within:ring-blue-50 transition-all overflow-hidden">
              <div className="pl-4 text-orange-400"><Thermometer size={14} /></div>
              <input type="text" defaultValue="36.5" className="w-full px-3 text-sm font-black text-slate-800 outline-none" />
              <span className="px-4 text-[10px] text-slate-400 font-black bg-slate-50 h-full flex items-center border-l border-slate-200 shrink-0">°C</span>
            </div>
          </div>
        </div>

        {/* 通路与状态组 */}
        <div className="lg:col-span-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 bg-white p-6 rounded-[32px] border border-slate-100 shadow-sm">
          <div className="lg:col-span-4 flex items-center space-x-2 mb-2">
            <Stethoscope size={18} className="text-emerald-500" />
            <h3 className="text-sm font-black text-slate-800 uppercase tracking-widest">血管通路与神志状态</h3>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1 flex items-center">
              A端位点
              <button className="ml-2 p-1 bg-blue-50 text-blue-600 rounded-lg hover:bg-blue-100 transition-colors" title="查看穿刺点位图">
                <MapPin size={10} />
              </button>
            </label>
            <div className="relative">
              <select className="w-full h-12 px-4 border border-slate-200 rounded-2xl text-sm font-black text-slate-800 outline-none bg-white appearance-none focus:border-blue-400 transition-all">
                <option>1号位</option>
                <option>2号位</option>
                <option>3号位</option>
              </select>
              <ChevronDown size={16} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">V端位点</label>
            <div className="relative">
              <select className="w-full h-12 px-4 border border-slate-200 rounded-2xl text-sm font-black text-slate-800 outline-none bg-white appearance-none focus:border-blue-400 transition-all">
                <option>1号位</option>
                <option>2号位</option>
                <option>3号位</option>
              </select>
              <ChevronDown size={16} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1 flex items-center">
              <Brain size={12} className="mr-1 text-purple-400" />
              神志状态
            </label>
            <div className="relative">
              <select className="w-full h-12 px-4 border border-slate-200 rounded-2xl text-sm font-black text-slate-800 outline-none bg-white appearance-none focus:border-blue-400 transition-all">
                <option>清醒 (Conscious)</option>
                <option>模糊 (Blurred)</option>
                <option>昏迷 (Coma)</option>
              </select>
              <ChevronDown size={16} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" />
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">护理分级</label>
            <div className="flex items-center h-12 px-4 bg-slate-50/50 rounded-2xl border border-slate-100 space-x-4">
              {['病危', '病重', '其他'].map((item) => (
                <label key={item} className="flex items-center space-x-2 cursor-pointer group">
                  <input type="radio" name="nurse_grade_final" defaultChecked={item === '其他'} className="w-4 h-4 text-blue-600 border-slate-300 focus:ring-blue-500" />
                  <span className="text-xs font-bold text-slate-600 group-hover:text-blue-600 transition-colors">{item}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="lg:col-span-2 space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">内瘘情况描述</label>
            <div className="flex-1 min-h-[48px] p-2 border border-slate-200 rounded-2xl flex flex-wrap gap-2 bg-white focus-within:border-blue-400 transition-all">
              {['杂音强', '震颤强', '搏动强'].map(tag => (
                <span key={tag} className="flex items-center px-3 py-1 bg-blue-50 text-blue-700 rounded-xl text-[10px] font-black border border-blue-100">
                  {tag} <X size={10} className="ml-2 text-blue-300 cursor-pointer hover:text-blue-500" />
                </span>
              ))}
              <input className="flex-1 min-w-[80px] outline-none text-xs font-medium bg-transparent px-2" placeholder="输入更多描述..." />
            </div>
          </div>

          <div className="lg:col-span-2 space-y-2">
            <label className="block text-[11px] font-black text-slate-400 uppercase tracking-wider ml-1">透前症状记录</label>
            <div className="flex-1 min-h-[48px] p-2 border border-slate-200 rounded-2xl flex flex-wrap gap-2 bg-white focus-within:border-blue-400 transition-all">
              <span className="flex items-center px-3 py-1 bg-slate-100 text-slate-700 rounded-xl text-[10px] font-black border border-slate-200">
                无明显症状 <X size={10} className="ml-2 text-slate-400 cursor-pointer hover:text-slate-600" />
              </span>
              <input className="flex-1 min-w-[80px] outline-none text-xs font-medium bg-transparent px-2" placeholder="录入新症状..." />
            </div>
          </div>
        </div>
      </div>

      {/* 底部流程人员区 */}
      <div className="mt-10 p-8 bg-slate-900 rounded-[40px] shadow-2xl relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-8 opacity-5 group-hover:scale-110 transition-transform duration-1000">
          <Clock size={120} className="text-white" />
        </div>
        
        <div className="relative z-10 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div className="space-y-2">
            <label className="block text-[10px] font-black text-slate-500 uppercase tracking-[2px]">治疗开始时间</label>
            <div className="flex items-center space-x-3 text-white">
              <Calendar size={16} className="text-blue-400" />
              <span className="text-sm font-black">2026-02-04 07:50</span>
            </div>
          </div>
          
          <div className="space-y-2">
            <label className="block text-[10px] font-black text-slate-500 uppercase tracking-[2px]">接诊医生</label>
            <div className="relative">
              <select className="w-full bg-transparent text-white text-sm font-black outline-none appearance-none cursor-pointer hover:text-blue-400 transition-colors">
                <option className="bg-slate-800">董婉颖 (主任医师)</option>
                <option className="bg-slate-800">王志远 (主治医师)</option>
              </select>
              <ChevronDown size={14} className="absolute right-0 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none" />
            </div>
          </div>

          <div className="lg:col-span-2 flex items-center justify-end space-x-3">
             <div className="text-right">
                <p className="text-[10px] font-black text-slate-500 uppercase tracking-widest">最后评估人</p>
                <p className="text-xs font-bold text-blue-400">郑九盈 (2026-02-04 07:38)</p>
             </div>
             <div className="w-10 h-10 rounded-full bg-blue-500/20 border border-blue-500/30 flex items-center justify-center text-blue-400 font-black text-xs">郑</div>
          </div>
        </div>
      </div>

      {/* 底部悬浮功能条 */}
      <div className="flex items-center justify-between mt-10 py-6 bg-white/80 backdrop-blur-xl sticky bottom-0 border-t border-slate-100 -mx-8 px-8 z-50">
        <div className="flex items-center space-x-8">
          <button className="flex items-center space-x-2 text-slate-600 hover:text-blue-600 transition-all group">
            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center group-hover:bg-blue-50 transition-colors">
              <Camera size={16} className="group-hover:scale-110 transition-transform" />
            </div>
            <span className="text-[11px] font-black uppercase tracking-wider">称重照片存档</span>
          </button>
          <div className="flex items-center space-x-2 text-[11px] text-slate-400 font-bold">
            <Info size={14} className="text-blue-400" />
            <span>所有评估项均已通过逻辑校验</span>
          </div>
        </div>
        <div className="flex space-x-4">
          <button className="px-8 py-4 border border-slate-200 text-slate-400 text-[12px] font-black rounded-2xl hover:bg-slate-50 transition-all">暂存草稿</button>
          <button className="px-12 py-4 bg-blue-600 text-white text-[12px] font-black rounded-2xl hover:bg-blue-700 shadow-xl shadow-blue-100 transition-all transform active:scale-[0.98]">提交并开始治疗</button>
        </div>
      </div>
    </div>
  );
};

export default PreAssessment;
