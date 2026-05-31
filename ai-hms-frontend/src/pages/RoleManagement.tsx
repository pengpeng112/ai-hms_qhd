import { useCallback, useEffect, useMemo, useState } from 'react'
import { message, Table, Button, Space, Modal, Form, Input, Tag, Tree, Input as AntInput } from 'antd'
import { Plus, Edit3, RefreshCw, Key } from 'lucide-react'
import { getErrorMessage } from '@/services/restClient'
import { roleManagementApi, type AppRoleApi } from '@/services/roleManagementApi'

interface PermissionNode {
  code: string
  name: string
  children?: PermissionNode[]
}

interface TreeNodeData {
  key: string
  title: React.ReactNode
  children?: TreeNodeData[]
}

function buildTreeData(nodes: PermissionNode[]): TreeNodeData[] {
  return nodes.map(node => ({
    key: node.code,
    title: <span><span className="font-medium">{node.name}</span><span className="text-xs text-slate-400 ml-2">{node.code}</span></span>,
    children: node.children?.length ? buildTreeData(node.children) : undefined,
  }))
}

export default function RoleManagement() {
  const [loading, setLoading] = useState(false)
  const [roles, setRoles] = useState<AppRoleApi[]>([])
  const [editVisible, setEditVisible] = useState(false)
  const [permissionVisible, setPermissionVisible] = useState(false)
  const [editingRole, setEditingRole] = useState<AppRoleApi | null>(null)
  const [form] = Form.useForm()
  const [permTree, setPermTree] = useState<TreeNodeData[]>([])
  const [checkedKeys, setCheckedKeys] = useState<string[]>([])
  const [permLoading, setPermLoading] = useState(false)

  const loadRoles = useCallback(async () => {
    setLoading(true)
    try {
      const res = await roleManagementApi.getRoleList()
      setRoles(res || [])
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void loadRoles() }, [loadRoles])

  const handleCreate = () => {
    setEditingRole(null)
    form.resetFields()
    setEditVisible(true)
  }

  const handleEdit = (record: AppRoleApi) => {
    setEditingRole(record)
    form.setFieldsValue({
      name: record.name,
      code: record.code,
      description: record.description || '',
    })
    setEditVisible(true)
  }

  const handleSubmitEdit = async () => {
    try {
      const values = await form.validateFields()
      if (editingRole) {
        await roleManagementApi.updateRole(editingRole.code, {
          name: values.name,
          description: values.description || undefined,
        })
        message.success('修改成功')
      } else {
        await roleManagementApi.createRole({
          code: values.code,
          name: values.name,
          description: values.description || undefined,
        })
        message.success('新增成功')
      }
      setEditVisible(false)
      void loadRoles()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(getErrorMessage(e))
    }
  }

  const handleDelete = (code: string) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除角色 ${code} 吗？`,
      onOk: async () => {
        try {
          await roleManagementApi.deleteRole(code)
          message.success('删除成功')
          void loadRoles()
        } catch (e) {
          message.error(getErrorMessage(e))
        }
      },
    })
  }

  const handleOpenPermission = async (record: AppRoleApi) => {
    setEditingRole(record)
    setPermLoading(true)
    setPermissionVisible(true)
    try {
      const [treeData, permData] = await Promise.all([
        roleManagementApi.getPermissionTree(),
        roleManagementApi.getRolePermissions(record.code),
      ])
      setPermTree(buildTreeData(treeData))
      setCheckedKeys(permData.permissionCodes || [])
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setPermLoading(false)
    }
  }

  const handleSubmitPermission = async () => {
    if (!editingRole) return
    try {
      await roleManagementApi.setRolePermissions(editingRole.code, checkedKeys)
      message.success('权限分配成功')
      setPermissionVisible(false)
    } catch (e) {
      message.error(getErrorMessage(e))
    }
  }

  const columns = useMemo(() => [
    { title: '角色名称', dataIndex: 'name', key: 'name', width: 160 },
    { title: '角色编码', dataIndex: 'code', key: 'code', width: 160 },
    { title: '描述', dataIndex: 'description', key: 'description', render: (v: string) => v || '-' },
    { title: '状态', dataIndex: 'isEnabled', key: 'isEnabled', width: 90, render: (v: boolean) => <Tag color={v ? 'blue' : 'default'}>{v ? '已启用' : '已禁用'}</Tag> },
    {
      title: '操作', key: 'action', width: 240,
      render: (_: unknown, record: AppRoleApi) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => handleEdit(record)}><Edit3 size={14} /> 编辑</Button>
          <Button type="link" size="small" onClick={() => handleOpenPermission(record)}><Key size={14} /> 权限</Button>
          <Button type="link" size="small" danger onClick={() => handleDelete(record.code)}>删除</Button>
        </Space>
      ),
    },
  // eslint-disable-next-line react-hooks/exhaustive-deps
  ], [])

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="bg-white rounded-lg p-6 shadow-sm">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h2 className="text-xl font-bold text-slate-800">角色管理</h2>
            <div className="flex items-center gap-1 mt-1 text-xs text-slate-400">
              <span>系统设置</span><span>/</span><span className="text-slate-600">角色管理</span>
            </div>
          </div>
          <Space>
            <Button onClick={loadRoles} icon={<RefreshCw size={16} />} loading={loading}>刷新</Button>
            <Button type="primary" onClick={handleCreate} icon={<Plus size={16} />}>新建角色</Button>
          </Space>
        </div>
        <Table dataSource={roles} columns={columns} rowKey="code" loading={loading} pagination={false} />
      </div>

      <Modal title={editingRole ? '编辑角色' : '新建角色'} open={editVisible} onCancel={() => setEditVisible(false)} onOk={handleSubmitEdit} width={480} destroyOnClose>
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="name" label="角色名称" rules={[{ required: true, message: '请输入角色名称' }]}>
            <Input placeholder="请输入角色名称" />
          </Form.Item>
          <Form.Item name="code" label="角色编码" rules={[{ required: true, message: '请输入角色编码' }]}>
            <Input placeholder="请输入角色编码" disabled={!!editingRole} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <AntInput.TextArea rows={3} placeholder="请输入描述" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={`分配权限 - ${editingRole?.name || ''}`} open={permissionVisible} onCancel={() => setPermissionVisible(false)} onOk={handleSubmitPermission} width={760} destroyOnClose>
        <div className="mb-3 p-3 bg-blue-50 rounded-md text-xs text-blue-700">
          权限按左侧菜单分组展示：勾选菜单控制入口可见，勾选下级操作用于后续按钮级控制。
        </div>
        {permLoading ? (
          <div className="text-center py-8 text-slate-400">加载权限数据...</div>
        ) : (
          <Tree checkable defaultExpandAll checkedKeys={checkedKeys} onCheck={(keys) => setCheckedKeys(keys as string[])} treeData={permTree} />
        )}
      </Modal>
    </div>
  )
}
