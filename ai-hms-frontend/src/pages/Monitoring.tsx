import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { MonitorDevice } from '../types/original'
import { restApi, type RestPatientOrder } from '../services/restClient'
import { useMonitoringData } from './monitoring/hooks/useMonitoringData'
import { useDeviceFilter } from './monitoring/hooks/useDeviceFilter'
import { useModalManager } from './monitoring/hooks/useModalManager'
import StatusGrid from './monitoring/StatusGrid'
import { cachedHistoryData } from './monitoring/types'
import {
  Monitor, Search, X, Activity, TrendingUp,
  ClipboardList, Trash2, Plus, ChevronDown, FileEdit, Edit3, BarChart3, Sparkles
} from 'lucide-react'
import {
  LineChart, Line, ResponsiveContainer, XAxis, YAxis, Tooltip,
  AreaChart, Area, CartesianGrid
} from 'recharts'

// --- 通用弹窗包装 ---
const ModalOverlay = ({
  children,
  onClose,
  maxWidth = 'max-w-5xl'
}: {
  children?: React.ReactNode
  onClose: () => void
  maxWidth?: string
}) => (
  <div
    className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
    onClick={onClose}
  >
    <div
      className={`bg-white rounded-lg shadow-2xl w-full ${maxWidth} max-h-[95vh] overflow-hidden m-4 flex flex-col`}
      onClick={(e) => e.stopPropagation()}
    >
      {children}
    </div>
  </div>
)

const ModalHeader = ({
  title,
  onClose,
  icon: Icon,
  saveButtonText
}: {
  title: string
  onClose: () => void
  icon: React.ElementType
  saveButtonText: string
}) => (
  <div className="flex justify-between items-center px-6 py-4 border-b border-gray-100 bg-gray-50 shrink-0">
    <h3 className="text-lg font-bold text-gray-800 flex items-center">
      <Icon className="mr-2 text-blue-600" size={20} /> {title}
    </h3>
    <div className="flex items-center gap-2">
      <button disabled className="cursor-not-allowed px-4 py-1.5 bg-slate-200 text-slate-500 rounded text-sm font-bold shadow-sm" title="该弹窗保存接口暂未开放">
        {saveButtonText}（未开放）
      </button>
      <button
        onClick={onClose}
        className="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-200 rounded-full transition-colors"
      >
        <X size={20} />
      </button>
    </div>
  </div>
)

const PrescriptionInput = ({
  label,
  suffix,
  defaultValue,
  width = 'w-24',
  readOnly,
  required
}: {
  label?: string
  suffix?: string
  defaultValue?: string
  width?: string
  readOnly?: boolean
  required?: boolean
}) => (
  <div className="flex items-center gap-2">
    {label && (
      <label className="text-sm text-gray-600 whitespace-nowrap">
        {required && <span className="text-red-500 mr-0.5">*</span>}
        {label}:
      </label>
    )}
    <div className={`relative ${width}`}>
      <input
        type="text"
        defaultValue={defaultValue}
        readOnly={readOnly}
        className={`w-full h-8 px-2 border rounded text-sm outline-none transition-all ${
          readOnly
            ? 'bg-gray-50 text-gray-500 border-gray-200'
            : 'bg-white border-gray-300 focus:ring-1 focus:ring-blue-500 focus:border-blue-500'
        }`}
      />
      {suffix && (
        <span className="absolute right-2 top-1/2 -translate-y-1/2 text-meta text-gray-400 pointer-events-none">
          {suffix}
        </span>
      )}
    </div>
  </div>
)

// --- 1. 缁煎悎閫忎腑鐩戞祴寮圭獥 ---
const ComprehensiveMonitorModal = ({
  device,
  onClose
}: {
  device: MonitorDevice
  onClose: () => void
}) => {
  const { t } = useTranslation(['monitoring', 'common'])
  // 浣跨敤妯″潡绾у埆棰勭敓鎴愮殑缂撳瓨鏁版嵁锛岄伩鍏嶆覆鏌撴湡闂磋皟鐢?Math.random
  const historyData = cachedHistoryData.get(device.id) || []

  return (
    <ModalOverlay onClose={onClose}>
      <ModalHeader
        title={t('monitoring:modal.realtimeMonitor', { name: device.patientName, bed: device.bedNumber })}
        onClose={onClose}
        icon={Activity}
        saveButtonText={t('monitoring:action.saveSubmit')}
      />
      <div className="p-6 overflow-y-auto space-y-6 bg-gray-50/50">
        <div className="bg-white rounded-lg border border-gray-200 p-5 shadow-sm">
          <h4 className="text-sm font-bold text-gray-700 mb-4 flex items-center">
            <Activity size={16} className="mr-2 text-red-500" /> {t('monitoring:chart.bpHrTrend')}
          </h4>
          <div className="h-[200px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={historyData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                <XAxis
                  dataKey="time"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 11 }}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 11 }}
                  domain={[40, 200]}
                />
                <Tooltip
                  contentStyle={{
                    borderRadius: '8px',
                    border: 'none',
                    boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)'
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="sbp"
                  name={t('monitoring:chart.sbp')}
                  stroke="#ef4444"
                  fill="#fef2f2"
                  strokeWidth={2}
                />
                <Area
                  type="monotone"
                  dataKey="dbp"
                  name={t('monitoring:chart.dbp')}
                  stroke="#3b82f6"
                  fill="#eff6ff"
                  strokeWidth={2}
                />
                <Line type="monotone" dataKey="hr" name={t('monitoring:chart.hr')} stroke="#10b981" strokeWidth={2} dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-5 shadow-sm">
          <h4 className="text-sm font-bold text-gray-700 mb-4 flex items-center">
            <TrendingUp size={16} className="mr-2 text-orange-500" /> {t('monitoring:chart.pressureTrend')}
          </h4>
          <div className="h-[200px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={historyData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                <XAxis
                  dataKey="time"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 11 }}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 11 }}
                  domain={[-200, 300]}
                />
                <Tooltip
                  contentStyle={{
                    borderRadius: '8px',
                    border: 'none',
                    boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)'
                  }}
                />
                <Line type="monotone" dataKey="ap" name={t('monitoring:chart.ap')} stroke="#8b5cf6" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="vp" name={t('monitoring:chart.vp')} stroke="#ec4899" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="tmp" name={t('monitoring:chart.tmp')} stroke="#f59e0b" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-5 shadow-sm">
          <h4 className="text-sm font-bold text-gray-700 mb-4 flex items-center">
            <BarChart3 size={16} className="mr-2 text-blue-500" /> {t('monitoring:chart.bfUfTrend')}
          </h4>
          <div className="h-[200px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={historyData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                <XAxis
                  dataKey="time"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 11 }}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 11 }}
                  domain={[0, 1000]}
                />
                <Tooltip
                  contentStyle={{
                    borderRadius: '8px',
                    border: 'none',
                    boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)'
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="bf"
                  name={t('monitoring:chart.bf')}
                  stroke="#0ea5e9"
                  fill="#f0f9ff"
                  strokeWidth={2}
                />
                <Line type="monotone" dataKey="uf" name={t('monitoring:chart.uf')} stroke="#6366f1" strokeWidth={2} dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </ModalOverlay>
  )
}

// --- 2. 澶勬柟璋冩暣寮圭獥 ---
const PrescriptionEditModal = ({
  device,
  onClose
}: {
  device: MonitorDevice
  onClose: () => void
}) => {
  const { t } = useTranslation(['monitoring', 'common'])
  const patientName = device.patientName || '--'
  const preBloodPressure =
    device.vitals.sbp > 0 && device.vitals.dbp > 0 ? `${device.vitals.sbp}/${device.vitals.dbp}mmHg` : ''
  const ufVolume = device.vitals.ufGoal > 0 ? (device.vitals.ufGoal / 1000).toFixed(1) : ''
  const [materials] = useState([
    // TODO: 从 MaterialCatalog API 加载
    { id: 1, name: 'JRHLL-025', category: '血路', count: 1, code: '', brand: '', spec: '', note: '' },
    { id: 2, name: '10ML注射器-10ML', category: '其他', count: 2, code: '', brand: '', spec: '10ML', note: '' },
    { id: 3, name: '15G', category: '透析器', count: 1, code: '', brand: 'NIPRO', spec: '', note: '' },
    { id: 4, name: '内瘘区', category: '护理区', count: 1, code: '1102011534', brand: '', spec: '', note: '' },
    { id: 5, name: '锐针-16G', category: '穿刺针', count: 2, code: '', brand: 'NIPRO', spec: '', note: '' }
  ])

  return (
    <ModalOverlay onClose={onClose} maxWidth="max-w-[1400px]">
      <div className="flex-1 overflow-y-auto p-6 bg-white space-y-6">
        {/* 鎮ｈ€呭鍚?*/}
        <div className="flex items-center text-lg font-bold text-gray-900 border-b pb-4 mb-4">
          {t('monitoring:prescription.patient')}: {patientName}
        </div>

        {/* 绗竴鎺掍綋寰佸弬鏁?*/}
        <div className="flex flex-wrap items-center gap-x-6 gap-y-4 text-[13px] text-gray-600">
          <span>{t('monitoring:prescription.dialysisMethod')}: HD</span>
          <span>{t('monitoring:prescription.preWeight')}: </span>
          <span>{t('monitoring:prescription.lastPostWeight')}: </span>
          <span>{t('monitoring:prescription.vsLastGain')}: </span>
          <span>{t('monitoring:prescription.currentGain', { pct: '' })}: </span>

          <div className="flex items-center gap-2">
            <span>{t('monitoring:prescription.dryWeight')}</span>
            <div className="flex items-center bg-blue-50/50 rounded pr-1 border border-blue-100">
              <input
                type="text"
                defaultValue=""
                className="w-16 h-7 bg-transparent text-center font-bold text-gray-800 outline-none"
              />
              {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（单位后缀） */}
              <span className="text-[10px] text-gray-400">kg</span>
            </div>
            <label className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                defaultChecked
                className="w-3 h-3 text-blue-600 rounded focus:ring-0 border-gray-300"
              />
              <span className="text-xs text-indigo-600 font-medium">{t('monitoring:action.syncToPlan')}</span>
            </label>
          </div>

          <PrescriptionInput label={t('monitoring:prescription.ufVolume', { pct: '' })} defaultValue={ufVolume} suffix="L" width="w-20" required />
          <span>{t('monitoring:prescription.preBP')}: {preBloodPressure}</span>
        </div>

        {/* 绗簩鎺掗€忔瀽鍙傛暟 */}
        <div className="flex flex-wrap items-center gap-x-8 gap-y-6">
          <PrescriptionInput label={t('monitoring:prescription.dialysisTime')} defaultValue="4" suffix="h" width="w-24" />
          <PrescriptionInput label={t('monitoring:prescription.standardBloodFlow')} defaultValue="200" suffix="ml/min" width="w-24" />

          <div className="flex items-center gap-3">
            <span className="text-sm text-gray-600">{t('monitoring:prescription.heparinType')}:</span>
            <div className="flex items-center gap-3">
              <label className="flex items-center gap-1 cursor-pointer">
                <input
                  type="radio"
                  name="hep"
                  defaultChecked
                  className="w-3.5 h-3.5 text-blue-600 border-gray-300 focus:ring-0"
                />
                <span className="text-sm text-gray-700">{t('monitoring:prescription.heparinNormal')}</span>
              </label>
              <label className="flex items-center gap-1 cursor-pointer">
                <input
                  type="radio"
                  name="hep"
                  className="w-3.5 h-3.5 text-blue-600 border-gray-300 focus:ring-0"
                />
                <span className="text-sm text-gray-700">{t('monitoring:prescription.heparinRelative')}</span>
              </label>
              <label className="flex items-center gap-1 cursor-pointer">
                <input
                  type="radio"
                  name="hep"
                  className="w-3.5 h-3.5 text-blue-600 border-gray-300 focus:ring-0"
                />
                <span className="text-sm text-gray-700">{t('monitoring:prescription.heparinAbsolute')}</span>
              </label>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">{t('monitoring:prescription.initialDrug')}:</label>
            <div className="relative">
              <select className="h-8 w-44 border rounded text-xs px-2 bg-white appearance-none outline-none focus:ring-1 focus:ring-blue-500">
                <option>那屈肝素钙注射液</option>
              </select>
              <ChevronDown
                className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
                size={12}
              />
            </div>
          </div>

          <PrescriptionInput label={t('monitoring:prescription.initialDose')} defaultValue="307.5" suffix="IU" width="w-28" />
        </div>

        {/* 绗笁鎺掔淮鎸佷笌鎬婚噺 */}
        <div className="flex flex-wrap items-center gap-x-8 gap-y-6">
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">{t('monitoring:prescription.maintenanceDrug')}:</label>
            <div className="relative">
              <select className="h-8 w-40 border rounded text-xs px-2 bg-white appearance-none outline-none focus:ring-1 focus:ring-blue-500 text-gray-400">
                <option>{t('monitoring:prescription.selectMaintenance')}</option>
              </select>
              <ChevronDown
                className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
                size={12}
              />
            </div>
          </div>
          <PrescriptionInput label={t('monitoring:prescription.infusionTime')} suffix="h" width="w-24" />
          <PrescriptionInput label={t('monitoring:prescription.infusionRate')} width="w-24" />
          <PrescriptionInput label={t('monitoring:prescription.maintenanceDose')} width="w-24" />
          <PrescriptionInput label={t('monitoring:prescription.totalDose')} defaultValue="307.5" suffix="IU" width="w-28" readOnly />
        </div>

        {/* 绗洓鎺掗€氳矾涓庡垎绫?*/}
        <div className="flex flex-wrap items-center gap-x-8 gap-y-6 pb-4 border-b border-gray-100">
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600 font-bold">
              <span className="text-red-500 mr-0.5">*</span>{t('monitoring:prescription.vascularAccess')}:
            </label>
            <div className="relative">
              <select className="h-8 w-40 border rounded text-xs px-2 bg-white appearance-none outline-none font-bold focus:ring-1 focus:ring-blue-500">
                <option>AVG-上臂</option>
              </select>
              <ChevronDown
                className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
                size={12}
              />
            </div>
            <button className="text-xs text-gray-500 hover:text-blue-600 px-2 underline decoration-dotted">
              {t('monitoring:action.view')}
            </button>
          </div>
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">{t('monitoring:prescription.dialysateCat')}:</label>
            <div className="relative">
              <select className="h-8 w-40 border rounded text-xs px-2 bg-white appearance-none outline-none focus:ring-1 focus:ring-blue-500">
                <option>A液+B液</option>
              </select>
              <ChevronDown
                className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
                size={12}
              />
            </div>
          </div>
        </div>

        {/* 鑰楁潗娣诲姞鍖?*/}
        <div className="flex items-center justify-between pt-2">
          <div className="flex items-center gap-2 w-[400px]">
            <div className="relative flex-1">
              <select className="h-9 w-full border rounded px-3 text-sm bg-white appearance-none outline-none text-gray-400 focus:ring-1 focus:ring-blue-500">
                <option>{t('monitoring:prescription.selectPlaceholder')}</option>
              </select>
              <ChevronDown
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
                size={14}
              />
            </div>
            <button className="px-4 py-1.5 border rounded text-sm text-gray-500 hover:bg-gray-50 flex items-center gap-1">
              <Plus size={14} /> {t('monitoring:action.add')}
            </button>
          </div>
          <button className="flex items-center gap-1 px-4 py-1.5 border rounded text-sm text-gray-400 hover:text-red-500 hover:border-red-200 transition-colors">
            <Trash2 size={14} /> {t('monitoring:action.delete')}
          </button>
        </div>

        {/* 鑰楁潗鍒楄〃琛ㄦ牸 */}
        <div className="bg-white rounded border border-gray-100 shadow-sm overflow-hidden">
          <table className="w-full text-left text-[13px] border-collapse">
            <thead className="bg-[#eef6ff] text-gray-700 font-bold border-b border-blue-100">
              <tr>
                <th className="px-4 py-3 w-10 text-center">
                  <input
                    type="checkbox"
                    className="w-3.5 h-3.5 rounded border-gray-300 text-blue-600 focus:ring-0"
                  />
                </th>
                <th className="px-4 py-3 w-16">{t('monitoring:materials.seqNo')}</th>
                <th className="px-4 py-3 w-64">{t('monitoring:materials.name')}</th>
                <th className="px-4 py-3 w-32 text-center">{t('monitoring:materials.category')}</th>
                <th className="px-4 py-3 w-32 text-center">{t('monitoring:materials.quantity')}</th>
                <th className="px-4 py-3">{t('monitoring:materials.code')}</th>
                <th className="px-4 py-3">{t('monitoring:materials.brand')}</th>
                <th className="px-4 py-3">{t('monitoring:materials.spec')}</th>
                <th className="px-4 py-3">{t('monitoring:materials.note')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {materials.map((m, idx) => (
                <tr key={m.id} className="hover:bg-blue-50/20 group">
                  <td className="px-4 py-3 text-center">
                    <input
                      type="checkbox"
                      className="w-3.5 h-3.5 rounded border-gray-300 text-blue-600 focus:ring-0"
                    />
                  </td>
                  <td className="px-4 py-3 text-gray-500">{idx + 1}</td>
                  <td className="px-4 py-3">
                    <div className="relative group/sel w-full">
                      <select className="w-full h-8 border rounded-sm px-2 bg-white appearance-none outline-none focus:ring-1 focus:ring-blue-500">
                        <option>{m.name}</option>
                      </select>
                      <ChevronDown
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-300 pointer-events-none group-hover/sel:text-blue-500 transition-colors"
                        size={12}
                      />
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center text-gray-600 font-medium">{m.category}</td>
                  <td className="px-4 py-3 text-center">
                    <input
                      type="number"
                      defaultValue={m.count}
                      className="w-20 h-8 border border-gray-200 rounded px-2 text-center text-sm outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </td>
                  <td className="px-4 py-3 text-gray-500 font-mono text-xs">{m.code}</td>
                  <td className="px-4 py-3 text-gray-600">{m.brand}</td>
                  <td className="px-4 py-3 text-gray-600">{m.spec}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-between">
                      <span className="text-gray-400 italic text-xs">{m.note || ''}</span>
                      <button className="text-indigo-600 hover:text-indigo-800 font-medium ml-2">{t('monitoring:action.modify')}</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* 搴曢儴鎸夐挳 */}
        <div className="flex justify-end gap-3 pt-4">
          <button
            onClick={onClose}
            className="px-8 py-2 border rounded-lg text-sm text-gray-600 hover:bg-gray-50"
          >
            {t('monitoring:action.close')}
          </button>
          <button className="px-8 py-2 bg-indigo-600 text-white rounded-lg text-sm font-bold shadow hover:bg-indigo-700">
            {t('monitoring:action.submit')}
          </button>
        </div>
      </div>
    </ModalOverlay>
  )
}

// --- 3. 鍖诲槺绠＄悊寮圭獥 ---
interface OrderItem {
  id: string
  content: string
  frequency: string
  doctor: string
  time: string
  status: 'ACTIVE' | 'STOPPED' | 'PENDING' | 'EXECUTED'
}

const OrderListModal = ({
  device,
  onClose
}: {
  device: MonitorDevice
  onClose: () => void
}) => {
  const { t } = useTranslation(['monitoring', 'common'])
  const [activeTab, setActiveTab] = useState<'LONG' | 'TEMP'>('LONG')
  const [longOrders, setLongOrders] = useState<OrderItem[]>([])
  const [tempOrders, setTempOrders] = useState<OrderItem[]>([])
  const [ordersLoading, setOrdersLoading] = useState(false)

  const currentOrders = activeTab === 'LONG' ? longOrders : tempOrders

  useEffect(() => {
    if (!device.patientId) {
      setLongOrders([])
      setTempOrders([])
      return
    }

    const toOrderItems = (items: RestPatientOrder[]): OrderItem[] =>
      items.map((item) => ({
        id: item.id,
        content: item.content || '--',
        frequency: item.frequency || '--',
        doctor: item.doctorName || '--',
        time: item.startTime || item.createdAt || '--',
        status:
          item.status === '停止' || item.status === '停用'
            ? 'STOPPED'
            : item.status === '已执行'
            ? 'EXECUTED'
            : item.status === '执行中'
            ? 'ACTIVE'
            : 'PENDING',
      }))

    setOrdersLoading(true)
    Promise.all([
      restApi.getPatientOrders(String(device.patientId), { type: 'LONG' }).catch(() => null),
      restApi.getPatientOrders(String(device.patientId), { type: 'TEMP' }).catch(() => null),
    ])
      .then(([longRes, tempRes]) => {
        setLongOrders(toOrderItems(longRes?.data ?? []))
        setTempOrders(toOrderItems(tempRes?.data ?? []))
      })
      .finally(() => setOrdersLoading(false))
  }, [device.patientId])

  return (
    <ModalOverlay onClose={onClose} maxWidth="max-w-[1100px]">
      <ModalHeader
        title={t('monitoring:modal.orders', { name: device.patientName, bed: device.bedNumber })}
        onClose={onClose}
        icon={ClipboardList}
        saveButtonText={t('monitoring:action.saveSubmit')}
      />
      <div className="flex flex-col flex-1 bg-gray-50 overflow-hidden">
        <div className="bg-white border-b border-gray-100 px-6 py-4 flex items-center justify-between shadow-sm shrink-0">
          <div className="flex bg-gray-100 p-1 rounded-lg">
            <button
              onClick={() => setActiveTab('LONG')}
              className={`px-8 py-2 text-sm font-bold rounded-lg transition-all ${
                activeTab === 'LONG' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:bg-gray-200'
              }`}
            >
              {t('monitoring:orders.longTerm')}
            </button>
            <button
              onClick={() => setActiveTab('TEMP')}
              className={`px-8 py-2 text-sm font-bold rounded-lg transition-all ${
                activeTab === 'TEMP' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:bg-gray-200'
              }`}
            >
              {t('monitoring:orders.temporary')}
            </button>
          </div>
          <button className="px-6 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-bold hover:bg-blue-700 shadow-lg">
            {t('monitoring:action.newOrder')}
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
            <table className="w-full text-left text-sm border-collapse">
              <thead className="bg-slate-50/80 backdrop-blur-sm text-gray-500 font-bold border-b border-gray-100 sticky top-0 z-10">
                <tr>
                  <th className="px-6 py-4 w-12 text-center"></th>
                  <th className="px-6 py-4">{t('monitoring:orders.content')}</th>
                  <th className="px-6 py-4 w-32 text-center">{t('monitoring:orders.frequency')}</th>
                  <th className="px-6 py-4 w-48">{t('monitoring:orders.doctorTime')}</th>
                  <th className="px-6 py-4 w-32 text-center">{t('monitoring:orders.status')}</th>
                  <th className="px-6 py-4 w-40 text-right">{t('monitoring:orders.operation')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {ordersLoading ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-sm text-gray-400">
                      {t('common:status.loading')}
                    </td>
                  </tr>
                ) : null}
                {currentOrders.map((order) => (
                  <tr
                    key={order.id}
                    className={`group hover:bg-blue-50/30 transition-all ${
                      order.status === 'STOPPED' ? 'opacity-50' : ''
                    }`}
                  >
                    <td className="px-6 py-4"></td>
                    <td className="px-6 py-4 font-bold text-gray-800">{order.content}</td>
                    <td className="px-6 py-4 text-center">
                      <span className="bg-gray-100 text-gray-600 px-2 py-1 rounded text-xs">
                        {order.frequency}
                      </span>
                    </td>
                    <td className="px-6 py-4 flex flex-col">
                      <span className="text-gray-700 font-bold">{order.doctor}</span>
                      <span className="text-meta text-gray-400 font-mono">{order.time}</span>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span
                        className={`px-3 py-1 rounded-full text-meta font-bold shadow-sm ${
                          order.status === 'ACTIVE'
                            ? 'bg-green-500 text-white'
                            : order.status === 'STOPPED'
                            ? 'bg-gray-400 text-white'
                            : order.status === 'EXECUTED'
                            ? 'bg-blue-500 text-white'
                            : 'bg-orange-400 text-white'
                        }`}
                      >
                        {order.status === 'ACTIVE'
                          ? t('monitoring:orders.status.active')
                          : order.status === 'STOPPED'
                          ? t('monitoring:orders.status.stopped')
                          : order.status === 'EXECUTED'
                          ? t('monitoring:orders.status.executed')
                          : t('monitoring:orders.status.pending')}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right opacity-0 group-hover:opacity-100 transition-opacity">
                      <button className="p-2 hover:bg-blue-100 text-blue-600 rounded-lg">
                        <Edit3 size={15} />
                      </button>
                      <button className="p-2 hover:bg-red-100 text-red-600 rounded-lg">
                        <Trash2 size={15} />
                      </button>
                    </td>
                  </tr>
                ))}
                {!ordersLoading && currentOrders.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-sm text-gray-400">
                      {t('common:empty.default')}
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </ModalOverlay>
  )
}

// --- 4. 濉啓灏忕粨寮圭獥 ---
const SummaryModal = ({
  device,
  onClose
}: {
  device: MonitorDevice
  onClose: () => void
}) => {
  const { t } = useTranslation(['monitoring', 'common'])
  const [summary, setSummary] = useState('')
  const templates = [
    t('monitoring:summary.template1'),
    t('monitoring:summary.template2'),
    t('monitoring:summary.template3'),
    t('monitoring:summary.template4')
  ]

  return (
    <ModalOverlay onClose={onClose} maxWidth="max-w-2xl">
      <ModalHeader
        title={t('monitoring:modal.summary', { name: device.patientName, bed: device.bedNumber })}
        onClose={onClose}
        icon={FileEdit}
        saveButtonText={t('monitoring:action.saveSubmit')}
      />
      <div className="p-6 space-y-6 bg-white overflow-y-auto">
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm font-bold text-gray-700 flex items-center">
              <Sparkles size={16} className="mr-1 text-indigo-500" /> {t('monitoring:summary.quickTemplate')}
            </label>
            <span className="text-meta text-gray-400">{t('monitoring:summary.clickToFill')}</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {templates.map((tpl, i) => (
              <button
                key={i}
                onClick={() => setSummary((prev) => (prev ? `${prev}\n${tpl}` : tpl))}
                className="text-meta bg-indigo-50 text-indigo-600 px-3 py-1.5 rounded-lg border border-indigo-100 hover:bg-indigo-100 transition-colors"
              >
                {tpl}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-sm font-bold text-gray-700 mb-2">{t('monitoring:summary.detailContent')}</label>
          <textarea
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            className="w-full h-48 p-4 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none text-sm leading-relaxed transition-all"
            placeholder={t('monitoring:summary.placeholder')}
          ></textarea>
        </div>

        <div className="flex items-center gap-4 p-4 bg-gray-50 rounded-lg border border-dashed border-gray-200">
          <div className="flex flex-col">
            <span className="text-meta text-gray-400 uppercase font-bold">{t('monitoring:summary.finalBP')}</span>
            <span className="text-sm font-bold text-gray-800">
              {device.vitals.sbp}/{device.vitals.dbp} mmHg
            </span>
          </div>
          <div className="w-px h-8 bg-gray-200"></div>
          <div className="flex flex-col">
            <span className="text-meta text-gray-400 uppercase font-bold">{t('monitoring:summary.finalUF')}</span>
            <span className="text-sm font-bold text-gray-800">{device.vitals.ufVolume.toFixed(2)} L</span>
          </div>
        </div>

        <div className="flex justify-end space-x-3 pt-2">
          <button
            onClick={onClose}
            className="px-6 py-2 border border-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors"
          >
            {t('monitoring:action.cancel')}
          </button>
          <button
            onClick={() => {
              alert(t('monitoring:summary.saved'))
              onClose()
            }}
            className="px-8 py-2 bg-indigo-600 text-white rounded-lg text-sm font-bold shadow-lg shadow-indigo-200 hover:bg-indigo-700 transition-all active:scale-95"
          >
            {t('monitoring:action.saveSummary')}
          </button>
        </div>
      </div>
    </ModalOverlay>
  )
}

// --- 主组件 ---
export default function Monitoring() {
  const { t } = useTranslation(['monitoring', 'common'])
  const { devices, loading, loadError } = useMonitoringData()
  const { filteredDevices, activeZone, setActiveZone, searchTerm, setSearchTerm } = useDeviceFilter(devices)
  const { activeModal, selectedDevice, openModal, closeModal } = useModalManager()

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'warning':
        return 'border-state-waiting bg-state-waiting-bg/30'
      case 'alarm':
        return 'border-state-alert bg-state-alert-bg'
      case 'offline':
        return 'border-gray-200 bg-state-offline-bg opacity-60'
      case 'unknown':
        return 'border-slate-200 bg-slate-50'
      default:
        return 'border-state-finished bg-white'
    }
  }

  return (
    <div className="h-full flex flex-col max-w-[1800px] mx-auto">
      <div className="mb-6 flex flex-col md:flex-row md:items-center justify-between gap-4 shrink-0">
        <div>
          <h2 className="text-2xl font-bold text-gray-800 flex items-center">
            <Monitor className="mr-3 text-blue-600" /> {t('monitoring:title')}
          </h2>
          <p className="text-gray-500 text-sm mt-1">
            {t('monitoring:subtitle.treating', { count: devices.filter((d) => !!d.patientName).length })}
          </p>
        </div>
        <div className="flex items-center space-x-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={16} />
            <input
              type="text"
              placeholder={t('monitoring:search.placeholder')}
              className="pl-9 pr-4 py-2 rounded-lg border border-gray-200 text-sm focus:ring-2 focus:ring-blue-500 outline-none w-48 bg-white"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="flex bg-white p-1 rounded-lg border border-gray-200 shadow-sm">
            {['ALL', 'A', 'B', 'C'].map((z) => (
              <button
                key={z}
                onClick={() => setActiveZone(z)}
                className={`px-3 py-1 rounded-md text-xs font-bold transition-all ${
                  activeZone === z ? 'bg-blue-600 text-white' : 'text-gray-500 hover:bg-gray-100'
                }`}
              >
                {z === 'ALL' ? t('monitoring:zone.all') : t('monitoring:zone.label', { zone: z })}
              </button>
            ))}
          </div>
        </div>
      </div>

      <StatusGrid
        devices={filteredDevices}
        loading={loading}
        loadError={loadError}
        getStatusColor={getStatusColor}
        onOpenModal={openModal}
        onReload={() => window.location.reload()}
      />

      {/* --- RENDER MODALS --- */}
      {activeModal === 'COMPREHENSIVE' && selectedDevice && (
        <ComprehensiveMonitorModal device={selectedDevice} onClose={() => closeModal()} />
      )}
      {activeModal === 'PRESCRIPTION' && selectedDevice && (
        <PrescriptionEditModal device={selectedDevice} onClose={() => closeModal()} />
      )}
      {activeModal === 'ORDERS' && selectedDevice && (
        <OrderListModal device={selectedDevice} onClose={() => closeModal()} />
      )}
      {activeModal === 'SUMMARY' && selectedDevice && (
        <SummaryModal device={selectedDevice} onClose={() => closeModal()} />
      )}
    </div>
  )
}


