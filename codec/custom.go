package codec

import (
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

type customCodec struct {
	enc ICodec
	dec ICodec
}

func CustomCodec(enc, dec ICodec) *customCodec {
	return &customCodec{
		enc: enc,
		dec: dec,
	}
}

func (c *customCodec) Decode(ctx *gin.Context, request interface{}) errs.IError {
	return c.dec.Decode(ctx, request)
}

func (c *customCodec) Encode(ctx *gin.Context, statuscode int, err errs.IError, response interface{}) errs.IError {
	return c.enc.Encode(ctx, statuscode, err, response)
}
