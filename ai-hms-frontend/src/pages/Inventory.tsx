import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
    Package,
    Printer,
    Search,
    Plus,
    Tag,
    ArrowDownCircle,
    ArrowUpCircle,
    AlertCircle,
    CheckCircle,
    Clock,
    Download
} from 'lucide-react'
import { EmptyState } from '@/components'

// 耗材类型定义
interface InventoryItem {
    id: string
    name: string
    spec: string
    category: string
    stock: number
    unit: string
    minStock: number
    maxStock: number
    alert: boolean
    location: string
    supplier: string
    lastUpdated: string
}

// 出入库记录类型
interface StockLog {
    id: string
    date: string
    type: 'in' | 'out'
    itemName: string
    quantity: number
    unit: string
    operator: string
    note: string
}

// 标签打印任务类型
interface LabelTask {
    id: string
    itemName: string
    spec: string
    quantity: number
    status: 'pending' | 'printing' | 'completed'
    createdAt: string
}

type TabType = 'STOCK' | 'LOG' | 'LABELS'

// Mock 数据
const MOCK_INVENTORY: InventoryItem[] = [
    {
        id: 'INV001',
        name: 'FX60 透析器',
        spec: '高通量',
        category: '透析耗材',
        stock: 124,
        unit: '个',
        minStock: 50,
        maxStock: 200,
        alert: false,
        location: '耗材库 A-01',
        supplier: '费森尤斯',
        lastUpdated: '2026-01-07 10:30'
    },
    {
        id: 'INV002',
        name: '血液管路',
        spec: '通用',
        category: '透析耗材',
        stock: 85,
        unit: '套',
        minStock: 100,
        maxStock: 300,
        alert: true,
        location: '耗材库 A-02',
        supplier: '费森尤斯',
        lastUpdated: '2026-01-07 09:15'
    },
    {
        id: 'INV003',
        name: '16G 穿刺针',
        spec: '锐针',
        category: '透析耗材',
        stock: 320,
        unit: '支',
        minStock: 200,
        maxStock: 500,
        alert: false,
        location: '耗材库 A-03',
        supplier: '日本尼普洛',
        lastUpdated: '2026-01-07 08:45'
    },
    {
        id: 'INV004',
        name: '碳酸氢盐透析液',
        spec: '5L',
        category: '药品',
        stock: 45,
        unit: '袋',
        minStock: 30,
        maxStock: 100,
        alert: false,
        location: '药品库 B-01',
        supplier: '百特',
        lastUpdated: '2026-01-07 07:20'
    },
    {
        id: 'INV005',
        name: '肝素钠',
        spec: '5000IU/ml',
        category: '药品',
        stock: 12,
        unit: '支',
        minStock: 20,
        maxStock: 80,
        alert: true,
        location: '药品库 B-02',
        supplier: '上海医药',
        lastUpdated: '2026-01-06 18:30'
    },
    {
        id: 'INV006',
        name: '一次性注射器',
        spec: '10ml',
        category: '医疗器械',
        stock: 580,
        unit: '个',
        minStock: 200,
        maxStock: 1000,
        alert: false,
        location: '耗材库 A-05',
        supplier: '江西洪达',
        lastUpdated: '2026-01-07 11:00'
    }
]

const MOCK_LOGS: StockLog[] = [
    {
        id: 'LOG001',
        date: '2026-01-07 10:30',
        type: 'in',
        itemName: 'FX60 透析器',
        quantity: 50,
        unit: '个',
        operator: '李护士',
        note: '月度采购'
    },
    {
        id: 'LOG002',
        date: '2026-01-07 09:15',
        type: 'out',
        itemName: '血液管路',
        quantity: 15,
        unit: '套',
        operator: '王护士',
        note: '治疗使用'
    },
    {
        id: 'LOG003',
        date: '2026-01-07 08:45',
        type: 'in',
        itemName: '16G 穿刺针',
        quantity: 100,
        unit: '支',
        operator: '张主任',
        note: '紧急采购'
    },
    {
        id: 'LOG004',
        date: '2026-01-06 18:30',
        type: 'out',
        itemName: '肝素钠',
        quantity: 8,
        unit: '支',
        operator: '赵护士',
        note: '夜班使用'
    },
    {
        id: 'LOG005',
        date: '2026-01-06 16:20',
        type: 'in',
        itemName: '碳酸氢盐透析液',
        quantity: 30,
        unit: '袋',
        operator: '李护士',
        note: '常规补货'
    }
]

const MOCK_LABEL_TASKS: LabelTask[] = [
    {
        id: 'LBL001',
        itemName: 'FX60 透析器',
        spec: '高通量',
        quantity: 50,
        status: 'completed',
        createdAt: '2026-01-07 10:35'
    },
    {
        id: 'LBL002',
        itemName: '16G 穿刺针',
        spec: '锐针',
        quantity: 100,
        status: 'printing',
        createdAt: '2026-01-07 09:00'
    },
    {
        id: 'LBL003',
        itemName: '碳酸氢盐透析液',
        spec: '5L',
        quantity: 30,
        status: 'pending',
        createdAt: '2026-01-06 16:25'
    }
]

export default function Inventory() {
    const { t } = useTranslation('inventory')
    const [activeTab, setActiveTab] = useState<TabType>('STOCK')
    const [searchTerm, setSearchTerm] = useState('')
    const [selectedItems, setSelectedItems] = useState<string[]>([])

    // 过滤库存数据
    const filteredInventory = MOCK_INVENTORY.filter(item =>
        item.name.includes(searchTerm) ||
        item.spec.includes(searchTerm) ||
        item.category.includes(searchTerm)
    )

    // 过滤出入库记录
    const filteredLogs = MOCK_LOGS.filter(log =>
        log.itemName.includes(searchTerm) ||
        log.operator.includes(searchTerm)
    )

    // 处理复选框选择
    const handleSelectItem = (id: string) => {
        setSelectedItems(prev =>
            prev.includes(id)
                ? prev.filter(i => i !== id)
                : [...prev, id]
        )
    }

    // 处理全选
    const handleSelectAll = () => {
        if (selectedItems.length === filteredInventory.length) {
            setSelectedItems([])
        } else {
            setSelectedItems(filteredInventory.map(item => item.id))
        }
    }

    // 处理入库登记
    const handleStockIn = () => {
        alert(t('alert.stockIn'))
    }

    // 处理批量打印标签
    const handleBatchPrint = () => {
        if (selectedItems.length === 0) {
            alert(t('alert.selectItems'))
            return
        }
        alert(t('alert.batchPrint', { count: selectedItems.length }))
    }

    // 获取库存状态样式
    const getStockStatusStyle = (item: InventoryItem) => {
        if (item.alert) {
            return 'px-2 py-0.5 bg-red-100 text-red-600 rounded-full text-[10px] font-bold'
        }
        if (item.stock < item.minStock * 1.2) {
            return 'px-2 py-0.5 bg-orange-100 text-orange-600 rounded-full text-[10px] font-bold'
        }
        return 'px-2 py-0.5 bg-green-100 text-green-600 rounded-full text-[10px] font-bold'
    }

    // 获取库存状态文本
    const getStockStatusText = (item: InventoryItem) => {
        if (item.alert) return t('stockAlert')
        if (item.stock < item.minStock * 1.2) return t('stockLow')
        return t('stockNormal')
    }

    // 渲染标签任务状态
    const getLabelStatusStyle = (status: LabelTask['status']) => {
        switch (status) {
            case 'completed':
                return 'bg-green-50 text-green-600 border-green-200'
            case 'printing':
                return 'bg-blue-50 text-blue-600 border-blue-200'
            case 'pending':
                return 'bg-orange-50 text-orange-600 border-orange-200'
        }
    }

    const getLabelStatusText = (status: LabelTask['status']) => {
        switch (status) {
            case 'completed':
                return t('label.completed')
            case 'printing':
                return t('label.printing')
            case 'pending':
                return t('label.pending')
        }
    }

    const getLabelStatusIcon = (status: LabelTask['status']) => {
        switch (status) {
            case 'completed':
                return <CheckCircle size={14} />
            case 'printing':
                return <Printer size={14} />
            case 'pending':
                return <Clock size={14} />
        }
    }

    return (
        <div className="h-full flex flex-col max-w-7xl mx-auto">
            {/* Header */}
            <div className="mb-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
                <h2 className="text-2xl font-bold text-gray-800 flex items-center">
                    <Package size={28} className="mr-3 text-indigo-600" />
                    {t('title')}
                </h2>
                <div className="flex gap-2">
                    <button
                        onClick={handleStockIn}
                        className="flex items-center px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-bold shadow-sm hover:bg-indigo-700 transition-all"
                    >
                        <Plus size={16} className="mr-2" />
                        {t('stockIn')}
                    </button>
                    <button
                        onClick={handleBatchPrint}
                        className="flex items-center px-4 py-2 bg-white border border-gray-200 text-gray-600 rounded-lg text-sm font-bold shadow-sm hover:bg-gray-50 transition-all"
                    >
                        <Tag size={16} className="mr-2" />
                        {t('batchPrint')}
                        {selectedItems.length > 0 && (
                            <span className="ml-2 px-1.5 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs">
                                {selectedItems.length}
                            </span>
                        )}
                    </button>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex gap-1 bg-gray-200 p-1 rounded-xl w-fit mb-6">
                {[
                    { key: 'STOCK', label: t('tab.stock') },
                    { key: 'LOG', label: t('tab.log') },
                    { key: 'LABELS', label: t('tab.labels') }
                ].map(tab => (
                    <button
                        key={tab.key}
                        onClick={() => setActiveTab(tab.key as TabType)}
                        className={`px-6 py-2 rounded-lg text-sm font-bold transition-all ${
                            activeTab === tab.key
                                ? 'bg-white text-indigo-600 shadow-sm'
                                : 'text-gray-500 hover:text-gray-700'
                        }`}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Main Content */}
            <div className="flex-1 bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden flex flex-col">
                {/* Search Bar */}
                <div className="p-4 border-b border-gray-100 flex justify-between items-center">
                    <div className="relative w-72">
                        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input
                            type="text"
                            placeholder={
                                activeTab === 'STOCK'
                                    ? t('search.stock')
                                    : activeTab === 'LOG'
                                    ? t('search.log')
                                    : t('search.labels')
                            }
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="w-full pl-9 pr-4 py-2 bg-gray-50 border-none rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 outline-none"
                        />
                    </div>
                    {activeTab === 'STOCK' && filteredInventory.length > 0 && (
                        <div className="text-sm text-gray-500">
                            {t('alertCount')}
                            <span className="ml-1 font-bold text-red-600">
                                {filteredInventory.filter(item => item.alert).length}
                            </span>
                            {' '}{t('items')}
                        </div>
                    )}
                </div>

                {/* Content Area */}
                <div className="flex-1 overflow-auto">
                    {/* 实时库存 Tab */}
                    {activeTab === 'STOCK' && (
                                <table className="w-full text-left">
                                    <thead className="bg-slate-50 text-gray-500 font-bold border-b border-gray-100 sticky top-0">
                                        <tr>
                                            <th className="px-6 py-4">
                                                <input
                                                    type="checkbox"
                                                    checked={
                                                        selectedItems.length === filteredInventory.length &&
                                                        filteredInventory.length > 0
                                                    }
                                                    onChange={handleSelectAll}
                                                    className="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                                />
                                            </th>
                                            <th className="px-6 py-4">{t('table.itemName')}</th>
                                            <th className="px-6 py-4">{t('table.spec')}</th>
                                            <th className="px-6 py-4">{t('table.category')}</th>
                                            <th className="px-6 py-4">{t('table.currentStock')}</th>
                                            <th className="px-6 py-4">{t('table.status')}</th>
                                            <th className="px-6 py-4">{t('table.location')}</th>
                                            <th className="px-6 py-4 text-right">{t('table.action')}</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-100">
                                        {filteredInventory.map(item => (
                                            <tr key={item.id} className="hover:bg-slate-50 transition-colors">
                                                <td className="px-6 py-4">
                                                    <input
                                                        type="checkbox"
                                                        checked={selectedItems.includes(item.id)}
                                                        onChange={() => handleSelectItem(item.id)}
                                                        className="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                                    />
                                                </td>
                                                <td className="px-6 py-4 font-bold text-gray-800">{item.name}</td>
                                                <td className="px-6 py-4 text-gray-500 text-sm">{item.spec}</td>
                                                <td className="px-6 py-4 text-gray-500 text-sm">{item.category}</td>
                                                <td className="px-6 py-4">
                                                    <span className="font-mono text-lg font-bold mr-1">{item.stock}</span>
                                                    <span className="text-xs text-gray-400">{item.unit}</span>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <span className={getStockStatusStyle(item)}>
                                                        {getStockStatusText(item)}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 text-gray-500 text-sm">{item.location}</td>
                                                <td className="px-6 py-4 text-right">
                                                    <button
                                                        onClick={() => alert(t('alert.viewDetail', { name: item.name }))}
                                                        className="text-indigo-600 hover:text-indigo-800 text-xs font-bold px-3 py-1 hover:bg-indigo-50 rounded transition-colors"
                                                    >
                                                        {t('table.detail')}
                                                    </button>
                                                </td>
                                            </tr>
                                        ))}
                                        {filteredInventory.length === 0 && (
                                            <tr>
                                                <td colSpan={8}>
                                                    <EmptyState
                                                        icon={Package}
                                                        message={t('empty.stock')}
                                                    />
                                                </td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            )}

                            {/* 出入库流水 Tab */}
                            {activeTab === 'LOG' && (
                                <table className="w-full text-left">
                                    <thead className="bg-slate-50 text-gray-500 font-bold border-b border-gray-100 sticky top-0">
                                        <tr>
                                            <th className="px-6 py-4">{t('table.time')}</th>
                                            <th className="px-6 py-4">{t('table.type')}</th>
                                            <th className="px-6 py-4">{t('table.itemName')}</th>
                                            <th className="px-6 py-4">{t('table.quantity')}</th>
                                            <th className="px-6 py-4">{t('table.operator')}</th>
                                            <th className="px-6 py-4">{t('table.note')}</th>
                                            <th className="px-6 py-4 text-right">{t('table.action')}</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-100">
                                        {filteredLogs.map(log => (
                                            <tr key={log.id} className="hover:bg-slate-50 transition-colors">
                                                <td className="px-6 py-4 text-gray-500 text-sm">{log.date}</td>
                                                <td className="px-6 py-4">
                                                    <span
                                                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-bold ${
                                                            log.type === 'in'
                                                                ? 'bg-green-100 text-green-600'
                                                                : 'bg-blue-100 text-blue-600'
                                                        }`}
                                                    >
                                                        {log.type === 'in' ? (
                                                            <>
                                                                <ArrowDownCircle size={12} className="mr-1" />
                                                                {t('log.in')}
                                                            </>
                                                        ) : (
                                                            <>
                                                                <ArrowUpCircle size={12} className="mr-1" />
                                                                {t('log.out')}
                                                            </>
                                                        )}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 font-bold text-gray-800">{log.itemName}</td>
                                                <td className="px-6 py-4">
                                                    <span className="font-mono text-base font-bold mr-1">
                                                        {log.type === 'in' ? '+' : '-'}
                                                        {log.quantity}
                                                    </span>
                                                    <span className="text-xs text-gray-400">{log.unit}</span>
                                                </td>
                                                <td className="px-6 py-4 text-gray-600">{log.operator}</td>
                                                <td className="px-6 py-4 text-gray-500 text-sm">{log.note}</td>
                                                <td className="px-6 py-4 text-right">
                                                    <button
                                                        onClick={() => alert(t('alert.viewLogDetail', { id: log.id }))}
                                                        className="text-indigo-600 hover:text-indigo-800 text-xs font-bold px-3 py-1 hover:bg-indigo-50 rounded transition-colors"
                                                    >
                                                        {t('table.detail')}
                                                    </button>
                                                </td>
                                            </tr>
                                        ))}
                                        {filteredLogs.length === 0 && (
                                            <tr>
                                                <td colSpan={7}>
                                                    <EmptyState
                                                        icon={ArrowDownCircle}
                                                        message={t('empty.log')}
                                                    />
                                                </td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            )}

                            {/* 标签打印 Tab */}
                            {activeTab === 'LABELS' && (
                                <div className="p-6">
                                    <div className="space-y-4">
                                        {MOCK_LABEL_TASKS.map(task => (
                                            <div
                                                key={task.id}
                                                className="p-4 border border-gray-200 rounded-lg hover:shadow-md transition-shadow"
                                            >
                                                <div className="flex items-center justify-between">
                                                    <div className="flex-1">
                                                        <div className="flex items-center space-x-3">
                                                            <h3 className="text-base font-bold text-gray-800">
                                                                {task.itemName}
                                                            </h3>
                                                            <span
                                                                className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-bold border ${getLabelStatusStyle(
                                                                    task.status
                                                                )}`}
                                                            >
                                                                {getLabelStatusIcon(task.status)}
                                                                <span className="ml-1">
                                                                    {getLabelStatusText(task.status)}
                                                                </span>
                                                            </span>
                                                        </div>
                                                        <div className="mt-2 flex items-center space-x-4 text-sm text-gray-500">
                                                            <span>{t('label.spec')}{task.spec}</span>
                                                            <span>{t('label.quantity')}{task.quantity} {t('label.sheets')}</span>
                                                            <span>{t('label.createdAt')}{task.createdAt}</span>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center space-x-2">
                                                        <button
                                                            onClick={() =>
                                                                alert(t('alert.downloadLabel', { id: task.id }))
                                                            }
                                                            className="flex items-center px-3 py-1.5 text-gray-600 border border-gray-200 rounded-lg text-xs font-bold hover:bg-gray-50 transition-colors"
                                                        >
                                                            <Download size={14} className="mr-1" />
                                                            {t('label.download')}
                                                        </button>
                                                        <button
                                                            onClick={() =>
                                                                alert(t('alert.reprintLabel', { id: task.id }))
                                                            }
                                                            className="flex items-center px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-xs font-bold hover:bg-indigo-700 transition-colors"
                                                        >
                                                            <Printer size={14} className="mr-1" />
                                                            {t('label.print')}
                                                        </button>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                        {MOCK_LABEL_TASKS.length === 0 && (
                                            <EmptyState icon={Tag} message={t('empty.labels')} />
                                        )}
                                    </div>
                                </div>
                            )}
                        </div>
                    

                {/* Footer Statistics */}
                {activeTab === 'STOCK' && filteredInventory.length > 0 && (
                    <div className="p-4 border-t border-gray-100 flex items-center justify-between text-sm">
                        <span className="text-gray-500">
                            {t('footer.showing', { count: filteredInventory.length, total: MOCK_INVENTORY.length })}
                        </span>
                        <div className="flex items-center space-x-6">
                            <div className="flex items-center">
                                <AlertCircle size={14} className="text-red-500 mr-1" />
                                <span className="text-gray-600">{t('alertCount')}</span>
                                <span className="ml-1 font-bold text-red-600">
                                    {filteredInventory.filter(item => item.alert).length}
                                </span>
                            </div>
                            <div className="flex items-center">
                                <CheckCircle size={14} className="text-green-500 mr-1" />
                                <span className="text-gray-600">{t('normalCount')}</span>
                                <span className="ml-1 font-bold text-green-600">
                                    {
                                        filteredInventory.filter(
                                            item => !item.alert && item.stock >= item.minStock * 1.2
                                        ).length
                                    }
                                </span>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}
