import 'i18next'

// 导入所有命名空间的 JSON 类型
import common from './locales/zh-CN/common.json'
import nav from './locales/zh-CN/nav.json'
import patient from './locales/zh-CN/patient.json'
import dialysis from './locales/zh-CN/dialysis.json'
import schedule from './locales/zh-CN/schedule.json'
import device from './locales/zh-CN/device.json'
import form from './locales/zh-CN/form.json'
import settings from './locales/zh-CN/settings.json'
import role from './locales/zh-CN/role.json'
import dashboard from './locales/zh-CN/dashboard.json'
import monitoring from './locales/zh-CN/monitoring.json'
import statistics from './locales/zh-CN/statistics.json'
import wardOverview from './locales/zh-CN/wardOverview.json'
import inventory from './locales/zh-CN/inventory.json'
import masterData from './locales/zh-CN/masterData.json'
import dialysisProcessing from './locales/zh-CN/dialysisProcessing.json'
import treatmentConfig from './locales/zh-CN/treatmentConfig.json'

declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common'
    resources: {
      common: typeof common
      nav: typeof nav
      patient: typeof patient
      dialysis: typeof dialysis
      schedule: typeof schedule
      device: typeof device
      form: typeof form
      settings: typeof settings
      role: typeof role
      dashboard: typeof dashboard
      monitoring: typeof monitoring
      statistics: typeof statistics
      wardOverview: typeof wardOverview
      inventory: typeof inventory
      masterData: typeof masterData
      dialysisProcessing: typeof dialysisProcessing
      treatmentConfig: typeof treatmentConfig
    }
  }
}
