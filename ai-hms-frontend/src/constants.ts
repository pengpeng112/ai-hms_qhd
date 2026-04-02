import { UserRole } from './types/original'
import type {
  Patient,
  DashboardCardConfig,
  MonitorDevice,
  StaffMember,
  Shift,
  PatientScheduleItem,
  TreatmentPlan,
  DetailedOrder,
  MedGroup,
} from './types/original'

// NOTE: MOCK_TREATMENT_PLAN is still used by non-Phase-1 pages (Dashboard, Schedule, DialysisProcessing).
// It is no longer imported by the patient detail main path.
export const MOCK_TREATMENT_PLAN: TreatmentPlan = {
  weeklyFrequency: 3,
  biweeklyFrequency: 3,
  duration: 4,
  dryWeight: 67.6,
  extraWeight: 0.6,
  vascularAccess: 'AVG-上臂',
  indicators: {
    mode: 'HD',
    bloodFlow: 200,
    bv: '',
    frequencyDesc: '一周三次',
    autoConfirm: true,
    status: '启用',
    notes: ''
  },
  anticoagulant: {
    initialDrug: '那屈肝素钙注射液(百力舒)-615axiu/ml',
    initialDose: '307.5',
    maintenanceDrug: '',
    infusionRate: '',
    infusionTime: '',
    maintenanceDose: '',
    totalDose: '307.5'
  },
  parameters: {
    dialysateType: 'A液+B液',
    dialysateGroup: 'A液16-含糖-K3\\Ca1.25',
    flowRate: 500,
    na: 140,
    ca: 1.25,
    k: 2.5,
    hco3: 31,
    glucose: '含糖 (1.1)',
    conductivity: 14.2,
    temp: 36.5,
    volume: 120
  },
  materials: [
    { id: '1', name: 'JRHLL-025', category: '血路管', count: 1, code: '', brand: '', spec: '', note: '' },
    { id: '2', name: '10ML注射器-10ML', category: '其他', count: 2, code: '', brand: '', spec: '10ML', note: '' },
    { id: '3', name: '15G', category: '透析、血滤器', count: 1, code: '', brand: 'NIPRO', spec: '', note: '' },
    { id: '4', name: '内瘘包', category: '护理包', count: 1, code: '1102011534', brand: '', spec: '', note: '' },
    { id: '5', name: '锐针-16G', category: '穿刺针', count: 2, code: '', brand: 'NIPRO', spec: '', note: '' },
  ],
  adjustmentHistory: [
    { id: 'adj1', date: '2023-10-25 08:30', content: '超滤量: 由【2.0】L 调整为【2.1】L', operator: '李医生', reason: '患者水肿加重' },
    { id: 'adj2', date: '2023-10-20 14:15', content: '干体重: 由【68.0】kg 调整为【67.6】kg', operator: '王医生', reason: '营养状况评估' }
  ]
};

// Added missing MOCK_DEVICE_INVENTORY
export const MOCK_DEVICE_INVENTORY = Array.from({ length: 15 }).map((_, i) => ({
  id: `DEV-${200 + i}`,
  serial: `SN-99${100 + i}`,
  brand: i % 2 === 0 ? 'NIPRO' : 'FRESENIUS',
  model: i % 2 === 0 ? 'SUREFLUX-15G' : '5008S',
  zone: i < 5 ? 'A区' : i < 10 ? 'B区' : 'C区',
  status: i === 2 ? 'maintenance' : i === 5 ? 'error' : 'normal',
  runningHours: 1200 + i * 50,
  purchaseDate: '2023-01-15'
}));

// Added missing MOCK_DEVICE_MAINTENANCE_LOGS
export const MOCK_DEVICE_MAINTENANCE_LOGS = [
  { id: 'LOG1', date: '2023-10-25', deviceId: 'DEV-200', type: 'PM', description: '预防性维护保养', engineer: '吴工程师', result: '完成' },
  { id: 'LOG2', date: '2023-10-20', deviceId: 'DEV-202', type: 'Repair', description: '更换透析液过滤器', engineer: '吴工程师', result: '完成' },
  { id: 'LOG3', date: '2023-10-18', deviceId: 'DEV-205', type: 'Daily', description: '日常性能检查', engineer: '吴工程师', result: '完成' },
];

// Added missing MOCK_SESSION_PROCESS
export const MOCK_SESSION_PROCESS = {
  monitoring: {
    records: [
      { time: '08:49', sbp: 133, dbp: 73, hr: 65, vp: 102, tmp: 55, ufVolume: '0.23', symptoms: '无', nurse: '李俊雅' },
      { time: '08:34', sbp: 104, dbp: 60, hr: 64, vp: 108, tmp: 69, ufVolume: '0.1', symptoms: '轻微头晕', nurse: '李俊雅' },
    ]
  }
};

// NOTE: MOCK_PATIENTS is still used by non-Phase-1 pages (Dashboard, Schedule, DialysisProcessing, mockHelpers).
// It is no longer imported by the patient detail main path.
export const MOCK_PATIENTS: Patient[] = [
  {
    id: 'P001',
    name: '张伟',
    age: 65,
    gender: '男',
    bedNumber: 'A01',
    diagnosis: '慢性肾脏病 5期',
    riskLevel: '高危',
    status: '透析中',
    patientType: '门诊',
    insuranceType: '职工医保',
    dryWeight: 65.5,
    defaultMode: 'HDF',
    doctorName: '王医生',
    vitals: { bp: '145/88', hr: 78, spO2: 98, weight: 68.5 },
    dialysisParams: { 
      timeRemaining: '01:30', 
      ufRate: 300, 
      targetUf: 2.5,
      accumulatedUf: 1.2,
      bloodFlow: 240,
      dialysateFlow: 500,
      mode: 'HDF'
    },
    vascularAccess: { 
      type: 'AVF', site: '左前臂', status: '震颤良好', 
      history: '2023-01-20 建立，初期流量不稳，现已扩张良好。' 
    },
    recentLabs: [
      { id: 'L1', name: '血红蛋白', value: '105', unit: 'g/L', date: '2023-10-25', isAbnormal: true, reference: '120-160' },
      { id: 'L2', name: '血钾', value: '5.2', unit: 'mmol/L', date: '2023-10-25', isAbnormal: false, reference: '3.5-5.5' },
    ],
    orders: [
      { id: 'O1', content: '低分子肝素 2500iu iv', type: '长期', status: '已执行', doctor: '王医生', startTime: '08:30' },
    ],
    infection: { hbsag: '阴性', hcvab: '阴性', hivab: '阴性', tpab: '阴性', updateDate: '2023-10-01' },
    documents: [],
    progressNotes: [],
    medicalHistory: {
      allergies: ['青霉素'],
      primaryDisease: '慢性肾小球肾炎',
      pathology: 'IgA肾病 IV期',
      tumorInfo: '无相关病史',
      medicalHistory: '高血压病史15年。',
      complications: ['肾性高血压']
    },
    outcome: { status: '治疗中' },
    treatmentPlan: MOCK_TREATMENT_PLAN
  },
  {
    id: 'P002',
    name: '李娜',
    age: 52,
    gender: '女',
    bedNumber: 'A02',
    diagnosis: '糖尿病肾病',
    riskLevel: '中危',
    status: '透析中',
    patientType: '住院',
    insuranceType: '居民医保',
    dryWeight: 53.0,
    defaultMode: 'HD',
    doctorName: '陈主任',
    vitals: { bp: '130/80', hr: 82, spO2: 97, weight: 55.2 },
    dialysisParams: { 
      timeRemaining: '02:15', 
      ufRate: 250, 
      targetUf: 2.0,
      accumulatedUf: 0.8,
      bloodFlow: 220,
      dialysateFlow: 500,
      mode: 'HD'
    },
    vascularAccess: { type: 'TCC', site: '右颈内', status: '固定良好' },
    recentLabs: [],
    orders: [],
    infection: { hbsag: '阴性', hcvab: '阴性', hivab: '阴性', tpab: '阴性', updateDate: '2023-10-01' },
    documents: [],
    progressNotes: [],
    medicalHistory: {
      allergies: ['无'],
      primaryDisease: '糖尿病肾病',
      pathology: '未活检',
      tumorInfo: '无',
      medicalHistory: '发现血糖高20年。',
      complications: ['糖尿病视网膜病变']
    },
    outcome: { status: '治疗中' },
    treatmentPlan: MOCK_TREATMENT_PLAN
  }
];

export const DASHBOARD_CARDS: DashboardCardConfig[] = [
  { id: 'dept_overview', title: '科室运行总览', type: 'stat', size: 'large', roles: [UserRole.DOCTOR_CHIEF, UserRole.NURSE_HEAD, UserRole.ENGINEER] },
  { id: 'active_patients', title: '我的责任患者', type: 'list', size: 'large', roles: [UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY] },
  { id: 'quality_stats', title: '透析质量达标率', type: 'chart', size: 'medium', roles: [UserRole.DOCTOR_CHIEF, UserRole.NURSE_HEAD] },
  { id: 'prescription_adjust', title: '待处理处方/医嘱', type: 'action', size: 'medium', roles: [UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY, UserRole.DOCTOR_CHIEF] },
  { id: 'duty_monitor', title: '全科实时监控', type: 'monitor', size: 'large', roles: [UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY, UserRole.DOCTOR_CHIEF] },
  { id: 'nurse_workload', title: '护士工作量统计', type: 'chart', size: 'medium', roles: [UserRole.NURSE_HEAD] },
  { id: 'staff_schedule', title: '今日人员排班', type: 'list', size: 'medium', roles: [UserRole.NURSE_HEAD, UserRole.NURSE_SCHEDULER] },
  { id: 'schedule_adjust', title: '排班调整请求', type: 'action', size: 'medium', roles: [UserRole.NURSE_HEAD, UserRole.NURSE_SCHEDULER] },
  { id: 'consumables_prep', title: '今日耗材准备', type: 'inventory', size: 'large', roles: [UserRole.NURSE_HEAD, UserRole.NURSE_MANAGER] },
  { id: 'my_duty_patients', title: '本班次负责患者', type: 'list', size: 'large', roles: [UserRole.NURSE_RESPONSIBLE] },
  { id: 'device_binding', title: '床位-设备绑定管理', type: 'binding', size: 'large', roles: [] },
  { id: 'device_status_eng', title: '设备实时状态监控', type: 'monitor', size: 'large', roles: [] },
  { id: 'maintenance_logs', title: '近期维修/保养记录', type: 'list', size: 'medium', roles: [] },
];

export const MOCK_STATS_DATA = [
  { name: '08:00', value: 12 },
  { name: '10:00', value: 45 },
  { name: '12:00', value: 38 },
  { name: '14:00', value: 55 },
  { name: '16:00', value: 20 },
];

export const MOCK_VITALS_HISTORY = [
  { time: '09:00', bpSys: 140, bpDia: 90, hr: 78, ufRate: 300 },
  { time: '09:30', bpSys: 138, bpDia: 88, hr: 80, ufRate: 300 },
  { time: '10:00', bpSys: 135, bpDia: 85, hr: 76, ufRate: 280 },
  { time: '10:30', bpSys: 142, bpDia: 92, hr: 82, ufRate: 250 },
  { time: '11:00', bpSys: 130, bpDia: 80, hr: 75, ufRate: 250 },
  { time: '11:30', bpSys: 128, bpDia: 78, hr: 74, ufRate: 200 },
];

export const MOCK_MONITOR_DEVICES: MonitorDevice[] = Array.from({length: 25}).map((_, i) => {
  const bedId = `${i < 16 ? 'A' : i < 21 ? 'B' : 'C'}${String((i % 16) + 1).padStart(2, '0')}`;
  const statusRoll = Math.random();
  let status: 'normal' | 'warning' | 'alarm' | 'offline' = 'normal';
  if (i > 18) status = 'offline';
  else if (statusRoll > 0.95) status = 'alarm';
  else if (statusRoll > 0.9) status = 'warning';

  return {
    id: `DEV-${100 + i}`,
    bedNumber: bedId,
    patientName: status === 'offline' ? '' : ['张伟', '李娜', '王强', '刘燕', '赵敏', '孙行', '周波', '吴用', '郑爽', '冯唐'][i % 10],
    status,
    mode: i % 3 === 0 ? 'HDF' : 'HD',
    timeRemaining: status === 'offline' ? '--:--' : `${Math.floor(Math.random() * 3)}:${String(Math.floor(Math.random() * 60)).padStart(2, '0')}`,
    vitals: {
      sbp: status === 'alarm' ? 85 : 130 + Math.floor(Math.random() * 20 - 10),
      dbp: status === 'alarm' ? 55 : 80 + Math.floor(Math.random() * 10 - 5),
      hr: 75 + Math.floor(Math.random() * 10),
      bf: status === 'offline' ? 0 : 220 + Math.floor(Math.random() * 40),
      tmp: status === 'warning' ? 180 : 120 + Math.floor(Math.random() * 20),
      ufGoal: 3.0,
      ufVolume: status === 'offline' ? 0 : 1.5 + Math.random(),
      conductivity: 14.0,
      temp: 36.5
    },
    alarms: status === 'alarm' ? ['低血压报警'] : status === 'warning' ? ['TMP 偏高'] : []
  };
});

export const MOCK_STAFF: StaffMember[] = [
  { id: 'S01', name: '刘护士长', role: '护士', level: 'N4' },
  { id: 'S02', name: '赵护士', role: '护士', level: 'N3' },
  { id: 'S03', name: '孙护士', role: '护士', level: 'N2' },
  { id: 'S04', name: '周护士', role: '护士', level: 'N2' },
  { id: 'S05', name: '吴护士', role: '护士', level: 'N1' },
];

export const MOCK_SCHEDULE: Shift[] = [];
const today = new Date();
MOCK_STAFF.forEach(staff => {
  for (let i = 0; i < 7; i++) {
    const d = new Date(today);
    d.setDate(today.getDate() + i);
    const dateStr = d.toISOString().split('T')[0];
    const rand = Math.random();
    let type: Shift['type'] = 'OFF';
    if (rand > 0.3) type = rand > 0.6 ? 'A' : 'P';
    
    MOCK_SCHEDULE.push({
      date: dateStr,
      staffId: staff.id,
      type,
      area: 'A区'
    });
  }
});

export const MOCK_PATIENT_SCHEDULE: PatientScheduleItem[] = [];
const patientNames = ['张伟', '李娜', '王强', '刘燕', '赵敏', '孙行', '周波', '吴用', '郑爽', '冯唐'];

Array.from({length: 25}).forEach((_, i) => {
  const bedNum = `${i < 16 ? 'A' : i < 21 ? 'B' : 'C'}${String((i % 16) + 1).padStart(2, '0')}`;
  for (let d = 0; d < 7; d++) {
    const date = new Date(today);
    date.setDate(today.getDate() + d);
    const dateStr = date.toISOString().split('T')[0];
    
    if (Math.random() > 0.3) {
      MOCK_PATIENT_SCHEDULE.push({
        id: `PS-${bedNum}-${d}-M`,
        bedNumber: bedNum,
        date: dateStr,
        shift: 'Morning',
        patientName: patientNames[(i + d) % patientNames.length],
        mode: (i + d) % 3 === 0 ? 'HDF' : 'HD'
      });
    }
    if (Math.random() > 0.6) {
      MOCK_PATIENT_SCHEDULE.push({
        id: `PS-${bedNum}-${d}-A`,
        bedNumber: bedNum,
        date: dateStr,
        shift: 'Afternoon',
        patientName: patientNames[(i + d + 5) % patientNames.length],
        mode: 'HD'
      });
    }
  }
});

export const MOCK_TREATMENT_HISTORY = [
   { id: 'H01', date: '2023-10-24', mode: 'HD', duration: '4.0h', weightLoss: 2.4, startBP: '140/85', endBP: '130/80', complications: '无', doctor: '王医生' },
   { id: 'H02', date: '2023-10-22', mode: 'HDF', duration: '4.0h', weightLoss: 2.2, startBP: '145/90', endBP: '110/70', complications: '低血压', doctor: '王医生' },
   { id: 'H03', date: '2023-10-20', mode: 'HD', duration: '4.0h', weightLoss: 2.5, startBP: '138/88', endBP: '135/82', complications: '无', doctor: '李医生' },
];

// ============ 医嘱相关 Mock 数据 ============

/** 详细医嘱列表 Mock 数据 */
export const MOCK_DETAILED_ORDERS: DetailedOrder[] = [
  { id: '1', name: '左卡尼汀注射液', dose: '1.0', unit: 'g', doctor: '王医生', startTime: '2024-05-04', status: '在用', type: '长期' },
  { id: '2', name: '那屈肝素钙注射液', dose: '307.5', unit: 'axiu', doctor: '陈主任', startTime: '2024-05-04', status: '在用', type: '长期' },
  { id: '3', name: '蔗糖铁注射液', dose: '100', unit: 'mg', doctor: '张哈', startTime: '2025-12-08', stopTime: '2025-12-22', status: '停用', type: '长期' },
  { id: '4', name: '氯化钠注射液', dose: '10', unit: 'ml', doctor: '张哈', startTime: '2025-12-08', stopTime: '2025-12-22', status: '停用', type: '长期' },
  { id: '5', name: '50% 葡萄糖注射液', dose: '20', unit: 'ml', doctor: '李医生', startTime: '2024-05-10 10:30', status: '在用', type: '临时' },
  { id: '6', name: '生理盐水', dose: '100', unit: 'ml', doctor: '王医生', startTime: '2024-05-01 09:00', stopTime: '2024-05-01 12:00', status: '停用', type: '临时' },
];

/** 医嘱单日期列表 Mock 数据 */
export const MOCK_ORDER_DATES = ['12-11', '09-01', '08-01', '07-15'];

/** 药品分组 Mock 数据 - 用于医嘱单 */
export const MOCK_MED_GROUPS: MedGroup[] = [
  {
    category: '改善贫血药物',
    items: [
      { drug: '人促红细胞生成素', dose: '10000 U', freq: '2 次/周' },
      { drug: '罗沙司他胶囊', dose: '120mg', freq: '3 次/周' },
      { drug: '蔗糖铁', dose: '', freq: '' },
      { drug: '其他：', dose: '', freq: '' }
    ]
  },
  {
    category: '磷结合剂',
    items: [
      { drug: '碳酸钙', dose: '', freq: '' },
      { drug: '碳酸镧咀嚼片 (餐中嚼服)', dose: '', freq: '' },
      { drug: '碳酸司维拉姆片', dose: '1.6g', freq: '3 次/天' },
      { drug: '其他：', dose: '', freq: '' }
    ]
  },
  {
    category: '拟钙剂',
    items: [
      { drug: '盐酸西那卡塞片', dose: '0mg', freq: '0 次/天' }
    ]
  },
  {
    category: '维生素 D',
    items: [
      { drug: '骨化三醇胶囊', dose: '', freq: '' },
      { drug: '帕立骨化醇注射液', dose: '', freq: '' },
      { drug: '其他：', dose: '', freq: '' }
    ]
  },
  {
    category: '降压药物',
    items: [
      { drug: '硝苯地平控释片', dose: '30mg', freq: '1 次/天' },
      { drug: '沙库巴曲缬沙坦', dose: '100mg', freq: '1 次/天' },
      { drug: '倍他乐克', dose: '47.5mg', freq: '1 次/天' }
    ]
  },
  { category: '降糖药物', items: [{ drug: '', dose: '', freq: '' }] },
  { category: '降脂药物', items: [{ drug: '', dose: '', freq: '' }] },
  { category: '降尿酸药物', items: [{ drug: '', dose: '', freq: '' }] },
  { category: '利尿剂', items: [{ drug: '呋塞米片', dose: '20mg', freq: '1 次/天' }] },
  { category: '其他用药', items: [{ drug: '', dose: '', freq: '' }] }
];
