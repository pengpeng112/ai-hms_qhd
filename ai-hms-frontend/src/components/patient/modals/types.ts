// 模态框相关类型定义

// 治疗转归记录接口
export interface OutcomeRecord {
  id: string;
  type: string;          // 存储字典 code
  typeName?: string;     // 显示名称（从字典获取）
  reason: string;        // 存储字典 code
  reasonName?: string;   // 显示名称（从字典获取）
  time: string;
  remarks: string;
  registrar: string;
  registrationTime: string;
  isDoorRule: boolean;  // 是否门规
}

// 治疗历史记录接口
export interface ExtendedTreatmentHistory {
  id: string;
  date: string;
  mode: string;
  duration: string;
  timeRange: string;
  weightLoss: number;
  startBP: string;
  endBP: string;
  complications: string;
  doctor: string;
  doctorSummary: string;
  treatmentSummary: string;
}
