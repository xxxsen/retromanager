package handler

import (
	"net/http"
	"reflect"
	"retromanager/codec"
	"retromanager/errs"
	"retromanager/server"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/log"
)

func WrapHandler(ptr interface{}, cc codec.ICodec, pfunc server.ProcessFunc) gin.HandlerFunc {
	creator := func() interface{} {
		typ := reflect.TypeOf(ptr)
		val := reflect.ValueOf(typ.Elem())
		return val.Interface()
	}
	return wrapHandler(server.NewHandler(creator, cc, pfunc))
}

func wrapHandler(h server.IHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err errs.IError = errs.ErrOK
		var statuscode int
		var rsp interface{}
		defer func() {
			log.Debugf("serve request finish, path:%s, statuscode:%d, busi code:%d, msg:%s",
				ctx.Request.URL.Path, statuscode, err.Code(), err.Error())
		}()
		req := h.Request()
		c := h.Codec()
		if req != nil {
			err = c.Decode(ctx, req)
		}
		if !errs.IsErrOK(err) {
			writeJson(ctx, err)
			log.Errorf("decode request fail, path:%s, err:%v", ctx.Request.URL.Path, err)
			return
		}
		statuscode, err, rsp = h.Process(ctx, req)
		if !errs.IsErrOK(err) {
			log.Errorf("process request fail, path:%s, err:%v", ctx.Request.URL.Path, err)
		}
		err = c.Encode(ctx, statuscode, err, rsp)
		if !errs.IsErrOK(err) {
			log.Errorf("encode request fail, path:%s, err:%w", ctx.Request.URL.Path, err)
			//这里就不回包了吧。。。
		}
	}
}

func writeJson(ctx *gin.Context, err errs.IError) {
	ctx.AbortWithStatusJSON(http.StatusOK, err)
}
