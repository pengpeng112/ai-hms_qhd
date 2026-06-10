#!/usr/bin/env python3
"""CI 编码闸门脚本：检测仓库中的非 UTF-8 文件与 GBK 乱码（双重编码）。

双重编码模式：GBK 字节被当作 Latin-1/Windows-1252 解读后以 UTF-8 存储。
典型症状：文件内容含 Unicode 私用区（PUA）字符（U+E000-U+F8FF）或外观为
"閰嶇疆" 的乱码（实为 "配置" 的 GBK 字节被错误 encode->decode）。

用法：
    python check_encoding.py [<目录>]

返回码：
    0 = 全部通过
    1 = 发现编码问题

配合 .editorconfig 确保新文件以 UTF-8 创建保存。
"""
import os
import re
import sys

PUA = re.compile(r"[\uE000-\uF8FF]")


def fixable(s: str):
    """尝试将外观为乱码的字符串逆向还原为正确中文。"""
    if not any("一" <= c <= "鿿" or "\u4e00" <= c <= "\u9fff" for c in s):
        return None
    for enc in ("gbk", "gb18030"):
        try:
            r = s.encode(enc).decode("utf-8")
            if r != s and any("一" <= c <= "\u9fff" for c in r):
                return r
        except (UnicodeEncodeError, UnicodeDecodeError):
            pass
    return None


def check(path: str) -> list[str]:
    issues = []
    try:
        with open(path, "rb") as f:
            raw = f.read()
    except Exception as e:
        return [f"{path}: 无法读取 ({e})"]

    # 检测 BOM
    if raw.startswith(b"\xef\xbb\xbf"):
        issues.append(f"{path}: 含 UTF-8 BOM（应移除以保持编码纯净）")

    # 检测非 UTF-8
    try:
        text = raw.decode("utf-8-sig")
    except UnicodeDecodeError:
        issues.append(f"{path}: 非 UTF-8 编码")
        return issues

    # 检测 PUA 字符（GBK 双重编码典型特征）
    for lineno, line in enumerate(text.split("\n"), 1):
        if PUA.search(line):
            issues.append(f"{path}:{lineno}: 含 Unicode 私用区(PUA)字符，疑似 GBK 双重编码乱码")

    # 检测可修复的乱码模式
    for lineno, line in enumerate(text.split("\n"), 1):
        r = fixable(line)
        if r is not None:
            issues.append(f"{path}:{lineno}: 疑似 GBK 双重编码乱码（可修复为: {r[:40]}）")

    return issues


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    issues_count = 0
    file_count = 0

    for dirpath, dirnames, filenames in os.walk(root):
        # 跳过依赖目录与产物
        dirnames[:] = [d for d in dirnames if d not in {
            "node_modules", "dist", ".git", ".vite", "tmp", "vendor",
            "ai-hms-v1.3-透析执行", "gorm", "old_system",
        }]
        for fn in filenames:
            if not fn.endswith((".ts", ".tsx", ".js", ".jsx", ".go", ".css", ".html", ".json", ".md", ".py")):
                continue
            file_count += 1
            path = os.path.join(dirpath, fn)
            issues = check(path)
            if issues:
                issues_count += len(issues)
                for issue in issues:
                    print(issue, file=sys.stderr)

    print(f"已扫描 {file_count} 个文件，发现 {issues_count} 个编码问题", file=sys.stderr)
    sys.exit(1 if issues_count > 0 else 0)


if __name__ == "__main__":
    main()
