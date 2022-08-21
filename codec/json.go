package codec

import (
	"retromanager/constants"
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

var JsonCodec = &jsonCodec{}

type jsonCodec struct {
}

type jsonMessageFrame struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (c *jsonCodec) Decode(gctx *gin.Context, request interface{}) errs.IError {
	if err := gctx.ShouldBindJSON(request); err != nil {
		return errs.Wrap(constants.ErrUnmarshal, "decode err", err)
	}
	return nil
}

func (c *jsonCodec) Encode(gctx *gin.Context, statuscode int, err errs.IError, response interface{}) errs.IError {
	frame := &jsonMessageFrame{}
	frame.Code = err.Code()
	frame.Message = err.Message()
	frame.Data = response
	gctx.JSON(statuscode, frame)
	return nil
}
