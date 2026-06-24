import { useState, useEffect, useCallback, useRef } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { message } from 'antd'
import { ArrowLeft, ChevronLeft, ChevronRight, Activity, User, Stethoscope, ClipboardList, FileText, FlaskConical, GitBranch, Clock, Calendar, ShieldAlert, HeartPulse, FileSignature, AlertTriangle, Pill, Scale, CreditCard } from 'lucide-react'
import { restApi, convertCoreResponseToPatient } from '@/services/restClient'
import { getErrorMessage } from '@/services/restClient'
import type { Patient } from '@/types/original'
import { LoadingState } from '@/components/ui'

import {
  OverviewTab, BasicInfoTab, TreatmentPlanTab, MedicalRecordTab,
  SchemeOrderTab, LabsExamsTab, VascularTab, AdverseTab, MedicationTab, DryWeightTab,
  NursingTab, ConsentTab, HistoryTab, MonthlySummaryTab, BillingTab,
} from './patient-detail/tabs'
import InfectiousPanel from '@/components/infectious/InfectiousPanel'
import PatientHeader from './patient-detail/PatientHeader'
import ClinicalFocusDrawer from './patient-detail/ClinicalFocusDrawer'
import TodoPopover from './patient-detail/TodoPopover'
import FloatingActionButtons from './patient-detail/FloatingActionButtons'

type SectionID =
  | 'overview'        // 全息透析概览
  | 'basic_info'      // 基本信息档案
  | 'treatment_plan'  // 治疗方案管理
  | 'medical_record'  // 临床病史档案
  | 'scheme_order'    // 长期方案/医嘱
  | 'labs_exams'      // 检查检验报告
  | 'vascular'        // 血管通路评估
  | 'adverse'         // 不良事件
  | 'medication'      // 用药与给药
  | 'dry_weight'      // 干体重评估
  | 'nursing'         // 护理文书
  | 'consent'         // 知情同意
  | 'billing'         // 收费归集
  | 'history'         // 治疗详情历史
  | 'monthly_summary' // 月份评估小结
  | 'infectious'      // 传染病筛查

const MENU_ITEMS: { key: SectionID; label: string; icon: React.ReactNode }[] = [
  { key: 'overview',        label: '全息透析概览', icon: <Activity size={18} /> },
  { key: 'basic_info',      label: '基本信息档案', icon: <User size={18} /> },
  { key: 'treatment_plan',  label: '治疗方案管理', icon: <Stethoscope size={18} /> },
  { key: 'medical_record',  label: '临床病史档案', icon: <ClipboardList size={18} /> },
  { key: 'scheme_order',    label: '长期方案/医嘱', icon: <FileText size={18} /> },
  { key: 'labs_exams',      label: '检查检验报告', icon: <FlaskConical size={18} /> },
  { key: 'infectious',      label: '传染病筛查',   icon: <ShieldAlert size={18} /> },
  { key: 'vascular',        label: '血管通路评估', icon: <GitBranch size={18} /> },
  { key: 'adverse',         label: '不良事件',     icon: <AlertTriangle size={18} /> },
  { key: 'medication',      label: '用药与给药',   icon: <Pill size={18} /> },
  { key: 'dry_weight',      label: '干体重评估',   icon: <Scale size={18} /> },
  { key: 'nursing',         label: '护理文书',     icon: <HeartPulse size={18} /> },
  { key: 'consent',         label: '知情同意',     icon: <FileSignature size={18} /> },
  { key: 'billing',         label: '收费归集',     icon: <CreditCard size={18} /> },
  { key: 'history',         label: '治疗详情历史', icon: <Clock size={18} /> },
  { key: 'monthly_summary', label: '月份评估小结', icon: <Calendar size={18} /> },
]

function SidebarMenu({ active, onSelect }: { active: SectionID; onSelect: (key: SectionID) => void }) {
  return (
    <aside className="w-[220px] shrink-0 bg-white border-r border-slate-200 flex flex-col h-full overflow-y-auto">
      <div className="px-3.5 pt-5 pb-4">
        <h2 className="text-xs font-bold text-blue-400 tracking-wide">诊疗全息视图</h2>
      </div>

      <nav className="flex-1 px-3 space-y-1">
        {MENU_ITEMS.map((item) => {
          const isActive = active === item.key
          return (
            <button
              key={item.key}
              type="button"
              onClick={() => onSelect(item.key)}
              className={`flex items-center gap-3 w-full h-11 px-3 rounded-xl text-sm font-medium transition ${
                isActive
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-600 hover:bg-blue-50 hover:text-slate-800'
              }`}
            >
              <span className={isActive ? 'text-white' : 'text-blue-500'}>{item.icon}</span>
              <span>{item.label}</span>
            </button>
          )
        })}
      </nav>

      {/* 底部医生信息卡片 */}
      <div className="px-3 pb-5 pt-3 shrink-0">
        <div className="flex items-center gap-3 rounded-xl bg-blue-50/70 px-3 py-3">
          <div className="w-9 h-9 rounded-full bg-blue-200 flex items-center justify-center shrink-0">
            <span className="text-xs font-bold text-blue-600">DR</span>
          </div>
          <div className="min-w-0">
            <div className="text-xs font-semibold text-slate-500 truncate">主治医生 / 团队</div>
          </div>
        </div>
      </div>
    </aside>
  )
}

export default function PatientDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { t } = useTranslation('patient')
  const [searchParams, setSearchParams] = useSearchParams()

  const activeSection = (searchParams.get('section') || 'overview') as SectionID

  const [loading, setLoading] = useState(true)
  const [avatarLoadFailed, setAvatarLoadFailed] = useState(false)
  const [patient, setPatient] = useState<Patient | null>(null)
  const [switchingPatient, setSwitchingPatient] = useState(false)

  // 临床焦点抽屉状态
  const [clinicalFocusOpen, setClinicalFocusOpen] = useState(false)

  // 待办任务弹窗状态
  const [todoOpen, setTodoOpen] = useState(false)
  const todoButtonRef = useRef<HTMLButtonElement>(null)

  const setSection = useCallback((key: SectionID) => {
    setSearchParams({ section: key }, { replace: true })
  }, [setSearchParams])

  // 打开临床焦点时关闭待办弹窗
  const handleOpenClinicalFocus = useCallback(() => {
    setTodoOpen(false)
    setClinicalFocusOpen(true)
  }, [])

  // 打开待办弹窗时关闭临床焦点
  const handleToggleTodo = useCallback(() => {
    setClinicalFocusOpen(false)
    setTodoOpen(prev => !prev)
  }, [])

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

  const renderContent = () => {
    switch (activeSection) {
      case 'overview':
        return <OverviewTab patient={patient} onNavigate={(key: string) => setSection(key as SectionID)} />
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
      case 'infectious':
        return <InfectiousPanel patientId={patient.id} />
      case 'vascular':
        return <VascularTab patient={patient} />
      case 'adverse':
        return <AdverseTab patient={patient} />
      case 'medication':
        return <MedicationTab patient={patient} />
      case 'dry_weight':
        return <DryWeightTab patient={patient} />
      case 'nursing':
        return <NursingTab patient={patient} />
      case 'consent':
        return <ConsentTab patient={patient} />
      case 'billing':
        return <BillingTab patient={patient} />
      case 'history':
        return <HistoryTab patient={patient} />
      case 'monthly_summary':
        return <MonthlySummaryTab patient={patient} />
      default:
        return <OverviewTab patient={patient} onNavigate={(key: string) => setSection(key as SectionID)} />
    }
  }

  // 判断是否有风险
  const hasRisk = patient.riskLevel === '高危' || patient.infection?.hbsag === '阳性'

  return (
    <div className="h-full flex flex-col max-w-[1800px] mx-auto overflow-hidden bg-[#f5f8fc]">
      {/* 顶部信息条 */}
      <div className="bg-white border-b border-[#e6ebf3] px-6 py-4 flex items-center justify-between shadow-sm z-20 shrink-0">
        <div className="flex items-center gap-4">
          <button onClick={handleBack} className="text-slate-400 hover:text-slate-900 transition-colors p-2.5 rounded-lg hover:bg-slate-100 active:scale-95 shrink-0">
            <ArrowLeft size={22} />
          </button>
          <PatientHeader patient={patient} avatarFailed={avatarLoadFailed} onAvatarError={() => setAvatarLoadFailed(true)} />
        </div>

        <div className="flex items-center gap-3 shrink-0">
          {/* 悬浮操作按钮 */}
          <FloatingActionButtons
            onClinicalFocusClick={handleOpenClinicalFocus}
            onTodoClick={handleToggleTodo}
            hasRisk={hasRisk}
            todoCount={3}
            todoButtonRef={todoButtonRef}
          />

          {/* 待办任务弹窗 */}
          <div className="relative">
            <TodoPopover
              open={todoOpen}
              onClose={() => setTodoOpen(false)}
              anchorRef={todoButtonRef}
            />
          </div>

          {/* 患者切换按钮 */}
          <div className="flex gap-1 ml-2">
            <button disabled={switchingPatient} onClick={() => handleSwitchPatient('prev')} className="p-2.5 bg-white border border-[#e6ebf3] rounded-lg hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
              <ChevronLeft size={20}/>
            </button>
            <button disabled={switchingPatient} onClick={() => handleSwitchPatient('next')} className="p-2.5 bg-white border border-[#e6ebf3] rounded-lg hover:bg-slate-50 text-slate-500 shadow-sm transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
              <ChevronRight size={20}/>
            </button>
          </div>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* 左侧垂直菜单 */}
        <SidebarMenu active={activeSection} onSelect={setSection} />

        {/* 主内容区 - 自动扩展占满剩余宽度 */}
        <div className="flex-1 overflow-y-auto no-scrollbar">
          <div className="p-6">
            {renderContent()}
          </div>
        </div>
      </div>

      {/* 临床焦点抽屉 */}
      <ClinicalFocusDrawer
        patient={patient}
        open={clinicalFocusOpen}
        onClose={() => setClinicalFocusOpen(false)}
      />
    </div>
  )
}
