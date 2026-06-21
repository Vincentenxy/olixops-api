package application

import (
	"github.com/gin-gonic/gin"

	"olixops/pkg/httpx"
	"olixops/pkg/pagination"
)

// Handler 处理应用相关 HTTP 请求。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/applications")
	g.POST("", h.create)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)

	g.GET("/:id/components", h.listComponents)
	g.POST("/:id/components", h.addComponent)
	g.DELETE("/:id/components/:component_id", h.deleteComponent)

	// 项目下的子资源
	pg := r.Group("/projects/:project_id/applications")
	pg.GET("", h.listByProject)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	a, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, a)
}

func (h *Handler) list(c *gin.Context) {
	q := pagination.FromGin(c)
	filter := ListFilter{
		ProjectID: c.Query("project_id"),
		Kind:      Kind(c.Query("kind")),
		Status:    Status(c.Query("status")),
		OwnerID:   c.Query("owner_id"),
	}
	items, total, err := h.svc.List(c.Request.Context(), q, filter)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Paged(c, items, total, q.Page, q.PageSize)
}

func (h *Handler) listByProject(c *gin.Context) {
	q := pagination.FromGin(c)
	filter := ListFilter{
		ProjectID: c.Param("project_id"),
		Kind:      Kind(c.Query("kind")),
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
	a, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, a)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	a, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, a)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) listComponents(c *gin.Context) {
	items, err := h.svc.ListComponents(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, items)
}

func (h *Handler) addComponent(c *gin.Context) {
	var in ComponentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, err)
		return
	}
	cmpt, err := h.svc.AddComponent(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, cmpt)
}

func (h *Handler) deleteComponent(c *gin.Context) {
	if err := h.svc.DeleteComponent(c.Request.Context(), c.Param("component_id")); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
