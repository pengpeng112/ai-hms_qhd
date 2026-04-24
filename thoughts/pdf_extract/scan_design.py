import re

text = open('数据库表设计.md', encoding='utf-8').read()
# 匹配所有可能的表标题格式：
# 1) ## N．中文名 EnglishName
# 2) **N．中文名 EnglishName**
# 3) ## N. 中文名 EnglishName
# 分隔符可能是中文 "．" 或半角 "."
pat = re.compile(
    r'^(?:##\s+|\*\*)\s*(\d+)\s*[．.]\s*(.+?)(?:\*\*|\n)',
    re.M,
)
matches = pat.findall(text)
print('Total matches:', len(matches))
for num, title in matches:
    print(f'  {num}. {title.strip()}')
