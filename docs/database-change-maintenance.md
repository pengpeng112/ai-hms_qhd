# 鏁版嵁搴撳彉鏇寸淮鎶ゆ竻鍗?
> 缁存姢鐩殑锛氭槑纭暟鎹簱鍙樻洿杈圭晫锛屽尯鍒嗏€淒BA 浜哄伐澶勭悊鐨勮€佽〃鍙樻洿鈥濆拰鈥滈儴缃查樁娈佃嚜鍔ㄦ墽琛岀殑鐙珛鏂拌〃鈥濄€?
> 閫傜敤搴擄細褰撳墠琛€閫?legacy PostgreSQL銆?
> 纭鍒欙細搴旂敤杩愯鏃剁姝?`AutoMigrate`銆乣DropTables` 鍜屼换浣曢殣寮?DDL銆?
## 1. SQL 鑴氭湰淇濈暀娓呭崟

| 鑴氭湰 | 鎵ц鏂?| 鐢ㄩ€?| 鏄惁淇濈暀 |
|---|---|---|---:|
| `docs/sql/deploy_new_tables.sql` | 閮ㄧ讲娴佺▼ | 鑷姩鍒涘缓鐙珛鏂拌〃锛屼笉淇敼鑰佽〃 | 鏄?|
| `docs/sql/old_table_extensions_dba.sql` | DBA | 鑰佽〃鎵╁瓧娈点€侀粯璁ゅ€间慨姝ｃ€佸彲閫夋暟鎹洖濉?| 鏄?|
| `ai-hms-backend/scripts/schedule_unique_indexes.sql` | DBA | 鑰佽〃鍞竴绱㈠紩锛屽惈閲嶅鏁版嵁鎺㈡祴 | 鏄?|

宸插垹闄?`docs/sql/` 涓巻鍙插弬鑰?閲嶅鑴氭湰锛歚v2_merge_legacy.sql`銆乣v1.3_v2_tables.sql`銆乣schedule_extension_tables.sql`銆乣schedule_patient_shift_unique_safety_net.sql`銆乣schedule_performance_indexes.sql`銆傝繖浜涘唴瀹瑰凡琚笂琛ㄤ繚鐣欒剼鏈垨褰撳墠浠ｇ爜瑙勫垯鏇夸唬銆?
## 2. 鎵ц杈圭晫

| 绫诲埆 | 鍙儴缃茶嚜鍔ㄦ墽琛?| 闇€ DBA 浜哄伐澶勭悊 | 璇存槑 |
|---|---:|---:|---|
| 鐙珛鏂拌〃 `CREATE TABLE IF NOT EXISTS` | 鏄?| 鍚?| 涓嶄慨鏀硅€佺郴缁熸棦鏈夎〃 |
| 鐙珛鏂拌〃绱㈠紩 | 鏄?| 鍚?| 闅忔柊琛ㄥ垱寤?|
| 鑰佽〃 `ALTER TABLE ADD COLUMN` | 鍚?| 鏄?| 浼氫慨鏀?legacy 鏃㈡湁琛?|
| 鑰佽〃榛樿鍊间慨鏀?| 鍚?| 鏄?| 浼氭敼鍙樺悗缁啓鍏ラ粯璁よ涓?|
| 鑰佽〃鍞竴绱㈠紩 | 鍚?| 鏄?| 蹇呴』鍏堝仛閲嶅鏁版嵁鎺㈡祴 |
| 鑰佽〃鏁版嵁鍥炲～/鎺ㄦ柇 | 鍚?| 鏄?| 闇€涓氬姟璇箟纭 |

## 3. 閮ㄧ讲鑷姩鏂板缓琛?
缁熶竴鑴氭湰锛歚docs/sql/deploy_new_tables.sql`

| 鏂拌〃 | 鍔熻兘 | 涓昏鍘熷洜 |
|---|---|---|
| `exam_reports` | HIS 妫€鏌ユ姤鍛婁富琛?| 淇濆瓨 HIS Oracle 妫€鏌ユ姤鍛婏紝涓嶆薄鏌撹€佹鏌?妫€楠岃〃 |
| `exam_report_items` | HIS 妫€鏌ユ姤鍛婇」鐩槑缁?| 鏀寔妫€鏌ラ」鐩粨鏋勫寲灞曠ず鍜屽悗缁粺璁?|
| `external_patient_mappings` | 澶栭儴鎮ｈ€呮槧灏?| 缁存姢 HIS `PATIENT_ID` 涓庢湰鍦版偅鑰?ID 鍏崇郴 |
| `sync_job_configs` | 鍚屾浠诲姟閰嶇疆 | 淇濆瓨鍚屾浠诲姟寮€鍏炽€佹壒閲忓ぇ灏忋€佹父鏍囩瓑閰嶇疆 |
| `sync_job_runs` | 鍚屾杩愯鍘嗗彶 | 淇濆瓨鍚屾鐘舵€併€佽€楁椂銆佹垚鍔?璺宠繃/澶辫触缁熻 |
| `sign_record` | 缁熶竴鐢靛瓙绛剧暀鐥?| 閬垮厤缁欏鏂?鏂规/灏忕粨鑰佽〃鍒嗗埆鍔犵鍚嶅瓧娈?|
| `Schedule_StaffDuty` | 浜哄姏鎺掔彮鏈堝熀绾?| 淇濆瓨鍖绘姢浜哄憳鏈堝害鍩虹褰撶彮瀹夋帓 |
| `Schedule_StaffDutyOverride` | 浜哄姏鎺掔彮褰撴棩瑕嗙洊 | 淇濆瓨椤剁彮銆佹崲鐝€佽鍋囧悗鐨勫疄闄呭綋鐝汉鍛?|
| `Schedule_Patient` | 鏅鸿兘鎺掔彮杞婚噺鎮ｈ€呮。妗?| 淇濆瓨鎺掔彮妯″潡鎵€闇€鎰熸煋鐘舵€?璞佸厤淇℃伅锛屼笉淇敼鎮ｈ€呬富妗ｈ€佽〃 |

## 4. DBA 鑰佽〃鎵╁瓧娈垫竻鍗?
缁熶竴鑴氭湰锛歚docs/sql/old_table_extensions_dba.sql`

### 4.1 `Schedule_Ward`

| 瀛楁 | SQL 鎽樿 | 鍘熷洜 |
|---|---|---|
| `ZoneType` | `ADD COLUMN IF NOT EXISTS "ZoneType" VARCHAR(8) NOT NULL DEFAULT 'A'` | 鏀寔鏅€?闅旂/HDF-CRRT 鍒嗗尯璇嗗埆 |
| `ParentWardId` | `ADD COLUMN IF NOT EXISTS "ParentWardId" BIGINT` | 鏀寔鐖跺瓙鐥呭尯缁撴瀯 |
| `IsSubZone` | `ADD COLUMN IF NOT EXISTS "IsSubZone" BOOLEAN DEFAULT false` | 鏍囪鏄惁瀛愬垎鍖?|

### 4.2 `Schedule_Shift`

| 瀛楁 | SQL 鎽樿 | 鍘熷洜 |
|---|---|---|
| `ShiftCode` | `ADD COLUMN IF NOT EXISTS "ShiftCode" VARCHAR(16) NOT NULL DEFAULT 'MORNING'` | 缁熶竴鏃?涓?鏅氱彮缂栫爜锛屽噺灏戝鑰?`Type` 瀛楁鐨勪緷璧?|

### 4.3 `Schedule_Bed`

| 瀛楁 | SQL 鎽樿 | 鍘熷洜 |
|---|---|---|
| `MachineType` | `ADD COLUMN IF NOT EXISTS "MachineType" VARCHAR(8) NOT NULL DEFAULT 'HD'` | 鏍囪瘑璁惧绫诲瀷 |
| `SupportedModes` | `ADD COLUMN IF NOT EXISTS "SupportedModes" VARCHAR(64) NOT NULL DEFAULT 'HD'` | 鏍囪瘑鏈轰綅鏀寔鐨勬不鐤楁ā寮?|
| `PositionIndex` | `ADD COLUMN IF NOT EXISTS "PositionIndex" INT NOT NULL DEFAULT 0` | 鏈轰綅灞曠ず/鎺掔彮鎺掑簭 |
| `LegacyBedName` | `ADD COLUMN IF NOT EXISTS "LegacyBedName" VARCHAR(256)` | 淇濈暀鑰佸簥浣嶅悕渚夸簬杩芥函 |
| `Code` | `ADD COLUMN IF NOT EXISTS "Code" VARCHAR(64)` | 璁惧/鏈轰綅缂栫爜 |

### 4.4 `Schedule_PatientShift`

| 瀛楁/淇敼 | SQL 鎽樿 | 鍘熷洜 |
|---|---|---|
| `DialysisMode` | `ADD COLUMN IF NOT EXISTS "DialysisMode" VARCHAR(8) NOT NULL DEFAULT 'HD'` | 淇濆瓨鏈鎺掔彮娌荤枟妯″紡 |
| `SourceType` | `ADD COLUMN IF NOT EXISTS "SourceType" SMALLINT NOT NULL DEFAULT 10` | 鏍囪瘑鎺掔彮鏉ユ簮 |
| `RecordForm` | `ADD COLUMN IF NOT EXISTS "RecordForm" SMALLINT NOT NULL DEFAULT 10` | 鏍囪瘑鏅€?涓存椂/琛ユ帓绛夎褰曞舰鎬?|
| `Confirm1At/2At/3At` | `ADD COLUMN IF NOT EXISTS ... TIMESTAMPTZ` | 涓夌骇纭鏃堕棿 |
| `Confirm1By/2By/3By` | `ADD COLUMN IF NOT EXISTS ... BIGINT` | 涓夌骇纭浜哄憳 |
| `IsBorrowedSlot` | `ADD COLUMN IF NOT EXISTS "IsBorrowedSlot" BOOLEAN DEFAULT false` | 鏍囪瘑鍊熺敤鏈轰綅 |
| `CancelReason` | `ADD COLUMN IF NOT EXISTS "CancelReason" VARCHAR(256)` | 淇濆瓨鎾ら攢/鍙栨秷鍘熷洜 |
| `SourceTemplateItemId` | `ADD COLUMN IF NOT EXISTS "SourceTemplateItemId" BIGINT` | 杩借釜鏉ユ簮妯℃澘椤?|
| `IsLocked` | `ADD COLUMN IF NOT EXISTS "IsLocked" BOOLEAN DEFAULT false` | 閿佸畾鎺掔彮闃叉璇敼 |
| `MachineId` | `ADD COLUMN IF NOT EXISTS "MachineId" BIGINT NOT NULL DEFAULT 0` | 寮曞叆鏈轰綅璇箟锛岄€愭鏇夸唬绾?`BedId` |
| `MakeupOfShiftId` | `ADD COLUMN IF NOT EXISTS "MakeupOfShiftId" BIGINT` | 璁板綍琛ユ帓鏉ユ簮 |
| `PatientPlanId/ShiftTiming/BedId/ShiftId` 榛樿鍊?| `ALTER COLUMN ... SET DEFAULT 0` | 閬垮厤 GORM 鏂板啓鍏ュ洜鑰佸垪鏃犻粯璁ゅ€煎け璐?|

### 4.5 `Schedule_PatientProfile`

| 瀛楁 | SQL 鎽樿 | 鍘熷洜 |
|---|---|---|
| `WeeklyCount` | `ADD COLUMN IF NOT EXISTS "WeeklyCount" SMALLINT` | 淇濆瓨姣忓懆閫忔瀽娆℃暟 |
| `PatientStatus` | `ADD COLUMN IF NOT EXISTS "PatientStatus" SMALLINT NOT NULL DEFAULT 10` | 鎺掔彮鎮ｈ€呯姸鎬?|
| `DischargeReason` | `ADD COLUMN IF NOT EXISTS "DischargeReason" VARCHAR(64)` | 杞嚭/鍋滄帓鍘熷洜 |
| `DischargedAt` | `ADD COLUMN IF NOT EXISTS "DischargedAt" TIMESTAMPTZ` | 杞嚭/鍋滄帓鏃堕棿 |
| `DischargedBy` | `ADD COLUMN IF NOT EXISTS "DischargedBy" BIGINT` | 杞嚭/鍋滄帓鎿嶄綔浜?|
| `FixedHdMachineId` | `ADD COLUMN IF NOT EXISTS "FixedHdMachineId" BIGINT` | 鍥哄畾 HD 鏈轰綅 |
| `FixedHdfMachineId` | `ADD COLUMN IF NOT EXISTS "FixedHdfMachineId" BIGINT` | 鍥哄畾 HDF 鏈轰綅 |

### 4.6 `Schedule_ScheduleTemplateItem`

| 瀛楁 | SQL 鎽樿 | 鍘熷洜 |
|---|---|---|
| `DefaultMode` | `ADD COLUMN IF NOT EXISTS "DefaultMode" VARCHAR(8) NOT NULL DEFAULT 'HD'` | 妯℃澘椤归粯璁ゆ不鐤楁ā寮?|
| `FixedHdMachineId` | `ADD COLUMN IF NOT EXISTS "FixedHdMachineId" BIGINT` | 妯℃澘鍥哄畾 HD 鏈轰綅 |
| `FixedHdfMachineId` | `ADD COLUMN IF NOT EXISTS "FixedHdfMachineId" BIGINT` | 妯℃澘鍥哄畾 HDF 鏈轰綅 |

### 4.7 鍏朵粬鎺掔彮鑰佽〃

| 鑰佽〃 | 瀛楁/淇敼 | 鍘熷洜 |
|---|---|---|
| `Schedule_Calendar` | `OpenWardIds`, `OpenMachineIds` | 鎸夋棩闄愬埗寮€鏀剧梾鍖?鏈轰綅 |
| `Schedule_MachineOutage` | `MachineId`; `BedId DEFAULT 0` | 璁惧鍋滄満寮曞叆鏈轰綅璇箟 |
| `Schedule_CrrtSession` | `MachineId`; `BedId DEFAULT 0` | CRRT 鍗犵敤寮曞叆鏈轰綅璇箟 |
| `Schedule_ConflictQueue` | `SuggestedDate`, `SuggestedShiftId`, `SuggestedBedId`, `SuggestedPatientShiftId`, `ResolvedBy`, `ResolvedAt` | 鍐茬獊寤鸿鍜屽鐞嗚褰?|

## 5. 鑰佽〃鍥炲～ SQL

鍥炲～ SQL 宸插啓鍦?`docs/sql/old_table_extensions_dba.sql` 灏鹃儴锛岄粯璁ゆ敞閲婏紝DBA 纭鍚庡啀鍙栨秷娉ㄩ噴鎵ц銆?
| 鍥炲～椤?| 鍘熷洜 | 鏄惁榛樿鎵ц |
|---|---|---:|
| `Schedule_BedMachineExt` 鈫?`Schedule_Bed` | 鍚堝苟鍘嗗彶璁惧鎵╁睍淇℃伅 | 鍚?|
| `Schedule_PatientShiftExt` 鈫?`Schedule_PatientShift` | 鍚堝苟鍘嗗彶鎺掔彮鎵╁睍淇℃伅 | 鍚?|
| `ZoneType` 鎺ㄦ柇 | 鑰佺梾鍖哄瓧娈佃涔変笉瀹屽叏纭畾锛岄渶涓氬姟纭 | 鍚?|
| `ShiftCode` 鎺ㄦ柇 | 鑰?`Type` 涓庢棭/涓?鏅氭槧灏勯渶纭 | 鍚?|
| `MachineId = BedId` 鍒濆鍖?| 灏嗚€佸簥浣嶈涔夎縼绉诲埌鏈轰綅瀛楁 | 鍚?|

## 6. 鑰佽〃鍞竴绱㈠紩

鑴氭湰锛歚ai-hms-backend/scripts/schedule_unique_indexes.sql`

| 绱㈠紩 | 琛?| 鍘熷洜 | 鍓嶇疆妫€鏌?|
|---|---|---|---|
| `uq_ps_patient_slot` | `Schedule_PatientShift` | 闃叉鍚屼竴鎮ｈ€呭悓鏃ュ悓鐝噸澶嶆湁鏁堟帓鐝?| 蹇呴』鍏堣窇閲嶅鏁版嵁鎺㈡祴 |
| `uq_ps_machine_slot` | `Schedule_PatientShift` | 闃叉鍚屼竴鏈轰綅鍚屾棩鍚岀彮閲嶅鏈夋晥鎺掔彮 | 蹇呴』鍏堢‘璁?`MachineId=0` 鑽夌鏄惁鍙備笌鍞竴鎬?|

## 7. 褰撳墠涓嶆墽琛岀殑璁捐棰勭暀椤?
| 琛?瀛楁 | 褰撳墠澶勭悊 |
|---|---|
| `Treatment_Treatment.Version` | 褰撳墠鏈嶅姟灞傛湭绋冲畾浣跨敤 `UpdateWithVersion`锛屾殏涓嶆墿鑰佽〃 |
| `App_IdempotencyKey` | 褰撳墠鏈彂鐜颁笟鍔′唬鐮佷娇鐢紝鏆備笉寤鸿〃 |
| `Audit_ClinicalWrite` | 褰撳墠鏈彂鐜颁笟鍔′唬鐮佷娇鐢紝鏆備笉寤鸿〃 |

## 8. 缁存姢瑙勫垯

1. 鏂板鐙珛琛ㄦ椂锛屽姞鍏?`docs/sql/deploy_new_tables.sql` 鍜屾湰鏂囩 3 鑺傘€?2. 淇敼鑰佽〃鏃讹紝鍔犲叆 `docs/sql/old_table_extensions_dba.sql` 鍜屾湰鏂囩 4 鑺傘€?3. 鑰佽〃鍞竴绱㈠紩蹇呴』鎻愪緵閲嶅鏁版嵁鎺㈡祴 SQL銆?4. 搴旂敤鍚姩鍜岃姹傝矾寰勪腑姘歌繙涓嶅緱鎵ц DDL銆?5. 娑夊強瀛楁璇箟涓嶇‘瀹氭椂锛屽厛鍐欏叆 `legacy-migration-uncertain-field-checklist.md`锛屼笉瑕佺寽娴嬨€?
