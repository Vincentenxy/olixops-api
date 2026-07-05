package user

import (
	"errors"
	"olixops/internal/platform/audit"
	"olixops/internal/platform/auth"
	"olixops/internal/platform/logger"
	"olixops/pkg/httpx"
	"olixops/pkg/pagination"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RegisterInput struct {
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	Email           string `json:"email" binding:"required"`
	DisplayName     string `json:"displayName" binding:"required"`
}

// Handler 处理用户相关 HTTP 请求
type Handler struct {
	svc           *Service
	recorder      audit.Recorder
	cookieManager *auth.CookieManager
}

// NewHandler 构造 handler。
func NewHandler(svc *Service, recorder audit.Recorder, cookieManager *auth.CookieManager) *Handler {
	return &Handler{
		svc:           svc,
		recorder:      recorder,
		cookieManager: cookieManager,
	}
}

// RegisterPub 注册免鉴权路由 (login / refresh)。
// 由 internal/modules/user/module.go 在 RegisterPub 中按路径挂载, 这里不直接挂载。
func (h *Handler) login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, errors.New("invalid input params"))
		return
	}
	in.IP = c.ClientIP()

	res, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		h.recordFailure(c, audit.UserLoginFailure, in.Username, err.Error())
		httpx.Fail(c, err)
		return
	}

	h.setAuthCookie(c, &res)
	h.recordSuccess(c, audit.UserLoginSuccess, res.User.ID, res.User.Username)
	httpx.OK(c, res)
}

func (h *Handler) setAuthCookie(c *gin.Context, res *LoginResult) {

	// set jwt into cookie
	h.cookieManager.SetWithExpires(
		c,
		"auth_token",
		res.Tokens.AccessToken,
		res.Tokens.AccessTokenExpiresAt,
	)
}

func (h *Handler) register(c *gin.Context) {
	log := logger.L()
	var req RegisterInput
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("invalid input prams: ", zap.Error(err))
		httpx.Fail(c, errors.New("参数解析失败！"))
		return
	}

	if req.Password != req.ConfirmPassword {
		httpx.Fail(c, errors.New("密码与确认密码必须一致！"))
		return
	}

	user, err := h.svc.Create(c, CreateInput{
		Username:    req.Username,
		Password:    req.Password,
		Email:       req.Email,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		h.recordFailure(c, audit.UserCreateFailure, req.Username, err.Error())
		httpx.Fail(c, err)
		return
	}

	httpx.OK(c, user)
}

func (h *Handler) refresh(c *gin.Context) {
	var in RefreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, err)
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.recordSuccess(c, "user.refresh", "", "")
	httpx.OK(c, pair)
}

// RegisterPrivate 注册需鉴权路由 (me / logout + CRUD)。
// 由 internal/modules/user/module.go 在 RegisterPrivate 中按路径挂载。
func (h *Handler) me(c *gin.Context) {
	u, err := h.svc.Me(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) logout(c *gin.Context) {
	sub, _ := auth.FromContext(c.Request.Context())
	_ = h.svc.Logout(c.Request.Context(), "")
	h.recordSuccess(c, "user.logout", sub.UserID, sub.Username)
	httpx.OK[any](c, nil)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, err)
		return
	}
	u, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) list(c *gin.Context) {
	var filter ListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		httpx.Fail(c, err)
	}

	items, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK[pagination.PageResult[*User]](
		c,
		pagination.PageResult[*User]{
			Total: total,
			List:  items,
		})
	return
}

func (h *Handler) get(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, err)
		return
	}
	u, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK[any](c, nil)
}

func (h *Handler) changePassword(c *gin.Context) {
	var in ChangePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), c.Param("id"), in); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK[any](c, nil)
}

func (h *Handler) recordSuccess(c *gin.Context, action, actorID, actorName string) {
	h.recorder.Record(c.Request.Context(), audit.Event{
		Action:    action,
		Actor:     actorID,
		ActorName: actorName,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Status:    "success",
	})
}

func (h *Handler) recordFailure(c *gin.Context, action, actorName, reason string) {
	h.recorder.Record(c.Request.Context(), audit.Event{
		Action:    action,
		ActorName: actorName,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Status:    "failure",
		Reason:    reason,
	})
}
