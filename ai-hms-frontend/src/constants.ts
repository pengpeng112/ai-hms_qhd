import { UserRole } from './types/original'
import type { DashboardCardConfig } from './types/original'

const ADMIN_ROLE = 'ADMIN' as UserRole

export const DASHBOARD_CARDS: DashboardCardConfig[] = [
  { id: 'dept_overview', title: 'Department Overview', type: 'stat', size: 'large', roles: [ADMIN_ROLE, UserRole.DOCTOR_CHIEF, UserRole.NURSE_HEAD, UserRole.ENGINEER] },
  { id: 'active_patients', title: 'Active Patients', type: 'list', size: 'large', roles: [ADMIN_ROLE, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY] },
  { id: 'quality_stats', title: 'Quality Stats', type: 'chart', size: 'medium', roles: [ADMIN_ROLE, UserRole.DOCTOR_CHIEF, UserRole.NURSE_HEAD] },
  { id: 'prescription_adjust', title: 'Prescription Tasks', type: 'action', size: 'medium', roles: [ADMIN_ROLE, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY, UserRole.DOCTOR_CHIEF] },
  { id: 'duty_monitor', title: 'Duty Monitor', type: 'monitor', size: 'large', roles: [ADMIN_ROLE, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY, UserRole.DOCTOR_CHIEF] },
  { id: 'nurse_workload', title: 'Nurse Workload', type: 'chart', size: 'medium', roles: [ADMIN_ROLE, UserRole.NURSE_HEAD] },
  { id: 'staff_schedule', title: 'Staff Schedule', type: 'list', size: 'medium', roles: [ADMIN_ROLE, UserRole.NURSE_HEAD, UserRole.NURSE_SCHEDULER] },
  { id: 'schedule_adjust', title: 'Schedule Adjustment', type: 'action', size: 'medium', roles: [ADMIN_ROLE, UserRole.NURSE_HEAD, UserRole.NURSE_SCHEDULER] },
  { id: 'consumables_prep', title: 'Consumables Prep', type: 'inventory', size: 'large', roles: [ADMIN_ROLE, UserRole.NURSE_HEAD, UserRole.NURSE_MANAGER] },
  { id: 'my_duty_patients', title: 'My Duty Patients', type: 'list', size: 'large', roles: [ADMIN_ROLE, UserRole.NURSE_RESPONSIBLE] },
  { id: 'device_binding', title: 'Device Binding', type: 'binding', size: 'large', roles: [] },
  { id: 'device_status_eng', title: 'Device Status', type: 'monitor', size: 'large', roles: [] },
  { id: 'maintenance_logs', title: 'Maintenance Logs', type: 'list', size: 'medium', roles: [] },
]
