// PatientDetail - 患者详情页面（U3 重构：4 主 Tab + 子 Tab）

import { useState, useEffect } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { message, Tabs } from 'antd'
import {
  ArrowLeft, User, AlertTriangle, ChevronRight, ChevronLeft,
  X, PanelRight,
  ShieldAlert, AlertCircle, Clock
} from 'lucide-react'
import { restApi, convertCoreResponseToPatient } from '@/services/restClient'
import { getErrorMessage } from '@/services/restClient'
import type { Patient } from '@/types/original'
import { LoadingState } from '@/components/ui'

// 导入 Tab 组件
import {
  OverviewTab,
  HistoryTab,
} from './patient-detail/tabs'
import TreatmentTabs from './patient-detail/TreatmentTabs'
import RecordsTabs from './patient-detail/RecordsTabs'

// 导入类型
import type { MainTabID, TreatmentSubTab, RecordsSubTab } from './patient-detail/types'

export default function PatientDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { t } = useTranslation('patient')
  const [searchParams, setSearchParams] = useSearchParams()

  // URL 同步的主 Tab 和子 Tab
  const mainTab = (searchParams.get('tab') || 'overview') as MainTabID
  const treatmentSub = (searchParams.get('sub') || 'plan') as TreatmentSubTab
  const recordsSub = (searchParams.get('sub') || 'basicInfo') as RecordsSubTab

  // 核心状态
  const [focusBarOpen, setFocusBarOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [avatarLoadFailed, setAvatarLoadFailed] = useState(false)
  const [patient, setPatient] = useState<Patient | null>(null)
  const [switchingPatient, setSwitchingPatient] = useState(false)

  const setMainTab = (tab: string) => {
    const next: Record<string, string> = { tab }
    // 设置子 tab 默认值
    if (tab === 'treatment') next.sub = treatmentSub || 'plan'
    if (tab === 'records') next.sub = recordsSub || 'basicInfo'
    setSearchParams(next, { replace: true })
  }

  const setTreatmentSub = (sub: TreatmentSubTab) => {
    setSearchParams({ tab: 'treatment', sub }, { replace: true })
  }

  const setRecordsSub = (sub: RecordsSubTab) => {
    setSearchParams({ tab: 'records', sub }, { replace: true })
  }

  // Load patient data
  useEffect(() => {
    const loadPatient = async () => {
      if (!id) return
      setLoading(true)
      setAvatarLoadFailed(false)
      try {
        const coreData = await restApi.getPatientCore(id)
        const partialPatient = convertCoreResponseToPatient(coreData)

        setPatient({
          id: partialPatient.id || id,
          avatar: partialPatient.avatar,
          name: partialPatient.name || '未知患者',
          age: partialPatient.age || 0,
          gender: partialPatient.gender || '男',
          bedNumber: partialPatient.bedNumber || '',
          status: partialPatient.status || '候诊',
          patientType: partialPatient.patientType || '门诊',
          insuranceType: partialPatient.insuranceType || '自费',
          dryWeight: partialPatient.dryWeight || 65,
          defaultMode: partialPatient.defaultMode || 'HD',
          doctorName: partialPatient.doctorName || '',
          diagnosis: partialPatient.diagnosis || '慢性肾脏病5期',
          riskLevel: partialPatient.riskLevel || '低危',
          vitals: partialPatient.vitals || { bp: '120/80', hr: 75, spO2: 98, weight: 65 },
          dialysisParams: partialPatient.dialysisParams || {
            timeRemaining: '00:00', ufRate: 0, targetUf: 2.5, accumulatedUf: 0,
            bloodFlow: 250, dialysateFlow: 500, mode: 'HD',
          },
          vascularAccess: partialPatient.vascularAccess || { type: '-', site: '-', status: '未知' },
          recentLabs: partialPatient.recentLabs || [],
          orders: partialPatient.orders || [],
          infection: partialPatient.infection,
          documents: partialPatient.documents || [],
          progressNotes: partialPatient.progressNotes || [],
          medicalHistory: partialPatient.medicalHistory || {
            allergies: [], primaryDisease: '', pathology: '', tumorInfo: '', medicalHistory: '', complications: [],
          },
          outcome: partialPatient.outcome || { status: '治疗中' },
          treatmentPlan: partialPatient.treatmentPlan,
        } as Patient)
      } catch (error) {
        console.error('Load patient detail failed:', error)
        message.error(getErrorMessage(error))
        setPatient(null)
      } finally {
        setLoading(false)
      }
    }
    loadPatient()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  if (loading) {
    return <LoadingState tip={t('loadDetail')} fullscreen />
  }

  if (!patient) {
    return (
      <div data-testid="patient-error-state" className="text-center py-20">
        <p className="text-gray-500 mb-4">{t('notFoundInfo')}</p>
        <button
          onClick={() => navigate('/patients')}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          {t('action.backToList')}
        </button>
      </div>
    )
  }

  const handleBack = () => navigate('/patients')
  const handleSwitchPatient = async (direction: 'prev' | 'next') => {
    if (!id || switchingPatient) return
    setSwitchingPatient(true)
    try {
      const response = await restApi.getPatientList({ page: 1, pageSize: 1000 })
      const patients = response.data.items
      const currentIndex = patients.findIndex(item => String(item.id) === String(id))
      if (currentIndex < 0) { message.warning(t('notFoundInfo')); return }
      const nextIndex = direction === 'prev' ? currentIndex - 1 : currentIndex + 1
      const targetPatient = patients[nextIndex]
      if (!targetPatient) { message.info(direction === 'prev' ? t('label.switchPrev') : t('label.switchNext')); return }
      navigate(`/patients/${targetPatient.id}`)
    } catch (error) {
      console.error('Switch patient failed:', error)
      message.error(getErrorMessage(error))
    } finally {
      setSwitchingPatient(false)
    }
  }

  // 主 Tab 项目
  const mainTabItems = [
    {
      key: 'overview',
      label: t('tab.main.overview'),
      children: <OverviewTab patient={patient} onNavigate={(id: string) => setMainTab(id)} />,
    },
    {
      key: 'treatment',
      label: t('tab.main.treatment'),
      children: <TreatmentTabs patient={patient} defaultSub={treatmentSub} onSubChange={setTreatmentSub} />,
    },
    {
      key: 'records',
      label: t('tab.main.records'),
      children: <RecordsTabs patient={patient} defaultSub={recordsSub} onSubChange={setRecordsSub} />,
    },
    {
      key: 'history',
      label: t('tab.main.history'),
      children: <HistoryTab patient={patient} />,
    },
  ]

  return (
    <div className="h-full flex flex-col max-w-[1800px] mx-auto overflow-hidden bg-slate-50/50">
      {/* 顶部全局信息条 */}
      <div className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between shadow-sm z-20 shrink-0">
        <div className="flex items-center gap-8">
          <button onClick={handleBack} className="text-slate-400 hover:text-slate-900 transition-colors p-2.5 rounded-lg hover:bg-slate-100 active:scale-95">
            <ArrowLeft size={22} />
          </button>
          <div className="flex items-center">
            <div className="relative">
              {patient.avatar && !avatarLoadFailed ? (
                <img
                  src={patient.avatar}
                  alt={`${patient.name} avatar`}
                  className="w-16 h-16 rounded-lg object-cover bg-slate-100 shadow-xl shadow-blue-100"
                  onError={() => setAvatarLoadFailed(true)}
                />
              ) : (
                <div className="w-16 h-16 rounded-lg bg-blue-600 flex items-center justify-center text-white shadow-xl shadow-blue-100">
                  <User size={32} strokeWidth={2.5}/>
                </div>
              )}
              {patient.riskLevel === '高危' && (
                <div className="absolute -top-1 -right-1 w-6 h-6 bg-red-500 border-2 border-white rounded-full flex items-center justify-center shadow-md">
                  <AlertTriangle size={12} className="text-white"/>
                </div>
              )}
            </div>
            <div className="ml-6">
              <div className="flex items-center">
                <h1 data-testid="patient-detail-name" className="text-h2 font-bold text-slate-900">{patient.name}</h1>
                <span className={`ml-4 px-2.5 py-0.5 rounded-lg text-meta font-black border ${patient.infection?.hbsag === '阳性' ? 'bg-red-50 text-red-600 border-red-100' : 'bg-green-50 text-green-600 border-green-100'}`}>
                  {t('info.infection')}: {patient.infection?.hbsag === '阳性' ? 'HBV(+)' : t('status.safe')}
                </span>
                <span className="ml-2 px-2.5 py-0.5 bg-slate-900 text-white rounded-lg text-meta font-black shadow-sm uppercase tracking-tighter">
                  {patient.bedNumber} {t('info.bedNumber')}
                </span>
              </div>
              <div className="text-meta text-foreground-muted mt-1.5 flex items-center gap-5 font-medium">
                <span>{patient.gender} · {patient.age}岁</span>
                <span className="w-px h-3 bg-slate-200"></span>
                <span className="flex items-center text-orange-600">
                  <AlertTriangle size={12} className="mr-1"/> {t('info.allergy')}: {patient.medicalHistory.allergies[0] || t('label.noAllergyRecord')}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-6">
          <div className="text-right">
            <p className="text-meta text-foreground-muted uppercase font-medium tracking-widest">{t('info.dialysisAge')}</p>
            <p className="font-bold text-xl text-slate-900">3.2 {/* density:strict 故意小字 */}<span className="text-[10px] font-bold">Years</span></p>
          </div>
          <div className="text-right">
            <p className="text-meta text-foreground-muted uppercase font-medium tracking-widest">{t('info.targetUF')}</p>
            <p className="font-bold text-xl text-blue-600">2.50 {/* density:strict 故意小字 */}<span className="text-[10px] font-bold">L</span></p>
          </div>
          <div className="h-10 w-px bg-slate-200 mx-2"></div>
          <div className="flex gap-1.5 items-center">
            <button
              onClick={() => setFocusBarOpen(!focusBarOpen)}
              className={`p-2.5 rounded-lg transition-all relative ${focusBarOpen ? 'bg-blue-600 text-white shadow-lg' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'}`}
              title={focusBarOpen ? t('label.collapseFocus') : t('label.expandFocus')}
            >
              <PanelRight size={20}/>
              {!focusBarOpen && (
                <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full border border-white"></span>
              )}
            </button>
            <div className="flex gap-1 ml-1">
              <button disabled={switchingPatient} onClick={() => handleSwitchPatient('prev')} className="p-2.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
                <ChevronLeft size={20}/>
              </button>
              <button disabled={switchingPatient} onClick={() => handleSwitchPatient('next')} className="p-2.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
                <ChevronRight size={20}/>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* 主内容区 */}
        <div className="flex-1 overflow-y-auto no-scrollbar">
          <div className="p-4">
            <Tabs
              activeKey={mainTab}
              onChange={setMainTab}
              items={mainTabItems}
              className="patient-detail-tabs"
            />
          </div>
        </div>

        {/* 右侧焦点栏 */}
        <aside
          className={`bg-white border-l border-slate-200 flex flex-col shrink-0 z-10 shadow-2xl transition-all duration-500 ease-in-out ${focusBarOpen ? 'translate-x-0' : 'translate-x-full overflow-hidden border-none'}`}
          style={{ width: focusBarOpen ? '340px' : '0' }}
        >
          <div className="p-6 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center sticky top-0 shrink-0">
            <h3 className="text-sm font-bold text-slate-800 flex items-center uppercase tracking-wider">
              <ShieldAlert size={20} className="mr-2 text-red-500"/> {t('section.clinicalFocus')}
            </h3>
            <button onClick={() => setFocusBarOpen(false)} className="p-1.5 text-gray-400 hover:bg-slate-200 rounded-full transition-colors">
              <X size={18}/>
            </button>
          </div>
          <div className="flex-1 overflow-y-auto p-4 space-y-4 no-scrollbar">
            <div className="space-y-3">
              <p className="text-meta font-bold text-foreground-muted uppercase tracking-widest opacity-70">{t('section.criticalValues')}</p>
              <div className="p-5 bg-state-alert-bg/50 border border-red-100 rounded-lg relative overflow-hidden group">
                <div className="absolute top-0 right-0 w-16 h-16 bg-red-100/50 rounded-bl-full animate-pulse opacity-50"></div>
                <p className="text-meta text-state-alert font-bold mb-1">{t('ai.bloodK')}</p>
                <div className="flex justify-between items-baseline">
                  <span className="text-3xl font-black text-state-alert font-mono">6.2</span>
                  <span className="text-meta text-state-alert font-bold uppercase">mmol/L</span>
                </div>
                <p className="text-meta text-state-alert font-bold mt-2 flex items-center italic">
                  <AlertCircle size={10} className="mr-1"/> {t('ai.suggestion')}
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <p className="text-meta font-bold text-foreground-muted uppercase tracking-widest opacity-70">{t('section.documentStatus')}</p>
              <div className="space-y-2">
                <div className="flex items-center justify-between p-4 bg-white border border-slate-100 rounded-lg shadow-sm hover:border-blue-200 transition-colors">
                  <span className="text-xs font-bold text-slate-700">{t('nav.monthlySummary')}</span>
                  <span className="text-meta text-state-waiting font-bold uppercase bg-state-waiting-bg px-2 py-0.5 rounded">{t('status.pending')}</span>
                </div>
                <div className="flex items-center justify-between p-4 bg-white border border-slate-100 rounded-lg shadow-sm opacity-60">
                  <span className="text-xs font-bold text-slate-700">{t('ai.treatmentConsent')}</span>
                  <span className="text-meta text-state-finished font-bold uppercase bg-state-finished-bg px-2 py-0.5 rounded">{t('status.done')}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6 bg-slate-50 border-t border-slate-100 shrink-0">
            <div className="flex items-center text-foreground-muted text-meta font-medium uppercase tracking-tight">
              <Clock size={12} className="mr-2"/> {t('label.lastSync')}: 14:32:05
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}
