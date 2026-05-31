// AI-HMS 全局类型定义
// 老库/HDIS PascalCase row 类型
// Patient → LegacyPatientRow，避免与 original.ts 中的 UI Patient 冲突

// 患者信息
export interface LegacyPatientRow {
    Id: number
    TenantId: number
    Name: string
    Spell: string
    Type: string
    TreatmentStatus: string
    OutComeStatus: string
    Gender: string
    BirthDate?: string
    PhoneNo?: string
    ExpenseType?: string
    DialysisNo?: string
}

// 治疗记录
export interface Treatment {
    Id: number
    TenantId: number
    PatientId: number
    ScheduleId: number
    SignInTime: string
    Status: string
    WardName?: string
    BedName?: string
    ShiftName?: string
}

// 体征数据
export interface VitalSigns {
    Id: number
    TreatmentId: number
    SBP: number  // 收缩压
    DBP: number  // 舒张压
    HeartRate: number
    BodyTemp?: number
    OperateTime: string
}

// 透析参数
export interface DialysisParam {
    Id: number
    TreatmentId: number
    TMP: number
    UFQuantity: number
    BF: number
    OperateTime: string
}

// 排班信息
export interface LegacyPatientShiftRow {
    Id: number
    PatientId: number
    TreatmentTime: string
    ShiftId: number
    WardId: number
    BedId: number
}

// API 响应通用格式
export interface ApiResponse<T> {
    data: T[]
    RowCount?: number
}
