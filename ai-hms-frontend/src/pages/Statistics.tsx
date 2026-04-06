import { useEffect, useMemo, useState } from 'react'
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
  Filter,
  Download
} from 'lucide-react'
import {
  restApi,
  type RestInfectionStatItem,
  type RestQualityStatItem,
  type RestVascularStatItem,
  type RestWorkloadStatItem,
} from '@/services/restClient'

type StatisticsTab = 'QUALITY' | 'INFECTION' | 'VASCULAR' | 'WORKLOAD'

export default function Statistics() {
  const { t } = useTranslation(['statistics'])
  const [activeTab, setActiveTab] = useState<StatisticsTab>('QUALITY')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [qualityData, setQualityData] = useState<RestQualityStatItem[]>([])
  const [infectionData, setInfectionData] = useState<RestInfectionStatItem[]>([])
  const [vascularData, setVascularData] = useState<RestVascularStatItem[]>([])
  const [workloadData, setWorkloadData] = useState<RestWorkloadStatItem[]>([])

  const currentYear = new Date().getFullYear()
  const currentYearMonth = `${currentYear}-${String(new Date().getMonth() + 1).padStart(2, '0')}`

  useEffect(() => {
    setLoading(true)
    setError('')

    Promise.all([
      restApi.getQualityStatistics({ year: currentYear }),
      restApi.getInfectionStatistics({ year: currentYear }),
      restApi.getVascularStatistics({ year: currentYear }),
      restApi.getWorkloadStatistics({ yearMonth: currentYearMonth }),
    ])
      .then(([qualityRes, infectionRes, vascularRes, workloadRes]) => {
        setQualityData(qualityRes.data.items || [])
        setInfectionData(infectionRes.data.items || [])
        setVascularData(vascularRes.data.items || [])
        setWorkloadData(workloadRes.data.items || [])
      })
      .catch(() => {
        setQualityData([])
        setInfectionData([])
        setVascularData([])
        setWorkloadData([])
        setError('数据加载失败，请刷新重试')
      })
      .finally(() => setLoading(false))
  }, [currentYear, currentYearMonth])

  const reportCategories = [
    { id: 'QUALITY' as const, label: t('statistics:tab.quality'), icon: Activity },
    { id: 'INFECTION' as const, label: t('statistics:tab.infection'), icon: AlertTriangle },
    { id: 'VASCULAR' as const, label: t('statistics:tab.vascular'), icon: FileText },
    { id: 'WORKLOAD' as const, label: t('statistics:tab.workload'), icon: TrendingUp },
  ]

  const monthLabel = (month: number) => {
    const map = [
      'statistics:month.jan', 'statistics:month.feb', 'statistics:month.mar', 'statistics:month.apr', 'statistics:month.may', 'statistics:month.jun',
      'statistics:month.jul', 'statistics:month.aug', 'statistics:month.sep', 'statistics:month.oct', 'statistics:month.nov', 'statistics:month.dec',
    ]
    return t(map[month - 1] || 'statistics:month.jan')
  }

  const qualityChartData = useMemo(() => qualityData.map(i => ({ ...i, month: monthLabel(i.month) })), [qualityData])
  const infectionChartData = useMemo(() => infectionData.map(i => ({ ...i, month: monthLabel(i.month) })), [infectionData])
  const vascularChartData = useMemo(() => vascularData.map(i => ({ ...i, month: monthLabel(i.month) })), [vascularData])

  const avg = (list: number[]) => list.length ? (list.reduce((a, b) => a + b, 0) / list.length).toFixed(1) : '--'
  const sum = (list: number[]) => list.reduce((a, b) => a + b, 0)

  const renderStatCards = () => {
    if (activeTab === 'QUALITY') {
      return [
        { label: t('statistics:quality.ktvRate'), value: qualityData.length ? `${avg(qualityData.map(i => i.ktv))}%` : '--' },
        { label: t('statistics:quality.hbRate'), value: qualityData.length ? `${avg(qualityData.map(i => i.hb))}%` : '--' },
        { label: t('statistics:quality.albRate'), value: qualityData.length ? `${avg(qualityData.map(i => i.alb))}%` : '--' },
        { label: t('statistics:quality.phosphorusRate'), value: '--' },
      ]
    }
    if (activeTab === 'INFECTION') {
      return [
        { label: t('statistics:infection.hbsagPositive'), value: `${sum(infectionData.map(i => i.hbsAg)) || '--'}` },
        { label: t('statistics:infection.hcvPositive'), value: `${sum(infectionData.map(i => i.hcv)) || '--'}` },
        { label: t('statistics:infection.hivPositive'), value: `${sum(infectionData.map(i => i.hiv)) || '--'}` },
        { label: t('statistics:infection.newPositive'), value: `${sum(infectionData.map(i => i.tp)) || '--'}` },
      ]
    }
    if (activeTab === 'VASCULAR') {
      const avf = sum(vascularData.map(i => i.avf))
      const avgValue = sum(vascularData.map(i => i.avg))
      const tcc = sum(vascularData.map(i => i.tcc))
      const total = avf + avgValue + tcc
      return [
        { label: t('statistics:vascular.avfRate'), value: total > 0 ? `${((avf / total) * 100).toFixed(1)}%` : '--' },
        { label: t('statistics:vascular.avgRate'), value: total > 0 ? `${((avgValue / total) * 100).toFixed(1)}%` : '--' },
        { label: t('statistics:vascular.catheterRate'), value: total > 0 ? `${((tcc / total) * 100).toFixed(1)}%` : '--' },
        { label: t('statistics:vascular.occlusionRate'), value: '--' },
      ]
    }
    const totalTreatments = sum(workloadData.map(i => i.treatments))
    const totalPunctures = sum(workloadData.map(i => i.punctures))
    return [
      { label: t('statistics:workload.totalTreatments'), value: totalTreatments || '--' },
      { label: t('statistics:workload.avgPerDay'), value: workloadData.length ? avg(workloadData.map(i => i.treatments)) : '--' },
      { label: t('statistics:workload.punctureSuccess'), value: totalPunctures || '--' },
      { label: t('statistics:workload.docCompletion'), value: '--' },
    ]
  }

  const renderChart = () => {
    if (activeTab === 'QUALITY') {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:quality.chartTitle')}</h3>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart data={qualityChartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{ fill: '#94a3b8' }} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: '#94a3b8' }} />
                <Tooltip />
                <Legend />
                <Area type="monotone" dataKey="ktv" name={t('statistics:quality.ktvLabel')} stroke="#3b82f6" fill="#dbeafe" strokeWidth={2} />
                <Line type="monotone" dataKey="hb" name={t('statistics:quality.hbLabel')} stroke="#8b5cf6" strokeWidth={2} />
                <Line type="monotone" dataKey="alb" name={t('statistics:quality.albLabel')} stroke="#10b981" strokeWidth={2} />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        </div>
      )
    }

    if (activeTab === 'INFECTION') {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:infection.chartTitle')}</h3>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={infectionChartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{ fill: '#94a3b8' }} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: '#94a3b8' }} />
                <Tooltip />
                <Legend />
                <Area type="monotone" dataKey="hbsAg" name={t('statistics:infection.hbsagLabel')} stroke="#ef4444" fill="#fef2f2" strokeWidth={2} />
                <Area type="monotone" dataKey="hcv" name={t('statistics:infection.hcvLabel')} stroke="#f97316" fill="#fff7ed" strokeWidth={2} />
                <Area type="monotone" dataKey="hiv" name={t('statistics:infection.hivLabel')} stroke="#8b5cf6" fill="#faf5ff" strokeWidth={2} />
                <Area type="monotone" dataKey="tp" name={t('statistics:infection.tpLabel')} stroke="#06b6d4" fill="#ecfeff" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )
    }

    if (activeTab === 'VASCULAR') {
      return (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          <h3 className="text-lg font-bold text-gray-800 mb-8">{t('statistics:vascular.chartTitle')}</h3>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={vascularChartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{ fill: '#94a3b8' }} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: '#94a3b8' }} />
                <Tooltip />
                <Legend />
                <Area type="monotone" dataKey="avf" name={t('statistics:vascular.avfLabel')} stroke="#3b82f6" fill="#dbeafe" strokeWidth={2} stackId="1" />
                <Area type="monotone" dataKey="avg" name={t('statistics:vascular.avgLabel')} stroke="#8b5cf6" fill="#ede9fe" strokeWidth={2} stackId="1" />
                <Area type="monotone" dataKey="tcc" name={t('statistics:vascular.catheterLabel')} stroke="#f97316" fill="#ffedd5" strokeWidth={2} stackId="1" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )
    }

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
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {workloadData.map((nurse) => (
                <tr key={nurse.userId} className="hover:bg-gray-50">
                  <td className="py-4 px-4 text-gray-900">{nurse.name || '--'}</td>
                  <td className="text-center py-4 px-4 text-gray-700">{nurse.treatments}</td>
                  <td className="text-center py-4 px-4 text-gray-700">{nurse.punctures}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col space-y-6 max-w-[1600px] mx-auto pb-10">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-800">{t('statistics:title')}</h2>
        <div className="flex gap-3">
          <button className="flex items-center px-4 py-2 bg-white border border-gray-200 rounded-xl text-sm font-medium hover:bg-gray-50 shadow-sm">
            <Filter size={16} className="mr-2" /> {t('statistics:action.filterDate')}
          </button>
          <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-xl text-sm font-bold hover:bg-blue-700 shadow-md">
            <Download size={16} className="mr-2" /> {t('statistics:action.exportReport')}
          </button>
        </div>
      </div>

      {loading ? (
        <div className="bg-white rounded-2xl border border-gray-100 p-10 text-center text-gray-500">Loading...</div>
      ) : error ? (
        <div className="bg-white rounded-2xl border border-red-200 p-10 text-center text-red-500">{error}</div>
      ) : (
        <>
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

          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            {renderStatCards().map((card) => (
              <div key={card.label} className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
                <p className="text-xs text-gray-400 font-bold mb-1">{card.label}</p>
                <h4 className="text-2xl font-bold text-blue-600">{card.value}</h4>
              </div>
            ))}
          </div>

          {(
            (activeTab === 'QUALITY' && qualityData.length === 0) ||
            (activeTab === 'INFECTION' && infectionData.length === 0) ||
            (activeTab === 'VASCULAR' && vascularData.length === 0) ||
            (activeTab === 'WORKLOAD' && workloadData.length === 0)
          ) ? (
            <div className="bg-white rounded-2xl border border-gray-100 p-10 text-center text-gray-400">暂无数据</div>
          ) : renderChart()}
        </>
      )}
    </div>
  )
}
