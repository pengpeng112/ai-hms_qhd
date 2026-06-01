import { useState, useEffect, useCallback } from 'react'
import { message, Spin } from 'antd'
import { FileText, Activity, Zap, ShieldCheck, Beaker, Briefcase, Edit3, Eye, Save, Calendar } from 'lucide-react'
import { SectionHeader, DetailCard, RadioGroup, SmallInput } from '@/components/ui'
import { SummaryPrintView } from '@/components/patient/modals'
import { getMonthlySummary, saveMonthlySummary, type MonthlySummaryData } from '@/services/monthlySummaryApi'
import type { Patient } from '@/types/original'

interface MonthlySummaryTabProps {
  patient: Patient
}

export default function MonthlySummaryTab({ patient }: MonthlySummaryTabProps) {
  const [localSummaryYear, setLocalSummaryYear] = useState('2025')
  const [localSummaryMonth, setLocalSummaryMonth] = useState('06')
  const [isEditing, setIsEditing] = useState(false)
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [data, setData] = useState<MonthlySummaryData | null>(null)
  const [content, setContent] = useState<Record<string, string>>({})

  const loadData = useCallback(async () => {
    if (!patient.id) return
    setLoading(true)
    try {
      const d = await getMonthlySummary(patient.id, parseInt(localSummaryYear), parseInt(localSummaryMonth))
      setData(d)
      setContent(typeof d.content === 'object' ? d.content as Record<string, string> : {})
    } catch {
      setData(null)
      setContent({})
    } finally {
      setLoading(false)
    }
  }, [patient.id, localSummaryYear, localSummaryMonth])

  useEffect(() => { setIsEditing(false); void loadData() }, [loadData])

  const handleSave = async () => {
    if (!patient.id) return
    setSaving(true)
    try {
      const d = await saveMonthlySummary(patient.id, parseInt(localSummaryYear), parseInt(localSummaryMonth), content)
      setData(d)
      setIsEditing(false)
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const setVal = (key: string, val: string) => setContent(prev => ({ ...prev, [key]: val }))
  const getVal = (key: string, def = '') => content[key] || def
  const hasData = data && data.id > 0
  const formKey = `${localSummaryYear}-${localSummaryMonth}-${hasData ? '1' : '0'}`

  return (
    <div className="flex gap-6 animate-fade-in pb-10 h-full">
      <div className="w-24 shrink-0 flex flex-col gap-4 sticky top-0 h-fit">
        {['2025', '2024', '2023'].map(year => (
          <div key={year} className="space-y-1">
            <div className="text-[10px] font-black text-slate-300 uppercase px-2 mb-1">{year}年</div>
            <div className="flex flex-col gap-1">
              {['12','11','10','09','08','07','06','05','04','03','02','01'].map(month => (
                <button key={month} onClick={() => { setLocalSummaryYear(year); setLocalSummaryMonth(month) }}
                  className={`w-full py-2.5 px-3 rounded-xl text-xs font-black transition-all ${
                    localSummaryYear === year && localSummaryMonth === month ? 'bg-blue-600 text-white shadow-lg' : 'bg-white text-slate-500 hover:bg-slate-100 border border-slate-100'}`}>
                  {month}月</button>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="flex-1 space-y-6">
        <div className="flex justify-between items-center px-2">
          <h3 className="text-lg font-black text-slate-800">
            <FileText size={22} className="mr-3 text-blue-600 inline" /> 月份小结
            <span className="ml-4 text-blue-600 font-mono">{localSummaryYear}-{localSummaryMonth}</span>
          </h3>
          <div className="flex gap-2">
            {hasData && <button onClick={() => setIsPreviewOpen(true)} className="px-6 py-2.5 bg-white border border-slate-200 text-slate-700 rounded-2xl text-sm font-black flex items-center gap-2 shadow-sm"><Eye size={18} /> 预览</button>}
            {isEditing ? (
              <button onClick={handleSave} disabled={saving} className="px-10 py-2.5 bg-blue-600 text-white rounded-2xl text-sm font-black flex items-center gap-2"><Save size={18} /> {saving ? '保存中...' : '保存'}</button>
            ) : (
              <button onClick={() => { if (!loading) setIsEditing(true) }} className="px-10 py-2.5 bg-slate-900 text-white rounded-2xl text-sm font-black flex items-center gap-2"><Edit3 size={18} /> 编辑</button>
            )}
          </div>
        </div>

        {loading ? <div className="text-center py-10"><Spin /></div> : (
          <div key={formKey} className="space-y-6">
            <DetailCard>
              <SectionHeader icon={Activity} title="一般情况评估" />
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-x-12 gap-y-8 mt-4">
                <RadioGroup label="自理能力" options={['正常','依赖','部分']} defaultValue={getVal('selfCare', '正常')} disabled={!isEditing} onChange={v => setVal('selfCare', v)} />
                <RadioGroup label="睡眠" options={['良好','一般','差']} defaultValue={getVal('sleep', '一般')} disabled={!isEditing} onChange={v => setVal('sleep', v)} />
                <RadioGroup label="饮食" options={['良好','一般','差']} defaultValue={getVal('diet', '良好')} disabled={!isEditing} onChange={v => setVal('diet', v)} />
                <RadioGroup label="营养" options={['良好','一般','差']} defaultValue={getVal('nutrition', '良好')} disabled={!isEditing} onChange={v => setVal('nutrition', v)} />
                <RadioGroup label="尿量" options={['无尿','少尿']} defaultValue={getVal('urineOutput', '无尿')} disabled={!isEditing} onChange={v => setVal('urineOutput', v)} />
                <RadioGroup label="用药依从" options={['依从','不依从']} defaultValue={getVal('medication', '依从')} disabled={!isEditing} onChange={v => setVal('medication', v)} />
                <RadioGroup label="血压监测" options={['自测','定期','偶尔']} defaultValue={getVal('bpMonitoring', '定期')} disabled={!isEditing} onChange={v => setVal('bpMonitoring', v)} />
                <RadioGroup label="血糖监测" options={['自测','定期','偶尔']} defaultValue={getVal('bgMonitoring', '自测')} disabled={!isEditing} onChange={v => setVal('bgMonitoring', v)} />
              </div>
            </DetailCard>

            <DetailCard>
              <SectionHeader icon={Zap} title="血透执行情况" />
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-12 mt-4">
                <SmallInput label="平均血流量" suffix="ml/min" defaultValue={getVal('avgBloodFlow')} readOnly={!isEditing} onChange={v => setVal('avgBloodFlow', v)} />
                <SmallInput label="干体重" suffix="kg" defaultValue={getVal('dryWeight')} readOnly={!isEditing} onChange={v => setVal('dryWeight', v)} />
                <RadioGroup label="间期增重" options={['>5kg','<5kg']} defaultValue={getVal('interdialyticWeightGain', '<5kg')} disabled={!isEditing} onChange={v => setVal('interdialyticWeightGain', v)} />
                <RadioGroup label="治疗依从" options={['良好','一般','差']} defaultValue={getVal('treatmentCompliance', '良好')} disabled={!isEditing} onChange={v => setVal('treatmentCompliance', v)} />
                <RadioGroup label="间期血压" options={['高','正常','低']} defaultValue={getVal('interdialyticBP', '正常')} disabled={!isEditing} onChange={v => setVal('interdialyticBP', v)} />
                <RadioGroup label="透中血压" options={['高','正常','低']} defaultValue={getVal('intradialyticBP', '正常')} disabled={!isEditing} onChange={v => setVal('intradialyticBP', v)} />
                <RadioGroup label="透后血压" options={['高','正常','低']} defaultValue={getVal('postDialysisBP', '正常')} disabled={!isEditing} onChange={v => setVal('postDialysisBP', v)} />
                <SmallInput label="透中并发症" defaultValue={getVal('intradialyticComplications')} readOnly={!isEditing} onChange={v => setVal('intradialyticComplications', v)} />
              </div>
            </DetailCard>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <DetailCard>
                <SectionHeader icon={ShieldCheck} title="充分性" />
                <div className="grid grid-cols-2 gap-6 mt-4">
                  <SmallInput label="CTR" suffix="%" defaultValue={getVal('ctr')} readOnly={!isEditing} onChange={v => setVal('ctr', v)} />
                  <SmallInput label="Hb" suffix="g/L" defaultValue={getVal('hb')} readOnly={!isEditing} onChange={v => setVal('hb', v)} />
                  <SmallInput label="URR" suffix="%" defaultValue={getVal('urr')} readOnly={!isEditing} onChange={v => setVal('urr', v)} />
                  <SmallInput label="Kt/V" suffix="%" defaultValue={getVal('ktv')} readOnly={!isEditing} onChange={v => setVal('ktv', v)} />
                </div>
              </DetailCard>
              <DetailCard>
                <SectionHeader icon={Beaker} title="骨病/钙磷" />
                <div className="grid grid-cols-1 gap-6 mt-4">
                  <SmallInput label="iPTH" suffix="pg/mL" defaultValue={getVal('iPTH')} readOnly={!isEditing} onChange={v => setVal('iPTH', v)} />
                  <div className="grid grid-cols-2 gap-4">
                    <SmallInput label="钙" suffix="mmol/L" defaultValue={getVal('calcium')} readOnly={!isEditing} onChange={v => setVal('calcium', v)} />
                    <SmallInput label="磷" suffix="mmol/L" defaultValue={getVal('phosphorus')} readOnly={!isEditing} onChange={v => setVal('phosphorus', v)} />
                  </div>
                </div>
              </DetailCard>
            </div>

            <DetailCard>
              <SectionHeader icon={Briefcase} title="就诊/转归" />
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 mt-4">
                <RadioGroup label="是否住院" options={['是','否']} defaultValue={getVal('hospitalized', '否')} disabled={!isEditing} onChange={v => setVal('hospitalized', v)} />
                <SmallInput label="入院日期" suffix={<Calendar size={12}/>} defaultValue={getVal('admissionDate')} readOnly={!isEditing} onChange={v => setVal('admissionDate', v)} />
                <SmallInput label="出院日期" suffix={<Calendar size={12}/>} defaultValue={getVal('dischargeDate')} readOnly={!isEditing} onChange={v => setVal('dischargeDate', v)} />
                <RadioGroup label="急诊透析" options={['是','否']} defaultValue={getVal('emergencyDialysis', '否')} disabled={!isEditing} onChange={v => setVal('emergencyDialysis', v)} />
              </div>
            </DetailCard>

            <DetailCard>
              <SectionHeader icon={Edit3} title="总体评价及建议" />
              <div className="mt-4">
                <textarea readOnly={!isEditing}
                  className={`w-full h-32 p-4 border rounded-2xl text-sm font-bold text-slate-800 outline-none transition-all resize-none ${isEditing ? 'bg-white border-slate-300' : 'bg-slate-50 border-slate-100 text-slate-700'}`}
                  placeholder="请输入总体评价及治疗建议..."
                  defaultValue={getVal('overallEvaluation')}
                  onChange={e => setVal('overallEvaluation', e.target.value)}
                />
              </div>
            </DetailCard>
          </div>
        )}
      </div>

      {isPreviewOpen && (
        <SummaryPrintView patient={patient} year={localSummaryYear} month={localSummaryMonth} data={content} onClose={() => setIsPreviewOpen(false)} />
      )}
    </div>
  )
}
