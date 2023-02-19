package mux

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

type SpecialContextResource struct {
	UserID   int64
	UserName string
	TenantID int64
	IsAdmin  bool
	RoleType int32
}

func CreateSpecialContext(r SpecialContextResource) Context {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx := &context{ctx: ginCtx}
	ctx.setUserID(r.UserID)
	ctx.setUserName(r.UserName)
	ctx.setTenantID(r.TenantID)
	is := 0
	if r.IsAdmin {
		is = 1
	}
	ctx.setIsAdmin(is)
	ctx.setRoleType(r.RoleType)
	return ctx
}
