package codec

import (
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

var NopCodec = &nopCodec{}

type nopCodec struct {
}

func (c *nopCodec) Decode(ctx *gin.Context, request interface{}) errs.IError {
	return nil
}

func (c *nopCodec) Encode(ctx *gin.Context, statuscode int, err errs.IError, response interface{}) errs.IError {
	return nil
}
