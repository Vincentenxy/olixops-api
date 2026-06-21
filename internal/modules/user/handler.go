package user

import (
	"github.com/gin-gonic/gin"

	"olixops/internal/platform/audit"
	"olixops/internal/platform/auth"
	"olixops/pkg/httpx"
	"olixops/pkg/pagination"
)

// Handler 处理用户相关 HTTP 请求。
type Handler struct {
	svc      *Service
	recorder audit.Recorder
}

// NewHandler 构造 handler。
func NewHandler(svc *Service, recorder audit.Recorder) *Handler {
	return &Handler{svc: svc, recorder: recorder}
}

// RegisterPub 注册免鉴权路由 (login / refresh)。
// 由 internal/modules/user/module.go 在 RegisterPub 中按路径挂载, 这里不直接挂载。
func (h *Handler) login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	in.IP = c.ClientIP()

	res, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		h.recordFailure(c, "user.login.failure", in.Username, err.Error())
		httpx.Error(c, err)
		return
	}
	h.recordSuccess(c, "user.login.success", res.User.ID, res.User.Username)
	httpx.OK(c, res)
}

func (h *Handler) refresh(c *gin.Context) {
	var in RefreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
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
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) logout(c *gin.Context) {
	sub, _ := auth.FromContext(c.Request.Context())
	_ = h.svc.Logout(c.Request.Context(), "")
	h.recordSuccess(c, "user.logout", sub.UserID, sub.Username)
	httpx.NoContent(c)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	u, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, u)
}

func (h *Handler) list(c *gin.Context) {
	q := pagination.FromGin(c)
	filter := ListFilter{
		Status: Status(c.Query("status")),
		Source: c.Query("source"),
	}
	items, total, err := h.svc.List(c.Request.Context(), q, filter)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Paged(c, items, total, q.Page, q.PageSize)
}

func (h *Handler) get(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	u, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, u)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) changePassword(c *gin.Context) {
	var in ChangePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), c.Param("id"), in); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
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
