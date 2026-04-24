
import React from 'react';
import { Patient } from '../../types';
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts';
import { 
  FileText, 
  Printer, 
  Clock, 
  Activity, 
  TrendingDown, 
  ShieldCheck, 
  Zap, 
  Droplets,
  FileCheck,
  ChevronRight,
  User,
  History,
  // Fix: Added missing icons to imports
  Check,
  CheckCircle
} from 'lucide-react';

const DialysisSummary: React.FC<{ patient: Patient }> = ({ patient }) => {
  // 模拟聚合数据
  const summaryData = {
    times: {
      start: '2026-02-09 13:02',
      end: '2026-02-09 17:02',
      duration: '4小时0分'
    },
    weight: {
      pre: 96.6,
      post: 94.5,
      loss: 2.1,
      ufTarget: 2.5,
      ufActual: 2.15,
      dry: 70.5
    },
    vitals: {
      bpPre: '176/71',
      bpPost: '148/80',
      hrPre: 82,
      hrPost: 65,
      tempPre: 36.5,
      tempPost: 36.6
    },
    outcomes: {
      clotting: '0级',
      access: '良好 (杂音/震颤/搏动强)',
      complications: '无',
      event: '否'
    }
  };

  // 模拟生命体征趋势数据
  const vitalTrendData = [
    { time: '13:00', sbp: 176, dbp: 71, hr: 82 },
    { time: '14:00', sbp: 168, dbp: 75, hr: 78 },
    { time: '15:00', sbp: 160, dbp: 78, hr: 72 },
    { time: '16:00', sbp: 152, dbp: 82, hr: 68 },
    { time: '17:00', sbp: 148, dbp: 80, hr: 65 },
  ];

  return (
    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500 space-y-6 pb-10">
      


      {/* 2. 数据聚合对比网格 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* 体重与容量看板 */}
        <section className="bg-white p-6 rounded-[32px] border border-slate-100 shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-slate-50 pb-4">
             <div className="flex items-center space-x-2">
               <TrendingDown size={16} className="text-emerald-500" />
               <span className="text-xs font-black text-slate-800">容量负荷评估</span>
             </div>
             <span className="text-[9px] text-slate-300 font-bold uppercase">Fluid Summary</span>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
             <div className="p-4 bg-slate-50/50 rounded-2xl border border-slate-100">
               <p className="text-[10px] font-bold text-slate-400 mb-1">透前/透后净重</p>
               <div className="flex items-baseline space-x-1">
                 <span className="text-lg font-black text-slate-700">{summaryData.weight.pre}</span>
                 <span className="text-[10px] text-slate-300">→</span>
                 <span className="text-lg font-black text-blue-600">{summaryData.weight.post}</span>
                 <span className="text-[9px] text-slate-400 font-bold ml-1">kg</span>
               </div>
             </div>
             <div className="p-4 bg-emerald-50/30 rounded-2xl border border-emerald-100">
               <p className="text-[10px] font-bold text-emerald-600 mb-1">本次体重丢失</p>
               <div className="flex items-baseline space-x-1">
                 <span className="text-2xl font-black text-emerald-700">{summaryData.weight.loss}</span>
                 <span className="text-[10px] text-emerald-400 font-bold">kg</span>
               </div>
             </div>
             <div className="col-span-2 p-4 bg-blue-50/20 rounded-2xl border border-blue-100/50 flex items-center justify-between">
                <div>
                   <p className="text-[10px] font-bold text-slate-400 mb-1">超滤总量 (目标/实际)</p>
                   <div className="flex items-baseline space-x-2">
                      <span className="text-lg font-black text-slate-400 line-through decoration-slate-300">{summaryData.weight.ufTarget}L</span>
                      <ChevronRight size={12} className="text-slate-300" />
                      <span className="text-xl font-black text-blue-600">{summaryData.weight.ufActual}L</span>
                   </div>
                </div>
                <div 
                   className="w-12 h-12 rounded-full flex items-center justify-center"
                   style={{ 
                      background: `conic-gradient(#2563eb ${Math.round((summaryData.weight.ufActual / summaryData.weight.ufTarget) * 100)}%, #dbeafe 0)` 
                   }}
                >
                   <div className="w-9 h-9 bg-white rounded-full flex items-center justify-center shadow-inner">
                      <span className="text-[9px] font-black text-blue-600">
                         {Math.round((summaryData.weight.ufActual / summaryData.weight.ufTarget) * 100)}%
                       </span>
                   </div>
                </div>
             </div>
          </div>
        </section>

        {/* 生命体征趋势看板 */}
        <section className="bg-white p-6 rounded-[32px] border border-slate-100 shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-slate-50 pb-4">
             <div className="flex items-center space-x-2">
               <Activity size={16} className="text-red-500" />
               <span className="text-xs font-black text-slate-800">关键生命体征趋势</span>
             </div>
             <div className="flex items-center space-x-3">
                <div className="flex items-center space-x-1">
                   <div className="w-2 h-2 rounded-full bg-red-500"></div>
                   <span className="text-[8px] font-bold text-slate-400">SBP</span>
                </div>
                <div className="flex items-center space-x-1">
                   <div className="w-2 h-2 rounded-full bg-orange-400"></div>
                   <span className="text-[8px] font-bold text-slate-400">DBP</span>
                </div>
                <div className="flex items-center space-x-1">
                   <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                   <span className="text-[8px] font-bold text-slate-400">HR</span>
                </div>
             </div>
          </div>
          <div className="h-48 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={vitalTrendData} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis 
                  dataKey="time" 
                  axisLine={false} 
                  tickLine={false} 
                  tick={{ fontSize: 9, fontWeight: 700, fill: '#94a3b8' }} 
                />
                <YAxis 
                  axisLine={false} 
                  tickLine={false} 
                  tick={{ fontSize: 9, fontWeight: 700, fill: '#94a3b8' }} 
                />
                <Tooltip 
                  contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1)', fontSize: '10px', fontWeight: 'bold' }}
                />
                <Line 
                  type="monotone" 
                  dataKey="sbp" 
                  stroke="#ef4444" 
                  strokeWidth={3} 
                  dot={{ r: 3, fill: '#ef4444', strokeWidth: 2, stroke: '#fff' }}
                  activeDot={{ r: 5 }}
                  name="收缩压"
                />
                <Line 
                  type="monotone" 
                  dataKey="dbp" 
                  stroke="#fb923c" 
                  strokeWidth={3} 
                  dot={{ r: 3, fill: '#fb923c', strokeWidth: 2, stroke: '#fff' }}
                  activeDot={{ r: 5 }}
                  name="舒张压"
                />
                <Line 
                  type="monotone" 
                  dataKey="hr" 
                  stroke="#3b82f6" 
                  strokeWidth={3} 
                  dot={{ r: 3, fill: '#3b82f6', strokeWidth: 2, stroke: '#fff' }}
                  activeDot={{ r: 5 }}
                  name="心率"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </section>

        {/* 质控结果看板 */}
        <section className="bg-white p-6 rounded-[32px] border border-slate-100 shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-slate-50 pb-4">
             <div className="flex items-center space-x-2">
               <ShieldCheck size={16} className="text-indigo-500" />
               <span className="text-xs font-black text-slate-800">质控与安全汇总</span>
             </div>
             <span className="text-[9px] text-slate-300 font-bold uppercase">Quality Assurance</span>
          </div>
          <div className="space-y-4">
             <div className="flex items-center justify-between p-3 bg-indigo-50/30 rounded-2xl border border-indigo-100">
                <div className="flex flex-col">
                   <span className="text-[9px] font-black text-indigo-400 uppercase">凝血分级评估</span>
                   <span className="text-sm font-black text-indigo-700">0 级 (无凝血)</span>
                </div>
                <Zap size={20} className="text-indigo-400" />
             </div>
             <div className="flex items-center justify-between p-3 bg-slate-50/50 rounded-2xl border border-slate-100">
                <div className="flex flex-col">
                   <span className="text-[9px] font-black text-slate-400 uppercase">血管通路状态</span>
                   <span className="text-[11px] font-black text-slate-700">{summaryData.outcomes.access}</span>
                </div>
             </div>
             <div className="flex items-center justify-between p-3 bg-slate-50/50 rounded-2xl border border-slate-100">
                <div className="flex flex-col">
                   <span className="text-[9px] font-black text-slate-400 uppercase">透析并发症</span>
                   <span className="text-[11px] font-black text-slate-700">{summaryData.outcomes.complications}</span>
                </div>
                {summaryData.outcomes.complications === '无' && <Check size={16} className="text-emerald-500" />}
             </div>
          </div>
        </section>
      </div>

      {/* 3. 自动化小结文本 */}
      <section className="bg-slate-900 text-white p-8 rounded-[40px] shadow-2xl relative overflow-hidden group">
         <div className="absolute top-0 right-0 p-8 opacity-10 group-hover:scale-110 transition-transform duration-700">
            <FileText size={120} />
         </div>
         <div className="relative z-10 space-y-4">
            <div className="flex items-center space-x-3">
               <div className="w-8 h-8 bg-blue-500 rounded-lg flex items-center justify-center">
                  <Zap size={16} className="text-white" />
               </div>
               <h3 className="text-sm font-black uppercase tracking-widest">智能化自动小结生成</h3>
            </div>
            <p className="text-sm text-slate-300 leading-relaxed font-medium">
               患者于 {summaryData.times.start} 开始 HDF 治疗，历时 {summaryData.times.duration}。
               透前体重 {summaryData.weight.pre}kg，干体重 {summaryData.weight.dry}kg。
               治疗过程中生命体征基本平稳，血压在 {summaryData.vitals.bpPre} 至 {summaryData.vitals.bpPost} 之间波动。
               实际超滤量为 {summaryData.weight.ufActual}L。
               下机观察透析器及血路管凝血分级为 {summaryData.outcomes.clotting}，穿刺点及管路固定良好，无明显渗血及过敏反应。
               患者自诉无不适，平稳下机。
            </p>
            <div className="flex items-center space-x-6 pt-4">
               <div className="flex items-center space-x-2">
                  <User size={14} className="text-blue-400" />
                  <span className="text-[10px] font-black text-slate-400 uppercase">审核医生: 董婉颖</span>
               </div>
               <div className="flex items-center space-x-2 border-l border-white/10 pl-6">
                  <User size={14} className="text-blue-400" />
                  <span className="text-[10px] font-black text-slate-400 uppercase">确认护士: 刘护士长</span>
               </div>
               <button className="ml-auto flex items-center space-x-2 text-[10px] font-black text-blue-400 hover:text-blue-300 transition-colors">
                  <span>点击手动编辑</span>
                  <ChevronRight size={12} />
               </button>
            </div>
         </div>
      </section>

      {/* 4. 医嘱执行汇总列表 */}
      <section className="bg-white border border-slate-100 rounded-[32px] overflow-hidden shadow-sm">
         <div className="px-8 py-5 border-b border-slate-50 flex items-center justify-between">
            <div className="flex items-center space-x-3">
               <Droplets size={16} className="text-blue-600" />
               <h3 className="text-xs font-black text-slate-800 uppercase tracking-widest">本次执行医嘱明细汇总</h3>
            </div>
            <div className="flex items-center space-x-2 text-[10px] text-slate-400 font-bold">
               <History size={12} />
               <span>Total: 3 Items executed</span>
            </div>
         </div>
         <div className="overflow-x-auto">
            <table className="w-full text-left">
               <thead>
                  <tr className="bg-slate-50/50">
                     <th className="px-8 py-3 text-[10px] font-black text-slate-400 uppercase w-20">类型</th>
                     <th className="px-8 py-3 text-[10px] font-black text-slate-400 uppercase">项目名称</th>
                     <th className="px-8 py-3 text-[10px] font-black text-slate-400 uppercase w-48">执行剂量与用法</th>
                     <th className="px-8 py-3 text-[10px] font-black text-slate-400 uppercase w-40">执行时间</th>
                     <th className="px-8 py-3 text-[10px] font-black text-slate-400 uppercase w-32">执行人</th>
                     <th className="px-8 py-3 text-[10px] font-black text-slate-400 uppercase w-20 text-center">状态</th>
                  </tr>
               </thead>
               <tbody className="divide-y divide-slate-50">
                  {[
                     { type: '长期', name: '血液透析滤过 (HDF)', usage: '4小时, 流量500ml/min', time: '13:02', nurse: '刘护士长' },
                     { type: '临时', name: '左卡尼汀注射液', usage: '1.0g, 静脉注射', time: '13:15', nurse: '刘护士长' },
                     { type: '临时', name: '普通肝素 (首剂)', usage: '2000 IU, 穿刺即刻', time: '13:00', nurse: '刘护士长' },
                  ].map((order, i) => (
                     <tr key={i} className="hover:bg-slate-50/50 transition-colors">
                        <td className="px-8 py-4">
                           <span className={`text-[9px] font-black px-2 py-0.5 rounded ${
                              order.type === '长期' ? 'bg-indigo-100 text-indigo-700' : 'bg-fuchsia-100 text-fuchsia-700'
                           }`}>{order.type}</span>
                        </td>
                        <td className="px-8 py-4 text-xs font-black text-slate-800">{order.name}</td>
                        <td className="px-8 py-4 text-[10px] text-slate-500 font-bold">{order.usage}</td>
                        <td className="px-8 py-4 text-[10px] text-slate-400 font-mono">{order.time}</td>
                        <td className="px-8 py-4 text-xs font-bold text-slate-600">{order.nurse}</td>
                        <td className="px-8 py-4 text-center">
                           <div className="w-5 h-5 bg-emerald-500 rounded-full flex items-center justify-center text-white mx-auto">
                              <CheckCircle size={12} />
                           </div>
                        </td>
                     </tr>
                  ))}
               </tbody>
            </table>
         </div>
      </section>

    </div>
  );
};

export default DialysisSummary;
