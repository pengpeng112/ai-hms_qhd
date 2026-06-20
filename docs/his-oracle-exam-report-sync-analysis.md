# HIS Oracle 妫€鏌ユ姤鍛婂悓姝ュ垎鏋愪笌寮€鍙戞柟妗?
> 鐢熸垚鏃堕棿锛?026-06-18
> 鐩爣锛氬垎鏋?HIS Oracle 妫€鏌ユ姤鍛婁笁琛ㄤ笌鏈郴缁熸鏌ユ姤鍛婃ā鍨嬬殑鏄犲皠鍏崇郴锛屽苟缁欏悗缁?AI/寮€鍙戜汉鍛樺鏍镐笌瀹炵幇浣跨敤銆?
## 1. 鐢ㄦ埛鎻愪緵鐨勮繛鎺ヤ俊鎭?
### HIS Oracle

- 鍦板潃锛歚10.10.8.216:1521/orcl`
- 鐢ㄦ埛鍚嶏細`HIS`
- 瀵嗙爜锛氬凡鐢辩敤鎴锋彁渚涳紝鏈枃妗ｄ笉鍐欏叆鏄庢枃瀵嗙爜锛岄伩鍏嶈繘鍏?Git 鍘嗗彶銆?- 涓昏琛細
  - `HIS.EXAM_MASTER a`
  - `HIS.EXAM_REPORT b`
  - `HIS.EXAM_ITEMS c`
- 鍏宠仈 SQL锛?
```sql
SELECT *
FROM his.EXAM_MASTER a
JOIN his.EXAM_REPORT b ON a.exam_no = b.exam_no
JOIN his.EXAM_ITEMS c ON c.exam_no = a.exam_no;
```

### 璺虫澘鏈?鏈嶅姟鍣?
- 鍦板潃锛歚10.10.8.53:40022`
- 鐢ㄦ埛鍚嶏細`root`
- 瀵嗙爜锛氬凡鐢辩敤鎴锋彁渚涳紝鏈枃妗ｄ笉鍐欏叆鏄庢枃瀵嗙爜銆?
## 2. 褰撳墠杩炴帴楠岃瘉缁撴灉

宸蹭粠鏈満楠岃瘉缃戠粶杩為€氭€э細

- `10.10.8.216:1521`锛歍CP 鍙揪銆?- `10.10.8.53:40022`锛歍CP 鍙揪銆?
宸插皾璇曟湰鏈鸿繛鎺?Oracle锛?
- 鏈満鏃?`sqlplus`銆?- 鏈満鏃?Oracle Instant Client銆?- Python `oracledb` 鍖呭彲鐢紝浣?thin 妯″紡杩炴帴澶辫触锛?
```text
DPY-6005: cannot connect to database
DPY-4011: the database or network closed the connection
```

宸插皾璇曢€氳繃璺虫澘鏈烘鏌ョ幆澧冿細

- 璺虫澘鏈哄彲閫氳繃 SSH key 鐧诲綍锛岀閽ヨ矾寰勭敱閮ㄧ讲鐜鎻愪緵锛屾湰鏂囦笉璁板綍鍏蜂綋鏂囦欢鍚嶃€?- 璺虫澘鏈烘棤 `sqlplus`銆?- 璺虫澘鏈烘棤 `tnsping`銆?- 璺虫澘鏈?Python 鍙敤銆?- 璺虫澘鏈哄瓨鍦?Oracle Instant Client锛歚/opt/oracle/instantclient_21`銆?- 璺虫澘鏈?Python `oracledb` thick mode 鍙垚鍔熻繛鎺?HIS Oracle銆?
缁撹锛?
- 鏈満缃戠粶灞傚彲杈撅紝浣嗘湰鏈?thin 妯″紡鏃犳硶杩炴帴 Oracle銆?- 姝ｇ‘閾捐矾涓猴細鏈満 Windows 鈫?SSH 璺虫澘鏈?`10.10.8.53:40022` 鈫?HIS Oracle `10.10.8.216:1521/orcl`銆?- 鍚庣画寮€鍙?璋冭瘯寤鸿鍦ㄨ烦鏉挎満鎴栧悓缃戞閮ㄧ讲鏈鸿繍琛屽悓姝ョ▼搴忋€?- 閫氳繃璺虫澘鏈哄凡鎴愬姛璇诲彇涓夊紶 HIS 妫€鏌ユ姤鍛婅〃鐨勫瓧娈靛拰缁熻淇℃伅銆?
## 2.1 璺虫澘鏈?Python 杩炴帴鏍蜂緥

娉ㄦ剰锛氬瘑鐮佷笉搴斿啓鍏?Git銆備互涓嬬ず渚嬪缓璁€氳繃鐜鍙橀噺娉ㄥ叆銆?
```python
import os
import oracledb

oracledb.init_oracle_client(
    lib_dir=os.environ.get("ORACLE_CLIENT", "/opt/oracle/instantclient_21")
)

conn = oracledb.connect(
    user=os.environ["HIS_ORACLE_USER"],
    password=os.environ["HIS_ORACLE_PASSWORD"],
    dsn=os.environ.get("HIS_ORACLE_DSN", "10.10.8.216:1521/orcl"),
)
```

## 2.2 HIS 涓夎〃瀹為檯缁熻

| 琛?| 琛屾暟 | 鍞竴 `EXAM_NO` 鏁?| 璇存槑 |
|---|---:|---:|---|
| `HIS.EXAM_MASTER` | 3,393,813 | 3,393,813 | 妫€鏌ョ敵璇?涓昏褰曡〃锛宍EXAM_NO` 鍞竴涓旈潪绌恒€?|
| `HIS.EXAM_REPORT` | 1,321,784 | 1,321,784 | 妫€鏌ユ姤鍛婄粨鏋滆〃锛宍EXAM_NO` 鍞竴涓旈潪绌恒€?|
| `HIS.EXAM_ITEMS` | 4,384,921 | 3,835,711 | 妫€鏌ラ」鐩〃锛屼竴浠芥鏌ュ彲瀵瑰簲澶氭潯椤圭洰銆?|
| 涓夎〃鍐呰繛鎺?| 1,411,501 | - | `MASTER JOIN REPORT JOIN ITEMS` 鍚庣殑鏄庣粏琛屾暟銆?|

鏃堕棿鑼冨洿锛?
| 瀛楁 | 鏈€灏忓€?| 鏈€澶у€?| 澶囨敞 |
|---|---|---|---|
| `EXAM_MASTER.REQ_DATE_TIME` | `2014-01-01 01:52:49` | `2026-06-18 22:44:00` | 鐢宠鏃堕棿锛岃鐩栨渶瀹屾暣锛岄€傚悎浣滀负鍏滃簳鏃堕棿銆?|
| `EXAM_MASTER.EXAM_DATE_TIME` | `0001-01-01 00:00:00` | `2026-06-18 22:55:01` | 妫€鏌ユ椂闂达紝瀛樺湪鏃犳晥闆舵棩鏈熴€?|
| `EXAM_MASTER.REPORT_DATE_TIME` | `0001-01-01 00:00:00` | `2026-06-18 20:27:45` | 鎶ュ憡鏃堕棿锛屽瓨鍦ㄦ棤鏁堥浂鏃ユ湡銆?|
| `EXAM_REPORT.REPORT_TIME` | `0001-01-01 00:00:00` | `2026-06-18 22:44:35` | 鎶ュ憡缁撴灉鏃堕棿锛屽瓨鍦ㄦ棤鏁堥浂鏃ユ湡銆?|
| `EXAM_REPORT.CREATEDATE` | `2025-04-25 14:34:53` | `2026-06-18 22:47:33` | 缁撴灉鍒涘缓鏃堕棿锛岄€傚悎鍋氳繎鏈熷閲忥紝浣嗗巻鍙蹭笉瀹屾暣銆?|

鐘舵€佸垎甯冿細

| 瀛楁 | 涓昏鍊?| 鏁伴噺 | 澶囨敞 |
|---|---|---:|---|
| `EXAM_MASTER.RESULT_STATUS` | `1` | 496 | 鏀跺埌鐢宠銆?|
| `EXAM_MASTER.RESULT_STATUS` | `2` | 2,108,986 | 宸叉墽琛屻€?|
| `EXAM_MASTER.RESULT_STATUS` | `3` | 255,769 | 鍒濇鎶ュ憡銆?|
| `EXAM_MASTER.RESULT_STATUS` | `4` | 1,026,608 | 纭鎶ュ憡銆?|
| `EXAM_MASTER.RESULT_STATUS` | `5` | 2,452 | PACS鍙栨秷銆?|
| `EXAM_MASTER.RESULT_STATUS` | `9` | - | 鍏朵粬锛堝綋鍓嶇粺璁℃湭瑙侊級銆?|
| `EXAM_REPORT.IS_ABNORMAL` | `1` | - | 闃虫€э紝鍗虫鏌ュ彲鑳芥湁鐥呭彉锛堝綋鍓嶇粺璁℃湭瑙侊級銆?|
| `EXAM_REPORT.IS_ABNORMAL` | `2` | 1,221,653 | 闈?1锛屾寜闃存€у鐞嗐€?|
| `EXAM_REPORT.IS_ABNORMAL` | `NULL` | 100,570 | 闈?1锛屾寜闃存€?鏈爣璇嗗鐞嗐€?|

瀛楃闆嗚鏄庯細

- 瀛楁鍚嶃€佺被鍨嬨€佹椂闂村拰鏁伴噺鍙甯歌鍙栥€?- 閮ㄥ垎涓枃鏋氫妇鍊?瀛楁娉ㄩ噴缁?SSH + Python 杈撳嚭鏃舵樉绀轰贡鐮侊紱灏濊瘯 `.AL32UTF8` 涓?`ZHS16GBK` 鍚庝粛鏃犳硶鍙潬杩樺師銆?- 涓嶅奖鍝嶅瓧娈垫槧灏勮璁★紱涓枃鏋氫妇鍚箟闇€瑕?HIS/DBA 琛ュ厖纭銆?
## 2.3 HIS 涓夎〃瀹為檯瀛楁

### `HIS.EXAM_MASTER`

| 搴忓彿 | 瀛楁 | 绫诲瀷 | 闀垮害/绮惧害 | 鍙┖ | 鍒濇鍚箟 |
|---:|---|---|---|---|---|
| 1 | `EXAM_NO` | `VARCHAR2` | 10 | N | 妫€鏌ュ彿/鎶ュ憡涓婚敭锛屽敮涓€銆?|
| 2 | `LOCAL_ID_CLASS` | `VARCHAR2` | 1 | Y | 鏈湴 ID 绫诲瀷銆?|
| 3 | `PATIENT_LOCAL_ID` | `VARCHAR2` | 10 | Y | 鎮ｈ€呮湰鍦版爣璇嗐€?|
| 4 | `PATIENT_ID` | `VARCHAR2` | 10 | Y | HIS 鎮ｈ€呮爣璇嗐€?|
| 5 | `VISIT_ID` | `NUMBER` | 2,0 | Y | 灏辫瘖/浣忛櫌鏍囪瘑銆?|
| 6 | `NAME` | `VARCHAR2` | 70 | Y | 濮撳悕銆?|
| 7 | `SEX` | `VARCHAR2` | 4 | Y | 鎬у埆銆?|
| 8 | `DATE_OF_BIRTH` | `DATE` | - | Y | 鍑虹敓鏃ユ湡銆?|
| 9 | `EXAM_CLASS` | `VARCHAR2` | 10 | Y | 妫€鏌ョ被鍒€?|
| 10 | `EXAM_SUB_CLASS` | `VARCHAR2` | 12 | Y | 妫€鏌ュ瓙绫诲埆銆?|
| 11 | `SPM_RECVED_DATE` | `DATE` | - | Y | 鏍囨湰/璧勬枡鎺ユ敹鏃堕棿銆?|
| 12 | `CLIN_SYMP` | `VARCHAR2` | 1000 | Y | 涓村簥鐥囩姸銆?|
| 13 | `PHYS_SIGN` | `VARCHAR2` | 1000 | Y | 浣撳緛銆?|
| 14 | `RELEVANT_LAB_TEST` | `VARCHAR2` | 200 | Y | 鐩稿叧鍖栭獙銆?|
| 15 | `RELEVANT_DIAG` | `VARCHAR2` | 400 | Y | 鐩稿叧璇婃柇銆?|
| 16 | `CLIN_DIAG` | `VARCHAR2` | 1000 | Y | 涓村簥璇婃柇銆?|
| 17 | `EXAM_MODE` | `VARCHAR2` | 1 | Y | 妫€鏌ユ柟寮忋€?|
| 18 | `EXAM_GROUP` | `VARCHAR2` | 16 | Y | 妫€鏌ョ粍銆?|
| 19 | `DEVICE` | `VARCHAR2` | 20 | Y | 璁惧銆?|
| 20 | `PERFORMED_BY` | `VARCHAR2` | 8 | Y | 鎵ц绉戝/鎵ц鑰呬唬鐮併€?|
| 21 | `PATIENT_SOURCE` | `VARCHAR2` | 1 | Y | 鎮ｈ€呮潵婧愩€?|
| 22 | `FACILITY` | `VARCHAR2` | 20 | Y | 鍖荤枟鏈烘瀯銆?|
| 23 | `REQ_DATE_TIME` | `DATE` | - | Y | 鐢宠鏃堕棿銆?|
| 24 | `REQ_DEPT` | `VARCHAR2` | 8 | Y | 鐢宠绉戝銆?|
| 25 | `REQ_PHYSICIAN` | `VARCHAR2` | 20 | Y | 鐢宠鍖荤敓銆?|
| 26 | `REQ_MEMO` | `VARCHAR2` | 60 | Y | 鐢宠澶囨敞銆?|
| 27 | `SCHEDULED_DATE_TIME` | `DATE` | - | Y | 棰勭害鏃堕棿銆?|
| 28 | `NOTICE` | `VARCHAR2` | 2000 | Y | 娉ㄦ剰浜嬮」銆?|
| 29 | `EXAM_DATE_TIME` | `DATE` | - | Y | 妫€鏌ユ椂闂淬€?|
| 30 | `REPORT_DATE_TIME` | `DATE` | - | Y | 鎶ュ憡鏃堕棿銆?|
| 31 | `TECHNICIAN` | `VARCHAR2` | 20 | Y | 鎶€甯堛€?|
| 32 | `REPORTER` | `VARCHAR2` | 20 | Y | 鎶ュ憡浜恒€?|
| 33 | `RESULT_STATUS` | `VARCHAR2` | 1 | Y | 妫€鏌ョ姸鎬併€?|
| 34 | `COSTS` | `NUMBER` | 8,2 | Y | 鎴愭湰銆?|
| 35 | `CHARGES` | `NUMBER` | 8,2 | Y | 璐圭敤銆?|
| 36 | `CHARGE_INDICATOR` | `NUMBER` | 1,0 | Y | 璁¤垂鏍囪銆?|
| 37 | `CHARGE_TYPE` | `VARCHAR2` | 8 | Y | 璐圭敤绫诲埆銆?|
| 38 | `IDENTITY` | `VARCHAR2` | 10 | Y | 韬唤/璐瑰埆銆?|
| 39 | `RPTSTATUS` | `VARCHAR2` | 50 | Y | 鎶ュ憡鐘舵€侊紝褰撳墠缁熻鍑犱箮鍏?NULL銆?|
| 40 | `PRINT_STATUS` | `VARCHAR2` | 50 | Y | 鎵撳嵃鐘舵€併€?|
| 41 | `EXAM_SUBDEPT` | `VARCHAR2` | 10 | Y | 妫€鏌ヤ簹绉戝銆?|
| 42 | `CONFIRM_DOCT` | `VARCHAR2` | 8 | Y | 纭鍖荤敓銆?|
| 43 | `STUDY_UID` | `VARCHAR2` | 128 | Y | PACS Study UID銆?|
| 44 | `VERIFIER` | `VARCHAR2` | 8 | Y | 瀹℃牳浜恒€?|
| 45 | `EXAM_REASON` | `VARCHAR2` | 200 | Y | 妫€鏌ュ師鍥犮€?|
| 46 | `CONFIRM_DATE_TIME` | `DATE` | - | Y | 纭鏃堕棿銆?|
| 47 | `PHOTO_STATUS` | `VARCHAR2` | 50 | Y | 鍥惧儚鐘舵€併€?|
| 48 | `AUDITING_DOCT` | `VARCHAR2` | 10 | Y | 瀹℃牳鍖荤敓銆?|
| 49 | `AUDITING_DATE_TIME` | `DATE` | - | Y | 瀹℃牳鏃堕棿銆?|
| 50 | `EVALUATE_PASS_FALG` | `VARCHAR2` | 1 | Y | 璇勪环閫氳繃鏍囪锛屽瓧娈靛悕鐤戜技鎷煎啓涓?FALG銆?|
| 51 | `RCPT_NO` | `VARCHAR2` | 20 | Y | 鏀舵嵁鍙枫€?|
| 52 | `DOCTOR_USER` | `VARCHAR2` | 8 | Y | 鍖荤敓宸ュ彿銆?|
| 53 | `DOCTOR_NAME` | `VARCHAR2` | 20 | Y | 鍖荤敓濮撳悕銆?|
| 54 | `ALLERGY` | `VARCHAR2` | 200 | Y | 杩囨晱鍙层€?|
| 55 | `ME_MO` | `VARCHAR2` | 1 | Y | 鍚箟寰呯‘璁ゃ€?|
| 56 | `INSUR_TYPE` | `NUMBER` | 1,0 | Y | 鍖讳繚绫诲瀷銆?|
| 57 | `ORDER_ID` | `NUMBER` | 21,0 | Y | 鍖诲槺 ID銆?|
| 58 | `HEART_UPDATE` | `NUMBER` | 1,0 | Y | 鍚箟寰呯‘璁ゃ€?|

### `HIS.EXAM_REPORT`

| 搴忓彿 | 瀛楁 | 绫诲瀷 | 闀垮害 | 鍙┖ | 鍒濇鍚箟 |
|---:|---|---|---:|---|---|
| 1 | `EXAM_NO` | `VARCHAR2` | 10 | N | 妫€鏌ュ彿锛屽敮涓€銆?|
| 2 | `EXAM_PARA` | `VARCHAR2` | 1000 | Y | 妫€鏌ュ弬鏁般€?|
| 3 | `DESCRIPTION` | `VARCHAR2` | 2000 | Y | 妫€鏌ユ弿杩?鎵€瑙併€?|
| 4 | `IMPRESSION` | `VARCHAR2` | 2000 | Y | 鍗拌薄/缁撹銆?|
| 5 | `RECOMMENDATION` | `VARCHAR2` | 2000 | Y | 寤鸿銆?|
| 6 | `IS_ABNORMAL` | `VARCHAR2` | 1 | Y | 鏄惁寮傚父銆?|
| 7 | `USE_IMAGE` | `VARCHAR2` | 150 | Y | 鍥惧儚寮曠敤/鏄惁浣跨敤鍥惧儚銆?|
| 8 | `MEMO` | `VARCHAR2` | 2000 | Y | 澶囨敞銆?|
| 9 | `TECHNICIAN` | `VARCHAR2` | 15 | Y | 鎶€甯堛€?|
| 10 | `REPORTER` | `VARCHAR2` | 15 | Y | 鎶ュ憡浜恒€?|
| 11 | `REPORT_TIME` | `DATE` | - | Y | 鎶ュ憡鏃堕棿銆?|
| 12 | `EXAM_ITEMS` | `VARCHAR2` | 500 | Y | 妫€鏌ラ」鐩枃鏈€?|
| 13 | `EXAM_DIAG` | `VARCHAR2` | 500 | Y | 妫€鏌ヨ瘖鏂€?|
| 14 | `CREATEDATE` | `DATE` | - | Y | 璁板綍鍒涘缓鏃堕棿銆?|

### `HIS.EXAM_ITEMS`

| 搴忓彿 | 瀛楁 | 绫诲瀷 | 闀垮害/绮惧害 | 鍙┖ | 鍒濇鍚箟 |
|---:|---|---|---|---|---|
| 1 | `EXAM_NO` | `VARCHAR2` | 10 | Y | 妫€鏌ュ彿銆?|
| 2 | `EXAM_ITEM_NO` | `NUMBER` | 2,0 | Y | 椤圭洰搴忓彿銆?|
| 3 | `EXAM_ITEM` | `VARCHAR2` | 100 | Y | 妫€鏌ラ」鐩悕绉般€?|
| 4 | `EXAM_ITEM_CODE` | `VARCHAR2` | 20 | Y | 妫€鏌ラ」鐩紪鐮併€?|
| 5 | `COSTS` | `NUMBER` | 8,2 | Y | 鎴愭湰銆?|
| 6 | `BILLING_INDICATOR` | `NUMBER` | 1,0 | Y | 璁¤垂鏍囪銆?|
| 7 | `ZFBL` | `NUMBER` | 5,2 | Y | 鑷粯姣斾緥銆?|
| 8 | `INSUR_CODE` | `VARCHAR2` | 20 | Y | 鍖讳繚缂栫爜銆?|
| 9 | `INSUR_NAME` | `VARCHAR2` | 100 | Y | 鍖讳繚鍚嶇О銆?|
| 10 | `ORDER_ID` | `NUMBER` | 21,0 | Y | 鍖诲槺 ID銆?|
| 11 | `ORDER_ITEM_ID` | `NUMBER` | 21,0 | Y | 鍖诲槺椤圭洰 ID銆?|
| 12 | `URGENT_SIGN` | `NUMBER` | 1,0 | Y | 鍔犳€ユ爣璁帮紝娉ㄩ噴鏄剧ず 0 鏅€?/ 1 鍔犳€ャ€?|

## 3. 鐜版湁鏈郴缁熸鏌ユ姤鍛婃ā鍨?
褰撳墠绯荤粺宸叉湁妫€鏌ユ姤鍛婃ā鍨嬶細`ai-hms-backend/internal/models/lab_report.go`銆?
鐩爣琛細`exam_reports`

瀛楁锛?
```go
type ExamReport struct {
    ID               string     `json:"id"`
    PatientID        string     `json:"patientId"`
    ExamDate         *time.Time `json:"examDate"`
    Title            string     `json:"title"`
    Conclusion       string     `json:"conclusion"`
    Department       string     `json:"department"`
    ExternalReportID *string    `json:"externalReportId,omitempty"`
    SourceSystem     string     `json:"sourceSystem"`
    SyncedAt         *time.Time `json:"syncedAt,omitempty"`
    CreatedAt        time.Time  `json:"createdAt"`
    UpdatedAt        time.Time  `json:"updatedAt"`
}
```

褰撳墠宸叉湁鍚庣鑳藉姏锛?
- 鏌ヨ锛歚GET /api/v1/patients/:id/exam-reports`
- 鍚屾鍏ュ彛锛歚POST /api/v1/patients/:id/exam-reports/sync`
- 鍚屾鏈嶅姟锛歚ExamReportSyncService`

褰撳墠闃诲锛?
- `ExamReportSyncService.getHDISPatientID()` 浠嶆槸 stub锛?
```go
return 0, errors.New("HDIS鎮ｈ€匢D鑾峰彇鏆備笉鍙敤锛氳€佸簱鏃燞disPatientID瀵瑰簲鍒楋紝patient_basic_infos琛ㄥ凡寮冪敤")
```

鍥犳锛岀幇鏈夋鏌ユ姤鍛婂悓姝ユ棤娉曠湡姝ｆ壘鍒板閮ㄧ郴缁熸偅鑰?ID銆?
## 4. HIS 涓夎〃鍒濇涓氬姟鐞嗚В

鐢ㄦ埛鎻愪緵鐨勫叧鑱斿叧绯昏〃鏄庯細

```text
EXAM_MASTER  1 鈹€鈹€ 1/澶?EXAM_REPORT
EXAM_MASTER  1 鈹€鈹€ 澶?  EXAM_ITEMS
```

鏍稿績鍏宠仈閿細

- `exam_no`

鎮ｈ€呭叧鑱旓細

- `patient_id`
- `visit_id`

涓氬姟鏁版嵁鐩爣锛?
- 妫€鏌ユ姤鍛?- 妫€鏌ユ弿杩?鎵€瑙?- 妫€鏌ヨ瘖鏂?- 妫€鏌ュ悕绉?- 妫€鏌ョ被鍒?- 妫€鏌ョ瀹?鎵ц绉戝
- 妫€鏌ユ椂闂?鎶ュ憡鏃堕棿/鏇存柊鏃堕棿

## 5. 寤鸿瀛楁鏄犲皠

浠ヤ笅鏄犲皠鍩轰簬宸插疄闄呰鍙栫殑 Oracle 瀛楁娓呭崟銆?
| 鏈郴缁熷瓧娈?| 鐩爣琛?| HIS 鍊欓€夋潵婧?| 璇存槑 |
|---|---|---|---|
| `external_report_id` | `exam_reports` | `EXAM_MASTER.EXAM_NO` / `EXAM_REPORT.EXAM_NO` | `EXAM_NO` 鍦?`EXAM_MASTER`銆乣EXAM_REPORT` 涓潎鍞竴锛屽缓璁綔涓哄閮ㄦ姤鍛?ID銆?|
| `patient_id` | `exam_reports` | `EXAM_MASTER.PATIENT_ID + VISIT_ID` 鈫?`external_patient_mappings` 鈫?鏈湴鑰佸簱鎮ｈ€?ID | 涓嶅簲鐩存帴鎶?HIS `PATIENT_ID` 鍐欏叆鏈郴缁?`patient_id`銆?|
| `exam_date` | `exam_reports` | `COALESCE(valid(EXAM_REPORT.REPORT_TIME), valid(EXAM_MASTER.REPORT_DATE_TIME), valid(EXAM_MASTER.EXAM_DATE_TIME), valid(EXAM_MASTER.REQ_DATE_TIME))` | `0001-01-01` 瑙嗕负鏃犳晥鏃堕棿銆?|
| `title` | `exam_reports` | `EXAM_REPORT.EXAM_ITEMS`锛屽厹搴曡仛鍚?`EXAM_ITEMS.EXAM_ITEM`锛屽啀鍏滃簳 `EXAM_MASTER.EXAM_CLASS / EXAM_SUB_CLASS` | 妫€鏌ュ悕绉?鏍囬銆?|
| `department` | `exam_reports` | `EXAM_MASTER.PERFORMED_BY` 鎴?`REQ_DEPT` | 褰撳墠涓轰唬鐮佸瓧娈碉紝鏄惁闇€瑕佸瓧鍏哥炕璇戝緟纭銆?|
| `conclusion` | `exam_reports` | `EXAM_REPORT.DESCRIPTION + IMPRESSION + EXAM_DIAG + RECOMMENDATION + MEMO` | 寤鸿淇濈暀鏍囩鍒嗘锛岄伩鍏嶈涔変涪澶便€?|
| `source_system` | `exam_reports` | 鍥哄畾鍊?| 寤鸿鏂板鏋氫妇 `HIS_ORACLE_EXAM`銆?|
| `synced_at` | `exam_reports` | 褰撳墠鍚屾鏃堕棿锛屾垨 `EXAM_REPORT.CREATEDATE` | HIS 鏃犳槑纭?`LAST_UPDATE_TIME`锛屽閲忕瓥鐣ラ渶璋ㄦ厧銆?|

鎺ㄨ崘 `conclusion` 鎷兼帴鏍煎紡锛?
```text
銆愭鏌ユ墍瑙併€憑DESCRIPTION}
銆愬嵃璞°€憑IMPRESSION}
銆愯瘖鏂€憑EXAM_DIAG}
銆愬缓璁€憑RECOMMENDATION}
銆愬娉ㄣ€憑MEMO}
```

鎺ㄨ崘 `exam_date` 鏈夋晥鏃堕棿鍒ゆ柇锛?
- Oracle 鏃ユ湡涓?`NULL`锛氭棤鏁堛€?- Oracle 鏃ユ湡灏忎簬 `1900-01-01`锛氳涓烘棤鏁堬紝閬垮厤 `0001-01-01` 杩涘叆鏈郴缁熴€?- 浼樺厛鎶ュ憡鏃堕棿锛屽叾娆℃鏌ユ椂闂达紝鏈€鍚庣敵璇锋椂闂淬€?
## 6. 鎮ｈ€呭尮閰嶈璁?
HIS 妫€鏌ユ姤鍛婇€氳繃 `patient_id`銆乣visit_id` 鍏宠仈鎮ｈ€咃紝浣嗘湰绯荤粺鎮ｈ€呬富閿槸鑰佸簱 `Register_PatientInfomation.Id`銆?
寤鸿鏂板缁熶竴澶栭儴鎮ｈ€呮槧灏勮〃锛歚external_patient_mappings`銆?
鐢ㄩ€旓細

- HIS patient_id / visit_id 鍒版湰鍦?legacy patient id 鐨勬槧灏勩€?- 鍚庣画鍚屾妫€楠屻€佹鏌ャ€佸尰鍢便€佽垂鐢ㄣ€佸缓妗ｄ俊鎭瓑閮藉鐢ㄣ€?
寤鸿瀛楁锛?
```sql
CREATE TABLE external_patient_mappings (
    id                  VARCHAR(36) PRIMARY KEY,
    tenant_id            BIGINT NOT NULL,
    legacy_patient_id    BIGINT NOT NULL,
    external_system      VARCHAR(32) NOT NULL,
    external_patient_id  VARCHAR(64) NOT NULL,
    external_visit_id    VARCHAR(64),
    id_no                VARCHAR(64),
    dialysis_no          VARCHAR(64),
    hosp_no              VARCHAR(64),
    case_no              VARCHAR(64),
    outpatient_no         VARCHAR(64),
    patient_name         VARCHAR(128),
    match_status         VARCHAR(32) NOT NULL DEFAULT 'confirmed',
    last_synced_at       TIMESTAMP,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_external_patient_mapping_unique
    ON external_patient_mappings (tenant_id, external_system, external_patient_id, COALESCE(external_visit_id, ''));

CREATE INDEX idx_external_patient_mapping_legacy
    ON external_patient_mappings (tenant_id, legacy_patient_id);
```

娉ㄦ剰锛歅ostgreSQL 琛ㄨ揪寮忕储寮曠殑 `COALESCE` 鍐欐硶闇€ DBA 鏍规嵁瀹為檯鐗堟湰纭锛涗篃鍙互鎷嗘垚鏅€氱储寮曞苟鍦ㄦ湇鍔″眰淇濊瘉鍞竴銆?
## 7. 妫€鏌ユ姤鍛婂瓨鍌ㄨ〃寤鸿

褰撳墠妯″瀷宸叉湁 `exam_reports`锛屼絾椤圭洰绂佹 AutoMigrate锛岄渶 DBA 寤鸿〃銆?
寤鸿 DDL 鑽夋锛?
```sql
CREATE TABLE exam_reports (
    id                  VARCHAR(36) PRIMARY KEY,
    patient_id           VARCHAR(36) NOT NULL,
    exam_date            TIMESTAMP,
    title                VARCHAR(200) NOT NULL,
    conclusion           TEXT,
    department           VARCHAR(100),
    external_report_id   VARCHAR(128),
    source_system        VARCHAR(32) NOT NULL DEFAULT 'HIS_ORACLE_EXAM',
    synced_at            TIMESTAMP,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exam_reports_patient_date
    ON exam_reports (patient_id, exam_date DESC);

CREATE UNIQUE INDEX idx_exam_reports_external_unique
    ON exam_reports (source_system, external_report_id, patient_id);
```

濡傛灉 HIS 涓€浠芥鏌ユ湁澶氭潯椤圭洰鏄庣粏锛屼笖鍓嶇闇€瑕佸睍绀烘槑缁嗭紝寤鸿鏂板 `exam_report_items`锛?
```sql
CREATE TABLE exam_report_items (
    id              VARCHAR(36) PRIMARY KEY,
    exam_report_id  VARCHAR(36) NOT NULL,
    item_name        VARCHAR(200) NOT NULL,
    item_category    VARCHAR(100),
    item_result      TEXT,
    sort_order       INT DEFAULT 0,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exam_report_items_report
    ON exam_report_items (exam_report_id, sort_order);
```

绗竴鏈熷鏋滃彧闇€瑕佲€滄姤鍛婃爣棰?+ 鎻忚堪 + 璇婃柇/缁撹鈥濓紝鍙互鍏堜笉寤?`exam_report_items`锛岀洿鎺ユ妸 `EXAM_ITEMS` 鑱氬悎鍒?`title` 鎴?`conclusion`銆?
## 8. Oracle 鍚屾鏌ヨ SQL 妯℃澘

璇峰鏍?AI 鎴?DBA 鍦?Oracle 鐜鎵ц浠ヤ笅 SQL锛屼互纭鐪熷疄瀛楁銆?
### 8.1 鏌ヨ瀛楁缁撴瀯

```sql
SELECT column_id,
       column_name,
       data_type,
       data_length,
       data_precision,
       data_scale,
       nullable
FROM all_tab_columns
WHERE owner = 'HIS'
  AND table_name IN ('EXAM_MASTER', 'EXAM_REPORT', 'EXAM_ITEMS')
ORDER BY table_name, column_id;
```

### 8.2 鏌ヨ琛屾暟

```sql
SELECT 'EXAM_MASTER' AS table_name, COUNT(*) AS cnt FROM his.EXAM_MASTER
UNION ALL
SELECT 'EXAM_REPORT' AS table_name, COUNT(*) AS cnt FROM his.EXAM_REPORT
UNION ALL
SELECT 'EXAM_ITEMS' AS table_name, COUNT(*) AS cnt FROM his.EXAM_ITEMS;
```

### 8.3 鏌ヨ鏃堕棿鑼冨洿鍊欓€夊瓧娈?
璇锋牴鎹瓧娈靛悕鏇挎崲浠ヤ笅鍊欓€夊垪锛?
```sql
SELECT MIN(report_date_time), MAX(report_date_time), COUNT(*)
FROM his.EXAM_REPORT;

SELECT MIN(req_date_time), MAX(req_date_time), COUNT(*)
FROM his.EXAM_MASTER;
```

### 8.4 鏌ヨ鑴辨晱鏍蜂緥

涓嶈瀵煎嚭濮撳悕銆佽韩浠借瘉銆佺數璇濈瓑鏁忔劅瀛楁銆傚缓璁粎鏌ョ粨鏋勬€у瓧娈碉細

```sql
SELECT a.exam_no,
       a.patient_id,
       a.visit_id,
       a.exam_class,
       a.exam_sub_class,
       c.exam_item,
       b.report_date_time
FROM his.EXAM_MASTER a
JOIN his.EXAM_REPORT b ON a.exam_no = b.exam_no
JOIN his.EXAM_ITEMS c ON c.exam_no = a.exam_no
WHERE ROWNUM <= 20;
```

### 8.5 鎺ㄨ崘鍚屾鏌ヨ妯℃澘

浠ヤ笅妯℃澘鐢ㄤ簬鎷夊彇妫€鏌ユ姤鍛婂ご + 鎶ュ憡鍐呭 + 椤圭洰鑱氬悎銆侽racle 鐗堟湰杈冭€佹椂鍙娇鐢?`LISTAGG`锛涜嫢鍗曟姤鍛婇」鐩繃澶氬鑷?`LISTAGG` 瓒呴暱锛岄渶瑕佹敼涓哄垎鎵规煡 `EXAM_ITEMS` 鎴栦娇鐢?XML 鑱氬悎銆?
```sql
SELECT a.exam_no,
       a.patient_id,
       a.visit_id,
       a.name,
       a.sex,
       a.date_of_birth,
       a.exam_class,
       a.exam_sub_class,
       a.performed_by,
       a.req_dept,
       a.req_physician,
       a.req_date_time,
       a.exam_date_time,
       a.report_date_time,
       a.result_status,
       a.study_uid,
       b.description,
       b.impression,
       b.recommendation,
       b.exam_diag,
       b.exam_items AS report_exam_items,
       b.is_abnormal,
       b.use_image,
       b.memo,
       b.reporter,
       b.report_time,
       b.createdate,
       LISTAGG(c.exam_item, '锛?) WITHIN GROUP (ORDER BY c.exam_item_no) AS item_names
  FROM his.exam_master a
  JOIN his.exam_report b ON a.exam_no = b.exam_no
  LEFT JOIN his.exam_items c ON c.exam_no = a.exam_no
 WHERE b.createdate >= :cursor_time
 GROUP BY a.exam_no,
          a.patient_id,
          a.visit_id,
          a.name,
          a.sex,
          a.date_of_birth,
          a.exam_class,
          a.exam_sub_class,
          a.performed_by,
          a.req_dept,
          a.req_physician,
          a.req_date_time,
          a.exam_date_time,
          a.report_date_time,
          a.result_status,
          a.study_uid,
          b.description,
          b.impression,
          b.recommendation,
          b.exam_diag,
          b.exam_items,
          b.is_abnormal,
          b.use_image,
          b.memo,
          b.reporter,
          b.report_time,
          b.createdate
 ORDER BY b.createdate ASC, a.exam_no ASC;
```

澧為噺瀛楁寤鸿锛?
- 绗竴閫夋嫨锛歚EXAM_REPORT.CREATEDATE`锛屼絾浠呰鐩?`2025-04-25` 涔嬪悗鐨勬暟鎹€?- 鍘嗗彶鍒濆鍖栵細鎸?`EXAM_MASTER.REQ_DATE_TIME` 鎴?`EXAM_MASTER.REPORT_DATE_TIME` 鍒嗘鎷夊彇銆?- 鏃ュ父澧為噺锛氬鏋?HIS 纭 `EXAM_REPORT.CREATEDATE` 琛ㄧず鎶ュ憡鍒涘缓/鏈€鍚庢洿鏂版椂闂达紝鍒欎娇鐢ㄥ畠锛涘惁鍒欓渶 HIS/DBA 鎻愪緵鐪熸鏇存柊鏃堕棿瀛楁銆?
## 9. 鍚屾绋嬪簭璁捐寤鸿

鎺ㄨ崘鏂板鐙珛鍚屾杩涚▼锛?
```text
ai-hms-backend/cmd/his-sync
```

杩愯鏂瑰紡锛?
```bash
his-sync --job his_exam_report --once
his-sync --job his_exam_report --from "2026-01-01 00:00:00" --to "2026-01-02 00:00:00"
his-sync --worker
```

涓嶈鎶?Oracle 瀹氭椂鍚屾鐩存帴濉炶繘 `cmd/server`锛屽師鍥狅細

- Oracle 杩炴帴鎱?鏂繛涓嶅簲褰卞搷 Web API銆?- 鍚屾浠诲姟闇€瑕佺嫭绔嬭秴鏃躲€侀噸璇曘€佹棩蹇楀拰璋冨害銆?- 鍚庣画涓嶅悓浠诲姟棰戠巼涓嶅悓锛岀嫭绔?worker 鏇存槗缁存姢銆?
## 10. 鍚屾妗嗘灦鎵╁睍琛?
涓烘敮鎸佸悗缁悓姝ュ叾浠栫被鍨嬫暟鎹紝寤鸿鏂板锛?
### `sync_job_configs`

```sql
CREATE TABLE sync_job_configs (
    id                VARCHAR(36) PRIMARY KEY,
    job_code          VARCHAR(64) NOT NULL UNIQUE,
    source_system     VARCHAR(32) NOT NULL,
    sync_type         VARCHAR(64) NOT NULL,
    enabled           BOOLEAN NOT NULL DEFAULT false,
    cron_expr         VARCHAR(64),
    interval_seconds  INT,
    batch_size        INT NOT NULL DEFAULT 500,
    timeout_seconds   INT NOT NULL DEFAULT 60,
    max_retry         INT NOT NULL DEFAULT 3,
    cursor_type       VARCHAR(32) NOT NULL DEFAULT 'time',
    cursor_value      VARCHAR(128),
    overwrite_policy  VARCHAR(32) NOT NULL DEFAULT 'fill_empty',
    last_run_at       TIMESTAMP,
    next_run_at       TIMESTAMP,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### `sync_job_runs`

```sql
CREATE TABLE sync_job_runs (
    id             VARCHAR(36) PRIMARY KEY,
    job_code       VARCHAR(64) NOT NULL,
    source_system  VARCHAR(32) NOT NULL,
    sync_type      VARCHAR(64) NOT NULL,
    status         VARCHAR(32) NOT NULL,
    started_at     TIMESTAMP NOT NULL,
    finished_at    TIMESTAMP,
    duration_ms    BIGINT,
    fetched_count  INT NOT NULL DEFAULT 0,
    created_count  INT NOT NULL DEFAULT 0,
    updated_count  INT NOT NULL DEFAULT 0,
    skipped_count  INT NOT NULL DEFAULT 0,
    failed_count   INT NOT NULL DEFAULT 0,
    cursor_before  VARCHAR(128),
    cursor_after   VARCHAR(128),
    error_message  TEXT,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## 11. Go 妯″潡寤鸿

寤鸿鏂板锛?
```text
ai-hms-backend/cmd/his-sync/main.go
ai-hms-backend/internal/integrations/his_oracle/client.go
ai-hms-backend/internal/integrations/his_oracle/exam_report_mapper.go
ai-hms-backend/internal/services/external_patient_mapping_service.go
ai-hms-backend/internal/services/his_exam_report_sync_service.go
ai-hms-backend/internal/services/sync_job_service.go
ai-hms-backend/internal/models/external_patient_mapping.go
ai-hms-backend/internal/models/sync_job.go
```

Oracle 椹卞姩寤鸿锛?
```go
github.com/godror/godror
```

閮ㄧ讲娉ㄦ剰锛?
- 闇€瑕?Oracle Instant Client銆?- 鍚屾杩涚▼閮ㄧ讲鏈哄櫒蹇呴』鑳借闂?`10.10.8.216:1521`銆?- 杩炴帴鍑嵁鏀剧幆澧冨彉閲忔垨涓撶敤閰嶇疆琛紝涓嶅啓鍏?Git銆?
## 12. HIS 妫€鏌ユ姤鍛婂悓姝ヤ吉浠ｇ爜

```go
func (s *HisExamReportSyncer) Run(ctx context.Context, job SyncJobConfig) (*SyncResult, error) {
    cursor := job.CursorValue
    rows := s.oracle.QueryExamReports(ctx, cursor, job.BatchSize)

    for _, row := range rows {
        legacyPatientID, matchErr := s.mapping.ResolveLegacyPatientID(row.PatientID, row.VisitID)
        if matchErr != nil {
            result.Failed++
            continue
        }

        report := mapHisExamToExamReport(row, legacyPatientID)
        createdOrUpdated := s.examRepo.UpsertByExternalID(report)
        result.Add(createdOrUpdated)
    }

    job.CursorValue = rows.MaxLastUpdateTime()
    s.jobs.UpdateCursor(job)
    return result, nil
}
```

## 13. 闇€瑕佺敤鎴?DBA 杩涗竴姝ョ‘璁?
璇风户缁彁渚涙垨璁?DBA 鎵ц绗?8 鑺?SQL 鍚庣粰鍑虹粨鏋溿€傚挨鍏堕渶瑕佺‘璁わ細

1. `EXAM_MASTER`銆乣EXAM_REPORT`銆乣EXAM_ITEMS` 鐨勫畬鏁村瓧娈垫竻鍗曘€?2. `exam_no` 鏄惁鍏ㄥ眬鍞竴銆?3. 鍝釜瀛楁浠ｈ〃妫€鏌ユ姤鍛婃渶缁堟洿鏂版椂闂淬€?4. 鍝釜瀛楁浠ｈ〃鎶ュ憡鐘舵€?浣滃簾鏍囪銆?5. `patient_id + visit_id` 濡備綍瀵瑰簲鏈湴鑰佸簱鎮ｈ€呫€?6. `EXAM_REPORT` 涓€滄弿杩?璇婃柇/缁撹鈥濈殑鐪熷疄瀛楁鍚嶃€?7. `EXAM_ITEMS` 鏄惁涓€浠芥姤鍛婂鏉℃鏌ラ」鐩€?8. 鏄惁闇€瑕佷繚瀛?PACS 鍥惧儚閾炬帴鎴?PDF 鎶ュ憡閾炬帴銆?
## 14. 鍒濇缁撹

- 妫€鏌ユ姤鍛婂姛鑳戒笉闇€瑕佸畬鍏ㄤ粠闆跺紑鍙戯紱鏈郴缁熷凡鏈?`exam_reports` 妯″瀷銆佸垪琛ㄦ帴鍙ｅ拰鍚屾鏈嶅姟楠ㄦ灦銆?- 鐪熸缂哄彛鏄細HIS Oracle 璇诲彇銆佸閮ㄦ偅鑰?ID 鏄犲皠銆丏BA 寤鸿〃銆佸悓姝ヤ换鍔¤皟搴﹀拰杩愯鏃ュ織銆?- 鎺ㄨ崘鍦ㄥ綋鍓嶄粨搴撳唴寮€鍙戠嫭绔?`his-sync` 鍚屾杩涚▼锛學eb API 鍙礋璐ｆ煡璇€佹墜鍔ㄨЕ鍙戝拰閰嶇疆灞曠ず銆?- 褰撳墠鎻愪緵鐨勪笁琛ㄥ叧绯昏冻澶熷惎鍔ㄨ璁★紝浣嗘寮忓紑鍙戝墠蹇呴』纭涓夎〃瀛楁娓呭崟鍜屽閲忔椂闂村瓧娈点€?
---

## 15. 澶嶆牳缁撹锛?026-06-19 瀹為檯杩炴帴 HIS Oracle 楠岃瘉锛?
浠ヤ笅閫氳繃璺虫澘鏈?`oracledb thick mode` 瀹為檯鏌ヨ HIS Oracle 鍚庡緱鍑恒€?
### 15.1 鏁版嵁楠岃瘉缁撴灉

| 楠岃瘉椤?| 缁撴灉 | 缁撹 |
|---|---|---|
| `EXAM_MASTER` 鎬昏鏁?| 3,393,813 | 閲忕骇澶э紝鍏ㄩ噺鍒濆鍖栭渶鍒嗘壒 |
| `EXAM_NO` 鍦?`EXAM_MASTER` 鍞竴鎬?| 3,393,813 distinct = 鎬绘暟 | 鉁?鍞竴 |
| `EXAM_NO` 鍦?`EXAM_REPORT` 鍞竴鎬?| 1,321,784 distinct = 鎬绘暟 | 鉁?鍞竴 |
| `EXAM_NO` 瀵瑰簲澶?`PATIENT_ID` | 0 鏉?| 鉁?涓€浠芥鏌ュ彧灞炰簬涓€涓偅鑰?|
| `EXAM_REPORT` 涓€瀵瑰 `EXAM_NO` | 0 鏉?| 鉁?鎶ュ憡涓庝富琛ㄤ弗鏍?1:1 |
| `EXAM_REPORT` 瀛ゅ効璁板綍锛堟棤 MASTER锛?| 54 鏉?| 鍙拷鐣ワ紝LEFT JOIN 澶勭悊 |
| `EXAM_ITEMS` 涓€浠芥姤鍛婂鏉￠」鐩?| 3,835,711 distinct vs 4,384,921 鎬绘暟 | 鉁?骞冲潎 1.14 鏉?鎶ュ憡 |
| 鏈€杩?7 澶?`CREATEDATE` 澧為噺 | 7,676 鏉?| 杩戞湡澧為噺鍙甯稿伐浣?|
| 鏈€杩?7 澶?`REPORT_DATE_TIME` 澧為噺 | 3,778 鏉?| 姣?CREATEDATE 灏?|
| 鏈€杩?7 澶?`REQ_DATE_TIME` 澧為噺 | 10,538 鏉?| 姣斿墠涓よ€呴兘澶?|
| `CREATEDATE` 鏈€杩?100 鏉℃槸鍚︽湁鍥炲～ | 0 鏉?| 鉁?杩戞湡鍗曡皟閫掑锛屽彲鍋氬閲忔父鏍?|
| `RESULT_STATUS` 涓昏鍒嗗竷 | `2`=2,108,723; `4`=1,026,311; `3`=255,657 | 鐮佸€煎惈涔夊緟 HIS 纭 |

### 15.2 鎮ｈ€呭尮閰嶉樆鏂闄╋紙鏈€楂樹紭鍏堢骇锛?
鏈郴缁熻€佸簱鎮ｈ€?ID 涓?HIS 鎮ｈ€?ID **鏍煎紡瀹屽叏涓嶅悓锛屾棤娉曠洿鎺ュ搴?*锛?
| 绯荤粺 | ID 鑼冨洿/鏍煎紡 | 鏍蜂緥 |
|---|---|---|
| 鏈湴鑰佸簱 `Register_PatientInfomation.Id` | 6 浣嶆暟瀛?`300002`-`300410` | `300002`, `300410` |
| HIS `EXAM_MASTER.PATIENT_ID` | 瀛楃涓诧紝鍚瓧姣嶅墠缂€锛?-10 浣?| `00093448`, `c0554459`, `00201578` |

缁撹锛?- 涓嶈兘鐢?`patient_id` 鐩存帴鍋氫富閿叧鑱斻€?- 蹇呴』閫氳繃韬唤璇佸彿/浣忛櫌鍙?闂ㄨ瘖鍙?閫忔瀽鍙风瓑妗ユ帴閿繘琛屾槧灏勩€?- 闇€纭 HIS 鎮ｈ€呰〃鏄惁瀛樺偍韬唤璇佸彿锛屽綋鍓?`EXAM_MASTER` 涓彧鏈?`PATIENT_ID`銆乣VISIT_ID`銆乣NAME`锛屾棤韬唤璇?浣忛櫌鍙峰瓧娈点€?- 濡傛灉 HIS 鏈夌嫭绔嬬殑鎮ｈ€呬富绱㈠紩琛紙濡?`PATIENT_INFO`銆乣PAT_VISIT`锛夛紝闇€瑕侀澶栬幏鍙栦互寤虹珛鏄犲皠銆?
杩欐槸鏁翠釜鍚屾鏂规鑳藉惁钀藉湴鐨勬渶澶у墠鎻愩€?
### 15.3 澧為噺鍚屾绛栫暐淇

`CREATEDATE` 浠呰鐩?`2025-04-25` 鑷充粖锛屽巻鍙茬害 100 涓囨潯鎶ュ憡鏃?`CREATEDATE`銆?
鎺ㄨ崘鍒嗘绛栫暐锛?
1. **鍘嗗彶鍒濆鍖?*锛氭寜 `EXAM_MASTER.REQ_DATE_TIME`锛堣鐩?2014 鑷充粖锛夊垎娈垫媺鍙栥€?2. **鏃ュ父澧為噺**锛氭寜 `EXAM_REPORT.CREATEDATE` 鎷夊彇锛堟渶杩戞暟鎹畬鏁翠笖鍗曡皟閫掑锛夈€?3. **琛ュ伩**锛氬畾鏃舵寜 `EXAM_MASTER.REPORT_DATE_TIME` 鎴?`REQ_DATE_TIME` 鍋氫簩娆℃牎楠岋紝鎹炲彇 `CREATEDATE` 涓虹┖浣嗚繎鏈熸湁鏇存柊鐨勮褰曘€?
`sync_job_configs.cursor_type` 搴旀敮鎸?`time` 鍜?`mixed` 涓ょ妯″紡銆?
### 15.4 鏂囨。褰撳墠涓嶈冻涔嬪

| # | 涓嶈冻 | 褰卞搷 | 寤鸿 |
|---|---|---|---|
| 1 | **鏈‘璁?HIS 鎮ｈ€呬富绱㈠紩琛?* | 鏃犳硶寤虹珛鎮ｈ€呮槧灏?| 闇€鏌ヨ HIS 鏄惁鏈?`PATIENT_INFO`/`PATIENT`/`PAT_VISIT` 绛夎〃 |
| 2 | **`RESULT_STATUS` 鐮佸€煎凡纭** | 鍙寜鐘舵€佽繃婊ょ‘璁ゆ姤鍛?鍙栨秷鎶ュ憡 | `1=鏀跺埌鐢宠,2=宸叉墽琛?3=鍒濇鎶ュ憡,4=纭鎶ュ憡,5=PACS鍙栨秷,9=鍏朵粬` |
| 3 | **`IS_ABNORMAL` 鐮佸€煎凡纭** | 鍙垽鏂槼鎬?闃存€?| `1=闃虫€э紝鍗虫鏌ュ彲鑳芥湁鐥呭彉锛涘叾浠?闃存€ |
| 4 | **涓枃鏋氫妇鍊间贡鐮?* | 鏃犳硶纭 `EXAM_CLASS`/`EXAM_SUB_CLASS` 鍚箟 | 闇€鍦?Oracle 鏈満鎴?DBA 瀵煎嚭纭 |
| 5 | **`study_uid` 鍙敤鎬?* | `STUDY_UID` 瀛樺湪浣嗗崰姣斿緟纭 | 鍙敤浜?PACS 娣遍摼锛岄渶纭 PACS 璁块棶鏂瑰紡 |
| 6 | **鏈€冭檻妫€鏌ユ洿鏂板満鏅?* | HIS 妫€鏌ユ姤鍛婁慨鏀瑰悗鏃?`LAST_UPDATE_TIME` | `CREATEDATE` 涓嶇瓑浜庢洿鏂版椂闂达紝淇敼鍚庡彲鑳芥紡鍚屾 |
| 7 | **`EXAM_REPORT` 涓?`EXAM_MASTER` 涓嶆槸鍏ㄩ噺瀵归綈** | `EXAM_MASTER` 339 涓?vs `EXAM_REPORT` 132 涓?| 绾?207 涓囨鏌ョ敵璇锋棤鎶ュ憡缁撴灉锛岄渶鍖哄垎"鐢宠"涓?鎶ュ憡" |
| 8 | **`LISTAGG` 鍙兘瓒呴暱** | `EXAM_ITEMS.EXAM_ITEM` 鑱氬悎鍙兘瓒?4000 瀛楃 | 鏀逛负鍒嗘壒鏌?`EXAM_ITEMS` 鎴栭檺鍒?`LISTAGG` 闀垮害 |

---

## 16. 鍓嶇鍚屾閰嶇疆鐣岄潰璁捐鏂规

### 16.1 鐜版湁鍓嶇涓嶈冻

褰撳墠绯荤粺鍓嶇鎯呭喌锛?
| 鍔熻兘 | 鐜扮姸 | 涓嶈冻 |
|---|---|---|
| 妫€鏌ユ姤鍛婂睍绀?| `LabsExamsTab.tsx` 宸叉湁鍒楄〃灞曠ず | 鍙睍绀烘爣棰?鏃ユ湡锛屼笉灞曠ず鎻忚堪/璇婃柇/缁撹 |
| 妫€鏌ユ姤鍛婂悓姝?| 杩涘叆椤甸潰鑷姩璋?`syncExamReports` | 璧?HDIS stub锛屽疄闄呬笉鍙敤锛涙棤 HIS Oracle 鍏ュ彛 |
| HDIS 閰嶇疆 | `Settings.tsx` 鏈夐泦鎴?Tab | 鍙敮鎸?HDIS锛屼笉鏀寔 HIS Oracle 杩炴帴閰嶇疆 |
| 鍚屾浠诲姟绠＄悊 | 鏃?| 鏃犳硶鏌ョ湅/鍚仠/鎵嬪姩瑙﹀彂 HIS 鍚屾浠诲姟 |
| 鍚屾鏃ュ織 | 鏃?| 鏃犳硶鏌ョ湅鍚屾缁撴灉銆侀敊璇€佽€楁椂 |
| 鎮ｈ€呮槧灏勭鐞?| 鏃?| 鏃犳硶鏌ョ湅/缁存姢澶栭儴鎮ｈ€呮槧灏?|

### 16.2 鏀归€犺寖鍥?
寤鸿鏂板涓や釜鍓嶇妯″潡锛?
**A. Settings 椤甸潰鏂板銆孒IS Oracle 鍚屾閰嶇疆銆峊ab**

鍦ㄧ幇鏈?`Settings.tsx` 鐨?Tab 鍒楄〃涓柊澧?`his-oracle` Tab锛?
```
Settings
鈹溾攢鈹€ 绯荤粺璁剧疆
鈹溾攢鈹€ HDIS 闆嗘垚锛堝凡鏈夛級
鈹溾攢鈹€ HIS Oracle 鍚屾锛堟柊澧烇級   鈫?鏈妭
鈹?  鈹溾攢鈹€ 鏁版嵁婧愯繛鎺ラ厤缃?鈹?  鈹溾攢鈹€ 鍚屾浠诲姟鍒楄〃
鈹?  鈹斺攢鈹€ 杩炴帴娴嬭瘯
鈹斺攢鈹€ 绯荤粺鏃ュ織
```

**B. 鎮ｈ€呰鎯呴〉妫€鏌ユ姤鍛婃敼閫?*

鏀归€?`LabsExamsTab.tsx` 妫€鏌ユ姤鍛婂尯鍩燂細

- 灞曞紑鏌ョ湅锛氭弿杩般€佸嵃璞°€佽瘖鏂€佸缓璁€佸娉?- 鏉ユ簮鏍囪锛歚HIS Oracle` / `HDIS` / `鎵嬪姩褰曞叆`
- 鎵嬪姩鍚屾鎸夐挳锛氬崟鐙Е鍙戣鎮ｈ€呮鏌ユ姤鍛婂悓姝?- PACS 閾炬帴锛堝鏋?`study_uid` 鍙敤锛?
**C. 鏂板銆屽悓姝ョ鐞嗕腑蹇冦€嶉〉闈紙鍙€夛紝寤鸿绗簩鏈燂級**

```
鍚屾绠＄悊涓績锛?sync-center锛?鈹溾攢鈹€ 鍚屾浠诲姟鍒楄〃
鈹?  鈹溾攢鈹€ 浠诲姟鐘舵€佸崱鐗囷紙鍚敤/鍋滅敤/涓婃杩愯/涓嬫杩愯/鎴愬姛/澶辫触锛?鈹?  鈹溾攢鈹€ 鎵嬪姩杩愯鎸夐挳
鈹?  鈹斺攢鈹€ 閰嶇疆缂栬緫
鈹溾攢鈹€ 鍚屾杩愯鍘嗗彶
鈹?  鈹溾攢鈹€ 姣忔杩愯鐨勭粺璁★紙鎷夊彇/鏂板/鏇存柊/璺宠繃/澶辫触锛?鈹?  鈹斺攢鈹€ 閿欒璇︽儏灞曞紑
鈹斺攢鈹€ 鏈尮閰嶆偅鑰呴槦鍒?    鈹溾攢鈹€ HIS 鎮ｈ€呭尮閰嶅け璐ョ殑璁板綍
    鈹斺攢鈹€ 浜哄伐缁戝畾鏈湴鎮ｈ€?```

### 16.3 HIS Oracle 杩炴帴閰嶇疆 Tab 璇︾粏璁捐

```
鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? HIS Oracle 鏁版嵁婧愯繛鎺?                           鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? Host          [10.10.8.216        ]            鈹?鈹? Port          [1521               ]            鈹?鈹? Service Name  [orcl               ]            鈹?鈹? Username      [his                ]            鈹?鈹? Password      [********           ]  (涓嶅洖鏄?  鈹?鈹?                                                 鈹?鈹? [娴嬭瘯杩炴帴]  [淇濆瓨]                              鈹?鈹?                                                 鈹?鈹? 杩炴帴鐘舵€? 鉁?宸茶繛鎺?(寤惰繜 12ms)                  鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?
鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 鍚屾浠诲姟                                        鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 浠诲姟            鈹?棰戠巼 鈹?鐘舵€?鈹?涓婃 鈹?鎿嶄綔    鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 妫€鏌ユ姤鍛婂悓姝?   鈹?10m  鈹?鍚敤 鈹?鎴愬姛 鈹?杩愯|閰嶇疆鈹?鈹? 鎮ｈ€呮。妗堝悓姝?   鈹?24h  鈹?鍋滅敤 鈹?-    鈹?閰嶇疆    鈹?鈹? 锛堟湭鏉ワ細妫€楠岋級  鈹?-    鈹?-    鈹?-    鈹?-       鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?```

浠诲姟閰嶇疆寮圭獥锛?
```
鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 閰嶇疆锛氭鏌ユ姤鍛婂悓姝?                鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 鍚敤          [鉁匽                鈹?鈹? 鍚屾棰戠巼      [姣?10 鍒嗛挓 鈻糫      鈹?鈹? 姣忔壒鏁伴噺      [500]               鈹?鈹? 瓒呮椂鏃堕棿      [60 绉抅             鈹?鈹? 鏈€澶ч噸璇?     [3 娆              鈹?鈹? 澧為噺瀛楁      [CREATEDATE 鈻糫      鈹?鈹? 娓告爣鍊?       [2026-06-18 22:..] 鈹?鈹? 瑕嗙洊绛栫暐      [浠呰ˉ绌哄瓧娈?鈻糫      鈹?鈹?                                   鈹?鈹? [淇濆瓨]  [绔嬪嵆杩愯]  [鍙栨秷]        鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?```

### 16.4 鍚屾杩愯鍘嗗彶璁捐

```
鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 妫€鏌ユ姤鍛婂悓姝?- 杩愯鍘嗗彶                              [鍒锋柊]  鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 寮€濮嬫椂闂?   鈹?鑰楁椂   鈹?鎷夊彇 鈹?鏂板 鈹?鏇存柊 鈹?澶辫触 鈹?鐘舵€?    鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 06-18 22:30 鈹?12.3s  鈹?45   鈹?3    鈹?0    鈹?0    鈹?鉁?鎴愬姛   鈹?鈹? 06-18 22:20 鈹?8.1s   鈹?12   鈹?0    鈹?0    鈹?0    鈹?鉁?鎴愬姛   鈹?鈹? 06-18 22:10 鈹?45.2s  鈹?500  鈹?23   鈹?2    鈹?1    鈹?鈿狅笍 閮ㄥ垎  鈹?鈹? 06-18 22:00 鈹?-      鈹?-    鈹?-    鈹?-    鈹?-    鈹?鉂?澶辫触   鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?
澶辫触琛屽睍寮€锛?  閿欒锛歄RA-03113: end-of-file on communication channel
  娓告爣锛?026-06-18 21:50:00 鈫?2026-06-18 22:00:00
  閲嶈瘯锛?/3锛堝凡鑰楀敖锛?```

### 16.5 鏈尮閰嶆偅鑰呴槦鍒楄璁?
```
鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 鏈尮閰嶆偅鑰呴槦鍒?                                     [鍒锋柊]  鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? HIS ID    鈹?濮撳悕      鈹?妫€鏌ユ暟   鈹?鎿嶄綔                     鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹尖攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? c0554459  鈹?寮?*      鈹?12       鈹?[鎼滅储鏈湴] [鏂板缓鎮ｈ€匽    鈹?鈹? 00201578  鈹?鏉?*      鈹?3        鈹?[鎼滅储鏈湴] [鏂板缓鎮ｈ€匽    鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹粹攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?
鐐瑰嚮 [鎼滅储鏈湴] 寮圭獥锛?鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? 缁戝畾鏈湴鎮ｈ€?                     鈹?鈹? HIS 鎮ｈ€? c0554459 / 寮?*         鈹?鈹? 鎼滅储: [寮犱笁          ]  [鎼滅储]    鈹?鈹? 鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹?鈹? 鈹?300012 寮犱笁 鐢?58宀?HD       鈹?鈹?鈹? 鈹?300045 寮犱笁涓?鐢?62宀?HDF    鈹?鈹?鈹? 鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹?鈹? [纭缁戝畾]  [鍙栨秷]               鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?```

### 16.6 鍓嶇鏂板璺敱鍜?API

鏂板璺敱锛?
```text
/sync-center          鈫?SyncCenterPage锛堝悓姝ョ鐞嗕腑蹇冿級
/settings             鈫?Settings锛堝凡鏈夛紝鏂板 HIS Oracle Tab锛?```

鏂板 API 鏂规硶锛坄restClient.ts`锛夛細

```typescript
// HIS Oracle 鏁版嵁婧愰厤缃?getHisOracleConfig()
updateHisOracleConfig(payload)
testHisOracleConnection()

// 鍚屾浠诲姟绠＄悊
getSyncJobs()
updateSyncJob(jobCode, payload)
runSyncJob(jobCode)

// 鍚屾杩愯鍘嗗彶
getSyncJobRuns(jobCode, params)

// 澶栭儴鎮ｈ€呮槧灏?getUnmatchedPatients(params)
bindExternalPatientMapping(externalPatientId, legacyPatientId)
```

### 16.7 鍓嶇鏀归€犱紭鍏堢骇

| 浼樺厛绾?| 鏀归€犻」 | 鐞嗙敱 |
|---|---|---|
| P0 | HIS Oracle 杩炴帴閰嶇疆 Tab | 鏃犺繛鎺ラ厤缃棤娉曡繍琛屽悓姝?|
| P0 | 妫€鏌ユ姤鍛婂垪琛ㄥ睍绀烘敼閫狅紙灞曞紑鎻忚堪/璇婃柇锛?| 褰撳墠灞曠ず澶畝闄?|
| P1 | 鍚屾浠诲姟鎵嬪姩杩愯鎸夐挳 | 鏂逛究杩愮淮瑙﹀彂 |
| P1 | 鍚屾杩愯鍘嗗彶 | 鏂逛究鎺掓煡鍚屾闂 |
| P2 | 鍚屾绠＄悊涓績椤甸潰 | 缁熶竴绠＄悊澶氫换鍔?|
| P2 | 鏈尮閰嶆偅鑰呴槦鍒?| 鍚屾鍚庢湡缁存姢闇€瑕?|

---

## 17. 瑙﹀彂鏂瑰紡銆佹偅鑰呰寖鍥翠笌鍏宠仈鍙鎬у鏍革紙2026-06-19锛?
### 17.1 鏄惁鍙互鍚屾椂鏀寔涓ょ鑾峰彇鏂瑰紡

鍙互锛屼笖寤鸿鍚屾椂鏀寔锛?
| 鏂瑰紡 | 瑙﹀彂鐐?| 閫傜敤鍦烘櫙 | 瀹炵幇寤鸿 |
|---|---|---|---|
| 瀹氭椂浠诲姟鍚屾 | `his-sync` 鐙珛杩涚▼鎸?`sync_job_configs` 鍛ㄦ湡鎵ц | 淇濇寔鏈湴妫€鏌ユ姤鍛婄紦瀛樻寔缁洿鏂?| 鎵归噺鎵弿宸插缓妗ｆ偅鑰呯殑澶栭儴鏄犲皠锛屾寜鏃堕棿娓告爣澧為噺鍚屾 |
| 鐐瑰嚮鏌ョ湅鏃惰嚜鍔ㄨ幏鍙?| 鎮ｈ€呰鎯呴〉杩涘叆鈥滄楠屾鏌モ€漈ab 鎴栫偣鍑烩€滃埛鏂版鏌ユ姤鍛娾€?| 寮ヨˉ瀹氭椂浠诲姟寤惰繜锛屽尰鐢熶复搴婃煡鐪嬫椂瀹炴椂琛ラ綈 | 浠呭褰撳墠鎮ｈ€呰Е鍙戠煭瓒呮椂鍚屾锛岀劧鍚庤鍙栨湰鍦?`exam_reports` 灞曠ず |

涓ょ鏂瑰紡涓嶅啿绐侊紝浣嗗繀椤诲叡鐢ㄥ悓涓€涓箓绛?upsert 閫昏緫锛?
- 鍞竴閿缓璁細`source_system + external_report_id`锛屾垨鍦ㄦ棤娉曚繚璇佸叏灞€鍞竴鏃朵娇鐢?`patient_id + source_system + external_report_id`銆?- 姣忔鍚屾鍏堟寜澶栭儴鎶ュ憡鍙锋煡宸叉湁璁板綍锛屽瓨鍦ㄥ垯鏇存柊锛屼笉瀛樺湪鍒欐柊澧炪€?- 鐐瑰嚮鏌ョ湅瑙﹀彂鐨勫悓姝ュ繀椤绘湁瓒呮椂鍜岄檷绾э細HIS Oracle 涓嶅彲鐢ㄦ椂锛屼笉闃绘柇椤甸潰灞曠ず鏈湴宸叉湁鎶ュ憡銆?- 鍓嶇涓嶅簲鐩存帴鏌?HIS锛涘墠绔彧璋冪敤鏈郴缁熷悗绔帴鍙ｏ紝鐢卞悗绔畬鎴愭潈闄愩€佹偅鑰呰寖鍥淬€佹槧灏勫拰瀹¤銆?
鐜版湁浠ｇ爜鐘舵€侊細

- `LabsExamsTab.tsx` 宸插湪鎮ｈ€呰鎯呭姞杞芥椂璋冪敤 `restApi.syncExamReports(patient.id)`锛屽叿澶団€滅偣鍑绘煡鐪嬭嚜鍔ㄨ姹傝幏鍙栤€濈殑璋冪敤褰㈡€併€?- `ExamReportSyncService.SyncPatientExamReports` 褰撳墠浠嶆槸 HDIS 楠ㄦ灦锛屽苟涓?`getHDISPatientID()` 鏄庣‘杩斿洖涓嶅彲鐢ㄣ€?- 鍚庣画闇€瑕佹妸 `POST /api/v1/patients/:id/exam-reports/sync` 鏀规垚 HIS Oracle 鎮ｈ€呯骇鍚屾锛岃€屼笉鏄户缁緷璧?HDIS PatientId銆?
### 17.2 鏄惁鍙悓姝ヨ閫忕郴缁熶腑宸插缓妗ｆ偅鑰?
蹇呴』鍙悓姝ュ凡寤烘。鎮ｈ€呯殑妫€鏌ユ姤鍛婏紝涓嶅仛 HIS 鍏ㄩ櫌鎮ｈ€呭叏閲忓叆搴撱€?
鍘熷洜锛?
- HIS `EXAM_MASTER` 鏈?339 涓囨鏌ョ敵璇凤紝杩滃ぇ浜庤閫忕郴缁熸偅鑰呰妯°€?- 鏈郴缁熷綋鍓?`TenantId=3` 琛€閫忓缓妗ｆ偅鑰呭彧鏈?365 浜恒€?- 鍏ㄩ噺鍚屾浼氬甫鏉ユ棤鍏虫暟鎹€侀殣绉佹毚闇层€佸瓨鍌ㄨ啫鑳€鍜屾偅鑰呭尮閰嶅櫔闊炽€?
鎺ㄨ崘鍚屾鑼冨洿瑙勫垯锛?
| 鍦烘櫙 | 鑼冨洿鎺у埗 |
|---|---|
| 瀹氭椂浠诲姟 | 鍙鐞?`Register_PatientInfomation` 涓?`TenantId=3` 鐨勬偅鑰咃紱浼樺厛澶勭悊宸叉湁 `external_patient_mappings` 鐨勬偅鑰?|
| 鎮ｈ€呰鎯呯偣鍑诲悓姝?| 鍙厑璁稿悓姝?URL 涓綋鍓?`patientId` 瀵瑰簲鐨勬湰鍦板缓妗ｆ偅鑰?|
| 鏈尮閰?HIS 鎶ュ憡 | 涓嶇洿鎺ュ叆 `exam_reports`锛屾渶澶氳繘鍏モ€滄湭鍖归厤闃熷垪/鍚屾閿欒鏄庣粏鈥?|
| HIS 鏂版偅鑰呭缓妗?| 璧板崟鐙€滃缓妗ｅ鍏?鍊欓€夋偅鑰呪€濇祦绋嬶紝涓嶈兘鐢辨鏌ユ姤鍛婂悓姝ヨ嚜鍔ㄥ垱寤烘偅鑰?|

瀹氭椂浠诲姟涓嶅缓璁娇鐢ㄢ€滄煡璇?HIS 鏈€杩戞墍鏈夋姤鍛婂啀鍙嶅悜鍖归厤鏈湴鎮ｈ€呪€濈殑鏂瑰紡浣滀负涓绘祦绋嬨€傛洿绋冲Ε鐨勬槸锛?
1. 鍏堣鍙栨湰鍦板凡寤烘。鎮ｈ€呭垪琛ㄦ垨 `external_patient_mappings`銆?2. 瀵规瘡涓凡鏄犲皠鎮ｈ€呮寜 HIS `PATIENT_ID/VISIT_ID` 鎷夊彇鎶ュ憡銆?3. 瀵规湭寤虹珛鏄犲皠鐨勬湰鍦版偅鑰咃紝浣跨敤璇佷欢鍙?閫忔瀽鍙?浣忛櫌鍙风瓑鍘?HIS 鎮ｈ€呬富绱㈠紩琛ㄦ煡鏄犲皠銆?4. 鍙妸鑳界‘瀹氬叧鑱斿埌鏈湴鎮ｈ€呯殑鎶ュ憡鍐欏叆 `exam_reports`銆?
### 17.3 涓庡綋鍓嶈閫忔偅鑰呬俊鎭殑鍏宠仈鍙鎬?
鏈湴鑰佽閫忓簱鏈夎冻澶熺殑鍊欓€夋ˉ鎺ュ瓧娈碉紝浣?HIS 妫€鏌ヤ笁琛ㄦ湰韬笉瓒充互瀹屾垚瀹夊叏鑷姩鍏宠仈銆?
宸茶ˉ鍏呯‘璁ょ殑 HIS 鎮ｈ€?灏辫瘖琛細

| HIS 琛?| 鍚箟 | 瀵瑰悓姝ョ殑浣滅敤 |
|---|---|---|
| `PAT_MASTER_INDEX` | 鎮ｈ€呬富绱㈠紩锛涘寘鍚偅鑰呮墍鏈夐棬璇婂強浣忛櫌寤烘。淇℃伅锛屽惈韬唤璇?| 鐢ㄤ簬閫氳繃 `PATIENT_ID` 鎵捐韩浠借瘉鍜屾偅鑰呬富妗ｆ |
| `PAT_VISIT` | 浣忛櫌灏辫瘖璁板綍锛涙瘡娆′綇闄骇鐢熶竴鏉?| 閫氳繃 `PATIENT_ID + VISIT_ID` 杈呭姪纭浣忛櫌灏辫瘖璁板綍锛涙偅鑰呬富鍏宠仈缁熶竴浣跨敤 `PATIENT_ID` |
| `CLINIC_MASTER` | 闂ㄨ瘖灏辫瘖璁板綍锛涙瘡娆￠棬璇婂氨璇婁骇鐢熶竴鏉?| 閫氳繃 `PATIENT_ID` 鍏宠仈闂ㄨ瘖灏辫瘖璁板綍锛涙偅鑰呬富鍏宠仈缁熶竴浣跨敤 `PATIENT_ID` |

鍥犳锛孒IS 渚у簲鍏堥€氳繃妫€鏌ユ姤鍛婄殑 `PATIENT_ID/VISIT_ID` 鍏宠仈涓婅堪鎮ｈ€?灏辫瘖琛紝鍐嶄笌鏈湴琛€閫忓缓妗ｆ偅鑰呭尮閰嶃€備笉瑕佸彧鎷?`EXAM_MASTER.NAME + DATE_OF_BIRTH` 鍋氳嚜鍔ㄧ粦瀹氥€?
鏈湴 `TenantId=3` 鎮ｈ€呮ˉ鎺ュ瓧娈佃鐩栫巼瀹炴煡缁撴灉锛?
| 瀛楁 | 瑕嗙洊鏁?/ 365 | 閲嶅閿暟閲?| 璇勪环 |
|---|---:|---:|---|
| `Register_IDInfomation.IDNo` 璇佷欢鍙?| 365 | 0 | 寮哄尮閰嶉閫?|
| `Register_PatientInfomation.DialysisNo` 閫忔瀽鍙?| 365 | 0 | 寮哄尮閰嶏紝浣嗛渶纭 HIS 鏄惁瀛樿瀛楁 |
| 濮撳悕 + 鍑虹敓鏃ユ湡 | 365 | 0 | 鍙綔浜哄伐鍊欓€?杈呭姪鍖归厤锛屼笉寤鸿鑷姩寮虹粦瀹?|
| `Register_Hospitalization.HospNo` 浣忛櫌鍙?| 178 | 2 | 鍙緟鍔╋紝瀛樺湪閲嶅 |
| `Register_Hospitalization.CaseNo` 灏辫瘖鍙?闂ㄨ瘖鍙?| 23 | 1 | 瑕嗙洊浣庯紝浠呰緟鍔?|
| `Register_Hospitalization.MedicalRecordNo` 鐥呮鍙?| 324 | 9 | 瑕嗙洊杈冮珮浣嗛噸澶嶈緝澶氾紝浠呰緟鍔?|

褰撳墠 HIS 妫€鏌ヤ笁琛ㄥ彲鐢ㄦ偅鑰呭瓧娈碉細

| HIS 瀛楁 | 鎵€鍦ㄨ〃 | 鏄惁瓒冲鑷姩鍏宠仈鏈湴鎮ｈ€?|
|---|---|---|
| `PATIENT_ID` | `EXAM_MASTER` | 鍚︼紝鏈湴鎮ｈ€?ID 鏍煎紡瀹屽叏涓嶅悓 |
| `VISIT_ID` | `EXAM_MASTER` | 鍚︼紝闇€瑕?HIS 灏辫瘖琛ㄨВ閲?|
| `NAME` | `EXAM_MASTER` | 鍚︼紝鍚屽悕椋庨櫓楂?|
| `SEX` | `EXAM_MASTER` | 鍚︼紝浠呰緟鍔?|
| `DATE_OF_BIRTH` | `EXAM_MASTER` | 鍚︼紝鍙笌濮撳悕缁勫悎浣滃€欓€?|

缁撹锛?
- 鎮ｈ€呬富鍏宠仈缁熶竴浣跨敤 HIS `PATIENT_ID`锛屼笉瑕佷娇鐢?`INP_NO` 浣滀负涓诲尮閰嶅瓧娈点€?- `PAT_MASTER_INDEX` 閫氳繃 `PATIENT_ID` 鍏宠仈鎮ｈ€呬富绱㈠紩锛宍PAT_VISIT` 閫氳繃 `PATIENT_ID + VISIT_ID` 瀹氫綅浣忛櫌灏辫瘖璁板綍锛宍CLINIC_MASTER` 閫氳繃 `PATIENT_ID` 鍏宠仈闂ㄨ瘖璁板綍銆?- `INP_NO`銆乣INP_SERIAL_NO`銆乣CLINIC_NO`銆乣MR_NO`銆乣VISIT_NO` 鍙綔涓鸿緟鍔╂牳瀵瑰瓧娈碉紝浣嗕笉浣滀负鎮ｈ€呬富鍏宠仈鏉′欢銆?- `DialysisNo` 鏄湰鍦拌閫忕郴缁熷瓧娈碉紝HIS 涓嶅瓨鍦ㄩ€忔瀽鍙锋椂涓嶈兘浣滀负 HIS 鑷姩鍖归厤鏉′欢锛屽彧鑳藉湪鏈湴渚у睍绀烘垨鐢ㄤ簬浜哄伐鏍稿銆?
鎺ㄨ崘鑷姩鍖归厤绛栫暐锛?
1. 浼樺厛澶嶇敤 `external_patient_mappings` 涓凡纭鐨?`HIS_ORACLE + PATIENT_ID` 鏄犲皠銆?2. 鏃犳槧灏勬椂锛岄€氳繃 `PAT_MASTER_INDEX.PATIENT_ID` 鑾峰彇 `ID_NO`锛屼笌鏈湴 `Register_IDInfomation.IDNo` 鑷姩鍖归厤銆?3. `PAT_VISIT` 鍜?`CLINIC_MASTER` 鍧囧洿缁?`PATIENT_ID` 鍏宠仈锛宍VISIT_ID` 鍙敤浜庤緟鍔╅檺瀹氭湰娆′綇闄㈣褰曘€?4. `INP_NO`銆乣INP_SERIAL_NO`銆乣CLINIC_NO`銆乣MR_NO`銆乣VISIT_NO` 鍙綔涓鸿緟鍔╂牳瀵?灞曠ず锛屼笉浣滀负涓诲尮閰嶅瓧娈点€?5. 鍖归厤鎴愬姛鍚庡啓鍏?`external_patient_mappings`锛屽悗缁悓姝ヤ紭鍏堜娇鐢ㄨ鏄犲皠锛岄伩鍏嶆瘡娆￠噸澶嶅尮閰嶃€?6. 浠讳綍澶氬€欓€夋垨鍐茬獊缁撴灉杩涘叆鏈尮閰?寰呯‘璁ら槦鍒楋紝涓嶈嚜鍔ㄥ叆搴撴鏌ユ姤鍛娿€?
### 17.4 鎺ュ彛琛屼负寤鸿

寤鸿鎶婅鍙栦笌鍚屾鎷嗗紑锛屼絾鍓嶇鐐瑰嚮鏌ョ湅鏃跺彲浠ョ粍鍚堣皟鐢細

```text
GET  /api/v1/patients/:id/exam-reports
POST /api/v1/patients/:id/exam-reports/sync
```

鎮ｈ€呰鎯呴〉鎵撳紑鏃讹細

1. 鍏?`GET` 鏈湴宸叉湁鎶ュ憡锛岄〉闈㈠揩閫熷睍绀恒€?2. 鍚庡彴 `POST /sync?mode=on_demand` 灏濊瘯鎷夊彇 HIS 鏈€鏂版姤鍛娿€?3. 鍚屾鎴愬姛鍚庡啀娆?`GET` 鍒锋柊鍒楄〃銆?4. 鍚屾澶辫触鍙彁绀衡€滃凡灞曠ず鏈湴缂撳瓨锛孒IS 鍚屾澶辫触鈥濓紝涓嶆竻绌哄凡鏈夋姤鍛娿€?
瀹氭椂浠诲姟锛?
```text
cmd/his-sync --job his_exam_report
```

浠诲姟鎵ц鏃讹細

1. 璇诲彇鍚敤鐘舵€併€侀鐜囥€佹壒澶у皬銆佹父鏍囥€?2. 璇诲彇 `TenantId=3` 宸插缓妗ｆ偅鑰呭強澶栭儴鏄犲皠銆?3. 鍙悓姝ュ凡鏄犲皠鎴栧彲纭畾鍖归厤鐨勬偅鑰呫€?4. 璁板綍杩愯缁熻鍜岄敊璇槑缁嗐€?5. 鏇存柊娓告爣銆?
### 17.5 褰撳墠宸茬‘璁ら」

1. 宸茬‘璁?HIS 瀛樺湪 `PAT_MASTER_INDEX`锛氭偅鑰呬富绱㈠紩锛屽寘鍚偅鑰呮墍鏈夐棬璇婂強浣忛櫌寤烘。淇℃伅锛屽苟鍖呭惈韬唤璇佸瓧娈?`ID_NO`銆?2. 宸茬‘璁?HIS 瀛樺湪 `PAT_VISIT`锛氫綇闄㈠氨璇婅褰曪紝姣忔浣忛櫌浜х敓涓€鏉★紝閫氳繃 `PATIENT_ID + VISIT_ID` 涓庢鏌ヨ褰曞叧鑱斻€?3. 宸茬‘璁?HIS 瀛樺湪 `CLINIC_MASTER`锛氶棬璇婂氨璇婅褰曪紝姣忔闂ㄨ瘖灏辫瘖浜х敓涓€鏉★紝閫氳繃 `PATIENT_ID` 涓庢偅鑰呬富绱㈠紩鍏宠仈銆?4. 宸茬‘璁ゆ偅鑰呬富鍏宠仈缁熶竴浣跨敤 `PATIENT_ID`锛屼笉瑕佷娇鐢?`INP_NO` 浣滀负涓诲尮閰嶅瓧娈点€?5. 宸茬‘璁?`RESULT_STATUS` 鐮佽〃锛歚1=鏀跺埌鐢宠, 2=宸叉墽琛? 3=鍒濇鎶ュ憡, 4=纭鎶ュ憡, 5=PACS鍙栨秷, 9=鍏朵粬`銆?6. 宸茬‘璁?`IS_ABNORMAL`锛歚1=闃虫€э紝鍗虫鏌ュ彲鑳芥湁鐥呭彉锛涘叾浠?闃存€銆?7. 宸茬‘璁ゅ悓姝ョ瓥鐣ワ細`RESULT_STATUS=3`锛堝垵姝ユ姤鍛婏級鍜?`4`锛堢‘璁ゆ姤鍛婏級閮藉啓鍏?`exam_reports`锛涘悗缁悓姝ユ椂鑻ユ姤鍛婁粠 `3` 鏇存柊涓?`4`锛屽垯瑕嗙洊鏇存柊銆?
## 18. P2 鍚屾绠＄悊涓績鏀跺熬锛?026-06-19锛?
### 18.1 鍐崇瓥璁板綍

| # | 浜嬮」 | 鍐崇瓥 |
|---|---|---|
| 1 | 鏈尮閰嶆偅鑰呭鍚嶈劚鏁?| 淇濇寔褰撳墠瑙勫垯锛氬*鍚嶏紱鍗曞瓧鍚嶆樉绀?*锛涘弻瀛楀悕鏄剧ず 濮? |
| 2 | 鏈尮閰嶆偅鑰呴槦鍒楀垎椤?| 鏀寔 `page`/`pageSize`/`keyword` 鍙傛暟锛屽悗绔?Oracle 鐢?ROWNUM 鍒嗛〉锛屽墠绔?Ant Design 鍒嗛〉缁勪欢 |
| 3 | 缁戝畾鍚庡鍚嶅啓鍏?| 涓嶅洖濉畬鏁?HIS 濮撳悕锛涘綋鍓?`BindMapping` 涓嶈缃?`patient_name` 瀛楁锛屼繚鎸?NULL |
| 4 | 瀹氭椂璋冨害 | 鎺ㄨ崘绯荤粺绾?cron/systemd timer 璋冪敤 `his-sync --job his_exam_report --once`锛涗笉瀹炵幇 Go 甯搁┗ worker |
| 5 | exam_report_items 鏄庣粏 | 鏆備笉鍚敤锛屽彧淇濈暀涓绘姤鍛婂悓姝ワ紱鍚庣画鍙惎鍔ㄥ仛妫€鏌ラ」鐩粨鏋勫寲灞曠ず |
| 6 | skipped 缁熻 | 鏃犳槧灏勬偅鑰呰鍏?skipped锛堜笉璁?failed锛夛紱鍓嶇杩愯鍘嗗彶灞曠ず created/updated/skipped/failed 鍥涘垪 |

### 18.2 鍚庣鏀归€犺杩?
#### 18.2.1 鏈尮閰嶆偅鑰?API 鍒嗛〉

**鎺ュ彛**: `GET /api/v1/sync/unmatched-patients`

**Query 鍙傛暟**:

| 鍙傛暟 | 绫诲瀷 | 榛樿鍊?| 璇存槑 |
|---|---|---|---|
| `page` | int | 1 | 椤电爜锛屾渶灏?1 |
| `pageSize` | int | 20 | 姣忛〉鏉℃暟锛屾渶澶?100 |
| `keyword` | string | (绌? | 鎼滅储 HIS 鎮ｈ€呭彿鎴栧鍚嶏紱涓嶆敮鎸佽韩浠借瘉鎼滅储 |

**鍝嶅簲缁撴瀯**:

```json
{
  "success": true,
  "data": {
    "items": [
      { "patientId": "00001234", "name": "寮?涓?, "examCnt": 15 }
    ],
    "total": 384449,
    "page": 1,
    "pageSize": 20
  }
}
```

- `items`锛氬綋鍓嶉〉鏈尮閰嶆偅鑰呭垪琛紝濮撳悕宸茶劚鏁?- `total`锛歄racle 绔尮閰嶇殑鎮ｈ€呮€绘暟锛堜笉鎵ｉ櫎宸叉槧灏勬暟锛岀敤浜庡垎椤靛弬鑰冿級
- `page`/`pageSize`锛氬綋鍓嶅垎椤典俊鎭?
**Oracle SQL 瀹炵幇**锛?
- 鍒嗛〉浣跨敤 `ROWNUM` 鍙屽眰宓屽锛堝吋瀹?Oracle 11g锛?- 姣忔浠?Oracle 鍙?`pageSize * 3` 鏉★紝鍦?Go 绔繃婊ゅ凡 confirmed 鏄犲皠锛屾渶缁堣繑鍥炰笉瓒呰繃 `pageSize` 鏉?- 璁℃暟 SQL 鐙珛鎵ц `SELECT COUNT(*) FROM (SELECT patient_id ... GROUP BY patient_id)`

**Oracle 瀹㈡埛绔柟娉?*锛?
- `QueryUnmatchedPatients(ctx, params UnmatchedPatientsParams) (*UnmatchedPatientsResult, error)`
- `UnmatchedPatientsParams{Page, PageSize, Keyword}`
- `UnmatchedPatientsResult{Items, Total}`
- keyword 鎼滅储锛歚a.name LIKE '%kw%' OR TO_CHAR(a.patient_id) LIKE '%kw%'`

#### 18.2.2 缁戝畾鎺ュ彛闅愮淇濇姢

`POST /api/v1/sync/external-mappings/bind` 涓嶅啓鍏?`patient_name` 瀛楁锛圢ULL锛夛紝涓嶅瓨鍌ㄥ畬鏁?HIS 濮撳悕銆?
#### 18.2.3 鍚屾鍘嗗彶瀛楁

`SyncJobRun` 妯″瀷宸插寘鍚?`CreatedCount`銆乣UpdatedCount`銆乣SkippedCount`銆乣FailedCount`锛孞SON tag 宸叉槧灏勪负 camelCase銆傚墠绔彲浠?`GET /api/v1/sync/jobs/:code/runs` 鐩存帴鑾峰彇銆?
### 18.3 鍓嶇鏀归€犺杩?
#### 18.3.1 SyncCenterPage 鏂板浜や簰

| 鍔熻兘 | 璇存槑 |
|---|---|
| 鏈尮閰嶆偅鑰呮悳绱?| 杈撳叆妗嗘敮鎸?HIS 鎮ｈ€呭彿 / 濮撳悕鎼滅储锛屽洖杞︽垨鐐瑰嚮鎼滅储鎸夐挳瑙﹀彂 |
| 鏈尮閰嶆偅鑰呭垎椤?| 鏄剧ず"鍏?X 涓偅鑰?路 绗?N/M 椤?锛屽乏鍙崇炕椤垫寜閽?|
| 杩愯鍘嗗彶鍒?| 灞曠ず 7 鍒楋細鏃堕棿銆佺姸鎬併€佽幏鍙栥€佹柊澧烇紙缁胯壊锛夈€佹洿鏂帮紙钃濊壊锛夈€佽烦杩囷紙鐏拌壊锛夈€佸け璐ワ紙绾㈣壊锛?|
| 缁戝畾鍚庡埛鏂?| 缁戝畾鎴愬姛鍚庤嚜鍔ㄥ埛鏂板綋鍓嶉〉鏈尮閰嶆偅鑰呮暟鎹?|

#### 18.3.2 restClient 绫诲瀷鏇存柊

鏂板 `UnmatchedPatientResponse` 绫诲瀷锛?
```typescript
export interface UnmatchedPatientResponse {
  items: UnmatchedPatientItem[]
  total: number
  page: number
  pageSize: number
}
```

`getUnmatchedPatients` 鏂规硶绛惧悕锛?
```typescript
async getUnmatchedPatients(params?: { page?: number; pageSize?: number; keyword?: string }): Promise<UnmatchedPatientResponse>
```

### 18.4 鐢熶骇瀹氭椂璋冨害寤鸿

鎺ㄨ崘浣跨敤 systemd timer 鎴?cron 璋冨害 `his-sync` 鐙珛杩涚▼锛?
```cron
# 姣?30 鍒嗛挓澧為噺鍚屾 HIS 妫€鏌ユ姤鍛?*/30 * * * * /opt/ai-hms/his-sync --job his_exam_report --once >> /var/log/his-sync.log 2>&1
```

涓嶆帹鑽愬湪 `cmd/server` 鍐呭疄鐜板父椹?worker锛屽師鍥狅細
- 鍚屾鏄壒閲忎换鍔★紝涓嶉渶瑕侀暱杩炴帴
- 绯荤粺璋冨害鎻愪緵鏇村ソ鐨勬棩蹇椼€佺洃鎺с€侀噸鍚兘鍔?- 閬垮厤 Go goroutine 寮傚父瀵艰嚧浠诲姟闈欓粯鍋滄憜

### 18.5 楠岃瘉缁撴灉锛?026-06-19锛?
| 妫€鏌?| 缁撴灉 |
|---|---|
| `go build ./cmd/server` | 閫氳繃 |
| `go build ./cmd/his-sync` | 閫氳繃 |
| `go vet ./internal/...` | 閫氳繃 |
| `go test ./internal/services -count=1` | 閫氳繃 |
| `npx tsc -b --noEmit` | 閫氳繃 |
| `npx eslint SyncCenterPage + restClient` | 閫氳繃 |

### 18.6 鍚庣画寰呯‘璁?
| # | 浜嬮」 | 璇存槑 |
|---|---|---|
| 1 | ~~`exam_report_items` 鏄庣粏鍚姩鏃舵満~~ | **宸插惎鐢?*锛?026-06-19锛夛紝瑙?18.7 |
| 2 | 鍘嗗彶鍏ㄩ噺鍒濆鍖?| 鍙悓姝ヤ竴骞村唴鏁版嵁锛坄CREATEDATE >= 2025-06-19`锛夛紝鐜版湁澧為噺娓告爣浠?2025-04-25 璧峰彲瑕嗙洊 |
| 3 | PACS/STUDY_UID 閾炬帴灞曠ず | 鍙敤浜庡墠绔烦杞?PACS 绯荤粺鏌ョ湅鍘熷浘 |
| 4 | ~~缁戝畾鍚庤嚜鍔ㄨЕ鍙戞姤鍛婂悓姝~ | **宸插疄鐜?*锛?026-06-19锛夛紝瑙?18.8 |

## 19. 2026-06-19 澧為噺鏀归€?
### 19.1 鍚敤 `exam_report_items` 鏄庣粏鍚屾

**鏁版嵁鏉ユ簮**: `HIS.EXAM_ITEMS` 琛紝瀛楁 `exam_no / exam_item / exam_item_code / exam_item_no`

**瀹炵幇鏂瑰紡**:
- 鏂板 `ExamReportItem` 妯″瀷锛坄internal/models/lab_report.go`锛?- 鏂板 `HisExamItemRow` 绫诲瀷锛坄internal/integrations/his_oracle/types.go`锛?- 鏂板 `QueryExamItems(ctx, examNos)` Oracle 鏂规硶锛屾寜 500 鏉″垎鎵?IN 鏌ヨ锛坄client.go`锛?- 鏂板 `syncExamItems` 鏂规硶锛氬厛鎵归噺鍒犻櫎鏃ф槑缁嗭紝鍐嶄粠 Oracle 鏌ヨ鍚庢壒閲忔彃鍏ワ紙`his_exam_report_sync_service.go`锛?- `SyncBatch` 鍜?`SyncPatientExamReports` 鏈熬鑷姩璋冪敤 `syncExamItems`

**鍚屾娴佺▼**:
1. 涓绘姤鍛婂啓鍏?鏇存柊 `exam_reports`锛堝凡鏈夐€昏緫锛?2. 鏀堕泦鏈娑夊強鐨?`exam_no 鈫?report_id` 鏄犲皠
3. 浠?HIS Oracle 鎵归噺鏌ヨ `EXAM_ITEMS`锛堟瘡鎵?500 涓?exam_no锛?4. 鍒犻櫎杩欎簺 `report_id` 鐨勬棫鏄庣粏锛堟瘡鎵?100 涓級
5. 鎵归噺鎻掑叆鏂版槑缁嗭紙姣忔壒 200 鏉★級

**鍒犻櫎 cleanup**: 姣忔鍚屾璇ユ姤鍛婃椂锛屽厛鍒犻櫎鏃ф槑缁嗗啀鍐欏叆鏂版槑缁嗭紝纭繚鏁版嵁涓?HIS 涓€鑷淬€?
### 19.2 缁戝畾鍚庤嚜鍔ㄨЕ鍙戞姤鍛婂悓姝?
`POST /api/v1/sync/external-mappings/bind` 鎴愬姛鍚庯紝寮傛 goroutine 璋冪敤 `SyncPatientExamReports`锛?
```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    syncSvc := services.NewHisExamReportSyncService(hcfg, tenantID)
    defer syncSvc.CloseOracleClient()
    result, err := syncSvc.SyncPatientExamReports(ctx, legacyPatientID)
    // log results
}()
```

- 寮傛鎵ц锛屼笉闃诲缁戝畾 API 鍝嶅簲
- 5 鍒嗛挓瓒呮椂淇濇姢
- 鍚屾瀹屾垚鍚庤嚜鍔ㄥ叧闂?Oracle 杩炴帴
- 缁撴灉鍐欏叆鏈嶅姟鍣ㄦ棩蹇?
### 19.3 璋冨害鍛ㄦ湡淇

鐢熶骇 cron 鏀逛负 **姣?30 鍒嗛挓**:

```cron
*/30 * * * * /opt/ai-hms/his-sync --job his_exam_report --once >> /var/log/his-sync.log 2>&1
```

### 19.4 楠岃瘉缁撴灉锛?026-06-19锛?
| 妫€鏌?| 缁撴灉 |
|---|---|
| `go build ./cmd/server` | 閫氳繃 |
| `go build ./cmd/his-sync` | 閫氳繃 |
| `go vet ./internal/...` | 閫氳繃 |
| `go test ./internal/services -count=1` | 閫氳繃 |
| `npx tsc -b --noEmit` | 閫氳繃 |
