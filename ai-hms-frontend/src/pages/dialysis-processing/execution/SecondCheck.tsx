import type { Patient } from '../types'

export default function SecondCheck({ patient }: { patient: Patient }) {
  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="text-lg font-black text-slate-800">二次核对 / 双人核对</div>
      <div className="mt-3 text-sm text-slate-600">患者：{patient.name}</div>
      <div className="mt-4 rounded-xl bg-amber-50 border border-amber-200 p-4 text-sm text-amber-700">
        <p className="font-semibold">此页面为占位</p>
        <p className="mt-1">双人核对表单已整合到「核查验证」页面。请在左侧菜单切换到核查验证 Tab 进行操作。</p>
      </div>
    </div>
  )
}
