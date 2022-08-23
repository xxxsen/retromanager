package codec

import (
	"retromanager/constants"
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

var MultipartCodec = &multipartCodec{}

type multipartCodec struct {
	nopCodec
}

func (c *multipartCodec) Decode(ctx *gin.Context, request interface{}) errs.IError {
	if err := ctx.ShouldBind(request); err != nil {
		return errs.Wrap(constants.ErrParam, "decode multipart fail", err)
	}
	return nil
}
