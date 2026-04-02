/**
 * 样式工具函数
 * 统一管理状态、风险等级、严重程度等样式
 */

/**
 * 获取治疗状态徽章样式
 */
export function getStatusBadgeStyles(status: string): string {
    switch (status) {
        case '透析中':
            return 'bg-blue-100 text-blue-700 border-blue-200'
        case '候诊':
        case '待透析':
            return 'bg-yellow-100 text-yellow-700 border-yellow-200'
        case '已结束':
        case '已完成':
            return 'bg-green-100 text-green-700 border-green-200'
        case '居家':
            return 'bg-gray-100 text-gray-500 border-gray-200'
        default:
            return 'bg-gray-100 text-gray-600 border-gray-200'
    }
}

/**
 * 获取风险等级背景色
 */
export function getRiskLevelBgColor(level: string): string {
    switch (level) {
        case '高危':
        case 'high':
            return 'bg-red-500'
        case '中危':
        case 'medium':
            return 'bg-orange-500'
        case '低危':
        case 'low':
            return 'bg-green-500'
        default:
            return 'bg-gray-400'
    }
}

/**
 * 获取风险等级文字样式
 */
export function getRiskLevelTextStyles(level: string): string {
    switch (level) {
        case '高危':
        case 'high':
            return 'text-red-600 bg-red-50 border-red-100'
        case '中危':
        case 'medium':
            return 'text-orange-600 bg-orange-50 border-orange-100'
        case '低危':
        case 'low':
            return 'text-green-600 bg-green-50 border-green-100'
        default:
            return 'text-gray-600 bg-gray-50 border-gray-100'
    }
}

/**
 * 获取严重程度样式（用于任务/告警）
 */
export function getSeverityStyles(severity: string): string {
    switch (severity) {
        case 'high':
            return 'bg-red-50 border-red-200 text-red-700 hover:bg-red-100'
        case 'medium':
            return 'bg-orange-50 border-orange-200 text-orange-700 hover:bg-orange-100'
        case 'low':
            return 'bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100'
        default:
            return 'bg-gray-50 border-gray-200 text-gray-700 hover:bg-gray-100'
    }
}

/**
 * 获取患者类型样式
 */
export function getPatientTypeStyles(type: string): string {
    switch (type) {
        case '住院':
            return 'bg-blue-50 text-blue-600 border-blue-100'
        case '门诊':
            return 'bg-green-50 text-green-600 border-green-100'
        default:
            return 'bg-gray-50 text-gray-600 border-gray-100'
    }
}

/**
 * 获取感染状态样式
 */
export function getInfectionStatusStyles(status: string): string {
    return status === '阳性'
        ? 'bg-red-100 text-red-600 border border-red-100'
        : 'bg-green-100 text-green-600 border border-green-100'
}

/**
 * 获取床位状态样式
 */
export function getBedStatusStyles(status: string): string {
    switch (status) {
        case '使用中':
        case 'occupied':
            return 'bg-blue-100 text-blue-700 border-blue-200'
        case '空闲':
        case 'available':
            return 'bg-green-100 text-green-700 border-green-200'
        case '维护':
        case 'maintenance':
            return 'bg-orange-100 text-orange-700 border-orange-200'
        default:
            return 'bg-gray-100 text-gray-600 border-gray-200'
    }
}
