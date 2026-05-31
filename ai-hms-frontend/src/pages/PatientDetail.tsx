// PatientDetail - 患者详情页面（U3 重构：4 主 Tab + 子 Tab + Header + FocusPanel）

import { useState, useEffect } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { message, Tabs } from 'antd'
import { ArrowLeft, ChevronLeft, ChevronRight } from 'lucide-react'
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
import PatientHeader from './patient-detail/PatientHeader'
import FocusPanel from './patient-detail/FocusPanel'

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
      {/* 顶部信息条 */}
      <div className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between shadow-sm z-20 shrink-0">
        <div className="flex items-center gap-4">
          <button onClick={handleBack} className="text-slate-400 hover:text-slate-900 transition-colors p-2.5 rounded-lg hover:bg-slate-100 active:scale-95 shrink-0">
            <ArrowLeft size={22} />
          </button>
          <PatientHeader patient={patient} avatarFailed={avatarLoadFailed} onAvatarError={() => setAvatarLoadFailed(true)} />
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <div className="flex gap-1">
            <button disabled={switchingPatient} onClick={() => handleSwitchPatient('prev')} className="p-2.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
              <ChevronLeft size={20}/>
            </button>
            <button disabled={switchingPatient} onClick={() => handleSwitchPatient('next')} className="p-2.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
              <ChevronRight size={20}/>
            </button>
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
        <FocusPanel patient={patient} />
      </div>
    </div>
  )
}
