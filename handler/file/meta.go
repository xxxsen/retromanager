package file

import (
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

const (
	maxListMetaSizePerRequest = 20
)

func Meta(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	//TODO: finish it
	panic(1)
}
