import React, { useState, useMemo, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { restApi, convertRestPatientList } from '@/services';
import type { RestTreatment } from '@/services/restClient';
import type { Patient } from '@/types/original';
import { useDictNameMaps, getNameFromMap } from '@/hooks/useDictName';
import { DICT_TYPES } from '@/services/dictApi';
import {
  Search, Activity, FileText, CheckSquare, Stethoscope,
  Zap, Droplet, BookOpen, CheckCircle2, User,
  Printer, Clock, AlertCircle, Syringe,
  ArrowRight, PenTool, History,
  ChevronDown, Image as ImageIcon, Calendar, TrendingUp, HelpCircle, X,
  Info, Plus, Trash2, ChevronsDown, Settings
} from 'lucide-react';

interface DialysisProcessingProps {
  initialPatientId?: string;
}

interface ExtendedMonitorRecord {
  id: string;
  time: string;
  sbp: string;
  dbp: string;
  hr: string;
  temp: string;
  resp: string;
  spO2: string;
  ufVolume: string;
  bloodFlow: string;
  arterialPressure: string;
  venousPressure: string;
  transmembranePressure: string;
  machineTemp: string;
  nurse: string;
  conductivity?: string;
  heparinFlow?: string;
  symptoms?: string;
}

// Helper Components
const InfoField = ({ label, value, unit, className, highlight, icon }: { label: string, value: string | number, unit?: string, className?: string, highlight?: boolean, icon?: React.ReactNode }) => (
    <div className={`flex flex-col ${className}`}>
        <span className="text-xs text-gray-500 mb-1">{label}</span>
        <span className={`text-sm font-medium ${highlight ? 'text-blue-700 font-bold' : 'text-gray-900'} truncate flex items-center`}>
            {value} {unit && <span className="text-xs text-gray-500 font-normal ml-0.5">{unit}</span>}
            {icon}
        </span>
    </div>
);

interface PrescriptionInputProps {
  label?: string;
  suffix?: string;
  defaultValue?: string;
  width?: string;
  readOnly?: boolean;
}

const PrescriptionInput = ({ label, suffix, defaultValue, width = "w-24", readOnly }: PrescriptionInputProps) => (
    <div className="flex items-center space-x-2">
        {label && <label className="text-sm text-gray-600 whitespace-nowrap">{label}:</label>}
        <div className={`relative ${width}`}>
            <input
                type="text"
                defaultValue={defaultValue}
                readOnly={readOnly}
                className={`w-full h-8 px-2 border rounded text-sm outline-none focus:ring-2 focus:ring-blue-500 ${readOnly ? 'bg-gray-50 text-gray-500 border-gray-200' : 'bg-white border-gray-300'}`}
            />
            {suffix && <span className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-500">{suffix}</span>}
        </div>
    </div>
);

const CheckRow = ({
  label,
  defaultChecked = true,
  normalText,
  abnormalText,
  abnormalPlaceholder,
  nurseOptions,
}: {
  label: string
  defaultChecked?: boolean
  normalText?: string
  abnormalText?: string
  abnormalPlaceholder?: string
  nurseOptions: React.ReactNode
}) => (
  <div className="flex items-center border-b border-gray-100 last:border-0 bg-white hover:bg-gray-50 transition-colors">
      <div className="w-12 py-3 flex justify-center items-center border-r border-gray-100 h-12 shrink-0">
          <input type="checkbox" defaultChecked={defaultChecked} className="w-4 h-4 rounded text-blue-600 focus:ring-blue-500 border-gray-300"/>
      </div>
      <div className="w-64 py-3 px-4 border-r border-gray-100 text-gray-700 font-medium shrink-0 flex items-center">
          {label}
      </div>
      <div className="flex-1 py-3 px-4 border-r border-gray-100 flex items-center space-x-4 min-w-0">
           <div className="flex items-center space-x-4 shrink-0">
               <label className="flex items-center space-x-1.5 cursor-pointer">
                   <input type="radio" name={`status-${label}`} defaultChecked className="w-4 h-4 text-blue-600 focus:ring-blue-500 border-gray-300"/>
                   <span className="text-gray-700 text-sm">{normalText}</span>
               </label>
               <label className="flex items-center space-x-1.5 cursor-pointer">
                   <input type="radio" name={`status-${label}`} className="w-4 h-4 text-blue-600 focus:ring-blue-500 border-gray-300"/>
                   <span className="text-gray-700 text-sm">{abnormalText}</span>
               </label>
           </div>
           <input type="text" placeholder={abnormalPlaceholder} className="flex-1 px-3 py-1.5 bg-gray-50 border border-gray-200 rounded text-sm outline-none focus:border-blue-400 focus:bg-white transition-all w-full min-w-0"/>
      </div>
      <div className="w-32 py-3 px-4 border-r border-gray-100 shrink-0">
           <div className="relative">
                <select className="w-full appearance-none bg-transparent border border-gray-200 rounded px-2 py-1 pr-6 text-gray-600 text-sm outline-none focus:border-blue-500">
                    {nurseOptions}
                </select>
               <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
           </div>
      </div>
      <div className="w-48 py-3 px-4 shrink-0 flex items-center space-x-2">
           <div className="relative flex-1">
               <input type="text" defaultValue={getCurrentDateTimeText()} className="w-full bg-gray-50 border border-gray-200 rounded px-2 py-1 text-gray-600 text-sm outline-none"/>
               <Calendar size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400"/>
           </div>
      </div>
  </div>
);

const FormField = ({ label, required, children, className = "" }: { label: string, required?: boolean, children?: React.ReactNode, className?: string }) => (
    <div className={`flex items-start ${className}`}>
        <label className="w-32 min-w-[8rem] text-right text-sm text-gray-600 mr-3 shrink-0 pt-2 leading-tight">
            {required && <span className="text-red-500 mr-1">*</span>}{label}
            {label && label.endsWith(":") ? "" : ":"}
        </label>
        <div className="flex-1 min-w-0">{children}</div>
    </div>
);

interface PreAssessInputProps {
  label: string;
  suffix?: string | React.ReactNode;
  required?: boolean;
  defaultValue?: string;
  readOnly?: boolean;
  warning?: string;
  className?: string;
  placeholder?: string;
}

const PreAssessInput = ({ label, suffix, required, defaultValue, readOnly, warning, className, placeholder }: PreAssessInputProps) => (
    <div className={`flex items-start ${className}`}>
        <label className="w-32 min-w-[8rem] text-right text-sm text-gray-600 mr-3 shrink-0 flex justify-end items-start pt-2 leading-tight">
            {required && <span className="text-red-500 mr-1">*</span>}
            {label}
            {label && ":"}
        </label>
        <div className="flex-1 relative">
            <div className="flex items-center">
                <input
                    type="text"
                    defaultValue={defaultValue}
                    readOnly={readOnly}
                    placeholder={placeholder}
                    className={`w-full h-9 px-3 border rounded-md text-sm outline-none transition-colors
                        ${readOnly ? 'bg-gray-50 text-gray-500 border-gray-200' : 'bg-white border-gray-300 focus:border-blue-500 focus:ring-1 focus:ring-blue-200'}
                        ${warning ? 'border-red-300 bg-red-50 text-red-600' : ''}
                    `}
                />
                {suffix && <span className="ml-2 text-sm text-gray-500 shrink-0 w-8">{suffix}</span>}
            </div>
            {warning && <div className="absolute top-full left-0 text-[10px] text-red-500 mt-0.5 whitespace-nowrap">{warning}</div>}
        </div>
    </div>
);

const MonitorInput = ({ label, suffix, required, defaultValue, className = "" }: { label: string, suffix?: string, required?: boolean, defaultValue?: string, className?: string }) => (
    <div className={`flex items-center ${className}`}>
        <label className="w-28 min-w-[7rem] text-right text-sm text-gray-600 mr-2 shrink-0 leading-tight">
            {required && <span className="text-red-500 mr-1">*</span>}{label}
            {label && ":"}
        </label>
        <div className="flex-1 relative flex items-center">
            <input
                type="text"
                defaultValue={defaultValue}
                className="w-full h-8 px-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 outline-none"
            />
            {suffix && <span className="ml-2 text-xs text-gray-500 shrink-0 w-8">{suffix}</span>}
        </div>
    </div>
);

interface PrintPatient {
  name: string;
  gender: string;
  age: number;
  id: string;
  insuranceType: string;
  bedNumber: string;
}

interface PrintMonitorRecord {
  time: string
  sbp: string
  dbp: string
  hr: string
  vp: string
  tmp: string
  ufVolume: string
  symptoms: string
  nurse: string
}

const getCurrentDateText = () => new Date().toISOString().slice(0, 10);
const getCurrentDateTimeText = () => new Date().toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-').slice(0, 16);
const getCurrentDateTimeLocal = () => new Date().toISOString().slice(0, 16);

const PrintView = ({ patient, onClose, t, dictNameMaps, monitoringRecords }: { patient: PrintPatient, onClose: () => void, t: TFunction<'dialysisProcessing'>, dictNameMaps: Record<string, Map<string, string>>, monitoringRecords: PrintMonitorRecord[] }) => {
    const signedNurse = monitoringRecords[0]?.nurse || '--';
    return (
        <div className="fixed inset-0 z-[100] bg-gray-900/50 backdrop-blur-sm overflow-y-auto flex justify-center py-8 print:p-0 print:bg-white print:static print:h-auto print:overflow-visible">
             <div className="bg-white shadow-2xl w-[210mm] min-h-[297mm] p-[10mm] relative print:shadow-none print:w-full print:h-auto print:p-0 print:mx-0 print:my-0">
                  <div className="absolute top-4 right-[-60px] flex flex-col gap-3 print:hidden">
                    <button onClick={() => window.print()} className="p-3 bg-blue-600 text-white rounded-full shadow-lg hover:bg-blue-700 transition-transform hover:scale-110" title={t('header.printExport')}>
                        <Printer size={24}/>
                    </button>
                    <button onClick={onClose} className="p-3 bg-white text-gray-600 rounded-full shadow-lg hover:text-red-600 transition-transform hover:scale-110" title={t('common.cancel')}>
                        <X size={24}/>
                    </button>
                  </div>

                  <div className="text-center mb-6">
                      <h1 className="text-2xl font-bold tracking-widest text-gray-900 mb-2">{t('print.title')}</h1>
                      <div className="text-sm text-gray-500 pb-2 border-b-2 border-gray-800">{t('print.subtitle')}</div>
                  </div>

                  <div className="flex flex-wrap text-sm mb-4 font-medium text-gray-900 gap-y-2">
                      <div className="w-1/4"><span className="text-gray-500">{t('print.name')}:</span> {patient.name}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.gender')}:</span> {patient.gender}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.age')}:</span> {patient.age}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.dialysisNo')}:</span> {patient.id}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.date')}:</span> {getCurrentDateText()}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.insurance')}:</span> {getNameFromMap(dictNameMaps[DICT_TYPES.INSURANCE_TYPE] || new Map(), patient.insuranceType)}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.bed')}:</span> {patient.bedNumber}</div>
                      <div className="w-1/4"><span className="text-gray-500">{t('print.dialysisCount')}:</span> 152</div>
                  </div>

                  <div className="border border-gray-400 mb-4">
                      <div className="flex border-b border-gray-400">
                          <div className="w-24 bg-gray-100 p-2 text-xs font-bold border-r border-gray-400 flex items-center justify-center">{t('print.preAssess')}</div>
                          <div className="flex-1 p-2 text-xs">
                              <div className="grid grid-cols-4 gap-2 mb-2">
                                  <div>{t('print.weight')}: 78.5 kg</div>
                                  <div>{t('print.bp')}: 145/88 mmHg</div>
                                  <div>{t('print.hr')}: 78 bpm</div>
                                  <div>{t('print.temp')}: 36.5 鈩</div>
                              </div>
                              <div className="grid grid-cols-2 gap-2">
                                  <div>{t('print.symptoms')}: 鏃犵壒娈婁笉閫傦紝绮剧灏氬彲</div>
                                  <div>{t('print.fistula')}: 闇囬ⅳ鑹ソ锛岀毊鑲ゅ畬鏁</div>
                              </div>
                          </div>
                      </div>
                       <div className="flex">
                          <div className="w-24 bg-gray-100 p-2 text-xs font-bold border-r border-gray-400 flex items-center justify-center">{t('print.prescription')}</div>
                          <div className="flex-1 p-2 text-xs">
                              <div className="grid grid-cols-4 gap-2 mb-2">
                                  <div>{t('print.method')}: HDF</div>
                                  <div>{t('print.duration')}: 4.0 h</div>
                                  <div>{t('print.bloodFlow')}: 240 ml/min</div>
                                  <div>{t('print.dialysateFlow')}: 500 ml/min</div>
                              </div>
                              <div className="grid grid-cols-2 gap-2">
                                  <div>{t('print.anticoagulant')}: 浣庡垎瀛愯倽绱?2500iu</div>
                                  <div>{t('print.dialyzer')}: FX60</div>
                                  <div>{t('print.targetUF')}: 2.50 L</div>
                                  <div>{t('print.dryWeight')}: 77.0 kg</div>
                              </div>
                          </div>
                      </div>
                  </div>

                  <div className="mb-4">
                      <h3 className="text-sm font-bold mb-1 border-l-4 border-gray-800 pl-2">{t('print.monitorRecord')}</h3>
                      <table className="w-full text-xs text-center border-collapse border border-gray-400">
                          <thead className="bg-gray-100">
                              <tr>
                                  <th className="border border-gray-400 p-1">{t('print.time')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.bp')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.hr')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.vp')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.tmp')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.ufVolume')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.symptomHandle')}</th>
                                  <th className="border border-gray-400 p-1">{t('print.signature')}</th>
                              </tr>
                          </thead>
                          <tbody>
                              {monitoringRecords.map((rec, i: number) => (
                                  <tr key={i}>
                                      <td className="border border-gray-400 p-1">{rec.time}</td>
                                      <td className="border border-gray-400 p-1">{rec.sbp}/{rec.dbp}</td>
                                      <td className="border border-gray-400 p-1">{rec.hr}</td>
                                      <td className="border border-gray-400 p-1">{rec.vp}</td>
                                      <td className="border border-gray-400 p-1">{rec.tmp}</td>
                                      <td className="border border-gray-400 p-1">{rec.ufVolume}</td>
                                      <td className="border border-gray-400 p-1 text-left px-2">{rec.symptoms}</td>
                                      <td className="border border-gray-400 p-1">{rec.nurse}</td>
                                  </tr>
                              ))}
                              {[1,2,3,4].map(k => (
                                  <tr key={`e-${k}`}>
                                      <td className="border border-gray-400 p-1 h-6"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                      <td className="border border-gray-400 p-1"></td>
                                  </tr>
                              ))}
                          </tbody>
                      </table>
                  </div>

                  <div className="border border-gray-400 mb-6">
                      <div className="flex border-b border-gray-400">
                          <div className="w-24 bg-gray-100 p-2 text-xs font-bold border-r border-gray-400 flex items-center justify-center">{t('print.postAssess')}</div>
                          <div className="flex-1 p-2 text-xs">
                              <div className="grid grid-cols-4 gap-2 mb-2">
                                  <div>{t('print.weight')}: 77.2 kg</div>
                                  <div>{t('print.bp')}: 130/80 mmHg</div>
                                  <div>{t('print.actualUF')}: 2.50 L</div>
                                  <div>{t('print.temp')}: 36.6 鈩</div>
                              </div>
                              <div>{t('print.punctureSite')}: 姝㈣鑹ソ锛屾棤娓楄锛屾棤琛€鑲裤€</div>
                          </div>
                      </div>
                       <div className="flex">
                          <div className="w-24 bg-gray-100 p-2 text-xs font-bold border-r border-gray-400 flex items-center justify-center">{t('print.dialysisSummary')}</div>
                          <div className="flex-1 p-2 text-xs min-h-[60px]">
                              鎮ｈ€呬粖鏃ラ€忔瀽杩囩▼椤哄埄锛屾棤鐗规畩涓嶉€備富璇夈€傜敓鍛戒綋寰佸钩绋筹紝鍐呯槝绌垮埡鐐规棤娓楄锛岄€忔瀽缁撴潫鍚庡帇杩琛€椤哄埄銆傛不鐤楁晥鏋滆瘎浠凤細绋冲畾銆?                          </div>
                      </div>
                  </div>

                  <div className="flex justify-between text-sm mt-8 px-8">
                      <div>
                          <span className="font-bold mr-2">{t('print.doctorSign')}:</span>
                          <span className="border-b border-gray-800 w-32 inline-block text-center font-script text-lg">鐜嬪尰鐢</span>
                      </div>
                      <div>
                          <span className="font-bold mr-2">{t('print.nurseSign')}:</span>
                          <span className="border-b border-gray-800 w-32 inline-block text-center font-script text-lg">{signedNurse}</span>
                      </div>
                      <div>
                          <span className="font-bold mr-2">{t('print.dateSign')}:</span>
                          {getCurrentDateText()}
                      </div>
                  </div>

                  <div className="text-center text-[10px] text-gray-400 mt-12">
                      Generated by AI-HMS (System Ver 2.4.0)
                  </div>
             </div>
        </div>
    )
}

const DialysisProcessing: React.FC<DialysisProcessingProps> = ({ initialPatientId }) => {
  const { t } = useTranslation('dialysisProcessing');

  // 瀛楀吀鍚嶇О鏄犲皠
  const dictTypeCodes = useMemo(() => [
    DICT_TYPES.INSURANCE_TYPE,
    DICT_TYPES.DIALYSIS_MODE,
  ], []);
  const dictNameMaps = useDictNameMaps(dictTypeCodes);

  // REST API 患者数据
  const [patients, setPatients] = useState<Partial<Patient>[]>([]);

  // 加载患者列表
  useEffect(() => {
    restApi.getPatientList({ page: 1, pageSize: 200 })
      .then(res => {
        const list = convertRestPatientList(res.data.items);
        setPatients(list);
        // 鏁版嵁鍔犺浇鍚庤缃粯璁ら€変腑
        if (!initialPatientId && list.length > 0) {
          const active = list.find(p => p.status === '透析中');
          setSelectedPatientId(active?.id || list[0]?.id || null);
        }
      })
      .catch(err => console.error('鍔犺浇鎮ｈ€呮暟鎹け璐?', err));
  }, [initialPatientId]);

  const [selectedPatientId, setSelectedPatientId] = useState<string | null>(initialPatientId || null);
  const [activeProcessStep, setActiveProcessStep] = useState<string>('pre');
  const [searchTerm, setSearchTerm] = useState('');
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [isEditingPrescription, setIsEditingPrescription] = useState(false);
  const [showOrderModal, setShowOrderModal] = useState(false);
  const [showMonitorModal, setShowMonitorModal] = useState(false);
  const [showPrintPreview, setShowPrintPreview] = useState(false);

  const [monitorRecords, setMonitorRecords] = useState<ExtendedMonitorRecord[]>([]);
  const [nurses, setNurses] = useState<{ id: string; name: string }[]>([]);
  const [currentTreatment, setCurrentTreatment] = useState<RestTreatment | null>(null);

  const renderNurseOptions = () => {
    if (nurses.length === 0) {
      return <option value="">--璇烽€夋嫨--</option>;
    }
    return nurses.map((nurse) => (
      <option key={nurse.id} value={nurse.id}>
        {nurse.name}
      </option>
    ));
  };

  const currentDateTimeText = getCurrentDateTimeText();
  const currentDateTimeLocal = getCurrentDateTimeLocal();

  useEffect(() => {
    restApi
      .getUserList({ status: 'active' })
      .then((users) => {
        const nurseList = users
          .filter((user) => {
            const role = String(user.role ?? '');
            return role.includes('护') || role.toLowerCase().includes('nurse');
          })
          .map((user) => ({
            id: String(user.id),
            name: user.realName || user.username,
          }));
        setNurses(nurseList);
      })
      .catch(() => setNurses([]));
  }, []);

  useEffect(() => {
    if (!selectedPatientId) return;
    const today = new Date().toISOString().slice(0, 10);
    restApi.getPatientTreatmentByDate(selectedPatientId, today)
      .then(res => {
        setCurrentTreatment(res.data ?? null);
        const params = res.data?.duringParams ?? [];
        setMonitorRecords(params.map((p, idx) => ({
          id: String(p.id ?? idx),
          time: p.recordTime ? new Date(p.recordTime).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) : '',
          sbp: '', dbp: '', hr: '', temp: '', resp: '', spO2: '',
          ufVolume: String(p.ufVolume ?? ''),
          bloodFlow: String(p.bloodFlow ?? ''),
          arterialPressure: String(p.arterialPressure ?? ''),
          venousPressure: String(p.venousPressure ?? ''),
          transmembranePressure: String(p.tmp ?? ''),
          machineTemp: String(p.temperature ?? ''),
          conductivity: String(p.conductivity ?? ''),
          nurse: '',
        })));
      })
      .catch(() => {
        setCurrentTreatment(null);
      });
  }, [selectedPatientId]);

  const [materials] = useState([
      { id: 1, name: '锐针-16G', category: '穿刺针', count: 2, code: '', brand: 'NIPRO', spec: '', note: '' },
      { id: 2, name: 'SD-15HF', category: '透析器', count: 1, code: '', brand: 'SWS德莱福', spec: '', note: '' },
      { id: 3, name: '内瘘区', category: '护理区', count: 1, code: '1102011534', brand: '', spec: '', note: '' },
      { id: 4, name: '10ML娉ㄥ皠鍣?10ML', category: '鍏朵粬', count: 1, code: '', brand: '', spec: '10ML', note: '' },
      { id: 5, name: 'JRHLL-025', category: '琛€璺', count: 1, code: '', brand: '', spec: '', note: '' },
  ]);

  const selectedPatient = patients.find(p => p.id === selectedPatientId) as Patient | undefined;
  const filteredPatients = patients.filter(p =>
    p.status !== '灞呭' &&
    (p.name?.includes(searchTerm) || p.bedNumber?.includes(searchTerm) || p.id?.includes(searchTerm))
  );

  const parseSelectedPatientId = (): number | null => {
    if (!selectedPatientId) return null;
    const parsed = Number(selectedPatientId);
    if (!Number.isFinite(parsed)) return null;
    return parsed;
  };

  const ensureTodayTreatment = async (status: number) => {
    const numericPatientId = parseSelectedPatientId();
    if (!selectedPatient || numericPatientId === null) return null;
    if (currentTreatment) {
      const updated = await restApi.updateTreatment(currentTreatment.id, {
        status,
        notes: currentTreatment.notes ?? '',
      });
      setCurrentTreatment(updated.data);
      return updated.data;
    }
    const created = await restApi.createTreatment({
      patientId: numericPatientId,
      treatmentDate: new Date().toISOString(),
      type: 1,
      status,
      notes: '// TODO: 补充治疗子表 API',
    });
    setCurrentTreatment(created.data);
    return created.data;
  };

  const handleStartDialysis = async () => {
    try {
      await ensureTodayTreatment(1);
      setActiveProcessStep('monitor');
    } catch (error) {
      console.error('start dialysis failed', error);
    }
  };

  const handleFinishDialysis = async () => {
    if (!currentTreatment) return;
    try {
      await restApi.updateTreatmentStatus(currentTreatment.id, 2);
      setCurrentTreatment({ ...currentTreatment, status: 2 });
      setActiveProcessStep('disinfect');
    } catch (error) {
      console.error('finish dialysis failed', error);
    }
  };

  const steps = [
        { id: 'pre', label: t('step.pre'), icon: Activity, badge: null },
        { id: 'prescription', label: t('step.prescription'), icon: FileText, badge: null },
        { id: 'check1', label: t('step.check1'), icon: CheckSquare, badge: '6/6' },
        { id: 'check2', label: t('step.check2'), icon: CheckSquare, badge: '5/5' },
        { id: 'orders_process', label: t('step.orders'), icon: Syringe, badge: '1/1' },
        { id: 'monitor', label: t('step.monitor'), icon: Activity, badge: '10/8' },
        { id: 'post', label: t('step.post'), icon: Stethoscope, badge: null },
        { id: 'disinfect', label: t('step.disinfect'), icon: Zap, badge: null },
        { id: 'consumables', label: t('step.consumables'), icon: Droplet, badge: null },
        { id: 'education', label: t('step.education'), icon: BookOpen, badge: null },
        { id: 'summary', label: t('step.summary'), icon: FileText, badge: null },
  ];

  const renderPatientHeader = () => {
      if (!selectedPatient) return null;
      return (
        <div className="bg-white border-b border-gray-200 px-6 py-4 shadow-sm shrink-0">
            <div className="flex flex-wrap items-center justify-between gap-4 mb-4">
                <div className="flex items-center gap-4">
                    <div className="flex items-center">
                        <span className="text-2xl font-bold text-gray-900 mr-3">{selectedPatient.name}</span>
                        <span className="bg-blue-600 text-white px-2.5 py-1 rounded text-sm font-bold shadow-sm flex items-center">
                            {selectedPatient.bedNumber || t('header.waiting')}
                        </span>
                    </div>
                    <div className="h-8 w-px bg-gray-200"></div>
                    <div className="flex gap-4 text-sm text-gray-600">
                        <div className="flex flex-col">
                            <span className="text-xs text-gray-400">{t('header.patientId')}</span>
                            <span className="font-mono font-medium">{selectedPatient.id}</span>
                        </div>
                        <div className="flex flex-col">
                            <span className="text-xs text-gray-400">{t('header.genderAge')}</span>
                            <span className="font-medium">{selectedPatient.gender} / {selectedPatient.age}{t('header.age')}</span>
                        </div>
                        <div className="flex flex-col">
                            <span className="text-xs text-gray-400">{t('header.insurance')}</span>
                            <span className="font-medium text-blue-600">{getNameFromMap(dictNameMaps[DICT_TYPES.INSURANCE_TYPE] || new Map(), selectedPatient.insuranceType)}</span>
                        </div>
                         <div className="flex flex-col">
                            <span className="text-xs text-gray-400">{t('header.dialysisAge')}</span>
                            <span className="font-medium">3骞?涓湀</span>
                        </div>
                    </div>
                </div>

                <div className="flex gap-6 items-center">
                     <div className="bg-blue-50 px-4 py-2 rounded-lg border border-blue-100 min-w-[100px]">
                        <span className="text-xs text-blue-500 block">{t('header.dryWeight')}</span>
                        <span className="text-xl font-bold text-blue-700">{selectedPatient.dryWeight} <span className="text-xs font-normal">kg</span></span>
                     </div>
                     <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-100 min-w-[100px]">
                        <span className="text-xs text-gray-500 block">{t('header.treatmentPlan')}</span>
                        <span className="text-xl font-bold text-gray-800">{getNameFromMap(dictNameMaps[DICT_TYPES.DIALYSIS_MODE] || new Map(), selectedPatient.defaultMode)}</span>
                     </div>

                     <button
                        onClick={() => setShowPrintPreview(true)}
                        className="flex flex-col items-center justify-center p-2 text-gray-500 hover:text-blue-600 hover:bg-gray-100 rounded-lg transition-colors border border-transparent hover:border-gray-200 ml-2"
                        title={t('header.printExport')}
                     >
                        <Printer size={20} />
                        <span className="text-[10px] mt-1">{t('header.printExport')}</span>
                     </button>
                </div>
            </div>

            <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                <div className="flex items-center gap-6 text-xs">
                    <div className="flex items-center text-gray-500 bg-gray-50 px-2 py-1 rounded">
                        <History size={12} className="mr-1"/>
                        {t('header.lastDialysis')}: <span className="font-mono text-gray-700 font-medium ml-1">2023-10-24</span>
                    </div>
                    <span className="text-gray-400">|</span>
                    <span className="text-gray-500">{t('header.postWeight')}: <span className="font-mono text-gray-700 font-medium">69.8 kg</span></span>
                    <span className="text-gray-400">|</span>
                    <span className="text-gray-500">{t('header.lastUF')}: <span className="font-mono text-gray-700 font-medium">2.4 L</span></span>
                </div>
                <div className="text-xs flex items-center">
                     {selectedPatient.medicalHistory?.allergies && selectedPatient.medicalHistory.allergies.length > 0 ? (
                        <span className="text-red-600 flex items-center bg-red-50 px-3 py-1 rounded-full border border-red-100 font-medium animate-pulse-slow">
                            <AlertCircle size={12} className="mr-1.5"/>
                            {t('header.allergy')}: {selectedPatient.medicalHistory.allergies.join(', ')}
                        </span>
                     ) : (
                        <span className="text-green-600 flex items-center bg-green-50 px-3 py-1 rounded-full border border-green-100">
                             <CheckCircle2 size={12} className="mr-1.5"/> {t('header.noAllergy')}
                        </span>
                     )}
                </div>
            </div>
        </div>
      );
  };

  const renderProcessForm = () => {
        if (!selectedPatient) return (
            <div className="h-full flex flex-col items-center justify-center text-gray-400 bg-gray-50">
                <User size={64} className="mb-4 opacity-10"/>
                <p>{t('selectPatient')}</p>
            </div>
        );

        switch (activeProcessStep) {
            case 'pre':
                return (
                    <div className="p-6 max-w-[1600px] mx-auto animate-fade-in space-y-6">
                        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-8">
                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <div className="flex items-center">
                                    <PreAssessInput label={t('pre.weight')} suffix="kg" required defaultValue="90.5"/>
                                    <label className="ml-3 flex items-center text-sm text-gray-500 cursor-pointer whitespace-nowrap">
                                        <input type="checkbox" className="mr-1 rounded text-blue-600"/> {t('pre.refuseMeasure')}
                                    </label>
                                    <label className="ml-3 flex items-center text-sm text-gray-500 cursor-pointer whitespace-nowrap">
                                        <input type="checkbox" className="mr-1 rounded text-blue-600"/> {t('pre.bedridden')}
                                    </label>
                                </div>
                                <PreAssessInput label={t('pre.extraWeight')} suffix="kg" defaultValue="12"/>
                                <PreAssessInput label={t('pre.dryWeight')} suffix="kg" defaultValue="77" readOnly/>
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <PreAssessInput label={t('pre.weightGain')} suffix="kg" defaultValue="1.5" readOnly warning={t('pre.weightWarning')}/>
                                <PreAssessInput label={t('pre.ufVolume')} suffix="L" required defaultValue="1.7"/>
                                <PreAssessInput label={t('pre.temp')} suffix="℃" required defaultValue="36.5"/>
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <PreAssessInput label={t('pre.hr')} suffix="次/分" required defaultValue="64"/>
                                <PreAssessInput label={t('pre.resp')} suffix="次/分" required defaultValue="16"/>
                                <div className="flex items-center">
                                    <label className="w-28 text-right text-sm text-gray-600 mr-3 shrink-0 whitespace-nowrap flex justify-end items-center"><span className="text-red-500 mr-1">*</span>{t('pre.bp')}:</label>
                                    <div className="flex-1 flex space-x-2 items-center">
                                        <input type="text" defaultValue="110" className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 text-center"/>
                                        <input type="text" defaultValue="60" className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 text-center"/>
                                        <span className="self-center text-sm text-gray-500 whitespace-nowrap">mmHg</span>
                                    </div>
                                </div>
                            </div>

                            <div className="h-px bg-gray-100 my-6"></div>

                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <FormField label={t('pre.bpSite')}>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-white">
                                            <option>{t('bp.rightArm')}</option>
                                            <option>{t('bp.leftArm')}</option>
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <div className="lg:col-span-1">
                                    <FormField label={t('pre.note')}>
                                        <textarea rows={1} placeholder={t('pre.note')} className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 resize-none h-9"></textarea>
                                    </FormField>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <FormField label={t('pre.consciousness')}>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-white">
                                            <option>{t('consciousness.alert')}</option>
                                            <option>{t('consciousness.drowsy')}</option>
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('pre.symptoms')}>
                                    <div className="flex items-center h-9 px-2 border border-gray-300 rounded-md bg-white overflow-hidden">
                                        <span className="bg-gray-100 text-gray-700 px-2 py-0.5 rounded text-xs mr-2 flex items-center whitespace-nowrap">
                                            {t('pre.lowBP')} <button className="ml-1 hover:text-red-500"><X size={10}/></button>
                                        </span>
                                        <input type="text" className="flex-1 outline-none text-sm min-w-0" placeholder={t('pre.selectSymptom')}/>
                                    </div>
                                </FormField>
                                <FormField label={t('pre.fistulaStatus')}>
                                    <div className="flex flex-wrap gap-2 h-9 items-center">
                                        {[t('pre.murmurStrong'), t('pre.thrillStrong'), t('pre.pulseStrong')].map(tag => (
                                            <span key={tag} className="flex items-center text-xs bg-gray-50 border border-gray-200 px-2 py-0.5 rounded text-gray-700">
                                                {tag} <button className="ml-1 text-gray-400 hover:text-gray-600"><X size={10}/></button>
                                            </span>
                                        ))}
                                    </div>
                                </FormField>
                            </div>

                             <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <FormField label={t('pre.aPoint')}>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-white text-gray-400">
                                            <option>{t('pre.selectAPoint')}</option>
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('pre.vPoint')}>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-white text-gray-400">
                                             <option>{t('pre.selectVPoint')}</option>
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('pre.fallRisk')}>
                                    <div className="flex items-center h-9 gap-4">
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="fall" className="text-blue-600"/> <span className="text-sm text-gray-700">{t('fallRisk.low')}</span></label>
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="fall" defaultChecked className="text-blue-600"/> <span className="text-sm text-gray-700">{t('fallRisk.high')}</span></label>
                                    </div>
                                </FormField>
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6">
                                <FormField label={t('pre.painScore')}>
                                    <input type="number" defaultValue="0" className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500"/>
                                </FormField>
                                <FormField label={t('pre.nursingLevel')}>
                                     <div className="flex items-center h-9 gap-4">
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="nursing" className="text-blue-600"/> <span className="text-sm text-gray-700">{t('nursing.critical')}</span></label>
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="nursing" className="text-blue-600"/> <span className="text-sm text-gray-700">{t('nursing.serious')}</span></label>
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="nursing" defaultChecked className="text-blue-600"/> <span className="text-sm text-gray-700">{t('nursing.other')}</span></label>
                                    </div>
                                </FormField>
                            </div>

                             <div className="grid grid-cols-1 lg:grid-cols-3 gap-x-12 gap-y-6 mb-6 pt-4">
                                <PreAssessInput label={t('pre.checkInTime')} className="lg:col-span-1" suffix={<Calendar size={14} className="text-gray-400"/>} defaultValue="07:46"/>
                                <PreAssessInput label={t('pre.admissionTime')} className="lg:col-span-1" suffix={<Calendar size={14} className="text-gray-400"/>} defaultValue="07:46"/>
                                <PreAssessInput label={t('pre.assessTime')} className="lg:col-span-1" required suffix={<Calendar size={14} className="text-gray-400"/>} defaultValue="07:47"/>

                                <FormField label={t('pre.doctor')}>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-white">
                                            <option>鏉庢晱</option>
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('pre.onMachineNurse')} required>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-white">
                                            {renderNurseOptions()}
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('pre.assessor')}>
                                    <div className="relative">
                                        <select className="w-full h-9 px-3 border border-gray-300 rounded-md text-sm outline-none focus:border-blue-500 bg-gray-50 text-gray-500" disabled>
                                            {renderNurseOptions()}
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>

                                <PreAssessInput label={t('pre.startTime')} className="lg:col-span-1" suffix={<Calendar size={14} className="text-gray-400"/>} defaultValue={currentDateTimeText}/>
                            </div>

                        </div>

                        <div className="flex justify-between items-center pt-2">
                             <button className="text-blue-600 text-sm flex items-center hover:underline">
                                <ImageIcon size={16} className="mr-1"/> {t('pre.weighPhoto')}
                                <ChevronDown size={12} className="ml-1"/>
                             </button>
                             <div className="flex gap-3">
                                <button className="px-6 py-2 bg-white border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 font-medium text-sm">{t('pre.tempSave')}</button>
                                <button
                                    onClick={() => setActiveProcessStep('prescription')}
                                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium text-sm flex items-center shadow-sm"
                                >
                                    {t('pre.submitNext')} <ArrowRight size={16} className="ml-2"/>
                                </button>
                             </div>
                        </div>
                    </div>
                );

            case 'prescription':
                return (
                    <div className="p-6 max-w-[1600px] mx-auto animate-fade-in space-y-6">

                        <div className="flex justify-between items-center bg-white px-6 py-4 rounded-xl border border-gray-200 shadow-sm">
                            <h2 className="text-xl font-bold text-gray-800 flex items-center">
                                {t('prescription.patient')}: {selectedPatient.name}
                            </h2>
                            <div className="flex space-x-3">
                                {isEditingPrescription ? (
                                    <>
                                        <button onClick={() => setIsEditingPrescription(false)} className="px-4 py-2 bg-white border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50 transition-colors">
                                            {t('prescription.cancel')}
                                        </button>
                                        <button onClick={() => setIsEditingPrescription(false)} className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 transition-colors shadow-sm flex items-center">
                                            <CheckCircle2 size={16} className="mr-2"/> {t('prescription.confirm')}
                                        </button>
                                    </>
                                ) : (
                                    <button onClick={() => setIsEditingPrescription(true)} className="flex items-center px-4 py-2 bg-white border border-blue-200 text-blue-600 rounded-lg text-sm hover:bg-blue-50 transition-colors shadow-sm">
                                        <PenTool size={16} className="mr-2"/> {t('prescription.edit')}
                                    </button>
                                )}
                                <button className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg"><HelpCircle size={20}/></button>
                            </div>
                        </div>

                        {!isEditingPrescription && (
                            <>
                                <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-8">
                                    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-y-8 gap-x-8">
                                        <InfoField label={t('prescription.method')} value="HDF" />
                                        <InfoField label={t('prescription.preWeight')} value="78.5" unit="kg" />
                                        <InfoField label={t('prescription.lastPostWeight')} value="77.2" unit="kg" />
                                        <InfoField label={t('prescription.weightGain')} value="1.3" unit="kg" />
                                        <InfoField label={t('prescription.dryWeight')} value="77" unit="kg" highlight icon={<TrendingUp size={14} className="ml-1 text-blue-600"/>} />
                                        <InfoField label={t('prescription.ufVolume')} value="1.7" unit="L (2.21%)" highlight />

                                        <InfoField label={t('prescription.thisWeightGain')} value="1.5" unit="kg (1.95%)" />
                                        <InfoField label={t('prescription.preBP')} value="110/60" unit="mmHg" />
                                        <InfoField label={t('prescription.duration')} value="3.5" unit="h" />
                                        <InfoField label={t('prescription.bloodFlow')} value="200" unit="ml/min" />
                                        <InfoField label="BV" value="-" />
                                        <InfoField label={t('prescription.heparinType')} value={t('prescription.heparinNormal')} />

                                        <div className="col-span-2 md:col-span-3 lg:col-span-6 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-y-8 gap-x-8 border-t border-gray-100 pt-6">
                                            <InfoField label={t('prescription.firstDoseName')} value="閭ｅ眻鑲濈礌閽欐敞灏勬恫(鐧惧姏鑸?-615axiu/ml" className="col-span-2" />
                                            <InfoField label={t('prescription.firstDose')} value="1230" unit="axiu" />
                                            <InfoField label={t('prescription.maintainDrug')} value="-" />
                                            <InfoField label={t('prescription.infusionRate')} value="-" />
                                            <InfoField label={t('prescription.infusionTime')} value="-" unit="h" />
                                            <InfoField label={t('prescription.maintainDose')} value="-" />
                                        </div>

                                        <div className="col-span-2 md:col-span-3 lg:col-span-6 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-y-8 gap-x-8 border-t border-gray-100 pt-6">
                                            <InfoField label={t('prescription.vascularAccess')} value="AVF-涓婅噦" />
                                            <InfoField label={t('prescription.dialysate')} value="A液+B液" />
                                            <InfoField label={t('prescription.dialysateFlow')} value="500" unit="ml/min" />
                                            <InfoField label={t('prescription.naConc')} value="138" unit="mmol/L" />
                                            <InfoField label={t('prescription.caConc')} value="1.5" unit="mmol/L" />
                                            <InfoField label={t('prescription.kConc')} value="2" unit="mmol/L" />

                                            <InfoField label={t('prescription.hco3Conc')} value="32" unit="mmol/L" />
                                            <InfoField label={t('prescription.glucoseConc')} value="1.1" unit="g/L" />
                                            <InfoField label={t('prescription.conductivity')} value="14.2" unit="mS/cm" />
                                            <InfoField label={t('prescription.dialysateTemp')} value="36.5" unit="掳C" />
                                            <InfoField label={t('prescription.dialysateVolume')} value="105" unit="L" />
                                            <InfoField label={t('prescription.notes')} value="-" />
                                        </div>
                                    </div>
                                    <div className="mt-8 pt-4 border-t border-gray-100 text-xs text-gray-400 text-right">
                                        {t('prescription.lastModified')}: {currentDateTimeText}
                                    </div>
                                </div>

                                <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                                    <table className="w-full text-left text-sm">
                                        <thead className="bg-blue-50 text-gray-700 font-bold border-b border-blue-100">
                                            <tr>
                                                <th className="px-6 py-3">{t('materials.index')}</th>
                                                <th className="px-6 py-3">{t('materials.name')}</th>
                                                <th className="px-6 py-3">{t('materials.category')}</th>
                                                <th className="px-6 py-3">{t('materials.quantity')}</th>
                                                <th className="px-6 py-3">{t('materials.code')}</th>
                                                <th className="px-6 py-3">{t('materials.brand')}</th>
                                                <th className="px-6 py-3">{t('materials.spec')}</th>
                                                <th className="px-6 py-3">{t('materials.note')}</th>
                                            </tr>
                                        </thead>
                                        <tbody className="divide-y divide-gray-100">
                                            {materials.map((m, i) => (
                                                <tr key={m.id} className="hover:bg-gray-50">
                                                    <td className="px-6 py-4 text-gray-500">{i + 1}</td>
                                                    <td className="px-6 py-4 font-medium">{m.name}</td>
                                                    <td className="px-6 py-4 text-gray-600">{m.category}</td>
                                                    <td className="px-6 py-4">{m.count}</td>
                                                    <td className="px-6 py-4 font-mono text-gray-500">{m.code}</td>
                                                    <td className="px-6 py-4">{m.brand}</td>
                                                    <td className="px-6 py-4">{m.spec}</td>
                                                    <td className="px-6 py-4 text-gray-400">{m.note || '-'}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>

                                <div className="space-y-4">
                                    <h3 className="text-sm font-bold text-gray-800 border-l-4 border-blue-500 pl-3">{t('adjustLog.title')}</h3>
                                    <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                                        <table className="w-full text-left text-sm">
                                            <thead className="bg-gray-50 text-gray-500 font-medium">
                                                <tr>
                                                    <th className="px-6 py-3">{t('materials.index')}</th>
                                                    <th className="px-6 py-3">{t('adjustLog.content')}</th>
                                                    <th className="px-6 py-3">{t('adjustLog.person')}</th>
                                                    <th className="px-6 py-3">{t('adjustLog.time')}</th>
                                                    <th className="px-6 py-3 text-right">{t('adjustLog.action')}</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-gray-100">
                                                <tr className="hover:bg-red-50/10">
                                                    <td className="px-6 py-4 text-gray-500">1</td>
                                                    <td className="px-6 py-4">
                                                        <span className="text-red-600 font-medium">瓒呮护閲? 鐢便€?銆慙 璋冩暣涓恒€?.7銆慙;</span>
                                                    </td>
                                                    <td className="px-6 py-4">{nurses[0]?.name ?? '--'}</td>
                                                    <td className="px-6 py-4 font-mono text-gray-500">{currentDateTimeText}</td>
                                                    <td className="px-6 py-4 text-right"><Settings size={14} className="inline text-gray-400"/></td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            </>
                        )}

                        {isEditingPrescription && (
                            <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-8 space-y-8 animate-fade-in">

                                <div className="flex flex-wrap gap-x-8 gap-y-4 items-center text-sm text-gray-700">
                                    <span>{t('prescription.method')}: HDF</span>
                                    <span>{t('prescription.preWeight')}: 78.5kg</span>
                                    <span>{t('prescription.lastPostWeight')}: 77.2kg</span>
                                    <span>{t('prescription.weightGain')}: 1.3kg</span>
                                    <span>{t('prescription.thisWeightGain')}(1.95%): 1.5kg</span>

                                    <PrescriptionInput label={t('prescription.dryWeight')} suffix="kg" defaultValue="77" width="w-28"/>
                                    <PrescriptionInput label={`${t('prescription.ufVolume')}(2.21%)`} suffix="L" defaultValue="1.7" width="w-28"/>
                                    <span>{t('prescription.preBP')}: 110/60mmHg</span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <input type="checkbox" defaultChecked className="rounded text-blue-600 focus:ring-blue-500 border-gray-300"/>
                                    <span className="text-sm text-gray-600">{t('prescription.syncToPlan')}</span>
                                </div>

                                <div className="flex flex-wrap gap-x-8 gap-y-6 items-center">
                                    <PrescriptionInput label={t('prescription.duration')} suffix="h" defaultValue="3.5"/>
                                    <PrescriptionInput label={t('prescription.bloodFlow')} suffix="ml/min" defaultValue="200" width="w-32"/>

                                    <div className="flex items-center space-x-4">
                                        <span className="text-sm text-gray-600">{t('prescription.heparinType')}:</span>
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="heparin" defaultChecked className="text-blue-600"/> <span className="text-sm">{t('prescription.heparinNormal')}</span></label>
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="heparin" className="text-blue-600"/> <span className="text-sm">{t('prescription.heparinRelative')}</span></label>
                                        <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="heparin" className="text-blue-600"/> <span className="text-sm">{t('prescription.heparinAbsolute')}</span></label>
                                    </div>

                                    <div className="flex items-center space-x-2">
                                        <label className="text-sm text-gray-600">{t('prescription.firstDoseName')}:</label>
                                        <div className="relative w-48">
                                            <select className="w-full h-8 border rounded text-sm px-2 bg-white appearance-none outline-none">
                                                <option>閭ｅ眻鑲濈礌閽欐敞灏?..</option>
                                            </select>
                                            <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                        </div>
                                    </div>
                                    <PrescriptionInput label={t('prescription.firstDose')} suffix="axiu" defaultValue="1230" width="w-32"/>
                                </div>

                                <div className="flex flex-wrap gap-x-8 gap-y-6 items-center">
                                    <div className="flex items-center space-x-2">
                                        <label className="text-sm text-gray-600">{t('prescription.maintName')}:</label>
                                        <div className="relative w-32">
                                            <select className="w-full h-8 border rounded text-sm px-2 bg-white appearance-none outline-none text-gray-400">
                                                <option>{t('prescription.selectMaint')}</option>
                                            </select>
                                            <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                        </div>
                                    </div>
                                    <PrescriptionInput label={t('prescription.infusionTime')} suffix="h" width="w-24"/>
                                    <PrescriptionInput label={t('prescription.infusionRate')} width="w-24"/>
                                    <PrescriptionInput label={t('prescription.maintainDose')} width="w-24"/>
                                    <PrescriptionInput label={t('prescription.totalDose')} suffix="axiu" defaultValue="1230" readOnly width="w-32"/>
                                </div>

                                <div className="flex flex-wrap gap-x-8 gap-y-6 items-center pb-6 border-b border-gray-100">
                                    <div className="flex items-center space-x-2">
                                        <label className="text-sm text-gray-600"><span className="text-red-500 mr-1">*</span>{t('prescription.vascularAccess')}:</label>
                                        <div className="relative w-32">
                                            <select className="w-full h-8 border rounded text-sm px-2 bg-white appearance-none outline-none">
                                                <option>AVF-涓婅噦</option>
                                            </select>
                                            <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                        </div>
                                        <button className="text-sm text-gray-600 hover:text-blue-600">{t('prescription.view')}</button>
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        <label className="text-sm text-gray-600">{t('prescription.dialysateType')}:</label>
                                        <div className="relative w-32">
                                            <select className="w-full h-8 border rounded text-sm px-2 bg-white appearance-none outline-none">
                                                <option>A娑?B娑</option>
                                            </select>
                                            <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                        </div>
                                    </div>
                                </div>

                                <div>
                                    <div className="flex justify-between items-center mb-4">
                                        <div className="flex items-center space-x-2 w-full max-w-md">
                                            <div className="relative flex-1">
                                                <select className="w-full h-9 border rounded-lg px-3 bg-white appearance-none outline-none text-gray-400 text-sm">
                                                    <option>{t('prescription.select')}</option>
                                                </select>
                                                <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                            </div>
                                            <button className="flex items-center px-4 py-2 bg-gray-50 border border-gray-200 text-gray-600 rounded-lg text-sm hover:bg-gray-100 transition-colors">
                                                <Plus size={16} className="mr-1"/> {t('prescription.add')}
                                            </button>
                                        </div>
                                        <button className="flex items-center px-4 py-2 text-red-500 hover:bg-red-50 rounded-lg text-sm transition-colors border border-transparent hover:border-red-100">
                                            <Trash2 size={16} className="mr-1"/> {t('prescription.delete')}
                                        </button>
                                    </div>

                                    <div className="bg-blue-50/30 rounded-xl border border-blue-100 overflow-hidden">
                                        <table className="w-full text-left text-sm">
                                            <thead className="bg-blue-100/50 text-gray-700 font-bold border-b border-blue-200">
                                                <tr>
                                                    <th className="px-4 py-3 w-12"><input type="checkbox" className="rounded text-blue-600"/></th>
                                                    <th className="px-4 py-3">{t('materials.index')}</th>
                                                    <th className="px-4 py-3">{t('materials.name')}</th>
                                                    <th className="px-4 py-3">{t('materials.category')}</th>
                                                    <th className="px-4 py-3">{t('materials.quantity')}</th>
                                                    <th className="px-4 py-3">{t('materials.code')}</th>
                                                    <th className="px-4 py-3">{t('materials.brand')}</th>
                                                    <th className="px-4 py-3">{t('materials.spec')}</th>
                                                    <th className="px-4 py-3 text-right">{t('materials.note')}</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-blue-100/50 bg-white">
                                                {materials.map((m, i) => (
                                                    <tr key={m.id} className="hover:bg-blue-50/30">
                                                        <td className="px-4 py-3"><input type="checkbox" className="rounded text-blue-600"/></td>
                                                        <td className="px-4 py-3 text-gray-500">{i + 1}</td>
                                                        <td className="px-4 py-3 font-medium text-gray-800">{m.name}</td>
                                                        <td className="px-4 py-3 text-gray-600">{m.category}</td>
                                                        <td className="px-4 py-3">
                                                            <input type="number" defaultValue={m.count} className="w-16 h-7 border border-gray-300 rounded px-2 text-center focus:ring-1 focus:ring-blue-500 outline-none"/>
                                                        </td>
                                                        <td className="px-4 py-3 font-mono text-gray-400 text-xs">{m.code}</td>
                                                        <td className="px-4 py-3">{m.brand}</td>
                                                        <td className="px-4 py-3">{m.spec}</td>
                                                        <td className="px-4 py-3 text-right">
                                                            <button className="text-blue-600 hover:underline text-xs">{t('prescription.modify')}</button>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            </div>
                        )}

                        {!isEditingPrescription && (
                            <div className="fixed bottom-8 right-8">
                                <button onClick={() => setActiveProcessStep('check1')} className="w-12 h-12 bg-white rounded-full text-gray-600 shadow-lg hover:text-blue-600 flex items-center justify-center transition-transform hover:scale-105 border border-gray-200">
                                    <User size={20}/>
                                </button>
                            </div>
                        )}
                    </div>
                );

            case 'check1':
                 return (
                    <div className="flex flex-col h-full bg-white animate-fade-in relative">
                        <div className="flex justify-between items-center px-6 py-3 bg-blue-50/50 border-b border-blue-100 shrink-0">
                            <div className="flex items-center text-sm text-blue-800"><Info size={16} className="mr-2 text-blue-600"/> {t('step.check1')}</div>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6 pb-24">
                            <div className="border border-gray-200 rounded-lg overflow-hidden shadow-sm mb-6">
                                <div className="flex bg-blue-100/50 text-sm font-bold text-gray-700 border-b border-gray-200"><div className="w-12 py-3 border-r"></div><div className="w-64 py-3 px-4 border-r">{t('check.content')}</div><div className="flex-1 py-3 px-4 border-r">{t('check.status')}</div><div className="w-32 py-3 px-4 border-r">{t('check.person')}</div><div className="w-48 py-3 px-4">{t('check.time')}</div></div>
                                <CheckRow nurseOptions={renderNurseOptions()} label={`${t('check.dialysisMode')} :HD`} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={`${t('check.dialyzer')} :SUREFLUX-15G`} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.bloodLine')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.priming')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.prescriptionCheck')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.patientId')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                            </div>
                        </div>
                        <div className="fixed bottom-8 right-8"><button onClick={() => setActiveProcessStep('check2')} className="w-12 h-12 bg-white rounded-full text-gray-600 shadow-lg hover:text-blue-600 flex items-center justify-center transition-transform hover:scale-105 border border-gray-200"><User size={20}/></button></div>
                    </div>
                );

            case 'check2':
                 return (
                    <div className="flex flex-col h-full bg-white animate-fade-in relative">
                        <div className="flex justify-between items-center px-6 py-3 bg-blue-50/50 border-b border-blue-100 shrink-0">
                            <div className="flex items-center text-sm text-blue-800"><Info size={16} className="mr-2 text-blue-600"/> {t('step.check2')}</div>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6 pb-24">
                            <div className="border border-gray-200 rounded-lg overflow-hidden shadow-sm mb-6">
                                <div className="flex bg-blue-100/50 text-sm font-bold text-gray-700 border-b border-gray-200"><div className="w-12 py-3 border-r"></div><div className="w-64 py-3 px-4 border-r">{t('check.content')}</div><div className="flex-1 py-3 px-4 border-r">{t('check.status')}</div><div className="w-32 py-3 px-4 border-r">{t('check.person')}</div><div className="w-48 py-3 px-4">{t('check.time')}</div></div>
                                <CheckRow nurseOptions={renderNurseOptions()} label={`${t('check.dialysisMode')} :HD`} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.prescriptionCheck')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.anticoagulant')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.vascularCheck')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                                <CheckRow nurseOptions={renderNurseOptions()} label={t('check.lineConnection')} normalText={t('check.normal')} abnormalText={t('check.abnormal')} abnormalPlaceholder={t('check.abnormalReason')} />
                            </div>
                            <div className="bg-blue-50/20 p-6 rounded-xl border border-blue-100/50 space-y-6">
                                <div className="grid grid-cols-1 md:grid-cols-3 gap-x-12 gap-y-6">
                                    <FormField label={t('check.onMachineNurse')} required><select className="w-full bg-white border border-gray-200 rounded-lg px-3 py-2">{renderNurseOptions()}</select></FormField>
                                    <FormField label={t('check.reCheckNurse')}><select className="w-full bg-white border border-gray-200 rounded-lg px-3 py-2">{renderNurseOptions()}</select></FormField>
                                    <FormField label={t('check.qcNurse')}><select className="w-full bg-white border border-gray-200 rounded-lg px-3 py-2">{renderNurseOptions()}</select></FormField>
                                </div>
                            </div>
                        </div>
                        <div className="fixed bottom-8 right-8"><button onClick={() => setActiveProcessStep('orders_process')} className="w-12 h-12 bg-white rounded-full text-gray-600 shadow-lg hover:text-blue-600 flex items-center justify-center transition-transform hover:scale-105 border border-gray-200"><Syringe size={20}/></button></div>
                    </div>
                );

            case 'orders_process':
                 return (
                    <div className="flex flex-col h-full bg-white animate-fade-in relative">
                        <div className="px-6 py-4 border-b border-gray-100 flex items-center space-x-3 bg-gray-50/50">
                            <button onClick={() => setShowOrderModal(true)} className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 shadow-sm transition-colors"><Plus size={16} className="mr-2"/> {t('orders.add')}</button>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6 space-y-8">
                            <div className="border border-gray-200 rounded-lg overflow-hidden shadow-sm">
                                <table className="w-full text-left text-sm">
                                    <thead className="bg-blue-100/50 text-gray-700 font-bold"><tr><th className="px-4 py-3">{t('orders.type')}</th><th className="px-4 py-3">{t('orders.content')}</th><th className="px-4 py-3">{t('orders.doctorTime')}</th><th className="px-4 py-3 text-center">{t('orders.status')}</th><th className="px-4 py-3 text-right">{t('orders.action')}</th></tr></thead>
                                    <tbody className="divide-y divide-gray-100">
                                        {([] as {id:number, type:string, content:string, doctor:string, time:string, status:string}[]).map(o => (
                                            <tr key={o.id} className="hover:bg-blue-50/30">
                                                <td className="px-4 py-3 text-gray-800">{o.type}</td>
                                                <td className="px-4 py-3 font-medium">{o.content}</td>
                                                <td className="px-4 py-3 text-gray-600 text-xs">{o.doctor} {o.time}</td>
                                                <td className="px-4 py-3 text-center text-xs">{o.status}</td>
                                                <td className="px-4 py-3 text-right"><button className="text-blue-600 hover:underline text-xs">{t('orders.execute')}</button></td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        {showOrderModal && (
                            <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in">
                                <div className="bg-white rounded-lg shadow-2xl w-full max-w-2xl p-6">
                                    <div className="flex justify-between mb-4"><h3 className="font-bold">{t('orders.addTitle')}</h3><button onClick={()=>setShowOrderModal(false)}><X/></button></div>
                                    <div className="space-y-4"><input className="w-full border p-2 rounded" placeholder={t('orders.contentPlaceholder')}/><textarea className="w-full border p-2 rounded" placeholder={t('orders.notePlaceholder')}></textarea></div>
                                    <div className="mt-4 flex justify-end gap-2"><button onClick={()=>setShowOrderModal(false)} className="px-4 py-2 border rounded">{t('common.cancel')}</button><button onClick={()=>setShowOrderModal(false)} className="px-4 py-2 bg-blue-600 text-white rounded">{t('common.ok')}</button></div>
                                </div>
                            </div>
                        )}
                        <div className="fixed bottom-8 right-8"><button onClick={handleStartDialysis} className="w-12 h-12 bg-white rounded-full text-gray-600 shadow-lg hover:text-blue-600 flex items-center justify-center transition-transform hover:scale-105 border border-gray-200"><Activity size={20}/></button></div>
                    </div>
                );

             case 'monitor':
                 return (
                    <div className="flex flex-col h-full bg-white animate-fade-in relative">
                        <div className="px-6 py-4 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
                            <div className="flex items-center space-x-3">
                                <button onClick={() => setShowMonitorModal(true)} className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 shadow-sm transition-colors"><Plus size={16} className="mr-2"/> {t('monitor.add')}</button>
                                <button className="flex items-center px-4 py-2 bg-white border border-gray-300 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors">{t('monitor.deleteCollectBP')}</button>
                            </div>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6 pb-24">
                            <div className="border border-gray-200 rounded-lg overflow-hidden shadow-sm">
                                <table className="w-full text-left text-sm whitespace-nowrap">
                                    <thead className="bg-blue-100/50 text-gray-700 font-bold">
                                        <tr>
                                            <th className="px-4 py-3">{t('monitor.index')}</th>
                                            <th className="px-4 py-3 flex items-center">{t('monitor.observeTime')} <ChevronsDown size={14} className="ml-1 text-gray-400"/></th>
                                            <th className="px-4 py-3">{t('monitor.sbp')}</th><th className="px-4 py-3">{t('monitor.dbp')}</th><th className="px-4 py-3">{t('monitor.hr')}</th><th className="px-4 py-3">{t('monitor.temp')}</th><th className="px-4 py-3">{t('monitor.resp')}</th><th className="px-4 py-3">{t('monitor.spo2')}</th><th className="px-4 py-3">{t('monitor.ufVolume')}</th><th className="px-4 py-3">{t('monitor.bloodFlow')}</th><th className="px-4 py-3">{t('monitor.ap')}</th><th className="px-4 py-3">{t('monitor.vp')}</th><th className="px-4 py-3">{t('monitor.tmp')}</th><th className="px-4 py-3">{t('monitor.machineTemp')}</th><th className="px-4 py-3">{t('monitor.nurse')}</th><th className="px-4 py-3 text-center sticky right-0 bg-blue-50">{t('monitor.action')}</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-100">
                                        {monitorRecords.map((record, i) => (
                                            <tr key={record.id} className="hover:bg-blue-50/30">
                                                <td className="px-4 py-3 text-gray-500">{monitorRecords.length - i}</td>
                                                <td className="px-4 py-3 font-mono text-gray-800">{record.time}</td>
                                                <td className="px-4 py-3 font-medium">{record.sbp}</td>
                                                <td className="px-4 py-3 font-medium">{record.dbp}</td>
                                                <td className="px-4 py-3">{record.hr}</td>
                                                <td className="px-4 py-3">{record.temp}</td>
                                                <td className="px-4 py-3">{record.resp}</td>
                                                <td className="px-4 py-3">{record.spO2}</td>
                                                <td className="px-4 py-3">{record.ufVolume}</td>
                                                <td className="px-4 py-3">{record.bloodFlow}</td>
                                                <td className="px-4 py-3">{record.arterialPressure}</td>
                                                <td className="px-4 py-3">{record.venousPressure}</td>
                                                <td className="px-4 py-3">{record.transmembranePressure}</td>
                                                <td className="px-4 py-3">{record.machineTemp}</td>
                                                <td className="px-4 py-3">{record.nurse}</td>
                                                <td className="px-4 py-3 text-center sticky right-0 bg-white group-hover:bg-blue-50/30">
                                                    <div className="flex items-center justify-center space-x-3">
                                                        <button onClick={() => setShowMonitorModal(true)} className="text-blue-600 hover:text-blue-800 text-xs">{t('monitor.edit')}</button>
                                                        <span className="text-gray-300">|</span>
                                                        <button className="text-red-500 hover:text-red-700 text-xs">{t('monitor.delete')}</button>
                                                    </div>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        {showMonitorModal && (
                            <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in p-4">
                                <div className="bg-white rounded-xl shadow-2xl w-full max-w-6xl overflow-hidden animate-scale-in flex flex-col max-h-[95vh]">
                                    <div className="flex justify-between items-center px-6 py-4 border-b border-gray-100 bg-gray-50 shrink-0">
                                        <h3 className="font-bold text-gray-800 text-lg">{t('monitor.title')}</h3>
                                        <button onClick={() => setShowMonitorModal(false)} className="text-gray-400 hover:text-gray-600 transition-colors"><X size={24}/></button>
                                    </div>

                                    <div className="p-8 overflow-y-auto bg-blue-50/10">
                                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-x-8 gap-y-6">
                                            <MonitorInput label={t('monitor.sbp')} suffix="mmHg" defaultValue="133"/>
                                            <MonitorInput label={t('monitor.dbp')} suffix="mmHg" defaultValue="73"/>
                                            <MonitorInput label={t('monitor.hr')} suffix="次/分" defaultValue="65"/>
                                            <MonitorInput label={t('monitor.temp')} suffix="℃"/>

                                            <MonitorInput label={t('monitor.resp')} suffix="次/分" required defaultValue="16"/>
                                            <MonitorInput label={t('monitor.spO2Full')} suffix="%"/>
                                            <MonitorInput label={t('monitor.bloodFlow')} suffix="ml/min" required defaultValue="240"/>
                                            <MonitorInput label={t('monitor.ap')} suffix="mmHg" defaultValue="-114"/>

                                            <MonitorInput label={t('monitor.vp')} suffix="mmHg" required defaultValue="102"/>
                                            <MonitorInput label={t('monitor.tmp')} suffix="mmHg" required defaultValue="55"/>
                                            <MonitorInput label={t('monitor.machineTemp')} suffix="℃" required defaultValue="35.8"/>
                                            <MonitorInput label={t('monitor.ufVolume')} suffix="ml" required defaultValue="234"/>

                                            <MonitorInput label={t('monitor.conductivity')} suffix="ms/cm" required defaultValue="13.86"/>
                                            <MonitorInput label={t('monitor.heparinFlow')} suffix="ml/h" defaultValue="0"/>
                                            <MonitorInput label={t('monitor.relativeBV')} suffix="%" defaultValue="100"/>
                                            <MonitorInput label={t('monitor.realTimeBV')} suffix="mL" defaultValue="0"/>

                                            <MonitorInput label={t('monitor.realTimeClearance')} suffix="mL/min" defaultValue="0"/>
                                            <MonitorInput label={t('monitor.arterialTemp')} suffix="℃" defaultValue="0"/>
                                            <MonitorInput label={t('monitor.symptoms')} className="col-span-1"/>
                                            <MonitorInput label={t('monitor.symptomType')} className="col-span-1"/>

                                            <div className="flex items-center col-span-2">
                                                <label className="w-24 text-right text-sm text-gray-600 mr-2 shrink-0">{t('monitor.doctor')}:</label>
                                                <div className="relative flex-1">
                                                    <select className="w-full h-8 border border-gray-300 rounded text-sm px-2 outline-none appearance-none bg-white text-gray-800">
                                                        <option>{t('monitor.selectDoctor')}</option>
                                                    </select>
                                                    <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                                </div>
                                            </div>

                                            <div className="flex items-center col-span-2">
                                                <label className="w-32 text-right text-sm text-gray-600 mr-2 shrink-0">{t('monitor.apFix')}:</label>
                                                <div className="flex space-x-4">
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="ap_fix" defaultChecked className="text-blue-600"/> <span className="text-sm">{t('monitor.good')}</span></label>
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="ap_fix" className="text-blue-600"/> <span className="text-sm">{t('monitor.bad')}</span></label>
                                                </div>
                                            </div>
                                            <div className="flex items-center col-span-2">
                                                <label className="w-32 text-right text-sm text-gray-600 mr-2 shrink-0">{t('monitor.apSeep')}:</label>
                                                <div className="flex space-x-4">
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="ap_seep" className="text-blue-600"/> <span className="text-sm">{t('monitor.yes')}</span></label>
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="ap_seep" defaultChecked className="text-blue-600"/> <span className="text-sm">{t('monitor.no')}</span></label>
                                                </div>
                                            </div>

                                            <div className="flex items-center col-span-2">
                                                <label className="w-32 text-right text-sm text-gray-600 mr-2 shrink-0">{t('monitor.vpFix')}:</label>
                                                <div className="flex space-x-4">
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="vp_fix" defaultChecked className="text-blue-600"/> <span className="text-sm">{t('monitor.good')}</span></label>
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="vp_fix" className="text-blue-600"/> <span className="text-sm">{t('monitor.bad')}</span></label>
                                                </div>
                                            </div>
                                            <div className="flex items-center col-span-2">
                                                <label className="w-32 text-right text-sm text-gray-600 mr-2 shrink-0">{t('monitor.vpSeep')}:</label>
                                                <div className="flex space-x-4">
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="vp_seep" className="text-blue-600"/> <span className="text-sm">{t('monitor.yes')}</span></label>
                                                    <label className="flex items-center space-x-1 cursor-pointer"><input type="radio" name="vp_seep" defaultChecked className="text-blue-600"/> <span className="text-sm">{t('monitor.no')}</span></label>
                                                </div>
                                            </div>

                                            <div className="col-span-2"><MonitorInput label={t('monitor.note')} className="w-full" /></div>

                                            <div className="flex items-center col-span-2">
                                                <label className="w-24 text-right text-sm text-gray-600 mr-2 shrink-0 whitespace-nowrap"><span className="text-red-500 mr-1">*</span>{t('monitor.observeTime')}:</label>
                                                <div className="relative flex-1">
                                                    <input type="text" defaultValue={currentDateTimeText} className="w-full h-8 px-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
                                                    <Calendar size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400"/>
                                                </div>
                                            </div>
                                            <div className="flex items-center col-span-2">
                                                <label className="w-24 text-right text-sm text-gray-600 mr-2 shrink-0 whitespace-nowrap"><span className="text-red-500 mr-1">*</span>{t('monitor.observer')}:</label>
                                                <div className="relative flex-1">
                                                    <select className="w-full h-8 border border-gray-300 rounded text-sm px-2 outline-none appearance-none bg-white text-gray-800">
                                                        {renderNurseOptions()}
                                                    </select>
                                                    <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-between items-center shrink-0">
                                        <div className="flex items-center space-x-4">
                                            <button className="px-4 py-2 bg-indigo-500 text-white rounded text-sm font-medium hover:bg-indigo-600 shadow-sm transition-colors">
                                                {t('monitor.getCollectData')}
                                            </button>
                                            <span className="text-xs text-blue-600 font-medium">{t('monitor.dialysisStart')}: {currentDateTimeText}</span>
                                        </div>
                                        <div className="flex space-x-3">
                                            <button onClick={() => setShowMonitorModal(false)} className="px-6 py-2 bg-white border border-gray-300 text-gray-700 rounded text-sm font-medium hover:bg-gray-100 transition-colors">{t('common.cancel')}</button>
                                            <button onClick={() => setShowMonitorModal(false)} className="px-6 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 shadow-sm transition-colors">{t('common.ok')}</button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                        <div className="fixed bottom-8 right-8"><button onClick={() => setActiveProcessStep('post')} className="w-12 h-12 bg-white rounded-full text-gray-600 shadow-lg hover:text-blue-600 flex items-center justify-center transition-transform hover:scale-105 border border-gray-200"><Stethoscope size={20}/></button></div>
                    </div>
                );

            case 'post':
                 return (
                    <div className="p-6 max-w-7xl mx-auto animate-fade-in space-y-6">
                        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-8">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-x-8 gap-y-6 mb-8">
                                <FormField label={t('post.startTime')} required>
                                    <div className="relative">
                                        <input type="datetime-local" defaultValue={currentDateTimeLocal} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('post.offMachineTime')} required>
                                    <div className="relative">
                                        <input type="datetime-local" defaultValue={currentDateTimeLocal} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('post.actualDuration')}>
                                    <input type="text" defaultValue="3小时30分" disabled className="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-500"/>
                                </FormField>

                                <FormField label={t('post.actualUF')} required>
                                    <div className="flex items-center gap-2">
                                        <div className="relative flex-1">
                                            <input type="number" defaultValue="337" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8"/>
                                            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">ml</span>
                                        </div>
                                        <span className="text-xs text-blue-500 whitespace-nowrap">{t('post.prescriptionUF')}:1700ml</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.actualReplacement')}>
                                    <div className="flex items-center gap-2">
                                        <div className="relative flex-1">
                                            <input type="number" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8"/>
                                            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">ml</span>
                                        </div>
                                        <span className="text-xs text-blue-500 whitespace-nowrap">{t('post.prescriptionReplacement')}:0ml</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.postWeight')}>
                                    <div className="flex items-center gap-2">
                                        <div className="relative flex-1">
                                            <input type="number" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8"/>
                                            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">kg</span>
                                        </div>
                                        <label className="flex items-center text-xs text-gray-600 whitespace-nowrap cursor-pointer">
                                            <input type="checkbox" className="mr-1 rounded text-blue-600"/> {t('pre.refuseMeasure')}
                                        </label>
                                         <label className="flex items-center text-xs text-gray-600 whitespace-nowrap cursor-pointer">
                                            <input type="checkbox" className="mr-1 rounded text-blue-600"/> {t('pre.bedridden')}
                                        </label>
                                    </div>
                                </FormField>

                                <FormField label={t('post.extraWeight')}>
                                    <div className="relative">
                                        <input type="number" defaultValue="12" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8"/>
                                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">kg</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.weightLoss')}>
                                    <div className="relative">
                                        <input type="number" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8 bg-gray-50"/>
                                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">kg</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.actualIntake')}>
                                    <div className="relative">
                                        <input type="number" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8"/>
                                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">ml</span>
                                    </div>
                                </FormField>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-3 gap-x-8 gap-y-6 mb-8">
                                <FormField label={t('post.bpSite')}>
                                    <div className="relative">
                                        <select className="w-full appearance-none border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
                                            <option>{t('bp.rightArm')}</option>
                                            <option>{t('bp.leftArm')}</option>
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('post.postTemp')} required>
                                    <div className="relative">
                                        <input type="number" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-8"/>
                                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">掳C</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.postResp')} required>
                                    <div className="relative">
                                        <input type="number" defaultValue="16" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-10"/>
                                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">娆?鍒</span>
                                    </div>
                                </FormField>

                                <FormField label={t('post.postHR')} required>
                                    <div className="relative">
                                        <input type="number" defaultValue="65" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none pr-10"/>
                                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs">娆?鍒</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.postBP')} required>
                                    <div className="flex gap-2 items-center">
                                        <input type="number" defaultValue="133" className="w-1/2 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none text-center"/>
                                        <span className="text-gray-400">/</span>
                                        <input type="number" defaultValue="73" className="w-1/2 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none text-center"/>
                                        <span className="text-xs text-gray-500 whitespace-nowrap ml-1">mmHg</span>
                                    </div>
                                </FormField>
                                <FormField label={t('post.note')}>
                                    <textarea placeholder={t('post.note')} rows={1} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none"/>
                                </FormField>
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-x-12 gap-y-6 mb-8 border-t border-gray-100 pt-6">
                                <FormField label={t('post.dialyzerCoag')} required>
                                    <div className="flex gap-4 pt-1.5">
                                        {[{key:'post.level0', val:t('post.level0')}, {key:'post.level1', val:t('post.level1')}, {key:'post.level2', val:t('post.level2')}, {key:'post.level3', val:t('post.level3')}].map(lvl => (
                                            <label key={lvl.key} className="flex items-center text-sm text-gray-700 cursor-pointer">
                                                <input type="radio" name="dialyzer_coag" className="mr-1.5 text-blue-600 focus:ring-blue-500 border-gray-300" defaultChecked={lvl.key==='post.level0'}/> {lvl.val}
                                            </label>
                                        ))}
                                    </div>
                                </FormField>
                                <FormField label={t('post.lineACoag')} required>
                                    <div className="flex gap-4 pt-1.5">
                                        {[{key:'post.level0', val:t('post.level0')}, {key:'post.level1', val:t('post.level1')}, {key:'post.level2', val:t('post.level2')}, {key:'post.level3', val:t('post.level3')}].map(lvl => (
                                            <label key={lvl.key} className="flex items-center text-sm text-gray-700 cursor-pointer">
                                                <input type="radio" name="line_a_coag" className="mr-1.5 text-blue-600 focus:ring-blue-500 border-gray-300" defaultChecked={lvl.key==='post.level0'}/> {lvl.val}
                                            </label>
                                        ))}
                                    </div>
                                </FormField>
                                <FormField label={t('post.lineVCoag')} required>
                                    <div className="flex gap-4 pt-1.5">
                                        {[{key:'post.level0', val:t('post.level0')}, {key:'post.level1', val:t('post.level1')}, {key:'post.level2', val:t('post.level2')}, {key:'post.level3', val:t('post.level3')}].map(lvl => (
                                            <label key={lvl.key} className="flex items-center text-sm text-gray-700 cursor-pointer">
                                                <input type="radio" name="line_v_coag" className="mr-1.5 text-blue-600 focus:ring-blue-500 border-gray-300" defaultChecked={lvl.key==='post.level0'}/> {lvl.val}
                                            </label>
                                        ))}
                                    </div>
                                </FormField>

                                <FormField label={t('post.fistulaCare')}>
                                    <div className="flex gap-6 pt-1.5">
                                        <label className="flex items-center text-sm text-gray-700 cursor-pointer">
                                            <input type="radio" name="care" className="mr-1.5 text-blue-600" defaultChecked/> {t('monitor.yes')}
                                        </label>
                                        <label className="flex items-center text-sm text-gray-700 cursor-pointer">
                                            <input type="radio" name="care" className="mr-1.5 text-blue-600"/> {t('monitor.no')}
                                        </label>
                                    </div>
                                </FormField>

                                <FormField label={t('post.accident')}>
                                    <div className="flex items-center gap-2">
                                        <div className="px-3 py-1.5 bg-gray-100 rounded text-sm text-gray-700 flex items-center">
                                            {t('post.none')} <button className="ml-2 text-gray-400 hover:text-gray-600"><X size={12}/></button>
                                        </div>
                                        <input type="text" className="flex-1 border-b border-gray-200 py-1 text-sm outline-none focus:border-blue-500" placeholder={t('post.otherInput')}/>
                                    </div>
                                </FormField>

                                <FormField label={t('post.dialysisEvent')}>
                                    <div className="flex gap-6 pt-1.5">
                                        <label className="flex items-center text-sm text-gray-700 cursor-pointer">
                                            <input type="radio" name="event" className="mr-1.5 text-blue-600"/> {t('monitor.yes')}
                                        </label>
                                        <label className="flex items-center text-sm text-gray-700 cursor-pointer">
                                            <input type="radio" name="event" className="mr-1.5 text-blue-600" defaultChecked/> {t('monitor.no')}
                                        </label>
                                    </div>
                                </FormField>
                            </div>

                            <div className="mb-8 pb-6 border-b border-gray-100">
                                <FormField label={t('post.fistulaStatus')}>
                                    <div className="flex flex-wrap gap-3">
                                        {[t('pre.murmurStrong'), t('pre.thrillStrong'), t('pre.pulseStrong')].map(tag => (
                                            <span key={tag} className="px-3 py-1.5 bg-gray-50 border border-gray-200 rounded text-sm text-gray-700 flex items-center">
                                                {tag} <button className="ml-2 text-gray-400 hover:text-gray-600"><X size={12}/></button>
                                            </span>
                                        ))}
                                        <button className="text-blue-600 text-sm flex items-center hover:bg-blue-50 px-2 rounded"><Plus size={14} className="mr-1"/> {t('prescription.add')}</button>
                                    </div>
                                </FormField>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-3 gap-8 items-end">
                                <div className="col-span-1 md:col-span-3 text-right">
                                    <span className="text-xs text-red-500 font-medium">{t('post.actualBPTime')}: {currentDateTimeText}</span>
                                </div>

                                <FormField label={t('post.assessTime')} required>
                                    <div className="relative">
                                        <input type="time" defaultValue="08:56" className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"/>
                                    </div>
                                </FormField>
                                <FormField label={t('post.onMachineNurse')}>
                                    <div className="relative">
                                        <select className="w-full appearance-none border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
                                            {renderNurseOptions()}
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                    <div className="mt-2 flex gap-4 text-xs text-blue-600">
                                        <button className="flex items-center hover:underline"><ImageIcon size={12} className="mr-1"/> {t('post.gradeRef')}</button>
                                        <button className="flex items-center hover:underline"><ImageIcon size={12} className="mr-1"/> {t('post.weighPhoto')}</button>
                                    </div>
                                </FormField>
                                <FormField label={t('post.assessor')} required>
                                    <div className="relative">
                                        <select className="w-full appearance-none border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
                                            {renderNurseOptions()}
                                        </select>
                                        <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"/>
                                    </div>
                                </FormField>
                            </div>
                        </div>
                        <div className="flex justify-end gap-3"><button onClick={handleFinishDialysis} className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium text-sm flex items-center shadow-sm">{t('common.submitNext')} <ArrowRight size={16} className="ml-2"/></button></div>
                    </div>
                );

            case 'disinfect':
            case 'consumables':
            case 'education':
            case 'summary':
                return (
                    <div className="h-full flex flex-col items-center justify-center text-gray-400 bg-gray-50">
                        <Clock size={64} className="mb-4 opacity-10"/>
                        <p>{t('developing')}</p>
                        <p className="text-sm mt-2">{t('developingHint')}</p>
                    </div>
                );

            default:
                return <div className="p-6 text-center text-gray-400">{t('selectStep')}</div>;
        }
  };

  return (
    <div className="h-full flex flex-col max-w-[1920px] mx-auto overflow-hidden bg-gray-50">
        <div className="flex flex-1 overflow-hidden">
            {sidebarOpen && (
                <div className="w-72 bg-white border-r border-gray-200 flex flex-col transition-all duration-300 shrink-0">
                    <div className="p-4 border-b border-gray-200 bg-gray-50/50">
                        <div className="relative">
                            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"/>
                            <input type="text" placeholder={t('search.placeholder')} className="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-lg outline-none bg-white" value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} />
                        </div>
                    </div>
                    <div className="flex-1 overflow-y-auto">
                        {filteredPatients.map(p => (
                            <div key={p.id} onClick={() => setSelectedPatientId(p.id ?? null)} className={`p-3 border-b border-gray-50 cursor-pointer hover:bg-blue-50 transition-colors ${selectedPatientId === p.id ? 'bg-blue-50 border-l-4 border-l-blue-600' : 'border-l-4 border-l-transparent'}`}>
                                <div className="flex justify-between items-start mb-1"><span className="font-bold text-gray-800 text-sm">{p.name}</span><span className="text-xs font-bold text-blue-600 bg-blue-100 px-1.5 py-0.5 rounded">{p.bedNumber}{t('header.bed')}</span></div>
                                <div className="flex justify-between items-center text-xs text-gray-500"><span>{p.gender} {p.age}{t('header.age')}</span><span>{p.status}</span></div>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            <div className="flex-1 flex flex-col min-w-0">
                <div className="bg-white border-b border-gray-200 flex items-center justify-between px-2 h-10 shrink-0">
                    <button onClick={() => setSidebarOpen(!sidebarOpen)} className="p-1.5 hover:bg-gray-100 rounded text-gray-500">{sidebarOpen ? <ChevronsDown size={16} className="rotate-90"/> : <ChevronsDown size={16} className="-rotate-90"/>}</button>
                    <div className="flex-1 overflow-x-auto no-scrollbar mx-4 flex space-x-1">
                        {steps.map(step => (
                            <button key={step.id} onClick={() => setActiveProcessStep(step.id)} className={`flex items-center px-3 py-1.5 rounded-md text-xs font-medium whitespace-nowrap transition-colors ${activeProcessStep === step.id ? 'bg-blue-600 text-white shadow-sm' : 'text-gray-600 hover:bg-gray-100'}`}>
                                <step.icon size={14} className={`mr-1.5 ${activeProcessStep === step.id ? 'text-white' : 'text-gray-400'}`}/> {step.label}
                            </button>
                        ))}
                    </div>
                </div>
                {renderPatientHeader()}
                <div className="flex-1 overflow-y-auto bg-gray-50/50">
                    {renderProcessForm()}
                </div>
            </div>
        </div>

        {showPrintPreview && selectedPatient && (
            <PrintView
              patient={selectedPatient}
              onClose={() => setShowPrintPreview(false)}
              t={t}
              dictNameMaps={dictNameMaps}
              monitoringRecords={monitorRecords.map(r => ({
                time: r.time,
                sbp: r.sbp,
                dbp: r.dbp,
                hr: r.hr,
                vp: r.venousPressure,
                tmp: r.transmembranePressure,
                ufVolume: r.ufVolume,
                symptoms: r.symptoms || '',
                nurse: r.nurse,
              }))}
            />
        )}
    </div>
  );
};

export default DialysisProcessing;



