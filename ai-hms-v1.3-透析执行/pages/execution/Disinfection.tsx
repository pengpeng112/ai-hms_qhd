
import React from 'react';
import { Patient } from '../../types';
import { Zap } from 'lucide-react';

const Disinfection: React.FC<{ patient: Patient }> = ({ patient }) => (
  <div className="p-8 border-2 border-dashed border-gray-100 rounded-3xl flex flex-col items-center justify-center text-gray-400 bg-gray-50/30">
    <Zap size={64} strokeWidth={1} className="mb-4 text-blue-100" />
    <h3 className="text-lg font-bold text-gray-600 mb-2">设备消毒记录</h3>
    <p className="text-xs max-w-xs text-center leading-relaxed">上机前后及换机位时的设备化学/热消毒流程核对与签字记录。</p>
  </div>
);

export default Disinfection;
