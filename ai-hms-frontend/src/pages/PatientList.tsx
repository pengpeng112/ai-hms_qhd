import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { restApi, convertRestPatientList } from '@/services'
import type { Patient } from '@/types/original'
import {
  Search, Plus, Filter, ArrowRight, User as UserIcon,
  BedDouble, Stethoscope, Activity, RefreshCw, AlertCircle, MoreHorizontal, Copy, Trash2
} from 'lucide-react'
import { LoadingState } from '@/components'
import { CreatePatientModal } from '@/components/patient'
import { message, Modal, Dropdown, Badge } from 'antd'
import { useDictNameMaps, getNameFromMap } from '@/hooks/useDictName'
import { DICT_TYPES } from '@/services/dictApi'
import { useAuth } from '@/contexts/AuthContext'
import { getErrorMessage } from '@/services/restClient'

type FilterType = 'all' | 'today' | 'active' | 'mine' | 'in_dept' | 'transferred'

export default function PatientList() {
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const [searchTerm, setSearchTerm] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [activeFilter, setActiveFilter] = useState<FilterType>('all')
  const [loading, setLoading] = useState(false)
  const [patients, setPatients] = useState<Partial<Patient>[]>([])
  const [apiError, setApiError] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)  // 用于强制刷新列表
  const [patientStats, setPatientStats] = useState<{totalCount:number; activeCount:number; outpatientCount:number; inpatientCount:number} | null>(null)
  const pageSize = 50

  // 搜索防抖：输入停止 300ms 后触发后端搜索
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  useEffect(() => {
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(searchTerm)
      setCurrentPage(1) // 搜索变化时重置到第一页
    }, 300)
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current)
    }
  }, [searchTerm])

  // 加载字典名称映射
  const dictTypeCodes = useMemo(() => [
    DICT_TYPES.INSURANCE_TYPE,
    DICT_TYPES.PATIENT_TYPE,
    DICT_TYPES.DIALYSIS_MODE,
  ], [])
  const dictNameMaps = useDictNameMaps(dictTypeCodes)

  // 构建后端查询参数：将可映射的筛选条件传给后端
  const buildQueryParams = useCallback(() => {
    const params: { page: number; pageSize: number; name?: string; status?: string; onlyActive?: boolean; onlyTransferred?: boolean } = {
      page: currentPage,
      pageSize,
    }
    if (debouncedSearch.trim()) {
      params.name = debouncedSearch.trim()
    }
    if (activeFilter === 'active') {
      params.status = '透析中'
    }
    if (activeFilter === 'in_dept') {
      params.onlyActive = true
    }
    if (activeFilter === 'transferred') {
      params.onlyTransferred = true
    }
    return params
  }, [currentPage, debouncedSearch, activeFilter])

  // Load patient data
  const loadPatients = useCallback(async () => {
    setLoading(true)
    setApiError(null)
    try {
      const params = buildQueryParams()
      const result = await restApi.getPatientList(params)
      const patients = convertRestPatientList(result.data.items)
      setPatients(patients)
      setTotal(result.data.pagination.total)
    } catch (error) {
      console.error('加载患者数据失败:', error)
      setApiError(getErrorMessage(error))
      setPatients([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [buildQueryParams])

  useEffect(() => {
    loadPatients()
  }, [loadPatients, refreshKey])  // refreshKey 变化时强制刷新

  // 加载患者统计
  useEffect(() => {
    restApi.getPatientStats().then(setPatientStats).catch(() => {})
  }, [refreshKey])

  // 稳定的 Modal 回调（符合 Vercel React 最佳实践：rerender-functional-setstate）
  const handleCloseModal = useCallback(() => {
    setCreateModalOpen(false)
  }, [])

  const handleCreateSuccess = useCallback(() => {
    setCreateModalOpen(false)
    // 重置到第一页并强制刷新列表
    setCurrentPage(1)
    setRefreshKey(k => k + 1)  // 递增 refreshKey 触发 useEffect 重新执行
  }, [])

  // 筛选切换时重置分页
  const handleFilterChange = useCallback((filter: FilterType) => {
    setActiveFilter(filter)
    setCurrentPage(1)
  }, [])

  // 本地过滤：仅处理后端无法表达的筛选条件（today/mine/transferred）
  const filteredPatients = patients.filter(p => {
    if (!p.name || !p.id) return false

    switch (activeFilter) {
      case 'today':
        return p.status !== '居家'
      case 'mine':
        if (currentUser?.role?.includes('DOCTOR')) {
          return p.doctorName === currentUser?.name
        } else if (currentUser?.role?.includes('NURSE')) {
          return p.status !== '居家'
        }
        return true
      case 'transferred':
        // 转出患者需要后端支持，当前显示全部但标记
        return true
      default:
        return true
    }
  })

  const handleSelectPatient = (id: string) => {
    navigate(`/patients/${id}`)
  }

  // 删除患者
  const handleDeletePatient = async (patient: Partial<Patient>, e: React.MouseEvent) => {
    e.stopPropagation() // 阻止事件冒泡，避免触发导航

    if (!patient.id) return

    const patientId = patient.id
    const patientName = patient.name

    Modal.confirm({
      title: '确认删除',
      content: `确定要删除患者"${patientName}"吗？此操作将删除患者及其所有相关数据，且无法恢复。`,
      okText: '确认',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await restApi.deletePatient(patientId)
          message.success('删除成功')
          loadPatients()
        } catch (error) {
          console.error('删除患者失败:', error)
          message.error(getErrorMessage(error))
        }
      },
    })
  }

  return (
    <div className="h-full flex flex-col max-w-[1600px] mx-auto">
      {/* Header */}
      <div className="mb-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 className="text-h2 font-bold text-foreground">全科患者管理</h2>
          {apiError && (
            <p className="text-orange-500 text-meta mt-1 inline-flex items-center">
              <AlertCircle size={12} className="mr-1" />
              {apiError}
            </p>
          )}
        </div>
        <div className="flex items-center space-x-3">
          <button
            onClick={loadPatients}
            disabled={loading}
            className="p-2 text-gray-500 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw size={18} className={loading ? 'animate-spin' : ''} />
          </button>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={16} />
            <input
              type="text"
              placeholder="搜索姓名、ID、床号..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-9 pr-4 py-2 rounded-lg border border-gray-200 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none w-64"
            />
          </div>
          <button
            onClick={() => setCreateModalOpen(true)}
            className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 shadow-sm transition-colors"
          >
            <Plus size={16} className="mr-2" /> 新增建档
          </button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
        <div className="bg-surface-sunken rounded-md p-3">
          <div className="text-h2 font-semibold text-foreground">{patientStats?.totalCount ?? '--'}</div>
          <div className="text-meta text-foreground-muted">总人数</div>
        </div>
        <div className="bg-surface-sunken rounded-md p-3">
          <div className="text-h2 font-semibold text-state-treating">{patientStats?.activeCount ?? '--'}</div>
          <div className="text-meta text-foreground-muted">在科活跃</div>
        </div>
        <div className="bg-surface-sunken rounded-md p-3">
          <div className="text-h2 font-semibold text-state-finished">{patientStats?.outpatientCount ?? '--'}</div>
          <div className="text-meta text-foreground-muted">门诊</div>
        </div>
        <div className="bg-surface-sunken rounded-md p-3">
          <div className="text-h2 font-semibold text-purple-600">{patientStats?.inpatientCount ?? '--'}</div>
          <div className="text-meta text-foreground-muted">住院</div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 flex flex-col flex-1 overflow-hidden">
        {/* Filter Tabs */}
        <div className="p-4 border-b border-gray-100 flex items-center space-x-2 overflow-x-auto">
          <button
            onClick={() => handleFilterChange('all')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors flex items-center whitespace-nowrap ${activeFilter === 'all' ? 'bg-gray-100 text-gray-900' : 'text-gray-500 hover:bg-gray-50'}`}
          >
            全部患者
          </button>
          <button
            onClick={() => handleFilterChange('in_dept')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors flex items-center whitespace-nowrap ${activeFilter === 'in_dept' ? 'bg-emerald-50 text-emerald-700' : 'text-gray-500 hover:bg-gray-50'}`}
          >
            在科患者
          </button>
          <button
            onClick={() => handleFilterChange('transferred')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors flex items-center whitespace-nowrap ${activeFilter === 'transferred' ? 'bg-rose-50 text-rose-700' : 'text-gray-500 hover:bg-gray-50'}`}
          >
            转出患者
          </button>

          <div className="flex-1"></div>
          <Badge count={activeFilter !== 'all' && activeFilter !== 'in_dept' && activeFilter !== 'transferred' ? 1 : 0} size="small">
            <Dropdown
              menu={{
                items: [
                  { key: 'today', label: '今日治疗', icon: <BedDouble size={14} /> },
                  { key: 'active', label: '透析中', icon: <Activity size={14} /> },
                  { key: 'mine', label: '我的患者', icon: <Stethoscope size={14} /> },
                ],
                onClick: ({ key }) => handleFilterChange(key as FilterType),
                selectedKeys: [activeFilter],
              }}
              trigger={['click']}
            >
              <button className="text-gray-400 hover:text-gray-600 p-2 rounded-lg hover:bg-gray-50 transition-colors">
                <Filter size={16} />
              </button>
            </Dropdown>
          </Badge>
        </div>

        {/* Table Content */}
        <div className="flex-1 overflow-auto">
          {loading ? (
            <LoadingState tip="加载中..." />
          ) : (
            <table data-testid="patient-list-table" className="w-full text-left text-sm">
              <thead className="bg-gray-50 text-gray-500 font-medium sticky top-0 z-10">
                <tr>
                  <th className="px-6 py-3">基本信息</th>
                  <th className="px-6 py-3">类型 / 医保</th>
                  <th className="px-6 py-3">治疗方案 (模式/干体重)</th>
                  <th className="px-6 py-3">当前状态</th>
                  <th className="px-6 py-3">主治医生</th>
                  <th className="px-6 py-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filteredPatients.map((patient) => (
                  <tr key={patient.id || ''} onClick={() => patient.id && handleSelectPatient(patient.id)} className="border-l-4 border-transparent hover:border-state-treating cursor-pointer group transition-colors">
                    <td className="px-6 py-5">
                      <div className="flex items-center">
                        <div className={`w-10 h-10 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center mr-3 border border-slate-200 shrink-0`}>
                          <UserIcon size={18} />
                        </div>
                        <div>
                          <div className="font-bold text-gray-800">{patient.name}</div>
                          <div className="text-xs text-gray-500">{patient.gender} · {patient.age}岁 · {patient.id}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="space-y-1">
                        <span className={`inline-block px-2 py-0.5 rounded text-meta border ${patient.patientType === '住院' ? 'bg-blue-50 text-blue-600 border-blue-100' : 'bg-green-50 text-green-600 border-green-100'
                          }`}>
                          {getNameFromMap(dictNameMaps[DICT_TYPES.PATIENT_TYPE] || new Map(), patient.patientType)}
                        </span>
                        <div className="text-xs text-gray-600">{getNameFromMap(dictNameMaps[DICT_TYPES.INSURANCE_TYPE] || new Map(), patient.insuranceType)}</div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="flex items-center space-x-3">
                        <div>
                          <div className="font-bold text-gray-800">{getNameFromMap(dictNameMaps[DICT_TYPES.DIALYSIS_MODE] || new Map(), patient.defaultMode)}</div>
                        </div>
                        <div className="h-6 w-px bg-gray-200"></div>
                        <div>
                          {/* eslint-disable no-restricted-syntax -- density:strict 故意小字（单位后缀） */}
                          <div className="font-bold text-gray-800">{Math.round(patient.dryWeight ?? 0)} <span className="text-[10px] font-normal text-gray-400">kg</span></div>
                          {/* eslint-enable no-restricted-syntax */}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="flex items-center space-x-2">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${
                          patient.status === '透析中' ? 'bg-blue-100 text-blue-700 border-blue-200' :
                          patient.status === '候诊' ? 'bg-yellow-100 text-yellow-700 border-yellow-200' :
                          'bg-gray-100 text-gray-600 border-gray-200'
                        }`}>
                          {patient.status}
                        </span>
                        {patient.bedNumber && (
                          <span className="font-mono font-bold text-gray-600 bg-gray-100 px-2 py-0.5 rounded text-xs border border-gray-200">
                            {patient.bedNumber}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <span className="text-gray-700 font-medium">{patient.doctorName}</span>
                    </td>
                    <td className="px-6 py-5 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Dropdown
                          menu={{
                            items: [
                              { key: 'copyId', label: '复制 ID', icon: <Copy size={14} /> },
                              { type: 'divider' as const },
                              { key: 'delete', label: '删除', icon: <Trash2 size={14} />, danger: true },
                            ],
                            onClick: ({ key, domEvent }) => {
                              domEvent.stopPropagation()
                              if (key === 'delete') handleDeletePatient(patient, domEvent as unknown as React.MouseEvent)
                              if (key === 'copyId') { navigator.clipboard.writeText(patient.id || ''); message.success('已复制') }
                            },
                          }}
                          trigger={['click']}
                        >
                          <button
                            onClick={(e) => e.stopPropagation()}
                            className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-full transition-colors"
                          >
                            <MoreHorizontal size={16} />
                          </button>
                        </Dropdown>
                        <button className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-full transition-colors">
                          <ArrowRight size={18} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {filteredPatients.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-12 text-center text-gray-400">
                      <div data-testid="patient-empty-state" className="flex flex-col items-center">
                        <div className="bg-gray-50 p-4 rounded-full mb-3">
                          <Filter size={24} className="text-gray-300" />
                        </div>
                        <p>未找到符合条件的患者</p>
                      </div>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {!loading && filteredPatients.length > 0 && (
          <div className="p-4 border-t border-gray-100 flex items-center justify-between text-sm text-gray-500">
            <span>
              显示 {filteredPatients.length} 条，共 {total || patients.length} 条
            </span>
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="px-3 py-1 rounded border border-gray-200 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              <span className="px-3 py-1">第 {currentPage} 页</span>
              <button
                onClick={() => setCurrentPage(p => p + 1)}
                disabled={patients.length < pageSize}
                className="px-3 py-1 rounded border border-gray-200 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                下一页
              </button>
            </div>
          </div>
        )}
      </div>

      {/* 新增患者弹窗 */}
      <CreatePatientModal
        open={createModalOpen}
        onClose={handleCloseModal}
        onSuccess={handleCreateSuccess}
      />
    </div>
  )
}
