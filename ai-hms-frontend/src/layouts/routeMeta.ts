export interface RouteMeta {
  title: string
  breadcrumb: string[]
  parentPath?: string
}

const routeMetaMap: Record<string, RouteMeta> = {
  '/': { title: '工作台', breadcrumb: ['工作台'] },
  '/dashboard': { title: '工作台', breadcrumb: ['工作台'] },
  '/patients': { title: '患者管理', breadcrumb: ['患者管理'] },
  '/monitoring': { title: '实时监控', breadcrumb: ['实时监控'] },
  '/schedule': { title: '排班管理', breadcrumb: ['排班管理'] },
  '/statistics': { title: '统计报表', breadcrumb: ['统计报表'] },
  '/settings': { title: '系统设置', breadcrumb: ['系统设置'] },
  '/ward-overview': { title: '病区概览', breadcrumb: ['病区概览'] },
  '/dialysis-processing': { title: '透析执行', breadcrumb: ['透析执行'] },
  '/inventory': { title: '耗材管理', breadcrumb: ['基础数据', '耗材管理'] },
  '/device-binding': { title: '设备管理', breadcrumb: ['基础数据', '设备管理'] },
  '/ward-management': { title: '病区管理', breadcrumb: ['基础数据', '病区管理'] },
  '/bed-management': { title: '床位管理', breadcrumb: ['基础数据', '床位管理'] },
  '/education-management': { title: '宣教内容管理', breadcrumb: ['系统设置', '宣教内容管理'] },
  '/master-data': { title: '主数据管理', breadcrumb: ['系统设置', '主数据管理'] },
  '/treatment-config': { title: '诊疗配置', breadcrumb: ['系统设置', '诊疗配置'] },
  '/dict-config': { title: '字典配置', breadcrumb: ['系统设置', '字典配置'] },
  '/user-management': { title: '用户管理', breadcrumb: ['系统设置', '用户管理'] },
  '/role-management': { title: '角色管理', breadcrumb: ['系统设置', '角色管理'] },
  '/schedule-templates': { title: '排班模板', breadcrumb: ['排班管理', '排班模板'] },
  '/schedule-templates/edit': { title: '编辑模板', breadcrumb: ['排班管理', '排班模板', '编辑模板'], parentPath: '/schedule-templates' },
  '/shift-config': { title: '班次设置', breadcrumb: ['排班管理', '班次设置'] },
}

export function getRouteMeta(pathname: string): RouteMeta {
  if (routeMetaMap[pathname]) {
    return routeMetaMap[pathname]
  }

  if (pathname.startsWith('/patients/')) {
    return { title: '患者详情', breadcrumb: ['患者管理', '患者详情'], parentPath: '/patients' }
  }

  return { title: '页面', breadcrumb: ['页面'] }
}
