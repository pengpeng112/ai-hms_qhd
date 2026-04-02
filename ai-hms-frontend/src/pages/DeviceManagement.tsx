/**
 * 设备管理页面 - 工程师专用
 * 功能：设备列表、设备详情、消毒记录、维护记录、设备统计
 */
import { useState, useEffect, useMemo, memo, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  Search, RefreshCw, Settings, Wrench,
  AlertTriangle, CheckCircle, XCircle, Clock, ChevronRight,
  X, Shield, Cpu
} from 'lucide-react'
import {
  getAllEquipments,
  getEquipmentDisinfections,
  getEquipmentStats,
  getEquipmentOverview
} from '@/services/equipment'
import type { EquipmentStats, EquipmentOverview } from '@/services/equipment'
import type { EquipmentInfo, EquipmentDisinfection } from '@/services/types/api'

// ============ 类型定义 ============
interface EquipmentCardProps {
  equipment: EquipmentInfo
  onClick: () => void
  t: TFunction<'device'>
}

interface DetailModalProps {
  equipment: EquipmentInfo
  overview: EquipmentOverview | null
  disinfections: EquipmentDisinfection[]
  loading: boolean
  onClose: () => void
  t: TFunction<'device'>
}

// ============ 辅助函数 ============
/** 设备状态映射（模拟，实际应从API获取） */
const getEquipmentStatus = (equipment: EquipmentInfo): 'normal' | 'warning' | 'error' | 'offline' => {
  // 基于设备ID模拟不同状态
  const statusMap = ['normal', 'normal', 'normal', 'warning', 'normal', 'error', 'normal', 'offline'] as const
  return statusMap[equipment.Id % statusMap.length]
}

/** 状态配置 */
const statusConfig = {
  normal: { labelKey: 'status.normal', color: 'text-green-600', bg: 'bg-green-50', border: 'border-green-200', icon: CheckCircle },
  warning: { labelKey: 'status.warning', color: 'text-amber-600', bg: 'bg-amber-50', border: 'border-amber-200', icon: AlertTriangle },
  error: { labelKey: 'status.error', color: 'text-red-600', bg: 'bg-red-50', border: 'border-red-200', icon: XCircle },
  offline: { labelKey: 'status.offline', color: 'text-gray-500', bg: 'bg-gray-50', border: 'border-gray-200', icon: Clock }
}

/** 透析方式映射 */
const dialysisMethodKeys: Record<string, string> = {
  'HD': 'method.HD',
  'HDF': 'method.HDF',
  'HP': 'method.HP',
  'CRRT': 'method.CRRT',
  'PD': 'method.PD'
}

// ============ 子组件 ============

/** 设备卡片 */
const EquipmentCard = memo(function EquipmentCard({ equipment, onClick, t }: EquipmentCardProps) {
  const status = getEquipmentStatus(equipment)
  const config = statusConfig[status]
  const StatusIcon = config.icon
  const methodKey = equipment.DialysisMethod ? dialysisMethodKeys[equipment.DialysisMethod] : null
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dialysisMethod = methodKey ? t(methodKey as any) : (equipment.DialysisMethod || t('info.unknown'))

  return (
    <div
      onClick={onClick}
      className={`group bg-white rounded-xl border ${config.border} p-4 cursor-pointer
        hover:shadow-md hover:border-blue-300 transition-all duration-200`}
    >
      {/* 头部：状态和编号 */}
      <div className="flex items-center justify-between mb-3">
        <div className={`flex items-center gap-1.5 px-2 py-1 rounded-full ${config.bg}`}>
          <StatusIcon className={`w-3.5 h-3.5 ${config.color}`} />
          {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
          <span className={`text-xs font-medium ${config.color}`}>{t(config.labelKey as any)}</span>
        </div>
        <span className="text-xs text-gray-400">#{equipment.IDNo || equipment.Id}</span>
      </div>

      {/* 设备名称 */}
      <h3 className="font-semibold text-gray-800 mb-2 truncate group-hover:text-blue-600 transition-colors">
        {equipment.Name || `Device ${equipment.Id}`}
      </h3>

      {/* 设备信息 */}
      <div className="space-y-1.5 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-gray-500">{t('info.brand')}</span>
          <span className="text-gray-700 font-medium">{equipment.Brand || '-'}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-gray-500">{t('info.model')}</span>
          <span className="text-gray-700">{equipment.ModelNo || '-'}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-gray-500">{t('info.dialysisMethod')}</span>
          <span className="text-gray-700">{String(dialysisMethod)}</span>
        </div>
      </div>

      {/* 底部操作提示 */}
      <div className="mt-3 pt-3 border-t border-gray-100 flex items-center justify-between">
        <span className="text-xs text-gray-400">{t('card.viewDetail')}</span>
        <ChevronRight className="w-4 h-4 text-gray-400 group-hover:text-blue-500 group-hover:translate-x-1 transition-all" />
      </div>
    </div>
  )
})

/** 设备详情弹窗 */
const DetailModal = memo(function DetailModal({
  equipment,
  disinfections,
  loading,
  onClose,
  t
}: DetailModalProps) {
  const status = getEquipmentStatus(equipment)
  const config = statusConfig[status]
  const StatusIcon = config.icon
  const methodKey = equipment.DialysisMethod ? dialysisMethodKeys[equipment.DialysisMethod] : null
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dialysisMethod = methodKey ? t(methodKey as any) : (equipment.DialysisMethod || t('info.unknown'))

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] overflow-hidden">
        {/* 头部 */}
        <div className="bg-gradient-to-r from-blue-600 to-teal-500 text-white p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
                <Cpu className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-xl font-bold">{equipment.Name || `Device ${equipment.Id}`}</h2>
                <p className="text-blue-100 text-sm">{t('info.serialNumber')}: {equipment.SerialNo || '-'}</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-2 hover:bg-white/20 rounded-lg transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* 内容 */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-180px)]">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <RefreshCw className="w-8 h-8 text-blue-500 animate-spin" />
            </div>
          ) : (
            <div className="space-y-6">
              {/* 状态卡片 */}
              <div className={`flex items-center gap-3 p-4 rounded-xl ${config.bg} ${config.border} border`}>
                <StatusIcon className={`w-6 h-6 ${config.color}`} />
                <div>
                  {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                  <p className={`font-semibold ${config.color}`}>{t(config.labelKey as any)}</p>
                  <p className="text-sm text-gray-600">{t('status.currentState')}</p>
                </div>
              </div>

              {/* 基本信息 */}
              <div>
                <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
                  <Settings className="w-4 h-4 text-blue-500" />
                  {t('section.basicInfo')}
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <InfoItem label={t('info.deviceId')} value={equipment.IDNo || '-'} />
                  <InfoItem label={t('info.name')} value={equipment.Name || '-'} />
                  <InfoItem label={t('info.brand')} value={equipment.Brand || '-'} />
                  <InfoItem label={t('info.model')} value={equipment.ModelNo || '-'} />
                  <InfoItem label={t('info.serialNumber')} value={equipment.SerialNo || '-'} />
                  <InfoItem label={t('info.dialysisMethod')} value={String(dialysisMethod)} />
                </div>
              </div>

              {/* 消毒记录 */}
              <div>
                <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
                  <Shield className="w-4 h-4 text-green-500" />
                  {t('section.disinfection')}
                </h3>
                {disinfections.length > 0 ? (
                  <div className="space-y-2">
                    {disinfections.slice(0, 5).map((record) => (
                      <div key={record.Id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
                            <CheckCircle className="w-4 h-4 text-green-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-800">{record.DisinfectWay || t('disinfect.routine')}</p>
                            <p className="text-xs text-gray-500">{record.Description || t('disinfect.noRemark')}</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <p className="text-sm text-gray-600">
                            {record.StartTime ? new Date(record.StartTime).toLocaleDateString() : '-'}
                          </p>
                          <p className="text-xs text-gray-400">
                            {record.StartTime ? new Date(record.StartTime).toLocaleTimeString() : ''}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-gray-400">
                    <Shield className="w-12 h-12 mx-auto mb-2 opacity-50" />
                    <p>{t('disinfect.noRecord')}</p>
                  </div>
                )}
              </div>

              {/* 维护建议 */}
              <div>
                <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
                  <Wrench className="w-4 h-4 text-amber-500" />
                  {t('section.maintenance')}
                </h3>
                <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
                  <div className="flex items-start gap-3">
                    <AlertTriangle className="w-5 h-5 text-amber-500 mt-0.5" />
                    <div>
                      <p className="font-medium text-amber-800">{t('maintenance.reminder')}</p>
                      <p className="text-sm text-amber-700 mt-1">
                        {t('maintenance.suggestion')}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 底部操作 */}
        <div className="border-t border-gray-100 p-4 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
          >
            {t('action.close')}
          </button>
          <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2">
            <Wrench className="w-4 h-4" />
            {t('action.createOrder')}
          </button>
        </div>
      </div>
    </div>
  )
})

/** 信息项组件 */
function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-50 rounded-lg p-3">
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className="font-medium text-gray-800">{value}</p>
    </div>
  )
}

/** 统计卡片 */
function StatCard({
  icon: Icon,
  label,
  value,
  color
}: {
  icon: React.ElementType
  label: string
  value: string | number
  color: string
}) {
  return (
    <div className={`bg-white rounded-xl border border-gray-100 p-4 hover:shadow-md transition-shadow`}>
      <div className="flex items-center gap-3">
        <div className={`w-10 h-10 rounded-lg ${color} flex items-center justify-center`}>
          <Icon className="w-5 h-5 text-white" />
        </div>
        <div>
          <p className="text-2xl font-bold text-gray-800">{value}</p>
          <p className="text-sm text-gray-500">{label}</p>
        </div>
      </div>
    </div>
  )
}

// ============ 主组件 ============
export default function DeviceManagement() {
  const { t } = useTranslation('device')

  // 状态
  const [equipments, setEquipments] = useState<EquipmentInfo[]>([])
  const [stats, setStats] = useState<EquipmentStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [brandFilter, setBrandFilter] = useState<string>('all')

  // 详情弹窗状态
  const [selectedEquipment, setSelectedEquipment] = useState<EquipmentInfo | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [overview, setOverview] = useState<EquipmentOverview | null>(null)
  const [disinfections, setDisinfections] = useState<EquipmentDisinfection[]>([])

  // 加载设备列表
  useEffect(() => {
    async function loadData() {
      setLoading(true)
      try {
        const [equipmentList, equipmentStats] = await Promise.all([
          getAllEquipments(),
          getEquipmentStats()
        ])
        setEquipments(equipmentList)
        setStats(equipmentStats)
      } catch (error) {
        console.error('加载设备数据失败:', error)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  // 打开设备详情
  const handleOpenDetail = useCallback(async (equipment: EquipmentInfo) => {
    setSelectedEquipment(equipment)
    setDetailLoading(true)
    try {
      const [overviewData, disinfectionData] = await Promise.all([
        getEquipmentOverview(equipment.Id),
        getEquipmentDisinfections(equipment.Id)
      ])
      setOverview(overviewData)
      setDisinfections(disinfectionData.data || [])
    } catch (error) {
      console.error('加载设备详情失败:', error)
    } finally {
      setDetailLoading(false)
    }
  }, [])

  // 关闭详情弹窗
  const handleCloseDetail = useCallback(() => {
    setSelectedEquipment(null)
    setOverview(null)
    setDisinfections([])
  }, [])

  // 刷新数据
  const handleRefresh = useCallback(async () => {
    setLoading(true)
    try {
      const [equipmentList, equipmentStats] = await Promise.all([
        getAllEquipments(),
        getEquipmentStats()
      ])
      setEquipments(equipmentList)
      setStats(equipmentStats)
    } catch (error) {
      console.error('刷新数据失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  // 获取品牌列表
  const brands = useMemo(() => {
    const brandSet = new Set(equipments.map(e => e.Brand).filter(Boolean))
    return Array.from(brandSet)
  }, [equipments])

  // 过滤设备
  const filteredEquipments = useMemo(() => {
    return equipments.filter(equipment => {
      // 搜索过滤
      if (searchTerm) {
        const term = searchTerm.toLowerCase()
        const matchName = equipment.Name?.toLowerCase().includes(term)
        const matchBrand = equipment.Brand?.toLowerCase().includes(term)
        const matchModel = equipment.ModelNo?.toLowerCase().includes(term)
        const matchSerial = equipment.SerialNo?.toLowerCase().includes(term)
        if (!matchName && !matchBrand && !matchModel && !matchSerial) return false
      }
      // 状态过滤
      if (statusFilter !== 'all') {
        const status = getEquipmentStatus(equipment)
        if (status !== statusFilter) return false
      }
      // 品牌过滤
      if (brandFilter !== 'all' && equipment.Brand !== brandFilter) return false
      return true
    })
  }, [equipments, searchTerm, statusFilter, brandFilter])

  // 统计各状态数量
  const statusCounts = useMemo(() => {
    const counts = { normal: 0, warning: 0, error: 0, offline: 0 }
    equipments.forEach(e => {
      const status = getEquipmentStatus(e)
      counts[status]++
    })
    return counts
  }, [equipments])

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      {/* 页面标题 */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800">{t('title')}</h1>
        <p className="text-gray-500 mt-1">{t('subtitle')}</p>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard
          icon={Cpu}
          label={t('stat.totalDevices')}
          value={stats?.total || equipments.length}
          color="bg-blue-500"
        />
        <StatCard
          icon={CheckCircle}
          label={t('stat.normalCount')}
          value={statusCounts.normal}
          color="bg-green-500"
        />
        <StatCard
          icon={AlertTriangle}
          label={t('stat.warningCount')}
          value={statusCounts.warning}
          color="bg-amber-500"
        />
        <StatCard
          icon={XCircle}
          label={t('stat.errorCount')}
          value={statusCounts.error}
          color="bg-red-500"
        />
      </div>

      {/* 工具栏 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-4 mb-6">
        <div className="flex flex-wrap items-center gap-4">
          {/* 搜索 */}
          <div className="flex-1 min-w-[200px]">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder={t('filter.searchPlaceholder')}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* 状态过滤 */}
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">{t('filter.allStatus')}</option>
            <option value="normal">{t('status.normal')}</option>
            <option value="warning">{t('status.warning')}</option>
            <option value="error">{t('status.error')}</option>
            <option value="offline">{t('status.offline')}</option>
          </select>

          {/* 品牌过滤 */}
          <select
            value={brandFilter}
            onChange={(e) => setBrandFilter(e.target.value)}
            className="px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">{t('filter.allBrands')}</option>
            {brands.map(brand => (
              <option key={brand} value={brand}>{brand}</option>
            ))}
          </select>

          {/* 刷新按钮 */}
          <button
            onClick={handleRefresh}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            {t('action.refresh')}
          </button>
        </div>
      </div>

      {/* 设备列表 */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <RefreshCw className="w-10 h-10 text-blue-500 animate-spin" />
        </div>
      ) : filteredEquipments.length === 0 ? (
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-12 text-center">
          <Cpu className="w-16 h-16 mx-auto mb-4 text-gray-300" />
          <p className="text-gray-500 text-lg">{t('empty.noDevice')}</p>
          <p className="text-gray-400 text-sm mt-1">{t('empty.adjustFilter')}</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredEquipments.map(equipment => (
            <EquipmentCard
              key={equipment.Id}
              equipment={equipment}
              onClick={() => handleOpenDetail(equipment)}
              t={t}
            />
          ))}
        </div>
      )}

      {/* 设备详情弹窗 */}
      {selectedEquipment && (
        <DetailModal
          equipment={selectedEquipment}
          overview={overview}
          disinfections={disinfections}
          loading={detailLoading}
          onClose={handleCloseDetail}
          t={t}
        />
      )}
    </div>
  )
}
