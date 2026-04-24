
import React, { useState } from 'react';
import PatientListSidebar from '../components/PatientListSidebar';
import PatientSummaryHeader from '../components/PatientSummaryHeader';
import { ExecutionTab, Patient } from '../types';
import { ChevronLeft, ChevronRight } from 'lucide-react';

// Import sub-pages
import PreAssessment from './execution/PreAssessment';
import TodayPrescription from './execution/TodayPrescription';
import Verification from './execution/Verification';
import MedicalOrders from './execution/MedicalOrders';
import MidMonitoring from './execution/MidMonitoring';
import PostAssessment from './execution/PostAssessment';
import HealthEducation from './execution/HealthEducation';
import DialysisSummary from './execution/DialysisSummary';

const patients: Patient[] = [
  { id: '1', name: '张梦琪', bedId: 'A01', gender: '男', age: 35, status: '透析中', patientId: 'P001', costType: '职工医保', dialysisAge: '3年2个月', dryWeight: 55, treatmentPlan: 'HDF' },
  { id: '2', name: '李芳', bedId: 'A02', gender: '女', age: 36, status: '透析中', patientId: 'P002', costType: '居民医保', dialysisAge: '1年5个月', dryWeight: 52, treatmentPlan: 'HD' },
  { id: '3', name: '王忆柳', bedId: 'A03', gender: '男', age: 37, status: '透析中', patientId: 'P003', costType: '职工医保', dialysisAge: '5年', dryWeight: 60, treatmentPlan: 'HDF' },
  { id: '4', name: '刘敏', bedId: 'A04', gender: '女', age: 38, status: '透析中', patientId: 'P004', costType: '公费', dialysisAge: '2年', dryWeight: 48, treatmentPlan: 'HD' },
  { id: '5', name: '陈之桃', bedId: 'A05', gender: '男', age: 39, status: '透析中', patientId: 'P005', costType: '职工医保', dialysisAge: '4年', dryWeight: 65, treatmentPlan: 'HDF' },
  { id: '6', name: '杨强', bedId: 'A06', gender: '女', age: 40, status: '候诊', patientId: 'P006', costType: '职工医保', dialysisAge: '6个月', dryWeight: 70, treatmentPlan: 'HD' },
];

const DialysisExecution: React.FC = () => {
  const [selectedPatient, setSelectedPatient] = useState<Patient>(patients[0]);
  const [activeTab, setActiveTab] = useState<ExecutionTab>(ExecutionTab.PRE_ASSESSMENT);
  const [isPatientListVisible, setIsPatientListVisible] = useState(true);

  const renderContent = () => {
    switch (activeTab) {
      case ExecutionTab.PRE_ASSESSMENT: return <PreAssessment patient={selectedPatient} />;
      case ExecutionTab.TODAY_PRESCRIPTION: return <TodayPrescription patient={selectedPatient} />;
      case ExecutionTab.DUAL_CHECK: return <Verification patient={selectedPatient} />;
      case ExecutionTab.MEDICAL_ORDERS: return <MedicalOrders patient={selectedPatient} />;
      case ExecutionTab.MID_MONITORING: return <MidMonitoring patient={selectedPatient} />;
      case ExecutionTab.POST_ASSESSMENT: return <PostAssessment patient={selectedPatient} />;
      case ExecutionTab.EDUCATION: return <HealthEducation patient={selectedPatient} />;
      case ExecutionTab.SUMMARY: return <DialysisSummary patient={selectedPatient} />;
      default: return <PreAssessment patient={selectedPatient} />;
    }
  };

  return (
    <div className="flex h-full relative overflow-hidden">
      <div className={`transition-all duration-300 ease-in-out border-r overflow-hidden ${isPatientListVisible ? 'w-64' : 'w-0'}`}>
        <PatientListSidebar 
          patients={patients} 
          selectedId={selectedPatient.id} 
          onSelect={setSelectedPatient} 
          isVisible={isPatientListVisible}
        />
      </div>

      <div 
        onClick={() => setIsPatientListVisible(!isPatientListVisible)}
        className="absolute top-1/2 -translate-y-1/2 z-20 cursor-pointer group transition-all duration-300"
        style={{ left: isPatientListVisible ? '244px' : '-4px' }}
      >
        <div className="w-5 h-20 bg-white border border-gray-200 rounded-full flex items-center justify-center shadow-sm hover:bg-blue-50 hover:border-blue-200 transition-colors">
          {isPatientListVisible ? (
            <ChevronLeft size={14} className="text-gray-400 group-hover:text-blue-500" />
          ) : (
            <ChevronRight size={14} className="text-gray-400 group-hover:text-blue-500 ml-1" />
          )}
        </div>
      </div>

      <div className="flex-1 flex flex-col bg-white overflow-hidden shadow-inner min-w-0">
        <div className="border-b bg-gray-50/50 p-4 shrink-0">
          <div className="flex items-center space-x-1 mb-4 overflow-x-auto no-scrollbar">
            {Object.values(ExecutionTab).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-4 py-2 text-xs font-medium whitespace-nowrap rounded transition-all duration-200 ${
                  activeTab === tab
                    ? 'bg-blue-600 text-white shadow-md'
                    : 'text-gray-500 hover:bg-white hover:text-blue-600'
                }`}
              >
                {tab}
              </button>
            ))}
          </div>
          <PatientSummaryHeader patient={selectedPatient} />
        </div>
        <div className="flex-1 overflow-y-auto p-6 bg-white no-scrollbar">
           {renderContent()}
        </div>
      </div>
    </div>
  );
};

export default DialysisExecution;
