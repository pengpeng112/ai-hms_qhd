// Basic Info Tab - 基本信息档案

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { UserCircle, ClipboardList, Briefcase, Phone, HeartHandshake, MapPin, Files, Edit3, Plus, Save, X, Trash2 } from 'lucide-react'
import { message, Spin, Cascader } from 'antd'
import { SectionHeader, DetailCard, FormField } from '@/components/ui'
import { ElectronicDocumentModal } from '@/components/patient/modals'
import { restApi, type PatientBasicInfoResponse } from '@/services/restClient'
import { dictCache, DICT_TYPES, type CascaderOption } from '@/services/dictApi'
import { useDictNameMaps, getNameFromMap } from '@/hooks/useDictName'
import type { Patient } from '@/types/original'

interface BasicInfoTabProps {
  patient: Patient
}

// 空值显示辅助函数
const formatValue = (value: string | undefined | null, suffix = ''): string => {
  if (!value || value === '') return '-'
  return suffix ? `${value}${suffix}` : value
}

const formatNumberValue = (value: number | undefined | null): string => {
  if (value === null || value === undefined) return '-'
  return String(value)
}

// 家属联系人类型
interface FamilyContact {
  id: string
  name: string
  phone: string
  type: 'primary' | 'family' | 'emergency'
  relation: string
}

export default function BasicInfoTab({ patient }: BasicInfoTabProps) {
  const { t } = useTranslation('patient')

  // 字典名称映射（用于只读模式显示名称而非代码）
  const dictTypeCodes = useMemo(() => [
    DICT_TYPES.ID_TYPE,
    DICT_TYPES.PATIENT_TYPE,
    DICT_TYPES.VISIT_CATEGORY,
    DICT_TYPES.BLOOD_TYPE_ABO,
    DICT_TYPES.BLOOD_TYPE_RH,
    DICT_TYPES.EDUCATION_LEVEL,
    DICT_TYPES.MARITAL_STATUS,
    DICT_TYPES.INSURANCE_TYPE,
  ], [])
  const dictNameMaps = useDictNameMaps(dictTypeCodes)

  // 性别转换函数
  const genderToChinese = (gender: string | undefined): string => {
    if (gender === 'M') return '男'
    if (gender === 'F') return '女'
    return gender || ''
  }

  const chineseToGender = (value: string | undefined | null): string | null => {
    if (!value) return null
    const trimmed = value.trim()
    if (trimmed === '男') return 'M'
    if (trimmed === '女') return 'F'
    // 如果已经是 M/F，直接返回
    if (trimmed === 'M' || trimmed === 'F') return trimmed
    return null
  }

  // 状态
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)
  const [basicInfo, setBasicInfo] = useState<PatientBasicInfoResponse | null>(null)
  const [isDocModalOpen, setIsDocModalOpen] = useState(false)

  // 字典选项状态
  const [dictOptions, setDictOptions] = useState<Record<string, Array<{ value: string; label: string }>>>({})
  // 医保类型树形选项（用于 Cascader）
  const [insuranceOptions, setInsuranceOptions] = useState<CascaderOption[]>([])

  // 家属联系人列表（本地状态）
  const [familyContacts, setFamilyContacts] = useState<FamilyContact[]>([])
  const [isContactModalOpen, setIsContactModalOpen] = useState(false)

  // 加载字典数据
  useEffect(() => {
    const loadDictData = async () => {
      try {
        const [
          idTypes,
          patientTypes,
          visitCategories,
          aboBloodTypes,
          rhBloodTypes,
          educationLevels,
          maritalStatuses,
          insuranceTree
        ] = await Promise.all([
          dictCache.getOptions(DICT_TYPES.ID_TYPE),
          dictCache.getOptions(DICT_TYPES.PATIENT_TYPE),
          dictCache.getOptions(DICT_TYPES.VISIT_CATEGORY),
          dictCache.getOptions(DICT_TYPES.BLOOD_TYPE_ABO),
          dictCache.getOptions(DICT_TYPES.BLOOD_TYPE_RH),
          dictCache.getOptions(DICT_TYPES.EDUCATION_LEVEL),
          dictCache.getOptions(DICT_TYPES.MARITAL_STATUS),
          dictCache.getCascaderOptions(DICT_TYPES.INSURANCE_TYPE),
        ])

        setDictOptions({
          [DICT_TYPES.ID_TYPE]: idTypes,
          [DICT_TYPES.PATIENT_TYPE]: patientTypes,
          [DICT_TYPES.VISIT_CATEGORY]: visitCategories,
          [DICT_TYPES.BLOOD_TYPE_ABO]: aboBloodTypes,
          [DICT_TYPES.BLOOD_TYPE_RH]: rhBloodTypes,
          [DICT_TYPES.EDUCATION_LEVEL]: educationLevels,
          [DICT_TYPES.MARITAL_STATUS]: maritalStatuses,
        })

        // 更新医保类型选项
        if (insuranceTree && insuranceTree.length > 0) {
          setInsuranceOptions(insuranceTree)
        }
      } catch (error) {
        console.error('加载字典数据失败:', error)
      }
    }
    loadDictData()
  }, [])

  // 获取数据
  const fetchBasicInfo = useCallback(async () => {
    if (!patient.id) return

    setLoading(true)
    try {
      const data = await restApi.getPatientBasicInfo(patient.id)
      setBasicInfo(data)

      // 构建默认联系人列表（只有紧急联系人，不包括本人）
      const contacts: FamilyContact[] = []
      if (data.contactInfo.contactName && data.contactInfo.contactPhone) {
        contacts.push({
          id: 'emergency',
          name: data.contactInfo.contactName,
          phone: data.contactInfo.contactPhone,
          type: 'emergency',
          relation: '-',
        })
      }
      setFamilyContacts(contacts)
    } catch (error) {
      console.error('获取基本信息失败:', error)
      message.warning('获取基本信息失败，显示部分数据')
    } finally {
      setLoading(false)
    }
  }, [patient.id])

  useEffect(() => {
    fetchBasicInfo()
  }, [fetchBasicInfo])

  // 保存基本信息
  const handleSave = async () => {
    if (!basicInfo) return

    setSaving(true)
    try {
      // 从 DOM 获取所有输入框的实际值
      const getInputValue = (selector: string): string | null => {
        const input = document.querySelector<HTMLInputElement>(selector)
        return input?.value?.trim() || null
      }

      // 清理单位信息的辅助函数
      const cleanUnit = (value: string | null, unit: string): string | null => {
        if (!value) return null
        const cleaned = value.replace(new RegExp(unit, 'g'), '').trim()
        return cleaned || null
      }

      const parseNullableInt = (value: string | null, fieldName: string): number | null => {
        if (!value) return null
        if (value === '-') return null
        const parsed = Number(value)
        if (!Number.isInteger(parsed) || parsed <= 0) {
          throw new Error(`${fieldName} 必须为正整数`)
        }
        return parsed
      }

      // 构建请求数据，使用 DOM 中输入框的实际值
      const requestData = {
        personalInfo: {
          name: getInputValue('input[data-field="name"]') || basicInfo.personalInfo.name,
          pinyin: getInputValue('input[data-field="pinyin"]'),
          birthday: getInputValue('input[data-field="birthday"]'),
          gender: chineseToGender(getInputValue('input[data-field="gender"]')) || basicInfo.personalInfo.gender,
          ethnicity: getInputValue('input[data-field="ethnicity"]'),
          idType: getInputValue('input[data-field="idType"]') || basicInfo.personalInfo.idType,
          idNumber: getInputValue('input[data-field="idNumber"]'),
          patientType: getInputValue('input[data-field="patientType"]'),
        },
        medicalInfo: {
          visitCategory: getInputValue('input[data-field="visitCategory"]'),
          admissionNo: getInputValue('input[data-field="admissionNo"]'),
          visitNo: getInputValue('input[data-field="visitNo"]'),
          medicalRecordNo: getInputValue('input[data-field="medicalRecordNo"]'),
          insuranceNo: getInputValue('input[data-field="insuranceNo"]'),
          hdisPatientId: parseNullableInt(getInputValue('input[data-field="hdisPatientId"]'), 'HDIS患者ID'),
          insuranceType: getInputValue('input[data-field="insuranceType"]'),
          dialysisNo: getInputValue('input[data-field="dialysisNo"]'),
          doctorName: getInputValue('input[data-field="doctorName"]'),
          nurseName: getInputValue('input[data-field="nurseName"]'),
          firstDialysisDate: getInputValue('input[data-field="firstDialysisDate"]'),
          firstHospitalDate: getInputValue('input[data-field="firstHospitalDate"]'),
          firstDialysisHospital: getInputValue('input[data-field="firstDialysisHospital"]'),
        },
        vitalSocialInfo: {
          height: cleanUnit(getInputValue('input[data-field="height"]'), 'cm'),
          dryWeight: basicInfo.vitalSocialInfo.dryWeight ?? null,
          aboBloodType: getInputValue('input[data-field="aboBloodType"]'),
          rhBloodType: getInputValue('input[data-field="rhBloodType"]'),
          educationLevel: getInputValue('input[data-field="educationLevel"]'),
          occupation: getInputValue('input[data-field="occupation"]'),
          maritalStatus: getInputValue('input[data-field="maritalStatus"]'),
          workplace: getInputValue('input[data-field="workplace"]'),
        },
        contactInfo: {
          phone: getInputValue('input[data-field="phone"]'),
          wechat: getInputValue('input[data-field="wechat"]'),
          landline: getInputValue('input[data-field="landline"]'),
          address: getInputValue('input[data-field="address"]'),
          district: getInputValue('input[data-field="district"]'),
          contactName: basicInfo.contactInfo.contactName || null,
          contactPhone: basicInfo.contactInfo.contactPhone || null,
        },
      }

      console.log('发送的数据:', JSON.stringify(requestData, null, 2))

      await restApi.updatePatientBasicInfo(patient.id, requestData)

      message.success('保存成功')
      setEditing(false)
      await fetchBasicInfo() // 刷新数据
    } catch (error) {
      console.error('保存失败:', error)
      // 打印错误详情
      if (error instanceof Error) {
        console.error('错误信息:', error.message)
        message.error(error.message)
        return
      }
      message.error('保存失败，请稍后重试')
    } finally {
      setSaving(false)
    }
  }

  // 添加联系人
  const handleAddContact = (contact: Omit<FamilyContact, 'id'>) => {
    const newContact: FamilyContact = {
      ...contact,
      id: Date.now().toString(),
    }
    setFamilyContacts([...familyContacts, newContact])
  }

  // 删除联系人
  const handleDeleteContact = (id: string) => {
    setFamilyContacts(familyContacts.filter(c => c.id !== id))
    message.success('联系人已删除')
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spin size="large" />
      </div>
    )
  }

  const info = basicInfo

  return (
    <div className="space-y-6 animate-fade-in pb-10">
      {/* 顶部工具栏 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-bold text-slate-800">基本信息档案</h2>
          {!editing && (
            <button
              onClick={() => setEditing(true)}
              className="px-3 py-1.5 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors flex items-center"
            >
              <Edit3 size={14} className="mr-1" />
              编辑
            </button>
          )}
        </div>
        {editing && (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setEditing(false)}
              disabled={saving}
              className="px-3 py-1.5 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 transition-colors flex items-center disabled:opacity-50"
            >
              <X size={14} className="mr-1" />
              取消
            </button>
            <button
              onClick={handleSave}
              disabled={saving}
              className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors flex items-center disabled:opacity-50"
            >
              {saving ? (
                <>
                  <div className="w-3 h-3 border-2 border-white border-t-transparent rounded-full mr-1.5 animate-spin"></div>
                  保存中...
                </>
              ) : (
                <>
                  <Save size={14} className="mr-1" />
                  保存
                </>
              )}
            </button>
          </div>
        )}
      </div>

      <div className="grid grid-cols-12 gap-6">
        <div className="col-span-12 lg:col-span-8 space-y-6">
          {/* 核心身份信息 */}
          <DetailCard>
            <SectionHeader icon={UserCircle} title={t('basicInfo.section.coreIdentity')} />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-y-5 gap-x-6 mt-4">
              <FormField
                label={t('basicInfo.field.patientName')}
                defaultValue={info?.personalInfo.name || patient.name}
                required
                dataField="name"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.namePinyin')}
                defaultValue={formatValue(info?.personalInfo.pinyin)}
                dataField="pinyin"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.internalName')}
                defaultValue={patient.name}
                required
                readOnly={true}
              />
              <FormField
                label={t('basicInfo.field.birthday')}
                defaultValue={formatValue(info?.personalInfo.birthday)}
                dataField="birthday"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.currentAge')}
                defaultValue={`${info?.personalInfo.age || patient.age} ${t('basicInfo.unit.yearsOld')}`}
                readOnly={true}
              />
              <FormField
                label={t('info.gender')}
                defaultValue={genderToChinese(info?.personalInfo.gender || patient.gender)}
                dataField="gender"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.ethnicity')}
                defaultValue={formatValue(info?.personalInfo.ethnicity)}
                dataField="ethnicity"
                readOnly={!editing}
              />
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.idType')}
                  defaultValue={info?.personalInfo.idType || t('basicInfo.value.idCard')}
                  required
                  options={dictOptions[DICT_TYPES.ID_TYPE] || [{ value: 'ID_CARD', label: '身份证' }]}
                  dataField="idType"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.idType')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.ID_TYPE] || new Map(), info?.personalInfo.idType) || t('basicInfo.value.idCard')}
                  required
                  dataField="idType"
                  readOnly={true}
                />
              )}
              <div className="col-span-2">
                <FormField
                  label={t('basicInfo.field.idNumber')}
                  defaultValue={formatValue(info?.personalInfo.idNumber)}
                  required
                  dataField="idNumber"
                  readOnly={!editing}
                />
              </div>
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.patientType')}
                  defaultValue={info?.personalInfo.patientType || patient.patientType}
                  options={dictOptions[DICT_TYPES.PATIENT_TYPE] || [{ value: 'OUTPATIENT', label: '门诊' }, { value: 'INPATIENT', label: '住院' }]}
                  dataField="patientType"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.patientType')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.PATIENT_TYPE] || new Map(), info?.personalInfo.patientType || patient.patientType)}
                  dataField="patientType"
                  readOnly={true}
                />
              )}
              <FormField
                label={t('label.autoConfirm')}
                defaultValue={t('label.yes')}
                readOnly={true}
              />
            </div>
          </DetailCard>

          {/* 医疗登记信息 */}
          <DetailCard>
            <SectionHeader icon={ClipboardList} title={t('basicInfo.section.medicalRegistration')} />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-y-5 gap-x-6 mt-4">
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.visitCategory')}
                  defaultValue={formatValue(info?.medicalInfo.visitCategory) || t('type.outpatient')}
                  required
                  options={dictOptions[DICT_TYPES.VISIT_CATEGORY] || [{ value: 'OUTPATIENT', label: '门诊' }, { value: 'INPATIENT', label: '住院' }]}
                  dataField="visitCategory"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.visitCategory')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.VISIT_CATEGORY] || new Map(), info?.medicalInfo.visitCategory) || t('type.outpatient')}
                  required
                  dataField="visitCategory"
                  readOnly={true}
                />
              )}
              <FormField
                label={t('basicInfo.field.admissionNo')}
                defaultValue={formatValue(info?.medicalInfo.admissionNo)}
                dataField="admissionNo"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.visitNo')}
                defaultValue={formatValue(info?.medicalInfo.visitNo)}
                required
                dataField="visitNo"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.medicalRecordNo')}
                defaultValue={formatValue(info?.medicalInfo.medicalRecordNo)}
                dataField="medicalRecordNo"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.insuranceNo')}
                defaultValue={formatValue(info?.medicalInfo.insuranceNo)}
                required
                dataField="insuranceNo"
                readOnly={!editing}
              />
              <FormField
                label="HDIS患者ID"
                defaultValue={editing
                  ? (info?.medicalInfo.hdisPatientId !== undefined && info?.medicalInfo.hdisPatientId !== null
                    ? String(info.medicalInfo.hdisPatientId)
                    : '')
                  : formatNumberValue(info?.medicalInfo.hdisPatientId)}
                dataField="hdisPatientId"
                readOnly={!editing}
              />
              {editing ? (
                <CascaderFormField
                  label={t('basicInfo.field.insuranceType')}
                  defaultValue={info?.medicalInfo.insuranceType || patient.insuranceType}
                  required
                  options={insuranceOptions.length > 0 ? insuranceOptions : [{ value: '自费', label: '自费' }]}
                  dataField="insuranceType"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.insuranceType')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.INSURANCE_TYPE] || new Map(), info?.medicalInfo.insuranceType || patient.insuranceType)}
                  required
                  dataField="insuranceType"
                  readOnly={true}
                />
              )}
              <FormField
                label={t('basicInfo.field.dialysisNo')}
                defaultValue={formatValue(info?.medicalInfo.dialysisNo)}
                dataField="dialysisNo"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.attendingDoctor')}
                defaultValue={info?.medicalInfo.doctorName || patient.doctorName}
                dataField="doctorName"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.responsibleDoctor')}
                defaultValue={info?.medicalInfo.doctorName || patient.doctorName}
                readOnly={true}
              />
              <FormField
                label={t('basicInfo.field.responsibleNurse')}
                defaultValue={formatValue(info?.medicalInfo.nurseName)}
                dataField="nurseName"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.hospFirstDialysisDate')}
                defaultValue={formatValue(info?.medicalInfo.firstHospitalDate)}
                dataField="firstHospitalDate"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.firstDialysisDate')}
                defaultValue={formatValue(info?.medicalInfo.firstDialysisDate)}
                dataField="firstDialysisDate"
                readOnly={!editing}
              />
              <div className="col-span-2">
                <FormField
                  label={t('basicInfo.field.firstDialysisHospital')}
                  defaultValue={formatValue(info?.medicalInfo.firstDialysisHospital)}
                  dataField="firstDialysisHospital"
                  readOnly={!editing}
                />
              </div>
              <FormField
                label={t('basicInfo.field.currentDialysisAge')}
                defaultValue={formatValue(info?.medicalInfo.currentDialysisAge)}
                readOnly={true}
              />
            </div>
          </DetailCard>

          {/* 体征与社会状态 */}
          <DetailCard>
            <SectionHeader icon={Briefcase} title={t('basicInfo.section.vitalSignsAndSocial')} />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-y-5 gap-x-6 mt-4">
              <FormField
                label={t('basicInfo.field.height')}
                defaultValue={formatValue(info?.vitalSocialInfo.height, ' cm')}
                required
                dataField="height"
                readOnly={!editing}
              />
              <FormField
                label={t('info.dryWeight')}
                defaultValue={`${info?.vitalSocialInfo.dryWeight || patient.dryWeight} kg`}
                required
                readOnly={true}
              />
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.aboBloodType')}
                  defaultValue={formatValue(info?.vitalSocialInfo.aboBloodType)}
                  options={dictOptions[DICT_TYPES.BLOOD_TYPE_ABO] || [{ value: 'A', label: 'A型' }, { value: 'B', label: 'B型' }, { value: 'AB', label: 'AB型' }, { value: 'O', label: 'O型' }]}
                  dataField="aboBloodType"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.aboBloodType')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.BLOOD_TYPE_ABO] || new Map(), info?.vitalSocialInfo.aboBloodType)}
                  dataField="aboBloodType"
                  readOnly={true}
                />
              )}
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.rhBloodType')}
                  defaultValue={formatValue(info?.vitalSocialInfo.rhBloodType)}
                  options={dictOptions[DICT_TYPES.BLOOD_TYPE_RH] || [{ value: 'POSITIVE', label: 'Rh阳性' }, { value: 'NEGATIVE', label: 'Rh阴性' }, { value: 'UNKNOWN', label: 'Rh未知' }]}
                  dataField="rhBloodType"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.rhBloodType')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.BLOOD_TYPE_RH] || new Map(), info?.vitalSocialInfo.rhBloodType)}
                  dataField="rhBloodType"
                  readOnly={true}
                />
              )}
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.educationLevel')}
                  defaultValue={formatValue(info?.vitalSocialInfo.educationLevel)}
                  options={dictOptions[DICT_TYPES.EDUCATION_LEVEL] || [
                    { value: 'PRIMARY', label: '小学' },
                    { value: 'JUNIOR_HIGH', label: '初中' },
                    { value: 'HIGH_SCHOOL', label: '高中' },
                    { value: 'VOCATIONAL', label: '中专' },
                    { value: 'COLLEGE', label: '大专' },
                    { value: 'UNDERGRADUATE', label: '本科' },
                    { value: 'POSTGRADUATE', label: '研究生' },
                    { value: 'OTHER', label: '其他' }
                  ]}
                  dataField="educationLevel"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.educationLevel')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.EDUCATION_LEVEL] || new Map(), info?.vitalSocialInfo.educationLevel)}
                  dataField="educationLevel"
                  readOnly={true}
                />
              )}
              <FormField
                label={t('basicInfo.field.occupation')}
                defaultValue={formatValue(info?.vitalSocialInfo.occupation)}
                dataField="occupation"
                readOnly={!editing}
              />
              {editing ? (
                <SelectFormField
                  label={t('basicInfo.field.maritalStatus')}
                  defaultValue={formatValue(info?.vitalSocialInfo.maritalStatus)}
                  options={dictOptions[DICT_TYPES.MARITAL_STATUS] || [
                    { value: 'UNMARRIED', label: '未婚' },
                    { value: 'MARRIED', label: '已婚' },
                    { value: 'DIVORCED', label: '离异' },
                    { value: 'WIDOWED', label: '丧偶' }
                  ]}
                  dataField="maritalStatus"
                  readOnly={!editing}
                />
              ) : (
                <FormField
                  label={t('basicInfo.field.maritalStatus')}
                  defaultValue={getNameFromMap(dictNameMaps[DICT_TYPES.MARITAL_STATUS] || new Map(), info?.vitalSocialInfo.maritalStatus)}
                  dataField="maritalStatus"
                  readOnly={true}
                />
              )}
              <FormField
                label={t('basicInfo.field.workplace')}
                defaultValue={formatValue(info?.vitalSocialInfo.workplace)}
                dataField="workplace"
                readOnly={!editing}
              />
            </div>
          </DetailCard>
        </div>

        <div className="col-span-12 lg:col-span-4 space-y-6">
          {/* 联系方式 */}
          <DetailCard className="bg-[#0f172a] text-white border-none relative overflow-hidden ring-1 ring-white/10 shadow-2xl">
            <div className="absolute top-0 right-0 w-32 h-32 bg-blue-500/10 rounded-full -translate-y-1/2 translate-x-1/2 blur-2xl"></div>
            <SectionHeader icon={Phone} title={t('basicInfo.section.contact')} dark />
            <div className="space-y-5 mt-4">
              <FormField
                label={t('basicInfo.field.mobilePhone')}
                defaultValue={formatValue(info?.contactInfo.phone)}
                required
                dark
                dataField="phone"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.wechat')}
                defaultValue={formatValue(info?.contactInfo.wechat)}
                dark
                dataField="wechat"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.landline')}
                defaultValue={formatValue(info?.contactInfo.landline)}
                dark
                dataField="landline"
                readOnly={!editing}
              />
            </div>
          </DetailCard>

          {/* 家属与紧急联系人 */}
          <DetailCard className="bg-slate-50/50 border-slate-100">
            <SectionHeader
              icon={HeartHandshake}
              title={t('basicInfo.section.familyContacts')}
              action={
                <button
                  onClick={() => setIsContactModalOpen(true)}
                  className="p-1 hover:bg-blue-50 text-blue-600 rounded-lg transition-all"
                >
                  <Plus size={16} />
                </button>
              }
            />
            <div className="space-y-3 mt-4">
              {familyContacts.length === 0 ? (
                <div className="p-4 bg-white rounded-2xl border border-dashed border-slate-200 text-center">
                  <p className="text-xs text-slate-400">暂无联系人</p>
                </div>
              ) : (
                familyContacts.map((contact) => (
                  <div key={contact.id} className="p-4 bg-white rounded-2xl border border-slate-100 flex items-center justify-between group hover:border-blue-300 hover:shadow-md transition-all">
                    <div className="flex-1">
                      <p className="text-sm font-black text-slate-800">{contact.name}
                        <span className="ml-2 text-[10px] text-slate-400 font-normal bg-slate-100 px-1.5 py-0.5 rounded">
                          {contact.type === 'primary' ? t('basicInfo.contactType.primary') :
                           contact.type === 'emergency' ? '紧急联系人' :
                           t('basicInfo.contactType.family')}
                        </span>
                      </p>
                      <p className="text-xs text-blue-600 font-mono mt-1 font-bold">{contact.phone}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] font-black text-slate-500 bg-slate-50 px-2 py-1 rounded-lg border border-slate-100 uppercase">
                        {contact.relation}
                      </span>
                      <button
                        onClick={() => handleDeleteContact(contact.id)}
                        className="p-1.5 hover:bg-red-50 text-red-400 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                        title="删除"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </DetailCard>

          {/* 地址信息 */}
          <DetailCard>
            <SectionHeader icon={MapPin} title={t('basicInfo.section.address')} />
            <div className="space-y-5 mt-4">
              <FormField
                label={t('basicInfo.field.region')}
                defaultValue={formatValue(info?.contactInfo.district)}
                dataField="district"
                readOnly={!editing}
              />
              <FormField
                label={t('basicInfo.field.detailedAddress')}
                defaultValue={formatValue(info?.contactInfo.address)}
                dataField="address"
                readOnly={!editing}
              />
            </div>
          </DetailCard>

          {/* 关联电子文书 */}
          <DetailCard className="bg-slate-50/50 border-slate-100">
            <SectionHeader icon={Files} title={t('basicInfo.section.electronicDocs')} action={<button onClick={() => setIsDocModalOpen(true)} className="p-1 hover:bg-blue-50 text-blue-600 rounded-lg transition-all"><Plus size={16} /></button>} />
            <div className="space-y-3 mt-4">
              <div className="p-4 bg-white rounded-2xl border border-dashed border-slate-200 text-center">
                <p className="text-xs text-slate-400">暂无电子文书</p>
              </div>
              <button className="w-full py-2 text-[10px] font-black text-slate-400 hover:text-blue-600 transition-colors border-t border-slate-100 mt-1 uppercase tracking-tighter">
                {t('basicInfo.action.viewMoreDocs')}
              </button>
            </div>
          </DetailCard>
        </div>
      </div>

      {/* 添加联系人弹窗 */}
      <AddFamilyContactModal
        isOpen={isContactModalOpen}
        onClose={() => setIsContactModalOpen(false)}
        onAdd={handleAddContact}
      />

      {/* 电子文书弹窗 */}
      <ElectronicDocumentModal
        isOpen={isDocModalOpen}
        onClose={() => setIsDocModalOpen(false)}
      />
    </div>
  )
}

// SelectFormField - 下拉选择字段组件（用于字典选择）
interface SelectFormFieldProps {
  label: string
  defaultValue?: string
  required?: boolean
  options: Array<{ value: string; label: string }>
  dataField: string
  readOnly?: boolean
  dark?: boolean
}

function SelectFormField({ label, defaultValue, required, options, dataField, readOnly = false, dark = false }: SelectFormFieldProps) {
  const labelColor = dark ? 'text-slate-400' : 'text-slate-500'

  const getSelectStyles = () => {
    if (dark) {
      return readOnly
        ? 'bg-slate-700 border-slate-600 text-slate-400'
        : 'bg-slate-800 border-slate-700 text-white focus:ring-1 focus:ring-blue-400 focus:border-blue-400'
    }
    return readOnly
      ? 'bg-slate-50 text-slate-400 border-slate-200'
      : 'bg-white border-slate-300 focus:ring-1 focus:ring-blue-500 focus:border-blue-500'
  }

  return (
    <div className="flex flex-col gap-1.5">
      <label className={`text-[11px] font-bold flex items-center ${labelColor}`}>
        {required && <span className="text-red-500 mr-0.5">*</span>}
        {label}
      </label>
      {/* 使用受控的 value 属性，并添加隐藏的 input 用于数据获取 */}
      <input
        type="hidden"
        data-field={dataField}
        defaultValue={defaultValue}
      />
      <select
        name={dataField}
        defaultValue={defaultValue}
        disabled={readOnly}
        className={`w-full h-10 px-3 border rounded-lg text-sm outline-none transition-all ${getSelectStyles()}`}
        onChange={(e) => {
          // 同步更新隐藏 input 的值
          const hiddenInput = document.querySelector<HTMLInputElement>(`input[data-field="${dataField}"]`)
          if (hiddenInput) {
            hiddenInput.value = e.target.value
          }
        }}
      >
        {options.map(opt => (
          <option key={opt.value} value={opt.label}>{opt.label}</option>
        ))}
      </select>
    </div>
  )
}

// CascaderFormField - 级联选择字段组件（用于医保类型等）
interface CascaderFormFieldProps {
  label: string
  defaultValue?: string
  required?: boolean
  options: CascaderOption[]
  dataField: string
  readOnly?: boolean
}

function CascaderFormField({ label, defaultValue, required, options, dataField, readOnly = false }: CascaderFormFieldProps) {
  // 解析默认值，转换为 Cascader 需要的数组格式
  const getDefaultValue = () => {
    if (!defaultValue) return []
    // 尝试在选项中查找匹配的值
    for (const parent of options) {
      if (parent.children && parent.children.length > 0) {
        const found = parent.children.find((child: CascaderOption) => child.label === defaultValue)
        if (found) return [parent.value, found.value]
      } else if (parent.label === defaultValue) {
        return [parent.value]
      }
    }
    return []
  }

  // 获取显示的文本值（用于保存到隐藏 input）
  const getDisplayValue = (value: string[]): string => {
    if (!value || value.length === 0) return defaultValue || ''
    if (value.length === 1) {
      // 只有一级，直接返回 label
      const parent = options.find(o => o.value === value[0])
      return parent?.label || ''
    }
    if (value.length === 2) {
      // 有二级，返回第二级的 label
      const parent = options.find(o => o.value === value[0])
      if (parent?.children) {
        const child = parent.children.find((c: CascaderOption) => c.value === value[1])
        return child?.label || ''
      }
    }
    return defaultValue || ''
  }

  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-[11px] font-bold flex items-center text-slate-500">
        {required && <span className="text-red-500 mr-0.5">*</span>}
        {label}
      </label>
      {readOnly ? (
        <div className="w-full h-10 px-3 bg-slate-50 text-slate-400 border border-slate-200 rounded-lg text-sm flex items-center">
          {defaultValue || '-'}
        </div>
      ) : (
        <>
          {/* 隐藏的 input 用于保存实际值 */}
          <input
            type="hidden"
            data-field={dataField}
            defaultValue={defaultValue}
          />
          <Cascader
            defaultValue={getDefaultValue()}
            options={options}
            placeholder="请选择医保类型"
            style={{ width: '100%' }}
            size="small"
            expandTrigger="click"
            changeOnSelect
            onChange={(value: string[]) => {
              // 将选中项的 label 保存到隐藏 input
              const displayValue = getDisplayValue(value)
              const input = document.querySelector<HTMLInputElement>(`input[data-field="${dataField}"]`)
              if (input) input.value = displayValue
            }}
          />
        </>
      )}
    </div>
  )
}

// 添加联系人弹窗组件
interface AddFamilyContactModalProps {
  isOpen: boolean
  onClose: () => void
  onAdd: (contact: Omit<FamilyContact, 'id'>) => void
}

function AddFamilyContactModal({ isOpen, onClose, onAdd }: AddFamilyContactModalProps) {
  const [contact, setContact] = useState({
    name: '',
    phone: '',
    type: 'family' as 'primary' | 'family' | 'emergency',
    relation: '',
  })

  const handleSubmit = () => {
    if (!contact.name.trim()) {
      message.error('请输入联系人姓名')
      return
    }
    if (!contact.phone.trim()) {
      message.error('请输入联系电话')
      return
    }
    if (contact.type !== 'primary' && !contact.relation.trim()) {
      message.error('请输入与患者关系')
      return
    }

    onAdd(contact)
    handleClose()
  }

  const handleClose = () => {
    setContact({
      name: '',
      phone: '',
      type: 'family',
      relation: '',
    })
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-2xl shadow-xl w-full max-w-md mx-4">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h3 className="text-lg font-bold text-gray-800">添加联系人</h3>
          <button
            onClick={handleClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X size={20} className="text-gray-400" />
          </button>
        </div>

        {/* Form */}
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              联系人类型 <span className="text-red-500">*</span>
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setContact({ ...contact, type: 'family' })}
                className={`flex-1 py-2 px-4 rounded-lg border-2 font-medium transition-all ${
                  contact.type === 'family'
                    ? 'border-blue-500 bg-blue-50 text-blue-700'
                    : 'border-gray-200 text-gray-600 hover:border-gray-300'
                }`}
              >
                家属
              </button>
              <button
                type="button"
                onClick={() => setContact({ ...contact, type: 'emergency' })}
                className={`flex-1 py-2 px-4 rounded-lg border-2 font-medium transition-all ${
                  contact.type === 'emergency'
                    ? 'border-red-500 bg-red-50 text-red-700'
                    : 'border-gray-200 text-gray-600 hover:border-gray-300'
                }`}
              >
                紧急联系人
              </button>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              姓名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={contact.name}
              onChange={(e) => setContact({ ...contact, name: e.target.value })}
              placeholder="请输入联系人姓名"
              className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              电话 <span className="text-red-500">*</span>
            </label>
            <input
              type="tel"
              value={contact.phone}
              onChange={(e) => setContact({ ...contact, phone: e.target.value })}
              placeholder="请输入联系电话"
              className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
            />
          </div>

          {contact.type !== 'primary' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                与患者关系 <span className="text-red-500">*</span>
              </label>
              <select
                value={contact.relation}
                onChange={(e) => setContact({ ...contact, relation: e.target.value })}
                className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
              >
                <option value="">请选择</option>
                <option value="配偶">配偶</option>
                <option value="父亲">父亲</option>
                <option value="母亲">母亲</option>
                <option value="子女">子女</option>
                <option value="兄弟">兄弟</option>
                <option value="姐妹">姐妹</option>
                <option value="其他">其他</option>
              </select>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100 bg-gray-50 rounded-b-2xl">
          <button
            onClick={handleClose}
            className="px-6 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-100 transition-colors"
          >
            取消
          </button>
          <button
            onClick={handleSubmit}
            className="px-6 py-2.5 bg-blue-600 rounded-lg text-sm font-medium text-white hover:bg-blue-700 transition-colors"
          >
            确认添加
          </button>
        </div>
      </div>
    </div>
  )
}
