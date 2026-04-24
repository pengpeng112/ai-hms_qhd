
import React from 'react';
import { Printer, Share2, Info } from 'lucide-react';
import { Patient } from '../types';

interface Props {
  patient: Patient;
}

const PatientSummaryHeader: React.FC<Props> = ({ patient }) => {
  return (
    <div className="flex items-start justify-between">
      <div className="flex items-center space-x-8">
        <div className="flex items-center space-x-3">
          <h1 className="text-2xl font-black text-slate-800 tracking-tight">{patient.name}</h1>
          <span className="bg-blue-600 text-white px-3 py-0.5 rounded-sm text-sm font-bold">{patient.bedId}</span>
        </div>

        <div className="flex items-center space-x-8 text-xs text-gray-500">
          <div>
            <p className="text-[10px] font-medium text-gray-400 uppercase mb-0.5 tracking-wider">患者ID</p>
            <p className="font-bold text-slate-700">{patient.patientId}</p>
          </div>
          <div>
            <p className="text-[10px] font-medium text-gray-400 uppercase mb-0.5 tracking-wider">性别/年龄</p>
            <p className="font-bold text-slate-700">{patient.gender} / {patient.age}岁</p>
          </div>
          <div>
            <p className="text-[10px] font-medium text-gray-400 uppercase mb-0.5 tracking-wider">费用</p>
            <p className="font-bold text-blue-600 underline cursor-pointer">{patient.costType}</p>
          </div>
          <div>
            <p className="text-[10px] font-medium text-gray-400 uppercase mb-0.5 tracking-wider">透析龄</p>
            <p className="font-bold text-slate-700">{patient.dialysisAge}</p>
          </div>
        </div>
      </div>

      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-12 px-6 py-2 border rounded-xl bg-white shadow-sm">
          <div className="text-center">
             <p className="text-[10px] text-gray-400 font-bold mb-1">干体重</p>
             <div className="flex items-baseline space-x-1 justify-center">
                <span className="text-xl font-black text-blue-600">{patient.dryWeight}</span>
                <span className="text-[10px] text-gray-500 font-bold uppercase">kg</span>
             </div>
          </div>
          <div className="h-10 w-px bg-gray-100"></div>
          <div className="text-center">
             <p className="text-[10px] text-gray-400 font-bold mb-1">治疗方案</p>
             <p className="text-xl font-black text-slate-800">{patient.treatmentPlan}</p>
          </div>
        </div>
        
        <button className="flex flex-col items-center justify-center p-3 border rounded-xl bg-white hover:bg-gray-50 transition-all shadow-sm group">
          <Printer size={18} className="text-gray-400 group-hover:text-blue-600 mb-1" />
          <span className="text-[10px] font-bold text-gray-500">打印/导出</span>
        </button>
      </div>
    </div>
  );
};

export default PatientSummaryHeader;
