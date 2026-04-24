import React, { useState } from 'react';
import { Patient } from '../../types';
import { 
  Plus, 
  Search, 
  Activity, 
  Clock, 
  Thermometer, 
  Heart, 
  Droplets, 
  Zap, 
  AlertCircle,
  Edit3,
  Trash2,
  ChevronDown,
  CheckCircle2,
  Info,
  MoreHorizontal,
  // Fix: Added History icon import from lucide-react to prevent collision with global History interface
  History
} from 'lucide-react';

const MidMonitoring: React.FC<{ patient: Patient }> = ({ patient }) => {
  const [showAddRow, setShowAddRow] = useState(false);

  // 模拟监测数据
  const [records, setRecords] = useState([
    {
      id: 1,
      time: '08:00',
      sbp: 125, dbp: 78, hr: 72, temp: 36.5, resp: 18, spo2: 99,
      uf: 0.5, flow: 250, artPres: -120, venPres: 140, tmp: 60,
      mTemp: 36.5, cond: 14.1, heparin: 2.0, rbv: '98%', rtv: '115L',
      clearance: '235', artBloodTemp: 36.2,
      symptom: '无', symptomType: '-', treatment: '-', result: '-',
      lastModified: '08:05:22',
      artFix: '良好', artOoze: '无', artOther: '-',
      venFix: '良好', venOoze: '无', venOther: '-',
      lineFix: '良好', lineOoze: '无',
      remark: '常规监测', nurse: '刘护士长'
    },
    {
      id: 2,
      time: '09:00',
      sbp: 118, dbp: 72, hr: 68, temp: 36.6, resp: 20, spo2: 98,
      uf: 1.0, flow: 250, artPres: -135, venPres: 155, tmp: 75,
      mTemp: 36.5, cond: 14.0, heparin: 2.0, rbv: '95%', rtv: '118L',
      clearance: '240', artBloodTemp: 36.3,
      symptom: '轻微头晕', symptomType: '透析低血压', treatment: '减慢超滤', result: '缓解',
      lastModified: '09:12:45',
      artFix: '良好', artOoze: '无', artOther: '-',
      venFix: '良好', venOoze: '无', venOther: '-',
      lineFix: '良好', lineOoze: '无',
      remark: '已处理并发症', nurse: '刘护士长'
    }
  ]);

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      
      {/* 1. 指标概览卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        {[
          { label: '平均动脉压', value: '93', unit: 'mmHg', icon: Activity, color: 'text-blue-600', bg: 'bg-blue-50' },
          { label: '实时跨膜压', value: '75', unit: 'mmHg', icon: Zap, color: 'text-amber-600', bg: 'bg-amber-50' },
          { label: '当前血流量', value: '250', unit: 'ml/min', icon: Droplets, color: 'text-indigo-600', bg: 'bg-indigo-50' },
          { label: '超滤速率', value: '0.45', unit: 'L/h', icon: Clock, color: 'text-emerald-600', bg: 'bg-emerald-50' },
          { label: '异常预警', value: '1', unit: '项', icon: AlertCircle, color: 'text-red-600', bg: 'bg-red-50' },
        ].map((card, i) => (
          <div key={i} className="bg-white p-4 rounded-2xl border border-gray-100 shadow-sm flex items-center space-x-4">
            <div className={`w-12 h-12 ${card.bg} rounded-xl flex items-center justify-center`}>
              <card.icon className={card.color} size={22} />
            </div>
            <div>
              <p className="text-[10px] font-black text-gray-400 uppercase tracking-tight">{card.label}</p>
              <div className="flex items-baseline space-x-1">
                <span className={`text-xl font-black ${card.color === 'text-red-600' ? 'animate-pulse' : 'text-slate-800'}`}>{card.value}</span>
                <span className="text-[10px] text-gray-400 font-bold">{card.unit}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 2. 监测记录列表 (超宽表) */}
      <div className="bg-white border border-gray-100 rounded-[28px] shadow-sm flex flex-col overflow-hidden">
        {/* Table Header Action */}
        <div className="px-6 py-4 border-b border-gray-100 flex items-center justify-between bg-white sticky left-0">
          <div className="flex items-center space-x-3">
             <div className="w-1.5 h-6 bg-blue-600 rounded-full"></div>
             <h3 className="text-sm font-black text-slate-800">实时监测记录流水</h3>
             <span className="text-[10px] px-2 py-0.5 bg-slate-100 text-slate-400 font-black rounded-lg uppercase tracking-widest">REAL-TIME FEED</span>
          </div>
          <button 
            onClick={() => setShowAddRow(true)}
            className="px-5 py-2 bg-blue-600 text-white text-[11px] font-black rounded-xl hover:bg-blue-700 shadow-lg shadow-blue-100 flex items-center space-x-2 transition-all transform active:scale-95"
          >
            <Plus size={14} />
            <span>录入新监测点</span>
          </button>
        </div>

        {/* Scrollable Container */}
        <div className="overflow-x-auto no-scrollbar relative">
          <table className="w-full text-left border-collapse min-w-[3600px] table-fixed">
            <thead>
              {/* Grouped Sub-header Rows */}
              <tr className="bg-slate-50/50 border-b border-gray-100">
                <th colSpan={2} className="px-4 py-2 text-[9px] font-black text-slate-400 uppercase tracking-widest text-center border-r border-gray-100 sticky left-0 bg-slate-50 z-20">基础</th>
                <th colSpan={6} className="px-4 py-2 text-[9px] font-black text-blue-400 uppercase tracking-widest text-center border-r border-gray-100">生命体征指标</th>
                <th colSpan={5} className="px-4 py-2 text-[9px] font-black text-indigo-400 uppercase tracking-widest text-center border-r border-gray-100">透析核心压力与流速</th>
                <th colSpan={7} className="px-4 py-2 text-[9px] font-black text-emerald-400 uppercase tracking-widest text-center border-r border-gray-100">设备机能与生化指标</th>
                <th colSpan={4} className="px-4 py-2 text-[9px] font-black text-amber-500 uppercase tracking-widest text-center border-r border-gray-100">并发症与症状处理</th>
                <th colSpan={1} className="px-4 py-2 text-[9px] font-black text-slate-400 uppercase tracking-widest text-center border-r border-gray-100">时效</th>
                <th colSpan={8} className="px-4 py-2 text-[9px] font-black text-fuchsia-400 uppercase tracking-widest text-center border-r border-gray-100">穿刺点与物理观察</th>
                <th colSpan={3} className="px-4 py-2 text-[9px] font-black text-slate-400 uppercase tracking-widest text-center sticky right-0 bg-slate-50 z-20">末尾项</th>
              </tr>
              <tr className="bg-slate-50 border-b border-gray-100">
                {/* 冻结列 (序号, 时间) */}
                <th className="px-4 py-3 text-[10px] font-black text-slate-500 w-16 text-center sticky left-0 bg-slate-50 z-20">序号</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-500 w-32 border-r border-gray-200 sticky left-16 bg-slate-50 z-20">观测时间</th>
                
                {/* 生命体征 */}
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">收缩压 (mmHg)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">舒张压 (mmHg)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24 text-red-500">心率 (bpm)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">体温 (°C)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">呼吸 (次/分)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24 border-r border-gray-200">血氧 (%)</th>
                
                {/* 透析参数 */}
                <th className="px-4 py-3 text-[10px] font-black text-blue-700 w-24">累计超滤 (L)</th>
                <th className="px-4 py-3 text-[10px] font-black text-blue-700 w-24">血流量 (ml/min)</th>
                <th className="px-4 py-3 text-[10px] font-black text-blue-700 w-24">动脉压 (mmHg)</th>
                <th className="px-4 py-3 text-[10px] font-black text-blue-700 w-24">静脉压 (mmHg)</th>
                <th className="px-4 py-3 text-[10px] font-black text-blue-700 w-24 border-r border-gray-200 text-orange-600">跨膜压 (mmHg)</th>
                
                {/* 机能生化 */}
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">机温 (°C)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">电导率 (mS/cm)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">肝素流量 (ml/h)</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">相对血容量</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">实时血容量</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24">实时清除率</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-24 border-r border-gray-200">动脉血温</th>
                
                {/* 症状处理 */}
                <th className="px-4 py-3 text-[10px] font-black text-amber-600 w-40">症状描述</th>
                <th className="px-4 py-3 text-[10px] font-black text-amber-600 w-40">症状类型</th>
                <th className="px-4 py-3 text-[10px] font-black text-amber-600 w-40">处理内容</th>
                <th className="px-4 py-3 text-[10px] font-black text-amber-600 w-32 border-r border-gray-200">处理结果</th>
                
                {/* 时效 */}
                <th className="px-4 py-3 text-[10px] font-black text-slate-400 w-32 border-r border-gray-200">修改时间</th>
                
                {/* 物理观察 */}
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">动脉点固定</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">动脉点渗血</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">动脉其他</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">静脉点固定</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">静脉点渗血</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">静脉其他</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32">血路管固定</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-700 w-32 border-r border-gray-200">血路管渗血</th>
                
                {/* 末尾冻结列 */}
                <th className="px-4 py-3 text-[10px] font-black text-slate-500 w-40 sticky right-24 bg-slate-50 z-20">备注</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-500 w-24 text-center sticky right-0 bg-slate-50 z-20 border-l border-gray-100 shadow-[-4px_0_8px_rgba(0,0,0,0.03)]">责任护士</th>
                <th className="px-4 py-3 text-[10px] font-black text-slate-500 w-24 text-center sticky right-0 bg-slate-50 z-30">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50 bg-white">
              {records.map((row, idx) => (
                <tr key={row.id} className="hover:bg-blue-50/20 transition-colors group">
                  <td className="px-4 py-4 text-xs font-black text-gray-300 text-center sticky left-0 bg-white group-hover:bg-blue-50/20 z-10">{idx + 1}</td>
                  <td className="px-4 py-4 text-xs font-black text-blue-600 sticky left-16 bg-white group-hover:bg-blue-50/20 z-10 border-r border-gray-200">
                    <div className="flex items-center space-x-2">
                       <Clock size={12} className="text-blue-300" />
                       <span>{row.time}</span>
                    </div>
                  </td>
                  
                  {/* Vital Signs Data */}
                  <td className="px-4 py-4 text-xs font-black text-slate-700">{row.sbp}</td>
                  <td className="px-4 py-4 text-xs font-black text-slate-700">{row.dbp}</td>
                  <td className="px-4 py-4 text-xs font-black text-red-500">{row.hr}</td>
                  <td className="px-4 py-4 text-xs font-black text-slate-700">{row.temp}</td>
                  <td className="px-4 py-4 text-xs font-black text-slate-700">{row.resp}</td>
                  <td className="px-4 py-4 text-xs font-black text-slate-700 border-r border-gray-100">{row.spo2}%</td>
                  
                  {/* Dialysis Params Data */}
                  <td className="px-4 py-4 text-xs font-black text-blue-700">{row.uf}</td>
                  <td className="px-4 py-4 text-xs font-black text-blue-700">{row.flow}</td>
                  <td className="px-4 py-4 text-xs font-black text-blue-700 italic">{row.artPres}</td>
                  <td className="px-4 py-4 text-xs font-black text-blue-700 italic">{row.venPres}</td>
                  <td className="px-4 py-4 text-xs font-black text-orange-500 border-r border-gray-100">{row.tmp}</td>
                  
                  {/* Machine/Bio Data */}
                  <td className="px-4 py-4 text-xs font-bold text-slate-500">{row.mTemp}</td>
                  <td className="px-4 py-4 text-xs font-bold text-slate-500">{row.cond}</td>
                  <td className="px-4 py-4 text-xs font-bold text-slate-500">{row.heparin}</td>
                  <td className="px-4 py-4 text-xs font-bold text-slate-500">{row.rbv}</td>
                  <td className="px-4 py-4 text-xs font-bold text-slate-500">{row.rtv}</td>
                  <td className="px-4 py-4 text-xs font-bold text-slate-500">{row.clearance}</td>
                  <td className="px-4 py-4 text-xs font-bold text-slate-500 border-r border-gray-100">{row.artBloodTemp}</td>

                  {/* Symptom Data */}
                  <td className="px-4 py-4">
                    <span className={`text-xs font-bold ${row.symptom === '无' ? 'text-gray-300' : 'text-amber-600'}`}>
                      {row.symptom}
                    </span>
                  </td>
                  <td className="px-4 py-4 text-xs text-slate-400">{row.symptomType}</td>
                  <td className="px-4 py-4 text-xs text-slate-400">{row.treatment}</td>
                  <td className="px-4 py-4 text-xs border-r border-gray-100">
                     {row.result !== '-' && <span className="px-2 py-0.5 bg-emerald-50 text-emerald-600 rounded text-[9px] font-black uppercase">Result: {row.result}</span>}
                  </td>

                  {/* 时效 */}
                  <td className="px-4 py-4 text-[10px] text-gray-400 font-mono border-r border-gray-100">{row.lastModified}</td>

                  {/* Physical Observation Data */}
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.artFix}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.artOoze}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.artOther}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.venFix}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.venOoze}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.venOther}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500">{row.lineFix}</td>
                  <td className="px-4 py-4 text-xs font-medium text-slate-500 border-r border-gray-100">{row.lineOoze}</td>

                  {/* Sticky End Cols */}
                  <td className="px-4 py-4 text-xs text-gray-400 italic sticky right-24 bg-white group-hover:bg-blue-50/20 z-10 truncate">{row.remark}</td>
                  <td className="px-4 py-4 text-center sticky right-0 bg-white group-hover:bg-blue-50/20 z-10 border-l border-gray-100 shadow-[-4px_0_8px_rgba(0,0,0,0.03)]">
                     <div className="flex flex-col items-center">
                        <div className="w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center text-[9px] font-black text-blue-600 mb-0.5">
                          {row.nurse.substring(0,1)}
                        </div>
                        <span className="text-[10px] font-bold text-slate-600">{row.nurse}</span>
                     </div>
                  </td>
                  <td className="px-4 py-4 text-center sticky right-0 bg-white group-hover:bg-blue-50/20 z-20">
                    <div className="flex items-center justify-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button className="p-1.5 text-blue-600 hover:bg-blue-100 rounded-lg"><Edit3 size={14} /></button>
                      <button className="p-1.5 text-red-400 hover:bg-red-50 rounded-lg"><Trash2 size={14} /></button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* 3. 分页与底部状态 */}
        <div className="px-6 py-4 bg-slate-50/50 border-t border-gray-100 flex items-center justify-between">
           <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2 text-[10px] font-black text-gray-400 uppercase tracking-widest">
                 <div className="w-2 h-2 rounded-full bg-emerald-500"></div>
                 <span>机位状态: 正常传输中</span>
              </div>
              <div className="flex items-center space-x-2 text-[10px] font-black text-gray-400 uppercase tracking-widest border-l pl-6">
                 <History size={12} />
                 <span>同步间隔: 60分钟/点</span>
              </div>
           </div>
           <div className="flex items-center space-x-2">
              <button className="px-3 py-1 text-[10px] font-black text-slate-500 bg-white border border-gray-200 rounded-lg shadow-sm hover:bg-gray-50">上一页</button>
              <button className="px-3 py-1 text-[10px] font-black text-blue-600 bg-white border border-blue-200 rounded-lg shadow-sm hover:bg-blue-50">下一页</button>
           </div>
        </div>
      </div>
      
      {/* 4. 操作提示 */}
      <div className="bg-amber-50 rounded-2xl p-4 border border-amber-100 flex items-start space-x-3">
         <Info size={16} className="text-amber-500 mt-0.5 shrink-0" />
         <p className="text-xs text-amber-700 font-medium leading-relaxed">
           监测数据支持自动采集与人工录入结合。系统已锁定历史监测记录，如需修改 15 分钟前的监测点，请联系护士长权限核准。
         </p>
      </div>
    </div>
  );
};

export default MidMonitoring;