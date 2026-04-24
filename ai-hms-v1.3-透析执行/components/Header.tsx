
import React from 'react';
import { Bell, ChevronDown, User, Maximize2, RefreshCw, Menu, PanelLeftClose, PanelLeftOpen } from 'lucide-react';

interface Props {
  onToggleSidebar: () => void;
  isSidebarVisible: boolean;
}

const Header: React.FC<Props> = ({ onToggleSidebar, isSidebarVisible }) => {
  return (
    <header className="h-14 bg-white border-b flex items-center justify-between px-6 z-10 shrink-0">
      <div className="flex items-center space-x-4">
        <button 
          onClick={onToggleSidebar}
          className="p-1.5 hover:bg-gray-100 rounded-md text-gray-500 transition-colors mr-2"
          title={isSidebarVisible ? "隐藏导航栏" : "显示导航栏"}
        >
          {isSidebarVisible ? <PanelLeftClose size={20} /> : <PanelLeftOpen size={20} />}
        </button>
        <div className="flex items-center space-x-2 bg-blue-50 px-3 py-1.5 rounded-full text-blue-600 text-xs font-medium border border-blue-100">
          <span className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></span>
          <span>当前科室: 肾内透析中心 · 第一病区</span>
        </div>
      </div>

      <div className="flex items-center space-x-6">
        <div className="flex items-center space-x-4">
          <button className="text-gray-400 hover:text-blue-600 transition-colors">
            <RefreshCw size={18} />
          </button>
          <button className="text-gray-400 hover:text-blue-600 transition-colors">
            <Maximize2 size={18} />
          </button>
          <div className="relative">
            <Bell size={18} className="text-gray-400" />
            <span className="absolute -top-1.5 -right-1.5 bg-red-500 text-white text-[10px] font-bold w-4 h-4 flex items-center justify-center rounded-full">1</span>
          </div>
        </div>

        <div className="h-6 w-px bg-gray-200"></div>

        <div className="flex items-center space-x-3 group cursor-pointer">
          <div className="text-right">
            <p className="text-sm font-semibold text-gray-800 leading-tight">刘护士长</p>
            <p className="text-[10px] text-gray-400 font-medium">NURSE_HEAD</p>
          </div>
          <div className="w-9 h-9 bg-gray-100 rounded-full flex items-center justify-center overflow-hidden border border-gray-200 group-hover:border-blue-300 transition-all">
            <img src="https://picsum.photos/seed/nurse/100/100" alt="Avatar" className="w-full h-full object-cover" />
          </div>
          <ChevronDown size={14} className="text-gray-400" />
        </div>
      </div>
    </header>
  );
};

export default Header;
