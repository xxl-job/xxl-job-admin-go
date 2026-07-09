package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/model"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginSession POST /api/auth/sessions —— 登录，返回 accessToken/refreshToken。
func (r *Router) loginSession(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	var u model.User
	err := r.db.Where("username = ?", req.Username).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		fail(c, 401, "用户名或密码错误")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		fail(c, 401, "用户名或密码错误")
		return
	}
	access, err := r.jwt.Sign(u.ID, u.Username, u.Role)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, gin.H{
		"accessToken":  access,
		"refreshToken": access, // 简化：refresh 同 access（P5 可拆分）
		"expiresIn":    int(r.jwt.TTL().Seconds()),
	})
}

// refreshSession POST /api/auth/tokens/refresh —— 刷新 token。
func (r *Router) refreshSession(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	_ = c.ShouldBindJSON(&req)
	claims, err := r.jwt.Parse(req.RefreshToken)
	if err != nil {
		fail(c, 401, "invalid token")
		return
	}
	var u model.User
	if err := r.db.First(&u, claims.UserID).Error; err != nil {
		fail(c, 401, "user not found")
		return
	}
	access, _ := r.jwt.Sign(u.ID, u.Username, u.Role)
	ok(c, gin.H{
		"accessToken":  access,
		"refreshToken": access,
		"expiresIn":    int(r.jwt.TTL().Seconds()),
	})
}

// logoutSession DELETE /api/auth/sessions —— 登出。
func (r *Router) logoutSession(c *gin.Context) {
	ok(c, nil)
}

// me GET /api/auth/users/me —— 当前用户信息（vue-admin UserProfile 格式）。
func (r *Router) me(c *gin.Context) {
	uid := c.GetUint64("uid")
	var u model.User
	if err := r.db.First(&u, uid).Error; err != nil {
		fail(c, 404, "user not found")
		return
	}
	ok(c, toProfile(&u))
}

// toProfile 转换为 vue-admin UserProfile：{id, username, nickname, roles[], permissions[]}。
func toProfile(u *model.User) gin.H {
	roles := []string{"user"}
	perms := []string{}
	if u.Role == 1 { // 管理员
		roles = []string{"super_admin"}
		perms = []string{"*"}
	} else if u.Permission != "" {
		for _, p := range strings.Split(u.Permission, ",") {
			if p = strings.TrimSpace(p); p != "" {
				perms = append(perms, p)
			}
		}
	}
	return gin.H{
		"id":          u.ID,
		"username":    u.Username,
		"nickname":    u.Username,
		"roles":       roles,
		"permissions": perms,
	}
}
