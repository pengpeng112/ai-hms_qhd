"""IDH 阈值重标定 · 提取+打分(团队在生产库跑,只读)。

对每个合格治疗的每个血压测点 t:取 t 前 30 行 Device_DMLog + 基本信息 + 当前 BP,
经【与线上同一套】features.build_feature_frame + 模型算 prob;真标签由 t→t+1 算。
导出小 CSV(treatment_id, operate_time, prob, label)交回分析端。

环境:子项目B 的 .venv312 + `pip install psycopg2-binary`。
连库(只读):设环境变量 PG_DSN，如
  PG_DSN="host=10.0.0.5 port=5432 dbname=hdis user=ro password=*** options='-c statement_timeout=0'"
运行:python recal_extract_score.py --tenant 3 --out recal_scored.csv
断点续:已在 out 里的 treatment_id 自动跳过。

⚠️ 关键一致性:线上 Device_DMLog 只有 9 个可用特征列(LogTime/TMP/UFVolume/VenousPressure/
   ArterialPressure/BF/Conductivity/TreatmentTime/UFSetVolume),其余训练列在线上恒为 0(reindex
   兜底)。本脚本同样只取这 9 列 → 重标定出的阈值对【线上实际特征】有效。
"""
from __future__ import annotations
import argparse
import csv
import os
import pickle
import sys

import psycopg2
import psycopg2.extras

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from features import build_feature_frame  # noqa: E402

BASE = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
# 线上可用的 DMLog 列(与 ai-hms loadIDHWindow 一致);键名须等于 features.FLATTEN_COLS。
DMLOG_COLS = ["LogTime", "TMP", "UFVolume", "VenousPressure", "ArterialPressure",
              "BF", "Conductivity", "TreatmentTime", "UFSetVolume"]


def load_model():
    model = pickle.load(open(os.path.join(BASE, "models", "CatBoost_model.pkl"), "rb"))
    pre = pickle.load(open(os.path.join(BASE, "models", "scaler.pkl"), "rb"))
    return model, pre, list(pre.feature_names_in_)


def map_val(s, d):
    if s is None or d is None:
        return None
    return (float(s) + 2.0 * float(d)) / 3.0


def is_idh(sbp, dbp, nsbp, ndbp):
    if nsbp is None:
        return None
    if float(nsbp) < 100:
        return 1
    if sbp is not None and (float(sbp) - float(nsbp)) > 20:
        return 1
    m0, m1 = map_val(sbp, dbp), map_val(nsbp, ndbp)
    if m0 is not None and m1 is not None and (m0 - m1) > 10:
        return 1
    return 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tenant", type=int, default=3)
    ap.add_argument("--out", default="recal_scored.csv")
    ap.add_argument("--limit-treatments", type=int, default=0, help="0=全部;>0 用于 dry-run 小批")
    args = ap.parse_args()

    model, pre, expected = load_model()
    conn = psycopg2.connect(os.environ["PG_DSN"])
    conn.set_session(readonly=True)

    done = set()
    write_header = True
    if os.path.exists(args.out) and os.path.getsize(args.out) > 0:
        with open(args.out, newline="", encoding="utf-8") as f:
            reader = csv.DictReader(f)
            if reader.fieldnames is None:
                # 文件存在但无表头（空文件或残留）→ 重建
                print(f"检测到 {args.out} 无表头,将重新写入。")
            else:
                for row in reader:
                    tid = row.get("treatment_id")
                    t_iso = row.get("operate_time")
                    if tid is None or t_iso is None:
                        continue  # 跳过半行/畸形尾行
                    try:
                        done.add((int(tid), t_iso))
                    except (ValueError, TypeError):
                        continue
                write_header = False
                print(f"续跑:已完成 {len(done)} 行,跳过。")
    elif os.path.exists(args.out):
        # 0 字节空文件 → 清空后重写表头
        os.remove(args.out)

    cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)
    # 合格治疗:同时有 DMLog 和 血压。
    cur.execute("""
        SELECT t."Id" AS tid
        FROM "Treatment_Treatment" t
        WHERE t."TenantId" = %(tn)s
          AND EXISTS (SELECT 1 FROM "Device_DMLog" d WHERE d."TreatmentId"=t."Id" AND d."TenantId"=%(tn)s)
          AND EXISTS (SELECT 1 FROM "Treatment_DuringSigns" s WHERE s."TreatmentId"=t."Id" AND s."TenantId"=%(tn)s AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL)
        ORDER BY t."Id"
    """, {"tn": args.tenant})
    tids = [r["tid"] for r in cur.fetchall()]
    if args.limit_treatments:
        tids = tids[:args.limit_treatments]
    print(f"合格治疗 {len(tids)} 个,开始打分…")

    out_f = open(args.out, "a", newline="", encoding="utf-8")
    writer = csv.writer(out_f)
    if write_header:
        writer.writerow(["treatment_id", "operate_time", "prob", "label"])

    dlcols = ", ".join(f'"{c}"' for c in DMLOG_COLS)
    n_rows = 0
    for i, tid in enumerate(tids):
        # 治疗级别的全量去重不再需要；改为按 (tid, operate_time) 逐行检查。
        _ = i, tid  # 保留变量名用于进度打印
        # 基本信息(一次)。pre-* 取该治疗透前体征;gender/age/dryweight 取患者/方案。
        cur.execute("""
            SELECT b."Weight" AS pre_weight, b."SBP" AS pre_sbp, b."DBP" AS pre_dbp,
                   p."Gender" AS gender, p."BirthDate" AS birth,
                   pl."DryWeight" AS dry, rx."DialysisMethod" AS method
            FROM "Treatment_Treatment" t
            LEFT JOIN LATERAL (SELECT * FROM "Treatment_BeforeSigns" bb
                               WHERE bb."TreatmentId"=t."Id" AND bb."TenantId"=t."TenantId"
                               ORDER BY bb."OperateTime" DESC NULLS LAST LIMIT 1) b ON true
            LEFT JOIN "Register_PatientInfomation" p ON p."Id"=t."PatientId" AND p."TenantId"=t."TenantId"
            LEFT JOIN "Plan_PatientPlan" pl ON pl."PatientId"=t."PatientId" AND pl."TenantId"=t."TenantId" AND COALESCE(pl."IsDisabled",false)=false
            LEFT JOIN LATERAL (SELECT "DialysisMethod" FROM "Plan_PatientPrescription" rr
                               WHERE rr."TreatmentId"=t."Id" AND rr."TenantId"=t."TenantId"
                               ORDER BY rr."CreateTime" DESC LIMIT 1) rx ON true
            WHERE t."Id"=%(tid)s AND t."TenantId"=%(tn)s
        """, {"tid": tid, "tn": args.tenant})
        info = cur.fetchone() or {}
        age = None
        if info.get("birth") is not None:
            from datetime import date
            b = info["birth"]
            age = (date.today() - b.date()).days / 365.25 if hasattr(b, "date") else None
        gender = None
        g = (str(info.get("gender") or "")).strip()
        if g in ("男", "M", "Male", "1"):
            gender = 1
        elif g in ("女", "F", "Female", "0"):
            gender = 0
        basic_base = {"Gender": gender, "Age": age, "DialysisMethod": info.get("method"),
                      "DryWeight": info.get("dry"), "UFQuantity_y": None,
                      "pre-Weight": info.get("pre_weight"), "pre-SBP": info.get("pre_sbp"),
                      "pre-DBP": info.get("pre_dbp")}

        # 该治疗全部 DMLog(升序,滤未来时间戳),内存切窗。
        cur.execute(f"""
            SELECT {dlcols} FROM "Device_DMLog"
            WHERE "TreatmentId"=%(tid)s AND "TenantId"=%(tn)s AND "LogTime" <= now()
            ORDER BY "LogTime" ASC
        """, {"tid": tid, "tn": args.tenant})
        dmlog = cur.fetchall()
        if not dmlog:
            continue

        # 该治疗全部血压(升序),构造 t→t+1 对。
        cur.execute("""
            SELECT "OperateTime" AS t, "SBP" AS sbp, "DBP" AS dbp FROM "Treatment_DuringSigns"
            WHERE "TreatmentId"=%(tid)s AND "TenantId"=%(tn)s AND "SBP" IS NOT NULL AND "DBP" IS NOT NULL
            ORDER BY "OperateTime" ASC
        """, {"tid": tid, "tn": args.tenant})
        bps = cur.fetchall()

        for k in range(len(bps) - 1):
            cur_bp, nxt = bps[k], bps[k + 1]
            label = is_idh(cur_bp["sbp"], cur_bp["dbp"], nxt["sbp"], nxt["dbp"])
            if label is None:
                continue
            t = cur_bp["t"]
            t_iso = t.isoformat()
            if (tid, t_iso) in done:
                continue
            window = [{c: row[c] for c in DMLOG_COLS} for row in dmlog if row["LogTime"] <= t][-30:]
            if not window:
                continue
            basic = dict(basic_base, SBP=cur_bp["sbp"], DBP=cur_bp["dbp"])
            frame = build_feature_frame(window, basic, expected)
            prob = float(model.predict_proba(pre.transform(frame))[0][1])
            writer.writerow([tid, t_iso, round(prob, 6), label])
            n_rows += 1

        if (i + 1) % 50 == 0:
            out_f.flush()
            print(f"  {i+1}/{len(tids)} 治疗, 累计 {n_rows} 行")

    out_f.close()
    conn.close()
    print(f"完成。写出 {n_rows} 行 → {args.out}")


if __name__ == "__main__":
    main()
