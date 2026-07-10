package auth

import "testing"

// TestHasPermission_Admin 管理员全量放行（无视 permission 列表）。
func TestHasPermission_Admin(t *testing.T) {
	if !HasPermission(RoleAdmin, "", "any-app") {
		t.Fatal("admin 应有全部权限")
	}
}

// TestHasPermission_UserAllowed 普通用户对授权 app 放行（含空白容错）。
func TestHasPermission_UserAllowed(t *testing.T) {
	if !HasPermission(RoleUser, "app-a, app-b", "app-b") {
		t.Fatal("普通用户对授权 app 应通过")
	}
	if !HasPermission(RoleUser, " app-a , app-b ", "app-a") {
		t.Fatal("permission 列表含空白应仍匹配")
	}
}

// TestHasPermission_UserDenied 普通用户对未授权/空 app 拒绝。
func TestHasPermission_UserDenied(t *testing.T) {
	if HasPermission(RoleUser, "app-a", "app-b") {
		t.Fatal("普通用户对未授权 app 应拒绝")
	}
	if HasPermission(RoleUser, "app-a", "") {
		t.Fatal("空 app 应拒绝")
	}
}
