import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
    Database,
    Pill,
    Droplet,
    BookOpen,
    CreditCard,
    Search,
    Plus,
    Edit2,
    Trash2,
    X,
    Save
} from 'lucide-react'

// 数据类型定义
type CatalogType = 'DRUG' | 'MATERIAL' | 'EDUCATION' | 'FEES'

interface CatalogItem {
    id: string
    name: string
    spec: string
    unit: string
    price: string
}

interface ModalState {
    visible: boolean
    mode: 'add' | 'edit'
    item: CatalogItem | null
}

import { materialCatalogApi, drugCatalogApi } from '@/services/treatmentConfigApi'

// EDUCATION / FEES 暂无后端接口，保留本地数据；DRUG / MATERIAL 从 REST 加载
const LOCAL_EDUCATION: CatalogItem[] = [
    { id: 'E001', name: '饮食管理指南', spec: '慢性肾脏病患者营养管理', unit: '篇', price: '0.00' },
    { id: 'E002', name: '透析并发症预防', spec: '低血压/肌肉痉挛/感染预防', unit: '篇', price: '0.00' },
    { id: 'E003', name: '血管通路护理', spec: '内瘘/导管日常维护', unit: '篇', price: '0.00' },
    { id: 'E004', name: '水分控制指导', spec: '透析间期体重管理', unit: '篇', price: '0.00' },
    { id: 'E005', name: '用药依从性教育', spec: '降磷剂/降压药服用', unit: '篇', price: '0.00' },
]
const LOCAL_FEES: CatalogItem[] = [
    { id: 'F001', name: '血液透析', spec: '4小时标准透析', unit: '次', price: '420.00' },
    { id: 'F002', name: '血液灌流', spec: '串联血液灌流治疗', unit: '次', price: '680.00' },
    { id: 'F003', name: '血管通路穿刺', spec: '动静脉内瘘穿刺', unit: '次', price: '20.00' },
    { id: 'F004', name: '生命体征监测', spec: '治疗期间监测', unit: '次', price: '15.00' },
    { id: 'F005', name: '抢救费', spec: '透析期间抢救费用', unit: '次', price: '150.00' },
]

const INITIAL_DATA: Record<CatalogType, CatalogItem[]> = {
    DRUG: [],
    MATERIAL: [],
    EDUCATION: LOCAL_EDUCATION,
    FEES: LOCAL_FEES,
}

interface CatalogConfigItem {
  id: CatalogType
  labelKey: string
  icon: React.ComponentType<{ size?: number; className?: string }>
}

const CATALOG_CONFIGS: CatalogConfigItem[] = [
    { id: 'DRUG', labelKey: 'tab.drug', icon: Pill },
    { id: 'MATERIAL', labelKey: 'tab.material', icon: Droplet },
    { id: 'EDUCATION', labelKey: 'tab.education', icon: BookOpen },
    { id: 'FEES', labelKey: 'tab.fees', icon: CreditCard },
]

export default function MasterData() {
    const { t } = useTranslation('masterData')
    const [activeTab, setActiveTab] = useState<CatalogType>('DRUG')
    const [searchTerm, setSearchTerm] = useState('')
    const [data, setData] = useState(INITIAL_DATA)

    useEffect(() => {
        drugCatalogApi.list({ page: 1, pageSize: 200 }).then(res => {
            const items: CatalogItem[] = res.items.map(d => ({
                id: String(d.id), name: d.name, spec: d.spec, unit: d.baseUnit, price: '0.00',
            }))
            setData(prev => ({ ...prev, DRUG: items }))
        }).catch((err) => console.error('[MasterData] 药品目录加载失败', err))
        materialCatalogApi.list({ page: 1, pageSize: 200 }).then(res => {
            const items: CatalogItem[] = res.items.map(m => ({
                id: String(m.id), name: m.name, spec: m.spec, unit: m.unit, price: '0.00',
            }))
            setData(prev => ({ ...prev, MATERIAL: items }))
        }).catch((err) => console.error('[MasterData] 耗材目录加载失败', err))
    }, [])
    const [modalState, setModalState] = useState<ModalState>({
        visible: false,
        mode: 'add',
        item: null,
    })
    const [formData, setFormData] = useState<Partial<CatalogItem>>({})

    // 获取当前标签的数据
    const currentData = data[activeTab]

    // 过滤数据
    const filteredData = currentData.filter(
        item =>
            item.name.includes(searchTerm) ||
            item.spec.includes(searchTerm) ||
            item.id.includes(searchTerm)
    )

    // 打开新增 Modal
    const handleAdd = () => {
        setModalState({ visible: true, mode: 'add', item: null })
        setFormData({
            name: '',
            spec: '',
            unit: '',
            price: '0.00',
        })
    }

    // 打开编辑 Modal
    const handleEdit = (item: CatalogItem) => {
        setModalState({ visible: true, mode: 'edit', item })
        setFormData(item)
    }

    // 删除条目
    const handleDelete = (id: string) => {
        if (confirm(t('confirm.delete'))) {
            setData(prev => ({
                ...prev,
                [activeTab]: prev[activeTab].filter(item => item.id !== id),
            }))
        }
    }

    // 保存条目
    const handleSave = () => {
        if (!formData.name || !formData.unit) {
            alert(t('validation.required'))
            return
        }

        if (modalState.mode === 'add') {
            // 生成新 ID
            const prefix = activeTab.charAt(0)
            const maxId = Math.max(
                0,
                ...currentData.map(item => parseInt(item.id.slice(1)) || 0)
            )
            const newId = `${prefix}${String(maxId + 1).padStart(3, '0')}`

            const newItem: CatalogItem = {
                id: newId,
                name: formData.name || '',
                spec: formData.spec || '',
                unit: formData.unit || '',
                price: formData.price || '0.00',
            }

            setData(prev => ({
                ...prev,
                [activeTab]: [...prev[activeTab], newItem],
            }))
        } else if (modalState.item) {
            // 编辑现有条目
            setData(prev => ({
                ...prev,
                [activeTab]: prev[activeTab].map(item =>
                    item.id === modalState.item!.id
                        ? { ...item, ...formData }
                        : item
                ),
            }))
        }

        setModalState({ visible: false, mode: 'add', item: null })
    }

    // 关闭 Modal
    const handleCloseModal = () => {
        setModalState({ visible: false, mode: 'add', item: null })
    }

    // 获取当前目录配置
    const activeConfig = CATALOG_CONFIGS.find(c => c.id === activeTab)!

    return (
        <div className="h-full flex flex-col space-y-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <h2 className="text-2xl font-bold text-gray-800 flex items-center">
                    <Database size={28} className="mr-3 text-indigo-600" />
                    {t('title')}
                </h2>
                <button
                    onClick={handleAdd}
                    className="flex items-center px-4 py-2 bg-indigo-600 text-white rounded-xl text-sm font-bold hover:bg-indigo-700 transition-all shadow-md"
                >
                    <Plus size={18} className="mr-2" />
                    {t('addItem')}
                </button>
            </div>

            <div className="flex flex-col lg:flex-row gap-6 flex-1 overflow-hidden">
                {/* Sidebar Tabs - 移动端顶部 / 桌面端左侧 */}
                <div className="w-full lg:w-64 bg-white rounded-2xl shadow-sm border border-gray-100 p-2 shrink-0">
                    <div className="flex lg:flex-col overflow-x-auto lg:overflow-visible gap-1">
                        {CATALOG_CONFIGS.map(cat => (
                            <button
                                key={cat.id}
                                onClick={() => setActiveTab(cat.id)}
                                className={`flex items-center px-4 py-3 rounded-xl transition-all whitespace-nowrap ${
                                    activeTab === cat.id
                                        ? 'bg-indigo-50 text-indigo-700 shadow-sm'
                                        : 'text-gray-500 hover:bg-gray-50'
                                }`}
                            >
                                <cat.icon size={18} className="mr-3 shrink-0" />
                                <span className="font-bold text-sm">{t(cat.labelKey as never)}</span>
                            </button>
                        ))}
                    </div>
                </div>

                {/* Content Table */}
                <div className="flex-1 bg-white rounded-2xl shadow-sm border border-gray-100 flex flex-col overflow-hidden">
                    {/* Search Bar */}
                    <div className="p-4 border-b border-gray-50 flex items-center">
                        <div className="relative flex-1 max-w-sm">
                            <Search
                                className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                                size={16}
                            />
                            <input
                                type="text"
                                placeholder={t('search.placeholder', { catalog: t(activeConfig.labelKey as 'tab.drug' | 'tab.material' | 'tab.education' | 'tab.fees') })}
                                value={searchTerm}
                                onChange={e => setSearchTerm(e.target.value)}
                                className="w-full pl-9 pr-4 py-2 bg-gray-50 border-none rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 outline-none"
                            />
                        </div>
                        <div className="ml-4 text-sm text-gray-500">
                            {t('totalCount', { count: filteredData.length })}
                        </div>
                    </div>

                    {/* Table */}
                    <div className="flex-1 overflow-auto">
                        <table className="w-full text-left">
                            <thead className="bg-gray-50/50 text-gray-400 font-bold text-xs tracking-wider sticky top-0 z-10">
                                <tr>
                                    <th className="px-6 py-4">{t('table.name')}</th>
                                    <th className="px-6 py-4">{t('table.spec')}</th>
                                    <th className="px-6 py-4">{t('table.unit')}</th>
                                    <th className="px-6 py-4">{t('table.price')}</th>
                                    <th className="px-6 py-4 text-right">{t('table.action')}</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-50">
                                {filteredData.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-12 text-center">
                                            <div className="flex flex-col items-center text-gray-400">
                                                <Database size={48} className="mb-3 opacity-30" />
                                                <p className="text-sm">{t('empty')}</p>
                                            </div>
                                        </td>
                                    </tr>
                                ) : (
                                    filteredData.map(item => (
                                        <tr
                                            key={item.id}
                                            className="hover:bg-indigo-50/20 transition-colors"
                                        >
                                            <td className="px-6 py-4 font-bold text-gray-800">
                                                {item.name}
                                            </td>
                                            <td className="px-6 py-4 text-gray-500 text-sm">
                                                {item.spec || '-'}
                                            </td>
                                            <td className="px-6 py-4 text-gray-500 text-sm">
                                                {item.unit}
                                            </td>
                                            <td className="px-6 py-4 font-mono font-bold text-indigo-600">
                                                {item.price}
                                            </td>
                                            <td className="px-6 py-4 text-right space-x-2">
                                                <button
                                                    onClick={() => handleEdit(item)}
                                                    className="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-all inline-flex"
                                                >
                                                    <Edit2 size={16} />
                                                </button>
                                                <button
                                                    onClick={() => handleDelete(item.id)}
                                                    className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all inline-flex"
                                                >
                                                    <Trash2 size={16} />
                                                </button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* Modal */}
            {modalState.visible && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-2xl shadow-xl max-w-lg w-full p-6 space-y-4">
                        {/* Modal Header */}
                        <div className="flex items-center justify-between">
                            <h3 className="text-xl font-bold text-gray-800">
                                {modalState.mode === 'add' ? t('modal.addTitle') : t('modal.editTitle')}
                            </h3>
                            <button
                                onClick={handleCloseModal}
                                className="p-2 text-gray-400 hover:bg-gray-100 rounded-lg transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        {/* Form */}
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    {t('form.name')} <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={formData.name || ''}
                                    onChange={e =>
                                        setFormData(prev => ({ ...prev, name: e.target.value }))
                                    }
                                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                                    placeholder={t('form.namePlaceholder')}
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    {t('form.spec')}
                                </label>
                                <input
                                    type="text"
                                    value={formData.spec || ''}
                                    onChange={e =>
                                        setFormData(prev => ({ ...prev, spec: e.target.value }))
                                    }
                                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                                    placeholder={t('form.specPlaceholder')}
                                />
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">
                                        {t('form.unit')} <span className="text-red-500">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.unit || ''}
                                        onChange={e =>
                                            setFormData(prev => ({ ...prev, unit: e.target.value }))
                                        }
                                        className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                                        placeholder={t('form.unitPlaceholder')}
                                    />
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">
                                        {t('form.price')}
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.price || ''}
                                        onChange={e =>
                                            setFormData(prev => ({ ...prev, price: e.target.value }))
                                        }
                                        className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                                        placeholder={t('form.pricePlaceholder')}
                                    />
                                </div>
                            </div>
                        </div>

                        {/* Modal Actions */}
                        <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-100">
                            <button
                                onClick={handleCloseModal}
                                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg text-sm font-medium transition-colors"
                            >
                                {t('action.cancel')}
                            </button>
                            <button
                                onClick={handleSave}
                                className="flex items-center px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
                            >
                                <Save size={16} className="mr-2" />
                                {t('action.save')}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
