package codec

import (
	"retromanager/constants"
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

var QueryCodec = &queryCodec{}

type queryCodec struct {
	nopCodec
}

func (c *queryCodec) Decode(ctx *gin.Context, request interface{}) errs.IError {
	if err := ctx.ShouldBind(request); err != nil {
		return errs.Wrap(constants.ErrParam, "bind query fail", err)
	}
	return nil
}
