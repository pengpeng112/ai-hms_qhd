// Scheme & Order Tab - 治疗医嘱管理
// 基于 UI 设计 v1.6 和 v1.7 完全重写，添加编辑医嘱单功能

import { useState, useEffect, useCallback } from 'react'
import {
  PlusCircle, History, PowerOff, Link, RotateCcw,
  ClipboardList, ScrollText, Edit3, Printer, Calendar, Copy,
  Plus, Save, Send, Pill, Zap, Loader2
} from 'lucide-react'
import { message } from 'antd'
import { DetailCard } from '@/components/ui'
import { OrderModal } from '@/components/patient/modals'
import type { Patient } from '@/types/original'
import { orderApi, prescriptionApi } from '@/services/orderApi'
import type { Order, Prescription, PrescriptionOrderItem } from '@/services/orderApi'

interface SchemeOrderTabProps {
  patient: Patient
}

// UI 展示用医嘱类型（后端状态 → 前端状态 adapter）
interface EnrichedOrder {
  id: string
  name: string
  dose: string
  unit: string
  route: string
  frequency: string
  timing: string
  execTiming?: string
  doctor: string
  startTime: string
  stopTime?: string
  status: '在用' | '停用'
  type: '长期' | '临时'
  lastExec?: string
  remark: string
  groupId?: string
  // 保留原始数据用于编辑
  _raw?: Order
}

// 后端状态 → UI 状态映射
function mapOrderStatus(status: string): '在用' | '停用' {
  if (status === '待执行' || status === '执行中') return '在用'
  return '停用'
}

// 后端 Order → UI EnrichedOrder
function toEnrichedOrder(o: Order): EnrichedOrder {
  return {
    id: o.id,
    name: o.name || o.content,
    dose: o.dose || '',
    unit: o.unit || '',
    route: o.route || '',
    frequency: o.frequency || '',
    timing: o.timing || '',
    execTiming: o.execTiming,
    doctor: o.doctorName,
    startTime: o.startTime ? new Date(o.startTime).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : '',
    stopTime: o.endTime ? new Date(o.endTime).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : undefined,
    status: mapOrderStatus(o.status),
    type: o.type as '长期' | '临时',
    lastExec: o.executedAt ? new Date(o.executedAt).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : undefined,
    remark: o.notes || '',
    groupId: o.groupId,
    _raw: o,
  }
}

interface OrderDisplayGroup {
  groupId?: string
  items: EnrichedOrder[]
}

export default function SchemeOrderTab({ patient }: SchemeOrderTabProps) {
  // 状态管理
  const [selectedLongIds, setSelectedLongIds] = useState<string[]>([])
  const [selectedTempIds, setSelectedTempIds] = useState<string[]>([])
  const [showLongHistory, setShowLongHistory] = useState(false)
  const [showTempHistory, setShowTempHistory] = useState(false)
  const [orderSubTab, setOrderSubTab] = useState<'DIALYSIS' | 'SHEET'>('DIALYSIS')
  const [isOrderModalOpen, setIsOrderModalOpen] = useState(false)
  const [orderModalType, setOrderModalType] = useState<'长期' | '临时'>('长期')
  const [editOrder, setEditOrder] = useState<Order | undefined>(undefined)

  // 医嘱单相关状态
  const [selectedPrescriptionId, setSelectedPrescriptionId] = useState<string>('')
  const [isEditingSheet, setIsEditingSheet] = useState(false)

  // API 数据
  const [orders, setOrders] = useState<EnrichedOrder[]>([])
  const [historyOrders, setHistoryOrders] = useState<EnrichedOrder[]>([])
  const [loading, setLoading] = useState(false)

  // 处方数据
  const [prescriptions, setPrescriptions] = useState<Prescription[]>([])
  const [selectedPrescription, setSelectedPrescription] = useState<Prescription | null>(null)
  const [sheetLoading, setSheetLoading] = useState(false)

  // 编辑表单状态
  const [editDialysisMode, setEditDialysisMode] = useState('')
  const [editFrequency, setEditFrequency] = useState('')
  const [editDuration, setEditDuration] = useState('')
  const [editAnticoagulant, setEditAnticoagulant] = useState('')
  const [editInitialDose, setEditInitialDose] = useState('')
  const [editMaintenanceDose, setEditMaintenanceDose] = useState('')
  const [editOrderItems, setEditOrderItems] = useState<PrescriptionOrderItem[]>([])

  // 加载医嘱数据
  const loadOrders = useCallback(async () => {
    if (!patient.id) return
    setLoading(true)
    try {
      const data = await orderApi.list(patient.id)
      const enriched = (data || []).map(toEnrichedOrder)
      setOrders(enriched.filter(o => o.status === '在用'))
    } catch (err) {
      message.error('加载医嘱失败: ' + (err instanceof Error ? err.message : '未知错误'))
    } finally {
      setLoading(false)
    }
  }, [patient.id])

  // 加载历史医嘱
  const loadHistoryOrders = useCallback(async () => {
    if (!patient.id) return
    try {
      const data = await orderApi.list(patient.id, { statuses: '已停止,已执行' })
      setHistoryOrders((data || []).map(toEnrichedOrder))
    } catch (err) {
      message.error('加载历史医嘱失败: ' + (err instanceof Error ? err.message : '未知错误'))
    }
  }, [patient.id])

  useEffect(() => {
    loadOrders()
  }, [loadOrders])

  useEffect(() => {
    if (showLongHistory || showTempHistory) {
      loadHistoryOrders()
    }
  }, [showLongHistory, showTempHistory, loadHistoryOrders])

  // 加载处方列表
  const loadPrescriptions = useCallback(async () => {
    if (!patient.id) return
    setSheetLoading(true)
    try {
      const data = await prescriptionApi.list(patient.id)
      const list = data || []
      setPrescriptions(list)
      // 如果当前选中项不在新列表中，重置到第一条
      if (list.length > 0) {
        const stillExists = selectedPrescriptionId && list.some(p => p.id === selectedPrescriptionId)
        if (!stillExists) {
          setSelectedPrescriptionId(list[0].id)
          setSelectedPrescription(list[0])
        } else {
          setSelectedPrescription(list.find(p => p.id === selectedPrescriptionId) || list[0])
        }
      } else {
        setSelectedPrescriptionId('')
        setSelectedPrescription(null)
      }
    } catch (err) {
      message.error('加载医嘱单失败: ' + (err instanceof Error ? err.message : '未知错误'))
    } finally {
      setSheetLoading(false)
    }
  }, [patient.id, selectedPrescriptionId])

  // 切换到 SHEET tab 时加载
  useEffect(() => {
    if (orderSubTab === 'SHEET') {
      loadPrescriptions()
    }
  }, [orderSubTab, loadPrescriptions])

  // 选中处方后加载详情
  const handleSelectPrescription = useCallback(async (id: string) => {
    setSelectedPrescriptionId(id)
    setIsEditingSheet(false)
    const found = prescriptions.find(p => p.id === id)
    if (found) {
      setSelectedPrescription(found)
    } else if (patient.id) {
      try {
        const detail = await prescriptionApi.get(patient.id, id)
        setSelectedPrescription(detail)
      } catch (err) {
        message.error('加载处方详情失败')
      }
    }
  }, [prescriptions, patient.id])

  // 进入编辑模式时预填表单
  const enterEditMode = () => {
    if (!selectedPrescription) return
    setEditDialysisMode(selectedPrescription.dialysisMode?.mode || 'HD')
    setEditFrequency(selectedPrescription.dialysisMode?.frequencyDesc || '3')
    setEditDuration(String(selectedPrescription.duration || 4))
    setEditAnticoagulant(selectedPrescription.anticoagulant?.initialDrug || '')
    setEditInitialDose(selectedPrescription.anticoagulant?.initialDose || '')
    setEditMaintenanceDose(selectedPrescription.anticoagulant?.maintenanceDose || '')
    setEditOrderItems(selectedPrescription.orderItems || [])
    setIsEditingSheet(true)
  }

  // 提取长嘱
  const handleExtractLongTermOrders = async () => {
    if (!patient.id) return
    const dateStr = new Date().toISOString().slice(0, 10)
    try {
      const p = await prescriptionApi.extract(patient.id, dateStr)
      message.success('提取长嘱成功')
      await loadPrescriptions()
      setSelectedPrescriptionId(p.id)
      setSelectedPrescription(p)
    } catch (err) {
      message.error('提取长嘱失败: ' + (err instanceof Error ? err.message : '未知错误'))
    }
  }

  // 保存草稿
  const handleSaveDraft = async () => {
    if (!patient.id || !selectedPrescription) return
    try {
      await prescriptionApi.update(patient.id, selectedPrescription.id, {
        duration: Number(editDuration) || undefined,
        dialysisMode: { ...selectedPrescription.dialysisMode, mode: editDialysisMode, frequencyDesc: editFrequency },
        anticoagulant: { ...selectedPrescription.anticoagulant, initialDrug: editAnticoagulant, initialDose: editInitialDose, maintenanceDose: editMaintenanceDose },
        orderItems: editOrderItems,
      })
      message.success('草稿已保存')
      loadPrescriptions()
    } catch (err) {
      message.error('保存失败: ' + (err instanceof Error ? err.message : '未知错误'))
    }
  }

  // 发布处方
  const handlePublish = async () => {
    if (!patient.id || !selectedPrescription) return
    try {
      // 先保存再发布
      await prescriptionApi.update(patient.id, selectedPrescription.id, {
        duration: Number(editDuration) || undefined,
        dialysisMode: { ...selectedPrescription.dialysisMode, mode: editDialysisMode, frequencyDesc: editFrequency },
        anticoagulant: { ...selectedPrescription.anticoagulant, initialDrug: editAnticoagulant, initialDose: editInitialDose, maintenanceDose: editMaintenanceDose },
        orderItems: editOrderItems,
      })
      await prescriptionApi.execute(patient.id, selectedPrescription.id)
      message.success('医嘱单已发布')
      setIsEditingSheet(false)
      loadPrescriptions()
    } catch (err) {
      message.error('发布失败: ' + (err instanceof Error ? err.message : '未知错误'))
    }
  }

  // 新增医嘱单
  const handleCreatePrescription = async () => {
    if (!patient.id) return
    const dateStr = new Date().toISOString().slice(0, 10)
    try {
      const p = await prescriptionApi.create(patient.id, { prescriptionDate: dateStr })
      message.success('新增医嘱单成功')
      await loadPrescriptions()
      setSelectedPrescriptionId(p.id)
      setSelectedPrescription(p)
      // 直接用返回的 p 预填编辑表单，不依赖 state 更新
      setEditDialysisMode(p.dialysisMode?.mode || 'HD')
      setEditFrequency(p.dialysisMode?.frequencyDesc || '3')
      setEditDuration(String(p.duration || 4))
      setEditAnticoagulant(p.anticoagulant?.initialDrug || '')
      setEditInitialDose(p.anticoagulant?.initialDose || '')
      setEditMaintenanceDose(p.anticoagulant?.maintenanceDose || '')
      setEditOrderItems(p.orderItems || [])
      setIsEditingSheet(true)
    } catch (err) {
      message.error('新增失败: ' + (err instanceof Error ? err.message : '未知错误'))
    }
  }

  const handleToggleSelection = (id: string, type: '长期' | '临时') => {
    if (type === '长期') {
      setSelectedLongIds(prev => (prev.includes(id) ? prev.filter(item => item !== id) : [...prev, id]))
      return
    }
    setSelectedTempIds(prev => (prev.includes(id) ? prev.filter(item => item !== id) : [...prev, id]))
  }

  const clearSelection = (type: '长期' | '临时') => {
    if (type === '长期') setSelectedLongIds([])
    else setSelectedTempIds([])
  }

  const refreshOrderViews = useCallback(async () => {
    await loadOrders()
    if (showLongHistory || showTempHistory) {
      await loadHistoryOrders()
    }
  }, [loadHistoryOrders, loadOrders, showLongHistory, showTempHistory])

  const handleGroup = async (type: '长期' | '临时') => {
    if (!patient.id) return
    const selected = type === '长期' ? selectedLongIds : selectedTempIds
    if (selected.length < 2) {
      message.warning('至少选择两条在用医嘱才能组合')
      return
    }
    try {
      await orderApi.group(patient.id, selected)
      message.success('组合成功')
      clearSelection(type)
      await refreshOrderViews()
    } catch (error) {
      message.error('组合失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  const handleUngroup = async (type: '长期' | '临时') => {
    if (!patient.id) return
    const selected = type === '长期' ? selectedLongIds : selectedTempIds
    if (selected.length === 0) {
      message.warning('请先选择要取消组合的医嘱')
      return
    }
    try {
      await orderApi.ungroup(patient.id, selected)
      message.success('取消组合成功')
      clearSelection(type)
      await refreshOrderViews()
    } catch (error) {
      message.error('取消组合失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  const handleOpenOrderModal = (type: '长期' | '临时') => {
    setOrderModalType(type)
    setEditOrder(undefined)
    setIsOrderModalOpen(true)
  }

  const handleEditOrder = (order: EnrichedOrder) => {
    if (!order._raw) return
    setOrderModalType(order.type)
    setEditOrder(order._raw)
    setIsOrderModalOpen(true)
  }

  const handleCopyOrder = async (order: EnrichedOrder) => {
    if (!patient.id) return
    try {
      await orderApi.copy(patient.id, order.id)
      message.success('已复制为新医嘱')
      await refreshOrderViews()
    } catch (error) {
      message.error('复制失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  const handleStopOrder = async (order: EnrichedOrder) => {
    if (!patient.id) return
    try {
      await orderApi.stop(patient.id, order.id)
      message.success('医嘱已停用')
      await refreshOrderViews()
    } catch (error) {
      message.error('停用失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  const handleOrderSaved = async () => {
    setIsOrderModalOpen(false)
    setEditOrder(undefined)
    await refreshOrderViews()
  }

  const buildDisplayGroups = (list: EnrichedOrder[]): OrderDisplayGroup[] => {
    const groups: OrderDisplayGroup[] = []
    const grouped = new Map<string, EnrichedOrder[]>()

    list.forEach(order => {
      if (!order.groupId) return
      const bucket = grouped.get(order.groupId) || []
      bucket.push(order)
      grouped.set(order.groupId, bucket)
    })

    const consumed = new Set<string>()
    list.forEach(order => {
      if (!order.groupId) {
        groups.push({ items: [order] })
        return
      }
      if (consumed.has(order.groupId)) return
      consumed.add(order.groupId)
      groups.push({
        groupId: order.groupId,
        items: grouped.get(order.groupId) || [order],
      })
    })

    return groups
  }

  const OrderTable = ({
    orders: tableOrders,
    historyList = [],
    type,
    selectedIds,
    showHistory = false,
  }: {
    orders: EnrichedOrder[]
    historyList?: EnrichedOrder[]
    type: '长期' | '临时'
    selectedIds: string[]
    showHistory?: boolean
  }) => {
    const isLong = type === '长期'
    const displayGroups = buildDisplayGroups(tableOrders)
    let displayIndex = 0

    const renderActiveRows = () => displayGroups.flatMap(group => {
      displayIndex += 1
      return group.items.map((order, index) => (
        <tr key={order.id} className="hover:bg-slate-50/50 transition-colors group">
          <td className={`py-3 px-2 text-center align-middle border-r border-slate-50 ${group.groupId ? 'bg-blue-50/10' : ''}`}>
            <input
              type="checkbox"
              checked={selectedIds.includes(order.id)}
              onChange={() => handleToggleSelection(order.id, type)}
              className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-0 cursor-pointer"
            />
          </td>
          {index === 0 && (
            <td rowSpan={group.items.length} className={`py-3 px-2 font-bold text-center align-middle border-r border-slate-50 text-slate-400 ${group.groupId ? 'border-r-2 border-blue-400 bg-blue-50/10' : ''}`}>
              {displayIndex}
            </td>
          )}
          <td className="py-3 px-3 font-black text-blue-600">{order.name}</td>
          <td className="py-3 px-2 text-center font-mono font-bold text-slate-700">{order.dose || '-'}</td>
          <td className="py-3 px-2 text-center text-slate-500 text-[10px] font-black">{order.unit || '-'}</td>
          <td className="py-3 px-2 text-slate-600 font-medium">{order.route || '-'}</td>
          {isLong ? (
            <>
              <td className="py-3 px-2 text-slate-600 font-medium">{order.frequency || '-'}</td>
              <td className="py-3 px-2 text-slate-600 font-medium">{order.timing || '-'}</td>
            </>
          ) : (
            <td className="py-3 px-2 text-emerald-600 font-black">{order.execTiming || '立即执行'}</td>
          )}
          <td className="py-3 px-2 font-mono text-[10px] text-slate-400 whitespace-nowrap">{order.startTime}</td>
          {isLong && <td className="py-3 px-2 font-mono text-[10px] text-red-400 whitespace-nowrap">{order.stopTime || '-'}</td>}
          <td className="py-3 px-2 text-slate-400 text-[11px] leading-relaxed italic truncate max-w-[120px]" title={order.remark}>{order.remark || '-'}</td>
          <td className="py-3 px-2 font-bold text-slate-700">{order.doctor}</td>
          <td className="py-3 px-2 text-right sticky right-0 bg-white group-hover:bg-slate-50 transition-colors shadow-[-4px_0_8px_rgba(0,0,0,0.05)]">
            <div className="flex justify-end gap-1.5">
              <button onClick={() => handleEditOrder(order)} className="p-1 text-slate-300 hover:text-blue-600 transition-all" title="编辑"><Edit3 size={14} /></button>
              <button onClick={() => handleStopOrder(order)} className="p-1 text-slate-300 hover:text-orange-500 transition-all" title="停用"><PowerOff size={14} /></button>
            </div>
          </td>
        </tr>
      ))
    })

    const renderHistoryRows = () => historyList.map((order, index) => (
      <tr key={`history-${order.id}`} className="bg-slate-50/30 hover:bg-slate-50 group">
        <td className="py-3 px-2 text-center align-middle border-r border-slate-50">
          <span className="text-slate-300">-</span>
        </td>
        <td className="py-3 px-2 font-bold text-center align-middle border-r border-slate-50 text-slate-400">
          H{index + 1}
        </td>
        <td className="py-3 px-3 font-black text-red-500 italic line-through opacity-70">{order.name}</td>
        <td className="py-3 px-2 text-center font-mono font-bold text-slate-500">{order.dose || '-'}</td>
        <td className="py-3 px-2 text-center text-slate-400 text-[10px] font-black">{order.unit || '-'}</td>
        <td className="py-3 px-2 text-slate-500 font-medium">{order.route || '-'}</td>
        {isLong ? (
          <>
            <td className="py-3 px-2 text-slate-500 font-medium">{order.frequency || '-'}</td>
            <td className="py-3 px-2 text-slate-500 font-medium">{order.timing || '-'}</td>
          </>
        ) : (
          <td className="py-3 px-2 text-emerald-500 font-bold">{order.execTiming || '立即执行'}</td>
        )}
        <td className="py-3 px-2 font-mono text-[10px] text-slate-400 whitespace-nowrap">{order.startTime}</td>
        {isLong && <td className="py-3 px-2 font-mono text-[10px] text-red-400 whitespace-nowrap">{order.stopTime || '-'}</td>}
        <td className="py-3 px-2 text-slate-400 text-[11px] leading-relaxed italic truncate max-w-[120px]" title={order.remark}>{order.remark || '-'}</td>
        <td className="py-3 px-2 font-bold text-slate-600">{order.doctor}</td>
        <td className="py-3 px-2 text-right sticky right-0 bg-slate-50/30 group-hover:bg-slate-50 transition-colors shadow-[-4px_0_8px_rgba(0,0,0,0.05)]">
          <div className="flex justify-end gap-1.5">
            <button onClick={() => handleCopyOrder(order)} className="p-1 text-slate-300 hover:text-blue-600 transition-all" title="复制"><Copy size={14} /></button>
          </div>
        </td>
      </tr>
    ))

    const colSpan = isLong ? 13 : 11

    return (
      <div className="overflow-x-auto custom-scrollbar-horizontal scroll-smooth border border-slate-100 rounded-2xl pb-2">
        <table className="w-full text-left text-xs border-separate border-spacing-0" style={{ minWidth: isLong ? '1040px' : '920px' }}>
          <thead className="bg-[#f8faff] text-slate-400 font-bold uppercase tracking-widest text-[9px] sticky top-0 z-10 shadow-sm border-b border-slate-100">
            <tr>
              <th className="py-3 px-2 w-10 text-center">选择</th>
              <th className="py-3 px-2 w-10 text-center">序号</th>
              <th className="py-3 px-3 w-36">药品 / 方法</th>
              <th className="py-3 px-2 w-12 text-center">剂量</th>
              <th className="py-3 px-2 w-12 text-center">单位</th>
              <th className="py-3 px-2 w-16">用法</th>
              {isLong ? (
                <>
                  <th className="py-3 px-2 w-16">频次</th>
                  <th className="py-3 px-2 w-16">使用时机</th>
                </>
              ) : (
                <th className="py-3 px-2 w-16">执行时机</th>
              )}
              <th className="py-3 px-2 w-24">医嘱时间</th>
              {isLong && <th className="py-3 px-2 w-24">停用日期</th>}
              <th className="py-3 px-2 w-32">备注</th>
              <th className="py-3 px-2 w-14">医生</th>
              <th className="py-3 px-2 w-14 text-right sticky right-0 bg-[#f8faff] shadow-[-4px_0_8px_rgba(0,0,0,0.05)]">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {tableOrders.length > 0 ? renderActiveRows() : (
              <tr><td colSpan={colSpan} className="py-12 text-center text-slate-300 font-bold italic">暂无医嘱数据</td></tr>
            )}
            {showHistory && historyList.length > 0 && renderHistoryRows()}
          </tbody>
        </table>
      </div>
    )
  }

  const renderDialysisOrders = () => {
    const activeLong = orders.filter(order => order.type === '长期')
    const activeTemp = orders.filter(order => order.type === '临时')
    const historyLong = historyOrders.filter(order => order.type === '长期')
    const historyTemp = historyOrders.filter(order => order.type === '临时')

    if (loading) {
      return (
        <div className="flex justify-center items-center py-20">
          <Loader2 size={32} className="animate-spin text-blue-400" />
          <span className="ml-3 text-slate-400 font-bold">加载医嘱数据...</span>
        </div>
      )
    }

    return (
      <div className="space-y-6 animate-fade-in pb-20">
        <DetailCard className="!p-0 overflow-hidden border-slate-200">
          <div className="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-white">
            <h4 className="font-black text-slate-800 flex items-center gap-2 text-lg">
              <div className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center"><Pill size={20} className="text-blue-600" /></div>
              长期医嘱
            </h4>
            <div className="flex items-center gap-3">
              <div className="flex bg-slate-100 p-1 rounded-xl">
                <button onClick={() => handleGroup('长期')} className="px-4 py-1.5 bg-white border border-slate-200 text-blue-600 rounded-lg text-xs font-bold shadow-sm flex items-center gap-1.5 hover:bg-blue-50 transition-all"><Link size={12} /> 组合</button>
                <button onClick={() => handleUngroup('长期')} className="px-4 py-1.5 text-slate-500 rounded-lg text-xs font-bold flex items-center gap-1.5 hover:text-slate-700 transition-all"><RotateCcw size={12} /> 取消组合</button>
              </div>
              <button onClick={() => handleOpenOrderModal('长期')} className="px-5 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-black shadow-lg shadow-blue-100 hover:bg-blue-700 transition-all flex items-center gap-2"><PlusCircle size={18} /> 新增医嘱</button>
            </div>
          </div>
          <OrderTable orders={activeLong} type="长期" selectedIds={selectedLongIds} showHistory={showLongHistory} historyList={historyLong} />
          <div className="py-4 border-t border-slate-50 flex flex-col items-center gap-3 bg-slate-50/30">
            <button onClick={() => setShowLongHistory(prev => !prev)} className="flex items-center gap-2 text-xs font-bold text-slate-400 hover:text-blue-600 transition-colors group">
              <History size={16} className={`${showLongHistory ? 'text-blue-600' : ''} group-hover:rotate-[-10deg] transition-transform`} />
              <span>{showLongHistory ? '收起历史库' : '展开历史库'}</span>
            </button>
          </div>
        </DetailCard>

        <DetailCard className="!p-0 overflow-hidden border-slate-200 shadow-sm">
          <div className="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-white">
            <h4 className="font-black text-slate-800 flex items-center gap-2 text-lg">
              <div className="w-10 h-10 rounded-full bg-amber-50 flex items-center justify-center"><Zap size={20} className="text-amber-500" /></div>
              临时医嘱
            </h4>
            <div className="flex items-center gap-3">
              <div className="flex bg-slate-100 p-1 rounded-xl">
                <button onClick={() => handleGroup('临时')} className="px-4 py-1.5 bg-white border border-slate-200 text-blue-600 rounded-lg text-xs font-bold shadow-sm flex items-center gap-1.5 hover:bg-blue-50 transition-all"><Link size={12} /> 组合</button>
                <button onClick={() => handleUngroup('临时')} className="px-4 py-1.5 text-slate-500 rounded-lg text-xs font-bold flex items-center gap-1.5 hover:text-slate-700 transition-all"><RotateCcw size={12} /> 取消组合</button>
              </div>
              <button onClick={() => handleOpenOrderModal('临时')} className="px-5 py-2.5 bg-amber-500 text-white rounded-xl text-sm font-black shadow-lg shadow-amber-100 hover:bg-amber-600 transition-all flex items-center gap-2"><PlusCircle size={18} /> 新增临时医嘱</button>
            </div>
          </div>
          <OrderTable orders={activeTemp} type="临时" selectedIds={selectedTempIds} showHistory={showTempHistory} historyList={historyTemp} />
          <div className="py-4 border-t border-slate-50 flex flex-col items-center gap-3 bg-slate-50/30">
            <button onClick={() => setShowTempHistory(prev => !prev)} className="flex items-center gap-2 text-xs font-bold text-slate-400 hover:text-amber-600 transition-colors group">
              <History size={16} className={`${showTempHistory ? 'text-amber-600' : ''} group-hover:rotate-[-10deg] transition-transform`} />
              <span>{showTempHistory ? '收起历史库' : '展开历史库'}</span>
            </button>
          </div>
        </DetailCard>
      </div>
    )
  }

  // 按 category 分组 orderItems（保留原始索引用于编辑回写）
  type IndexedItem = PrescriptionOrderItem & { _idx: number }
  const groupOrderItems = (items: PrescriptionOrderItem[]) => {
    const groups: { category: string; items: IndexedItem[] }[] = []
    items.forEach((item, idx) => {
      const cat = item.category || '其他'
      const indexed = { ...item, _idx: idx }
      const existing = groups.find(g => g.category === cat)
      if (existing) existing.items.push(indexed)
      else groups.push({ category: cat, items: [indexed] })
    })
    return groups
  }

  // 编辑医嘱单表单
  const renderEditForm = () => {
    const groups = groupOrderItems(editOrderItems)
    const prescriptionDate = selectedPrescription?.prescriptionDate
      ? new Date(selectedPrescription.prescriptionDate).toISOString().slice(0, 10)
      : new Date().toISOString().slice(0, 10)

    const handleItemChange = (index: number, field: 'dose' | 'frequency', value: string) => {
      setEditOrderItems(prev => prev.map((item, i) => i === index ? { ...item, [field]: value } : item))
    }

    const handleDeleteItem = (index: number) => {
      setEditOrderItems(prev => prev.filter((_, i) => i !== index))
    }

    return (
      <div className="flex-1 space-y-4 animate-fade-in">
        <div className="bg-white rounded-3xl border border-slate-200 p-6 shadow-sm space-y-6">
          {/* 顶部控制栏 */}
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <span className="text-sm font-bold text-slate-500">医嘱日期:</span>
                <div className="flex items-center bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-200">
                  <span className="text-sm font-mono font-bold w-28">{prescriptionDate}</span>
                  <Calendar size={14} className="text-slate-300 ml-2"/>
                </div>
              </div>
              <span className="px-2 py-0.5 bg-green-50 text-green-600 text-[10px] font-black rounded border border-green-100">正在编辑医嘱单</span>
            </div>
            <div className="flex gap-2">
              <button onClick={handleCreatePrescription} className="px-4 py-2 bg-blue-600 text-white text-xs font-black rounded-xl shadow-lg hover:bg-blue-700 flex items-center gap-1.5">
                <Plus size={14}/> 新增医嘱单
              </button>
              <button onClick={handleExtractLongTermOrders} className="px-4 py-2 bg-indigo-600 text-white text-xs font-black rounded-xl shadow-lg hover:bg-indigo-700">提取长嘱</button>
              <button onClick={handleSaveDraft} className="px-4 py-2 bg-slate-200 text-slate-600 text-xs font-black rounded-xl hover:bg-slate-300 flex items-center gap-1.5">
                <Save size={14}/> 保存草稿
              </button>
              <button onClick={handlePublish} className="px-4 py-2 bg-blue-500 text-white text-xs font-black rounded-xl shadow-lg hover:bg-blue-600 flex items-center gap-1.5">
                <Send size={14}/> 发布
              </button>
            </div>
          </div>

          {/* 计划区块 */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-x-12 gap-y-4 pt-4 border-t border-slate-50">
            <div className="flex items-center gap-3">
              <span className="text-sm font-bold text-slate-600 whitespace-nowrap">透析计划:透析模式</span>
              <input type="text" value={editDialysisMode} onChange={e => setEditDialysisMode(e.target.value)} className="flex-1 h-9 px-3 border border-slate-200 rounded-lg text-sm font-bold bg-slate-50/30"/>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm font-bold text-slate-600 whitespace-nowrap">频率:</span>
              <div className="flex-1 flex items-center bg-slate-50/30 border border-slate-200 rounded-lg pr-3">
                <input type="text" value={editFrequency} onChange={e => setEditFrequency(e.target.value)} className="flex-1 h-9 px-3 bg-transparent text-sm font-bold outline-none"/>
                <span className="text-xs text-slate-400 font-bold">次/周</span>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm font-bold text-slate-600 whitespace-nowrap">透析时间:</span>
              <div className="flex-1 flex items-center bg-slate-50/30 border border-slate-200 rounded-lg pr-3">
                <input type="text" value={editDuration} onChange={e => setEditDuration(e.target.value)} className="flex-1 h-9 px-3 bg-transparent text-sm font-bold outline-none"/>
                <span className="text-xs text-slate-400 font-bold">时/次</span>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm font-bold text-slate-600 whitespace-nowrap">抗凝计划:抗凝剂种</span>
              <input type="text" value={editAnticoagulant} onChange={e => setEditAnticoagulant(e.target.value)} className="flex-1 h-9 px-3 border border-slate-200 rounded-lg text-sm font-bold bg-slate-50/30"/>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm font-bold text-slate-600 whitespace-nowrap">首剂:</span>
              <div className="flex-1 flex items-center bg-slate-50/30 border border-slate-200 rounded-lg pr-3">
                <input type="text" value={editInitialDose} onChange={e => setEditInitialDose(e.target.value)} className="flex-1 h-9 px-3 bg-transparent text-sm font-bold outline-none"/>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm font-bold text-slate-600 whitespace-nowrap">追加:</span>
              <div className="flex-1 flex items-center bg-slate-50/30 border border-slate-200 rounded-lg pr-3">
                <input type="text" value={editMaintenanceDose} onChange={e => setEditMaintenanceDose(e.target.value)} className="flex-1 h-9 px-3 bg-transparent text-sm font-bold outline-none"/>
              </div>
            </div>
          </div>

          {/* 药品表格 */}
          <div className="mt-8 border border-slate-100 rounded-2xl overflow-hidden shadow-sm">
            <table className="w-full text-left text-sm border-collapse">
              <thead className="bg-[#eef6ff] text-slate-700 font-black border-b border-slate-100">
                <tr>
                  <th className="py-4 px-6 w-40">种类</th>
                  <th className="py-4 px-4 w-64">药物</th>
                  <th className="py-4 px-4">剂量</th>
                  <th className="py-4 px-4">频次</th>
                  <th className="py-4 px-6 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50">
                {editOrderItems.length === 0 ? (
                  <tr><td colSpan={5} className="py-12 text-center text-slate-300 font-bold italic">暂无药品数据，点击"提取长嘱"导入</td></tr>
                ) : (
                  groups.map((group) => {
                    return group.items.map((item, dIdx) => {
                      return (
                        <tr key={`${group.category}-${dIdx}`} className="hover:bg-slate-50/30 group/row">
                          {dIdx === 0 && (
                            <td rowSpan={group.items.length} className="py-4 px-6 font-bold text-slate-800 bg-slate-50/30 align-top">
                              {group.category}
                            </td>
                          )}
                          <td className="py-4 px-4 font-medium text-slate-700">{item.name}</td>
                          <td className="py-4 px-4">
                            <input type="text" value={item.dose} onChange={e => handleItemChange(item._idx, 'dose', e.target.value)} className="w-24 h-8 px-2 border border-slate-100 rounded bg-transparent focus:bg-white focus:border-blue-200 outline-none font-mono text-slate-500"/>
                          </td>
                          <td className="py-4 px-4">
                            <input type="text" value={item.frequency} onChange={e => handleItemChange(item._idx, 'frequency', e.target.value)} className="w-24 h-8 px-2 border border-slate-100 rounded bg-transparent focus:bg-white focus:border-blue-200 outline-none font-bold text-slate-600"/>
                          </td>
                          <td className="py-4 px-6 text-right">
                            <button onClick={() => handleDeleteItem(item._idx)} className="text-red-400 hover:underline text-xs font-bold">删除</button>
                          </td>
                        </tr>
                      )
                    })
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    )
  }

  // 医嘱单预览视图 (A4)
  const renderPreviewSheet = () => {
    if (!selectedPrescription) {
      return (
        <div className="flex-1 flex justify-center items-center py-20">
          <span className="text-slate-300 font-bold italic">请选择一条医嘱单</span>
        </div>
      )
    }

    const p = selectedPrescription
    const groups = groupOrderItems(p.orderItems || [])
    const prescriptionDate = p.prescriptionDate ? new Date(p.prescriptionDate) : new Date()
    const dateStr = `${prescriptionDate.getFullYear()} 年 ${prescriptionDate.getMonth() + 1} 月 ${prescriptionDate.getDate()} 日`
    const canEdit = p.status === '待执行'

    return (
      <div className="flex flex-col items-center bg-slate-100/50 p-10 min-h-full space-y-6 animate-fade-in">
        {/* 操作工具栏 */}
        <div className="flex gap-4 w-[210mm] justify-end no-print">
          {canEdit && (
            <button
              onClick={enterEditMode}
              className="flex items-center gap-2 px-6 py-2 bg-white border border-slate-200 text-slate-700 rounded-xl text-sm font-black hover:bg-slate-50 shadow-sm transition-all"
            >
              <Edit3 size={16} className="text-blue-500" /> 编辑医嘱单
            </button>
          )}
          <button
            onClick={() => window.print()}
            className="flex items-center gap-2 px-6 py-2 bg-blue-600 text-white rounded-xl text-sm font-black hover:bg-blue-700 shadow-lg shadow-blue-200 transition-all"
          >
            <Printer size={16} /> 打印医嘱单
          </button>
        </div>

        {/* 状态标签 */}
        <div className="w-[210mm] flex justify-start">
          <span className={`px-3 py-1 text-xs font-black rounded-full border ${
            p.status === '待执行' ? 'bg-yellow-50 text-yellow-600 border-yellow-200' :
            p.status === '执行中' ? 'bg-blue-50 text-blue-600 border-blue-200' :
            p.status === '已执行' ? 'bg-green-50 text-green-600 border-green-200' :
            'bg-red-50 text-red-500 border-red-200'
          }`}>{p.status}</span>
        </div>

        {/* A4 预览容器 */}
        <div className="bg-white shadow-2xl w-[210mm] min-h-[297mm] p-[20mm] flex flex-col font-serif border border-slate-200 print:shadow-none print:m-0 print:border-none">
          {/* 标题 */}
          <h1 className="text-2xl font-bold text-center mb-10 tracking-widest">门诊患者医嘱单</h1>

          {/* 基本信息 */}
          <div className="flex justify-between items-center mb-6 text-sm font-bold border-b border-slate-300 pb-2">
            <span>姓名：{patient.name}</span>
            <span>性别：{patient.gender}</span>
            <span>年龄：{patient.age}岁</span>
            <span>主管医师：{p.doctorName || patient.doctorName}</span>
          </div>

          {/* 计划描述 */}
          <div className="space-y-2 mb-6 text-sm leading-relaxed">
            <p>1. 透析计划: 透析模式: {p.dialysisMode?.mode || '-'}, 频率: {p.dialysisMode?.frequencyDesc || '-'} 次/周, 透析时间: {p.duration || '-'} 时/次</p>
            <p>2. 抗凝计划: 抗凝剂种类: {p.anticoagulant?.initialDrug || '-'}, 首剂: {p.anticoagulant?.initialDose || '-'}, 追加: {p.anticoagulant?.maintenanceDose || '-'}</p>
            <p>3. 药物治疗:</p>
          </div>

          {/* 核心药品表格 */}
          <table className="w-full border-collapse border border-black text-[13px]">
            <thead>
              <tr>
                <th className="border border-black p-2 w-32 bg-slate-50/50">种类</th>
                <th className="border border-black p-2 bg-slate-50/50">药物</th>
                <th className="border border-black p-2 w-32 bg-slate-50/50 text-center">剂量</th>
                <th className="border border-black p-2 w-32 bg-slate-50/50 text-center">频次</th>
              </tr>
            </thead>
            <tbody>
              {groups.length === 0 ? (
                <tr><td colSpan={4} className="border border-black p-4 text-center text-slate-400 italic">暂无药品数据</td></tr>
              ) : (
                groups.map((cat) =>
                  cat.items.map((item, itemIdx) => (
                    <tr key={`${cat.category}-${itemIdx}`}>
                      {itemIdx === 0 && (
                        <td rowSpan={cat.items.length} className="border border-black p-2 text-center align-middle font-bold bg-slate-50/10">
                          {cat.category}
                        </td>
                      )}
                      <td className="border border-black p-2 min-h-[30px]">{item.name}</td>
                      <td className="border border-black p-2 text-center font-mono">{item.dose}</td>
                      <td className="border border-black p-2 text-center">{item.frequency}</td>
                    </tr>
                  ))
                )
              )}
            </tbody>
          </table>

          {/* 页脚签名 */}
          <div className="mt-auto space-y-4 pt-10">
            <div className="flex justify-between items-end text-sm">
              <span>联系电话：0531-8888****</span>
              <div className="flex flex-col items-end gap-2">
                <div className="flex items-end gap-2">
                  <span className="font-bold">医师签名：</span>
                  <span className="text-xl font-bold italic font-serif border-b border-black w-32 text-center pb-0.5">{p.doctorName || patient.doctorName}</span>
                </div>
                <span className="mt-2 tracking-tight">日期：{dateStr}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // 医嘱单视图（包含日期时间轴 + 编辑/预览切换）
  const renderOrderSheet = () => {
    if (sheetLoading) {
      return (
        <div className="flex justify-center items-center py-20">
          <Loader2 size={32} className="animate-spin text-blue-400" />
          <span className="ml-3 text-slate-400 font-bold">加载医嘱单...</span>
        </div>
      )
    }

    if (prescriptions.length === 0) {
      return (
        <div className="flex flex-col justify-center items-center py-20 gap-4">
          <span className="text-slate-300 font-bold italic text-lg">暂无医嘱单</span>
          <div className="flex gap-3">
            <button onClick={handleCreatePrescription} className="px-5 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-black shadow-lg hover:bg-blue-700 flex items-center gap-2">
              <Plus size={16}/> 新增医嘱单
            </button>
            <button onClick={handleExtractLongTermOrders} className="px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-black shadow-lg hover:bg-indigo-700">
              提取长嘱
            </button>
          </div>
        </div>
      )
    }

    return (
      <div className="flex gap-4 animate-fade-in pb-10 h-full">
        {/* 左侧日期时间轴 */}
        <div className="w-24 shrink-0 flex flex-col gap-2 pt-2 no-print">
          {prescriptions.map((p) => {
            const dateLabel = p.prescriptionDate
              ? new Date(p.prescriptionDate).toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
              : '未知'
            const statusColor = p.status === '待执行' ? 'text-yellow-500' : p.status === '已执行' ? 'text-green-500' : p.status === '已取消' ? 'text-red-400' : 'text-blue-500'
            return (
              <button
                key={p.id}
                onClick={() => handleSelectPrescription(p.id)}
                className={`w-full py-2.5 rounded-lg text-xs font-black border transition-all flex flex-col items-center gap-0.5 ${
                  selectedPrescriptionId === p.id
                    ? 'bg-white border-blue-200 text-blue-600 shadow-sm'
                    : 'bg-slate-50 border-transparent text-slate-400 hover:bg-white hover:border-slate-100'
                }`}
              >
                <span>{dateLabel}</span>
                <span className={`text-[9px] ${statusColor}`}>{p.status}</span>
              </button>
            )
          })}
        </div>

        {/* 右侧内容区 */}
        {isEditingSheet ? renderEditForm() : renderPreviewSheet()}
      </div>
    )
  }

  return (
    <div className="space-y-8 animate-fade-in pb-10 h-full">
      {/* 顶部二级导航 */}
      <div className="flex justify-center no-print">
        <div className="flex bg-slate-200/50 p-1 rounded-2xl shadow-inner ring-1 ring-slate-100">
          <button
            onClick={() => setOrderSubTab('DIALYSIS')}
            className={`flex items-center px-8 py-2.5 rounded-xl text-sm font-black transition-all duration-300 ${
              orderSubTab === 'DIALYSIS'
              ? 'bg-white text-blue-600 shadow-lg ring-1 ring-blue-50'
              : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            <ClipboardList size={16} className="mr-2"/> 透析医嘱
          </button>
          <button
            onClick={() => setOrderSubTab('SHEET')}
            className={`flex items-center px-8 py-2.5 rounded-xl text-sm font-black transition-all duration-300 ${
              orderSubTab === 'SHEET'
              ? 'bg-white text-blue-600 shadow-lg ring-1 ring-blue-50'
              : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            <ScrollText size={16} className="mr-2"/> 医嘱单
          </button>
        </div>
      </div>

      <div className="flex-1">
        {orderSubTab === 'DIALYSIS' ? renderDialysisOrders() : renderOrderSheet()}
      </div>

      {/* 新增/编辑医嘱弹窗 */}
      <OrderModal
        isOpen={isOrderModalOpen}
        onClose={() => { setIsOrderModalOpen(false); setEditOrder(undefined) }}
        type={orderModalType}
        patientId={patient.id}
        editOrder={editOrder}
        onSave={handleOrderSaved}
      />
    </div>
  )
}
