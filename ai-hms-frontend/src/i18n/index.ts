import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

// 导入中文翻译文件
import commonZh from './locales/zh-CN/common.json'
import navZh from './locales/zh-CN/nav.json'
import patientZh from './locales/zh-CN/patient.json'
import dialysisZh from './locales/zh-CN/dialysis.json'
import deviceZh from './locales/zh-CN/device.json'
import formZh from './locales/zh-CN/form.json'
import settingsZh from './locales/zh-CN/settings.json'
import roleZh from './locales/zh-CN/role.json'
import dashboardZh from './locales/zh-CN/dashboard.json'
import monitoringZh from './locales/zh-CN/monitoring.json'
import statisticsZh from './locales/zh-CN/statistics.json'
import wardOverviewZh from './locales/zh-CN/wardOverview.json'
import inventoryZh from './locales/zh-CN/inventory.json'
import masterDataZh from './locales/zh-CN/masterData.json'
import dialysisProcessingZh from './locales/zh-CN/dialysisProcessing.json'
import treatmentConfigZh from './locales/zh-CN/treatmentConfig.json'

// 导入英文翻译文件
import commonEn from './locales/en-US/common.json'
import navEn from './locales/en-US/nav.json'
import patientEn from './locales/en-US/patient.json'
import dialysisEn from './locales/en-US/dialysis.json'
import deviceEn from './locales/en-US/device.json'
import formEn from './locales/en-US/form.json'
import settingsEn from './locales/en-US/settings.json'
import roleEn from './locales/en-US/role.json'
import dashboardEn from './locales/en-US/dashboard.json'
import monitoringEn from './locales/en-US/monitoring.json'
import statisticsEn from './locales/en-US/statistics.json'
import wardOverviewEn from './locales/en-US/wardOverview.json'
import inventoryEn from './locales/en-US/inventory.json'
import masterDataEn from './locales/en-US/masterData.json'
import dialysisProcessingEn from './locales/en-US/dialysisProcessing.json'
import treatmentConfigEn from './locales/en-US/treatmentConfig.json'

// 导入类型定义
import './types'

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      'zh-CN': {
        common: commonZh,
        nav: navZh,
        patient: patientZh,
        dialysis: dialysisZh,
        device: deviceZh,
        form: formZh,
        settings: settingsZh,
        role: roleZh,
        dashboard: dashboardZh,
        monitoring: monitoringZh,
        statistics: statisticsZh,
        wardOverview: wardOverviewZh,
        inventory: inventoryZh,
        masterData: masterDataZh,
        dialysisProcessing: dialysisProcessingZh,
        treatmentConfig: treatmentConfigZh,
      },
      'en-US': {
        common: commonEn,
        nav: navEn,
        patient: patientEn,
        dialysis: dialysisEn,
        device: deviceEn,
        form: formEn,
        settings: settingsEn,
        role: roleEn,
        dashboard: dashboardEn,
        monitoring: monitoringEn,
        statistics: statisticsEn,
        wardOverview: wardOverviewEn,
        inventory: inventoryEn,
        masterData: masterDataEn,
        dialysisProcessing: dialysisProcessingEn,
        treatmentConfig: treatmentConfigEn,
      },
    },
    defaultNS: 'common',
    fallbackLng: 'zh-CN',
    supportedLngs: ['zh-CN', 'en-US'],
    interpolation: {
      escapeValue: false, // React 已处理 XSS
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: 'i18n_language',
    },
    // 开发环境下输出缺失的翻译
    saveMissing: import.meta.env.DEV,
    missingKeyHandler: (lng, ns, key) => {
      if (import.meta.env.DEV) {
        console.warn(`Missing translation: [${lng}] ${ns}:${key}`)
      }
    },
  })

export default i18n
