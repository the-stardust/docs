package middlewares

import (
	"github.com/gin-gonic/gin"
	"interview/common"
)

type Recovery struct {
	Middleware
}

func (sf *Recovery) HandlerFunc(c *gin.Context, err interface{}) {
	if err != nil {
		sf.Error(common.CodeServerBusy, c, "异常panic!")
		return
	}

}
