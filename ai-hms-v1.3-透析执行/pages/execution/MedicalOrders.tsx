
import React, { useState, useMemo } from 'react';
import { Patient, MedicalOrder } from '../../types';
import { 
  Plus, 
  X, 
  ChevronDown, 
  Calendar, 
  FileText, 
  Trash2, 
  Edit3, 
  CheckCircle2,
  Activity,
  ClipboardList,
  Search,
  LayoutGrid,
  Zap,
  Clock,
  Check
} from 'lucide-react';

// 模拟医嘱项目库
const MASTER_ITEMS = [
  { id: 'm1', name: '左卡尼汀注射液', type: 'drug', spec: '1.0g/5ml', route: '静脉注射', unit: 'ml' },
  { id: 'm2', name: '蔗糖铁注射液', type: 'drug', spec: '100mg/5ml', route: '静脉注射', unit: 'ml' },
  { id: 'm3', name: '促红细胞生成素 (EPO)', type: 'drug', spec: '3000IU/支', route: '皮下注射', unit: '支' },
  { id: 'm4', name: '血液透析滤过 (HDF)', type: 'non-drug', spec: '每次', route: '体外循环', unit: '小时' },
  { id: 'm5', name: '血液透析 (HD)', type: 'non-drug', spec: '每次', route: '体外循环', unit: '小时' },
];

const MedicalOrders: React.FC<{ patient: Patient }> = ({ patient }) => {
  const [showModal, setShowModal] = useState(false);
  
  // 弹窗状态管理
  const [activeCategory, setActiveCategory] = useState<'long' | 'short'>('long'); // 长期 vs 临时
  const [entryMode, setEntryMode] = useState<'direct' | 'template'>('direct');   // 直接 vs 模板
  const [selectedItemId, setSelectedItemId] = useState<string>('');
  
  const selectedItem = useMemo(() => 
    MASTER_ITEMS.find(item => item.id === selectedItemId), 
    [selectedItemId]
  );
  
  const isDrug = selectedItem ? selectedItem.type === 'drug' : true;

  // 模拟医嘱列表数据
  const [orders, setOrders] = useState<any[]>([
    { id: 1, type: '长期', content: '血液透析滤过 (HDF)', usage: '体外循环, 4小时, 每次透析', doctor: '董婉颖', orderTime: '2026-01-14 09:00', recentExecution: '2026-02-02 08:30', checked: true, executed: true, weekCount: 2, lastModified: '2026-02-04 07:45' },
    { id: 2, type: '临时', content: '左卡尼汀注射液 1g', usage: '静脉注射, 1支, 每次', doctor: '王志远', orderTime: '2026-02-04 08:15', recentExecution: '-', checked: false, executed: false, weekCount: 0, lastModified: '2026-02-04 08:15' }
  ]);

  const handleAddOrder = () => {
    if(!selectedItem && entryMode === 'direct') return;
    const newOrder: any = {
      id: orders.length + 1,
      type: activeCategory === 'long' ? '长期' : '临时',
      content: selectedItem?.name || '模板医嘱',
      usage: isDrug ? `${selectedItem?.route || '静脉'}, 1支` : '4小时, 每次',
      doctor: '刘护士长',
      orderTime: new Date().toLocaleString().slice(0, 16),
      recentExecution: '-',
      checked: false,
      executed: false,
      weekCount: 0,
      lastModified: new Date().toLocaleString().slice(0, 16)
    };
    setOrders([newOrder, ...orders]);
    setShowModal(false);
    setSelectedItemId('');
  };

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500">
      
      {/* 1. 操作顶栏 */}
      <div className="flex items-center justify-end bg-white p-3 rounded-2xl border border-gray-100 shadow-sm">
        <button 
          onClick={() => setShowModal(true)}
          className="bg-blue-600 text-white px-5 py-2 rounded-xl text-xs font-black shadow-lg shadow-blue-100 flex items-center space-x-2 hover:bg-blue-700 transition-all active:scale-95"
        >
          <Plus size={16} />
          <span>新增透析医嘱</span>
        </button>
      </div>

      {/* 2. 医嘱明细列表 - 支持横向拖动 */}
      <div className="space-y-3">
        <div className="flex items-center justify-between px-1">
          <div className="flex items-center space-x-2">
            <FileText size={16} className="text-blue-500" />
            <h3 className="text-xs font-black text-slate-800 tracking-tight">透析医嘱明细</h3>
          </div>
          <span className="text-[9px] text-gray-400 font-bold flex items-center uppercase tracking-tighter">
            <Search size={10} className="mr-1" />
            Horizontal scroll enabled for details
          </span>
        </div>
        
        <div className="bg-white border border-gray-100 rounded-2xl shadow-sm overflow-hidden flex flex-col group">
          <div className="overflow-x-auto scroll-smooth custom-scrollbar">
            <table className="w-full text-left border-collapse min-w-[1050px] table-fixed">
              <thead>
                <tr className="bg-slate-50 border-b border-gray-100">
                  <th className="px-2 py-3 text-[10px] font-black text-gray-400 uppercase w-[50px] text-center sticky left-0 bg-slate-50 z-10 border-r border-gray-100">序号</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[60px]">类型</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[160px]">医嘱内容</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[160px]">使用描述</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[140px]">医生/下嘱时间</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[120px]">最近执行</th>
                  <th className="px-2 py-3 text-[10px] font-black text-gray-400 uppercase w-[60px] text-center">核对</th>
                  <th className="px-2 py-3 text-[10px] font-black text-gray-400 uppercase w-[60px] text-center">执行</th>
                  <th className="px-2 py-3 text-[10px] font-black text-gray-400 uppercase w-[90px] text-center">本周执行</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[130px]">最后修改</th>
                  <th className="px-3 py-3 text-[10px] font-black text-gray-400 uppercase w-[80px] text-center sticky right-0 bg-slate-50 shadow-[-4px_0_10px_rgba(0,0,0,0.02)] z-10">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {orders.map((order, index) => (
                  <tr key={order.id} className="hover:bg-blue-50/20 transition-colors group/row">
                    <td className="px-2 py-3.5 text-xs font-bold text-gray-400 text-center sticky left-0 bg-white group-hover/row:bg-blue-50/20 z-10 border-r border-gray-100">{index + 1}</td>
                    <td className="px-3 py-3.5">
                      <span className={`px-2 py-0.5 rounded-[4px] text-[9px] font-black ${
                        order.type === '长期' ? 'bg-indigo-100 text-indigo-700' : 'bg-fuchsia-100 text-fuchsia-700'
                      }`}>
                        {order.type}
                      </span>
                    </td>
                    <td className="px-3 py-3.5 text-xs font-black text-slate-800 truncate" title={order.content}>{order.content}</td>
                    <td className="px-3 py-3.5 text-[10px] text-slate-500 font-medium truncate" title={order.usage}>{order.usage}</td>
                    <td className="px-3 py-3.5">
                      <div className="text-[10px] font-black text-slate-700 leading-tight">{order.doctor}</div>
                      <div className="text-[8px] text-gray-400 font-mono">{order.orderTime}</div>
                    </td>
                    <td className="px-3 py-3.5 text-[10px] text-slate-400 font-mono italic">{order.recentExecution}</td>
                    <td className="px-2 py-3.5 text-center">
                      {order.checked ? <CheckCircle2 size={15} className="text-emerald-500 mx-auto" /> : <div className="w-1.5 h-1.5 rounded-full bg-gray-200 mx-auto"></div>}
                    </td>
                    <td className="px-2 py-3.5 text-center">
                      {order.executed ? <CheckCircle2 size={15} className="text-emerald-500 mx-auto" /> : <div className="w-1.5 h-1.5 rounded-full bg-gray-200 mx-auto"></div>}
                    </td>
                    <td className="px-2 py-3.5 text-center text-xs font-black text-slate-700">{order.weekCount} <span className="text-[9px] text-gray-400">次</span></td>
                    <td className="px-3 py-3.5 text-[10px] text-gray-400 font-mono">{order.lastModified}</td>
                    <td className="px-3 py-3.5 text-center sticky right-0 bg-white group-hover/row:bg-blue-50/20 shadow-[-4px_0_10px_rgba(0,0,0,0.02)] z-10">
                      <div className="flex items-center justify-center space-x-1">
                        <button className="p-1 text-gray-400 hover:text-blue-600 rounded transition-colors"><Edit3 size={13} /></button>
                        <button className="p-1 text-gray-400 hover:text-red-500 rounded transition-colors"><Trash2 size={13} /></button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {/* 底部横向滚动提示条 */}
          <div className="h-[2px] bg-gray-100 w-full relative">
             <div className="absolute top-0 left-0 h-full bg-blue-500 w-24 rounded-full"></div>
          </div>
          <style dangerouslySetInnerHTML={{ __html: `
            .custom-scrollbar::-webkit-scrollbar { height: 6px; }
            .custom-scrollbar::-webkit-scrollbar-track { background: #f8fafc; }
            .custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; border: 1px solid #f8fafc; }
            .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #cbd5e1; }
          `}} />
        </div>
      </div>

      {/* 3. 新增透析医嘱弹窗 - 重构后的多级布局 */}
      {showModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-300">
          <div className="bg-white w-[780px] rounded-[36px] shadow-2xl overflow-hidden flex flex-col animate-in zoom-in-95 duration-300 border border-white/20 max-h-[90vh]">
            
            {/* Top Switcher: Long vs Short */}
            <div className="px-10 pt-10 pb-6 flex items-center justify-between bg-white relative">
              <div className="flex bg-slate-100 p-1.5 rounded-2xl shadow-inner border border-slate-200/50">
                <button 
                  onClick={() => setActiveCategory('long')}
                  className={`flex items-center space-x-3 px-10 py-3 rounded-xl text-xs font-black transition-all ${
                    activeCategory === 'long' ? 'bg-white text-indigo-600 shadow-xl shadow-indigo-100 ring-1 ring-slate-200' : 'text-gray-400 hover:text-gray-500'
                  }`}
                >
                  <Clock size={16} />
                  <span>新增长期医嘱</span>
                </button>
                <button 
                  onClick={() => setActiveCategory('short')}
                  className={`flex items-center space-x-3 px-10 py-3 rounded-xl text-xs font-black transition-all ${
                    activeCategory === 'short' ? 'bg-white text-fuchsia-600 shadow-xl shadow-fuchsia-100 ring-1 ring-slate-200' : 'text-gray-400 hover:text-gray-500'
                  }`}
                >
                  <Zap size={16} />
                  <span>新增临时医嘱</span>
                </button>
              </div>
              <button onClick={() => setShowModal(false)} className="p-2.5 hover:bg-red-50 rounded-full text-gray-300 hover:text-red-500 transition-all transform hover:rotate-90">
                <X size={24} />
              </button>
            </div>

            {/* Entry Mode Switcher: Direct vs Template */}
            <div className="px-10 pb-4 border-b border-gray-50 flex items-center space-x-10">
               <button 
                onClick={() => setEntryMode('direct')}
                className={`relative pb-3 text-[13px] font-black transition-all ${
                  entryMode === 'direct' ? 'text-blue-600' : 'text-gray-400'
                }`}
              >
                直接新增录入
                {entryMode === 'direct' && <div className="absolute bottom-0 left-0 w-full h-1 bg-blue-600 rounded-full animate-in slide-in-from-left duration-300"></div>}
              </button>
              <button 
                onClick={() => setEntryMode('template')}
                className={`relative pb-3 text-[13px] font-black transition-all ${
                  entryMode === 'template' ? 'text-blue-600' : 'text-gray-400'
                }`}
              >
                从模板组调取
                {entryMode === 'template' && <div className="absolute bottom-0 left-0 w-full h-1 bg-blue-600 rounded-full animate-in slide-in-from-left duration-300"></div>}
              </button>
            </div>

            {/* Content Area */}
            <div className="p-10 space-y-6 flex-1 overflow-y-auto no-scrollbar bg-slate-50/30">
              
              {entryMode === 'direct' ? (
                <div className="animate-in fade-in slide-in-from-bottom-2 duration-300 space-y-6">
                  {/* 项目选择 */}
                  <div className="flex items-center space-x-6">
                    <label className="w-24 text-xs font-black text-slate-500 text-right uppercase tracking-widest">医嘱项目</label>
                    <div className="flex-1 relative">
                      <select 
                        value={selectedItemId}
                        onChange={(e) => setSelectedItemId(e.target.value)}
                        className="w-full h-12 px-5 border border-slate-200 rounded-2xl text-[13px] font-black text-slate-800 outline-none bg-white appearance-none focus:ring-4 focus:ring-blue-100 transition-all shadow-sm"
                      >
                        <option value="">点击搜索并选择医嘱项...</option>
                        {MASTER_ITEMS.map(item => <option key={item.id} value={item.id}>{item.name} ({item.spec})</option>)}
                      </select>
                      <ChevronDown size={16} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
                    </div>
                  </div>

                  {/* 动态细节表单 */}
                  <div className="ml-30 grid grid-cols-2 gap-x-8 gap-y-5">
                    {isDrug ? (
                      <>
                        <div className="space-y-1.5">
                          <label className="text-[10px] font-black text-slate-400 uppercase tracking-[2px]">单次剂量</label>
                          <div className="flex items-center space-x-2">
                            <input className="flex-1 h-11 px-5 border border-slate-200 rounded-xl text-xs font-black bg-white focus:border-blue-400 outline-none" placeholder="数值" />
                            <span className="text-[11px] text-slate-400 font-bold">{selectedItem?.unit || 'ml'}</span>
                          </div>
                        </div>
                        <div className="space-y-1.5">
                          <label className="text-[10px] font-black text-slate-400 uppercase tracking-[2px]">给药途径</label>
                          <select className="w-full h-11 px-5 border border-slate-200 rounded-xl text-xs font-black bg-white outline-none appearance-none">
                            <option>{selectedItem?.route || '静脉注射'}</option>
                            <option>皮下注射</option>
                            <option>体外循环入壶</option>
                          </select>
                        </div>
                      </>
                    ) : (
                      <div className="col-span-2 space-y-1.5">
                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-[2px]">执行标准/时长</label>
                        <div className="flex items-center space-x-2">
                          <input className="flex-1 h-11 px-5 border border-slate-200 rounded-xl text-xs font-black bg-white" placeholder="例如：4" />
                          <span className="text-[11px] text-slate-400 font-bold">小时 / 每次</span>
                        </div>
                      </div>
                    )}
                    
                    <div className="space-y-1.5">
                      <label className="text-[10px] font-black text-slate-400 uppercase tracking-[2px]">执行频次</label>
                      <select className="w-full h-11 px-5 border border-slate-200 rounded-xl text-xs font-black bg-white appearance-none">
                        <option>每次透析 (QHD)</option>
                        <option>每周一次 (QW)</option>
                        <option>仅限本次 (ST)</option>
                      </select>
                    </div>
                    <div className="space-y-1.5">
                      <label className="text-[10px] font-black text-slate-400 uppercase tracking-[2px]">下嘱日期</label>
                      <div className="relative">
                        <input type="text" defaultValue="2026-02-04" className="w-full h-11 px-5 border border-slate-200 rounded-xl text-xs font-black bg-white outline-none" />
                        <Calendar size={14} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300" />
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center space-x-6 pt-4 border-t border-dashed border-slate-200">
                    <label className="w-24 text-xs font-black text-slate-500 text-right">备注说明</label>
                    <div className="flex-1 relative">
                      <input className="w-full h-12 px-5 border border-slate-200 rounded-2xl text-xs font-medium bg-white outline-none focus:ring-4 focus:ring-slate-100" placeholder="录入备注细节或注意事项..." />
                      <ClipboardList size={16} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300" />
                    </div>
                  </div>
                </div>
              ) : (
                <div className="animate-in fade-in slide-in-from-bottom-2 duration-300 grid grid-cols-2 gap-6">
                  {/* 模拟模板组数据 */}
                  {[
                    { title: '标准诱导期模板', content: 'HD + 1级护理 + 2.5h', icon: '⚡' },
                    { title: '高通量HDF常规模板', content: 'HDF + 普通肝素 + 4h', icon: '💎' },
                    { title: '无肝素透析套餐', content: 'HD + 盐水冲管 + 3.5h', icon: '🛡️' },
                    { title: '常态贫血治疗组', content: '左卡 + 蔗糖铁 + EPO', icon: '💊' }
                  ].map((tpl, i) => (
                    <div 
                      key={i} 
                      className="p-6 bg-white border border-slate-100 rounded-3xl hover:border-blue-400 hover:shadow-xl hover:shadow-blue-50 transition-all cursor-pointer group relative overflow-hidden"
                    >
                      <div className="flex items-start justify-between mb-3">
                         <div className="w-10 h-10 bg-blue-50 rounded-2xl flex items-center justify-center text-xl group-hover:scale-110 transition-transform">{tpl.icon}</div>
                         <div className="w-5 h-5 border-2 border-slate-200 rounded-full flex items-center justify-center group-hover:border-blue-500 group-hover:bg-blue-500 transition-all">
                           <Check size={12} className="text-white scale-0 group-hover:scale-100 transition-transform" />
                         </div>
                      </div>
                      <h4 className="text-[13px] font-black text-slate-800 mb-1">{tpl.title}</h4>
                      <p className="text-[10px] text-slate-400 font-bold uppercase tracking-tight">{tpl.content}</p>
                      
                      <div className="absolute top-0 right-0 w-16 h-16 bg-blue-50/30 rounded-full -mr-8 -mt-8 scale-0 group-hover:scale-100 transition-transform duration-500"></div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="px-10 py-8 bg-white border-t border-slate-50 flex items-center justify-between">
              <div className="flex items-center space-x-4 bg-slate-50 px-6 py-3 rounded-[24px] border border-slate-100 shadow-inner">
                 <div className={`w-9 h-9 rounded-xl flex items-center justify-center text-white text-[13px] font-black shadow-lg ${
                   activeCategory === 'long' ? 'bg-indigo-600 shadow-indigo-100' : 'bg-fuchsia-600 shadow-fuchsia-100'
                 }`}>
                   {patient.name.substring(0, 1)}
                 </div>
                 <div className="flex flex-col">
                   <span className="text-[11px] font-black text-slate-700 leading-tight">拟开嘱医生: 董婉颖</span>
                   <span className="text-[9px] text-gray-400 uppercase tracking-widest font-bold">Waiting for signature</span>
                 </div>
              </div>
              <div className="flex space-x-4">
                <button 
                  onClick={() => setShowModal(false)} 
                  className="px-10 py-4 bg-white border border-slate-200 text-xs font-black text-gray-400 rounded-2xl hover:bg-slate-50 transition-all active:scale-95"
                >
                  取消返回
                </button>
                <button 
                  onClick={handleAddOrder}
                  className="px-16 py-4 bg-blue-600 text-white text-xs font-black rounded-2xl hover:bg-blue-700 shadow-2xl shadow-blue-100 transition-all transform active:scale-[0.98]"
                >
                  {entryMode === 'template' ? '应用模板并提交' : '确认并保存医嘱'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 4. 抗凝剂列表 - 保持不变 */}
      <div className="space-y-3 pt-2">
        <div className="flex items-center space-x-2 px-1">
          <Activity size={16} className="text-amber-500" />
          <h3 className="text-xs font-black text-slate-800 tracking-tight">抗凝剂</h3>
        </div>
        <div className="bg-white border border-gray-100 rounded-2xl shadow-sm overflow-hidden">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-amber-50/30 border-b border-gray-100">
                <th className="px-6 py-3 text-[10px] font-black text-gray-400 uppercase w-32">类别</th>
                <th className="px-6 py-3 text-[10px] font-black text-gray-400 uppercase w-48">名称</th>
                <th className="px-6 py-3 text-[10px] font-black text-gray-400 uppercase w-48">剂量</th>
                <th className="px-6 py-3 text-[10px] font-black text-gray-400 uppercase w-40">医生</th>
                <th className="px-6 py-3 text-[10px] font-black text-gray-400 uppercase w-40">执行</th>
                <th className="px-6 py-3 text-[10px] font-black text-gray-400 uppercase w-40">核对</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {[
                { category: '首剂', name: '普通肝素', dose: '2000 IU', doctor: '董婉颖', executor: '刘护士长', checker: '张护士' },
                { category: '维持', name: '普通肝素', dose: '1000 IU/h', doctor: '董婉颖', executor: '刘护士长', checker: '张护士' }
              ].map((item, idx) => (
                <tr key={idx} className="hover:bg-amber-50/10 transition-colors">
                  <td className="px-6 py-4">
                    <span className={`px-2 py-0.5 rounded text-[10px] font-black ${
                      item.category === '首剂' ? 'bg-amber-100 text-amber-700' : 'bg-orange-100 text-orange-700'
                    }`}>
                      {item.category}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-xs font-black text-slate-700">{item.name}</td>
                  <td className="px-6 py-4 text-xs font-black text-blue-600 tracking-tight">{item.dose}</td>
                  <td className="px-6 py-4 text-xs font-bold text-slate-600">{item.doctor}</td>
                  <td className="px-6 py-4 text-xs font-bold text-slate-600">{item.executor}</td>
                  <td className="px-6 py-4 text-xs font-bold text-slate-600">{item.checker}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default MedicalOrders;
