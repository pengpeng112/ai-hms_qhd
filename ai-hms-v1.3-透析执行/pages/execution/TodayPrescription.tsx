
import React, { useState } from 'react';
import { Patient } from '../../types';
import { 
  Settings, 
  Droplets, 
  Zap, 
  FileText,
  Clock,
  History,
  Package,
  Activity,
  ArrowRight,
  ClipboardList,
  ChevronDown,
  X,
  Save,
  Plus,
  Trash2,
  Lock
} from 'lucide-react';

interface Material {
  id: number;
  name: string;
  cat: string;
  qty: number;
  code: string;
  brand: string;
  spec: string;
  remark: string;
}

const TodayPrescription: React.FC<{ patient: Patient }> = ({ patient }) => {
  const [isEditing, setIsEditing] = useState(false);

  // 1. 处方参数状态
  const [data, setData] = useState({
    treatmentMethod: patient.treatmentPlan, // 固定
    preNetWeight: 96.6, // 固定
    lastPostNetWeight: 94.3, // 固定
    increaseVsLastPost: 2.3, // 固定
    dryWeight: patient.dryWeight,
    ufTarget: 2.5,
    currentWeightIncrease: 2.1, // 固定
    preBP: '176/71', // 固定
    duration: 4.0,
    stdBloodFlow: 230,
    heparinType: '普通肝素',
    initialDoseName: '普通肝素',
    initialDoseValue: 2000,
    maintenanceDoseName: '普通肝素',
    infusionRate: 2.0,
    infusionTime: 3.5,
    maintenanceValue: 1000,
    vascularAccess: '左侧自体动静脉内瘘',
    dialysateType: '标准',
    dialysateFlow: 500,
    caIon: 1.5,
    kIon: 2.0,
    glucose: 5.5,
    conductivity: 14.0,
    temperature: 36.5,
    totalVolume: 120
  });

  // 2. 材料列表状态
  const [materials, setMaterials] = useState<Material[]>([
    { id: 1, name: 'Fresenius FX100', cat: '透析器', qty: 1, code: 'EQ-0012', brand: '费森尤斯', spec: '1.8m²', remark: '-' },
    { id: 2, name: 'AV-Set 2000A', cat: '血路管', qty: 1, code: 'EQ-0943', brand: '尼普洛', spec: '成人标准', remark: '-' },
    { id: 3, name: '16G 穿刺针', cat: '穿刺针', qty: 2, code: 'EQ-0211', brand: 'JMS', spec: '16G', remark: '动脉、静脉各一' },
    { id: 4, name: '无菌透析护理包', cat: '护理包', qty: 1, code: 'EQ-0552', brand: '恒健', spec: '标准型', remark: '-' },
  ]);

  const handleParamChange = (key: keyof typeof data, value: any) => {
    setData(prev => ({ ...prev, [key]: value }));
  };

  // 材料操作：新增
  const handleAddMaterial = () => {
    const newId = materials.length > 0 ? Math.max(...materials.map(m => m.id)) + 1 : 1;
    setMaterials([...materials, { id: newId, name: '', cat: '', qty: 1, code: '', brand: '', spec: '', remark: '' }]);
  };

  // 材料操作：删除
  const handleDeleteMaterial = (id: number) => {
    setMaterials(materials.filter(m => m.id !== id));
  };

  // 材料操作：更新
  const handleUpdateMaterial = (id: number, field: keyof Material, value: any) => {
    setMaterials(materials.map(m => m.id === id ? { ...m, [field]: value } : m));
  };

  const handleSave = () => {
    setIsEditing(false);
    console.log('保存处方:', data);
    console.log('保存材料:', materials);
  };

  // 辅助组件：参数字段渲染
  const RenderField = ({ label, value, unit, fieldKey, type = "text", options, isFixed = false }: { 
    label: string, value: any, unit?: string, fieldKey: keyof typeof data, type?: "text" | "number" | "select", options?: string[], isFixed?: boolean 
  }) => {
    if (!isEditing || isFixed) {
      return (
        <div className={`p-2 rounded-lg transition-colors border border-transparent ${isFixed && isEditing ? 'bg-gray-100/40 border-dashed border-gray-200 opacity-60' : ''}`}>
          <div className="flex items-center space-x-1 mb-0.5">
            <p className="text-[10px] font-bold text-gray-400 truncate">{label}</p>
            {isFixed && isEditing && <Lock size={8} className="text-gray-300" />}
          </div>
          <p className={`text-[13px] font-black leading-tight ${isFixed ? 'text-slate-500' : 'text-slate-700'}`}>
            {value} <span className="text-[9px] text-gray-400 font-bold ml-0.5">{unit}</span>
          </p>
        </div>
      );
    }

    return (
      <div className="space-y-1 p-1">
        <label className="text-[10px] font-bold text-blue-600 truncate block">{label}</label>
        {type === "select" ? (
          <div className="relative">
            <select 
              value={value} 
              onChange={(e) => handleParamChange(fieldKey, e.target.value)}
              className="w-full h-8 px-2 border border-blue-100 rounded text-xs font-black text-slate-700 outline-none bg-blue-50/30 appearance-none"
            >
              {options?.map(opt => <option key={opt}>{opt}</option>)}
            </select>
            <ChevronDown size={12} className="absolute right-2 top-1/2 -translate-y-1/2 text-blue-400 pointer-events-none" />
          </div>
        ) : (
          <div className="flex items-center h-8 border border-blue-100 rounded bg-blue-50/30 overflow-hidden focus-within:ring-1 focus-within:ring-blue-400">
            <input 
              type={type} 
              value={value} 
              onChange={(e) => handleParamChange(fieldKey, type === "number" ? parseFloat(e.target.value) : e.target.value)}
              className="w-full px-2 text-xs font-black text-slate-700 outline-none bg-transparent"
            />
            {unit && <span className="px-1 text-[9px] text-blue-400 font-bold border-l border-blue-100 bg-white/50 h-full flex items-center shrink-0">{unit}</span>}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500 pb-8">
      
      {/* 1. 顶部指标区 */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        <div className={`col-span-1 rounded-2xl p-3 text-white shadow-lg flex flex-col justify-between transition-colors ${isEditing ? 'bg-slate-700 shadow-slate-100' : 'bg-blue-600 shadow-blue-100'}`}>
          <div className="flex items-center justify-between">
            <p className="text-white/60 text-[10px] font-bold uppercase tracking-wider">透析方法</p>
            {isEditing && <Lock size={10} className="text-white/40" />}
          </div>
          <div className="flex items-baseline space-x-1">
            <h2 className="text-xl font-black">{data.treatmentMethod}</h2>
            <span className="text-[9px] opacity-60">固定值</span>
          </div>
        </div>
        
        {[
          { label: '本次体重增加量', value: data.currentWeightIncrease, unit: 'kg', color: 'text-red-500/80', isFixed: true },
          { label: '目标超滤量', value: data.ufTarget, unit: 'L', color: 'text-blue-600', isEditable: true, field: 'ufTarget' },
          { label: '透前血压', value: data.preBP, unit: 'mmHg', color: 'text-slate-400', isFixed: true },
          { label: '透析时间', value: data.duration, unit: 'h', color: 'text-slate-800', isEditable: true, field: 'duration' },
        ].map((item, idx) => (
          <div key={idx} className="bg-white border border-gray-100 rounded-2xl p-3 shadow-sm">
            <p className="text-[10px] text-gray-400 font-bold mb-1">{item.label}</p>
            <div className="flex items-baseline space-x-1">
              {isEditing && item.isEditable ? (
                <input 
                  type="number" 
                  value={item.value as number} 
                  onChange={(e) => handleParamChange(item.field as any, parseFloat(e.target.value))}
                  className="w-16 h-7 text-lg font-black text-blue-600 bg-blue-50/50 rounded outline-none border border-blue-100 px-1"
                />
              ) : (
                <span className={`text-xl font-black ${item.color}`}>{item.value}</span>
              )}
              <span className="text-[10px] text-gray-400 font-bold uppercase">{item.unit}</span>
              {isEditing && item.isFixed && <Lock size={8} className="ml-1 text-gray-200" />}
            </div>
          </div>
        ))}
      </div>

      {/* 2. 详情参数区 - 调整为更紧凑的 3-4 列布局 */}
      <div className="flex flex-col space-y-4">
        {/* 第一排：体重基础 + 抗凝方案 */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          {/* 体重与基础参数 */}
          <section className={`bg-white border rounded-2xl overflow-hidden shadow-sm transition-all ${isEditing ? 'border-blue-200 ring-1 ring-blue-50 shadow-md' : 'border-gray-100'}`}>
            <div className="bg-gray-50 px-4 py-2 border-b flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Activity size={13} className="text-blue-600" />
                <span className="text-[11px] font-black text-gray-700 uppercase tracking-tight">体重详情与基础标准</span>
              </div>
            </div>
            <div className="p-3 grid grid-cols-3 gap-2">
              <RenderField label="透前净重" value={data.preNetWeight} unit="kg" fieldKey="preNetWeight" isFixed={true} />
              <RenderField label="干体重" value={data.dryWeight} unit="kg" fieldKey="dryWeight" type="number" />
              <RenderField label="上次透后" value={data.lastPostNetWeight} unit="kg" fieldKey="lastPostNetWeight" isFixed={true} />
              <RenderField label="较前增量" value={data.increaseVsLastPost} unit="kg" fieldKey="increaseVsLastPost" isFixed={true} />
              <RenderField label="标准血流" value={data.stdBloodFlow} unit="ml/min" fieldKey="stdBloodFlow" type="number" />
            </div>
          </section>

          {/* 抗凝方案 */}
          <section className={`bg-white border rounded-2xl overflow-hidden shadow-sm transition-all ${isEditing ? 'border-blue-200 ring-1 ring-blue-50 shadow-md' : 'border-gray-100'}`}>
            <div className="bg-gray-50 px-4 py-2 border-b flex items-center space-x-2">
              <Droplets size={13} className="text-blue-600" />
              <span className="text-[11px] font-black text-gray-700 uppercase tracking-tight">抗凝方案设定</span>
            </div>
            <div className="p-3 space-y-2">
              <div className="grid grid-cols-4 gap-2">
                <div className="col-span-2">
                  <RenderField 
                    label="肝素类型" 
                    value={data.heparinType} 
                    fieldKey="heparinType" 
                    type="select" 
                    options={['普通肝素', '低分子肝素', '无肝素', '枸橼酸抗凝']} 
                  />
                </div>
                <RenderField label="首剂名称" value={data.initialDoseName} fieldKey="initialDoseName" />
                <RenderField label="首剂量" value={data.initialDoseValue} unit="IU" fieldKey="initialDoseValue" type="number" />
                
                <RenderField label="维持剂" value={data.maintenanceDoseName} fieldKey="maintenanceDoseName" />
                <RenderField label="维持量" value={data.maintenanceValue} unit="IU/h" fieldKey="maintenanceValue" type="number" />
                <RenderField label="注入速率" value={data.infusionRate} unit="ml/h" fieldKey="infusionRate" type="number" />
                <RenderField label="注入时间" value={data.infusionTime} unit="h" fieldKey="infusionTime" type="number" />
              </div>
            </div>
          </section>
        </div>

        {/* 第二排：透析液参数 (全宽展示以容纳更多列) */}
        <section className={`bg-white border rounded-2xl overflow-hidden shadow-sm transition-all ${isEditing ? 'border-blue-200 ring-1 ring-blue-50 shadow-md' : 'border-gray-100'}`}>
          <div className="bg-gray-50 px-4 py-2 border-b flex items-center space-x-2">
            <Settings size={13} className="text-blue-600" />
            <span className="text-[11px] font-black text-gray-700 uppercase tracking-tight">透析液及通路设定</span>
          </div>
          <div className="p-3 grid grid-cols-4 md:grid-cols-5 xl:grid-cols-9 gap-2">
            <div className="col-span-2">
              <RenderField 
                label="血管通路" 
                value={data.vascularAccess} 
                fieldKey="vascularAccess" 
                type="select" 
                options={['左侧自体动静脉内瘘', '右侧自体动静脉内瘘', '颈内静脉置管', '股静脉置管']} 
              />
            </div>
            <RenderField label="透析液" value={data.dialysateType} fieldKey="dialysateType" />
            <RenderField label="透析流速" value={data.dialysateFlow} unit="ml/min" fieldKey="dialysateFlow" type="number" />
            <RenderField label="Ca浓度" value={data.caIon} unit="mmol/L" fieldKey="caIon" type="number" />
            <RenderField label="K浓度" value={data.kIon} unit="mmol/L" fieldKey="kIon" type="number" />
            <RenderField label="葡萄糖" value={data.glucose} unit="mmol/L" fieldKey="glucose" type="number" />
            <RenderField label="电导度" value={data.conductivity} unit="mS/cm" fieldKey="conductivity" type="number" />
            <RenderField label="液温" value={data.temperature} unit="°C" fieldKey="temperature" type="number" />
          </div>
        </section>
      </div>

      {/* 3. 透析材料管理列表 */}
      <section className={`bg-white border rounded-2xl overflow-hidden shadow-sm transition-all ${isEditing ? 'border-blue-200 shadow-md' : 'border-gray-100'}`}>
        <div className="bg-gray-50 px-4 py-2.5 border-b flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Package size={13} className="text-blue-600" />
            <span className="text-[11px] font-black text-gray-700 uppercase tracking-wider">透析材料清单</span>
          </div>
          {isEditing && (
            <button 
              onClick={handleAddMaterial}
              className="flex items-center space-x-1 px-3 py-1 bg-blue-600 text-white rounded-lg text-[9px] font-black hover:bg-blue-700 transition-colors shadow-sm"
            >
              <Plus size={10} />
              <span>新增材料</span>
            </button>
          )}
        </div>
        <div className="overflow-x-auto no-scrollbar">
          <table className="w-full text-left border-collapse min-w-[800px]">
            <thead>
              <tr className="bg-slate-50 border-b border-gray-100">
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-10 text-center">#</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-40">材料名称</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-20">分类</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-16 text-center">数量</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-28">编码</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-20">品牌</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase w-28">规格</th>
                <th className="px-3 py-2 text-[9px] font-bold text-gray-400 uppercase">备注</th>
                {isEditing && <th className="px-3 py-2 text-[9px] font-bold text-red-400 uppercase w-12 text-center">删</th>}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {materials.map((item, index) => (
                <tr key={item.id} className="hover:bg-slate-50 transition-colors">
                  <td className="px-3 py-1.5 text-[11px] text-center font-bold text-gray-400">{index + 1}</td>
                  <td className="px-3 py-1.5">
                    {isEditing ? (
                      <input 
                        className="w-full h-7 border rounded border-blue-100 bg-blue-50/20 px-2 text-[11px] font-bold text-slate-700 focus:bg-white outline-none"
                        value={item.name}
                        onChange={(e) => handleUpdateMaterial(item.id, 'name', e.target.value)}
                      />
                    ) : (
                      <span className="text-[11px] font-black text-slate-700 truncate block max-w-[150px]">{item.name || '-'}</span>
                    )}
                  </td>
                  <td className="px-3 py-1.5 text-[11px] text-slate-500">{item.cat || '-'}</td>
                  <td className="px-3 py-1.5 text-center">
                    {isEditing ? (
                      <input 
                        type="number"
                        className="w-12 h-7 border rounded border-blue-100 bg-blue-50/20 px-1 text-center text-[11px] font-bold text-slate-700 focus:bg-white outline-none"
                        value={item.qty}
                        onChange={(e) => handleUpdateMaterial(item.id, 'qty', parseInt(e.target.value))}
                      />
                    ) : (
                      <span className="text-[11px] font-black text-blue-600">{item.qty}</span>
                    )}
                  </td>
                  <td className="px-3 py-1.5 text-[9px] text-gray-400 font-mono">{item.code || '-'}</td>
                  <td className="px-3 py-1.5 text-[11px] text-slate-500">{item.brand || '-'}</td>
                  <td className="px-3 py-1.5 text-[11px] text-slate-500">{item.spec || '-'}</td>
                  <td className="px-3 py-1.5 text-[9px] text-gray-400 italic truncate max-w-[100px]">{item.remark || '-'}</td>
                  {isEditing && (
                    <td className="px-3 py-1.5 text-center">
                      <button 
                        onClick={() => handleDeleteMaterial(item.id)}
                        className="p-1 text-red-400 hover:text-red-600 rounded transition-all"
                      >
                        <Trash2 size={12} />
                      </button>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* 4. 调整记录 - 缩小行高 */}
      <section className="bg-white border rounded-2xl overflow-hidden shadow-sm">
        <div className="bg-gray-50 px-4 py-2 border-b flex items-center space-x-2">
          <History size={13} className="text-amber-500" />
          <span className="text-[11px] font-black text-gray-700 uppercase tracking-wider">当日处方动态调整记录</span>
        </div>
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-slate-50 border-b border-gray-100">
              <th className="px-4 py-1.5 text-[9px] font-bold text-gray-400 uppercase w-10 text-center">#</th>
              <th className="px-4 py-1.5 text-[9px] font-bold text-gray-400 uppercase">调整内容</th>
              <th className="px-4 py-1.5 text-[9px] font-bold text-gray-400 uppercase w-32">调整人</th>
              <th className="px-4 py-1.5 text-[9px] font-bold text-gray-400 uppercase w-40">调整时间</th>
              <th className="px-4 py-1.5 text-[9px] font-bold text-gray-400 uppercase w-16 text-right"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {[
              { id: 1, content: '目标超滤量从 2.2L 调整至 2.5L', person: '王志远 (医师)', time: '2026-02-04 07:45' },
              { id: 2, content: '抗凝剂注入速率根据患者凝血现状微调 +0.2ml/h', person: '刘护士长', time: '2026-02-04 08:12' },
            ].map((log) => (
              <tr key={log.id} className="hover:bg-amber-50/20 transition-colors">
                <td className="px-4 py-2 text-[11px] text-center font-bold text-gray-400">{log.id}</td>
                <td className="px-4 py-2 text-[11px] font-black text-slate-700 flex items-center">
                   <div className="w-1 h-1 rounded-full bg-amber-400 mr-2 shrink-0"></div>
                   {log.content}
                </td>
                <td className="px-4 py-2 text-[11px] text-slate-600 font-bold">{log.person}</td>
                <td className="px-4 py-2 text-[9px] text-gray-400 font-mono">{log.time}</td>
                <td className="px-4 py-2 text-right">
                  <button className="text-blue-600 hover:underline text-[9px] font-black uppercase">Detail</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      {/* 5. 底部功能栏 */}
      <div className="flex items-center justify-between pt-4 border-t border-gray-100 mt-2">
        <div className="flex items-center space-x-6">
          <div className="flex items-center space-x-2 text-[9px] text-gray-400 font-bold uppercase tracking-wider">
            <Clock size={11} className="text-gray-300" />
            <span>Update: 2026-02-04 08:12:05</span>
          </div>
          <div className="flex items-center space-x-2 text-[9px] text-gray-400 font-bold uppercase tracking-wider">
            <ClipboardList size={11} className="text-gray-300" />
            <span className={isEditing ? 'text-amber-500' : 'text-gray-400'}>
              Status: {isEditing ? 'Editing...' : 'Pending Verification'}
            </span>
          </div>
        </div>
        <div className="flex space-x-2">
          {isEditing ? (
            <>
              <button 
                onClick={() => setIsEditing(false)}
                className="px-5 py-2 border border-red-200 rounded-xl text-[10px] font-black text-red-500 hover:bg-red-50 transition-all flex items-center space-x-2"
              >
                <X size={12} />
                <span>CANCEL</span>
              </button>
              <button 
                onClick={handleSave}
                className="px-7 py-2 bg-emerald-600 text-white rounded-xl text-[10px] font-black hover:bg-emerald-700 shadow-lg shadow-emerald-100 flex items-center space-x-2 transition-all"
              >
                <Save size={12} />
                <span>SAVE & LOCK</span>
              </button>
            </>
          ) : (
            <>
              <button 
                onClick={() => setIsEditing(true)}
                className="px-5 py-2 border rounded-xl text-[10px] font-black text-gray-500 hover:bg-gray-50 transition-all shadow-sm flex items-center space-x-2"
              >
                <Settings size={12} className="text-gray-400" />
                <span>EDIT</span>
              </button>
              <button className="px-8 py-2 bg-blue-600 text-white rounded-xl text-[10px] font-black hover:bg-blue-700 shadow-lg shadow-blue-100 flex items-center space-x-2 transition-all transform active:scale-[0.98]">
                <span>VERIFY & EXECUTE</span>
                <ArrowRight size={12} />
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default TodayPrescription;
