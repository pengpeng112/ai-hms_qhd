"""
合并老血透 3 份源 → 老血透数据库表结构-合并版.md

源：
  - 数据库表结构.md     字段权威（类型/主键/非空/默认值）
  - 数据库表设计.md     中文业务含义、枚举、弃用状态
  - 数据库表ER图.docx   【弃用】关系线为绘图对象，文本提取丢失拓扑

输出：./老血透数据库表结构-合并版.md
"""
import re, os, collections, sys

STRUCT_PATH = '数据库表结构.md'
DESIGN_PATH = '数据库表设计.md'
OUT_PATH    = '老血透数据库表结构-合并版.md'

# ---------------------------------------------------------------
# 1) 解析 数据库表结构.md → {PhysicalName: [col rows]}
# ---------------------------------------------------------------
def parse_struct(path):
    text = open(path, encoding='utf-8').read()
    out = collections.OrderedDict()
    headings = re.findall(r'^## \d+\.\d+\.\s+(.+)$', text, flags=re.M)
    bodies = re.split(r'^## \d+\.\d+\.\s+.+$', text, flags=re.M)[1:]
    for name, body in zip(headings, bodies):
        name = name.strip()
        rows = []
        for line in body.splitlines():
            if not line.startswith('|'):
                continue
            cells = [c.strip() for c in line.strip('|').split('|')]
            if len(cells) < 6 or cells[0] in ('字段名', 'Field'):
                continue
            # 跳过分隔行
            if set(cells[0]) <= set('-'):
                continue
            rows.append({
                'name': cells[0],
                'type': cells[1],
                'pk':   cells[2] == 'YES',
                'notnull': cells[3] == 'YES',
                'default': cells[4],
                'comment': cells[5],
            })
        out[name] = rows
    return out

# ---------------------------------------------------------------
# 2) 解析 数据库表设计.md → {CamelName: {'zh': 中文表名, 'fields': {col_name: 中文注释}, 'status': 正常/弃用/未完}}
# ---------------------------------------------------------------
DESIGN_TITLE_RE = re.compile(
    # 三种格式都兼容：
    # ## N．中文 EnglishName
    # ## N. 中文 EnglishName
    # **N．中文 EnglishName**
    # **N.** **中文 EnglishName**
    r'^(?:##\s*|\*{2})\s*(\d+)\s*[．.]\s*(?:\*{2}\s*)?(.+?)(?:\*{2})?\s*$',
    re.M,
)

STATUS_MARKS = {
    '弃用': 'deprecated',
    '尚未完': 'incomplete',
    '未完': 'incomplete',
    '改用': 'replaced',
}

def parse_design(path):
    text = open(path, encoding='utf-8').read()
    out = {}
    # 按标题切分
    positions = []
    for m in DESIGN_TITLE_RE.finditer(text):
        title = m.group(2).strip().strip('*').strip()
        # 跳过"模块或概念英文名称"这类非表标题（没有英文表名的）
        if not re.search(r'[A-Za-z_]{4,}', title):
            continue
        # 行首是字段表格 `| 序号 |` 的不是标题（防误抓）
        positions.append((m.start(), m.end(), int(m.group(1)), title))

    for i, (s, e, num, title) in enumerate(positions):
        # 英文表名：找所有英文 token，取 **最长且以大写开头** 的那个（通常就是表名）
        tokens = re.findall(r'[A-Za-z_][A-Za-z0-9_]{3,}', title)
        # 过滤掉 JsonData 这类频繁出现的普通英文词，只认首字母大写并长度 ≥5 的
        candidates = [t for t in tokens if t[0].isupper() and len(t) >= 5]
        # 常见英文"噪声词"：这些词可能出现在标题的状态描述里
        NOISE = {'JsonData', 'Json'}
        candidates = [t for t in candidates if t not in NOISE] or candidates
        if not candidates:
            continue
        english = max(candidates, key=len)
        # 中文名：去掉所有英文 token 与状态描述
        chinese = re.sub(r'[A-Za-z_][A-Za-z0-9_]*', '', title)
        chinese = re.sub(r'[-—–].*$', '', chinese)
        chinese = chinese.strip('*，,。.（(）) \t')

        status_flags = [v for k, v in STATUS_MARKS.items() if k in title]
        # 切取 body 到下一个标题
        body_end = positions[i+1][0] if i+1 < len(positions) else len(text)
        body = text[e:body_end]

        # 解析 md 表格：| 序号 | 字段 | 类型 | 长度 | 描述 |
        fields = collections.OrderedDict()
        in_table = False
        for line in body.splitlines():
            if not line.lstrip().startswith('|'):
                if in_table:
                    in_table = False
                continue
            cells = [c.strip() for c in line.strip().strip('|').split('|')]
            # 表头
            if '字段' in cells and ('描述' in cells or '含义' in cells):
                in_table = True
                continue
            # 分隔线 | --- | --- |
            if all(set(c) <= set('-') for c in cells if c):
                continue
            if not in_table:
                continue
            if len(cells) < 5:
                continue
            # cells: [序号, 字段, 类型, 长度, 描述]
            fname = cells[1]
            if not fname or not re.match(r'^[A-Za-z_]', fname):
                continue
            desc = cells[4] if len(cells) > 4 else ''
            desc = desc.replace('<br>', ' / ').strip()
            # 同一字段名若出现多次取最后一个非空
            if fname in fields and not desc:
                continue
            fields[fname] = desc

        entry = out.get(english, {'zh': chinese, 'fields': {}, 'status': status_flags, 'title': title, 'seq': num})
        # 合并字段（同一英文名两张表的情况——如 MedicalHistory 新旧版——让后者覆盖/补充）
        entry['fields'].update(fields)
        if status_flags:
            entry['status'] = list(set(entry.get('status', []) + status_flags))
        out[english] = entry
    return out

# ---------------------------------------------------------------
# 3) 表名匹配：struct 用 `Register_PatientInfomation`；design 用 `RegisterPatientInfomation`
#    规则：去下划线 + 大小写无关 + 容忍常见 OCR 错字（1↔l、0↔o、i↔l↔1）
# ---------------------------------------------------------------
def norm(name):
    """规范化名称用于匹配，容错常见 OCR 字形混淆。"""
    s = name.replace('_', '').lower()
    # 1↔l↔i 混淆
    s = s.replace('1', 'l').replace('i', 'l')
    # 0↔o 混淆
    s = s.replace('0', 'o')
    return s

# ---------------------------------------------------------------
# 4) 模块分组
# ---------------------------------------------------------------
MODULE_LABELS = [
    ('Register',          '患者档案 Register'),
    ('Plan',              '治疗方案 Plan'),
    ('Schedule',          '排班 Schedule'),
    ('Order',             '医嘱 Order'),
    ('Treatment',         '治疗记录 Treatment'),
    ('Auxiliary',         '基础数据 / 辅助资料 Auxiliary'),
    ('LIS',               '检验接口 LIS'),
    ('Device',            '设备日志 Device'),
    ('Stock',             '库存 Stock'),
    ('Cost',              '费用 Cost'),
    ('QualityEvaluation', '质控评估 QualityEvaluation'),
    ('Notify',            '通知 Notify'),
    ('Message',           '消息 Message'),
    ('MessageBoard',      '留言板 MessageBoard'),
    ('Log',               '系统日志 Log'),
    ('Applications',      '应用配置 Applications'),
    ('CodeDictionary',    '代码字典 CodeDictionary'),
    ('TenantConfig',      '租户配置 TenantConfig'),
    ('Report',            '报表 Report'),
    ('User',              '用户 User'),
]

def module_of(name):
    for prefix, label in MODULE_LABELS:
        if name.startswith(prefix):
            return prefix, label
    return 'Other', '其他 Other'

# ---------------------------------------------------------------
# 5) 通用字段注释推断（用于 design 未覆盖的字段）
# ---------------------------------------------------------------
GENERIC_COMMENTS = {
    'Id':              '主键 ID',
    'TenantId':        '租户 ID',
    'PatientId':       '患者 ID（外键 → `Register_PatientInfomation.Id`）',
    'CreatorId':       '创建人 ID',
    'CreateTime':      '创建时间',
    'LastModifyTime':  '最后修改时间',
    'IsDisabled':      '是否禁用 / 软删除（false=启用）',
    'Note':            '备注',
    'Name':            '名称',
    'Type':            '类型 / 分类',
    'Status':          '状态（枚举）',
    'ImageBase64String': '图片 Base64 字符串',
    'Sort':            '排序序号',
    'BizId':           '业务对象 ID',
}

def infer_comment(col_name, design_comment):
    if design_comment:
        return design_comment, 'design.md'
    if col_name in GENERIC_COMMENTS:
        return GENERIC_COMMENTS[col_name], '通用约定'
    if col_name.endswith('Id') and col_name != 'Id':
        return f'外键 / 关联 ID（`{col_name}` 待确认指向）', '规则推断 TODO'
    if col_name.endswith('Time') or col_name.endswith('Date'):
        return '时间字段（待补语义）', '规则推断 TODO'
    if col_name.startswith('Is') or col_name.startswith('Has'):
        return '布尔标志（待补语义）', '规则推断 TODO'
    return '—（待人工补注）', 'TODO'

# ---------------------------------------------------------------
# 6) FK 候选：非主键且以 Id 结尾（Id 自身除外，TenantId/CreatorId 等通用字段不计入 FK）
# ---------------------------------------------------------------
GENERIC_ID_FIELDS = {'Id', 'TenantId', 'CreatorId'}
def fk_candidates(cols):
    out = []
    for c in cols:
        n = c['name']
        if n in GENERIC_ID_FIELDS:
            continue
        if c['pk']:
            continue
        if n.endswith('Id'):
            out.append(n)
    return out

# ---------------------------------------------------------------
# 7) 渲染
# ---------------------------------------------------------------
STATUS_BADGE = {
    'deprecated': '⚠️ 弃用',
    'incomplete': '🚧 未完',
    'replaced':   '🔁 被 JsonData 替代',
}

def render(struct, design, out_path):
    # design 的 key 规范化到同一命名空间
    design_by_norm = {norm(k): v for k, v in design.items()}

    grouped = collections.defaultdict(list)
    for t in struct:
        mod, label = module_of(t)
        grouped[(mod, label)].append(t)

    # 保持 MODULE_LABELS 顺序
    lines = []
    lines.append('# 老血透数据库表结构 · 合并版')
    lines.append('')
    lines.append('> 本文件整合老血透系统 3 份源文档：')
    lines.append('>')
    lines.append('> | 来源 | 提供 | 权威性 |')
    lines.append('> |------|------|--------|')
    lines.append('> | `数据库表结构.md` | 字段名、物理类型、主键、非空、默认值 | ✅ 字段定义权威 |')
    lines.append('> | `数据库表设计.md` | 中文业务含义、枚举值、弃用/未完状态 | ✅ 业务注释权威 |')
    lines.append('> | `数据库表ER图.docx` | 理论上是 ER 关系 | ❌ 弃用（关系线为绘图对象，文本提取后拓扑丢失） |')
    lines.append('>')
    lines.append(f'> **统计**：共 {len(struct)} 张表；其中 {sum(1 for t in struct if norm(t) in design_by_norm)} 张在设计 md 中有业务注释。')
    lines.append('>')
    lines.append('> **使用方式**：后端代码从新血透模型迁移到老血透库时，以本文档为字段映射源。表名差异说明：')
    lines.append('>   - 物理表名带下划线（`Register_PatientInfomation`）——用于 GORM TableName 与 SQL')
    lines.append('>   - 设计文档命名去下划线（`RegisterPatientInfomation`）——仅文档引用，代码侧不用')
    lines.append('>')
    lines.append('> **字段业务含义列的"来源"标注**：')
    lines.append('>   - `design.md` — 来自设计文档原文')
    lines.append('>   - `通用约定` — 通用字段（Id/TenantId/CreateTime 等）按系统规范推断')
    lines.append('>   - `规则推断 TODO` — 按字段名启发式推断，**需人工确认**')
    lines.append('>   - `TODO` — 无法推断，**需人工补注**')
    lines.append('')

    # 目录
    lines.append('## 目录')
    lines.append('')
    for prefix, label in MODULE_LABELS + [('Other', '其他 Other')]:
        tbls = grouped.get((prefix, label), [])
        if not tbls:
            continue
        anchor = re.sub(r'[^a-z0-9]+', '-', label.lower()).strip('-')
        lines.append(f'- [{label}（{len(tbls)} 张）](#{anchor})')
    lines.append('- [附录 A：设计 md 中出现但 md 结构未收录的表](#附录-a设计-md-中出现但-md-结构未收录的表)')
    lines.append('- [附录 B：通用字段约定](#附录-b通用字段约定)')
    lines.append('- [附录 C：字段业务注释覆盖率](#附录-c字段业务注释覆盖率)')
    lines.append('')

    # 模块展开
    all_prefixes = [p for p, _ in MODULE_LABELS] + ['Other']
    for prefix in all_prefixes:
        label = dict(MODULE_LABELS + [('Other', '其他 Other')])[prefix]
        tbls = grouped.get((prefix, label), [])
        if not tbls:
            continue
        lines.append(f'## {label}')
        lines.append('')

        for t in tbls:
            cols = struct[t]
            design_entry = design_by_norm.get(norm(t), {})
            design_fields = design_entry.get('fields', {})
            status_flags = design_entry.get('status', [])
            zh_name = design_entry.get('zh', '')

            # 标题
            title_parts = [f'`{t}`']
            if zh_name:
                title_parts.append(f'— {zh_name}')
            if status_flags:
                badges = ' '.join(STATUS_BADGE.get(s, s) for s in status_flags)
                title_parts.append(badges)
            lines.append(f'### ' + ' '.join(title_parts))
            lines.append('')
            pk_cols = [c['name'] for c in cols if c['pk']]
            fks = fk_candidates(cols)
            lines.append(f'- 字段数：{len(cols)}')
            if pk_cols:
                lines.append(f'- 主键：`{", ".join(pk_cols)}`')
            if fks:
                lines.append(f'- 外键候选（字段名启发）：`{", ".join(fks)}`')
            lines.append(f'- 业务注释来源：{"设计 md ✅" if design_entry else "无，按规则推断 ⚠️"}')
            lines.append('')
            lines.append('| # | 字段 | 类型 | PK | NN | 默认值 | 业务含义 | 来源 |')
            lines.append('|---|------|------|----|----|--------|----------|------|')
            for i, c in enumerate(cols, 1):
                pk = '✓' if c['pk'] else ''
                nn = '✓' if c['notnull'] else ''
                default = c['default'] or ''
                design_cmt = design_fields.get(c['name'], '') or c['comment']
                cmt, src = infer_comment(c['name'], design_cmt)
                # md 转义 | 符号
                cmt_safe = cmt.replace('|', '\\|')
                lines.append(f'| {i} | `{c["name"]}` | {c["type"]} | {pk} | {nn} | {default} | {cmt_safe} | {src} |')
            lines.append('')

    # 附录 A
    not_in_struct = [k for k in design if norm(k) not in {norm(t) for t in struct}]
    lines.append('## 附录 A：设计 md 中出现但 md 结构未收录的表')
    lines.append('')
    if not_in_struct:
        lines.append('以下表在 `数据库表设计.md` 中有业务描述，但 `数据库表结构.md` 里没有对应表——通常是**老系统规划但未落库**的表。')
        lines.append('')
        for n in not_in_struct:
            entry = design[n]
            zh = entry.get('zh', '')
            lines.append(f'- `{n}` {zh} —— 字段数 {len(entry["fields"])}')
    else:
        lines.append('（无）')
    lines.append('')

    # 附录 B
    lines.append('## 附录 B：通用字段约定')
    lines.append('')
    lines.append('系统内所有表共享以下通用字段（来自多表对照观察）：')
    lines.append('')
    lines.append('| 字段 | 典型类型 | 含义 |')
    lines.append('|------|----------|------|')
    typical = {
        'Id': 'bigint (snowflake)',
        'TenantId': 'bigint',
        'PatientId': 'bigint',
        'CreatorId': 'bigint',
        'CreateTime': 'timestamp',
        'LastModifyTime': 'timestamp',
        'IsDisabled': 'boolean',
        'Note': 'varchar(1024)',
        'Name': 'varchar(64~256)',
        'Type': 'varchar(64)',
        'Status': 'varchar(64) / integer',
        'ImageBase64String': 'text',
        'Sort': 'numeric',
        'BizId': 'bigint',
    }
    for k, v in GENERIC_COMMENTS.items():
        lines.append(f'| `{k}` | {typical.get(k, "—")} | {v} |')
    lines.append('')

    # 附录 C
    covered = sum(1 for t in struct if norm(t) in design_by_norm)
    total_fields = sum(len(cols) for cols in struct.values())
    commented_fields = 0
    todo_fields = 0
    for t, cols in struct.items():
        design_fields = design_by_norm.get(norm(t), {}).get('fields', {})
        for c in cols:
            cmt, src = infer_comment(c['name'], design_fields.get(c['name'], '') or c['comment'])
            if src == 'design.md' or src == '通用约定':
                commented_fields += 1
            else:
                todo_fields += 1
    lines.append('## 附录 C：字段业务注释覆盖率')
    lines.append('')
    lines.append(f'- 总表数：{len(struct)}')
    lines.append(f'- 设计 md 覆盖表数：{covered}（{covered*100/len(struct):.1f}%）')
    lines.append(f'- 总字段数：{total_fields}')
    lines.append(f'- 已注释字段：{commented_fields}（{commented_fields*100/total_fields:.1f}%）')
    lines.append(f'- 待人工补注字段：{todo_fields}（{todo_fields*100/total_fields:.1f}%）')
    lines.append('')

    with open(out_path, 'w', encoding='utf-8') as f:
        f.write('\n'.join(lines))

def main():
    struct = parse_struct(STRUCT_PATH)
    design = parse_design(DESIGN_PATH)
    render(struct, design, OUT_PATH)
    print(f'OK wrote {OUT_PATH}')
    print(f'struct tables: {len(struct)}')
    print(f'design tables: {len(design)}')
    matched = sum(1 for t in struct if norm(t) in {norm(k) for k in design})
    print(f'matched: {matched} ({matched*100/len(struct):.1f}%)')

if __name__ == '__main__':
    main()
