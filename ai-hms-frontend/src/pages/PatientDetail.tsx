// PatientDetail - 患者详情页面（重构后）
// 所有 Tab 内容已拆分为独立组件

import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { message } from 'antd'
import {
  ArrowLeft, Activity, User, UserCircle, AlertTriangle, ChevronRight, ChevronLeft,
  Layers, FileHeart, Beaker, Syringe, History, Files, X, PanelRight,
  ShieldAlert, AlertCircle, Clock, ClipboardList
} from 'lucide-react'
import { restApi, convertCoreResponseToPatient } from '@/services/restClient'
import type { Patient } from '@/types/original'
import { LoadingState } from '@/components/ui'

// 导入 Tab 组件
import {
  OverviewTab,
  BasicInfoTab,
  TreatmentPlanTab,
  MedicalRecordTab,
  SchemeOrderTab,
  LabsExamsTab,
  VascularTab,
  HistoryTab,
  MonthlySummaryTab
} from './patient-detail/tabs'

// 导入类型
import type { TabID } from './patient-detail/types'

export default function PatientDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { t } = useTranslation('patient')

  // 核心状态
  const [activeTab, setActiveTab] = useState<TabID>('overview')
  const [focusBarOpen, setFocusBarOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [patient, setPatient] = useState<Patient | null>(null)

  // Load patient data
  useEffect(() => {
    const loadPatient = async () => {
      if (!id) return
      setLoading(true)
      try {
        // 使用 /core 接口获取患者核心数据
        const coreData = await restApi.getPatientCore(id)
        const partialPatient = convertCoreResponseToPatient(coreData)

        // 确保数据符合 Patient 类型（添加必要的默认值）
        setPatient({
          id: partialPatient.id || id,
          name: partialPatient.name || '未知患者',
          age: partialPatient.age || 0,
          gender: partialPatient.gender || '男',
          bedNumber: partialPatient.bedNumber || '',
          status: partialPatient.status || '候诊',
          patientType: partialPatient.patientType || '门诊',
          insuranceType: partialPatient.insuranceType || '自费',
          dryWeight: partialPatient.dryWeight || 65,
          defaultMode: partialPatient.defaultMode || 'HD',
          doctorName: partialPatient.doctorName || '王医生',
          diagnosis: partialPatient.diagnosis || '慢性肾脏病5期',
          riskLevel: partialPatient.riskLevel || '低危',
          vitals: partialPatient.vitals || {
            bp: '120/80',
            hr: 75,
            spO2: 98,
            weight: 65,
          },
          dialysisParams: partialPatient.dialysisParams || {
            timeRemaining: '00:00',
            ufRate: 0,
            targetUf: 2.5,
            accumulatedUf: 0,
            bloodFlow: 250,
            dialysateFlow: 500,
            mode: 'HD',
          },
          vascularAccess: partialPatient.vascularAccess || {
            type: '-',
            site: '-',
            status: '未知',
          },
          recentLabs: partialPatient.recentLabs || [],
          orders: partialPatient.orders || [],
          infection: partialPatient.infection,  // 允许 undefined，由 OverviewTab 显示空状态
          documents: partialPatient.documents || [],
          progressNotes: partialPatient.progressNotes || [],
          medicalHistory: partialPatient.medicalHistory || {
            allergies: [],
            primaryDisease: '',
            pathology: '',
            tumorInfo: '',
            medicalHistory: '',
            complications: [],
          },
          outcome: partialPatient.outcome || {
            status: '治疗中',
          },
          treatmentPlan: partialPatient.treatmentPlan,
        } as Patient)
      } catch (error) {
        console.error('Load patient detail failed:', error)
        message.error(t('apiLoadFailed'))
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

  // Navigation handlers
  const handleBack = () => navigate('/patients')
  const handleSwitchPatient = (direction: 'prev' | 'next') => {
    message.info(direction === 'prev' ? t('label.switchPrev') : t('label.switchNext'))
  }

  // Tab 导航配置
  const navItems = [
    { id: 'overview' as const, label: t('nav.overview'), icon: Activity },
    { id: 'basic_info' as const, label: t('nav.basicInfo'), icon: UserCircle },
    { id: 'treatment_plan' as const, label: t('nav.treatmentPlan'), icon: Layers },
    { id: 'medical_record' as const, label: t('nav.medicalRecord'), icon: FileHeart },
    { id: 'scheme_order' as const, label: t('nav.schemeOrder'), icon: ClipboardList },
    { id: 'labs_exams' as const, label: t('nav.labsExams'), icon: Beaker },
    { id: 'vascular' as const, label: t('nav.vascular'), icon: Syringe },
    { id: 'history' as const, label: t('nav.history'), icon: History },
    { id: 'monthly_summary' as const, label: t('nav.monthlySummary'), icon: Files },
  ]

  // 渲染当前 Tab 内容
  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return <OverviewTab patient={patient} onNavigate={setActiveTab} />
      case 'basic_info':
        return <BasicInfoTab patient={patient} />
      case 'treatment_plan':
        return <TreatmentPlanTab patientId={patient.id} patientName={patient.name} />
      case 'medical_record':
        return <MedicalRecordTab patient={patient} />
      case 'scheme_order':
        return <SchemeOrderTab patient={patient} />
      case 'labs_exams':
        return <LabsExamsTab patient={patient} />
      case 'vascular':
        return <VascularTab patient={patient} />
      case 'history':
        return <HistoryTab patient={patient} />
      case 'monthly_summary':
        return <MonthlySummaryTab patient={patient} />
      default:
        return null
    }
  }

  return (
    <div className="h-full flex flex-col max-w-[1800px] mx-auto overflow-hidden bg-slate-50/50">
      {/* 顶部全局信息条 */}
      <div className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between shadow-sm z-20 shrink-0">
        <div className="flex items-center gap-8">
          <button onClick={handleBack} className="text-slate-400 hover:text-slate-900 transition-colors p-2.5 rounded-2xl hover:bg-slate-100 active:scale-95">
            <ArrowLeft size={22} />
          </button>
          <div className="flex items-center">
            <div className="relative">
              <div className="w-16 h-16 rounded-3xl bg-blue-600 flex items-center justify-center text-white shadow-xl shadow-blue-100">
                <User size={32} strokeWidth={2.5}/>
              </div>
              {patient.riskLevel === '高危' && (
                <div className="absolute -top-1 -right-1 w-6 h-6 bg-red-500 border-2 border-white rounded-full flex items-center justify-center animate-bounce shadow-md">
                  <AlertTriangle size={12} className="text-white"/>
                </div>
              )}
            </div>
            <div className="ml-6">
              <div className="flex items-center">
                <h1 data-testid="patient-detail-name" className="text-2xl font-black text-slate-900">{patient.name}</h1>
                <span className={`ml-4 px-2.5 py-0.5 rounded-lg text-[10px] font-black border ${patient.infection?.hbsag === '阳性' ? 'bg-red-50 text-red-600 border-red-100' : 'bg-green-50 text-green-600 border-green-100'}`}>
                  {t('info.infection')}: {patient.infection?.hbsag === '阳性' ? 'HBV(+)' : t('status.safe')}
                </span>
                <span className="ml-2 px-2.5 py-0.5 bg-slate-900 text-white rounded-lg text-[10px] font-black shadow-sm uppercase tracking-tighter">
                  {patient.bedNumber} {t('info.bedNumber')}
                </span>
              </div>
              <div className="text-[11px] text-slate-500 mt-1.5 flex items-center gap-5 font-black uppercase">
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
            <p className="text-[10px] text-slate-400 uppercase font-black tracking-widest">{t('info.dialysisAge')}</p>
            <p className="font-black text-xl text-slate-900">3.2 <span className="text-[10px] font-bold">Years</span></p>
          </div>
          <div className="text-right">
            <p className="text-[10px] text-slate-400 uppercase font-black tracking-widest">{t('info.targetUF')}</p>
            <p className="font-black text-xl text-blue-600">2.50 <span className="text-[10px] font-bold">L</span></p>
          </div>
          <div className="h-10 w-px bg-slate-200 mx-2"></div>
          <div className="flex gap-1.5 items-center">
            <button
              onClick={() => setFocusBarOpen(!focusBarOpen)}
              className={`p-2.5 rounded-2xl transition-all relative ${focusBarOpen ? 'bg-blue-600 text-white shadow-lg' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'}`}
              title={focusBarOpen ? t('label.collapseFocus') : t('label.expandFocus')}
            >
              <PanelRight size={20}/>
              {!focusBarOpen && (
                <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full border border-white"></span>
              )}
            </button>
            <div className="flex gap-1 ml-1">
              <button onClick={() => handleSwitchPatient('prev')} className="p-2.5 bg-white border border-slate-200 rounded-2xl hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95">
                <ChevronLeft size={20}/>
              </button>
              <button onClick={() => handleSwitchPatient('next')} className="p-2.5 bg-white border border-slate-200 rounded-2xl hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95">
                <ChevronRight size={20}/>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* 左侧导航 */}
        <div className="bg-white border-r border-slate-200 flex flex-col shrink-0 z-10 w-52">
          <div className="w-52 h-full flex flex-col p-5 shrink-0 overflow-y-auto no-scrollbar">
            <nav className="space-y-1">
              <p className="text-[10px] font-black text-slate-400 uppercase px-4 mb-3 tracking-[0.2em] opacity-60">{t('nav.holisticView')}</p>
              {navItems.map(nav => (
                <button
                  key={nav.id}
                  onClick={() => setActiveTab(nav.id)}
                  className={`w-full flex items-center px-4 py-4 text-sm font-black rounded-2xl transition-all duration-300 ${
                    activeTab === nav.id
                      ? 'bg-blue-600 text-white shadow-xl shadow-blue-200 translate-x-1'
                      : 'text-slate-500 hover:bg-slate-50'
                  }`}
                >
                  <nav.icon size={20} className={`mr-3 ${activeTab === nav.id ? 'text-white' : 'text-blue-500 opacity-70'}`}/>
                  {nav.label}
                </button>
              ))}
            </nav>

            <div className="mt-auto p-4 bg-slate-50 rounded-3xl border border-slate-100 flex items-center gap-3">
              <div className="w-10 h-10 rounded-2xl bg-blue-100 text-blue-600 flex items-center justify-center font-black uppercase text-xs">DR</div>
              <div className="truncate">
                <p className="text-xs font-black text-slate-800 truncate">{patient.doctorName}</p>
                <p className="text-[10px] text-slate-400 font-bold uppercase tracking-tighter">{t('info.attendingDoctor')}</p>
              </div>
            </div>
          </div>
        </div>

        {/* 中部内容区 */}
        <div className="flex-1 overflow-y-auto p-8 scroll-smooth no-scrollbar">
          <div className={`${activeTab === 'scheme_order' || activeTab === 'monthly_summary' ? 'max-w-none' : 'max-w-6xl'} mx-auto pb-10 h-full transition-all duration-300`}>
            {renderTabContent()}
          </div>
        </div>

        {/* 右侧焦点栏 */}
        <aside
          className={`bg-white border-l border-slate-200 flex flex-col shrink-0 z-10 shadow-2xl transition-all duration-500 ease-in-out ${focusBarOpen ? 'w-85 translate-x-0' : 'w-0 translate-x-full overflow-hidden border-none'}`}
          style={{ width: focusBarOpen ? '340px' : '0' }}
        >
          <div className="p-6 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center sticky top-0 shrink-0">
            <h3 className="text-sm font-black text-slate-800 flex items-center uppercase tracking-wider">
              <ShieldAlert size={20} className="mr-2 text-red-500"/> {t('section.clinicalFocus')}
            </h3>
            <button onClick={() => setFocusBarOpen(false)} className="p-1.5 text-gray-400 hover:bg-slate-200 rounded-full transition-colors">
              <X size={18}/>
            </button>
          </div>
          <div className="flex-1 overflow-y-auto p-6 space-y-8 no-scrollbar">
            <div className="space-y-3">
              <p className="text-[10px] font-black text-slate-400 uppercase tracking-widest opacity-70">{t('section.criticalValues')}</p>
              <div className="p-5 bg-red-50/50 border border-red-100 rounded-3xl relative overflow-hidden group">
                <div className="absolute top-0 right-0 w-16 h-16 bg-red-100/50 rounded-bl-full animate-pulse opacity-50"></div>
                <p className="text-[10px] text-red-400 font-black mb-1">{t('ai.bloodK')}</p>
                <div className="flex justify-between items-baseline">
                  <span className="text-3xl font-black text-red-600 font-mono">6.2</span>
                  <span className="text-[10px] text-red-400 font-black uppercase">mmol/L</span>
                </div>
                <p className="text-[10px] text-red-400 font-bold mt-2 flex items-center italic">
                  <AlertCircle size={10} className="mr-1"/> {t('ai.suggestion')}
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <p className="text-[10px] font-black text-slate-400 uppercase tracking-widest opacity-70">{t('section.documentStatus')}</p>
              <div className="space-y-2">
                <div className="flex items-center justify-between p-4 bg-white border border-slate-100 rounded-2xl shadow-sm hover:border-blue-200 transition-colors">
                  <span className="text-xs font-black text-slate-700">{t('nav.monthlySummary')}</span>
                  <span className="text-[10px] text-orange-500 font-black uppercase bg-orange-50 px-2 py-0.5 rounded">{t('status.pending')}</span>
                </div>
                <div className="flex items-center justify-between p-4 bg-white border border-slate-100 rounded-2xl shadow-sm opacity-60">
                  <span className="text-xs font-black text-slate-700">{t('ai.treatmentConsent')}</span>
                  <span className="text-[10px] text-green-500 font-black uppercase bg-green-50 px-2 py-0.5 rounded">{t('status.done')}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6 bg-slate-50 border-t border-slate-100 shrink-0">
            <div className="flex items-center text-slate-400 text-[10px] font-black uppercase tracking-tight">
              <Clock size={12} className="mr-2"/> {t('label.lastSync')}: 14:32:05
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}
