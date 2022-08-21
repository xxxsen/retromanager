package server

import (
	"retromanager/codec"
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

type ProcessFunc func(c *gin.Context, req interface{}) (int, errs.IError, interface{})
type RequestCreatorFunc func() interface{}

type IHandler interface {
	Request() interface{}
	Codec() codec.ICodec
	Process(c *gin.Context, req interface{}) (int, errs.IError, interface{})
}

type HandlerRegisterFunc func(engine *gin.Engine)

type DefaultHandler struct {
	reqc  RequestCreatorFunc
	codec codec.ICodec
	pfunc ProcessFunc
}

func NewHandler(reqc RequestCreatorFunc, codec codec.ICodec, proc ProcessFunc) *DefaultHandler {
	return &DefaultHandler{
		reqc:  reqc,
		codec: codec,
		pfunc: proc,
	}
}

func (c *DefaultHandler) Request() interface{} {
	return c.reqc()
}

func (c *DefaultHandler) Codec() codec.ICodec {
	return c.codec
}

func (c *DefaultHandler) Process(ctx *gin.Context, req interface{}) (int, errs.IError, interface{}) {
	return c.pfunc(ctx, req)
}
