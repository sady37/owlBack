所有用户管理相关的权限检查已实现：
角色层级检查（创建/修改/删除用户）
assigned_only 检查（Caregiver/Nurse 只能查看自己）
branch_only 检查（Manager 只能查看同 branch 的用户，含空值匹配）
自己操作例外（修改密码/email/phone 不受限制）
所有后端权限检查已完成。
