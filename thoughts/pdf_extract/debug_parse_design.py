import re, sys
sys.path.insert(0, 'thoughts/pdf_extract')
from build_legacy_merge import parse_design

design = parse_design('数据库表设计.md')
# Look for treatment tables
for k in sorted(design):
    if 'Treatment' in k or 'BeforeCheck' in k or 'BeforeSymptom' in k:
        e = design[k]
        print(k, '| zh=', e.get('zh'), '| fields=', len(e.get('fields', {})), '| status=', e.get('status'))
