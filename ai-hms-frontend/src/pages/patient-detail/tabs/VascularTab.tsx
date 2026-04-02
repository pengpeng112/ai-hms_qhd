// Vascular Access Tab - 血管通路评估

import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Syringe, Activity, Edit3, Wrench, Trash2, ShieldX, Loader2, Plus, ClipboardList, X } from 'lucide-react'
import { message, Modal } from 'antd'
import { SectionHeader, DetailCard, LabelValue } from '@/components/ui'
import { VascularAccessModal, VascularInterventionModal, type InterventionFormData } from '@/components/patient/modals'
import { restApi, type VascularAccessApi, type VascularAccessInterventionApi, type VascularAccessInterventionCreateRequest } from '@/services/restClient'
import { dictCache, DICT_TYPES } from '@/services/dictApi'
import type { VascularFormData } from '@/components/patient/modals/VascularAccessModal'
import type { Patient } from '@/types/original'
import type { VascularRecord } from '../types'

interface VascularTabProps {
  patient: Patient
}

// API 数据转换为组件使用的格式
function convertApiToRecord(api: VascularAccessApi): VascularRecord {
  return {
    id: api.id,
    accessType: api.accessType,
    site: api.site,
    artery: api.artery || [],
    vein: api.vein || [],
    side: api.side,
    hospital: api.hospital,
    surgeon: api.surgeon,
    surgeryDate: api.surgeryDate,
    firstUseDate: api.firstUseDate,
    accessNumber: api.accessNumber,
    interventionCount: api.interventionCount,
    interventionDate: api.interventionDate,
    catheterMethod: api.catheterMethod,
    catheterDepth: api.catheterDepth,
    vPuncturePosition: api.vPuncturePosition || [],
    aPuncturePosition: api.aPuncturePosition || [],
    notes: api.notes,
    images: api.images || [],
    isDefault: api.isDefault,
    isDisabled: api.isDisabled,
    createdAt: api.createdAt,
  }
}

// 组件表单数据转换为 API 请求格式
function convertFormToApi(form: VascularFormData) {
  return {
    accessType: form.accessType,
    site: form.site,
    artery: form.artery,
    vein: form.vein,
    side: form.side,
    hospital: form.hospital,
    surgeon: form.surgeon,
    surgeryDate: form.surgeryDate,
    firstUseDate: form.firstUseDate,
    accessNumber: form.accessNumber,
    interventionCount: form.interventionCount,
    interventionDate: form.interventionDate,
    catheterMethod: form.catheterMethod || undefined,
    catheterDepth: form.catheterDepth || undefined,
    vPuncturePosition: form.vPuncturePosition,
    aPuncturePosition: form.aPuncturePosition,
    notes: form.remark,
    images: form.images,
    isDefault: form.isDefault,
    isDisabled: form.isDisabled,
  }
}

// Record 转换为表单初始数据
function convertRecordToForm(record: VascularRecord): Partial<VascularFormData> {
  return {
    accessType: record.accessType,
    site: record.site,
    artery: record.artery,
    vein: record.vein,
    side: record.side,
    hospital: record.hospital,
    surgeon: record.surgeon,
    surgeryDate: record.surgeryDate,
    firstUseDate: record.firstUseDate,
    accessNumber: record.accessNumber,
    interventionCount: record.interventionCount,
    interventionDate: record.interventionDate,
    catheterMethod: record.catheterMethod || '',
    catheterDepth: record.catheterDepth || '',
    vPuncturePosition: record.vPuncturePosition,
    aPuncturePosition: record.aPuncturePosition,
    remark: record.notes,
    images: record.images,
    isDefault: record.isDefault,
    isDisabled: record.isDisabled,
  }
}

export default function VascularTab({ patient }: VascularTabProps) {
  const { t } = useTranslation('patient')

  // 状态管理
  const [isVascularModalOpen, setIsVascularModalOpen] = useState(false)
  const [isInterventionModalOpen, setIsInterventionModalOpen] = useState(false)
  const [isInterventionDetailModalOpen, setIsInterventionDetailModalOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [vascularHistory, setVascularHistory] = useState<VascularRecord[]>([])
  const [interventionRecords, setInterventionRecords] = useState<VascularAccessInterventionApi[]>([])
  const [editingRecord, setEditingRecord] = useState<VascularRecord | null>(null)
  const [selectedIntervention, setSelectedIntervention] = useState<VascularAccessInterventionApi | null>(null)
  const [modalMode, setModalMode] = useState<'create' | 'edit'>('create')

  // 字典选项映射（代码 -> 名称）
  const [dictNames, setDictNames] = useState<Record<string, Record<string, string>>>({})

  // 加载血管通路数据
  const loadVascularAccesses = useCallback(async () => {
    if (!patient.id) return

    setLoading(true)
    try {
      const data = await restApi.getVascularAccesses(patient.id)
      setVascularHistory(data.map(convertApiToRecord))
    } catch (error) {
      console.error('加载血管通路数据失败:', error)
      message.error('加载血管通路数据失败')
    } finally {
      setLoading(false)
    }
  }, [patient.id])

  // 加载干预记录数据
  const loadInterventionRecords = useCallback(async () => {
    if (!patient.id) return

    try {
      const data = await restApi.getVascularAccessInterventions(patient.id)
      setInterventionRecords(data)
    } catch (error) {
      console.error('加载干预记录失败:', error)
      // 不显示错误消息，仅记录日志
    }
  }, [patient.id])

  useEffect(() => {
    loadVascularAccesses()
    loadInterventionRecords()
  }, [loadVascularAccesses, loadInterventionRecords])

  // 加载字典数据（用于显示名称）
  useEffect(() => {
    const loadDictNames = async () => {
      try {
        const dictTypes = [
          DICT_TYPES.VASCULAR_ACCESS,
          DICT_TYPES.VASCULAR_SITE,
          DICT_TYPES.ARTERY_TYPE,
          DICT_TYPES.VEIN_TYPE,
          DICT_TYPES.HOSPITAL,
          DICT_TYPES.DOCTOR,
          DICT_TYPES.SURGERY_TYPE,
        ] as const

        const results = await Promise.all(
          dictTypes.map(type => dictCache.getOptions(type).catch(() => []))
        )

        const maps: Record<string, Record<string, string>> = {}
        dictTypes.forEach((type, index) => {
          const map: Record<string, string> = {}
          results[index].forEach(item => {
            map[item.value] = item.label
          })
          maps[type] = map
        })

        setDictNames(maps)
      } catch (error) {
        console.error('加载字典数据失败:', error)
      }
    }

    loadDictNames()
  }, [])

  // 当前启用的通路（未禁用且为默认）
  const activeAccess = vascularHistory.find(v => !v.isDisabled && v.isDefault)

  const hasInterventionRecords = interventionRecords.length > 0

  // 打开新增弹窗
  const handleOpenCreate = () => {
    setEditingRecord(null)
    setModalMode('create')
    setIsVascularModalOpen(true)
  }

  // 打开编辑弹窗
  const handleOpenEdit = (record: VascularRecord) => {
    setEditingRecord(record)
    setModalMode('edit')
    setIsVascularModalOpen(true)
  }

  // 保存血管通路
  const handleSave = async (formData: VascularFormData): Promise<void> => {
    setSaving(true)
    try {
      const apiData = convertFormToApi(formData)

      if (modalMode === 'edit' && editingRecord) {
        await restApi.updateVascularAccess(patient.id, editingRecord.id, apiData)
        message.success('更新成功')
      } else {
        await restApi.createVascularAccess(patient.id, apiData)
        message.success('创建成功')
      }

      await loadVascularAccesses()
      setIsVascularModalOpen(false)
    } catch (error) {
      console.error('保存血管通路失败:', error)
      message.error('保存失败')
      throw error
    } finally {
      setSaving(false)
    }
  }

  // 删除血管通路
  const handleDelete = (record: VascularRecord) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除这条血管通路记录吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await restApi.deleteVascularAccess(patient.id, record.id)
          message.success('删除成功')
          await loadVascularAccesses()
        } catch (error) {
          console.error('删除血管通路失败:', error)
          message.error('删除失败')
        }
      },
    })
  }

  // 字典 code 转名称
  const dictName = (type: string, code: string | undefined) => {
    if (!code) return '-'
    return dictNames[type]?.[code] || code
  }

  // 字典 code 数组转名称
  const dictArrayNames = (type: string, codes: string[] | undefined) => {
    if (!codes || codes.length === 0) return '-'
    return codes.map(code => dictNames[type]?.[code] || code).join('、')
  }

  // 格式化左右
  const formatSide = (side: string) => {
    if (side === 'L') return '左'
    if (side === 'R') return '右'
    return side || '-'
  }

  // 保存干预记录
  const handleSaveIntervention = async (formData: InterventionFormData): Promise<void> => {
    if (!activeAccess) {
      message.error('没有选择血管通路')
      return
    }

    setSaving(true)
    try {
      const requestData: VascularAccessInterventionCreateRequest = {
        vascularAccessId: activeAccess.id,
        accessType: formData.accessType || activeAccess.accessType,
        avgBloodFlow: formData.avgBloodFlow,
        usageDays: formData.usageDays,
        surgeryType: formData.surgeryType,
        interventionReason: formData.interventionReason,
        doctor: formData.doctor,
        interventionDate: formData.interventionDate,
        description: formData.description,
      }

      await restApi.createVascularAccessIntervention(patient.id, requestData)
      message.success('保存成功')
      setIsInterventionModalOpen(false)
      await loadInterventionRecords()
      await loadVascularAccesses() // 刷新以更新干预次数
    } catch (error) {
      console.error('保存干预记录失败:', error)
      message.error('保存失败')
      throw error
    } finally {
      setSaving(false)
    }
  }

  // 查看干预记录详情
  const handleViewIntervention = (record: VascularAccessInterventionApi) => {
    setSelectedIntervention(record)
    setIsInterventionDetailModalOpen(true)
  }

  // 删除干预记录
  const handleDeleteIntervention = (record: VascularAccessInterventionApi) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除 ${record.interventionDate} 的干预记录吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await restApi.deleteVascularAccessIntervention(patient.id, record.id)
          message.success('删除成功')
          await loadInterventionRecords()
          await loadVascularAccesses() // 刷新以更新干预次数
        } catch (error) {
          console.error('删除干预记录失败:', error)
          message.error('删除失败')
        }
      },
    })
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="animate-spin text-blue-500" size={32} />
        <span className="ml-3 text-slate-500">加载中...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in pb-10">
      {/* 上部区域：启用通路 + 干预记录 */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* 左上：当前启用血管通路 */}
        <div className="lg:col-span-8">
          <DetailCard className="h-full">
            <div className="flex justify-between items-center mb-8 px-1">
              <h3 className="text-sm font-black uppercase tracking-wider flex items-center text-slate-800">
                <Syringe size={18} className="mr-2 text-blue-600" /> {t('vascular.currentActiveAccess')}
              </h3>
              {/* 只有存在启用通路时才显示操作按钮 */}
              {activeAccess && (
                <div className="flex gap-2">
                  <button onClick={() => handleOpenEdit(activeAccess)} className="px-4 py-1.5 bg-blue-50 text-blue-600 rounded-lg text-xs font-black hover:bg-blue-100 transition-all flex items-center gap-1.5 border border-blue-100 shadow-sm"><Edit3 size={14}/> {t('action.edit')}</button>
                  <button onClick={() => setIsInterventionModalOpen(true)} className="px-4 py-1.5 bg-orange-50 text-orange-600 rounded-lg text-xs font-black hover:bg-orange-100 transition-all flex items-center gap-1.5 border border-orange-100 shadow-sm"><Wrench size={14}/> {t('vascular.intervention')}</button>
                  <button onClick={() => handleDelete(activeAccess)} className="px-4 py-1.5 bg-red-50 text-red-600 rounded-lg text-xs font-black hover:bg-red-100 transition-all flex items-center gap-1.5 border border-red-100 shadow-sm"><Trash2 size={14}/> {t('action.delete')}</button>
                </div>
              )}
            </div>
            {activeAccess ? (
              <div className="grid grid-cols-4 gap-y-10 mt-6 px-4">
                <LabelValue label={t('label.accessType')} value={<span className="text-2xl font-black text-blue-600">{dictName(DICT_TYPES.VASCULAR_ACCESS, activeAccess.accessType)}</span>} />
                <LabelValue label={t('vascular.accessSite')} value={dictName(DICT_TYPES.VASCULAR_SITE, activeAccess.site)} />
                <LabelValue label={t('label.leftRight')} value={formatSide(activeAccess.side)} />
                <LabelValue label={t('label.artery')} value={dictArrayNames(DICT_TYPES.ARTERY_TYPE, activeAccess.artery)} />
                <LabelValue label={t('label.vein')} value={dictArrayNames(DICT_TYPES.VEIN_TYPE, activeAccess.vein)} />
                <LabelValue label={t('vascular.surgeryDoctor')} value={dictName(DICT_TYPES.DOCTOR, activeAccess.surgeon)} />
                <LabelValue label={t('vascular.surgeryTime')} value={activeAccess.surgeryDate || '-'} />
                <LabelValue label="手术医院" value={dictName(DICT_TYPES.HOSPITAL, activeAccess.hospital)} />
                <LabelValue label={t('vascular.isDefault')} value={<span className="text-green-600 font-black">{t('label.yes')}</span>} />
                <LabelValue label={t('label.currentStatus')} value={<span className="text-green-600 font-black">{t('vascular.statusEnabled')}</span>} />
                <div className="col-span-3"></div>
                <div className="col-span-full mt-2"><LabelValue label={t('label.notes')} value={activeAccess.notes || '-'} /></div>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-20 text-slate-300">
                <ShieldX size={48} className="mb-4 opacity-20"/>
                <p className="font-bold">{t('vascular.noActiveAccess')}</p>
                <button onClick={handleOpenCreate} className="mt-4 px-6 py-2 bg-blue-600 text-white rounded-xl text-xs font-black shadow-lg">{t('vascular.addNow')}</button>
              </div>
            )}
          </DetailCard>
        </div>

        {/* 右上：当前通路干预记录 */}
        <div className="lg:col-span-4">
          <DetailCard className="h-full bg-slate-50/50 border-slate-100 flex flex-col relative">
            <SectionHeader
              icon={Activity}
              title={t('vascular.interventionRecords')}
            />
            <div className="mt-4 flex-1 overflow-y-auto max-h-[280px]">
              {hasInterventionRecords ? (
                <div className="space-y-2">
                  {interventionRecords.map((record) => (
                    <div
                      key={record.id}
                      className="bg-white rounded-lg p-3 border border-slate-100 hover:border-orange-200 transition-all group relative cursor-pointer"
                      onClick={() => handleViewIntervention(record)}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] font-black text-orange-600 bg-orange-50 px-2 py-0.5 rounded">
                          {dictName(DICT_TYPES.SURGERY_TYPE, record.surgeryType)}
                        </span>
                        <span className="text-[10px] text-slate-400">{record.interventionDate}</span>
                      </div>
                      <div className="text-xs text-slate-600 mb-1">{record.interventionReason}</div>
                      {record.doctor && (
                        <div className="text-[10px] text-slate-400">医生: {dictName(DICT_TYPES.DOCTOR, record.doctor)}</div>
                      )}
                      {/* 删除按钮 - hover时显示 */}
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDeleteIntervention(record)
                        }}
                        className="absolute top-2 right-2 w-6 h-6 bg-red-100 text-red-500 rounded-full opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center hover:bg-red-200"
                        title="删除"
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-10 text-slate-300">
                  <Activity size={32} className="mb-3 opacity-30"/>
                  <p className="text-xs font-bold">{t('vascular.noInterventionRecords')}</p>
                </div>
              )}
            </div>
          </DetailCard>
        </div>
      </div>

      {/* 下方：其他血管通路 */}
      <div className="w-full">
        <DetailCard>
          <SectionHeader
            icon={ClipboardList}
            title="血管通路列表"
            action={
              <button
                onClick={handleOpenCreate}
                className="px-4 py-1.5 bg-blue-600 text-white rounded-lg text-xs font-black hover:bg-blue-700 transition-all flex items-center gap-1.5 shadow-sm"
              >
                <Plus size={14}/> 新增
              </button>
            }
          />
          <div className="mt-4 border border-slate-100 rounded-2xl overflow-hidden shadow-sm">
            <table className="w-full text-left text-xs border-collapse">
              <thead className="bg-[#f8faff] text-slate-500 font-black uppercase tracking-widest text-[9px] border-b border-slate-100">
                <tr>
                  <th className="py-4 px-5">{t('label.accessType')}</th>
                  <th className="py-4 px-5">{t('vascular.accessSite')}</th>
                  <th className="py-4 px-5">{t('label.artery')}</th>
                  <th className="py-4 px-5">{t('label.vein')}</th>
                  <th className="py-4 px-5 text-center">{t('label.leftRight')}</th>
                  <th className="py-4 px-5 text-center">{t('vascular.isDefault')}</th>
                  <th className="py-4 px-5 text-center">{t('label.currentStatus')}</th>
                  <th className="py-4 px-5 text-right">{t('vascular.operation')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50">
                {vascularHistory.length > 0 ? vascularHistory.map((item) => (
                  <tr key={item.id} className={`hover:bg-slate-50 transition-colors group ${item.isDisabled ? 'opacity-60' : ''}`}>
                    <td className="py-5 px-5"><span className={`px-2 py-1 rounded-lg font-black text-[11px] border ${item.isDisabled ? 'bg-slate-100 text-slate-500 border-slate-200' : 'bg-blue-50 text-blue-600 border-blue-100'}`}>{dictName(DICT_TYPES.VASCULAR_ACCESS, item.accessType)}</span></td>
                    <td className="py-5 px-5 font-bold text-slate-700">{dictName(DICT_TYPES.VASCULAR_SITE, item.site)}</td>
                    <td className="py-5 px-5 text-slate-500">{dictArrayNames(DICT_TYPES.ARTERY_TYPE, item.artery)}</td>
                    <td className="py-5 px-5 text-slate-500">{dictArrayNames(DICT_TYPES.VEIN_TYPE, item.vein)}</td>
                    <td className="py-5 px-5 text-center font-black text-slate-400">{formatSide(item.side)}</td>
                    <td className="py-5 px-5 text-center"><span className="text-slate-300">{item.isDefault ? t('label.yes') : '-'}</span></td>
                    <td className="py-5 px-5 text-center">
                      {item.isDisabled
                        ? <span className="px-2 py-0.5 rounded-lg text-[10px] font-black uppercase bg-slate-100 text-slate-400 border border-slate-200">{t('vascular.statusDisabled')}</span>
                        : <span className="px-2 py-0.5 rounded-lg text-[10px] font-black uppercase bg-green-50 text-green-600 border border-green-100">{t('vascular.statusEnabled')}</span>
                      }
                    </td>
                    <td className="py-5 px-5 text-right whitespace-nowrap">
                      <div className="flex items-center justify-end gap-1.5 opacity-0 group-hover:opacity-100 transition-all transform translate-x-2 group-hover:translate-x-0">
                        <button onClick={() => handleOpenEdit(item)} className="flex items-center gap-1 px-3 py-1.5 bg-white border border-slate-100 text-blue-600 rounded-xl text-[10px] font-black hover:bg-blue-600 hover:text-white transition-all shadow-sm"><Edit3 size={12}/> {t('action.edit')}</button>
                        <button onClick={() => setIsInterventionModalOpen(true)} className="flex items-center gap-1 px-3 py-1.5 bg-white border border-orange-100 text-orange-600 rounded-xl text-[10px] font-black hover:bg-orange-100 transition-all shadow-sm"><Wrench size={12}/> {t('vascular.intervention')}</button>
                        <button onClick={() => handleDelete(item)} className="flex items-center gap-1 px-3 py-1.5 bg-white border border-red-100 text-red-400 rounded-xl text-[10px] font-black hover:bg-red-600 hover:text-white transition-all shadow-sm"><Trash2 size={12}/> {t('action.delete')}</button>
                      </div>
                    </td>
                  </tr>
                )) : (
                  <tr><td colSpan={8} className="py-10 text-center text-slate-300 italic font-bold">{t('vascular.noDisabledAccess')}</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </DetailCard>
      </div>

      {/* 血管通路编辑弹窗 */}
      <VascularAccessModal
        isOpen={isVascularModalOpen}
        onClose={() => setIsVascularModalOpen(false)}
        onSave={handleSave}
        initialData={editingRecord ? convertRecordToForm(editingRecord) : undefined}
        mode={modalMode}
        saving={saving}
      />

      {/* 血管通路干预弹窗 */}
      <VascularInterventionModal
        isOpen={isInterventionModalOpen}
        onClose={() => setIsInterventionModalOpen(false)}
        vascularAccessType={activeAccess?.accessType}
        onSave={handleSaveIntervention}
      />

      {/* 干预记录详情弹窗 */}
      {isInterventionDetailModalOpen && selectedIntervention && (
        <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in" onClick={() => setIsInterventionDetailModalOpen(false)}>
          <div className="bg-white rounded-[12px] shadow-2xl w-full max-w-md overflow-hidden animate-scale-in" onClick={(e) => e.stopPropagation()}>
            {/* Header */}
            <div className="bg-[#fff7ed] px-6 py-4 flex items-center justify-between border-b border-orange-100">
              <h3 className="text-base font-bold text-slate-800 flex items-center">
                <Activity size={18} className="mr-2 text-orange-600" /> 干预记录详情
              </h3>
              <button
                onClick={() => setIsInterventionDetailModalOpen(false)}
                className="p-1.5 hover:bg-white/50 rounded-full transition-all text-slate-400 hover:text-slate-600"
              >
                <X size={20} />
              </button>
            </div>

            {/* Content */}
            <div className="p-6 space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-slate-500">手术类型</span>
                <span className="text-sm font-bold text-orange-600 bg-orange-50 px-3 py-1 rounded-lg">
                  {dictName(DICT_TYPES.SURGERY_TYPE, selectedIntervention.surgeryType)}
                </span>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-slate-500">干预日期</span>
                <span className="text-sm font-bold text-slate-800">{selectedIntervention.interventionDate}</span>
              </div>

              <div className="border-t border-slate-100 pt-4">
                <span className="text-sm font-medium text-slate-500 block mb-2">干预原因</span>
                <p className="text-sm text-slate-700 bg-slate-50 rounded-lg p-3">
                  {selectedIntervention.interventionReason}
                </p>
              </div>

              {selectedIntervention.doctor && (
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-slate-500">干预医生</span>
                  <span className="text-sm font-bold text-slate-800">{dictName(DICT_TYPES.DOCTOR, selectedIntervention.doctor)}</span>
                </div>
              )}

              {selectedIntervention.avgBloodFlow > 0 && (
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-slate-500">平均血流量</span>
                  <span className="text-sm font-bold text-slate-800">{selectedIntervention.avgBloodFlow} ml/min</span>
                </div>
              )}

              {selectedIntervention.usageDays > 0 && (
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-slate-500">使用天数</span>
                  <span className="text-sm font-bold text-slate-800">{selectedIntervention.usageDays} 天</span>
                </div>
              )}

              {selectedIntervention.description && (
                <div className="border-t border-slate-100 pt-4">
                  <span className="text-sm font-medium text-slate-500 block mb-2">干预描述</span>
                  <p className="text-sm text-slate-700 bg-slate-50 rounded-lg p-3">
                    {selectedIntervention.description}
                  </p>
                </div>
              )}
            </div>

            {/* Footer */}
            <div className="bg-slate-50 px-6 py-4 flex justify-end">
              <button
                onClick={() => setIsInterventionDetailModalOpen(false)}
                className="px-6 py-2 rounded-lg bg-slate-200 text-slate-700 text-sm font-bold hover:bg-slate-300 transition-all"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
