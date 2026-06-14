import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { restApi, convertRestPatientList } from '@/services'
import type { Patient } from '@/types/original'
import {
  Search, Plus, Filter, ArrowRight, User as UserIcon,
  BedDouble, Stethoscope, Activity, RefreshCw, AlertCircle, Copy, Trash2
} from 'lucide-react'
import { LoadingState } from '@/components'
import { CreatePatientModal } from '@/components/patient'
import { message, Modal, Dropdown, Badge } from 'antd'
import { useDictNameMaps, getNameFromMap } from '@/hooks/useDictName'
import { DICT_TYPES } from '@/services/dictApi'
import { useAuth } from '@/contexts/AuthContext'
import { getErrorMessage } from '@/services/restClient'

type FilterType = 'all' | 'today' | 'active' | 'mine' | 'in_dept' | 'transferred'

const FILTER_ITEMS: Array<{ key: FilterType; label: string; shortLabel: string; icon?: React.ReactNode }> = [
  { key: 'all', label: '全部患者', shortLabel: '全部' },
  { key: 'in_dept', label: '在科患者', shortLabel: '在科' },
  { key: 'today', label: '今日治疗', shortLabel: '今日' },
  { key: 'active', label: '透析中', shortLabel: '透析中' },
  { key: 'mine', label: '我的患者', shortLabel: '我的' },
  { key: 'transferred', label: '转出患者', shortLabel: '转出' },
]

const STAT_CARDS = [
  { key: 'totalCount', label: '总人数', desc: '全部建档患者', color: 'bg-blue-500' },
  { key: 'activeCount', label: '在科活跃', desc: '当前在院/在科患者', color: 'bg-emerald-500' },
  { key: 'outpatientCount', label: '门诊', desc: '门诊透析患者', color: 'bg-teal-500' },
  { key: 'inpatientCount', label: '住院', desc: '住院患者', color: 'bg-indigo-500' },
] as const

export default function PatientList() {
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const [searchTerm, setSearchTerm] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [activeFilter, setActiveFilter] = useState<FilterType>('in_dept')
  const [loading, setLoading] = useState(false)
  const [patients, setPatients] = useState<Partial<Patient>[]>([])
  const [apiError, setApiError] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [patientStats, setPatientStats] = useState<{totalCount:number; activeCount:number; outpatientCount:number; inpatientCount:number} | null>(null)
  const pageSize = 50

  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  useEffect(() => {
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(searchTerm)
      setCurrentPage(1)
    }, 300)
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current)
    }
  }, [searchTerm])

  const dictTypeCodes = useMemo(() => [
    DICT_TYPES.INSURANCE_TYPE,
    DICT_TYPES.PATIENT_TYPE,
    DICT_TYPES.DIALYSIS_MODE,
  ], [])
  const dictNameMaps = useDictNameMaps(dictTypeCodes)

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
  }, [loadPatients, refreshKey])

  useEffect(() => {
    restApi.getPatientStats().then(setPatientStats).catch(() => {})
  }, [refreshKey])

  const handleCloseModal = useCallback(() => {
    setCreateModalOpen(false)
  }, [])

  const handleCreateSuccess = useCallback(() => {
    setCreateModalOpen(false)
    setCurrentPage(1)
    setRefreshKey(k => k + 1)
  }, [])

  const handleFilterChange = useCallback((filter: FilterType) => {
    setActiveFilter(filter)
    setCurrentPage(1)
  }, [])

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
        return true
      default:
        return true
    }
  })

  const handleSelectPatient = (id: string) => {
    navigate(`/patients/${id}`)
  }

  const handleDeletePatient = async (patient: Partial<Patient>, e: React.MouseEvent) => {
    e.stopPropagation()

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
          setRefreshKey(k => k + 1)
        } catch (error) {
          console.error('删除患者失败:', error)
          message.error(getErrorMessage(error))
        }
      },
    })
  }

  const filterCounts: Record<string, number | undefined> = {
    all: patientStats?.totalCount,
    in_dept: patientStats?.activeCount,
  }

  return (
    <div className="h-full flex flex-col max-w-[1600px] mx-auto space-y-4">
      <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p className="text-xs font-black uppercase tracking-[0.28em] text-blue-600">Patient Management</p>
            <h1 className="mt-1 text-2xl font-black tracking-tight text-slate-900">全科患者管理</h1>
            <p className="mt-1 text-sm text-slate-500">统一维护患者档案、透析方案、医保类型、当前状态与床位信息。</p>
            {apiError && (
              <p className="mt-2 inline-flex items-center gap-1 text-xs text-orange-600">
                <AlertCircle size={12} />
                {apiError}
              </p>
            )}
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={loadPatients}
              disabled={loading}
              className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-500 transition hover:bg-slate-50 disabled:opacity-50"
            >
              <RefreshCw size={16} className={loading ? 'animate-spin' : ''} />
            </button>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={15} />
              <input
                type="text"
                placeholder="搜索姓名、ID、床号..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="h-9 w-56 rounded-lg border border-slate-200 bg-slate-50 pl-9 pr-4 text-sm outline-none transition focus:border-blue-400 focus:bg-white"
              />
            </div>
            <button
              onClick={() => setCreateModalOpen(true)}
              className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-bold text-white transition hover:bg-blue-700 shadow-sm"
            >
              <Plus size={16} /> 新增建档
            </button>
          </div>
        </div>
      </section>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        {STAT_CARDS.map((card) => (
          <div key={card.key} className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
            <div className={`h-1 w-full ${card.color}`} />
            <div className="p-4">
              <div className="text-2xl font-black text-slate-900">
                {patientStats?.[card.key] ?? '--'}
              </div>
              <div className="mt-0.5 text-sm font-bold text-slate-500">{card.label}</div>
              <div className="mt-0.5 text-[11px] text-slate-400">{card.desc}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="flex flex-1 flex-col overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-1 overflow-x-auto border-b border-slate-100 px-3 py-2">
          {FILTER_ITEMS.map((item) => {
            const active = activeFilter === item.key
            const count = filterCounts[item.key]
            return (
              <button
                key={item.key}
                onClick={() => handleFilterChange(item.key)}
                className={`flex shrink-0 items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-semibold whitespace-nowrap transition ${
                  active
                    ? 'bg-blue-600 text-white shadow-sm'
                    : 'text-slate-500 hover:bg-slate-50 hover:text-slate-700'
                }`}
              >
                {item.label}
                {count !== undefined && (
                  <span className={`rounded-full px-1.5 py-0.5 text-[10px] font-bold ${active ? 'bg-blue-500 text-blue-100' : 'bg-slate-100 text-slate-400'}`}>
                    {count}
                  </span>
                )}
              </button>
            )
          })}
          <div className="flex-1" />
          <Badge count={0} size="small">
            <Dropdown
              menu={{
                items: [
                  { key: 'today', label: '今日治疗', icon: <BedDouble size={14} /> },
                  { key: 'active', label: '透析中', icon: <Activity size={14} /> },
                  { key: 'mine', label: '我的患者', icon: <Stethoscope size={14} /> },
                ],
                onClick: ({ key }) => handleFilterChange(key as FilterType),
                selectedKeys: [],
              }}
              trigger={['click']}
            >
              <button className="flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold text-slate-400 transition hover:bg-slate-50 hover:text-slate-600">
                <Filter size={13} />
                更多筛选
              </button>
            </Dropdown>
          </Badge>
        </div>

        <div className="flex-1 overflow-auto">
          {loading ? (
            <LoadingState tip="加载中..." />
          ) : (
            <table data-testid="patient-list-table" className="w-full text-left text-sm">
              <thead className="sticky top-0 z-10 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-5 py-3">基本信息</th>
                  <th className="px-5 py-3">类型 / 医保</th>
                  <th className="px-5 py-3">治疗方案</th>
                  <th className="px-5 py-3">当前状态</th>
                  <th className="px-5 py-3">主治医生</th>
                  <th className="px-5 py-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {filteredPatients.map((patient) => (
                  <tr key={patient.id || ''} tabIndex={0} role="row"
                    onClick={() => patient.id && handleSelectPatient(patient.id)}
                    onKeyDown={(e) => { if (e.key === 'Enter' && patient.id) handleSelectPatient(patient.id) }}
                    className="group cursor-pointer border-l-[3px] border-l-transparent transition-colors hover:border-l-blue-500 hover:bg-blue-50/30 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
                  >
                    <td className="px-5 py-3.5">
                      <div className="flex items-center gap-3">
                        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-slate-100 text-slate-400 overflow-hidden">
                          {patient.avatar ? (
                            <img src={patient.avatar} alt={patient.name} className="h-full w-full object-cover" />
                          ) : (
                            <UserIcon size={16} />
                          )}
                        </div>
                        <div className="min-w-0">
                          <div className="text-sm font-bold text-slate-900 truncate">{patient.name}</div>
                          <div className="mt-0.5 text-[11px] text-slate-400">
                            {patient.gender} · {patient.age}岁 · {patient.id}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="space-y-1">
                        <span className={`inline-block rounded-md px-2 py-0.5 text-[11px] font-medium border ${
                          patient.patientType === '住院' ? 'bg-blue-50 text-blue-600 border-blue-100' : 'bg-green-50 text-green-600 border-green-100'
                        }`}>
                          {getNameFromMap(dictNameMaps[DICT_TYPES.PATIENT_TYPE] || new Map(), patient.patientType)}
                        </span>
                        <div className="text-[11px] text-slate-500">{getNameFromMap(dictNameMaps[DICT_TYPES.INSURANCE_TYPE] || new Map(), patient.insuranceType)}</div>
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-bold text-slate-800">
                          {getNameFromMap(dictNameMaps[DICT_TYPES.DIALYSIS_MODE] || new Map(), patient.defaultMode)}
                        </span>
                        <span className="text-slate-300">|</span>
                        <span className="text-sm font-semibold text-slate-600">
                          {Math.round(patient.dryWeight ?? 0)}
                          <span className="text-[10px] font-normal text-slate-400"> kg</span>
                        </span>
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="flex items-center gap-1.5">
                        <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-medium border ${
                          patient.status === '透析中' ? 'bg-blue-100 text-blue-700 border-blue-200' :
                          patient.status === '候诊' ? 'bg-yellow-100 text-yellow-700 border-yellow-200' :
                          'bg-slate-100 text-slate-600 border-slate-200'
                        }`}>
                          {patient.status}
                        </span>
                        {patient.bedNumber && (
                          <span className="rounded-md bg-slate-100 px-1.5 py-0.5 text-[11px] font-mono font-bold text-slate-600">
                            {patient.bedNumber}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <span className="text-sm text-slate-500 font-medium">
                        {patient.doctorName || '--'}
                      </span>
                    </td>
                    <td className="px-5 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={(e) => { e.stopPropagation(); navigator.clipboard.writeText(patient.id || ''); message.success('已复制') }}
                          className="rounded-lg p-2 text-slate-400 transition hover:bg-slate-100 hover:text-blue-600"
                          title="复制ID"
                        >
                          <Copy size={14} />
                        </button>
                        <button
                          onClick={(e) => handleDeletePatient(patient, e)}
                          className="rounded-lg p-2 text-slate-400 transition hover:bg-red-50 hover:text-red-600"
                          title="删除"
                        >
                          <Trash2 size={14} />
                        </button>
                        <button
                          className="rounded-lg p-2 text-slate-400 transition hover:bg-blue-50 hover:text-blue-600"
                          title="进入详情"
                        >
                          <ArrowRight size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {filteredPatients.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-16 text-center text-slate-400">
                      <div data-testid="patient-empty-state" className="flex flex-col items-center gap-2">
                        <div className="rounded-full bg-slate-100 p-4">
                          <Filter size={24} className="text-slate-300" />
                        </div>
                        <p className="text-sm font-semibold text-slate-500">未找到符合条件的患者</p>
                        <p className="text-xs text-slate-400">请调整搜索关键词或清空筛选条件</p>
                        {debouncedSearch && (
                          <button
                            type="button"
                            onClick={() => { setSearchTerm(''); setDebouncedSearch('') }}
                            className="mt-2 rounded-lg border border-slate-200 px-3 py-1 text-xs text-slate-500 transition hover:bg-slate-50"
                          >
                            清空搜索
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </div>

        {!loading && (
          <div className="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-xs text-slate-500">
            <span>
              显示 {filteredPatients.length} 条，共 {total || patients.length} 条
              <span className="ml-2 text-slate-300">|</span>
              <span className="ml-2">50 条/页</span>
            </span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="rounded-lg border border-slate-200 px-3 py-1.5 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                上一页
              </button>
              <span className="rounded-lg bg-slate-100 px-3 py-1.5 font-semibold text-slate-700">第 {currentPage} 页</span>
              <button
                onClick={() => setCurrentPage(p => p + 1)}
                disabled={currentPage * pageSize >= (total || patients.length)}
                className="rounded-lg border border-slate-200 px-3 py-1.5 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                下一页
              </button>
            </div>
          </div>
        )}
      </div>

      <CreatePatientModal
        open={createModalOpen}
        onClose={handleCloseModal}
        onSuccess={handleCreateSuccess}
      />
    </div>
  )
}
