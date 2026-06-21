package project

import (
	"github.com/gin-gonic/gin"

	"olixops/pkg/httpx"
	"olixops/pkg/pagination"
)

// Handler 处理项目相关 HTTP 请求。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/projects")
	g.POST("", h.create)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)

	g.GET("/:id/members", h.listMembers)
	g.POST("/:id/members", h.addMember)
	g.DELETE("/:id/members/:user_id", h.removeMember)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	p, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, p)
}

func (h *Handler) list(c *gin.Context) {
	q := pagination.FromGin(c)
	filter := ListFilter{
		OwnerID:    c.Query("owner_id"),
		Status:     Status(c.Query("status")),
		Visibility: Visibility(c.Query("visibility")),
	}
	items, total, err := h.svc.List(c.Request.Context(), q, filter)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Paged(c, items, total, q.Page, q.PageSize)
}

func (h *Handler) get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, p)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	p, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, p)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) listMembers(c *gin.Context) {
	ms, err := h.svc.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, ms)
}

func (h *Handler) addMember(c *gin.Context) {
	var in AddMemberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	m, err := h.svc.AddMember(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, m)
}

func (h *Handler) removeMember(c *gin.Context) {
	if err := h.svc.RemoveMember(c.Request.Context(), c.Param("id"), c.Param("user_id")); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
