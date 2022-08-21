package codec

import (
	"retromanager/constants"
	"retromanager/errs"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"
)

var QueryCodec = &queryCodec{}

type queryCodec struct {
}

func (c *queryCodec) Decode(ctx *gin.Context, request interface{}) errs.IError {
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(request, ctx.Request.URL.Query()); err != nil {
		return errs.Wrap(constants.ErrParam, "invalid query", err)
	}
	return nil
}

func (c *queryCodec) Encode(ctx *gin.Context, statuscode int, err errs.IError, response interface{}) errs.IError {
	return nil
}
