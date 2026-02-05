-- SystemOperator 增加 cards R，仅用于前端显示 overview 路由；后端对该角色仍返回空列表
INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, permission_scope)
VALUES ('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'cards', 'R', 'A')
ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
DO UPDATE SET permission_scope = EXCLUDED.permission_scope;
