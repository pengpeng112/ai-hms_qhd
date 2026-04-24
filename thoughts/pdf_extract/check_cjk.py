import os, glob
for p in sorted(glob.glob('thoughts/pdf_extract/db_*.txt')):
    t = open(p, encoding='utf-8', errors='ignore').read()
    cjk = sum(1 for c in t if '\u4e00' <= c <= '\u9fff')
    print(p, 'size=', len(t), 'cjk=', cjk)
    # print first 300 CJK chars
    cjk_chars = ''.join(c for c in t if '\u4e00' <= c <= '\u9fff')
    print('  first-cjk-sample:', cjk_chars[:200])
