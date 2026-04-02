// Monthly Summary Tab - 月份小结

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { FileText, Activity, Zap, ShieldCheck, Beaker, Briefcase, Edit3, Eye, Save, CheckCircle2, Calendar } from 'lucide-react'
import { SectionHeader, DetailCard, RadioGroup, SmallInput } from '@/components/ui'
import { SummaryPrintView } from '@/components/patient/modals'
import type { Patient } from '@/types/original'

interface MonthlySummaryTabProps {
  patient: Patient
}

export default function MonthlySummaryTab({ patient }: MonthlySummaryTabProps) {
  const { t } = useTranslation('patient')

  // 状态管理
  const [localSummaryYear, setLocalSummaryYear] = useState('2024')
  const [localSummaryMonth, setLocalSummaryMonth] = useState('12')
  const [isMonthlySummaryEditing, setIsMonthlySummaryEditing] = useState(false)
  const [isMonthlySummaryPreviewOpen, setIsMonthlySummaryPreviewOpen] = useState(false)

  return (
    <div className="flex gap-6 animate-fade-in pb-10 h-full">
      {/* 左侧月份选择 */}
      <div className="w-24 shrink-0 flex flex-col gap-4 sticky top-0 h-fit">
        {['2025', '2024'].map(year => (
          <div key={year} className="space-y-1">
            <div className="text-[10px] font-black text-slate-300 uppercase px-2 mb-1 flex items-center justify-between">
              {year}{t('monthlySummary.yearSuffix')} <div className="w-1.5 h-1.5 bg-slate-200 rounded-full"></div>
            </div>
            <div className="flex flex-col gap-1">
              {['12','11','10','09','08','07','06','05','04','03','02','01'].map(month => (
                <button
                  key={month}
                  onClick={() => { setLocalSummaryYear(year); setLocalSummaryMonth(month); }}
                  className={`w-full py-2.5 px-3 rounded-xl text-xs font-black transition-all flex items-center justify-between group ${
                    localSummaryYear === year && localSummaryMonth === month
                      ? 'bg-blue-600 text-white shadow-lg'
                      : 'bg-white text-slate-500 hover:bg-slate-100 border border-slate-100'
                  }`}
                >
                  {month}{t('monthlySummary.monthSuffix')}
                  {parseInt(month) < 6 && (
                    <CheckCircle2 size={12} className={localSummaryYear === year && localSummaryMonth === month ? 'text-white' : 'text-slate-200 group-hover:text-slate-400'} />
                  )}
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* 右侧内容区 */}
      <div className="flex-1 space-y-6">
        {/* 顶部操作栏 */}
        <div className="flex justify-between items-center px-2">
          <div className="flex items-center gap-4">
            <h3 className="text-lg font-black text-slate-800 flex items-center">
              <FileText size={22} className="mr-3 text-blue-600" /> {t('monthlySummary.title')}
              <span className="ml-4 text-blue-600 font-mono">{localSummaryYear}-{localSummaryMonth}</span>
            </h3>
            <div className="flex items-center gap-2 px-3 py-1 bg-green-50 text-green-600 text-[10px] font-black rounded-full border border-green-100">
              <CheckCircle2 size={12}/> {t('status.done')}
            </div>
          </div>
          <div className="flex gap-2">
            <button onClick={() => setIsMonthlySummaryPreviewOpen(true)} className="px-6 py-2.5 bg-white border border-slate-200 text-slate-700 rounded-2xl text-sm font-black hover:bg-slate-50 transition-all flex items-center gap-2 shadow-sm">
              <Eye size={18} className="text-blue-500"/> {t('monthlySummary.preview')}
            </button>
            {isMonthlySummaryEditing ? (
              <button onClick={() => setIsMonthlySummaryEditing(false)} className="px-10 py-2.5 bg-blue-600 text-white rounded-2xl text-sm font-black shadow-xl shadow-blue-100 hover:bg-blue-700 transition-all flex items-center gap-2">
                <Save size={18}/> {t('monthlySummary.save')}
              </button>
            ) : (
              <button onClick={() => setIsMonthlySummaryEditing(true)} className="px-10 py-2.5 bg-slate-900 text-white rounded-2xl text-sm font-black shadow-xl hover:bg-slate-800 transition-all flex items-center gap-2">
                <Edit3 size={18}/> {t('action.edit')}
              </button>
            )}
          </div>
        </div>

        <div className="space-y-6">
          {/* 一般情况评估 */}
          <DetailCard>
            <SectionHeader icon={Activity} title={t('monthlySummary.generalAssessment')} />
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-x-12 gap-y-8 mt-4">
              <RadioGroup label={t('monthlySummary.selfCare')} options={[t('monthlySummary.normal'), t('monthlySummary.dependent'), t('monthlySummary.partial'), t('monthlySummary.complete')]} defaultValue={t('monthlySummary.normal')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.sleep')} options={[t('monthlySummary.good'), t('monthlySummary.average'), t('monthlySummary.poor')]} defaultValue={t('monthlySummary.average')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.diet')} options={[t('monthlySummary.good'), t('monthlySummary.average'), t('monthlySummary.poor')]} defaultValue={t('monthlySummary.good')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.nutrition')} options={[t('monthlySummary.good'), t('monthlySummary.average'), t('monthlySummary.poor')]} defaultValue={t('monthlySummary.good')} disabled={!isMonthlySummaryEditing} />
              <div className="flex items-end gap-2">
                <div className="flex-1">
                  <RadioGroup label={t('monthlySummary.urineOutput')} options={[t('monthlySummary.anuria'), t('monthlySummary.oliguria')]} defaultValue={t('monthlySummary.anuria')} disabled={!isMonthlySummaryEditing} />
                </div>
                <div className="w-24">
                  <SmallInput label={t('monthlySummary.urineAmount')} suffix="ml/d" readOnly={!isMonthlySummaryEditing} />
                </div>
              </div>
              <RadioGroup label={t('monthlySummary.medication')} options={[t('monthlySummary.compliant'), t('monthlySummary.nonCompliant')]} defaultValue={t('monthlySummary.compliant')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.bpMonitoring')} options={[t('monthlySummary.selfMonitor'), t('monthlySummary.regular'), t('monthlySummary.occasional'), t('monthlySummary.notSelfMonitor')]} defaultValue={t('monthlySummary.regular')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.bgMonitoring')} options={[t('monthlySummary.selfMonitor'), t('monthlySummary.regular'), t('monthlySummary.occasional'), t('monthlySummary.notSelfMonitor')]} defaultValue={t('monthlySummary.selfMonitor')} disabled={!isMonthlySummaryEditing} />
            </div>
          </DetailCard>

          {/* 血透执行情况 */}
          <DetailCard>
            <SectionHeader icon={Zap} title={t('monthlySummary.dialysisExecution')} />
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-12 mt-4">
              <div className="lg:col-span-1">
                <SmallInput label={t('monthlySummary.avgBloodFlow')} suffix="ml/min" defaultValue="230" readOnly={!isMonthlySummaryEditing} />
              </div>
              <div className="lg:col-span-1">
                <SmallInput label={t('info.dryWeight')} suffix="kg" defaultValue="67.5" readOnly={!isMonthlySummaryEditing} />
              </div>
              <RadioGroup label={t('monthlySummary.interdialyticWeightGain')} options={['>5kg', '<5kg']} defaultValue="<5kg" disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.treatmentCompliance')} options={[t('monthlySummary.good'), t('monthlySummary.average'), t('monthlySummary.poor')]} defaultValue={t('monthlySummary.good')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.interdialyticBP')} options={[t('monthlySummary.high'), t('monthlySummary.normalBP'), t('monthlySummary.low')]} defaultValue={t('monthlySummary.normalBP')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.intradialyticBP')} options={[t('monthlySummary.high'), t('monthlySummary.normalBP'), t('monthlySummary.low')]} defaultValue={t('monthlySummary.normalBP')} disabled={!isMonthlySummaryEditing} />
              <RadioGroup label={t('monthlySummary.postDialysisBP')} options={[t('monthlySummary.high'), t('monthlySummary.normalBP'), t('monthlySummary.low')]} defaultValue={t('monthlySummary.normalBP')} disabled={!isMonthlySummaryEditing} />
              <SmallInput label={t('monthlySummary.intradialyticComplications')} placeholder={t('monthlySummary.noneIfEmpty')} readOnly={!isMonthlySummaryEditing} />
              <div className="flex items-end gap-3 lg:col-span-1">
                <div className="shrink-0">
                  <RadioGroup label={t('monthlySummary.interdialyticEdema')} options={[t('label.yes'), t('label.no')]} defaultValue={t('label.no')} disabled={!isMonthlySummaryEditing} />
                </div>
                <div className="flex-1">
                  <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">{t('monthlySummary.edemaLocation')}</label>
                  <select disabled={!isMonthlySummaryEditing} className="w-full h-9 px-2 border border-slate-200 rounded-lg text-xs font-bold outline-none bg-slate-50">
                    <option>{t('monthlySummary.none')}</option>
                    <option>{t('monthlySummary.lowerLimbs')}</option>
                    <option>{t('monthlySummary.wholeBody')}</option>
                  </select>
                </div>
              </div>
            </div>
          </DetailCard>

          {/* 充分性与骨病指标 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <DetailCard>
              <SectionHeader icon={ShieldCheck} title={t('monthlySummary.adequacyAnemia')} />
              <div className="grid grid-cols-2 gap-6 mt-4">
                <SmallInput label="CTR" suffix="%" readOnly={!isMonthlySummaryEditing} />
                <SmallInput label="Hb" suffix="g/L" readOnly={!isMonthlySummaryEditing} />
                <SmallInput label="URR" suffix="%" readOnly={!isMonthlySummaryEditing} />
                <SmallInput label="Kt/V" suffix="%" readOnly={!isMonthlySummaryEditing} />
                <div className="col-span-2">
                  <RadioGroup label={t('monthlySummary.dialysisAdequacy')} options={[t('monthlySummary.adequate'), t('monthlySummary.inadequate')]} defaultValue={t('monthlySummary.adequate')} disabled={!isMonthlySummaryEditing} />
                </div>
              </div>
            </DetailCard>
            <DetailCard>
              <SectionHeader icon={Beaker} title={t('monthlySummary.boneDisease')} />
              <div className="grid grid-cols-1 gap-6 mt-4">
                <SmallInput label="iPTH" suffix="pg/mL" defaultValue="235.4" readOnly={!isMonthlySummaryEditing} />
                <div className="grid grid-cols-2 gap-4">
                  <SmallInput label={t('monthlySummary.calcium')} suffix="mmol/L" defaultValue="2.25" readOnly={!isMonthlySummaryEditing} />
                  <SmallInput label={t('monthlySummary.phosphorus')} suffix="mmol/L" defaultValue="1.42" readOnly={!isMonthlySummaryEditing} />
                </div>
                <SmallInput label={t('label.notes')} placeholder={t('monthlySummary.specialNotesPlaceholder')} readOnly={!isMonthlySummaryEditing} />
              </div>
            </DetailCard>
          </div>

          {/* 就诊转归与其他 */}
          <DetailCard>
            <SectionHeader icon={Briefcase} title={t('monthlySummary.visitOutcome')} />
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 mt-4">
              <RadioGroup label={t('monthlySummary.hospitalized')} options={[t('label.yes'), t('label.no')]} defaultValue={t('label.no')} disabled={!isMonthlySummaryEditing} />
              <SmallInput label={t('monthlySummary.admissionDate')} suffix={<Calendar size={12}/>} readOnly={!isMonthlySummaryEditing} />
              <SmallInput label={t('monthlySummary.dischargeDate')} suffix={<Calendar size={12}/>} readOnly={!isMonthlySummaryEditing} />
              <div className="flex flex-col gap-2">
                <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">{t('monthlySummary.outcomeStatus')}</label>
                <select disabled={!isMonthlySummaryEditing} className="h-9 px-3 border border-slate-200 rounded-lg text-xs font-bold outline-none bg-white">
                  <option>{t('monthlySummary.maintainDialysis')}</option>
                  <option>{t('monthlySummary.transfer')}</option>
                  <option>{t('monthlySummary.kidneyTransplant')}</option>
                  <option>{t('monthlySummary.death')}</option>
                </select>
              </div>
              <div className="lg:col-span-2">
                <SmallInput label={t('monthlySummary.mainVisitReason')} readOnly={!isMonthlySummaryEditing} />
              </div>
              <div className="flex items-end gap-3 lg:col-span-2">
                <div className="shrink-0">
                  <RadioGroup label={t('monthlySummary.emergencyDialysis')} options={[t('label.yes'), t('label.no')]} defaultValue={t('label.no')} disabled={!isMonthlySummaryEditing} />
                </div>
                <div className="flex-1">
                  <RadioGroup label={t('monthlySummary.emergencyReason')} options={[t('monthlySummary.hyperkalemia'), t('monthlySummary.heartFailure'), t('monthlySummary.other')]} disabled={!isMonthlySummaryEditing} />
                </div>
              </div>
            </div>
          </DetailCard>

          {/* 本阶段透析总评价以及治疗建议 */}
          <DetailCard>
            <SectionHeader icon={Edit3} title={t('monthlySummary.overallEvaluation')} />
            <div className="mt-4">
              <textarea
                readOnly={!isMonthlySummaryEditing}
                className={`w-full h-32 p-4 border rounded-2xl text-sm font-bold text-slate-800 outline-none transition-all resize-none leading-relaxed ${
                  isMonthlySummaryEditing
                    ? 'bg-white border-slate-300 focus:ring-1 focus:ring-blue-500'
                    : 'bg-slate-50 border-slate-100 text-slate-700'
                }`}
                placeholder={t('monthlySummary.evaluationPlaceholder')}
                defaultValue={t('monthlySummary.evaluationDefaultValue')}
              />
            </div>
          </DetailCard>
        </div>
      </div>

      {/* 打印预览 Modal */}
      {isMonthlySummaryPreviewOpen && (
        <SummaryPrintView
          patient={patient}
          year={localSummaryYear}
          month={localSummaryMonth}
          data={{}}
          onClose={() => setIsMonthlySummaryPreviewOpen(false)}
        />
      )}
    </div>
  )
}
