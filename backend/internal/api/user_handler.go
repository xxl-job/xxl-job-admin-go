package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"xxl-job-admin-go/internal/model"
)

// listUsers GET /api/users
func (r *Router) listUsers(c *gin.Context) {
	var users []model.User
	if err := r.db.Order("id desc").Find(&users).Error; err != nil {
		fail(c, 500, "查询用户列表失败: "+err.Error())
		return
	}
	ok(c, gin.H{"list": users})
}

// createUser POST /api/users
func (r *Router) createUser(c *gin.Context) {
	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Role       int8   `json:"role"`
		Permission string `json:"permission"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 500, err.Error())
		return
	}
	if req.Username == "" || req.Password == "" {
		fail(c, 400, "用户名和密码必填")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	u := model.User{Username: req.Username, Password: string(hash), Role: req.Role, Permission: req.Permission}
	if err := r.db.Create(&u).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, u)
}

// updateUser PUT /api/users/:id（密码留空不修改）
func (r *Router) updateUser(c *gin.Context) {
	var u model.User
	if err := r.db.First(&u, c.Param("id")).Error; err != nil {
		fail(c, 404, "user not found")
		return
	}
	var req struct {
		Role       *int8   `json:"role"`
		Permission *string `json:"permission"`
		Password   *string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 500, err.Error())
		return
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if req.Permission != nil {
		u.Permission = *req.Permission
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			fail(c, 500, "密码加密失败: "+err.Error())
			return
		}
		u.Password = string(hash)
	}
	if err := r.db.Save(&u).Error; err != nil {
		fail(c, 500, "保存用户失败: "+err.Error())
		return
	}
	ok(c, u)
}

// deleteUser DELETE /api/users/:id（禁止删除 id=1 超级管理员）
func (r *Router) deleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, 400, "无效的用户 id")
		return
	}
	if id == 1 {
		fail(c, 400, "不能删除超级管理员")
		return
	}
	if err := r.db.Delete(&model.User{}, id).Error; err != nil {
		fail(c, 500, "删除用户失败: "+err.Error())
		return
	}
	ok(c, nil)
}
