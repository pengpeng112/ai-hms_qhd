
import React, { useState } from 'react';
import { Patient } from '../../types';
import { 
  ChevronDown, 
  Clock, 
  X, 
  Camera, 
  CheckCircle2, 
  Scale,
  Thermometer,
  Heart,
  Droplets,
  Plus,
  Stethoscope,
  Info,
  ShieldAlert,
  Image as ImageIcon,
  Search
} from 'lucide-react';

interface Props {
  patient: Patient;
}

const PostAssessment: React.FC<Props> = ({ patient }) => {
  const [clottingGrade, setClottingGrade] = useState('0');
  const [hasEvent, setHasEvent] = useState(false);
  const [eventDescription, setEventDescription] = useState('');
  
  // 内瘘情况状态管理
  const [fistulaTags, setFistulaTags] = useState(['杂音强', '震颤强']);
  const [showFistulaMenu, setShowFistulaMenu] = useState(false);

  const toggleFistulaTag = (tag: string) => {
    if (fistulaTags.includes(tag)) {
      setFistulaTags(fistulaTags.filter(t => t !== tag));
    } else {
      setFistulaTags([...fistulaTags, tag]);
    }
  };

  return (
    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500 pb-4 space-y-4 max-w-[1400px] mx-auto">
      
      {/* 1. 治疗时间与总量摘要 - 超紧凑横条 */}
      <div className="bg-slate-50 border border-slate-200 rounded-[20px] px-6 py-3 flex items-center justify-between shadow-sm">
        <div className="flex items-center space-x-4 border-r border-slate-200 pr-6">
          <div className="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center text-white shadow-lg shadow-indigo-100">
            <Clock size={16} />
          </div>
          <div>
            <p className="text-[9px] font-black text-slate-400 uppercase tracking-widest">治疗时间</p>
            <p className="text-xs font-black text-slate-800">13:02 ~ 17:02 <span className="text-indigo-600 ml-1">(4小时)</span></p>
          </div>
        </div>

        <div className="flex-1 grid grid-cols-2 gap-10 px-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Droplets size={14} className="text-blue-500" />
              <span className="text-xs font-bold text-slate-600">实际超滤</span>
            </div>
            <div className="flex items-center h-8 w-28 bg-white border border-slate-200 rounded-lg px-2 focus-within:border-blue-400 transition-all">
              <input type="text" defaultValue="2153" className="w-full text-xs font-black text-slate-800 outline-none text-right" />
              <span className="text-[9px] text-slate-400 font-bold ml-1 uppercase">ml</span>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <ImageIcon size={14} className="text-emerald-500" />
              <span className="text-xs font-bold text-slate-600">实际置换</span>
            </div>
            <div className="flex items-center h-8 w-28 bg-white border border-slate-200 rounded-lg px-2 focus-within:border-blue-400 transition-all">
              <input type="text" placeholder="0" className="w-full text-xs font-black text-slate-800 outline-none text-right" />
              <span className="text-[9px] text-slate-400 font-bold ml-1 uppercase">ml</span>
            </div>
          </div>
        </div>
      </div>

      {/* 2. 体重与生命体征 - 紧凑 2x4 网格 */}
      <div className="bg-white border border-blue-100 rounded-[24px] p-5 shadow-sm relative">
        <div className="flex items-center justify-between mb-5">
          <div className="flex items-center space-x-2">
            <div className="w-1 h-4 bg-blue-500 rounded-full"></div>
            <h3 className="text-xs font-black text-slate-800 uppercase tracking-wider">体重与生命体征</h3>
          </div>
          <button className="flex items-center space-x-1.5 text-[9px] font-black text-blue-600 hover:text-blue-700 bg-blue-50 px-3 py-1.5 rounded-lg transition-all">
            <ImageIcon size={12} />
            <span>查看下机图片</span>
          </button>
        </div>

        <div className="grid grid-cols-4 gap-x-8 gap-y-5">
          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-400 uppercase tracking-tighter">透后体重 (KG)</label>
            <div className="flex items-center space-x-2">
              <div className="flex-1 h-9 flex items-center px-3 border border-slate-200 rounded-lg bg-slate-50/30 focus-within:bg-white focus-within:border-blue-400 transition-all">
                <input type="text" placeholder="值" className="w-full text-xs font-black text-slate-800 outline-none bg-transparent" />
                <Scale size={14} className="text-slate-300" />
              </div>
              <label className="flex items-center space-x-1 cursor-pointer shrink-0">
                <input type="checkbox" className="w-3.5 h-3.5 rounded border-slate-300 text-blue-600" />
                <span className="text-[10px] font-bold text-slate-400">卧床</span>
              </label>
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-400 uppercase tracking-tighter">体重丢失 (KG)</label>
            <div className="h-9 flex items-center px-3 bg-slate-50 border border-slate-100 rounded-lg text-xs font-black text-slate-400 italic">
              -- 自动计算 --
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-tighter flex items-center">
              <span className="text-red-500 mr-1">*</span>透后血压 (MMHG)
            </label>
            <div className="flex items-center space-x-1.5">
              <input type="text" defaultValue="148" className="w-full h-9 border border-slate-200 rounded-lg text-center text-xs font-black text-slate-800 outline-none focus:border-blue-400" />
              <span className="text-slate-300 font-bold">/</span>
              <input type="text" defaultValue="80" className="w-full h-9 border border-slate-200 rounded-lg text-center text-xs font-black text-slate-800 outline-none focus:border-blue-400" />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-tighter">测压部位</label>
            <div className="relative">
              <select className="w-full h-9 px-3 border border-slate-200 rounded-lg text-xs font-black text-slate-800 bg-white appearance-none outline-none focus:border-blue-400">
                <option>右上肢</option>
                <option>左上肢</option>
              </select>
              <ChevronDown size={12} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-tighter">额外体重 (KG)</label>
            <div className="h-9 flex items-center px-3 bg-white border border-slate-200 rounded-lg focus-within:border-blue-400">
              <input type="text" defaultValue="0" className="w-full text-xs font-black text-slate-800 outline-none" />
              <span className="text-[9px] text-slate-300 font-black">KG</span>
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-tighter">实际摄入 (ML)</label>
            <div className="h-9 flex items-center px-3 bg-white border border-slate-200 rounded-lg focus-within:border-blue-400">
              <input type="text" placeholder="0" className="w-full text-xs font-black text-slate-800 outline-none" />
              <Droplets size={12} className="text-blue-200" />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-tighter flex items-center">
              <span className="text-red-500 mr-1">*</span>透后心率
            </label>
            <div className="h-9 flex items-center px-3 bg-white border border-slate-200 rounded-lg focus-within:border-blue-400">
              <input type="text" defaultValue="65" className="w-full text-xs font-black text-slate-800 outline-none" />
              <Heart size={14} className="text-red-400" />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-tighter flex items-center">
              <span className="text-red-500 mr-1">*</span>透后体温 (°C)
            </label>
            <div className="h-9 flex items-center px-3 bg-white border border-slate-200 rounded-lg focus-within:border-blue-400">
              <input type="text" defaultValue="36.5" className="w-full text-xs font-black text-slate-800 outline-none" />
              <Thermometer size={14} className="text-orange-400" />
            </div>
          </div>
        </div>
      </div>

      {/* 3. 临床观察与内瘘 - 左右分栏紧凑版 */}
      <div className="bg-slate-50/50 border border-slate-100 rounded-[24px] p-5 space-y-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="w-1 h-4 bg-orange-500 rounded-full"></div>
            <h3 className="text-xs font-black text-slate-800 uppercase tracking-wider">临床观察与记录</h3>
          </div>
          <button className="flex items-center space-x-1.5 text-[9px] font-black text-blue-500 bg-white border border-blue-100 px-3 py-1.5 rounded-lg hover:bg-blue-50 transition-all">
            <Stethoscope size={12} />
            <span>查看凝血分级参照图</span>
          </button>
        </div>

        {/* 凝血分级 - 紧凑行 */}
        <div className="bg-white border border-slate-100 rounded-xl px-5 py-3 flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <label className="text-[11px] font-black text-slate-500">凝血分级：</label>
            <div className="flex items-center space-x-6">
              {['0级', '1级', '2级', '3级'].map((level, i) => (
                <label key={level} className="flex items-center space-x-2 cursor-pointer group">
                  <div className={`w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${
                    clottingGrade === String(i) ? 'border-blue-600 bg-blue-600' : 'border-slate-200 bg-white'
                  }`}>
                    {clottingGrade === String(i) && <div className="w-1 h-1 bg-white rounded-full"></div>}
                  </div>
                  <input type="radio" className="hidden" checked={clottingGrade === String(i)} onChange={() => setClottingGrade(String(i))} />
                  <span className={`text-xs font-black ${clottingGrade === String(i) ? 'text-blue-600' : 'text-slate-400'}`}>{level}</span>
                </label>
              ))}
            </div>
          </div>
          <div className="flex items-center space-x-2 text-slate-400 text-[10px] italic">
             <Info size={12} />
             <span>0级: 滤器及管路无凝血</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-8">
          {/* 事件记录 */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <label className="text-[11px] font-black text-slate-600">发生透析事件：</label>
              <div className="flex bg-white p-1 rounded-lg border border-slate-100 shadow-sm">
                <button 
                  onClick={() => setHasEvent(true)}
                  className={`px-4 py-1 text-[10px] font-black rounded-md transition-all ${hasEvent ? 'bg-red-500 text-white' : 'text-slate-400'}`}
                >
                  是
                </button>
                <button 
                  onClick={() => { setHasEvent(false); setEventDescription(''); }}
                  className={`px-4 py-1 text-[10px] font-black rounded-md transition-all ${!hasEvent ? 'bg-slate-600 text-white' : 'text-slate-400'}`}
                >
                  否
                </button>
              </div>
            </div>
            {hasEvent && (
              <textarea 
                className="w-full h-24 border border-red-100 bg-white rounded-xl p-3 text-xs font-medium text-slate-700 outline-none focus:border-red-300 resize-none shadow-inner animate-in slide-in-from-top-1" 
                placeholder="描述低血压、抽搐、过敏等异常..."
                value={eventDescription}
                onChange={(e) => setEventDescription(e.target.value)}
              ></textarea>
            )}
            {!hasEvent && (
              <div className="h-10 border border-dashed border-slate-200 rounded-xl flex items-center justify-center text-slate-300 space-x-2 bg-white/40">
                <ShieldAlert size={14} className="opacity-30" />
                <span className="text-[10px] font-bold">无透析负面事件</span>
              </div>
            )}
          </div>

          {/* 内瘘与意外 */}
          <div className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-[11px] font-black text-slate-600">内瘘情况：</label>
              <div className="relative">
                <div 
                  onClick={() => setShowFistulaMenu(!showFistulaMenu)}
                  className="min-h-[36px] p-1.5 border border-slate-200 rounded-lg bg-white flex flex-wrap gap-1.5 items-center cursor-pointer hover:border-blue-300"
                >
                  {fistulaTags.length > 0 ? (
                    fistulaTags.map(tag => (
                      <span key={tag} className="flex items-center px-2 py-0.5 bg-blue-50 text-blue-700 rounded-md text-[10px] font-black border border-blue-100">
                        {tag}
                        <X size={10} className="ml-1 text-blue-300 hover:text-red-500" onClick={(e) => { e.stopPropagation(); toggleFistulaTag(tag); }} />
                      </span>
                    ))
                  ) : (
                    <span className="text-slate-400 text-[10px] px-1">请评估内瘘...</span>
                  )}
                  <ChevronDown size={12} className="ml-auto text-slate-400" />
                </div>
                {showFistulaMenu && (
                  <div className="absolute z-50 w-full mt-1 bg-white border border-slate-100 rounded-xl shadow-xl p-3 animate-in fade-in zoom-in-95">
                    <div className="grid grid-cols-2 gap-1.5">
                      {['杂音强', '杂音弱', '震颤强', '震颤弱', '搏动强', '搏动弱', '红肿', '渗血'].map(option => (
                        <div key={option} onClick={() => toggleFistulaTag(option)} className={`px-3 py-1.5 rounded-lg text-[10px] font-bold cursor-pointer ${fistulaTags.includes(option) ? 'bg-blue-600 text-white' : 'hover:bg-slate-50 text-slate-600'}`}>{option}</div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-black text-slate-600">其他意外情况：</label>
              <div className="h-9 border border-slate-200 rounded-lg bg-white flex items-center px-3 space-x-2 focus-within:border-blue-400">
                <span className="flex items-center px-1.5 py-0.5 bg-slate-100 text-slate-500 rounded text-[9px] font-black border border-slate-200">
                  无意外 <X size={10} className="ml-1 text-slate-300 cursor-pointer" />
                </span>
                <input className="flex-1 outline-none text-[11px] font-bold text-slate-700 bg-transparent" placeholder="管路折叠、滑脱等..." />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 4. 底部执行确认 - 扁平版 */}
      <div className="pt-4 border-t border-slate-100 flex items-center justify-between">
        <div className="flex space-x-10">
          <div className="flex flex-col">
            <label className="text-[9px] font-black text-slate-400 uppercase tracking-widest">评估时间</label>
            <div className="flex items-center space-x-1.5 text-xs font-black text-slate-700 mt-0.5">
              <Clock size={14} className="text-slate-300" />
              <span>16:30</span>
            </div>
          </div>

          <div className="flex items-center space-x-3 pl-8 border-l border-slate-100">
            <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white text-[11px] font-black">
              郑
            </div>
            <div className="flex flex-col">
              <p className="text-[9px] font-black text-slate-400 uppercase tracking-widest">下机护士</p>
              <span className="text-xs font-black text-slate-700 mt-0.5">郑九盈</span>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button className="px-6 py-2.5 border border-slate-200 text-slate-400 text-[11px] font-black rounded-xl hover:bg-slate-50">
            暂存报告
          </button>
          <button className="px-12 py-3 bg-blue-600 text-white text-[11px] font-black rounded-xl hover:bg-blue-700 shadow-lg shadow-blue-100 flex items-center space-x-2 transition-all active:scale-95">
            <CheckCircle2 size={16} />
            <span>提交结项</span>
          </button>
        </div>
      </div>

    </div>
  );
};

export default PostAssessment;
