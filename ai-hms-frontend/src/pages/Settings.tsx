import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { logout } from '@/services/auth'
import { useTheme } from '@/contexts/useTheme'
import {
  getErrorMessage,
  restApi,
  type HdisIntegrationSettingsUpdatePayload,
  type HisOracleConnectionTestResult,
  type SyncJobConfig,
  type SyncJobRun,
  type SystemLogEntry,
  type SystemLogLevel,
  type SystemLogSource,
  type SystemLogsResponse,
} from '@/services/restClient'
import {
  User as UserIcon, Bell, Shield, Monitor,
  LogOut, Lock, Mail, Save, Smartphone, Check, Link2, RefreshCw, AlertCircle, FileText, Play, Pause, Database
} from 'lucide-react'

// Static Toggle component (定义在组件外部)
interface ToggleProps {
  checked: boolean
  onChange: () => void
  disabled?: boolean
}

const Toggle = ({ checked, onChange, disabled }: ToggleProps) => (
  <label className={`relative inline-flex items-center ${disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'}`}>
    <input type="checkbox" className="sr-only peer" checked={checked} onChange={onChange} disabled={disabled} />
    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
  </label>
)

type TabId = 'account' | 'notifications' | 'display' | 'security' | 'integration' | 'his-oracle' | 'logs'

type LogLevelFilter = '' | SystemLogLevel

const isMergedLogResponse = (value: SystemLogsResponse | null): value is Extract<SystemLogsResponse, { merged: SystemLogEntry[] }> => {
  return !!value && 'merged' in value
}

const getLogLevelBadgeClass = (level: SystemLogLevel) => {
  switch (level) {
    case 'ERROR':
      return 'bg-rose-900/60 text-rose-200'
    case 'WARN':
      return 'bg-amber-900/60 text-amber-200'
    case 'INFO':
      return 'bg-sky-900/60 text-sky-200'
    default:
      return 'bg-slate-800 text-slate-200'
  }
}

// Static TabButton component (定义在组件外部)
interface TabButtonProps {
  id: TabId
  label: string
  icon: React.ElementType
  activeTab: TabId
  onClick: (id: TabId) => void
}

const TabButton = ({ id, label, icon: Icon, activeTab, onClick }: TabButtonProps) => (
  <button
    onClick={() => onClick(id)}
    className={`w-full flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors mb-1 ${
      activeTab === id ? 'bg-blue-50 text-blue-700' : 'text-gray-600 hover:bg-gray-50'
    }`}
  >
    <Icon size={18} className="mr-3" />
    {label}
  </button>
)

export default function Settings() {
  const { t, i18n } = useTranslation(['settings', 'common'])
  const { theme, setTheme } = useTheme()
  const [activeTab, setActiveTab] = useState<TabId>('account')
  const [integrationLoaded, setIntegrationLoaded] = useState(false)
  const [integrationLoading, setIntegrationLoading] = useState(false)
  const [integrationSaving, setIntegrationSaving] = useState(false)
  const [integrationRefreshing, setIntegrationRefreshing] = useState(false)
  const [notifications, setNotifications] = useState({
    email: true,
    sms: false,
    app: true,
  })
  const [integrationForm, setIntegrationForm] = useState<HdisIntegrationSettingsUpdatePayload>({
    webcmdUrl: '',
    graphqlUrl: '',
    authUrl: '',
    clientId: '',
    serviceUsername: '',
    servicePassword: '',
    autoRefreshEnabled: true,
    refreshLeadSeconds: 1800,
  })
  const [integrationStatus, setIntegrationStatus] = useState({
    tokenExpiresAt: null as string | null,
    tokenStatus: 'MISSING' as 'MISSING' | 'UNKNOWN' | 'VALID' | 'EXPIRING' | 'EXPIRED',
    lastError: '',
    servicePasswordConfigured: false,
  })
  const [hisOracleForm, setHisOracleForm] = useState({
    host: '',
    port: 1521,
    service: 'orcl',
    username: '',
    password: '',
  })
  const [hisOracleTesting, setHisOracleTesting] = useState(false)
  const [hisOracleTestResult, setHisOracleTestResult] = useState<HisOracleConnectionTestResult | null>(null)
  const [hisOracleJobs, setHisOracleJobs] = useState<SyncJobConfig[]>([])
  const [hisOracleJobsLoaded, setHisOracleJobsLoaded] = useState(false)
  const [hisOracleRuns, setHisOracleRuns] = useState<Record<string, SyncJobRun[]>>({})
  const [hisOracleRunning, setHisOracleRunning] = useState<Record<string, boolean>>({})
  const [logsLoading, setLogsLoading] = useState(false)
  const [logsRefreshing, setLogsRefreshing] = useState(false)
  const [logsAutoRefresh, setLogsAutoRefresh] = useState(true)
  const [logsSource, setLogsSource] = useState<SystemLogSource>('all')
  const [logsLevel, setLogsLevel] = useState<LogLevelFilter>('')
  const [logsLines, setLogsLines] = useState(200)
  const [logsKeywordInput, setLogsKeywordInput] = useState('')
  const [logsKeyword, setLogsKeyword] = useState('')
  const [logsData, setLogsData] = useState<SystemLogsResponse | null>(null)
  const [logsError, setLogsError] = useState('')

  const toggleNotification = (key: keyof typeof notifications) => {
    setNotifications(prev => ({ ...prev, [key]: !prev[key] }))
  }

  const handleLogout = () => {
    logout()
  }

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng)
  }

  const loadIntegrationSettings = useCallback(async () => {
    setIntegrationLoading(true)
    try {
      const settings = await restApi.getHdisIntegrationSettings()
      setIntegrationForm(prev => ({
        ...prev,
        webcmdUrl: settings.webcmdUrl || '',
        graphqlUrl: settings.graphqlUrl || '',
        authUrl: settings.authUrl || '',
        clientId: settings.clientId || '',
        serviceUsername: settings.serviceUsername || '',
        servicePassword: '',
        autoRefreshEnabled: settings.autoRefreshEnabled,
        refreshLeadSeconds: settings.refreshLeadSeconds || 1800,
      }))
      setIntegrationStatus({
        tokenExpiresAt: settings.tokenExpiresAt,
        tokenStatus: settings.tokenStatus,
        lastError: settings.lastError || '',
        servicePasswordConfigured: settings.servicePasswordConfigured,
      })
      setIntegrationLoaded(true)
    } catch (error) {
      console.error('加载 HDIS 设置失败:', error)
      alert(t('settings:integration.loadFailed'))
    } finally {
      setIntegrationLoading(false)
    }
  }, [t])

  useEffect(() => {
    if (activeTab === 'integration' && !integrationLoaded) {
      void loadIntegrationSettings()
    }
  }, [activeTab, integrationLoaded, loadIntegrationSettings])

  const saveIntegrationSettings = async () => {
    setIntegrationSaving(true)
    try {
      const payload: HdisIntegrationSettingsUpdatePayload = {
        ...integrationForm,
        servicePassword: integrationForm.servicePassword?.trim() || undefined,
      }
      const saved = await restApi.updateHdisIntegrationSettings(payload)
      setIntegrationForm(prev => ({
        ...prev,
        webcmdUrl: saved.webcmdUrl || '',
        graphqlUrl: saved.graphqlUrl || '',
        authUrl: saved.authUrl || '',
        clientId: saved.clientId || '',
        serviceUsername: saved.serviceUsername || '',
        servicePassword: '',
        autoRefreshEnabled: saved.autoRefreshEnabled,
        refreshLeadSeconds: saved.refreshLeadSeconds || 1800,
      }))
      setIntegrationStatus({
        tokenExpiresAt: saved.tokenExpiresAt,
        tokenStatus: saved.tokenStatus,
        lastError: saved.lastError || '',
        servicePasswordConfigured: saved.servicePasswordConfigured,
      })
      alert(t('settings:integration.saveSuccess'))
    } catch (error) {
      console.error('保存 HDIS 设置失败:', error)
      alert(t('settings:integration.saveFailed'))
    } finally {
      setIntegrationSaving(false)
    }
  }

  const refreshIntegrationToken = async () => {
    setIntegrationRefreshing(true)
    try {
      await restApi.refreshHdisToken()
      await loadIntegrationSettings()
      alert(t('settings:integration.refreshSuccess'))
    } catch (error) {
      console.error('刷新 HDIS Token 失败:', error)
      alert(t('settings:integration.refreshFailed'))
    } finally {
      setIntegrationRefreshing(false)
    }
  }

  const testHisOracleConnection = async () => {
    setHisOracleTesting(true)
    setHisOracleTestResult(null)
    try {
      const result = await restApi.testHisOracleConnection({
        host: hisOracleForm.host,
        port: hisOracleForm.port,
        service: hisOracleForm.service,
        username: hisOracleForm.username,
        password: hisOracleForm.password,
      })
      setHisOracleTestResult(result)
    } catch (error) {
      setHisOracleTestResult({ connected: false, error: getErrorMessage(error) })
    } finally {
      setHisOracleTesting(false)
    }
  }

  const loadHisOracleJobs = useCallback(async () => {
    try {
      const jobs = await restApi.getSyncJobs()
      setHisOracleJobs(jobs)
      setHisOracleJobsLoaded(true)
    } catch {
      setHisOracleJobs([])
      setHisOracleJobsLoaded(true)
    }
  }, [])

  const initHisOracleJobs = async () => {
    try {
      await restApi.seedSyncJobs()
      alert(t('settings:hisOracle.initJobsSuccess'))
      await loadHisOracleJobs()
    } catch {
      alert(t('settings:hisOracle.saveFailed'))
    }
  }

  const handleRunJob = async (code: string) => {
    setHisOracleRunning(prev => ({ ...prev, [code]: true }))
    try {
      await restApi.runSyncJob(code)
      alert(`${code} 同步已启动`)
      setTimeout(() => {
        handleLoadRuns(code)
        loadHisOracleJobs()
      }, 2000)
    } catch (e) {
      alert(getErrorMessage(e))
    } finally {
      setHisOracleRunning(prev => ({ ...prev, [code]: false }))
    }
  }

  const handleLoadRuns = async (code: string) => {
    try {
      const runs = await restApi.getSyncJobRuns(code)
      setHisOracleRuns(prev => ({ ...prev, [code]: runs }))
    } catch {
      // ignore
    }
  }

  useEffect(() => {
    if (activeTab === 'his-oracle' && !hisOracleJobsLoaded) {
      void loadHisOracleJobs()
    }
  }, [activeTab, hisOracleJobsLoaded, loadHisOracleJobs])

  const loadSystemLogs = useCallback(async (silent = false) => {
    if (silent) {
      setLogsRefreshing(true)
    } else {
      setLogsLoading(true)
    }
    setLogsError('')
    try {
      const data = await restApi.getSystemLogs({
        source: logsSource,
        lines: logsLines,
        keyword: logsKeyword.trim() || undefined,
        level: logsLevel || undefined,
      })
      setLogsData(data)
    } catch (error) {
      console.error('加载系统日志失败:', error)
      setLogsError(getErrorMessage(error))
    } finally {
      if (silent) {
        setLogsRefreshing(false)
      } else {
        setLogsLoading(false)
      }
    }
  }, [logsSource, logsLines, logsKeyword, logsLevel])

  useEffect(() => {
    if (activeTab !== 'logs') {
      return
    }
    void loadSystemLogs(false)
  }, [activeTab, loadSystemLogs])

  useEffect(() => {
    if (activeTab !== 'logs' || !logsAutoRefresh) {
      return
    }

    const timer = window.setInterval(() => {
      if (document.visibilityState !== 'visible') {
        return
      }
      void loadSystemLogs(true)
    }, 10_000)

    return () => {
      window.clearInterval(timer)
    }
  }, [activeTab, logsAutoRefresh, loadSystemLogs])

  const tokenStatusText = (() => {
    switch (integrationStatus.tokenStatus) {
      case 'VALID':
        return t('settings:integration.status.valid')
      case 'EXPIRING':
        return t('settings:integration.status.expiring')
      case 'EXPIRED':
        return t('settings:integration.status.expired')
      case 'UNKNOWN':
        return t('settings:integration.status.unknown')
      default:
        return t('settings:integration.status.missing')
    }
  })()

  const displayedLogEntries = (() => {
    if (!logsData) {
      return [] as SystemLogEntry[]
    }
    if (isMergedLogResponse(logsData)) {
      return logsData.merged
    }
    return logsData.entries
  })()

  return (
    <div className="h-full flex flex-col max-w-[1200px] mx-auto">
      <h2 className="text-2xl font-bold text-gray-800 mb-6">{t('settings:title')}</h2>

      <div className="flex flex-col md:flex-row gap-8 flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="w-full md:w-64 shrink-0">
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-2">
            <TabButton id="account" label={t('settings:tab.account')} icon={UserIcon} activeTab={activeTab} onClick={setActiveTab} />
            <TabButton id="notifications" label={t('settings:tab.notifications')} icon={Bell} activeTab={activeTab} onClick={setActiveTab} />
            <TabButton id="display" label={t('settings:tab.display')} icon={Monitor} activeTab={activeTab} onClick={setActiveTab} />
            <TabButton id="security" label={t('settings:tab.security')} icon={Shield} activeTab={activeTab} onClick={setActiveTab} />
            <TabButton id="integration" label={t('settings:tab.integration')} icon={Link2} activeTab={activeTab} onClick={setActiveTab} />
            <TabButton id="his-oracle" label={t('settings:tab.hisOracle')} icon={Database} activeTab={activeTab} onClick={setActiveTab} />
            <TabButton id="logs" label={t('settings:tab.logs')} icon={FileText} activeTab={activeTab} onClick={setActiveTab} />
            <div className="h-px bg-gray-100 my-2"></div>
            <button
              onClick={handleLogout}
              className="w-full flex items-center px-4 py-3 text-sm font-medium rounded-lg text-red-600 hover:bg-red-50 transition-colors"
            >
              <LogOut size={18} className="mr-3" />
              {t('settings:action.logout')}
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 bg-white rounded-xl shadow-sm border border-gray-200 p-8 overflow-y-auto">
          {activeTab === 'account' && (
            <div className="max-w-xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">{t('settings:account.title')}</h3>
              <div className="mb-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-700">
                个人资料保存接口暂未开放，当前页面仅展示配置入口。
              </div>
              <div className="space-y-6">
                <div className="flex items-center gap-6">
                  <div className="w-20 h-20 rounded-full bg-gray-200 flex items-center justify-center text-gray-400 border-2 border-white shadow-sm overflow-hidden relative">
                    <UserIcon size={28} className="text-gray-500" />
                  </div>
                  <div>
                    <button disabled className="px-4 py-2 cursor-not-allowed bg-gray-50 border border-gray-200 rounded-lg text-sm font-medium text-gray-400">
                      {t('settings:account.avatar')}
                    </button>
                    <p className="text-xs text-gray-500 mt-2">{t('settings:account.avatarHint')}</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:account.name')}</label>
                    <input
                      type="text"
                      placeholder="暂未接入个人资料接口"
                      disabled
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500 text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:account.employeeId')}</label>
                    <input
                      type="text"
                      placeholder="暂未接入"
                      disabled
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500 text-sm"
                    />
                  </div>
                  <div className="col-span-2">
                    <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:account.email')}</label>
                    <input
                      type="email"
                      placeholder="暂未接入个人资料接口"
                      disabled
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500 text-sm"
                    />
                  </div>
                  <div className="col-span-2">
                    <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:account.signature')}</label>
                    <textarea
                      rows={3}
                      disabled
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500 text-sm"
                      placeholder="签名维护暂未开放"
                    ></textarea>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'notifications' && (
            <div className="max-w-xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">{t('settings:notifications.title')}</h3>
              <div className="mb-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-700">
                通知偏好保存接口暂未开放，开关已临时禁用。
              </div>
              <div className="space-y-4">
                <div className="flex items-center justify-between py-3 border-b border-gray-50">
                  <div className="flex items-start">
                    <Mail className="text-gray-400 mt-1 mr-3" size={20} />
                    <div>
                      <p className="font-medium text-gray-800">{t('settings:notifications.email')}</p>
                      <p className="text-sm text-gray-500">{t('settings:notifications.emailDesc')}</p>
                    </div>
                  </div>
                  <Toggle checked={notifications.email} onChange={() => toggleNotification('email')} disabled />
                </div>

                <div className="flex items-center justify-between py-3 border-b border-gray-50">
                  <div className="flex items-start">
                    <Smartphone className="text-gray-400 mt-1 mr-3" size={20} />
                    <div>
                      <p className="font-medium text-gray-800">{t('settings:notifications.sms')}</p>
                      <p className="text-sm text-gray-500">{t('settings:notifications.smsDesc')}</p>
                    </div>
                  </div>
                  <Toggle checked={notifications.sms} onChange={() => toggleNotification('sms')} disabled />
                </div>

                <div className="flex items-center justify-between py-3 border-b border-gray-50">
                  <div className="flex items-start">
                    <Bell className="text-gray-400 mt-1 mr-3" size={20} />
                    <div>
                      <p className="font-medium text-gray-800">{t('settings:notifications.app')}</p>
                      <p className="text-sm text-gray-500">{t('settings:notifications.appDesc')}</p>
                    </div>
                  </div>
                  <Toggle checked={notifications.app} onChange={() => toggleNotification('app')} disabled />
                </div>
              </div>
            </div>
          )}

          {activeTab === 'display' && (
            <div className="max-w-xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">{t('settings:display.title')}</h3>
              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-3">{t('settings:display.theme')}</label>
                  <div className="grid grid-cols-2 gap-4">
                    <button
                      onClick={() => setTheme('light')}
                      className={`relative rounded-lg p-3 bg-white text-left transition-all ${
                        theme === 'light'
                          ? 'border-2 border-blue-500 shadow-md'
                          : 'border border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                      }`}
                    >
                      {theme === 'light' && (
                        <div className="absolute top-2 right-2 w-5 h-5 bg-blue-500 rounded-full flex items-center justify-center">
                          <Check size={12} className="text-white" />
                        </div>
                      )}
                      <div className="w-full h-8 bg-gray-100 mb-2 rounded border border-gray-200"></div>
                      <span className={`text-sm font-medium ${theme === 'light' ? 'text-blue-600' : 'text-gray-600'}`}>
                        {t('settings:theme.light')}
                      </span>
                    </button>
                    <button
                      onClick={() => setTheme('high-contrast')}
                      className={`relative rounded-lg p-3 bg-blue-50 text-left transition-all ${
                        theme === 'high-contrast'
                          ? 'border-2 border-blue-500 shadow-md'
                          : 'border border-blue-100 hover:bg-blue-100'
                      }`}
                    >
                      {theme === 'high-contrast' && (
                        <div className="absolute top-2 right-2 w-5 h-5 bg-blue-500 rounded-full flex items-center justify-center">
                          <Check size={12} className="text-white" />
                        </div>
                      )}
                      <div className="w-full h-8 bg-white mb-2 rounded border border-blue-200"></div>
                      <span className={`text-sm font-medium ${theme === 'high-contrast' ? 'text-blue-600' : 'text-gray-800'}`}>
                        {t('settings:theme.highContrast')}
                      </span>
                    </button>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings:display.language')}</label>
                  <select
                    value={i18n.language}
                    onChange={(e) => changeLanguage(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="zh-CN">{t('settings:language.zhCN')}</option>
                    <option value="en-US">{t('settings:language.enUS')}</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'security' && (
            <div className="max-w-xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">{t('settings:security.title')}</h3>
              <div className="space-y-6">
                <div className="p-4 bg-yellow-50 rounded-lg border border-yellow-100">
                  <div className="flex items-start">
                    <Lock className="text-yellow-600 mt-1 mr-3" size={20} />
                    <div>
                      <p className="font-bold text-yellow-800">{t('settings:security.changePassword')}</p>
                      <p className="text-sm text-yellow-700 mt-1">{t('settings:security.passwordHint')}</p>
                      <button disabled className="mt-3 cursor-not-allowed px-4 py-2 bg-white border border-yellow-100 text-yellow-300 rounded-lg text-sm font-medium">
                        修改密码暂未开放
                      </button>
                    </div>
                  </div>
                </div>
                <div>
                  <h4 className="font-medium text-gray-800 mb-2">{t('settings:security.loginActivity')}</h4>
                  <div className="text-sm text-gray-600 bg-gray-50 p-3 rounded-lg border border-gray-200">
                    <p>登录活动审计暂未接入，暂不展示模拟 IP、设备和时间。</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'logs' && (
            <div className="max-w-5xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">
                {t('settings:logs.title')}
              </h3>

              <div className="space-y-4">
                <div className="p-4 border border-gray-200 rounded-lg bg-gray-50/60">
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:logs.source')}</label>
                      <select
                        value={logsSource}
                        onChange={(e) => setLogsSource(e.target.value as SystemLogSource)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      >
                        <option value="all">{t('settings:logs.sourceAll')}</option>
                        <option value="app">{t('settings:logs.sourceApp')}</option>
                        <option value="error">{t('settings:logs.sourceError')}</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:logs.lines')}</label>
                      <select
                        value={logsLines}
                        onChange={(e) => setLogsLines(Number(e.target.value))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      >
                        <option value={100}>100</option>
                        <option value={200}>200</option>
                        <option value={500}>500</option>
                        <option value={1000}>1000</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:logs.level')}</label>
                      <select
                        value={logsLevel}
                        onChange={(e) => setLogsLevel(e.target.value as LogLevelFilter)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      >
                        <option value="">{t('settings:logs.levelAll')}</option>
                        <option value="INFO">INFO</option>
                        <option value="WARN">WARN</option>
                        <option value="ERROR">ERROR</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:logs.keyword')}</label>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          value={logsKeywordInput}
                          onChange={(e) => setLogsKeywordInput(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                              setLogsKeyword(logsKeywordInput.trim())
                            }
                          }}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                        <button
                          type="button"
                          onClick={() => setLogsKeyword(logsKeywordInput.trim())}
                          className="px-3 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-100"
                        >
                          {t('settings:logs.apply')}
                        </button>
                      </div>
                    </div>
                  </div>

                  <div className="flex justify-end gap-3 mt-4">
                    <button
                      type="button"
                      onClick={() => setLogsAutoRefresh(prev => !prev)}
                      className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 inline-flex items-center gap-2"
                    >
                      {logsAutoRefresh ? <Pause size={16} /> : <Play size={16} />}
                      {logsAutoRefresh ? t('settings:logs.autoRefreshOn') : t('settings:logs.autoRefreshOff')}
                    </button>
                    <button
                      type="button"
                      onClick={() => void loadSystemLogs(true)}
                      disabled={logsRefreshing}
                      className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 disabled:opacity-60 inline-flex items-center gap-2"
                    >
                      <RefreshCw size={16} className={logsRefreshing ? 'animate-spin' : ''} />
                      {t('settings:logs.refresh')}
                    </button>
                  </div>
                </div>

                <div className="text-xs text-gray-500">
                  {t('settings:logs.redactedHint')}
                  {logsData?.meta?.fetchedAt ? ` · ${t('settings:logs.lastFetchedAt')}: ${logsData.meta.fetchedAt}` : ''}
                </div>

                {logsError && (
                  <div className="p-3 border border-red-200 rounded-lg bg-red-50 text-red-700 text-sm">
                    {t('settings:logs.loadFailed')}: {logsError}
                  </div>
                )}

                <div className="border border-gray-200 rounded-lg overflow-hidden">
                  <div className="max-h-[560px] overflow-auto bg-slate-950">
                    {logsLoading ? (
                      <div className="text-sm text-slate-300 p-4">{t('settings:logs.loading')}</div>
                    ) : displayedLogEntries.length === 0 ? (
                      <div className="text-sm text-slate-400 p-4">{t('settings:logs.empty')}</div>
                    ) : (
                      <div className="divide-y divide-slate-800">
                        {displayedLogEntries.map((entry, index) => (
                          <div key={`${entry.source}-${entry.timestamp || 'na'}-${index}`} className="px-4 py-3 font-mono text-xs text-slate-100 whitespace-pre-wrap break-words">
                            <div className="flex flex-wrap items-center gap-2 text-[10px] text-slate-400 mb-1">
                              <span>{entry.timestamp || '-'}</span>
                              <span className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-200">{entry.source}</span>
                              {entry.level && (
                                <span className={`px-1.5 py-0.5 rounded ${getLogLevelBadgeClass(entry.level)}`}>{entry.level}</span>
                              )}
                            </div>
                            <div>{entry.raw}</div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'integration' && (
            <div className="max-w-3xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">
                {t('settings:integration.title')}
              </h3>

              {integrationLoading ? (
                <div className="text-sm text-gray-500 py-6">{t('settings:integration.loading')}</div>
              ) : (
                <div className="space-y-6">
                  <div className="p-4 border border-gray-200 rounded-lg bg-gray-50/60">
                    <h4 className="font-semibold text-gray-800 mb-4">{t('settings:integration.configTitle')}</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.webcmdUrl')}</label>
                        <input
                          type="text"
                          value={integrationForm.webcmdUrl}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, webcmdUrl: e.target.value }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.graphqlUrl')}</label>
                        <input
                          type="text"
                          value={integrationForm.graphqlUrl}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, graphqlUrl: e.target.value }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.authUrl')}</label>
                        <input
                          type="text"
                          value={integrationForm.authUrl}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, authUrl: e.target.value }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.clientId')}</label>
                        <input
                          type="text"
                          value={integrationForm.clientId}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, clientId: e.target.value }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.serviceUsername')}</label>
                        <input
                          type="text"
                          value={integrationForm.serviceUsername}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, serviceUsername: e.target.value }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.servicePassword')}</label>
                        <input
                          type="password"
                          value={integrationForm.servicePassword || ''}
                          placeholder={integrationStatus.servicePasswordConfigured ? t('settings:integration.passwordHintConfigured') : t('settings:integration.passwordHintEmpty')}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, servicePassword: e.target.value }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                      <div className="flex items-center justify-between border border-gray-200 rounded-lg px-3 py-2">
                        <span className="text-sm text-gray-700">{t('settings:integration.autoRefreshEnabled')}</span>
                        <Toggle
                          checked={integrationForm.autoRefreshEnabled}
                          onChange={() => setIntegrationForm(prev => ({ ...prev, autoRefreshEnabled: !prev.autoRefreshEnabled }))}
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:integration.refreshLeadSeconds')}</label>
                        <input
                          type="number"
                          min={60}
                          max={86400}
                          value={integrationForm.refreshLeadSeconds}
                          onChange={(e) => setIntegrationForm(prev => ({ ...prev, refreshLeadSeconds: Number(e.target.value || 1800) }))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                        />
                      </div>
                    </div>
                  </div>

                  <div className="p-4 border border-gray-200 rounded-lg">
                    <h4 className="font-semibold text-gray-800 mb-3">{t('settings:integration.tokenStatusTitle')}</h4>
                    <div className="space-y-2 text-sm">
                      <p className="text-gray-700">
                        {t('settings:integration.tokenStatus')}: <span className="font-semibold">{tokenStatusText}</span>
                      </p>
                      <p className="text-gray-700">
                        {t('settings:integration.tokenExpiresAt')}: <span className="font-mono">{integrationStatus.tokenExpiresAt || '-'}</span>
                      </p>
                      {integrationStatus.lastError && (
                        <p className="text-red-600 flex items-start gap-2">
                          <AlertCircle size={16} className="mt-0.5 shrink-0" />
                          <span>{integrationStatus.lastError}</span>
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="flex justify-end gap-3">
                    <button
                      type="button"
                      onClick={() => void refreshIntegrationToken()}
                      disabled={integrationRefreshing}
                      className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 disabled:opacity-60 inline-flex items-center gap-2"
                    >
                      <RefreshCw size={16} className={integrationRefreshing ? 'animate-spin' : ''} />
                      {t('settings:integration.actionRefreshToken')}
                    </button>
                    <button
                      type="button"
                      onClick={() => void saveIntegrationSettings()}
                      disabled={integrationSaving}
                      className="px-5 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-60 inline-flex items-center gap-2"
                    >
                      <Save size={16} />
                      {t('settings:integration.actionSave')}
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'his-oracle' && (
            <div className="max-w-3xl">
              <h3 className="text-lg font-bold text-gray-800 mb-6 pb-2 border-b border-gray-100">
                {t('settings:hisOracle.title')}
              </h3>

              <div className="space-y-6">
                <div className="p-4 border border-gray-200 rounded-lg bg-gray-50/60">
                  <h4 className="font-semibold text-gray-800 mb-4">{t('settings:hisOracle.connectionTitle')}</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:hisOracle.host')}</label>
                      <input
                        type="text"
                        value={hisOracleForm.host}
                        onChange={(e) => setHisOracleForm(prev => ({ ...prev, host: e.target.value }))}
                        placeholder="10.10.8.216"
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:hisOracle.port')}</label>
                      <input
                        type="number"
                        value={hisOracleForm.port}
                        onChange={(e) => setHisOracleForm(prev => ({ ...prev, port: Number(e.target.value || 1521) }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      />
                    </div>
                    <div className="md:col-span-2">
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:hisOracle.service')}</label>
                      <input
                        type="text"
                        value={hisOracleForm.service}
                        onChange={(e) => setHisOracleForm(prev => ({ ...prev, service: e.target.value }))}
                        placeholder="orcl"
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:hisOracle.username')}</label>
                      <input
                        type="text"
                        value={hisOracleForm.username}
                        onChange={(e) => setHisOracleForm(prev => ({ ...prev, username: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">{t('settings:hisOracle.password')}</label>
                      <input
                        type="password"
                        value={hisOracleForm.password}
                        onChange={(e) => setHisOracleForm(prev => ({ ...prev, password: e.target.value }))}
                        placeholder={t('settings:hisOracle.passwordHint')}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm"
                      />
                    </div>
                  </div>

                  <div className="flex items-center gap-3 mt-4">
                    <button
                      type="button"
                      onClick={() => void testHisOracleConnection()}
                      disabled={hisOracleTesting}
                      className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 disabled:opacity-60 inline-flex items-center gap-2 text-sm"
                    >
                      <Link2 size={16} className={hisOracleTesting ? 'animate-spin' : ''} />
                      {hisOracleTesting ? t('settings:hisOracle.testing') : t('settings:hisOracle.testConnection')}
                    </button>
                  </div>

                  {hisOracleTestResult && (
                    <div className={`mt-3 p-3 rounded-lg text-sm ${hisOracleTestResult.connected ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                      {hisOracleTestResult.connected
                        ? `${t('settings:hisOracle.connected')} · ${t('settings:hisOracle.latency')}: ${hisOracleTestResult.latency_ms}ms`
                        : `${t('settings:hisOracle.connectFailed')}: ${hisOracleTestResult.error}`}
                    </div>
                  )}
                </div>

                <div className="p-4 border border-gray-200 rounded-lg">
                  <div className="flex items-center justify-between mb-4">
                    <h4 className="font-semibold text-gray-800">{t('settings:hisOracle.syncJobsTitle')}</h4>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => void initHisOracleJobs()}
                        className="px-3 py-1.5 bg-white border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 text-xs"
                      >
                        {t('settings:hisOracle.actionInitJobs')}
                      </button>
                    </div>
                  </div>
                  {hisOracleJobs.length === 0 ? (
                    <div className="text-sm text-gray-500 py-4">
                      {hisOracleJobsLoaded ? '暂无同步任务' : '加载中...'}
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {hisOracleJobs.map((job) => (
                        <div key={job.jobCode} className="flex items-center justify-between p-3 border border-gray-200 rounded-lg">
                          <div>
                            <div className="text-sm font-medium text-gray-800">{job.jobCode}</div>
                            <div className="text-xs text-gray-500 mt-0.5">
                              {job.enabled ? t('settings:hisOracle.jobEnabled') : t('settings:hisOracle.jobDisabled')}
                              {job.lastRunAt ? ` · ${t('settings:hisOracle.jobLastRun')}: ${new Date(job.lastRunAt).toLocaleString()}` : ` · ${t('settings:hisOracle.jobNeverRun')}`}
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className={`px-2 py-1 rounded-full text-xs font-medium ${job.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                              {job.enabled ? t('settings:hisOracle.jobEnabled') : t('settings:hisOracle.jobDisabled')}
                            </span>
                            <button
                              type="button"
                              disabled={!job.enabled || hisOracleRunning[job.jobCode]}
                              onClick={() => void handleRunJob(job.jobCode)}
                              className={`px-3 py-1 rounded-lg text-xs font-medium inline-flex items-center gap-1 ${
                                !job.enabled ? 'bg-gray-100 text-gray-400 cursor-not-allowed' :
                                'bg-blue-50 text-blue-600 hover:bg-blue-100'
                              }`}
                            >
                              <Play size={12} className={hisOracleRunning[job.jobCode] ? 'animate-spin' : ''} />
                              运行
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="p-4 border border-gray-200 rounded-lg">
                  <h4 className="font-semibold text-gray-800 mb-3">运行历史</h4>
                  {hisOracleJobs.map((job) => {
                    const runs = hisOracleRuns[job.jobCode]
                    if (!runs || runs.length === 0) return null
                    return (
                      <div key={job.jobCode} className="mb-4 last:mb-0">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm font-medium text-gray-700">{job.jobCode}</span>
                          <button
                            type="button"
                            onClick={() => void handleLoadRuns(job.jobCode)}
                            className="text-xs text-blue-600 hover:text-blue-800"
                          >
                            刷新
                          </button>
                        </div>
                        <div className="overflow-x-auto">
                          <table className="w-full text-xs">
                            <thead>
                              <tr className="text-left text-gray-500 border-b border-gray-100">
                                <th className="pb-1 font-medium">时间</th>
                                <th className="pb-1 font-medium">状态</th>
                                <th className="pb-1 font-medium">耗时</th>
                                <th className="pb-1 font-medium">获取</th>
                                <th className="pb-1 font-medium">新增</th>
                                <th className="pb-1 font-medium">失败</th>
                              </tr>
                            </thead>
                            <tbody>
                              {runs.slice(0, 5).map((run) => (
                                <tr key={run.id} className="border-b border-gray-50">
                                  <td className="py-1 text-gray-600">{new Date(run.startedAt).toLocaleString()}</td>
                                  <td className="py-1">
                                    <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${
                                      run.status === 'success' ? 'bg-green-100 text-green-600' :
                                      run.status === 'partial' ? 'bg-amber-100 text-amber-600' :
                                      run.status === 'failed' ? 'bg-red-100 text-red-600' :
                                      'bg-blue-100 text-blue-600'
                                    }`}>{run.status}</span>
                                  </td>
                                  <td className="py-1 text-gray-500">{run.durationMs ? `${(run.durationMs / 1000).toFixed(1)}s` : '-'}</td>
                                  <td className="py-1 text-gray-600">{run.fetchedCount}</td>
                                  <td className="py-1 text-gray-600">{run.createdCount}</td>
                                  <td className="py-1 text-gray-600">{run.failedCount}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    )
                  })}
                  {hisOracleJobs.every((j) => !hisOracleRuns[j.jobCode] || hisOracleRuns[j.jobCode].length === 0) && (
                    <div className="text-sm text-gray-400 py-2">暂无运行记录，点击「运行」启动同步</div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Sticky Footer */}
          {activeTab === 'display' && (
            <div className="mt-10 pt-6 border-t border-gray-100 text-right text-sm font-semibold text-gray-400">
              显示设置会立即应用，无需保存。
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
