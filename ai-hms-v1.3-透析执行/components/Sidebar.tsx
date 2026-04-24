
import React from 'react';
import { 
  LayoutDashboard, 
  Users, 
  Activity, 
  Calendar, 
  Package, 
  BarChart3, 
  Database, 
  Settings, 
  Monitor 
} from 'lucide-react';

const menuItems = [
  { icon: LayoutDashboard, label: '工作台', path: '/dashboard' },
  { icon: Monitor, label: '病区概览', path: '/ward' },
  { icon: Users, label: '患者管理', path: '/patients' },
  { icon: Activity, label: '透析执行', path: '/execution', active: true },
  { icon: Calendar, label: '排班管理', path: '/schedule' },
  { icon: Package, label: '耗材管理', path: '/consumables' },
  { icon: BarChart3, label: '统计报表', path: '/reports' },
  { icon: Database, label: '主数据管理', path: '/data' },
  { icon: Settings, label: '系统设置', path: '/settings' },
];

interface Props {
  isVisible: boolean;
}

const Sidebar: React.FC<Props> = ({ isVisible }) => {
  return (
    <div className={`h-full bg-[#001529] text-gray-300 flex flex-col overflow-hidden whitespace-nowrap transition-all duration-300 ${isVisible ? 'w-52' : 'w-0'}`}>
      <div className="p-4 flex items-center space-x-2 border-b border-white/10 shrink-0">
        <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center font-bold text-white shrink-0">AI</div>
        <span className="text-lg font-bold text-white tracking-tight">智能透析</span>
      </div>
      <nav className="flex-1 py-4 overflow-y-auto no-scrollbar overflow-x-hidden">
        {menuItems.map((item) => (
          <div
            key={item.label}
            className={`flex items-center space-x-3 px-6 py-3 cursor-pointer transition-colors group ${
              item.active ? 'bg-blue-600 text-white' : 'hover:bg-white/5 hover:text-white'
            }`}
          >
            <item.icon size={18} className="shrink-0" />
            <span className="text-sm font-medium">{item.label}</span>
          </div>
        ))}
      </nav>
      <div className="p-4 border-t border-white/10 shrink-0">
        <div className="text-[10px] uppercase text-gray-500 font-semibold mb-2">当前登录身份</div>
        <div className="text-xs font-medium text-blue-400">NURSE_HEAD</div>
      </div>
    </div>
  );
};

export default Sidebar;
