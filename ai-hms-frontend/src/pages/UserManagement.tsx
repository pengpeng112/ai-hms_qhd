import { useCallback, useEffect, useMemo, useState } from 'react'
import { message, Table, Button, Space, Input, Modal, Form, Select, Switch, Popconfirm } from 'antd'
import { Plus, RefreshCw, Search } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { userApi, type RestUser, type CreateUserRequest } from '@/services/userApi'
import { roleManagementApi, type AppRoleApi } from '@/services/roleManagementApi'

function ageFromBirthdate(birthdate?: string): number | null {
  if (!birthdate) return null
  const d = new Date(birthdate)
  if (isNaN(d.getTime())) return null
  const diff = Date.now() - d.getTime()
  return Math.floor(diff / 31557600000)
}

export default function UserManagement() {
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<RestUser[]>([])
  const [keyword, setKeyword] = useState('')
  const [roles, setRoles] = useState<AppRoleApi[]>([])
  const [editVisible, setEditVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<RestUser | null>(null)
  const [form] = Form.useForm()
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10, total: 0 })

  const loadUsers = useCallback(async () => {
    setLoading(true)
    try {
      const res = await userApi.getList({
        keyword: keyword || undefined,
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
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [keyword, pagination.current, pagination.pageSize])

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
      content: (
        <Input.Password id="reset-pwd-input" placeholder="请输入新密码" />
      ),
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
    setEditVisible(true)
  }

  const handleOpenEdit = (record: RestUser) => {
    setEditingUser(record)
    form.setFieldsValue({
      username: record.username,
      realName: record.realName,
      gender: record.gender || '',
      phone: record.phone || '',
      email: record.email || '',
      type: record.type || record.accountType || '',
      roles: record.roles || (record.role ? [record.role] : []),
    })
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
        roles: values.roles || undefined,
      }
      if (editingUser) {
        // When editing, password is optional
        if (values.password) {
          payload.password = values.password
        }
        await userApi.update(editingUser.id, payload)
        message.success('修改成功')
      } else {
        if (!values.password) {
          message.warning('请输入密码')
          return
        }
        payload.password = values.password
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

  const roleOptions = useMemo(() => roles.map(r => ({ label: r.name, value: r.code })), [roles])

  const columns = useMemo(() => [
    { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
    {
      title: '真实姓名',
      dataIndex: 'realName',
      key: 'realName',
      width: 100,
      render: (v: string) => v || '-',
    },
    {
      title: '性别',
      key: 'gender',
      width: 60,
      render: (_: unknown, r: RestUser) => {
        if (r.gender === 'M') return '男'
        if (r.gender === 'F') return '女'
        return r.gender || '-'
      },
    },
    {
      title: '年龄',
      key: 'age',
      width: 60,
      render: (_: unknown, r: RestUser) => {
        const a = ageFromBirthdate(r.birthdate)
        return a !== null ? String(a) : '-'
      },
    },
    {
      title: '人员类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (v: string, r: RestUser) => v || r.accountType || '-',
    },
    {
      title: '角色',
      dataIndex: 'roles',
      key: 'roles',
      width: 180,
      render: (vals: string[], r: RestUser) => {
        const codes = vals || (r.role ? [r.role] : [])
        const names = codes
          .map((code: string) => roles.find(item => item.code === code)?.name || code)
          .filter(Boolean)
        return names.join('、') || '-'
      },
    },
    { title: '手机号', dataIndex: 'phone', key: 'phone', width: 130, render: (v: string) => v || '-' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: string, record: RestUser) => (
        <Switch
          size="small"
          checked={v === 'active'}
          onChange={() => handleToggleStatus(record.id, v)}
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: RestUser) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => handleOpenEdit(record)}>编辑</Button>
          <Button type="link" size="small" onClick={() => handleResetPassword(record.id)}>重置密码</Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  // eslint-disable-next-line react-hooks/exhaustive-deps
  ], [roles])

  return (
    <div className="max-w-[1400px] mx-auto">
      <div className="bg-white rounded-lg p-6 shadow-sm">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h2 className="text-xl font-bold text-slate-800">人员管理</h2>
            <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
              <span>系统设置</span>
              <span>/</span>
              <span className="text-slate-600">用户管理</span>
            </div>
          </div>
          <Space>
            <Input
              placeholder="搜索姓名/用户名/拼音"
              value={keyword}
              onChange={e => setKeyword(e.target.value)}
              onPressEnter={() => { setPagination(prev => ({ ...prev, current: 1 })); void loadUsers() }}
              prefix={<Search size={16} className="text-slate-400" />}
              allowClear
              className="w-56"
            />
            <Button onClick={() => { setPagination(prev => ({ ...prev, current: 1 })); void loadUsers() }} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={handleOpenCreate} icon={<Plus size={16} />}>新增人员</Button>
          </Space>
        </div>
        <Table
          dataSource={users}
          columns={columns}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1100 }}
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
        width={560}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input placeholder="请输入用户名" disabled={!!editingUser} />
          </Form.Item>
          <Form.Item name="realName" label="真实姓名" rules={[{ required: true, message: '请输入真实姓名' }]}>
            <Input placeholder="请输入真实姓名" />
          </Form.Item>
          <Form.Item name="password" label={editingUser ? '密码（留空不修改）' : '密码'} rules={editingUser ? [] : [{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder={editingUser ? '留空则保持原密码' : '请输入密码'} />
          </Form.Item>
          <Space className="w-full" size="middle">
            <Form.Item name="gender" label="性别" className="flex-1">
              <Select placeholder="请选择" allowClear options={[{ label: '男', value: 'M' }, { label: '女', value: 'F' }]} />
            </Form.Item>
            <Form.Item name="phone" label="手机号" className="flex-1">
              <Input placeholder="请输入" />
            </Form.Item>
            <Form.Item name="email" label="邮箱" className="flex-1">
              <Input placeholder="请输入" />
            </Form.Item>
          </Space>
          <Space className="w-full" size="middle">
            <Form.Item name="type" label="人员类型" className="flex-1">
              <Select
                placeholder="请选择"
                allowClear
                options={[
                  { label: '医生', value: 'DOCTOR' },
                  { label: '护士', value: 'NURSE' },
                  { label: '工程师', value: 'ENGINEER' },
                  { label: '管理员', value: 'ADMIN' },
                ]}
              />
            </Form.Item>
            <Form.Item name="roles" label="角色" className="flex-1">
              <Select mode="multiple" placeholder="请选择角色" allowClear options={roleOptions} />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}
