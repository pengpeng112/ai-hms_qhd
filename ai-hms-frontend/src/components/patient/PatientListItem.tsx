/**
 * PatientListItem - 患者列表项组件
 * 用于 Dashboard 和 PatientList 中的患者展示
 */

import { ArrowRight } from 'lucide-react'
import type { Patient } from '@/types/original'
import { getRiskLevelBgColor, getStatusBadgeStyles } from '@/utils/styles'

export interface PatientListItemProps {
    /** 患者数据 */
    patient: Patient
    /** 点击回调 */
    onClick?: (patient: Patient) => void
    /** 展示模式 */
    variant?: 'compact' | 'normal' | 'detailed'
}

/**
 * 紧凑模式 - 用于 Dashboard 侧边列表
 */
function CompactItem({ patient, onClick }: PatientListItemProps) {
    return (
        <div
            onClick={() => onClick?.(patient)}
            className="flex items-center p-3 rounded-xl border border-gray-100 hover:border-blue-200 hover:shadow-sm hover:bg-blue-50/30 cursor-pointer transition-all bg-white group"
        >
            <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold shadow-sm shrink-0 mr-3 ${getRiskLevelBgColor(patient.riskLevel)}`}>
                {patient.bedNumber || '-'}
            </div>
            <div className="flex-1 min-w-0">
                <div className="flex items-center mb-0.5">
                    <h4 className="font-bold text-gray-900 text-sm truncate mr-2">{patient.name}</h4>
                </div>
                <p className="text-xs text-gray-500 truncate">{patient.diagnosis}</p>
            </div>
            <div className="text-right shrink-0 ml-2">
                <span className={`inline-block px-2 py-0.5 rounded text-[10px] font-bold mb-0.5 ${
                    patient.status === '透析中' ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-500'
                }`}>
                    {patient.status}
                </span>
            </div>
        </div>
    )
}

/**
 * 标准模式 - 用于完整列表
 */
function NormalItem({ patient, onClick }: PatientListItemProps) {
    return (
        <div
            onClick={() => onClick?.(patient)}
            className="flex items-center p-4 rounded-xl border border-gray-100 hover:border-blue-200 hover:shadow-md hover:bg-blue-50/20 cursor-pointer transition-all bg-white group"
        >
            {/* 头像/床号 */}
            <div className={`w-12 h-12 rounded-xl flex items-center justify-center text-white font-bold shadow-sm shrink-0 mr-4 ${getRiskLevelBgColor(patient.riskLevel)}`}>
                {patient.bedNumber || '-'}
            </div>

            {/* 基本信息 */}
            <div className="flex-1 min-w-0">
                <div className="flex items-center mb-1">
                    <h4 className="font-bold text-gray-900 truncate mr-2">{patient.name}</h4>
                    <span className="text-xs text-gray-400">{patient.gender} · {patient.age}岁</span>
                </div>
                <p className="text-xs text-gray-500 truncate">{patient.diagnosis}</p>
            </div>

            {/* 状态 */}
            <div className="flex items-center gap-3 shrink-0 ml-4">
                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${getStatusBadgeStyles(patient.status)}`}>
                    {patient.status}
                </span>
                <button className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-full transition-colors opacity-0 group-hover:opacity-100">
                    <ArrowRight size={18} />
                </button>
            </div>
        </div>
    )
}

/**
 * 详细模式 - 用于表格行
 */
function DetailedItem({ patient, onClick }: PatientListItemProps) {
    return (
        <tr
            onClick={() => onClick?.(patient)}
            className="hover:bg-blue-50/30 cursor-pointer group transition-colors"
        >
            <td className="px-6 py-4">
                <div className="flex items-center">
                    <div className={`w-10 h-10 rounded-full flex items-center justify-center mr-3 border shrink-0 ${getRiskLevelBgColor(patient.riskLevel)} bg-opacity-20`}>
                        <span className="text-gray-600 font-bold text-xs">{patient.bedNumber?.slice(0, 3) || '-'}</span>
                    </div>
                    <div>
                        <div className="font-bold text-gray-800">{patient.name}</div>
                        <div className="text-xs text-gray-500">{patient.gender} · {patient.age}岁 · {patient.id}</div>
                    </div>
                </div>
            </td>
            <td className="px-6 py-4">
                <div className="space-y-1">
                    <span className={`inline-block px-2 py-0.5 rounded text-[10px] border ${
                        patient.patientType === '住院' ? 'bg-blue-50 text-blue-600 border-blue-100' : 'bg-green-50 text-green-600 border-green-100'
                    }`}>
                        {patient.patientType}
                    </span>
                    <div className="text-xs text-gray-600">{patient.insuranceType}</div>
                </div>
            </td>
            <td className="px-6 py-4">
                <div className="flex items-center space-x-3">
                    <div>
                        <div className="font-bold text-gray-800">{patient.defaultMode}</div>
                        <div className="text-xs text-gray-400">默认模式</div>
                    </div>
                    <div className="h-6 w-px bg-gray-200"></div>
                    <div>
                        <div className="font-bold text-gray-800">{patient.dryWeight.toFixed(1)} <span className="text-[10px] font-normal text-gray-400">kg</span></div>
                        <div className="text-xs text-gray-400">干体重</div>
                    </div>
                </div>
            </td>
            <td className="px-6 py-4">
                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${getStatusBadgeStyles(patient.status)}`}>
                    {patient.status}
                </span>
            </td>
            <td className="px-6 py-4">
                <span className="text-gray-700 font-medium">{patient.doctorName}</span>
            </td>
            <td className="px-6 py-4 text-right">
                <button className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-full transition-colors">
                    <ArrowRight size={18} />
                </button>
            </td>
        </tr>
    )
}

export default function PatientListItem({
    patient,
    onClick,
    variant = 'normal'
}: PatientListItemProps) {
    switch (variant) {
        case 'compact':
            return <CompactItem patient={patient} onClick={onClick} />
        case 'detailed':
            return <DetailedItem patient={patient} onClick={onClick} />
        case 'normal':
        default:
            return <NormalItem patient={patient} onClick={onClick} />
    }
}
