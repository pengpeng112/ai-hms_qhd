import { UserRole } from './types/original'
import type { DashboardCardConfig } from './types/original'

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
]
