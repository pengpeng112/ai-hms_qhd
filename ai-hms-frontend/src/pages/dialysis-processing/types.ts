export interface Patient {
  id: string
  name: string
  bedId: string
  gender: string
  age: number
  status: string
  patientId: string
  costType: string
  dialysisAge: string
  dryWeight: number
  treatmentPlan: string
}

export interface PreAssessmentFormValue {
  weight: string
  extraWeight: string
  targetUf: string
  sbp: string
  dbp: string
  heartRate: string
  respiration: string
  temperature: string
  pressurePoint: string
  aSite: string
  vSite: string
  consciousness: string
  nurseLevel: string
  notes: string
  symptoms: string[]
}

export const ExecutionTab = {
  PRE_ASSESSMENT: '透前评估',
  TODAY_PRESCRIPTION: '当日处方',
  DUAL_CHECK: '双人核对',
  MEDICAL_ORDERS: '透析医嘱',
  MID_MONITORING: '透中监测',
  POST_ASSESSMENT: '透后评估',
  EDUCATION: '健康宣教',
  SUMMARY: '透析小结',
} as const

export type ExecutionTab = typeof ExecutionTab[keyof typeof ExecutionTab]
