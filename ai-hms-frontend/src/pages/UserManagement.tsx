import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { message, Table, Button, Space, Input, Modal, Form, Select, Switch, Popconfirm, Tag } from 'antd'
import { Plus, RefreshCw, Search, Upload, X, Eye, Settings2 } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { userApi, type RestUser, type CreateUserRequest } from '@/services/userApi'
import { roleManagementApi, type AppRoleApi } from '@/services/roleManagementApi'

export default function UserManagement() {
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<RestUser[]>([])
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [typeFilter, setTypeFilter] = useState<string>('')
  const [roles, setRoles] = useState<AppRoleApi[]>([])
  const [editVisible, setEditVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<RestUser | null>(null)
  const [form] = Form.useForm()
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10, total: 0 })
  const [signaturePreview, setSignaturePreview] = useState<string>('')
  const [signatureChanged, setSignatureChanged] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const loadUsers = useCallback(async () => {
    setLoading(true)
    try {
      const res = await userApi.getList({
        keyword: keyword || undefined,
        status: statusFilter || undefined,
        type: typeFilter || undefined,
        page: pagination.current,
        pageSize: pagination.pageSize,
      })
      setUsers(res.items || [])
      setPagination(prev => ({ ...prev, total: res.total }))
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [keyword, statusFilter, typeFilter, pagination.current, pagination.pageSize])

  const loadRoles = useCallback(async () => {
    try {
      const items = await roleManagementApi.getRoleList()
      setRoles(items)
    } catch { /* ignore */ }
  }, [])

  useEffect(() => { void loadUsers() }, [loadUsers])
  useEffect(() => { void loadRoles() }, [loadRoles])

  const handleDelete = async (id: string) => {
    try {
      await userApi.remove(id)
      message.success('删除成功')
      void loadUsers()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const handleResetPassword = async (id: string) => {
    Modal.confirm({
      title: '重置密码',
      content: <Input.Password id="reset-pwd-input" placeholder="请输入新密码" />,
      onOk: async () => {
        const input = document.getElementById('reset-pwd-input') as HTMLInputElement | null
        const pwd = input?.value?.trim()
        if (!pwd) { message.warning('请输入新密码'); return Promise.reject() }
        try {
          await userApi.resetPassword(id, pwd)
          message.success('密码重置成功')
        } catch (e) {
          message.error(getErrorMessage(e))
          return Promise.reject()
        }
      },
    })
  }

  const handleToggleStatus = async (id: string, currentStatus: string) => {
    const newStatus = currentStatus === 'active' ? 'disabled' : 'active'
    try {
      await userApi.updateStatus(id, newStatus)
      message.success('状态已更新')
      void loadUsers()
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const handleOpenCreate = () => {
    setEditingUser(null)
    form.resetFields()
    setSignaturePreview('')
    setSignatureChanged(false)
    setEditVisible(true)
  }

  const handleOpenEdit = async (record: RestUser) => {
    setEditingUser(record)
    form.setFieldsValue({
      username: record.username,
      realName: record.realName,
      gender: record.gender || '',
      phone: record.phone || '',
      email: record.email || '',
      type: record.type || '',
      sort: record.sort ?? 0,
      idNumber: record.idNumber || '',
      icNumber: record.icNumber || '',
      avatar: record.avatar || '',
      birthdate: record.birthdate ? record.birthdate.slice(0, 10) : '',
      roles: record.roleNames || [],
    })

    setSignatureChanged(false)
    if (record.hasSignature) {
      try {
        const sig = await userApi.getSignature(record.id)
        setSignaturePreview(sig.signatureImage || '')
      } catch {
        setSignaturePreview('')
      }
    } else {
      setSignaturePreview('')
    }
    setEditVisible(true)
  }

  const handleSubmitEdit = async () => {
    try {
      const values = await form.validateFields()
      const payload: CreateUserRequest = {
        username: values.username,
        realName: values.realName,
        gender: values.gender || undefined,
        phone: values.phone || undefined,
        email: values.email || undefined,
        type: values.type || undefined,
        sort: values.sort ?? undefined,
        idNumber: values.idNumber || undefined,
        icNumber: values.icNumber || undefined,
        avatar: values.avatar || undefined,
        birthdate: values.birthdate || undefined,
        roles: values.roles || undefined,
      }
      if (editingUser) {
        if (values.password) {
          payload.password = values.password
        }
        if (signatureChanged) {
          payload.signatureImage = signaturePreview || ''
        }
        await userApi.update(editingUser.id, payload)
        message.success('修改成功')
      } else {
        if (!values.password) {
          message.warning('请输入密码')
          return
        }
        payload.password = values.password
        if (signaturePreview) {
          payload.signatureImage = signaturePreview
        }
        await userApi.create(payload)
        message.success('新增成功')
      }
      setEditVisible(false)
      void loadUsers()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleSignatureUpload = () => {
    fileInputRef.current?.click()
  }

  const handleSignatureFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.type.match(/^image\/(png|jpeg|bmp|gif)$/)) {
      message.error('仅支持 PNG/JPEG/BMP/GIF 格式')
      return
    }
    if (file.size > 2 * 1024 * 1024) {
      message.error('签名图片不能超过 2MB')
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      setSignaturePreview(reader.result as string)
      setSignatureChanged(true)
    }
    reader.readAsDataURL(file)
  }

  const handleRemoveSignature = async () => {
    if (editingUser) {
      try {
        await userApi.deleteSignature(editingUser.id)
        message.success('签名已移除')
      } catch (e) {
        message.error(getErrorMessage(e))
      }
    }
    setSignaturePreview('')
    setSignatureChanged(true)
  }

  const roleOptions = useMemo(() => roles.map(r => ({ label: r.name, value: r.name })), [roles])

  const columns = useMemo(() => [
    {
      title: '序号', key: 'index', width: 60,
      render: (_: unknown, __: unknown, idx: number) => (pagination.current - 1) * pagination.pageSize + idx + 1,
    },
    {
      title: '姓名', dataIndex: 'realName', key: 'realName', width: 100,
      render: (v: string) => v || '-',
    },
    {
      title: '性别', key: 'gender', width: 60,
      render: (_: unknown, r: RestUser) => r.gender || '-',
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (v: string, record: RestUser) => (
        <Switch size="small" checked={v === 'active'} onChange={() => handleToggleStatus(record.id, v)} />
      ),
    },
    {
      title: '人员类型', dataIndex: 'type', key: 'type', width: 90,
      render: (v: string) => v || '-',
    },
    {
      title: '账户类型', key: 'accountType', width: 90,
      render: (_: unknown, r: RestUser) => {
        const names = r.roleNames || (r.role ? [r.role] : [])
        if (names.length === 0) return '-'
        const labels = names
          .map((n: string) => roles.find(item => item.name === n))
          .filter(Boolean)
        return labels.length > 0 ? labels[0]!.name || '普通用户' : '普通用户'
      },
    },
    {
      title: '用户名', dataIndex: 'username', key: 'username', width: 100,
      render: (v: string) => v,
    },
    {
      title: '绑定', dataIndex: 'bindStatus', key: 'bindStatus', width: 70,
      render: (v: string) => (
        <Tag color={v === 'bound' ? 'blue' : undefined}>{v === 'bound' ? '已绑定' : '未绑定'}</Tag>
      ),
    },
    {
      title: '同步', dataIndex: 'syncStatus', key: 'syncStatus', width: 70,
      render: (v: string) => (
        <Tag color={v === 'synced' ? 'green' : 'red'}>{v === 'synced' ? '已同步' : '未同步'}</Tag>
      ),
    },
    {
      title: '操作', key: 'action', width: 260,
      render: (_: unknown, record: RestUser) => (
        <Space size="small">
          <Button type="link" size="small" icon={<Eye size={13} />}>查看权限</Button>
          <Button type="link" size="small" icon={<Settings2 size={13} />}>管理账号</Button>
          <Button type="link" size="small" onClick={() => { void handleOpenEdit(record) }}>编辑</Button>
          <Button type="link" size="small" onClick={() => handleResetPassword(record.id)}>重置密码</Button>
          <Popconfirm title="确定删除该人员？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ], [roles, pagination])

  return (
    <div className="max-w-[1400px] mx-auto">
      <div className="bg-white rounded-lg p-6 shadow-sm">
        <div className="flex flex-col gap-3 mb-4">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-xl font-bold text-slate-800">人员管理</h2>
              <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
                <span>系统设置</span><span>/</span><span className="text-slate-600">用户管理</span>
              </div>
            </div>
            <Space>
              <Button onClick={() => { setPagination(prev => ({ ...prev, current: 1 })); void loadUsers() }} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
              <Button type="primary" onClick={handleOpenCreate} icon={<Plus size={16} />}>新增人员</Button>
            </Space>
          </div>

          <div className="flex items-center gap-3 flex-wrap">
            <Select
              placeholder="状态：全部"
              allowClear
              value={statusFilter || undefined}
              onChange={(v) => { setStatusFilter(v || ''); setPagination(prev => ({ ...prev, current: 1 })) }}
              className="w-28"
              options={[
                { label: '全部', value: '' },
                { label: '启用', value: 'active' },
                { label: '禁用', value: 'disabled' },
              ]}
            />
            <Select
              placeholder="人员类型：全部"
              allowClear
              value={typeFilter || undefined}
              onChange={(v) => { setTypeFilter(v || ''); setPagination(prev => ({ ...prev, current: 1 })) }}
              className="w-36"
              options={[
                { label: '全部', value: '' },
                { label: '医生', value: '医生' },
                { label: '护士', value: '护士' },
                { label: '患者', value: '患者' },
                { label: '运维', value: '运维' },
              ]}
            />
            <Input
              placeholder="搜索姓名/用户名/手机号/身份证号"
              value={keyword}
              onChange={e => setKeyword(e.target.value)}
              onPressEnter={() => { setPagination(prev => ({ ...prev, current: 1 })); void loadUsers() }}
              prefix={<Search size={16} className="text-slate-400" />}
              allowClear
              className="w-64"
            />
          </div>
        </div>

        <Table
          dataSource={users}
          columns={columns}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1200 }}
          size="small"
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 人`,
            onChange: (page, pageSize) => setPagination(prev => ({ ...prev, current: page, pageSize })),
          }}
        />
      </div>

      <Modal
        title={editingUser ? '编辑人员' : '新增人员'}
        open={editVisible}
        onCancel={() => setEditVisible(false)}
        onOk={handleSubmitEdit}
        width={720}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          <div className="grid grid-cols-2 gap-x-6 gap-y-2">
            <div>
              <Form.Item name="realName" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
                <Input placeholder="请输入姓名" />
              </Form.Item>
              <Form.Item name="gender" label="性别" rules={[{ required: true, message: '请选择性别' }]}>
                <Select placeholder="请选择" allowClear options={[{ label: '男', value: '男' }, { label: '女', value: '女' }]} />
              </Form.Item>
              <Form.Item name="type" label="人员类型">
                <Select placeholder="请选择" allowClear options={[
                  { label: '医生', value: '医生' },
                  { label: '护士', value: '护士' },
                  { label: '患者', value: '患者' },
                  { label: '运维', value: '运维' },
                ]} />
              </Form.Item>
              <Form.Item name="idNumber" label="身份证号">
                <Input placeholder="请输入身份证号" />
              </Form.Item>
              <Form.Item name="email" label="电子邮箱">
                <Input placeholder="请输入电子邮箱" />
              </Form.Item>

              <Form.Item label="手写签名">
                <div className="space-y-2">
                  {signaturePreview ? (
                    <div className="relative border rounded-lg p-2 bg-white inline-block">
                      <img src={signaturePreview} alt="签名预览" className="max-h-32 max-w-full object-contain" />
                      <button
                        type="button"
                        onClick={handleRemoveSignature}
                        className="absolute -top-2 -right-2 rounded-full bg-red-500 text-white p-0.5 hover:bg-red-600"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ) : (
                    <div className="border border-dashed border-slate-300 rounded-lg p-4 text-center text-xs text-slate-400">
                      暂无签名
                    </div>
                  )}
                  <div className="flex gap-2">
                    <Button size="small" icon={<Upload size={13} />} onClick={handleSignatureUpload}>上传</Button>
                    {signaturePreview && (
                      <Button size="small" danger onClick={handleRemoveSignature}>移除</Button>
                    )}
                  </div>
                  <input ref={fileInputRef} type="file" accept="image/png,image/jpeg,image/bmp,image/gif" className="hidden" onChange={handleSignatureFile} />
                </div>
              </Form.Item>
            </div>

            <div>
              <Form.Item label="头像">
                <div className="space-y-2">
                  <div className="w-20 h-20 rounded-full border-2 border-dashed border-slate-300 flex items-center justify-center bg-slate-50 text-slate-400">
                    <Upload size={20} />
                  </div>
                  <p className="text-[11px] text-slate-400">头像上传承接表未确认，暂不开放</p>
                </div>
              </Form.Item>

              <Form.Item name="sort" label="排序" initialValue={0}>
                <Input type="number" placeholder="排序号" />
              </Form.Item>
              <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }]}>
                <Input placeholder="请输入手机号" />
              </Form.Item>
              {editingUser && (
                <Form.Item label="创建时间">
                  <Input value={editingUser.createdAt || '-'} disabled />
                </Form.Item>
              )}
              <Form.Item label="同步云中心">
                <Button size="small" disabled className="text-[11px]">同步云中心（暂未开放）</Button>
              </Form.Item>

              <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
                <Input placeholder="请输入用户名" disabled={!!editingUser} />
              </Form.Item>
              <Form.Item name="password" label={editingUser ? '密码（留空不修改）' : '密码'} rules={editingUser ? [] : [{ required: true, message: '请输入密码' }]}>
                <Input.Password placeholder={editingUser ? '留空则保持原密码' : '请输入密码'} />
              </Form.Item>
              <Form.Item name="roles" label="角色">
                <Select mode="multiple" placeholder="请选择角色" allowClear options={roleOptions} />
              </Form.Item>
            </div>
          </div>
        </Form>
      </Modal>
    </div>
  )
}
