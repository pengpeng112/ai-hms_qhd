#!/usr/bin/env python3
"""GBK 乱码就地修复脚本。

适用场景：原本的 UTF-8 中文被当作 GBK 解码后又存为 UTF-8，产生乱码
（如 "閰嶇疆" 实为 "配置"）。本脚本对每一行尝试 gbk/gb18030 逆转还原；
正确的中文会解码失败而被自动跳过，因此对正常文件无副作用。

⚠️ 会原地改写文件。请在干净的 git 工作区运行，改后用 git diff 复核再提交。

用法：
    python fix_encoding.py <文件1> [<文件2> ...]
例：
    python fix_encoding.py ai-hms-frontend/src/services/restClient.ts

配套检测脚本：scripts/check_encoding.py（CI 闸门，发现乱码即失败）。
"""
import re
import sys


def looks_cjk(s: str) -> bool:
    return any("一" <= c <= "鿿" for c in s)


def fixable(s: str):
    if not looks_cjk(s):
        return None
    for enc in ("gbk", "gb18030"):
        try:
            r = s.encode(enc).decode("utf-8")
            if r != s and looks_cjk(r):
                return r
        except (UnicodeEncodeError, UnicodeDecodeError):
            pass
    return None


def fix_file(path: str) -> int:
    raw = open(path, "rb").read()
    bom = raw.startswith(b"\xef\xbb\xbf")
    text = raw.decode("utf-8-sig")
    parts = re.split(r"(\r\n|\r|\n)", text)  # 保留各行原换行符
    n = 0
    for i in range(0, len(parts), 2):
        fx = fixable(parts[i])
        if fx is not None:
            parts[i] = fx
            n += 1
    out = "".join(parts).encode("utf-8")
    if bom:
        out = b"\xef\xbb\xbf" + out
    open(path, "wb").write(out)
    return n


def main():
    if len(sys.argv) < 2:
        sys.stderr.write(__doc__)
        sys.exit(2)
    for path in sys.argv[1:]:
        n = fix_file(path)
        print("%s: 修复 %d 行" % (path, n))


if __name__ == "__main__":
    main()
