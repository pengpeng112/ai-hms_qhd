
import React from 'react';
import { Search } from 'lucide-react';
import { Patient } from '../types';

interface Props {
  patients: Patient[];
  selectedId: string;
  onSelect: (patient: Patient) => void;
  isVisible: boolean;
}

const PatientListSidebar: React.FC<Props> = ({ patients, selectedId, onSelect, isVisible }) => {
  return (
    <div className={`w-64 h-full bg-white flex flex-col transition-opacity duration-200 ${isVisible ? 'opacity-100' : 'opacity-0'}`}>
      <div className="p-4 border-b shrink-0">
        <div className="relative">
          <input
            type="text"
            placeholder="搜索姓名..."
            className="w-full pl-9 pr-4 py-2 text-xs border rounded bg-gray-50 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
          <Search className="absolute left-3 top-2.5 text-gray-400" size={14} />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto no-scrollbar whitespace-nowrap">
        {patients.map((p) => (
          <div
            key={p.id}
            onClick={() => onSelect(p)}
            className={`p-4 border-b cursor-pointer transition-all hover:bg-gray-50 ${
              selectedId === p.id ? 'bg-blue-50 border-l-4 border-l-blue-600' : ''
            }`}
          >
            <div className="flex justify-between items-start mb-1 overflow-hidden">
              <span className={`text-sm font-bold truncate mr-2 ${selectedId === p.id ? 'text-blue-700' : 'text-gray-800'}`}>
                {p.name}
              </span>
              <span className={`text-[10px] px-1.5 py-0.5 rounded font-bold shrink-0 ${
                p.status === '透析中' ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-500'
              }`}>
                {p.bedId}床
              </span>
            </div>
            <div className="flex justify-between items-end overflow-hidden">
              <span className="text-xs text-gray-500 truncate mr-2">{p.gender} {p.age}岁</span>
              <span className={`text-[10px] shrink-0 ${p.status === '透析中' ? 'text-blue-500' : 'text-gray-400'}`}>
                {p.status}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default PatientListSidebar;
