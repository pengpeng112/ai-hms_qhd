import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Line,
  Legend,
  ComposedChart
} from 'recharts'
import {
  FileText,
  TrendingUp,
  AlertTriangle,
  Activity,
  ArrowUpRight,
  ArrowDownRight,
  Filter,
  Download
} from 'lucide-react'

export default function Statistics() {
  const { t } = useTranslation(['statistics'])
  const [activeTab, setActiveTab] = useState('QUALITY')

  const reportCategories = [
    { id: 'QUALITY', label: t('statistics:tab.quality'), icon: Activity },
    { id: 'INFECTION', label: t('statistics:tab.infection'), icon: AlertTriangle },
    { id: 'VASCULAR', label: t('statistics:tab.vascular'), icon: FileText },
    { id: 'WORKLOAD', label: t('statistics:tab.workload'), icon: TrendingUp },
  ]

  // Quality data for charts
  const qualityData = [
    { month: t('statistics:month.jan'), ktv: 92, hb: 85, alb: 88 },
    { month: t('statistics:month.feb'), ktv: 94, hb: 88, alb: 90 },
    { month: t('statistics:month.mar'), ktv: 93, hb: 84, alb: 92 },
    { month: t('statistics:month.apr'), ktv: 96, hb: 90, alb: 91 },
    { month: t('statistics:month.may'), ktv: 95, hb: 87, alb: 89 },
    { month: t('statistics:month.jun'), ktv: 94, hb: 89, alb: 93 },
  ]

  // Infection markers data
  const infectionData = [
    { month: t('statistics:month.jan'), hbsag: 2, hcv: 1, hiv: 0, tp: 0 },
    { month: t('statistics:month.feb'), hbsag: 2, hcv: 1, hiv: 0, tp: 1 },
    { month: t('statistics:month.mar'), hbsag: 3, hcv: 1, hiv: 0, tp: 1 },
    { month: t('statistics:month.apr'), hbsag: 3, hcv: 2, hiv: 0, tp: 1 },
    { month: t('statistics:month.may'), hbsag: 3, hcv: 2, hiv: 0, tp: 1 },
    { month: t('statistics:month.jun'), hbsag: 3, hcv: 2, hiv: 0, tp: 1 },
  ]

  // Vascular access data
  const vascularData = [
    { month: t('statistics:month.jan'), avf: 85, avg: 10, tcc: 5 },
    { month: t('statistics:month.feb'), avf: 84, avg: 11, tcc: 5 },
    { month: t('statistics:month.mar'), avf: 86, avg: 10, tcc: 4 },
    { month: t('statistics:month.apr'), avf: 87, avg: 9, tcc: 4 },
    { month: t('statistics:month.may'), avf: 88, avg: 8, tcc: 4 },
    { month: t('statistics:month.jun'), avf: 88, avg: 9, tcc: 3 },
  ]

  // Nurse workload data
  const workloadData = [
    { name: '刘护士长', treatments: 45, punctures: 42, documentation: 95 },
    { name: '赵护士', treatments: 68, punctures: 65, documentation: 88 },
    { name: '孙护士', treatments: 72, punctures: 70, documentation: 92 },
    { name: '周护士', treatments: 65, punctures: 62, documentation: 85 },
    { name: '吴护士', treatments: 58, punctures: 55, documentation: 90 },
  ]

  // Render stat cards based on active tab
  const renderStatCards = () => {
    if (activeTab === 'QUALITY') {
      return (
        <>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:quality.ktvRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-blue-600">92.4%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 2.1%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:quality.hbRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-indigo-600">85.6%</h4>
              <span className="text-xs text-red-500 flex items-center font-bold"><ArrowDownRight size={12}/> 0.5%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:quality.albRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-emerald-600">89.2%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 1.3%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:quality.phosphorusRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-orange-600">78.5%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 0.8%</span>
            </div>
          </div>
        </>
      )
    } else if (activeTab === 'INFECTION') {
      return (
        <>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:infection.hbsagPositive')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-red-600">3 {t('statistics:unit.person')}</h4>
              <span className="text-xs text-gray-500 font-bold">{t('statistics:infection.ratio')} 2.1%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:infection.hcvPositive')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-orange-600">2 {t('statistics:unit.person')}</h4>
              <span className="text-xs text-gray-500 font-bold">{t('statistics:infection.ratio')} 1.4%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:infection.hivPositive')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-green-600">0 {t('statistics:unit.person')}</h4>
              <span className="text-xs text-gray-500 font-bold">{t('statistics:infection.ratio')} 0%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:infection.newPositive')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-green-600">0 {t('statistics:unit.person')}</h4>
              <span className="text-xs text-green-500 font-bold">{t('statistics:infection.stable')}</span>
            </div>
          </div>
        </>
      )
    } else if (activeTab === 'VASCULAR') {
      return (
        <>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:vascular.avfRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-blue-600">88%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 1.2%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:vascular.avgRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-indigo-600">9%</h4>
              <span className="text-xs text-gray-500 font-bold">{t('statistics:vascular.stable')}</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:vascular.catheterRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-orange-600">3%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowDownRight size={12}/> 0.5%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:vascular.occlusionRate')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-emerald-600">1.2%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowDownRight size={12}/> 0.3%</span>
            </div>
          </div>
        </>
      )
    } else {
      return (
        <>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:workload.totalTreatments')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-blue-600">308 {t('statistics:unit.times')}</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 5.2%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:workload.avgPerDay')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-indigo-600">12.8 {t('statistics:unit.times')}</h4>
              <span className="text-xs text-gray-500 font-bold">{t('statistics:workload.onTarget')}</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:workload.punctureSuccess')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-emerald-600">96.5%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 0.8%</span>
            </div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
            <p className="text-xs text-gray-400 font-bold mb-1">{t('statistics:workload.docCompletion')}</p>
            <div className="flex items-baseline justify-between">
              <h4 className="text-2xl font-bold text-orange-600">90.2%</h4>
              <span className="text-xs text-green-500 flex items-center font-bold"><ArrowUpRight size={12}/> 2.1%</span>
            </div>
          </div>
        </>
      )
    }
  }

  // Render chart based on active tab
  const renderChart = () => {
    if (activeTab === 'QUALITY') {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:quality.chartTitle')}</h3>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart data={qualityData}>
                <defs>
                  <linearGradient id="colorKtv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.1}/>
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{fill: '#94a3b8'}} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{fill: '#94a3b8'}} domain={[70, 100]} />
                <Tooltip contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1)'}} />
                <Legend />
                <Area type="monotone" dataKey="ktv" name={t('statistics:quality.ktvLabel')} stroke="#3b82f6" fillOpacity={1} fill="url(#colorKtv)" strokeWidth={3} />
                <Line type="monotone" dataKey="hb" name={t('statistics:quality.hbLabel')} stroke="#8b5cf6" strokeWidth={2} />
                <Line type="monotone" dataKey="alb" name={t('statistics:quality.albLabel')} stroke="#10b981" strokeWidth={2} />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        </div>
      )
    } else if (activeTab === 'INFECTION') {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:infection.chartTitle')}</h3>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={infectionData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{fill: '#94a3b8'}} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{fill: '#94a3b8'}} />
                <Tooltip contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1)'}} />
                <Legend />
                <Area type="monotone" dataKey="hbsag" name={t('statistics:infection.hbsagLabel')} stroke="#ef4444" fill="#fef2f2" strokeWidth={2} />
                <Area type="monotone" dataKey="hcv" name={t('statistics:infection.hcvLabel')} stroke="#f97316" fill="#fff7ed" strokeWidth={2} />
                <Area type="monotone" dataKey="hiv" name={t('statistics:infection.hivLabel')} stroke="#8b5cf6" fill="#faf5ff" strokeWidth={2} />
                <Area type="monotone" dataKey="tp" name={t('statistics:infection.tpLabel')} stroke="#06b6d4" fill="#ecfeff" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )
    } else if (activeTab === 'VASCULAR') {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:vascular.chartTitle')}</h3>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={vascularData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{fill: '#94a3b8'}} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{fill: '#94a3b8'}} domain={[0, 100]} />
                <Tooltip contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1)'}} />
                <Legend />
                <Area type="monotone" dataKey="avf" name={t('statistics:vascular.avfLabel')} stroke="#3b82f6" fill="#dbeafe" strokeWidth={2} stackId="1" />
                <Area type="monotone" dataKey="avg" name={t('statistics:vascular.avgLabel')} stroke="#8b5cf6" fill="#ede9fe" strokeWidth={2} stackId="1" />
                <Area type="monotone" dataKey="tcc" name={t('statistics:vascular.catheterLabel')} stroke="#f97316" fill="#ffedd5" strokeWidth={2} stackId="1" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )
    } else {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:workload.chartTitle')}</h3>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left py-3 px-4 text-sm font-bold text-gray-500">{t('statistics:workload.nurseName')}</th>
                  <th className="text-center py-3 px-4 text-sm font-bold text-gray-500">{t('statistics:workload.treatments')}</th>
                  <th className="text-center py-3 px-4 text-sm font-bold text-gray-500">{t('statistics:workload.punctures')}</th>
                  <th className="text-center py-3 px-4 text-sm font-bold text-gray-500">{t('statistics:workload.documentation')}</th>
                  <th className="text-center py-3 px-4 text-sm font-bold text-gray-500">{t('statistics:workload.efficiency')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {workloadData.map((nurse, index) => (
                  <tr key={index} className="hover:bg-gray-50">
                    <td className="py-4 px-4">
                      <div className="flex items-center">
                        <div className="w-8 h-8 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-xs font-bold mr-3">
                          {nurse.name[0]}
                        </div>
                        <span className="font-medium text-gray-900">{nurse.name}</span>
                      </div>
                    </td>
                    <td className="text-center py-4 px-4 text-gray-700">{nurse.treatments}</td>
                    <td className="text-center py-4 px-4 text-gray-700">{nurse.punctures}</td>
                    <td className="text-center py-4 px-4">
                      <span className={`px-2 py-1 rounded-full text-xs font-bold ${
                        nurse.documentation >= 90 ? 'bg-green-100 text-green-700' :
                        nurse.documentation >= 80 ? 'bg-yellow-100 text-yellow-700' :
                        'bg-red-100 text-red-700'
                      }`}>
                        {nurse.documentation}%
                      </span>
                    </td>
                    <td className="text-center py-4 px-4">
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-blue-600 h-2 rounded-full"
                          style={{ width: `${Math.min((nurse.treatments / 80) * 100, 100)}%` }}
                        ></div>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )
    }
  }

  return (
    <div className="h-full flex flex-col space-y-6 max-w-[1600px] mx-auto pb-10">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-800">{t('statistics:title')}</h2>
        <div className="flex gap-3">
          <button className="flex items-center px-4 py-2 bg-white border border-gray-200 rounded-xl text-sm font-medium hover:bg-gray-50 shadow-sm">
            <Filter size={16} className="mr-2"/> {t('statistics:action.filterDate')}
          </button>
          <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-xl text-sm font-bold hover:bg-blue-700 shadow-md">
            <Download size={16} className="mr-2"/> {t('statistics:action.exportReport')}
          </button>
        </div>
      </div>

      {/* Tab Switcher */}
      <div className="flex gap-2 p-1 bg-gray-200/50 rounded-2xl w-fit">
        {reportCategories.map(cat => (
          <button
            key={cat.id}
            onClick={() => setActiveTab(cat.id)}
            className={`flex items-center px-6 py-2.5 rounded-xl text-sm font-bold transition-all ${
              activeTab === cat.id ? 'bg-white text-blue-600 shadow-md' : 'text-gray-500 hover:bg-gray-100'
            }`}
          >
            <cat.icon size={16} className="mr-2" />
            {cat.label}
          </button>
        ))}
      </div>

      {/* Stat Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {renderStatCards()}
      </div>

      {/* Chart */}
      {renderChart()}
    </div>
  )
}
