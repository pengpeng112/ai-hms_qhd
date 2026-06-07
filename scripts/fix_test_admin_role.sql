-- test_admin 管理员角色修复脚本
-- 原因：角色权限从 Identity_Roles 迁移到 Authorization_Roles 后，test_admin 的关联未同步

-- 第一步：确认 test_admin 用户存在
SELECT "Id", "UserName" FROM "Identity_Users" WHERE "UserName" = 'test_admin';

-- 第二步：确认管理员角色存在于 Authorization_Roles
SELECT "Id", "Name" FROM "Authorization_Roles" WHERE "Name" IN ('ADMIN', '管理员', '安全管理员', '运维管理员');

-- 第三步：插入缺失的关联（幂等，不会重复插入）
INSERT INTO "Authorization_RoleUsers" ("UserId", "RoleId")
SELECT u."Id", r."Id"
FROM "Identity_Users" u, "Authorization_Roles" r
WHERE u."UserName" = 'test_admin'
  AND r."Name" = '管理员'
  AND NOT EXISTS (
    SELECT 1 FROM "Authorization_RoleUsers" ru
    WHERE ru."UserId" = u."Id" AND ru."RoleId" = r."Id"
  );

-- 如果需要补上所有管理员角色，改下面这句：
--   AND r."Name" IN ('ADMIN', '管理员', '安全管理员', '运维管理员')

-- 第四步：验证
SELECT u."UserName", r."Name" AS "RoleName"
FROM "Identity_Users" u
JOIN "Authorization_RoleUsers" ru ON ru."UserId" = u."Id"
JOIN "Authorization_Roles" r ON r."Id" = ru."RoleId"
WHERE u."UserName" = 'test_admin';
