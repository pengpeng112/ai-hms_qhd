// TreatmentConfig - 诊疗配置中心
// 包含：方案模版、医嘱模版、材料目录、药品目录

import { useState, useEffect, startTransition, useMemo } from 'react'
import { FileText, Syringe, Container, Pill, CheckCircle2, Clock } from 'lucide-react'
import { dictCache, DICT_TYPES } from '@/services/dictApi'
import { PlanTab, OrderTab, MaterialTab, DrugTab } from '@pages/TreatmentConfig/tabs'

type SubView = 'PLAN' | 'ORDER' | 'MATERIAL' | 'DRUG'

// Tab 配置常量
const TABS = [
  { id: 'PLAN' as const, label: '方案模板', icon: FileText },
  { id: 'ORDER' as const, label: '医嘱模板', icon: Syringe },
  { id: 'MATERIAL' as const, label: '材料目录', icon: Container },
  { id: 'DRUG' as const, label: '药品目录', icon: Pill },
] as const

export default function TreatmentConfig() {
  const [activeSubView, setActiveSubView] = useState<SubView>('PLAN')

  // 字典数据状态
  const [dictOptions, setDictOptions] = useState<Record<string, Array<{ value: string; label: string }>>>({})

  // 加载字典数据
  const loadDictData = async (forceRefresh = false) => {
    try {
      const options = await Promise.all([
        dictCache.getOptions(DICT_TYPES.DIALYSIS_MODE, forceRefresh),
        dictCache.getOptions(DICT_TYPES.ANTICOAGULANT, forceRefresh),
        dictCache.getOptions(DICT_TYPES.DIALYSATE_TYPE, forceRefresh),
        dictCache.getOptions(DICT_TYPES.DIALYSATE_GROUP, forceRefresh),
        dictCache.getOptions(DICT_TYPES.DIALYSATE_FLOW, forceRefresh),
        dictCache.getOptions(DICT_TYPES.GLUCOSE, forceRefresh),
        dictCache.getOptions(DICT_TYPES.MATERIAL_CATEGORY, forceRefresh),
        dictCache.getOptions(DICT_TYPES.DRUG_CATEGORY, forceRefresh),
        dictCache.getOptions(DICT_TYPES.ORDER_TYPE, forceRefresh),
        dictCache.getOptions(DICT_TYPES.ORDER_CATEGORY, forceRefresh),
        dictCache.getOptions(DICT_TYPES.ORDER_ROUTE, forceRefresh),
        dictCache.getOptions(DICT_TYPES.ORDER_FREQUENCY, forceRefresh),
        dictCache.getOptions(DICT_TYPES.ORDER_TIMING, forceRefresh),
      ])

      // 使用 startTransition 标记为非紧急更新，避免级联渲染
      startTransition(() => {
        setDictOptions({
          [DICT_TYPES.DIALYSIS_MODE]: options[0],
          [DICT_TYPES.ANTICOAGULANT]: options[1],
          [DICT_TYPES.DIALYSATE_TYPE]: options[2],
          [DICT_TYPES.DIALYSATE_GROUP]: options[3],
          [DICT_TYPES.DIALYSATE_FLOW]: options[4],
          [DICT_TYPES.GLUCOSE]: options[5],
          [DICT_TYPES.MATERIAL_CATEGORY]: options[6],
          [DICT_TYPES.DRUG_CATEGORY]: options[7],
          [DICT_TYPES.ORDER_TYPE]: options[8],
          [DICT_TYPES.ORDER_CATEGORY]: options[9],
          [DICT_TYPES.ORDER_ROUTE]: options[10],
          [DICT_TYPES.ORDER_FREQUENCY]: options[11],
          [DICT_TYPES.ORDER_TIMING]: options[12],
        })
      })
    } catch (error) {
      console.error('加载字典数据失败:', error)
    }
  }

  // 初始化时加载字典数据
  useEffect(() => {
    loadDictData()
  }, [])

  const handleRefreshDict = () => loadDictData(true)

  // 缓存当前日期字符串，避免每次渲染创建新对象
  const currentDate = useMemo(() => new Date().toLocaleDateString(), [])

  // 渲染内容
  const renderContent = () => {
    switch (activeSubView) {
      case 'PLAN':
        return <PlanTab dictOptions={dictOptions} onRefreshDict={handleRefreshDict} />
      case 'ORDER':
        return <OrderTab dictOptions={dictOptions} onRefreshDict={handleRefreshDict} />
      case 'MATERIAL':
        return <MaterialTab dictOptions={dictOptions} onRefreshDict={handleRefreshDict} />
      case 'DRUG':
        return <DrugTab dictOptions={dictOptions} onRefreshDict={handleRefreshDict} />
    }
  }

  return (
    <div className="h-full flex flex-col bg-slate-50/50">
       <div className="px-8 py-3 bg-white border-b border-slate-100 flex items-center gap-4 shrink-0 shadow-sm z-10">
          <div className="flex bg-slate-100/80 p-1 rounded-2xl shadow-inner">
             {TABS.map(tab => (
               <button
                 key={tab.id}
                 onClick={() => setActiveSubView(tab.id)}
                 className={`flex items-center gap-2 px-6 py-2 rounded-xl text-xs font-black transition-all ${activeSubView === tab.id ? 'bg-white text-blue-600 shadow-md ring-1 ring-slate-200' : 'text-slate-400 hover:text-slate-600'}`}
               >
                 <tab.icon size={14} />
                 {tab.label}
               </button>
             ))}
          </div>
          <div className="h-6 w-px bg-slate-200 mx-2"></div>
          <div className="flex items-center gap-4 text-[11px] font-bold text-slate-400">
             <span className="flex items-center gap-1.5"><CheckCircle2 size={14} className="text-emerald-500"/> 系统正常</span>
             <span className="flex items-center gap-1.5"><Clock size={14}/> {currentDate}</span>
          </div>
       </div>

       {renderContent()}
    </div>
  )
}
