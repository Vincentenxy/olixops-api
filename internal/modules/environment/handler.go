package environment

import (
	"github.com/gin-gonic/gin"

	"olixops/pkg/httpx"
	"olixops/pkg/pagination"
)

// Handler 处理环境相关 HTTP 请求。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册路由,挂在项目路由下。
func (h *Handler) Register(r *gin.RouterGroup) {
	// 全局列表入口
	r.GET("/environments/:id", h.get)
	r.PUT("/environments/:id", h.update)
	r.DELETE("/environments/:id", h.delete)

	// 项目下的子资源
	g := r.Group("/projects/:project_id/environments")
	g.POST("", h.create)
	g.GET("", h.list)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	in.ProjectID = c.Param("project_id")
	e, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, e)
}

func (h *Handler) list(c *gin.Context) {
	q := pagination.FromGin(c)
	filter := ListFilter{
		ProjectID: c.Param("project_id"),
		Type:      Type(c.Query("type")),
		Status:    Status(c.Query("status")),
	}
	items, total, err := h.svc.List(c.Request.Context(), q, filter)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Paged(c, items, total, q.Page, q.PageSize)
}

func (h *Handler) get(c *gin.Context) {
	e, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, e)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	e, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, e)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
