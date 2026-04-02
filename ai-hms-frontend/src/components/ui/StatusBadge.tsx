/**
 * StatusBadge - 状态徽章组件
 * 用于展示各种状态（治疗状态、风险等级、床位状态等）
 */

import { getStatusBadgeStyles, getRiskLevelTextStyles, getPatientTypeStyles, getBedStatusStyles } from '@/utils/styles'

export type BadgeType = 'status' | 'risk' | 'patientType' | 'bed' | 'custom'

export interface StatusBadgeProps {
    /** 徽章类型 */
    type?: BadgeType
    /** 状态值 */
    status: string
    /** 自定义样式 class（当 type='custom' 时使用） */
    customStyles?: string
    /** 尺寸 */
    size?: 'sm' | 'md' | 'lg'
}

export default function StatusBadge({
    type = 'status',
    status,
    customStyles,
    size = 'sm'
}: StatusBadgeProps) {
    // 根据类型获取样式
    const getStyles = () => {
        if (type === 'custom' && customStyles) {
            return customStyles
        }
        switch (type) {
            case 'risk':
                return getRiskLevelTextStyles(status)
            case 'patientType':
                return getPatientTypeStyles(status)
            case 'bed':
                return getBedStatusStyles(status)
            case 'status':
            default:
                return getStatusBadgeStyles(status)
        }
    }

    // 尺寸样式
    const sizeStyles = {
        sm: 'px-2 py-0.5 text-[10px]',
        md: 'px-2.5 py-1 text-xs',
        lg: 'px-3 py-1.5 text-sm'
    }

    return (
        <span className={`inline-flex items-center rounded font-bold border ${getStyles()} ${sizeStyles[size]}`}>
            {status}
        </span>
    )
}
