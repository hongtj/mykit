package router

import (
	"mykit/core/transfer"

	"github.com/gin-gonic/gin"
)

func Init(e *gin.Engine) {
	transfer.RegRouter(e,
		apiGroup,
	)
}

func apiGroup(e *gin.Engine) {
	group := e.Group("/api/:app/:method")
	{
		group.POST("", nil)
	}
}
