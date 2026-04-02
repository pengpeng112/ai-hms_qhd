import React, { useState, useMemo, memo, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { MonitorDevice } from '../types/original'
import { MOCK_MONITOR_DEVICES } from '../constants'
import {
  Monitor,
  Search,
  X,
  Activity,
  Clock,
  AlertOctagon,
  TrendingUp,
  Droplet,
  ClipboardList,
  Wifi,
  Trash2,
  Plus,
  ChevronDown,
  FileEdit,
  Edit3,
  BarChart3,
  Sparkles
} from 'lucide-react'
import {
  LineChart,
  Line,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  AreaChart,
  Area
} from 'recharts'

type ModalType = 'COMPREHENSIVE' | 'PRESCRIPTION' | 'ORDERS' | 'SUMMARY' | null

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
      className={`bg-white rounded-xl shadow-2xl w-full ${maxWidth} max-h-[95vh] overflow-hidden m-4 flex flex-col`}
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
      <button className="px-4 py-1.5 bg-blue-600 text-white rounded text-sm font-bold hover:bg-blue-700 transition-colors shadow-sm">
        {saveButtonText}
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
        <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-gray-400 pointer-events-none">
          {suffix}
        </span>
      )}
    </div>
  </div>
)

// --- 1. 综合透中监测弹窗 ---
const ComprehensiveMonitorModal = ({
  device,
  onClose
}: {
  device: MonitorDevice
  onClose: () => void
}) => {
  const { t } = useTranslation(['monitoring', 'common'])
  // 使用模块级别预生成的缓存数据，避免渲染期间调用 Math.random
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
        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
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
        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
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
        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
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

// --- 2. 处方调整弹窗 ---
const PrescriptionEditModal = ({
  device,
  onClose
}: {
  device: MonitorDevice
  onClose: () => void
}) => {
  const { t } = useTranslation(['monitoring', 'common'])
  const [materials] = useState([
    { id: 1, name: 'JRHLL-025', category: '血路管', count: 1, code: '', brand: '', spec: '', note: '' },
    { id: 2, name: '10ML注射器-10ML', category: '其他', count: 2, code: '', brand: '', spec: '10ML', note: '' },
    { id: 3, name: '15G', category: '透析、血滤器', count: 1, code: '', brand: 'NIPRO', spec: '', note: '' },
    { id: 4, name: '内瘘包', category: '护理包', count: 1, code: '1102011534', brand: '', spec: '', note: '' },
    { id: 5, name: '锐针-16G', category: '穿刺针', count: 2, code: '', brand: 'NIPRO', spec: '', note: '' }
  ])

  return (
    <ModalOverlay onClose={onClose} maxWidth="max-w-[1400px]">
      <div className="flex-1 overflow-y-auto p-6 bg-white space-y-6">
        {/* 患者姓名 */}
        <div className="flex items-center text-lg font-bold text-gray-900 border-b pb-4 mb-4">
          {t('monitoring:prescription.patient')}: {device.patientName || '高敬兰'}
        </div>

        {/* 第一排体征参数 */}
        <div className="flex flex-wrap items-center gap-x-6 gap-y-4 text-[13px] text-gray-600">
          <span>{t('monitoring:prescription.dialysisMethod')}: HD</span>
          <span>{t('monitoring:prescription.preWeight')}: 69.5kg</span>
          <span>{t('monitoring:prescription.lastPostWeight')}: 67.2kg</span>
          <span>{t('monitoring:prescription.vsLastGain')}: 2.3kg</span>
          <span>{t('monitoring:prescription.currentGain', { pct: '2.81' })}: 1.9kg</span>

          <div className="flex items-center gap-2">
            <span>{t('monitoring:prescription.dryWeight')}</span>
            <div className="flex items-center bg-blue-50/50 rounded pr-1 border border-blue-100">
              <input
                type="text"
                defaultValue="67.6"
                className="w-16 h-7 bg-transparent text-center font-bold text-gray-800 outline-none"
              />
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

          <PrescriptionInput label={t('monitoring:prescription.ufVolume', { pct: '3.11' })} defaultValue="2.1" suffix="L" width="w-20" required />
          <span>{t('monitoring:prescription.preBP')}: 137/68mmHg</span>
        </div>

        {/* 第二排透析参数 */}
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
                <option>那屈肝素钙注射...</option>
              </select>
              <ChevronDown
                className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
                size={12}
              />
            </div>
          </div>

          <PrescriptionInput label={t('monitoring:prescription.initialDose')} defaultValue="307.5" suffix="axiau" width="w-28" />
        </div>

        {/* 第三排维持与总量 */}
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
          <PrescriptionInput label={t('monitoring:prescription.totalDose')} defaultValue="307.5" suffix="axiau" width="w-28" readOnly />
        </div>

        {/* 第四排通路与分类 */}
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

        {/* 耗材添加区 */}
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

        {/* 耗材列表表格 */}
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

        {/* 底部按钮 */}
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

// --- 3. 医嘱管理弹窗 ---
interface OrderItem {
  id: number
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

  const [longOrders] = useState<OrderItem[]>([
    {
      id: 1,
      content: '0.9% 氯化钠注射液 100ml',
      frequency: '每次透析',
      doctor: '王医生',
      time: '2025-12-01 08:30',
      status: 'ACTIVE'
    },
    {
      id: 2,
      content: '左卡尼汀注射液 1.0g',
      frequency: '每次透析',
      doctor: '王医生',
      time: '2025-12-01 08:31',
      status: 'ACTIVE'
    },
    {
      id: 3,
      content: '那屈肝素钙 2500iu iv',
      frequency: '每次透析',
      doctor: '李医生',
      time: '2025-11-15 09:00',
      status: 'STOPPED'
    }
  ])

  const [tempOrders] = useState<OrderItem[]>([
    {
      id: 4,
      content: '50% 葡萄糖注射液 20ml iv',
      frequency: 'ST',
      doctor: '王医生',
      time: '2025-12-13 10:15',
      status: 'EXECUTED'
    },
    {
      id: 5,
      content: '去乙酰毛花苷 0.2mg iv',
      frequency: 'ST',
      doctor: '陈主任',
      time: '2025-12-13 11:20',
      status: 'PENDING'
    }
  ])

  const currentOrders = activeTab === 'LONG' ? longOrders : tempOrders

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
          <div className="flex bg-gray-100 p-1 rounded-xl">
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
          <button className="px-6 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-bold hover:bg-blue-700 shadow-lg">
            {t('monitoring:action.newOrder')}
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
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
                      <span className="text-[10px] text-gray-400 font-mono">{order.time}</span>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span
                        className={`px-3 py-1 rounded-full text-[11px] font-bold shadow-sm ${
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
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </ModalOverlay>
  )
}

// --- 4. 填写小结弹窗 ---
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
            <span className="text-[10px] text-gray-400">{t('monitoring:summary.clickToFill')}</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {templates.map((tpl, i) => (
              <button
                key={i}
                onClick={() => setSummary((prev) => (prev ? `${prev}\n${tpl}` : tpl))}
                className="text-[11px] bg-indigo-50 text-indigo-600 px-3 py-1.5 rounded-lg border border-indigo-100 hover:bg-indigo-100 transition-colors"
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
            className="w-full h-48 p-4 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none text-sm leading-relaxed transition-all"
            placeholder={t('monitoring:summary.placeholder')}
          ></textarea>
        </div>

        <div className="flex items-center gap-4 p-4 bg-gray-50 rounded-xl border border-dashed border-gray-200">
          <div className="flex flex-col">
            <span className="text-[10px] text-gray-400 uppercase font-bold">{t('monitoring:summary.finalBP')}</span>
            <span className="text-sm font-bold text-gray-800">
              {device.vitals.sbp}/{device.vitals.dbp} mmHg
            </span>
          </div>
          <div className="w-px h-8 bg-gray-200"></div>
          <div className="flex flex-col">
            <span className="text-[10px] text-gray-400 uppercase font-bold">{t('monitoring:summary.finalUF')}</span>
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

// --- 预生成图表数据（避免每次渲染重新计算）---
const generateMiniGraphData = (device: MonitorDevice) => {
  return Array.from({ length: 12 }).map(() => ({
    sbp: device.vitals.sbp + Math.floor(Math.random() * 8 - 4),
    hr: device.vitals.hr + Math.floor(Math.random() * 6 - 3)
  }))
}

// 预生成历史数据（用于弹窗图表）
const generateHistoryData = (device: MonitorDevice) => {
  const now = new Date()
  return Array.from({ length: 15 }).map((_, i) => {
    const time = new Date(now)
    time.setMinutes(now.getMinutes() - (14 - i) * 15)
    return {
      time: time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      sbp: device.vitals.sbp + Math.floor(Math.random() * 10 - 5),
      dbp: device.vitals.dbp + Math.floor(Math.random() * 8 - 4),
      hr: device.vitals.hr + Math.floor(Math.random() * 10 - 5),
      ap: -120 + Math.floor(Math.random() * 20),
      vp: 110 + Math.floor(Math.random() * 15),
      tmp: 130 + Math.floor(Math.random() * 30),
      bf: 240 + Math.floor(Math.random() * 10 - 5),
      uf: 500 + Math.floor(Math.random() * 100 - 50)
    }
  })
}

// 在模块级别缓存所有设备的图表数据
const cachedGraphData = new Map<string, { sbp: number; hr: number }[]>()
const cachedHistoryData = new Map<string, ReturnType<typeof generateHistoryData>>()
MOCK_MONITOR_DEVICES.forEach((device) => {
  cachedGraphData.set(device.id, generateMiniGraphData(device))
  cachedHistoryData.set(device.id, generateHistoryData(device))
})

// --- Mini图表组件（使用memo防止不必要的重渲染）---
const MiniVitalsChart = memo(({ deviceId }: { deviceId: string }) => {
  const data = cachedGraphData.get(deviceId) || []
  const containerRef = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(180)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const updateWidth = () => {
      setWidth(container.clientWidth || 180)
    }

    updateWidth()

    const resizeObserver = new ResizeObserver(() => {
      // 防抖 200ms
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
      timeoutRef.current = setTimeout(updateWidth, 200)
    })

    resizeObserver.observe(container)
    return () => {
      resizeObserver.disconnect()
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
    }
  }, [])

  return (
    <div ref={containerRef} className="h-10 w-full mt-1" style={{ contain: 'layout style paint' }}>
      <AreaChart data={data} width={width} height={40}>
        <Area
          type="monotone"
          dataKey="sbp"
          stroke="#ef4444"
          fill="#fef2f2"
          strokeWidth={1.5}
          dot={false}
          isAnimationActive={false}
        />
        <Line
          type="monotone"
          dataKey="hr"
          stroke="#10b981"
          strokeWidth={1.5}
          dot={false}
          isAnimationActive={false}
        />
      </AreaChart>
    </div>
  )
})

MiniVitalsChart.displayName = 'MiniVitalsChart'

// --- 主组件 ---
export default function Monitoring() {
  const { t } = useTranslation(['monitoring', 'common'])
  const [activeModal, setActiveModal] = useState<ModalType>(null)
  const [selectedDevice, setSelectedDevice] = useState<MonitorDevice | null>(null)
  const [activeZone, setActiveZone] = useState('ALL')
  const [searchTerm, setSearchTerm] = useState('')

  const filteredDevices = useMemo(() => {
    return MOCK_MONITOR_DEVICES.filter((d) => {
      const zoneMatch = activeZone === 'ALL' || d.bedNumber.startsWith(activeZone)
      const searchMatch =
        (d.patientName || '').includes(searchTerm) || d.bedNumber.includes(searchTerm)
      return zoneMatch && searchMatch
    })
  }, [activeZone, searchTerm])


  const openModal = (device: MonitorDevice, type: ModalType) => {
    setSelectedDevice(device)
    setActiveModal(type)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'warning':
        return 'border-orange-400 bg-orange-50/30'
      case 'alarm':
        return 'border-red-500 bg-red-50'
      case 'offline':
        return 'border-gray-200 bg-gray-50 opacity-60'
      default:
        return 'border-green-400 bg-white'
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
            {t('monitoring:subtitle.treating', { count: MOCK_MONITOR_DEVICES.filter((d) => d.status !== 'offline').length })}
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

      <div className="flex-1 overflow-y-auto pr-2 pb-10">
        <div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4"
          style={{ contain: 'layout style' }}
        >
          {filteredDevices.map((device, i) => {
            const gender = i % 2 === 0 ? t('monitoring:label.male') : t('monitoring:label.female')
            const age = 45 + (i % 30)
            const dryWeight = 65.5 + (i % 10)
            const weightGainVal = 1.2 + (i % 2)
            const weightGainPct = ((weightGainVal / dryWeight) * 100).toFixed(1)
            const access = i % 3 === 0 ? 'AVF' : 'TCC'

            return (
              <div
                key={device.id}
                className={`rounded-xl border-2 p-3 flex flex-col shadow-sm relative group ${getStatusColor(
                  device.status
                )}`}
                style={{ contain: 'layout style paint' }}
              >
                <div className="flex justify-between items-start mb-2 gap-1">
                  <div className="flex items-center min-w-0 flex-1">
                    <div
                      className={`w-8 h-8 rounded-lg flex items-center justify-center font-bold text-sm mr-2 shadow-sm shrink-0 ${
                        device.status === 'alarm'
                          ? 'bg-red-600 text-white'
                          : device.status === 'warning'
                          ? 'bg-orange-500 text-white'
                          : 'bg-slate-800 text-white'
                      }`}
                    >
                      {device.bedNumber}
                    </div>
                    <div className="min-w-0 flex-1">
                      <h4 className="font-bold text-gray-900 text-sm flex items-center truncate">
                        <span className="truncate">{device.patientName || t('monitoring:card.idle')}</span>
                        {device.patientName && (
                          <span className="text-[10px] text-gray-500 ml-1 font-normal whitespace-nowrap">
                            ({gender} · {age}{t('monitoring:label.age')})
                          </span>
                        )}
                      </h4>
                      <div className="flex items-center text-[10px] text-gray-500 font-medium">
                        <Wifi
                          size={10}
                          className={`mr-1 shrink-0 ${device.status === 'offline' ? 'text-gray-300' : 'text-green-500'}`}
                        />
                        <span className="text-blue-600">{device.mode}</span>
                        <span className="mx-1 opacity-40">·</span>
                        <Clock size={10} className="mr-0.5 shrink-0" /> {device.timeRemaining}
                      </div>
                    </div>
                  </div>
                  <div className="flex shrink-0">
                    <button
                      onClick={() => openModal(device, 'ORDERS')}
                      className="p-1.5 hover:bg-white rounded text-gray-400 hover:text-blue-600 transition-colors"
                      title={t('monitoring:action.viewOrders')}
                    >
                      <ClipboardList size={16} />
                    </button>
                    {device.patientName && (
                      <button
                        onClick={() => openModal(device, 'SUMMARY')}
                        className="p-1.5 hover:bg-white rounded text-gray-400 hover:text-indigo-600 transition-colors"
                        title={t('monitoring:action.writeSummary')}
                      >
                        <FileEdit size={16} />
                      </button>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-1 mb-2 py-1.5 border-t border-b border-gray-100 text-[10px] font-medium text-gray-600">
                  <div className="flex flex-col">
                    <span className="text-gray-400 scale-90 origin-left">{t('monitoring:card.dryWeight')}</span>
                    <span className="text-gray-800 font-bold">{dryWeight.toFixed(1)}kg</span>
                  </div>
                  <div className="flex flex-col">
                    <span className="text-gray-400 scale-90 origin-left">{t('monitoring:card.gainPct')}</span>
                    <span className="text-blue-600 font-bold">
                      {weightGainVal.toFixed(1)}kg ({weightGainPct}%)
                    </span>
                  </div>
                  <div className="flex flex-col items-end">
                    <span className="text-gray-400 scale-90 origin-right">{t('monitoring:card.vascularAccess')}</span>
                    <span className="text-gray-800 font-bold">{access}</span>
                  </div>
                </div>

                <div
                  onClick={() => device.status !== 'offline' && openModal(device, 'COMPREHENSIVE')}
                  className="bg-white/80 rounded-lg p-2 mb-2 cursor-pointer hover:bg-white hover:shadow-md transition-colors border border-transparent hover:border-blue-200 group/vitals"
                >
                  <div className="flex justify-between items-end mb-1 px-1">
                    <div className="flex flex-col">
                      <span className="text-[9px] font-bold text-gray-400 uppercase">{t('monitoring:card.bp')}</span>
                      <div className="flex items-baseline">
                        <span
                          className={`text-base font-bold font-mono leading-none ${
                            device.status === 'alarm' ? 'text-red-600' : 'text-gray-900'
                          }`}
                        >
                          {device.vitals.sbp}/{device.vitals.dbp}
                        </span>
                        <span className="text-[9px] text-gray-400 ml-0.5 font-bold">mmHg</span>
                      </div>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className="text-[9px] font-bold text-gray-400 uppercase">{t('monitoring:card.hr')}</span>
                      <div className="flex items-baseline">
                        <span className="text-base font-bold font-mono leading-none text-gray-900">
                          {device.vitals.hr}
                        </span>
                        <span className="text-[9px] text-gray-400 ml-0.5 font-bold">bpm</span>
                      </div>
                    </div>
                  </div>
                  <MiniVitalsChart deviceId={device.id} />
                </div>

                <div
                  onClick={() => device.status !== 'offline' && openModal(device, 'PRESCRIPTION')}
                  className="flex-1 cursor-pointer group/pres pt-1"
                >
                  <div className="flex justify-between items-center text-[10px] mb-1 px-1 font-bold">
                    <span className="text-gray-500 flex items-center">
                      <Droplet size={10} className="mr-1 text-blue-500" />{' '}
                      {device.vitals.ufVolume.toFixed(2)}L / {device.vitals.ufGoal.toFixed(1)}L
                    </span>
                    <span className="text-blue-700">
                      {Math.round((device.vitals.ufVolume / device.vitals.ufGoal) * 100)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 h-1.5 rounded-full overflow-hidden mb-2 relative shadow-inner">
                    <div
                      className={`h-full rounded-full ${
                        device.status === 'alarm' ? 'bg-red-500' : 'bg-blue-600'
                      }`}
                      style={{
                        width: `${Math.min(100, (device.vitals.ufVolume / device.vitals.ufGoal) * 100)}%`
                      }}
                    ></div>
                  </div>
                </div>
                {device.status === 'alarm' && (
                  <div className="absolute bottom-1 right-1 text-red-600 flex items-center bg-red-100 p-0.5 rounded-full">
                    <AlertOctagon size={14} className="animate-bounce" />
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* --- RENDER MODALS --- */}
      {activeModal === 'COMPREHENSIVE' && selectedDevice && (
        <ComprehensiveMonitorModal device={selectedDevice} onClose={() => setActiveModal(null)} />
      )}
      {activeModal === 'PRESCRIPTION' && selectedDevice && (
        <PrescriptionEditModal device={selectedDevice} onClose={() => setActiveModal(null)} />
      )}
      {activeModal === 'ORDERS' && selectedDevice && (
        <OrderListModal device={selectedDevice} onClose={() => setActiveModal(null)} />
      )}
      {activeModal === 'SUMMARY' && selectedDevice && (
        <SummaryModal device={selectedDevice} onClose={() => setActiveModal(null)} />
      )}
    </div>
  )
}
