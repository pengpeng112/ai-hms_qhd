/**
 * 病区概览页面
 *
 * 提供科室运行全景视图，包括：
 * - 核心指标卡片（患者、工作量、设备、时长）
 * - 治疗进度统计（饼图）
 * - 设备与床位运行矩阵
 */

import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Users, Monitor, Activity, CheckCircle2,
  AlertCircle, Clock, Battery, RefreshCw,
} from 'lucide-react'
import {
  PieChart, Pie, Cell, ResponsiveContainer, Tooltip,
} from 'recharts'
import { restApi } from '@/services'
import { getAllEquipments } from '@/services/equipment'
import { getActiveShifts } from '@/services/schedule'

// 颜色配置
const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444']

// 治疗状态类型
type TreatmentStatus = 'waiting' | 'dialysis' | 'disinfect' | 'completed'

interface ProcessDataItem {
  name: string
  value: number
  status: TreatmentStatus
}

interface BedStatus {
  id: string
  label: string
  status: 'active' | 'empty' | 'alarm' | 'maintenance'
}

export default function WardOverview() {
  const { t, i18n } = useTranslation('wardOverview')
  const [loading, setLoading] = useState(true)
  const [lastUpdate, setLastUpdate] = useState(new Date())

  // 统计数据
  const [stats, setStats] = useState({
    scheduledPatients: 0,
    totalCapacity: 50,
    workCompletion: 85,
    deviceUtilization: 92.8,
    avgDialysisHours: 3.8,
    onSiteCount: 0,
    alarmDevices: 0,
  })

  // 治疗进度数据
  const [processData, setProcessData] = useState<ProcessDataItem[]>([
    { name: t('process.waiting'), value: 0, status: 'waiting' },
    { name: t('process.dialysis'), value: 0, status: 'dialysis' },
    { name: t('process.disinfect'), value: 0, status: 'disinfect' },
    { name: t('process.completed'), value: 0, status: 'completed' },
  ])

  // 床位状态数据
  const [bedStatuses, setBedStatuses] = useState<BedStatus[]>([])

  // 加载数据
  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      // 并行请求多个 API
      const [patientsRes, equipmentsRes, shiftsRes, dashboardRes] = await Promise.all([
        restApi.getPatientList({ page: 1, pageSize: 100 }).catch(() => null),
        getAllEquipments().catch(() => null),
        getActiveShifts().catch(() => null),
        restApi.getDashboardStats().catch(() => null),
      ])

      // 计算统计数据
      const scheduledPatients = shiftsRes?.length || dashboardRes?.todaySchedules || 0
      const totalEquipments = equipmentsRes?.length || dashboardRes?.equipmentCount || 0
      const totalPatients = patientsRes?.data?.pagination?.total || dashboardRes?.activePatients || 0

      // 使用 Dashboard Stats 的真实数据
      const activeCount = dashboardRes?.todaySchedules || scheduledPatients
      const alarmCount = dashboardRes?.alertItems || 0

      setStats({
        scheduledPatients: activeCount,
        totalCapacity: totalEquipments,
        workCompletion: dashboardRes?.todayTreatments ? Math.round((dashboardRes.todayTreatments / Math.max(activeCount, 1)) * 100) : 0,
        deviceUtilization: totalEquipments > 0
          ? Math.round((activeCount / totalEquipments) * 1000) / 10
          : 0,
        avgDialysisHours: 3.8, // 暂无真实数据，保留默认值
        onSiteCount: totalPatients,
        alarmDevices: alarmCount,
      })

      // 治疗进度数据使用真实数据
      const waitingCount = Math.max(0, activeCount - (dashboardRes?.todayTreatments || 0))
      const completedCount = dashboardRes?.todayTreatments || 0

      setProcessData([
        { name: t('process.waiting'), value: waitingCount, status: 'waiting' },
        { name: t('process.dialysis'), value: activeCount, status: 'dialysis' },
        { name: t('process.disinfect'), value: 0, status: 'disinfect' },
        { name: t('process.completed'), value: completedCount, status: 'completed' },
      ])

      // 生成床位状态矩阵
      const beds: BedStatus[] = Array.from({ length: totalEquipments || 32 }).map((_, i) => {
        // 基于设备数量分配状态
        let status: BedStatus['status'] = 'active'

        if (i < alarmCount) {
          status = 'alarm'
        } else if (i >= activeCount) {
          status = 'empty'
        }

        return {
          id: `bed-${i}`,
          label: `A${String(i + 1).padStart(2, '0')}`,
          status,
        }
      })
      setBedStatuses(beds)

      setLastUpdate(new Date())
    } catch (error) {
      console.error(t('loadError'), error)
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleRefresh = () => {
    loadData()
  }

  // 获取床位状态样式
  const getBedStatusStyles = (status: BedStatus['status']) => {
    switch (status) {
      case 'alarm':
        return 'bg-red-50 border-red-200 text-red-600 animate-pulse'
      case 'empty':
        return 'bg-gray-50 border-gray-100 text-gray-400'
      case 'maintenance':
        return 'bg-orange-50 border-orange-200 text-orange-600'
      case 'active':
      default:
        return 'bg-blue-50 border-blue-100 text-blue-600'
    }
  }

  // 格式化时间
  const formatTime = (date: Date) => {
    const locale = i18n.language === 'en-US' ? 'en-US' : 'zh-CN'
    return date.toLocaleString(locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // 转换数据为 Recharts 格式
  const chartData = processData.map(item => ({
    name: item.name,
    value: item.value,
  }))

  return (
    <div className="h-full flex flex-col space-y-6 max-w-[1600px] mx-auto pb-10">
      {/* Header */}
      <div className="flex justify-between items-end">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">{t('title')}</h2>
          <p className="text-sm text-gray-500 mt-1">
            {t('lastUpdate', { time: formatTime(lastUpdate) })}
          </p>
        </div>
        <div className="flex gap-4">
          <button
            onClick={handleRefresh}
            disabled={loading}
            className="bg-white px-4 py-2 rounded-lg border border-gray-200 shadow-sm flex items-center hover:bg-gray-50 transition-colors disabled:opacity-50"
          >
            <RefreshCw size={16} className={`mr-2 ${loading ? 'animate-spin' : ''}`} />
            <span className="text-sm font-medium">{t('refresh')}</span>
          </button>
          <div className="bg-white px-4 py-2 rounded-lg border border-gray-200 shadow-sm flex items-center">
            <div className="w-2 h-2 rounded-full bg-green-500 mr-2 animate-pulse"></div>
            <span className="text-sm font-bold">{t('onSiteCount', { count: stats.onSiteCount })}</span>
          </div>
          {stats.alarmDevices > 0 && (
            <div className="bg-white px-4 py-2 rounded-lg border border-red-200 shadow-sm flex items-center">
              <AlertCircle size={16} className="text-red-500 mr-2" />
              <span className="text-sm font-bold text-red-600">{t('alarmDevices', { count: stats.alarmDevices })}</span>
            </div>
          )}
        </div>
      </div>

      {/* 核心指标卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 flex items-center">
          <div className="p-4 bg-blue-50 rounded-2xl mr-4">
            <Users className="text-blue-600" size={24} />
          </div>
          <div>
            <p className="text-xs text-gray-400 font-bold uppercase">{t('stats.scheduledPatients')}</p>
            <p className="text-2xl font-bold text-gray-800">
              {stats.scheduledPatients}
              <span className="text-xs font-normal text-gray-400"> / {stats.totalCapacity}</span>
            </p>
          </div>
        </div>

        <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 flex items-center">
          <div className="p-4 bg-green-50 rounded-2xl mr-4">
            <CheckCircle2 className="text-green-600" size={24} />
          </div>
          <div>
            <p className="text-xs text-gray-400 font-bold uppercase">{t('stats.workCompletion')}</p>
            <p className="text-2xl font-bold text-gray-800">{stats.workCompletion}%</p>
          </div>
        </div>

        <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 flex items-center">
          <div className="p-4 bg-purple-50 rounded-2xl mr-4">
            <Battery className="text-purple-600" size={24} />
          </div>
          <div>
            <p className="text-xs text-gray-400 font-bold uppercase">{t('stats.deviceUtilization')}</p>
            <p className="text-2xl font-bold text-gray-800">{stats.deviceUtilization}%</p>
          </div>
        </div>

        <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 flex items-center">
          <div className="p-4 bg-orange-50 rounded-2xl mr-4">
            <Clock className="text-orange-600" size={24} />
          </div>
          <div>
            <p className="text-xs text-gray-400 font-bold uppercase">{t('stats.avgDialysisHours')}</p>
            <p className="text-2xl font-bold text-gray-800">{stats.avgDialysisHours}h</p>
          </div>
        </div>
      </div>

      {/* 图表区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 治疗进度饼图 */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
          <h3 className="text-sm font-bold text-gray-800 mb-6 flex items-center">
            <Activity size={18} className="mr-2 text-blue-600" />
            {t('chart.treatmentProgress')}
          </h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={chartData}
                  innerRadius={60}
                  outerRadius={80}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {chartData.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value) => [t('process.count', { count: value as number }), '']}
                  contentStyle={{
                    backgroundColor: 'white',
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4">
            {processData.map((item, i) => (
              <div key={item.status} className="flex items-center text-xs">
                <div
                  className="w-3 h-3 rounded-full mr-2"
                  style={{ backgroundColor: COLORS[i] }}
                ></div>
                <span className="text-gray-500">{item.name}:</span>
                <span className="ml-auto font-bold">{t('process.count', { count: item.value })}</span>
              </div>
            ))}
          </div>
        </div>

        {/* 设备与床位运行矩阵 */}
        <div className="lg:col-span-2 bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
          <h3 className="text-sm font-bold text-gray-800 mb-6 flex items-center">
            <Monitor size={18} className="mr-2 text-blue-600" />
            {t('chart.equipmentMatrix')}
          </h3>

          {/* 图例 */}
          <div className="flex gap-4 mb-4 text-xs">
            <div className="flex items-center">
              <div className="w-3 h-3 rounded bg-blue-500 mr-1.5"></div>
              <span className="text-gray-500">{t('bedStatus.active')}</span>
            </div>
            <div className="flex items-center">
              <div className="w-3 h-3 rounded bg-gray-300 mr-1.5"></div>
              <span className="text-gray-500">{t('bedStatus.empty')}</span>
            </div>
            <div className="flex items-center">
              <div className="w-3 h-3 rounded bg-red-500 mr-1.5"></div>
              <span className="text-gray-500">{t('bedStatus.alarm')}</span>
            </div>
            <div className="flex items-center">
              <div className="w-3 h-3 rounded bg-orange-500 mr-1.5"></div>
              <span className="text-gray-500">{t('bedStatus.maintenance')}</span>
            </div>
          </div>

          {/* 床位矩阵 */}
          <div className="grid grid-cols-8 gap-3">
            {bedStatuses.map((bed) => (
              <div
                key={bed.id}
                className={`h-16 rounded-xl flex flex-col items-center justify-center border transition-all hover:scale-105 cursor-pointer ${getBedStatusStyles(bed.status)}`}
                title={`${bed.label} - ${t(`bedStatus.${bed.status}`)}`}
              >
                <span className="text-[10px] font-bold opacity-50">{bed.label}</span>
                <Monitor size={16} />
              </div>
            ))}
          </div>

          {/* 统计摘要 */}
          <div className="mt-4 pt-4 border-t border-gray-100 flex gap-6 text-sm">
            <div>
              <span className="text-gray-500">{t('summary.active')}: </span>
              <span className="font-bold text-blue-600">
                {bedStatuses.filter(b => b.status === 'active').length}
              </span>
            </div>
            <div>
              <span className="text-gray-500">{t('summary.empty')}: </span>
              <span className="font-bold text-gray-600">
                {bedStatuses.filter(b => b.status === 'empty').length}
              </span>
            </div>
            <div>
              <span className="text-gray-500">{t('summary.alarm')}: </span>
              <span className="font-bold text-red-600">
                {bedStatuses.filter(b => b.status === 'alarm').length}
              </span>
            </div>
            <div>
              <span className="text-gray-500">{t('summary.maintenance')}: </span>
              <span className="font-bold text-orange-600">
                {bedStatuses.filter(b => b.status === 'maintenance').length}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
