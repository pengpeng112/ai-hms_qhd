import { message } from 'antd'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { restApi } from '@/services'
import type { RestPatient, RestTreatment } from '@/services'
import { getErrorMessage, getRequestErrorKind, getTreatmentLoadErrorMessage } from '@/services/restClient'
import type {
  TreatmentBeforeSignsRequest,
  TreatmentAfterSignsRequest,
  TreatmentDuringParamRequest,
  TreatmentFirstCheckRequest,
  TreatmentSecondCheckRequest,
  TreatmentDisinfectionRequest,
} from '@/services/restClient'
import PatientListSidebar from './components/PatientListSidebar'
import PatientSummaryHeader from './components/PatientSummaryHeader'
import DialysisSummary from './execution/DialysisSummary'
import HealthEducation from './execution/HealthEducation'
import MedicalOrders from './execution/MedicalOrders'
import MidMonitoring from './execution/MidMonitoring'
import PostAssessment from './execution/PostAssessment'
import PreAssessment from './execution/PreAssessment'
import TodayPrescription from './execution/TodayPrescription'
import Verification from './execution/Verification'
import { ExecutionTab } from './types'
import type { ExecutionTab as ExecutionTabValue, Patient } from './types'

const getTodayDateParam = () => new Date().toISOString().slice(0, 10)

type TreatmentLoadState = 'idle' | 'loading' | 'ready' | 'missing' | 'server-error' | 'network-error'

function mapRestPatientToExecutionPatient(patient: RestPatient): Patient {
  const statusMap: Record<string, string> = {
    active: '透析中',
    inactive: '候诊',
    discharged: '已结束',
  }

  return {
    id: String(patient.id),
    name: patient.name || '-',
    bedId: patient.bedNumber || '未排床',
    gender: patient.gender === 'M' ? '男' : patient.gender === 'F' ? '女' : patient.gender || '-',
    age: patient.age || 0,
    status: statusMap[patient.status] || patient.status || '候诊',
    patientId: String(patient.id),
    costType: patient.insuranceType || '自费',
    dialysisAge: '待补充',
    dryWeight: patient.dryWeight ?? 0,
    treatmentPlan: patient.defaultMode || 'HD',
  }
}

export default function DialysisExecution() {
  const [searchParams] = useSearchParams()
  const requestedPatientId = searchParams.get('patientId') || ''
  const [patients, setPatients] = useState<Patient[]>([])
  const [selectedPatientId, setSelectedPatientId] = useState('')
  const [activeTab, setActiveTab] = useState<ExecutionTabValue>(ExecutionTab.PRE_ASSESSMENT)
  const [isPatientListVisible, setIsPatientListVisible] = useState(true)
  const [loadingPatients, setLoadingPatients] = useState(true)
  const [loadingTreatment, setLoadingTreatment] = useState(false)
  const [currentTreatment, setCurrentTreatment] = useState<RestTreatment | null>(null)
  const [treatmentLoadState, setTreatmentLoadState] = useState<TreatmentLoadState>('idle')
  const treatmentRequestIdRef = useRef(0)

  const selectedPatient = useMemo(
    () => patients.find((item) => item.id === selectedPatientId) ?? null,
    [patients, selectedPatientId]
  )

  const handleSelectPatient = (patientId: string) => {
    if (patientId === selectedPatientId) return

    treatmentRequestIdRef.current += 1
    setLoadingTreatment(true)
    setCurrentTreatment(null)
    setTreatmentLoadState('loading')
    setSelectedPatientId(patientId)
  }

  useEffect(() => {
    const loadPatients = async () => {
      setLoadingPatients(true)
      try {
        const res = await restApi.getPatientList({ page: 1, pageSize: 500 })
        const items = (res.data.items || []).map(mapRestPatientToExecutionPatient)
        setPatients(items)
        setSelectedPatientId((prev) => {
          if (requestedPatientId && items.some((item) => item.id === requestedPatientId)) {
            return requestedPatientId
          }
          return prev || items[0]?.id || ''
        })
      } catch (error) {
        console.error('[DialysisExecution] load patients failed', error)
        message.error(getErrorMessage(error))
      } finally {
        setLoadingPatients(false)
      }
    }

    void loadPatients()
  }, [requestedPatientId])

  useEffect(() => {
    if (!selectedPatientId) {
      setCurrentTreatment(null)
      setLoadingTreatment(false)
      setTreatmentLoadState('idle')
      return
    }

    const requestId = ++treatmentRequestIdRef.current
    setCurrentTreatment(null)
    setLoadingTreatment(true)
    setTreatmentLoadState('loading')

    const loadTodayTreatment = async () => {
      let shouldClearLoading = true

      try {
        const res = await restApi.getPatientTreatmentByDate(selectedPatientId, getTodayDateParam())
        if (treatmentRequestIdRef.current !== requestId) {
          shouldClearLoading = false
          return
        }
        setCurrentTreatment(res.data ?? null)
        setTreatmentLoadState(res.data ? 'ready' : 'missing')
      } catch (error) {
        if (treatmentRequestIdRef.current !== requestId) {
          shouldClearLoading = false
          return
        }
        console.error('[DialysisExecution] load treatment failed', error)
        const errorKind = getRequestErrorKind(error)
        if (errorKind === 'auth' || errorKind === 'forbidden') {
          setTreatmentLoadState('idle')
          message.error(getErrorMessage(error))
          return
        }
        if (errorKind === 'not_found') {
          setCurrentTreatment(null)
          setTreatmentLoadState('missing')
          return
        }
        setCurrentTreatment(null)
        setTreatmentLoadState(errorKind === 'network' ? 'network-error' : 'server-error')
        message.error(getTreatmentLoadErrorMessage(error))
      } finally {
        if (shouldClearLoading) {
          setLoadingTreatment(false)
        }
      }
    }

    void loadTodayTreatment()
  }, [selectedPatientId])

  const reloadTodayTreatment = async () => {
    if (!selectedPatientId) {
      setCurrentTreatment(null)
      setLoadingTreatment(false)
      return null
    }

    const requestId = ++treatmentRequestIdRef.current
    setLoadingTreatment(true)
    setTreatmentLoadState('loading')
    let shouldClearLoading = true

    try {
      const refreshed = await restApi.getPatientTreatmentByDate(selectedPatientId, getTodayDateParam())
      if (treatmentRequestIdRef.current !== requestId) {
        shouldClearLoading = false
        return null
      }
      setCurrentTreatment(refreshed.data ?? null)
      setTreatmentLoadState(refreshed.data ? 'ready' : 'missing')
      return refreshed.data ?? null
      } catch (error) {
        if (treatmentRequestIdRef.current !== requestId) {
          shouldClearLoading = false
          return null
        }
        const errorKind = getRequestErrorKind(error)
        if (errorKind === 'auth' || errorKind === 'forbidden') {
          setTreatmentLoadState('idle')
          message.error(getErrorMessage(error))
          return null
        }
        if (errorKind === 'not_found') {
          setCurrentTreatment(null)
          setTreatmentLoadState('missing')
        return null
      }
      setTreatmentLoadState(errorKind === 'network' ? 'network-error' : 'server-error')
      message.error(getTreatmentLoadErrorMessage(error))
      return null
    } finally {
      if (shouldClearLoading) {
        setLoadingTreatment(false)
      }
    }
  }

  const ensuringRef = useRef(false)

  const ensureTodayTreatment = async (status: number) => {
    if (!selectedPatient) return null
    const numericPatientId = Number(selectedPatient.id)
    if (!Number.isFinite(numericPatientId)) return null

    if (ensuringRef.current) return null
    ensuringRef.current = true
    try {
      if (currentTreatment) {
        const updated = await restApi.updateTreatment(currentTreatment.id, {
          status,
          notes: currentTreatment.notes ?? '',
        })
        setCurrentTreatment(updated.data)
        setTreatmentLoadState('ready')
        return updated.data
      }

      if (treatmentLoadState === 'server-error' || treatmentLoadState === 'network-error') {
        message.error(treatmentLoadState === 'network-error' ? '网络异常，请检查连接' : '治疗记录加载失败，请重试')
        return null
      }

      const created = await restApi.createTreatment({
        patientId: String(numericPatientId),
        treatmentDate: new Date().toISOString(),
        type: 1,
        status,
        notes: '// TODO: 补充治疗子表 API',
      })
      setCurrentTreatment(created.data)
      setTreatmentLoadState('ready')
      return created.data
    } catch (error) {
      console.error('[DialysisExecution] ensureTodayTreatment failed', error)
      message.error(getErrorMessage(error))
      return null
    } finally {
      ensuringRef.current = false
    }
  }

  const handleCreateTodayTreatment = async () => {
    const treatment = await ensureTodayTreatment(0)
    if (!treatment) return
    message.success('治疗记录已创建')
    await reloadTodayTreatment()
  }

  const handleSavePreAssessment = async (payload: TreatmentBeforeSignsRequest) => {
    const treatment = await ensureTodayTreatment(0)
    if (!treatment) return
    await restApi.saveTreatmentBeforeSigns(treatment.id, payload)
    await reloadTodayTreatment()
    message.success('透前评估已保存')
  }

  const handleSaveFirstCheck = async (payload: TreatmentFirstCheckRequest) => {
    const treatment = await ensureTodayTreatment(0)
    if (!treatment) return
    await restApi.saveTreatmentFirstCheck(treatment.id, payload)
    await reloadTodayTreatment()
    message.success('首次核对已保存')
  }

  const handleSaveSecondCheck = async (payload: TreatmentSecondCheckRequest) => {
    const treatment = await ensureTodayTreatment(0)
    if (!treatment) return
    await restApi.saveTreatmentSecondCheck(treatment.id, payload)
    await reloadTodayTreatment()
    message.success('二次核对已保存')
  }

  const handleSaveDisinfection = async (payload: TreatmentDisinfectionRequest) => {
    const treatment = await ensureTodayTreatment(1)
    if (!treatment) return
    await restApi.saveTreatmentDisinfection(treatment.id, payload)
    await reloadTodayTreatment()
    message.success('消毒登记已保存')
  }

  const handleCreateDuringParam = async (payload: TreatmentDuringParamRequest) => {
    const treatment = await ensureTodayTreatment(1)
    if (!treatment) return
    await restApi.createTreatmentDuringParam(treatment.id, payload)
    await reloadTodayTreatment()
  }

  const handleUpdateDuringParam = async (paramId: number, payload: TreatmentDuringParamRequest) => {
    const treatment = await ensureTodayTreatment(1)
    if (!treatment) return
    await restApi.updateTreatmentDuringParam(treatment.id, paramId, payload)
    await reloadTodayTreatment()
  }

  const handleDeleteDuringParam = async (paramId: number) => {
    if (!currentTreatment) return
    await restApi.deleteTreatmentDuringParam(currentTreatment.id, paramId)
    await reloadTodayTreatment()
  }

  const handleSavePostAssessment = async (payload: TreatmentAfterSignsRequest) => {
    const treatment = await ensureTodayTreatment(1)
    if (!treatment) return
    await restApi.saveTreatmentAfterSigns(treatment.id, payload)
    await reloadTodayTreatment()
  }

  const handleSubmitPostAssessment = async (payload: TreatmentAfterSignsRequest) => {
    const treatment = await ensureTodayTreatment(1)
    if (!treatment) return
    await restApi.submitTreatmentPostAssessment(treatment.id, payload)
    await reloadTodayTreatment()
  }

  const tabs = useMemo(() => Object.values(ExecutionTab), [])

  const content = (() => {
    if (!selectedPatient) return null

    switch (activeTab) {
      case ExecutionTab.PRE_ASSESSMENT:
        return (
          <PreAssessment
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            saving={loadingTreatment}
            treatmentLoading={loadingTreatment}
            onSave={handleSavePreAssessment}
          />
        )
      case ExecutionTab.TODAY_PRESCRIPTION:
        return (
          <TodayPrescription
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            treatmentLoading={loadingTreatment}
          />
        )
      case ExecutionTab.DUAL_CHECK:
        return (
          <Verification
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            treatmentLoading={loadingTreatment}
            onSaveFirstCheck={handleSaveFirstCheck}
            onSaveSecondCheck={handleSaveSecondCheck}
            onSaveDisinfection={handleSaveDisinfection}
          />)
      case ExecutionTab.MEDICAL_ORDERS:
        return <MedicalOrders patient={selectedPatient} />
      case ExecutionTab.MID_MONITORING:
        return (
          <MidMonitoring
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            treatmentLoading={loadingTreatment}
            onCreate={handleCreateDuringParam}
            onUpdate={handleUpdateDuringParam}
            onDelete={handleDeleteDuringParam}
          />
        )
      case ExecutionTab.POST_ASSESSMENT:
        return (
          <PostAssessment
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            treatmentLoading={loadingTreatment}
            onSave={handleSavePostAssessment}
            onSubmit={handleSubmitPostAssessment}
          />
        )
      case ExecutionTab.EDUCATION:
        return (
          <HealthEducation
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            treatmentLoading={loadingTreatment}
          />
        )
      case ExecutionTab.SUMMARY:
        return (
          <DialysisSummary
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            treatmentLoading={loadingTreatment}
            onTreatmentUpdated={(t) => setCurrentTreatment(t)}
          />
        )
      default:
        return (
          <PreAssessment
            key={`${activeTab}-${selectedPatientId}`}
            patient={selectedPatient}
            treatment={currentTreatment}
            saving={loadingTreatment}
            treatmentLoading={loadingTreatment}
            onSave={handleSavePreAssessment}
          />
        )
    }
  })()

  return (
    <div className="relative flex h-full overflow-hidden bg-slate-100">
      <div
        className={`overflow-hidden border-r border-slate-200 transition-all duration-300 ease-in-out ${
          isPatientListVisible ? 'w-60' : 'w-0'
        }`}
      >
        <PatientListSidebar
          patients={patients}
          selectedId={selectedPatientId}
          onSelect={(patient) => handleSelectPatient(patient.id)}
          isVisible={isPatientListVisible}
        />
      </div>

      <button
        type="button"
        onClick={() => setIsPatientListVisible((value) => !value)}
        className="absolute top-1/2 z-20 -translate-y-1/2 transition-all duration-300"
        style={{ left: isPatientListVisible ? '244px' : '-4px' }}
      >
        <span className="inline-flex h-20 w-5 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-400 shadow-sm hover:bg-blue-50 hover:text-blue-600">
          {isPatientListVisible ? <ChevronLeft size={14} /> : <ChevronRight size={14} />}
        </span>
      </button>

      <div className="flex min-w-0 flex-1 flex-col bg-white">
        <div className="shrink-0 border-b border-slate-200 bg-white px-5 py-2.5">
          <div className="mb-3 flex items-center gap-[10px] overflow-x-auto no-scrollbar whitespace-nowrap">
            {tabs.map((tab) => (
              <button
                key={tab}
                type="button"
                onClick={() => setActiveTab(tab)}
                className={`flex h-9 shrink-0 items-center rounded-[11px] px-4 text-[15px] font-black transition-colors ${
                  activeTab === tab
                    ? 'bg-blue-600 text-white shadow-[0_8px_20px_rgba(29,99,255,0.22)]'
                    : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                }`}
              >
                {tab}
              </button>
            ))}
          </div>
          {selectedPatient ? <PatientSummaryHeader patient={selectedPatient} /> : null}
        </div>

        <div className="flex-1 overflow-y-auto bg-[#f5f7fb] p-5">
          {loadingPatients ? (
            <div className="rounded-lg border border-slate-200 bg-white p-10 text-center text-slate-500">
              正在加载患者列表...
            </div>
          ) : selectedPatient ? (
            <>
              {!loadingTreatment && treatmentLoadState === 'missing' ? (
                <section className="mb-3 flex min-h-[58px] items-center gap-3 rounded-2xl border border-emerald-300 bg-emerald-50 px-4 py-2">
                  <div className="flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-xl bg-emerald-600 text-lg font-black text-white">+</div>
                  <div className="min-w-0 flex-1">
                    <div className="text-[15px] font-black text-emerald-800">暂无治疗记录</div>
                    <div className="text-[13px] font-semibold text-emerald-700">可先创建今日治疗记录，再继续录入和查看。</div>
                  </div>
                  <button
                    type="button"
                    onClick={handleCreateTodayTreatment}
                    className="ml-auto h-[34px] shrink-0 rounded-[10px] bg-emerald-600 px-4 text-[13px] font-black text-white hover:bg-emerald-700"
                  >
                    创建治疗记录
                  </button>
                </section>
              ) : null}

              {!loadingTreatment && (treatmentLoadState === 'server-error' || treatmentLoadState === 'network-error') ? (
                <div className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-rose-900">
                  <div className="text-sm font-semibold">
                    {treatmentLoadState === 'network-error' ? '网络异常，请检查连接' : '治疗记录加载失败，请重试'}
                  </div>
                  <button
                    type="button"
                    onClick={() => void reloadTodayTreatment()}
                    className="mt-2 rounded-lg bg-rose-600 px-4 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-rose-700"
                  >
                    重试加载
                  </button>
                </div>
              ) : null}

              {content}
            </>
          ) : (
            <div className="rounded-lg border border-slate-200 bg-white p-10 text-center text-slate-500">
              暂无可用患者
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
