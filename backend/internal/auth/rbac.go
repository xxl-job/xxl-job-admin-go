package auth

import "strings"

// HasPermission 检查用户对 app 的执行器权限。
// 管理员全量；普通用户需在 permission（逗号分隔的 app_name 列表）中。
func HasPermission(role int8, permission string, app string) bool {
	if role == RoleAdmin {
		return true
	}
	if app == "" {
		return false
	}
	for _, p := range strings.Split(permission, ",") {
		if strings.TrimSpace(p) == app {
			return true
		}
	}
	return false
}
