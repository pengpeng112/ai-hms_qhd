// SummaryPrintView - 月份小结打印预览组件

import { Printer, X } from 'lucide-react'
import type { Patient } from '@/types/original'

interface SummaryPrintViewProps {
  patient: Patient
  year: string
  month: string
  data: Record<string, unknown>
  onClose: () => void
}

export default function SummaryPrintView({ patient, year, month, onClose }: SummaryPrintViewProps) {
  return (
    <div className="fixed inset-0 z-[500] bg-slate-900/80 backdrop-blur-md overflow-y-auto flex justify-center py-10 print:p-0 print:bg-white print:static print:h-auto print:overflow-visible">
      <div className="bg-white shadow-2xl w-[210mm] min-h-[297mm] p-[15mm] relative print:shadow-none print:w-full print:h-auto print:p-0 print:mx-0 print:my-0 flex flex-col font-serif">
        {/* 控制按钮 */}
        <div className="absolute top-4 right-[-80px] flex flex-col gap-4 print:hidden">
          <button onClick={() => window.print()} className="p-4 bg-blue-600 text-white rounded-full shadow-2xl hover:scale-110 transition-transform">
            <Printer size={24} />
          </button>
          <button onClick={onClose} className="p-4 bg-white text-slate-600 rounded-full shadow-2xl hover:scale-110 transition-transform">
            <X size={24} />
          </button>
        </div>

        {/* 标题 */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold tracking-widest border-b-2 border-black pb-2 inline-block">血液净化月份小结</h1>
        </div>

        {/* 表格容器 */}
        <div className="border border-black flex-1 flex flex-col text-[13px] leading-relaxed">
          {/* 第一行: 基础信息 */}
          <div className="flex border-b border-black">
            <div className="flex-1 px-3 py-2 border-r border-black">
              姓名: <span className="font-bold ml-1">{patient.name}</span>
            </div>
            <div className="w-24 px-3 py-2 border-r border-black">
              性别: <span className="ml-1">{patient.gender}</span>
            </div>
            <div className="w-24 px-3 py-2 border-r border-black">
              年龄: <span className="ml-1">{patient.age}</span>
            </div>
            <div className="flex-1 px-3 py-2">
              治疗时间: <span className="font-bold ml-1">{year} 年 {month} 月</span>
            </div>
          </div>

          {/* 一般情况 */}
          <div className="flex border-b border-black">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">一般情况</div>
            <div className="flex-1 p-3 space-y-2">
              <div className="flex flex-wrap gap-x-6 gap-y-1">
                <span>生活自理: □高 ☑正常 □低</span>
                <span>睡眠: □好 ☑一般 □差</span>
                <span>饮食: ☑好 □一般 □差</span>
                <span>营养: ☑好 □一般 □差</span>
              </div>
              <div className="flex flex-wrap gap-x-6 gap-y-1">
                <span>尿量: □无尿 ☑少尿 ml/日</span>
                <span>服药: ☑遵医嘱 □不遵医嘱</span>
              </div>
              <div className="flex flex-wrap gap-x-6 gap-y-1">
                <span>血压: ☑自测(□规律 □偶) □未自测</span>
                <span>血糖: □自测(□规律 □偶) ☑未自测</span>
              </div>
            </div>
          </div>

          {/* 血透情况 */}
          <div className="flex border-b border-black h-12">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">血透情况</div>
            <div className="flex-1 flex items-center px-3">
              平均血流速: <span className="font-bold ml-1 mx-2">220ml/min</span>
            </div>
          </div>

          {/* 干体重与依从性 */}
          <div className="flex border-b border-black h-12">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">干体重</div>
            <div className="flex-1 flex items-center px-3 gap-10">
              <span>透析间期平均体重增加: (□&gt;5kg ☑&lt;5kg)</span>
              <span>治疗依从性: ☑一般 □差</span>
            </div>
          </div>

          {/* 血压评估 */}
          <div className="flex border-b border-black min-h-[60px]">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">血 压</div>
            <div className="flex-1 p-3 space-y-2">
              <div className="flex gap-8">
                <span>1.透析间期: □高 ☑正常 □低</span>
                <span>2.透中: □高 ☑正常 □低</span>
                <span>3.透后: □高 ☑正常 □低</span>
              </div>
            </div>
          </div>

          {/* 并发症与水肿 */}
          <div className="flex border-b border-black min-h-[100px]">
            <div className="flex-1 flex flex-col">
              <div className="flex border-b border-black flex-1">
                <div className="flex-1 p-3">
                  <span className="font-bold block mb-1">透析中并发症:</span>
                  <span className="opacity-80">□低血压 □肌肉痉挛 □低血糖 □心律失常 □心绞痛 □心肌梗塞 □肺栓塞 □透析器反应 □致热源反应 □失衡综合征 ☑无 □其他</span>
                </div>
              </div>
              <div className="flex flex-1">
                <div className="flex-1 p-3">
                  <div className="flex gap-10 items-center">
                    <span>透析间期水肿: (□有 ☑无)</span>
                    <span>水肿部位: (□颜面 □下肢 □心包 □胸腔 □腹腔)</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* CTR */}
          <div className="flex border-b border-black h-16">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">CTR</div>
            <div className="flex-1 flex border-r border-black items-center px-3">%</div>
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">备注</div>
            <div className="flex-[2] flex items-center px-3"></div>
          </div>

          {/* 透析充分性 */}
          <div className="flex border-b border-black h-16">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">透析充分性</div>
            <div className="flex-1 p-3">
              <div className="flex gap-10">
                <span>URR: %; Kt/V: %</span>
                <span>□充分 □不充分</span>
              </div>
            </div>
          </div>

          {/* 骨病 */}
          <div className="flex border-b border-black h-16">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">骨病</div>
            <div className="flex-1 p-3">
              <div className="flex flex-wrap gap-x-10 gap-y-1">
                <span>iPTH: pg/mL</span>
                <span>Ca: mmol/L</span>
                <span>P: mmol/L</span>
              </div>
            </div>
          </div>

          {/* 贫血 */}
          <div className="flex border-b border-black h-12">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">贫血</div>
            <div className="flex-1 flex items-center px-3">Hb: 104g/L</div>
          </div>

          {/* 其他 */}
          <div className="flex border-b border-black min-h-[80px]">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">其他</div>
            <div className="flex-1 p-3 space-y-3">
              <div className="flex flex-wrap gap-x-10 gap-y-2">
                <span>住院: □是 ☑否</span>
                <span>住院日期:</span>
                <span>出院日期:</span>
                <span>转归: □好转 □恶化 □转院 □死亡</span>
              </div>
              <div>主要就诊原因: </div>
            </div>
          </div>

          {/* 急诊透析 */}
          <div className="flex border-b border-black h-16">
            <div className="w-24 bg-gray-50 flex items-center justify-center font-bold border-r border-black text-center px-2">急诊透析</div>
            <div className="flex-1 p-3">
              <div className="flex gap-10">
                <span>□有 ☑无</span>
                <span>(原因: □高钾血症 □心力衰竭 □其他)</span>
              </div>
            </div>
          </div>

          {/* 总结建议 */}
          <div className="flex-1 flex flex-col">
            <div className="p-4 border-b border-black min-h-[200px]">
              <p className="font-bold mb-2 underline">本阶段透析总评价以及治疗建议:</p>
              <p className="leading-relaxed">患者近期一般状况可，继续当前治疗。</p>
            </div>

            <div className="mt-auto px-10 py-10 flex justify-end items-end gap-2">
              <span className="font-bold">医师签名: </span>
              <div className="border-b border-black w-32 text-center pb-1 text-lg italic">王医生</div>
            </div>
          </div>
        </div>

        {/* 系统标识 */}
        <div className="text-[10px] text-gray-400 mt-4 flex justify-between print:hidden">
          <span>AI-HMS 智能血液透析管理系统</span>
          <span>打印日期: {new Date().toISOString().slice(0, 10)}</span>
        </div>
      </div>
    </div>
  )
}
