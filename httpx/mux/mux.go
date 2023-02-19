package mux

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Resource struct {
	Logger              *zap.Logger
	ProjectListen       string
	Env                 string
	HeaderLoginToken    string
	HeaderSignToken     string
	HeaderSignTokenDate string
}

var _ IMux = (*Mux)(nil)

// IMux http Mux
type IMux interface {
	http.Handler
	GetEngine() *gin.Engine
	Group(relativePath string, handlers ...HandlerFunc) RouterGroup
}

type Mux struct {
	Engine *gin.Engine
}

func (m *Mux) GetEngine() *gin.Engine {
	return m.Engine
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.Engine.ServeHTTP(w, req)
}

func (m *Mux) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	return &router{
		group: m.Engine.Group(relativePath, WrapHandlers(handlers...)...),
	}
}

// RouterGroup 包装gin的RouterGroup
type RouterGroup interface {
	Group(string, ...HandlerFunc) RouterGroup
	IRoutes
}

var _ IRoutes = (*router)(nil)

// IRoutes 包装gin的IRoutes
type IRoutes interface {
	Any(string, ...HandlerFunc)
	GET(string, ...HandlerFunc)
	POST(string, ...HandlerFunc)
	DELETE(string, ...HandlerFunc)
	PATCH(string, ...HandlerFunc)
	PUT(string, ...HandlerFunc)
	OPTIONS(string, ...HandlerFunc)
	HEAD(string, ...HandlerFunc)
}

type router struct {
	group *gin.RouterGroup
}

func (r *router) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	group := r.group.Group(relativePath, WrapHandlers(handlers...)...)
	return &router{group: group}
}

func (r *router) Any(relativePath string, handlers ...HandlerFunc) {
	r.group.Any(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) GET(relativePath string, handlers ...HandlerFunc) {
	r.group.GET(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) POST(relativePath string, handlers ...HandlerFunc) {
	r.group.POST(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) DELETE(relativePath string, handlers ...HandlerFunc) {
	r.group.DELETE(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) PATCH(relativePath string, handlers ...HandlerFunc) {
	r.group.PATCH(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) PUT(relativePath string, handlers ...HandlerFunc) {
	r.group.PUT(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	r.group.OPTIONS(relativePath, WrapHandlers(handlers...)...)
}

func (r *router) HEAD(relativePath string, handlers ...HandlerFunc) {
	r.group.HEAD(relativePath, WrapHandlers(handlers...)...)
}

func WrapHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler
		funcs[i] = func(c *gin.Context) {
			ctx := NewContext(c)
			defer ReleaseContext(ctx)

			handler(ctx)
		}
	}

	return funcs
}
