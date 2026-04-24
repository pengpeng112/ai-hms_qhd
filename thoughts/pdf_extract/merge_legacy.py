"""
合并老血透数据库 3 份源文档到一个结构化 MD。

输入：
  - 数据库表结构.md          字段级权威（类型/主键/非空/默认值）
  - 数据库表设计.pdf          模块分组与表顺序（中文注释提取失败）
  - 数据库表ER图.pdf          外键多重度（* 标注）

输出：
  - docs/legacy-db-schema.md  结构化合并结果
"""
import re, os, sys, io, collections

ROOT = '.'
MD_PATH = os.path.join(ROOT, '数据库表结构.md')
DESIGN_TXT = os.path.join(ROOT, 'thoughts/pdf_extract/db_design_raw.txt')
ER_TXT = os.path.join(ROOT, 'thoughts/pdf_extract/db_er_raw.txt')
OUT_PATH = os.path.join(ROOT, 'docs/legacy-db-schema.md')

# --------- 1. 解析 md 字段定义 ---------
def parse_md(path):
    text = open(path, encoding='utf-8').read()
    tables = collections.OrderedDict()
    # 每张表以 `## N.M. TableName` 作为分节
    sections = re.split(r'^## \d+\.\d+\.\s+', text, flags=re.M)[1:]
    headings = re.findall(r'^## \d+\.\d+\.\s+(.+)$', text, flags=re.M)
    for name, body in zip(headings, sections):
        name = name.strip()
        rows = []
        for line in body.splitlines():
            line = line.rstrip()
            if not line.startswith('|'):
                continue
            if set(line.replace('|','').replace('-','').strip()) <= set(' '):
                continue
            cells = [c.strip() for c in line.strip('|').split('|')]
            if len(cells) < 6:
                continue
            if cells[0] in ('字段名','Field','字段'):
                continue
            rows.append({
                'name': cells[0],
                'type': cells[1],
                'pk':   cells[2] == 'YES',
                'notnull': cells[3] == 'YES',
                'default': cells[4],
                'comment': cells[5],
            })
        tables[name] = rows
    return tables

# --------- 2. 解析设计 PDF 的表顺序/分组 ---------
def parse_design_order(path):
    """返回 [(seq, table_name)]；设计 PDF 未覆盖的表不在此列。"""
    order = []
    seen = set()
    with open(path, encoding='utf-8', errors='ignore') as f:
        for line in f:
            m = re.match(r'^(\d+)\.\s+([A-Za-z_][\w_ ]*)', line)
            if not m:
                continue
            seq = int(m.group(1))
            name = m.group(2).strip().rstrip('_').replace(' ', '')
            if not name or name in seen:
                continue
            seen.add(name)
            order.append((seq, name))
    return order

# --------- 3. 解析 ER PDF 中 * 前缀的外键候选 ---------
def parse_er_fk(path):
    """从 ER raw.txt 里找出 `*FieldName type` 模式，标注为 FK 候选。
    注意：raw 文本里 * 与字段名会被分到相邻行，这里做宽松匹配。
    返回 set 字段名（全局性，不分表）——仅用于给最终表里的字段打 FK 标。"""
    fks = set()
    with open(path, encoding='utf-8', errors='ignore') as f:
        text = f.read()
    # 模式：独占一行的 "*" 后紧跟 "字段名 类型"
    # raw extraction 常常每 token 一行
    lines = [l.strip() for l in text.splitlines()]
    for i, line in enumerate(lines):
        if line == '*':
            # 向后找第一条像 "XxxId xxx" 的行
            for j in range(i+1, min(i+5, len(lines))):
                m = re.match(r'^([A-Za-z_]\w*Id)\s+', lines[j])
                if m:
                    fks.add(m.group(1))
                    break
    return fks

# --------- 4. 模块前缀分组 ---------
MODULE_LABELS = {
    'Register': '患者档案（Register）',
    'Plan': '治疗计划（Plan）',
    'Order': '医嘱（Order）',
    'Schedule': '排班（Schedule）',
    'Treatment': '治疗记录（Treatment）',
    'Auxiliary': '基础数据 / 辅助资料（Auxiliary）',
    'Applications': '应用配置（Applications）',
    'CodeDictionary': '代码字典（CodeDictionary）',
    'Cost': '费用（Cost）',
    'Device': '设备日志（Device）',
    'LIS': '检验接口（LIS）',
    'Log': '系统日志（Log）',
    'MessageBoard': '留言板（MessageBoard）',
    'Message': '消息（Message）',
    'Notify': '通知（Notify）',
    'QualityEvaluation': '质控评估（QualityEvaluation）',
    'Stock': '库存（Stock）',
    'TenantConfig': '租户配置（TenantConfig）',
    'User': '用户（User）',
    'Report': '报表（Report）',
}

def module_of(table_name):
    for p in sorted(MODULE_LABELS.keys(), key=len, reverse=True):
        if table_name.startswith(p):
            return p
    return 'Other'

# --------- 5. 生成合并 MD ---------
def render(md_tables, design_order, fks, out_path):
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    design_set = {name for _, name in design_order}
    grouped = collections.defaultdict(list)
    for name in md_tables:
        grouped[module_of(name)].append(name)

    order_of_module = [
        'Register', 'Plan', 'Schedule', 'Order', 'Treatment',
        'Auxiliary', 'LIS', 'Device', 'Stock', 'Cost',
        'QualityEvaluation', 'Notify', 'Message', 'MessageBoard',
        'Log', 'Applications', 'CodeDictionary', 'TenantConfig',
        'Report', 'User', 'Other',
    ]

    lines = []
    lines.append('# 老血透数据库表结构（合并版）')
    lines.append('')
    lines.append('> **来源合并**')
    lines.append('> - `数据库表结构.md` — 字段级权威（字段名、类型、主键、非空、默认值）')
    lines.append('> - `数据库表设计.pdf` — 模块分组、表顺序与业务含义（**中文注释因 PDF 字体缺少 ToUnicode CMap，`pdftotext` 提取丢失**，需人工/OCR 回填）')
    lines.append('> - `数据库表ER图.pdf` — 外键多重度标记（`*FieldId` → 指向父表主键）')
    lines.append('>')
    lines.append('> **本文件用途**：作为后端代码对齐老血透数据库的权威参考。后续新血透 `DATABASE_DESIGN.md` 将以此做字段级对照。')
    lines.append('>')
    lines.append(f'> **统计**：{len(md_tables)} 张表；设计 PDF 覆盖 {len(design_set)} 张；ER 图候选外键字段 {len(fks)} 个。')
    lines.append('')
    lines.append('## 模块索引')
    lines.append('')
    for mod in order_of_module:
        if mod not in grouped or not grouped[mod]:
            continue
        label = MODULE_LABELS.get(mod, mod)
        anchor = re.sub(r'[^a-z0-9]+', '-', label.lower()).strip('-')
        lines.append(f'- [{label}（{len(grouped[mod])} 张）](#{anchor})')
    lines.append('')

    # --------- 每个模块 ---------
    for mod in order_of_module:
        tables = grouped.get(mod, [])
        if not tables:
            continue
        label = MODULE_LABELS.get(mod, mod)
        lines.append(f'## {label}')
        lines.append('')
        # 每张表一个子节
        for t in tables:
            cols = md_tables[t]
            in_design = '✅' if t in design_set else '—'
            lines.append(f'### `{t}`')
            lines.append('')
            lines.append(f'- 字段数：{len(cols)}')
            lines.append(f'- 设计文档覆盖：{in_design}')
            pk_cols = [c["name"] for c in cols if c["pk"]]
            if pk_cols:
                lines.append(f'- 主键：`{", ".join(pk_cols)}`')
            # FK 候选（md 字段名在 ER * 集合里）
            fk_hits = [c['name'] for c in cols if c['name'] in fks]
            if fk_hits:
                lines.append(f'- 外键候选（来自 ER 图 `*` 标记）：`{", ".join(fk_hits)}`')
            lines.append('')
            lines.append('| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |')
            lines.append('|---|------|------|------|------|--------|------------------|')
            for i, c in enumerate(cols, 1):
                pk = '✓' if c['pk'] else ''
                nn = '✓' if c['notnull'] else ''
                default = c['default'] or ''
                comment = c['comment'] or ''
                # 对明显命名的字段给出机器推断注释占位
                if not comment:
                    if c['name'] == 'Id':
                        comment = '主键 ID（snowflake / bigint）'
                    elif c['name'] == 'TenantId':
                        comment = '租户 ID'
                    elif c['name'] == 'PatientId':
                        comment = '患者 ID → `Register_PatientInfomation.Id`'
                    elif c['name'] == 'CreatorId':
                        comment = '创建人 ID'
                    elif c['name'] == 'CreateTime':
                        comment = '创建时间'
                    elif c['name'] == 'LastModifyTime':
                        comment = '最近修改时间'
                    elif c['name'] == 'IsDisabled':
                        comment = '是否禁用/软删除'
                    elif c['name'] == 'Note':
                        comment = '备注'
                    elif c['name'] == 'ImageBase64String':
                        comment = '图像 Base64'
                    elif c['name'].endswith('Time'):
                        comment = '时间字段（待补语义）'
                    elif c['name'].endswith('Id'):
                        comment = 'ID 外键（待核对）'
                    else:
                        comment = '— 待补'
                lines.append(f'| {i} | `{c["name"]}` | {c["type"]} | {pk} | {nn} | {default} | {comment} |')
            lines.append('')

    # --------- 附录：设计 PDF 有但 md 里没有的表 ---------
    extra = [name for _, name in design_order if name not in md_tables]
    lines.append('## 附录 A：设计 PDF 覆盖但 md 未收录的表')
    lines.append('')
    if extra:
        for n in extra:
            lines.append(f'- `{n}`（仅见于设计 PDF，字段需从 PDF 人工提取）')
    else:
        lines.append('（无）')
    lines.append('')

    # --------- 附录：md 有但 design PDF 未覆盖 ---------
    only_md = [n for n in md_tables if n not in design_set]
    lines.append('## 附录 B：md 收录但设计 PDF 未覆盖的表')
    lines.append('')
    if only_md:
        for n in only_md:
            lines.append(f'- `{n}`')
    else:
        lines.append('（无）')
    lines.append('')

    lines.append('## 附录 C：中文业务注释回填建议')
    lines.append('')
    lines.append('PDF 中文字形未嵌入 ToUnicode CMap，`pdftotext` 提取后字段注释列全部为空。建议回填方式（任选其一）：')
    lines.append('')
    lines.append('1. 使用 OCR（`tesseract` + 中文简体模型）对 `数据库表设计.pdf` 每页扫描，结合字段序号回填。')
    lines.append('2. 将 PDF 导出为图片后人工补注。')
    lines.append('3. 对接老系统仓库或字典表（如 `CodeDictionary_CodeDictionarys`）反查业务枚举值含义。')
    lines.append('')

    with open(out_path, 'w', encoding='utf-8') as f:
        f.write('\n'.join(lines))
    return out_path

def main():
    md = parse_md(MD_PATH)
    design = parse_design_order(DESIGN_TXT)
    fks = parse_er_fk(ER_TXT)
    out = render(md, design, fks, OUT_PATH)
    print(f'OK wrote {out}')
    print(f'md tables: {len(md)}')
    print(f'design PDF tables parsed: {len(design)}')
    print(f'ER * FK candidates: {len(fks)}')

if __name__ == '__main__':
    main()
