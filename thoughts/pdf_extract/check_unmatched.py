import re, sys
sys.path.insert(0, 'thoughts/pdf_extract')
from build_legacy_merge import parse_struct, parse_design, norm

struct = parse_struct('数据库表结构.md')
design = parse_design('数据库表设计.md')

design_norms = {norm(k): k for k in design}
struct_norms = {norm(k): k for k in struct}

print('=== struct 中没被 design 匹配的 ===')
for t in struct:
    if norm(t) not in design_norms:
        print(' ', t)
print()
print('=== design 中没被 struct 匹配的 ===')
for t in design:
    if norm(t) not in struct_norms:
        print(' ', t)
