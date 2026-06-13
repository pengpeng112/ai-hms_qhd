import { useCallback, useEffect, useMemo, useState } from 'react'
import { message, Table, Button, Space, Modal, Form, Input, Select, Switch, Popconfirm, InputNumber, Tag } from 'antd'
import { Plus, Edit3, Trash2, RefreshCw, Building2, Search, Bed, AlertTriangle } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import {
  wardManagementApi, type WardItem, type WardPayload,
  bedManagementApi, type BedItem, type BedPayload,
} from '@/services/managementApi'
import { userApi } from '@/services/userApi'
import { dictCache, DICT_TYPES, type DictItem } from '@/services/dictApi'

function isGarbledText(value?: string | null): boolean {
  const text = String(value || '').trim()
  if (!text) return false
  return /\?{2,}|\uFFFD|锟/.test(text)
}

type BedStatusFilter = 'ALL' | 'enabled' | 'disabled' | 'garbled'

export default function WardBedManagement() {
  const [loading, setLoading] = useState(false)
  const [wards, setWards] = useState<WardItem[]>([])
  const [beds, setBeds] = useState<BedItem[]>([])
  const [selectedWardId, setSelectedWardId] = useState<string>('ALL')
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<BedStatusFilter>('ALL')

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [wardItems, bedItems] = await Promise.all([
        wardManagementApi.list(),
        bedManagementApi.list(),
      ])
      setWards(wardItems)
      setBeds(bedItems)
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void loadData() }, [loadData])

  const summary = useMemo(() => {
    const wardTotal = wards.length
    const wardEnabled = wards.filter(w => !w.isDisabled).length
    const bedTotal = beds.length
    const bedEnabled = beds.filter(b => !b.isDisabled).length
    const bedNoWard = beds.filter(b => !b.wardId || Number(b.wardId) === 0 || !wards.some(w => String(w.id) === String(b.wardId))).length
    const garbledWards = wards.filter(w => isGarbledText(w.name)).length
    const garbledBeds = beds.filter(b => isGarbledText(b.name) || isGarbledText(b.wardName)).length
    return { wardTotal, wardEnabled, bedTotal, bedEnabled, bedNoWard, garbled: garbledWards + garbledBeds }
  }, [wards, beds])

  const selectedWard = useMemo(() => wards.find(w => String(w.id) === selectedWardId) ?? null, [wards, selectedWardId])

  const filteredBeds = useMemo(() => beds.filter(bed => {
    const wardOk = selectedWardId === 'ALL' || String(bed.wardId) === selectedWardId
    const kw = keyword.trim()
    const keywordOk = !kw || bed.name.includes(kw) || bed.wardName.includes(kw) || String(bed.id).includes(kw)
    const garbled = isGarbledText(bed.name) || isGarbledText(bed.wardName)
    const statusOk =
      statusFilter === 'ALL' ||
      (statusFilter === 'enabled' && !bed.isDisabled) ||
      (statusFilter === 'disabled' && bed.isDisabled) ||
      (statusFilter === 'garbled' && garbled)
    return wardOk && keywordOk && statusOk
  }), [beds, selectedWardId, keyword, statusFilter])

  // ---- Ward Modal ----
  const [wardEditVisible, setWardEditVisible] = useState(false)
  const [editingWard, setEditingWard] = useState<WardItem | null>(null)
  const [wardForm] = Form.useForm()
  const [userOptions, setUserOptions] = useState<Array<{ label: string; value: string }>>([])
  const [infectionTypeDict, setInfectionTypeDict] = useState<DictItem[]>([])

  useEffect(() => {
    userApi.getList({ page: 1, pageSize: 200 }).then(res => {
      setUserOptions((res.items || []).map(u => ({ label: u.realName || u.username, value: String(u.id) })))
    }).catch(() => {})
  }, [])

  useEffect(() => {
    dictCache.getItems(DICT_TYPES.INFECTION_TYPE).then(setInfectionTypeDict).catch(() => {})
  }, [])

  const openWardCreate = () => { setEditingWard(null); wardForm.resetFields(); setWardEditVisible(true) }
  const openWardEdit = (record: WardItem) => {
    setEditingWard(record)
    wardForm.setFieldsValue({
      name: record.name, sort: record.sort,
      patientType: record.patientType || '',
      infectionType: record.infectionType || '',
      responsibleUsers: record.responsibleUsers ? record.responsibleUsers.split(',').map(s => s.trim()).filter(Boolean) : [],
      note: record.note || '', isDisabled: record.isDisabled,
    })
    setWardEditVisible(true)
  }

  const handleWardSubmit = async () => {
    try {
      const values = await wardForm.validateFields()
      const payload: WardPayload = {
        name: values.name, sort: values.sort ?? 0,
        patientType: values.patientType || undefined,
        infectionType: values.infectionType || undefined,
        responsibleUsers: Array.isArray(values.responsibleUsers) ? values.responsibleUsers.join(',') : (values.responsibleUsers || undefined),
        note: values.note || undefined, isDisabled: values.isDisabled ?? false,
      }
      if (editingWard) {
        await wardManagementApi.update(editingWard.id, payload)
        message.success('病区修改成功')
      } else {
        await wardManagementApi.create(payload)
        message.success('病区新建成功')
      }
      setWardEditVisible(false)
      void loadData()
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleWardDelete = async (id: string) => {
    try { await wardManagementApi.remove(id); message.success('已删除'); void loadData() } catch (e) { message.error(getErrorMessage(e)) }
  }

  // ---- Bed Modal ----
  const [bedEditVisible, setBedEditVisible] = useState(false)
  const [editingBed, setEditingBed] = useState<BedItem | null>(null)
  const [bedForm] = Form.useForm()

  const wardOptions = useMemo(() => wards.map(w => ({ label: w.name, value: String(w.id) })), [wards])

  const openBedCreate = () => {
    setEditingBed(null)
    bedForm.resetFields()
    if (selectedWardId !== 'ALL') { bedForm.setFieldsValue({ wardId: selectedWardId }) }
    setBedEditVisible(true)
  }
  const openBedEdit = (record: BedItem) => {
    setEditingBed(record)
    bedForm.setFieldsValue({
      name: record.name, wardId: record.wardId ? String(record.wardId) : undefined,
      sort: record.sort, fepId: record.fepId ?? undefined,
      acquisiteConnectId: record.acquisiteConnectId ?? undefined,
      note: record.note || '', isDisabled: record.isDisabled,
    })
    setBedEditVisible(true)
  }

  const handleBedSubmit = async () => {
    try {
      const values = await bedForm.validateFields()
      const payload: BedPayload = {
        name: values.name, wardId: Number(values.wardId), sort: values.sort ?? 0,
        fepId: values.fepId ? Number(values.fepId) : undefined,
        acquisiteConnectId: values.acquisiteConnectId ? Number(values.acquisiteConnectId) : undefined,
        note: values.note || undefined, isDisabled: values.isDisabled ?? false,
      }
      if (editingBed) {
        await bedManagementApi.update(editingBed.id, payload)
        message.success('床位修改成功')
      } else {
        await bedManagementApi.create(payload)
        message.success('床位新建成功')
      }
      setBedEditVisible(false)
      void loadData()
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleBedDelete = async (id: string) => {
    try { await bedManagementApi.remove(id); message.success('已删除'); void loadData() } catch (e) { message.error(getErrorMessage(e)) }
  }

  const bedColumns = [
    { title: '床位名称', dataIndex: 'name', key: 'name', width: 160,
      render: (v: string) => isGarbledText(v) ? <span className="text-red-500">{v} <Tag color="red" className="ml-1">疑似乱码</Tag></span> : v },
    { title: '所属病区', dataIndex: 'wardName', key: 'wardName', width: 140,
      render: (v: string) => isGarbledText(v) ? <span className="text-red-500">{v}</span> : (v || '-') },
    { title: '默认设备', dataIndex: 'defaultEquipmentName', key: 'defaultEquipmentName', width: 120, render: (v: string) => v || '-' },
    { title: '设备数', dataIndex: 'equipmentCount', key: 'equipmentCount', width: 70 },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 70 },
    { title: '状态', dataIndex: 'isDisabled', key: 'isDisabled', width: 70, render: (v: boolean) => <Switch size="small" checked={!v} disabled /> },
    {
      title: '操作', key: 'action', width: 160,
      render: (_: unknown, record: BedItem) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => openBedEdit(record)}><Edit3 size={14} /> 编辑</Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleBedDelete(record.id)}>
            <Button type="link" size="small" danger><Trash2 size={14} /> 删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const summaryCards = [
    { label: '病区总数', value: summary.wardTotal, icon: Building2, color: 'bg-blue-50 text-blue-600' },
    { label: '启用病区', value: summary.wardEnabled, icon: Building2, color: 'bg-emerald-50 text-emerald-600' },
    { label: '床位总数', value: summary.bedTotal, icon: Bed, color: 'bg-indigo-50 text-indigo-600' },
    { label: '启用床位', value: summary.bedEnabled, icon: Bed, color: 'bg-green-50 text-green-600' },
    { label: '未分配病区', value: summary.bedNoWard, icon: AlertTriangle, color: 'bg-amber-50 text-amber-500' },
    { label: '疑似乱码', value: summary.garbled, icon: AlertTriangle, color: summary.garbled > 0 ? 'bg-red-50 text-red-500' : 'bg-slate-50 text-slate-400' },
  ]

  return (
    <div className="h-full flex flex-col max-w-[1440px] mx-auto" style={{ background: '#f5f7fb' }}>
      <div className="px-6 pt-6 pb-4 shrink-0">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 mb-5">
          <div>
            <h2 className="text-2xl font-extrabold text-slate-800 flex items-center">
              <Building2 className="mr-3 text-blue-600" size={28} /> 病区床位管理
            </h2>
            <p className="text-sm text-slate-500 mt-1.5">统一维护病区、床位与床位所属关系{summary.garbled > 0 ? ` · 检测到 ${summary.garbled} 条疑似乱码数据` : ''}</p>
          </div>
          <Space>
            <Button onClick={loadData} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={openWardCreate} icon={<Plus size={16} />}>新建病区</Button>
            <Button onClick={openBedCreate} icon={<Plus size={16} />}>新建床位</Button>
          </Space>
        </div>

        <div className="grid grid-cols-3 lg:grid-cols-6 gap-3 mb-5">
          {summaryCards.map(s => (
            <div key={s.label} className="bg-white rounded-[14px] border border-slate-100 px-4 py-3 flex items-center gap-3 shadow-sm">
              <div className={`w-10 h-10 rounded-[12px] ${s.color} flex items-center justify-center shrink-0`}><s.icon size={20} /></div>
              <div>
                <p className="text-xs text-slate-400 font-medium">{s.label}</p>
                <p className={`text-xl font-extrabold ${s.label === '疑似乱码' && s.value > 0 ? 'text-red-600' : 'text-slate-800'}`}>{s.value}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex-1 flex gap-4 px-6 pb-6 min-h-0">
        {/* 左侧病区列表 */}
        <div className="w-[280px] shrink-0 flex flex-col gap-2 overflow-y-auto" style={{ scrollbarWidth: 'thin' }}>
          <button
            onClick={() => setSelectedWardId('ALL')}
            className={`text-left px-4 py-3 rounded-[12px] font-bold text-sm transition-all ${selectedWardId === 'ALL' ? 'bg-blue-600 text-white shadow-md' : 'bg-white text-slate-600 hover:bg-slate-50 border border-slate-100'}`}
          >
            <span className="flex items-center gap-2"><Building2 size={16} /> 全部病区</span>
            <span className="block text-xs mt-0.5 opacity-70">{wards.length} 个病区 · {beds.length} 个床位</span>
          </button>
          {wards.map(ward => {
            const wardBedCount = beds.filter(b => String(b.wardId) === String(ward.id)).length
            const garbled = isGarbledText(ward.name)
            const active = selectedWardId === String(ward.id)
            return (
              <div key={ward.id} className="group relative">
                <button
                  onClick={() => setSelectedWardId(String(ward.id))}
                  className={`w-full text-left px-4 py-3 rounded-[12px] transition-all ${active ? 'bg-blue-600 text-white shadow-md' : ward.isDisabled ? 'bg-white text-slate-400 border border-slate-100 opacity-70' : 'bg-white text-slate-700 border border-slate-100 hover:bg-slate-50'}`}
                >
                  <div className="flex items-center gap-2 font-bold text-sm">
                    {garbled ? <span className={active ? 'text-red-200' : 'text-red-500'}>{ward.name}</span> : <span className="truncate">{ward.name}</span>}
                    {garbled && <Tag color="red" className="text-xs">乱码</Tag>}
                  </div>
                  <div className={`flex items-center gap-3 text-xs mt-1 ${active ? 'opacity-80' : 'opacity-50'}`}>
                    <span>{wardBedCount} 床位</span>
                    <span>{ward.patientTypeLabel || ward.patientType || '-'}</span>
                    <span>{ward.isDisabled ? '停用' : '启用'}</span>
                  </div>
                </button>
                <div className="absolute right-2 top-1/2 -translate-y-1/2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button type="text" size="small" className={active ? 'text-white hover:text-blue-200' : 'text-slate-400 hover:text-blue-600'} onClick={(e) => { e.stopPropagation(); openWardEdit(ward) }}><Edit3 size={13} /></Button>
                  <Popconfirm title="确定删除？" onConfirm={() => handleWardDelete(ward.id)}>
                    <Button type="text" size="small" danger className={active ? 'text-white hover:text-red-200' : ''} onClick={e => e.stopPropagation()}><Trash2 size={13} /></Button>
                  </Popconfirm>
                </div>
              </div>
            )
          })}
        </div>

        {/* 右侧床位区 */}
        <div className="flex-1 min-w-0 flex flex-col min-h-0">
          <div className="bg-white rounded-[14px] border border-slate-100 shadow-sm p-4 mb-4 shrink-0">
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-[12px] bg-blue-50 text-blue-600 flex items-center justify-center shrink-0"><Building2 size={20} /></div>
                <div>
                  <p className="font-bold text-sm text-slate-800">{selectedWard ? selectedWard.name : '全部病区'}</p>
                  <p className="text-xs text-slate-400">{selectedWard ? `患者类型 ${selectedWard.patientTypeLabel || '-'} · ${selectedWard.infectionType || '无感染标记'}` : `共 ${filteredBeds.length} 个床位`}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Input prefix={<Search size={14} className="text-slate-400" />} placeholder="搜索床位名或病区" value={keyword} onChange={e => setKeyword(e.target.value)} className="w-52" allowClear />
                <Select value={statusFilter} onChange={v => setStatusFilter(v)} className="w-28" options={[
                  { label: '全部', value: 'ALL' }, { label: '启用', value: 'enabled' }, { label: '停用', value: 'disabled' }, { label: '疑似乱码', value: 'garbled' },
                ]} />
              </div>
            </div>
          </div>

          <div className="flex-1 min-h-0 overflow-auto bg-white rounded-[14px] border border-slate-100 shadow-sm p-4">
            <Table dataSource={filteredBeds} columns={bedColumns} rowKey="id" loading={loading} pagination={false} size="middle" />
          </div>
        </div>
      </div>

      {/* ---- Ward Modal ---- */}
      <Modal title={editingWard ? '编辑病区' : '新建病区'} open={wardEditVisible} onCancel={() => setWardEditVisible(false)} onOk={handleWardSubmit} width={520} destroyOnClose>
        <Form form={wardForm} layout="vertical" className="mt-4">
          <Form.Item name="name" label="病区名称" rules={[{ required: true, message: '请输入病区名称' }]}>
            <Input placeholder="请输入病区名称" />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="sort" label="排序" className="flex-1"><InputNumber className="w-full" placeholder="序号" min={0} /></Form.Item>
            <Form.Item name="patientType" label="患者类型" className="flex-1">
              <Select placeholder="请选择" allowClear options={[{ label: '长期患者', value: '长期患者' }, { label: '临时患者', value: '临时患者' }]} />
            </Form.Item>
          </Space>
          <Space className="w-full" size="middle">
            <Form.Item name="infectionType" label="感染类型" className="flex-1">
              <Select placeholder="请选择" allowClear options={infectionTypeDict.map(d => ({ label: d.name, value: d.name }))} />
            </Form.Item>
            <Form.Item name="responsibleUsers" label="负责医护" className="flex-1">
              <Select mode="multiple" placeholder="请选择负责医护" options={userOptions} allowClear />
            </Form.Item>
          </Space>
          <Form.Item name="note" label="备注"><Input.TextArea rows={3} placeholder="请输入备注" /></Form.Item>
          <Form.Item name="isDisabled" label="是否禁用" valuePropName="checked"><Switch /></Form.Item>
        </Form>
      </Modal>

      {/* ---- Bed Modal ---- */}
      <Modal title={editingBed ? '编辑床位' : '新建床位'} open={bedEditVisible} onCancel={() => setBedEditVisible(false)} onOk={handleBedSubmit} width={560} destroyOnClose>
        <Form form={bedForm} layout="vertical" className="mt-4">
          <Form.Item name="name" label="床位名称" rules={[{ required: true, message: '请输入床位名称' }]}>
            <Input placeholder="请输入床位名称（如 床位01）" />
          </Form.Item>
          <Form.Item name="wardId" label="所属病区" rules={[{ required: true, message: '请选择所属病区' }]}>
            <Select placeholder="请选择病区" options={wardOptions} />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="sort" label="排序" className="flex-1"><InputNumber className="w-full" placeholder="序号" min={0} /></Form.Item>
          </Space>
          <Space className="w-full" size="middle">
            <Form.Item name="fepId" label="FEP 设备" className="flex-1"><InputNumber className="w-full" placeholder="FEP设备 ID" /></Form.Item>
            <Form.Item name="acquisiteConnectId" label="采集连接" className="flex-1"><InputNumber className="w-full" placeholder="采集连接 ID" /></Form.Item>
          </Space>
          <Form.Item name="note" label="备注"><Input.TextArea rows={3} placeholder="请输入备注" /></Form.Item>
          <Form.Item name="isDisabled" label="是否禁用" valuePropName="checked"><Switch /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
