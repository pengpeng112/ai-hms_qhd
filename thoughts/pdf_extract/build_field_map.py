"""
自动生成「新血透模型 → 老血透物理表」字段级对照表。

- 新侧源：DATABASE_DESIGN.md（第 3 章以下的每张表字段表格）
- 老侧源：老血透数据库表结构-合并版.md（复用 parse_struct 的结果）

输出：docs/migration-field-map.md
"""
import re, os, sys, collections
sys.path.insert(0, 'thoughts/pdf_extract')
from build_legacy_merge import parse_struct

NEW_MD = 'DATABASE_DESIGN.md'
LEGACY_MERGED = '老血透数据库表结构-合并版.md'

# -----------------------------------------------------------------
# 表级映射：新表 → 老表（若老库无对应，value=None 并标注原因）
# -----------------------------------------------------------------
TABLE_MAP = {
    # 新表名                              : (老表名 | None, 类别, 说明)
    'users':                                (None, 'app-only',    '老库无独立用户表，旧系统使用 Identity/Organ 体系；保留应用层表'),
    'patients':                             ('Register_PatientInfomation', 'rename',   '老库合并了档案扩展，字段最丰富的主表'),
    'patient_basic_infos':                  ('Register_PatientInfomation + Register_Hospitalization + Register_IDInfomation + Register_FamilyMember', 'fold', '新表为独立 1:1 扩展，老库字段分散在 4 张 Register 表'),
    'medical_histories':                    ('Register_MedicalHistory + Register_Allergen + Register_Complication + Register_Diagnosis + Register_Pathology + Register_Protopathy + Register_Tumor', 'split-to-many', '新表一张扁平 33 列；老库为 7 张专项表，每类病史一张'),
    'infection_infos':                      ('Register_Infection', 'rename-text-parse', '老表 InfectionDesc/OtherDesc/Note 是自由文本，当前已在 patient_core_service 中做关键字解析'),
    'vascular_accesses':                    ('Register_VascularAccess', 'rename', '已完成 TableName 切换，字段差异待对齐'),
    'vascular_access_interventions':        ('Register_VascularAccessChange', 'rename', '已完成 TableName 切换'),
    'outcome_records':                      ('Register_OutCome', 'rename', '已完成 TableName 切换'),
    'hospitalizations':                     ('Register_Hospitalization', 'rename', '字段几乎一一对应'),
    'treatment_plans':                      ('Plan_PatientPlan', 'rewrite', '新表 4 列 JSONB（dialysisMode/anticoagulant/parameters/materials） vs 老表扁平多列；重点改造'),
    'orders':                               ('Order_PatientOrder', 'rename-fields', 'patient_core_service.buildActiveOrders 已部分使用；字段别名与状态字段需统一'),
    'prescriptions':                        ('Plan_PatientPrescription + Plan_PatientPrescriptionMaterial', 'rewrite+child', '新表 materials/orderItems 为 JSONB；老库 material 为子表'),
    'adjustment_records':                   ('Plan_PatientPlanPrescriptionAdjustment', 'rename', '字段少，简单改名'),
    'wards':                                ('Schedule_Ward', 'rename', ''),
    'beds':                                 ('Schedule_Bed', 'rename', ''),
    'shifts':                               ('Schedule_Shift', 'rename', ''),
    'patient_shifts':                       ('Schedule_PatientShift', 'rename', ''),
    'Treatment_Treatment':                  ('Treatment_Treatment', 'same-name', '同名表，字段需核对'),
    'Treatment_BeforeCheck':                ('Treatment_BeforeCheck', 'same-name', '同名表'),
    'Treatment_BeforeSigns':                ('Treatment_BeforeSigns', 'same-name', '同名表'),
    'Treatment_DuringParam':                ('Treatment_DuringParam', 'same-name', '同名表'),
    'Treatment_AfterSigns':                 ('Treatment_AfterSigns', 'same-name', '同名表'),
    'Treatment_Alarm':                      ('Treatment_Alarm', 'same-name', '同名表'),
    'plan_templates':                       ('Plan_PlanTPL + Plan_PlanTPLMaterial', 'rewrite+child', '新表 templateContent 为 JSONB，老库为父表 + 材料子表'),
    'material_catalogs':                    ('Auxiliary_MaterialInfomation', 'rename', ''),
    'drug_catalogs':                        ('Auxiliary_DrugInfomation', 'rename', ''),
    'order_templates':                      ('Order_OrderTPL', 'rename', '老库可能仅一张表包含模板；需确认是否含子条目'),
    'order_template_items':                 ('Order_OrderTPL (同表)', 'fold-to-parent', '需确认：若老库无子表，items 以 JSON/多行形式存在父表'),
    'dict_types':                           ('CodeDictionary_CodeDictionarys', 'rewrite', '新表为 type/item 两表；老库单表树形（parent + category 分类）'),
    'dict_items':                           ('CodeDictionary_CodeDictionarys', 'rewrite', '同上'),
    'lab_reports':                          ('LIS_Examination', 'rename-fields', '已部分在 patient_core_service.buildLabTrends 使用'),
    'lab_report_items':                     ('LIS_ExaminationItem', 'rename-fields', '同上'),
    'exam_reports':                         (None, 'app-only+sync', '老库无本地表，数据来自 HDIS 同步；保留新表'),
    'patient_key_indicators':               (None, 'app-only+sync', '老库无；来自 HDIS Record 同步；保留新表'),
    'integration_hdis_settings':            (None, 'app-only',    '新系统配置表，与老库无关'),
    # 补充新代码里额外的应用表
    'permissions':                          (None, 'app-only', '权限定义，应用层'),
    'role_permissions':                     (None, 'app-only', '角色-权限关联'),
    'clinical_tasks':                       (None, 'app-only?', '待评估：老库可能无对应；如纯应用层则保留'),
    'devices':                              ('Auxiliary_EquipmentInfomation + Schedule_BedEquipmentRel + Schedule_Bed + Schedule_Ward', 'multi-join', '已由 device_service 通过多表 join 实现；model.TableName 仍为 devices，需修正'),
    'inventory_items':                      ('Stock_Stock + Stock_Storage', 'rewrite', '老库库存拆为 Stock_Stock（总账） + Stock_Storage（仓库）'),
    'stock_logs':                           ('Stock_InOutStorage + Stock_InOutStorageDetail', 'rewrite', '出入库单 + 明细'),
    'label_tasks':                          (None, 'app-only', '条码标签任务，应用层'),
}

# -----------------------------------------------------------------
# 解析新表字段
# -----------------------------------------------------------------
def parse_new_md(path):
    text = open(path, encoding='utf-8').read()
    # 查找所有 #### `xxx` 或 #### xxx — 描述 标题
    pat = re.compile(r'^####\s+`?([A-Za-z_][A-Za-z0-9_]*)`?\s*(?:—|-)?\s*(.*?)$', re.M)
    tables = collections.OrderedDict()
    matches = list(pat.finditer(text))
    for i, m in enumerate(matches):
        name = m.group(1)
        zh = m.group(2).strip()
        start = m.end()
        end = matches[i+1].start() if i+1 < len(matches) else len(text)
        body = text[start:end]
        # 字段表格
        fields = []
        in_table = False
        for line in body.splitlines():
            if not line.strip().startswith('|'):
                if in_table:
                    in_table = False
                continue
            cells = [c.strip() for c in line.strip().strip('|').split('|')]
            if len(cells) < 4:
                continue
            if cells[0] == '字段' and '类型' in cells:
                in_table = True
                continue
            if all(set(c) <= set('-') for c in cells if c):
                continue
            if not in_table:
                continue
            # cells: [字段, 类型, 约束, 说明]
            fields.append({
                'name': cells[0].strip('`'),
                'type': cells[1],
                'constraints': cells[2] if len(cells) > 2 else '',
                'desc': cells[3] if len(cells) > 3 else '',
            })
        if fields:  # 只保留真的有字段的
            tables[name] = {'zh': zh, 'fields': fields}
    return tables

# -----------------------------------------------------------------
# 字段名启发式映射
# -----------------------------------------------------------------
# 新字段（snake_case）→ 老字段（PascalCase）同义词
SYNONYM = {
    'created_at': 'CreateTime',
    'updated_at': 'LastModifyTime',
    'create_time': 'CreateTime',
    'last_modify_time': 'LastModifyTime',
    'tenant_id': 'TenantId',
    'creator_id': 'CreatorId',
    'patient_id': 'PatientId',
    'doctor_id': 'ResponsibilityDrId',     # 主诊医生
    'doctor_name': 'AttendDr',              # 经治医生
    'nurse_name': 'ResponsibilityNurseId',  # 责任护士ID（注意类型差异）
    'is_disabled': 'IsDisabled',
    'notes': 'Note',
    'note': 'Note',
    'name': 'Name',
    'age': '',            # 老库无年龄字段，需通过 BirthDate 计算
    'gender': 'Gender',
    'bed_number': '',     # 老库无冗余字段，通过 BedId → Schedule_Bed.Name
    'diagnosis': 'DiagnosisDesc',  # 在 Register_Diagnosis 表
    'risk_level': '',     # 老库无，需确认
    'status': 'Status',   # 可能需要枚举转换
    'patient_type': 'PatientType',
    'insurance_type': 'ExpenseType',
    'dry_weight': 'PredictWeight',
    'default_mode': 'DialysisMethod',
    'admission_date': 'FirstDialysisDate',
    'discharge_date': '',
    'id_type': 'IDType',
    'id_number': 'IDNo',
    'medical_record_no': 'MedicalRecordNo',
    'insurance_no': 'SSN',
    'dialysis_no': 'DialysisNo',
    'birthday': 'BirthDate',
    'ethnicity': 'Nation',
    'pinyin': 'Spell',
    'abo_blood_type': 'ABOType',
    'rh_blood_type': 'RHType',
    'education_level': 'EducationLevel',
    'occupation': 'Occupation',
    'marital_status': 'MaritalStatus',
    'workplace': 'Workunit',
    'phone': 'PhoneNo',
    'wechat': 'WeChatNo',
    'landline': 'HomePhoneNo',
    'address': 'Address',
    'first_dialysis_date': 'FirstDialysisDate',
    'first_hospital_date': 'OurHospitalFirstDialysisDate',
    'first_dialysis_hospital': 'FirstDialysisHospital',
    'height': 'Height',
    # 排班
    'schedule_date': 'TreatmentTime',
    'ward_id': 'WardId',
    'bed_id': 'BedId',
    'shift_id': 'ShiftId',
    # 治疗方案
    'weekly_frequency': 'OddWeekFrequency',     # 单周
    'biweekly_frequency': 'EvenWeekFrequency',  # 双周
    'duration': 'DialysisDuration',
    # 医嘱
    'type': 'Type',
    'category': 'Classification',
    'content': 'Content',
    'dose': 'Dose',
    'unit': 'Unit',
    'route': 'Route',
    'frequency': 'Frequency',
    'timing': 'Timing',
    'start_time': 'StartTime',
    'end_time': 'EndTime',
    'priority': 'Priority',
    'executed_at': 'ExecuteTime',
    'executed_by': 'ExecutorId',
    'stop_reason': 'StopReason',
    # 检验
    'item_code': 'ItemCode',
    'item_name': 'ItemName',
    'result_value': 'Result',
    'reference_range': 'Reference',
    'abnormal_flag': 'ResultSign',
    'tested_at': 'ResultTime',
    'specimen_type': 'SpecimenType',
    'urgency': 'Urgency',
    'request_doctor': 'RequestDoctor',
    'requested_at': 'RequestTime',
    'sampled_at': 'SampleTime',
    'received_at': 'ReceiveTime',
    'reported_at': 'ReportTime',
    'external_report_id': '',  # 老库无外部 id 字段（因为老库就是数据源）
    'source_system': '',
    'synced_at': '',
}

def snake_to_pascal(s):
    return ''.join(p.title() for p in s.split('_')) if s else s

def map_field(new_name, legacy_cols):
    """给一个新字段名，在老表字段列表里找对应。"""
    if new_name in SYNONYM:
        guess = SYNONYM[new_name]
        if guess == '':
            return None, '老库无对应（需业务决策）'
        # 确认 guess 是否在 legacy_cols 里
        for c in legacy_cols:
            if c['name'] == guess:
                return c['name'], 'synonym'
        # 别名没命中真实列
        return guess, 'synonym-不在目标表（可能跨表/计算得到）'
    pascal = snake_to_pascal(new_name)
    for c in legacy_cols:
        if c['name'] == pascal:
            return c['name'], 'pascal-exact'
    # 忽略大小写/下划线的模糊匹配
    target = new_name.replace('_', '').lower()
    for c in legacy_cols:
        if c['name'].lower() == target:
            return c['name'], 'case-insensitive'
    return None, '无匹配'

# -----------------------------------------------------------------
# 生成
# -----------------------------------------------------------------
def get_primary_legacy(legacy_name_str):
    """从 `Table1 + Table2 + ...` 中取第一个作为主表（fold/join 类映射时）。"""
    if legacy_name_str is None:
        return None
    return legacy_name_str.split(' + ')[0].split(' (')[0].strip()

def render(new_tables, legacy_tables, out_path):
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    lines = []
    lines.append('# 新血透 → 老血透 · 字段级迁移对照表')
    lines.append('')
    lines.append('> 本表为 `docs/migration-plan-legacy.md` 的附录。每张新表列出字段级映射，Codex 据此重写 GORM 模型与 SQL。')
    lines.append('')
    lines.append('## 映射类别')
    lines.append('')
    lines.append('| 类别 | 含义 | 典型操作 |')
    lines.append('|------|------|---------|')
    lines.append('| `rename` | 仅表名不同，字段相近 | 改 `TableName()`、按映射补 `gorm:"column:..."` |')
    lines.append('| `rename-fields` | 表名改且字段名差异较多 | 上 + 逐字段 column tag |')
    lines.append('| `same-name` | 表名相同 | 核对字段类型与命名差异 |')
    lines.append('| `rewrite` | 结构差异大（JSONB ↔ 扁平列） | 重写 service 读写逻辑 + 可能增加领域对象 |')
    lines.append('| `rewrite+child` | 重写 + 拆出子表 | rewrite + 额外处理子表 CRUD |')
    lines.append('| `fold` | 老库由多张表拼成新表 | service 层 join 或多次查询组装 |')
    lines.append('| `split-to-many` | 新表被老库拆成多张专项表 | service 层分散写入/聚合读取 |')
    lines.append('| `multi-join` | 新表通过老库多表 join | 已有实现 / 查询构造器 |')
    lines.append('| `fold-to-parent` | 子表合并到父表（JSON 或多行） | 视情况处理 |')
    lines.append('| `app-only` | 老库无对应，应用层自管 | **不改 TableName，保留新表**（若老库存在该表则不冲突） |')
    lines.append('')

    # 表清单
    lines.append('## 迁移类别汇总')
    lines.append('')
    lines.append('| # | 新表 | 老表 | 类别 | 说明 |')
    lines.append('|---|------|------|------|------|')
    for i, (new_name, (old, cat, note)) in enumerate(TABLE_MAP.items(), 1):
        lines.append(f'| {i} | `{new_name}` | `{old or "—"}` | `{cat}` | {note} |')
    lines.append('')

    # 每张表的字段映射
    for new_name, (old_str, cat, note) in TABLE_MAP.items():
        lines.append(f'## `{new_name}` → `{old_str or "—（应用层保留）"}`')
        lines.append('')
        lines.append(f'**类别：** `{cat}`  ')
        if note:
            lines.append(f'**说明：** {note}')
            lines.append('')

        if cat in ('app-only', 'app-only+sync'):
            lines.append('> 🟢 **此表保留为应用层独立表，无需迁移。**')
            if cat == 'app-only+sync':
                lines.append('> 数据由外部系统（HDIS/LIS）同步写入；老库无本地副本。')
            lines.append('')
            continue

        # 定位老表字段（取第一个主表）
        primary = get_primary_legacy(old_str) if old_str else None
        legacy_cols = legacy_tables.get(primary, [])
        new_fields = new_tables.get(new_name, {}).get('fields', [])
        if not new_fields:
            lines.append(f'> ⚠️ 未在 DATABASE_DESIGN.md 找到 `{new_name}` 字段定义，跳过字段级映射。')
            lines.append('')
            continue

        lines.append(f'**新表字段数：** {len(new_fields)}  **老表（主） `{primary}` 字段数：** {len(legacy_cols)}')
        lines.append('')
        lines.append('| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |')
        lines.append('|--------|--------|--------|--------|----------|------|')
        for f in new_fields:
            mapped, how = map_field(f['name'], legacy_cols)
            legacy_type = ''
            if mapped:
                for c in legacy_cols:
                    if c['name'] == mapped:
                        legacy_type = c['type']
                        break
            mark = ''
            if not mapped:
                mark = '❌ TODO'
            elif '不在目标表' in how:
                mark = '⚠️ 跨表'
            elif how == 'synonym':
                mark = '✅'
            elif how == 'pascal-exact':
                mark = '✅'
            elif how == 'case-insensitive':
                mark = '✅'
            lines.append(f'| `{f["name"]}` | {f["type"]} | {f"`{mapped}`" if mapped else "—"} | {legacy_type} | {how} | {mark} {f["desc"]} |')
        lines.append('')

        # 列出老表里未被映射的字段（在新表里没出现，老库独有）
        used = set()
        for f in new_fields:
            m, _ = map_field(f['name'], legacy_cols)
            if m:
                used.add(m)
        extra = [c for c in legacy_cols if c['name'] not in used]
        if extra:
            lines.append(f'<details><summary>老表 `{primary}` 有 {len(extra)} 个未被新表消费的字段</summary>')
            lines.append('')
            lines.append('| 老字段 | 类型 | 业务含义 |')
            lines.append('|--------|------|----------|')
            for c in extra:
                lines.append(f'| `{c["name"]}` | {c["type"]} | {c["comment"] or "—"} |')
            lines.append('')
            lines.append('</details>')
            lines.append('')

    with open(out_path, 'w', encoding='utf-8') as f:
        f.write('\n'.join(lines))
    return out_path

def main():
    new_tables = parse_new_md(NEW_MD)
    # 复用 struct 解析（合并版 md 结构与 struct.md 相同的表格形式时不可直接用）——
    # 直接读老血透合并版的字段表其实更准，但它格式改了。复用 parse_struct 指向 数据库表结构.md
    legacy_tables = parse_struct('数据库表结构.md')
    print(f'new tables parsed: {len(new_tables)}')
    print(f'legacy tables parsed: {len(legacy_tables)}')
    out = render(new_tables, legacy_tables, 'docs/migration-field-map.md')
    print(f'wrote {out}')

if __name__ == '__main__':
    main()
