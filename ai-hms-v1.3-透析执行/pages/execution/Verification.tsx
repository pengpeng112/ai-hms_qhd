
import React, { useState, useEffect } from 'react';
import { Patient } from '../../types';
import { 
  ShieldCheck, 
  UserCheck, 
  CheckCircle, 
  User, 
  Check, 
  ClipboardCheck, 
  Stethoscope, 
  Activity, 
  Shield,
  Droplets,
  Clock,
  ChevronDown
} from 'lucide-react';

interface Props {
  patient: Patient;
}

const Verification: React.FC<Props> = ({ patient }) => {
  const [firstChecked, setFirstChecked] = useState(false);
  const [secondChecked, setSecondChecked] = useState(false);
  
  // 状态管理：核对人与登记人
  const [firstChecker, setFirstChecker] = useState('刘护士');
  const [secondChecker, setSecondChecker] = useState('张护士');
  const [registrant, setRegistrant] = useState('刘护士');

  // 当首次核对人改变时，自动同步登记人（但允许手动修改）
  useEffect(() => {
    setRegistrant(firstChecker);
  }, [firstChecker]);

  const [personnel, setPersonnel] = useState({
    priming: '刘护士',
    puncture: '刘护士',
    machine: '刘护士',
    qc: '王护士长',
    inspection: '陈医生'
  });

  const staffOptions = ['刘护士', '张护士', '王护士长', '陈医生', '李护士'];

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-2 duration-500">
      
      {/* 顶部三项并排布局 */}
      <div className="grid grid-cols-3 gap-6">
        
        {/* 首次核对 */}
        <div className="flex flex-col border border-gray-100 rounded-2xl bg-white shadow-sm hover:shadow-md transition-all overflow-hidden">
          <div className="px-5 py-4 flex items-center space-x-2 border-b border-gray-50">
            <div className="w-5 h-5 bg-blue-50 rounded-full flex items-center justify-center">
              <UserCheck size={14} className="text-blue-600" />
            </div>
            <span className="text-sm font-black text-slate-800 tracking-tight">首次核对</span>
          </div>
          <div className="p-5 flex-1 flex flex-col space-y-5">
            <div className="text-[11px] text-gray-500 font-medium leading-relaxed bg-gray-50/50 p-4 rounded-xl border border-gray-100">
              核对项目：透析模式、处方参数、耗材规格、患者身份、管路连接安全性。
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">核对人</label>
                <div className="relative">
                  <select 
                    value={firstChecker}
                    onChange={(e) => setFirstChecker(e.target.value)}
                    className="w-full text-xs font-bold border border-gray-100 bg-white px-3 py-2 rounded-lg outline-none focus:ring-2 focus:ring-blue-100 focus:border-blue-400 appearance-none transition-all"
                  >
                    {staffOptions.map(staff => <option key={staff}>{staff}</option>)}
                  </select>
                  <ChevronDown size={12} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">核对时间</label>
                <input type="time" defaultValue="08:45" className="w-full text-xs font-bold border border-gray-100 bg-white px-3 py-2 rounded-lg outline-none focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all" />
              </div>
            </div>
            <button 
              onClick={() => setFirstChecked(!firstChecked)}
              className={`w-full py-3 rounded-xl text-xs font-black transition-all transform active:scale-[0.98] ${
                firstChecked ? 'bg-green-500 text-white shadow-lg shadow-green-100' : 'bg-blue-600 text-white hover:bg-blue-700 shadow-lg shadow-blue-100'
              }`}
            >
              {firstChecked ? '已完成首次核对' : '确认并完成核对'}
            </button>
          </div>
        </div>

        {/* 二次核对 */}
        <div className="flex flex-col border border-gray-100 rounded-2xl bg-white shadow-sm hover:shadow-md transition-all overflow-hidden">
          <div className="px-5 py-4 flex items-center space-x-2 border-b border-gray-50">
            <div className="w-5 h-5 bg-orange-50 rounded-full flex items-center justify-center">
              <ShieldCheck size={14} className="text-orange-600" />
            </div>
            <span className="text-sm font-black text-slate-800 tracking-tight">二次核对</span>
          </div>
          <div className="p-5 flex-1 flex flex-col space-y-5">
            <div className="text-[11px] text-gray-500 font-medium leading-relaxed bg-gray-50/50 p-4 rounded-xl border border-gray-100">
              二次核对重点：核实透析参数、处方调整项、耗材效期及批号一致性。
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">核对人</label>
                <div className="relative">
                  <select 
                    value={secondChecker}
                    onChange={(e) => setSecondChecker(e.target.value)}
                    className="w-full text-xs font-bold border border-gray-100 bg-white px-3 py-2 rounded-lg outline-none focus:ring-2 focus:ring-orange-100 focus:border-orange-400 appearance-none transition-all"
                  >
                    {staffOptions.map(staff => <option key={staff}>{staff}</option>)}
                  </select>
                  <ChevronDown size={12} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">核对时间</label>
                <input type="time" defaultValue="08:47" className="w-full text-xs font-bold border border-gray-100 bg-white px-3 py-2 rounded-lg outline-none focus:ring-2 focus:ring-orange-100 focus:border-orange-400 transition-all" />
              </div>
            </div>
            <button 
              onClick={() => setSecondChecked(!secondChecked)}
              className={`w-full py-3 rounded-xl text-xs font-black transition-all transform active:scale-[0.98] ${
                secondChecked ? 'bg-green-500 text-white shadow-lg shadow-green-100' : 'bg-orange-500 text-white hover:bg-orange-600 shadow-lg shadow-orange-100'
              }`}
            >
              {secondChecked ? '已完成二次核对' : '确认并完成核对'}
            </button>
          </div>
        </div>

        {/* 机表消毒登记 */}
        <div className="flex flex-col border border-emerald-100 rounded-2xl bg-white shadow-sm hover:shadow-md transition-all overflow-hidden">
          <div className="px-5 py-4 flex items-center space-x-2 border-b border-emerald-50 bg-emerald-50/20">
            <div className="w-5 h-5 bg-emerald-100 rounded-full flex items-center justify-center">
              <ClipboardCheck size={14} className="text-emerald-700" />
            </div>
            <span className="text-sm font-black text-slate-800 tracking-tight">机表消毒登记</span>
          </div>
          <div className="p-5 flex-1 flex flex-col justify-between">
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">消毒类型</label>
                  <div className="w-full text-xs font-black text-emerald-700 border border-emerald-100 bg-emerald-50/30 px-3 py-2.5 rounded-xl">
                    机表
                  </div>
                </div>
                <div className="space-y-1.5">
                  <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">消毒液</label>
                  <div className="w-full text-xs font-black text-emerald-700 border border-emerald-100 bg-emerald-50/30 px-3 py-2.5 rounded-xl truncate">
                    500mg/L含氯消毒液
                  </div>
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-gray-400 uppercase tracking-wider flex items-center">
                  消毒时间 <span className="text-[8px] text-gray-300 ml-1">(时:分)</span>
                </label>
                <div className="relative">
                  <input 
                    type="time" 
                    defaultValue="08:30" 
                    className="w-full text-xs font-bold border border-gray-100 bg-gray-50/50 px-4 py-2.5 rounded-xl outline-none focus:ring-2 focus:ring-emerald-100 focus:border-emerald-400 transition-all" 
                  />
                  <Clock size={14} className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-300" />
                </div>
              </div>
            </div>

            {/* 登记人部分 */}
            <div className="mt-6 bg-emerald-50/50 border border-emerald-100 rounded-xl p-3 flex items-center justify-between group transition-all">
              <div className="flex items-center space-x-3">
                <div className="w-8 h-8 bg-emerald-500 rounded-full flex items-center justify-center text-white font-black text-xs shadow-sm ring-2 ring-white">
                  {registrant.substring(0, 1)}
                </div>
                <div className="flex flex-col">
                  <span className="text-[10px] font-bold text-emerald-600 uppercase">登记人</span>
                  <div className="relative">
                    <select 
                      value={registrant}
                      onChange={(e) => setRegistrant(e.target.value)}
                      className="text-xs font-black text-slate-700 bg-transparent border-none p-0 focus:ring-0 outline-none cursor-pointer appearance-none pr-4"
                    >
                      {staffOptions.map(staff => <option key={staff}>{staff}</option>)}
                    </select>
                    <ChevronDown size={10} className="absolute right-0 top-1/2 -translate-y-1/2 text-emerald-400 pointer-events-none" />
                  </div>
                </div>
              </div>
              <CheckCircle size={20} className="text-emerald-500 opacity-60 group-hover:opacity-100 transition-opacity" />
            </div>
          </div>
        </div>
      </div>

      {/* 底部横排：人员配置登记 */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2 border-l-4 border-slate-700 pl-3">
            <h3 className="text-xs font-black text-slate-800 tracking-tight">人员配置与角色登记</h3>
          </div>
          <span className="text-[10px] text-gray-400 font-bold italic">提示：点击下方角色卡片可快速切换对应人员</span>
        </div>
        
        <div className="bg-slate-50 border border-gray-100 rounded-2xl p-4 shadow-inner">
          <div className="flex items-center justify-between gap-4">
            {[
              { label: '预冲护士', key: 'priming', icon: Droplets },
              { label: '穿刺/连管', key: 'puncture', icon: Stethoscope },
              { label: '上机护士', key: 'machine', icon: User },
              { label: '质控护士', key: 'qc', icon: Shield },
              { label: '质检医生', key: 'inspection', icon: Activity },
            ].map((role) => (
              <div 
                key={role.key} 
                className="flex-1 bg-white border border-gray-100 rounded-xl p-3.5 flex flex-col items-center hover:border-blue-400 hover:shadow-md transition-all cursor-pointer group relative overflow-hidden"
              >
                {/* 角色图标与标签 */}
                <div className="flex items-center space-x-2 mb-2.5 z-0">
                  <div className="w-5 h-5 bg-gray-50 rounded flex items-center justify-center group-hover:bg-blue-50 transition-colors">
                    <role.icon size={11} className="text-slate-400 group-hover:text-blue-500 transition-colors" />
                  </div>
                  <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{role.label}</span>
                </div>
                
                {/* 姓名显示 */}
                <div className="text-xs font-black text-slate-800 text-center z-0 transition-colors group-hover:text-blue-600">
                  {personnel[role.key as keyof typeof personnel]}
                </div>

                {/* 透明选择器覆盖整个卡片区域 */}
                <select 
                  className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10 appearance-none"
                  value={personnel[role.key as keyof typeof personnel]}
                  onChange={(e) => setPersonnel({...personnel, [role.key]: e.target.value})}
                  title={`点击更改${role.label}`}
                >
                  {staffOptions.map(staff => <option key={staff} value={staff}>{staff}</option>)}
                </select>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between pt-4 pb-4 border-t border-gray-100">
        <div className="flex items-center space-x-2 text-xs text-gray-400">
          <div className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-pulse"></div>
          <span>系统已根据当前排班与核对记录自动建议登记人员，请在提交前进行最后确认。</span>
        </div>
        <div className="flex space-x-4">
          <button className="px-10 py-2.5 border border-gray-300 text-gray-500 text-xs font-bold rounded-xl hover:bg-gray-50 hover:text-gray-700 transition-all">暂存修改</button>
          <button className="px-12 py-2.5 bg-blue-600 text-white text-xs font-bold rounded-xl hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 flex items-center space-x-2 transform active:scale-[0.97]">
            <CheckCircle size={16} />
            <span>提交并生效</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Verification;
