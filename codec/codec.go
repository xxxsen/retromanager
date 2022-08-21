package codec

import (
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

type ICodec interface {
	Decode(c *gin.Context, request interface{}) errs.IError
	Encode(c *gin.Context, statuscode int, err errs.IError, response interface{}) errs.IError
}
