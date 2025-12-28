-- Fix device_store permission for Manager: assigned_only should be TRUE
UPDATE role_permissions
SET assigned_only = TRUE
WHERE role_code = 'Manager'
  AND resource_type = 'device_store'
  AND tenant_id = '00000000-0000-0000-0000-000000000001';
