
export interface Patient {
  id: string;
  name: string;
  bedId: string;
  gender: '男' | '女';
  age: number;
  status: '透析中' | '候诊' | '已完成';
  patientId: string;
  costType: string;
  dialysisAge: string;
  dryWeight: number;
  treatmentPlan: string;
}

export interface MedicalOrder {
  id: number;
  type: '长期' | '临时';
  content: string;
  usage: string;
  doctor: string;
  orderTime: string;
  validTime: string;
  recentExecution: string;
  checked: boolean;
  executed: boolean;
  weekCount: number;
  nextScheduleDays: number;
  lastModified: string;
}

export enum ExecutionTab {
  PRE_ASSESSMENT = '透前评估',
  TODAY_PRESCRIPTION = '当日处方',
  DUAL_CHECK = '双人核对',
  MEDICAL_ORDERS = '透析医嘱',
  MID_MONITORING = '透中监测',
  POST_ASSESSMENT = '透后评估',
  EDUCATION = '健康宣教',
  SUMMARY = '透析小结'
}
