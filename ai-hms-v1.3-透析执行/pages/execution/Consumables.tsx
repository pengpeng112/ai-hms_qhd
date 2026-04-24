
import React from 'react';
import { Patient } from '../../types';
import { PackageSearch } from 'lucide-react';

const Consumables: React.FC<{ patient: Patient }> = ({ patient }) => (
  <div className="p-8 border-2 border-dashed border-gray-100 rounded-3xl flex flex-col items-center justify-center text-gray-400 bg-gray-50/30">
    <PackageSearch size={64} strokeWidth={1} className="mb-4 text-blue-100" />
    <h3 className="text-lg font-bold text-gray-600 mb-2">耗材批号核对</h3>
    <p className="text-xs max-w-xs text-center leading-relaxed">通过扫码或手动选择，记录透析器、管路、补液及针头的具体规格与有效期。</p>
  </div>
);

export default Consumables;
