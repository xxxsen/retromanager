package handler

import (
	"fmt"
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
		if ptr == nil {
			return nil
		}
		typ := reflect.TypeOf(ptr)
		val := reflect.New(typ.Elem())
		return val.Interface()
	}
	return wrapHandler(server.NewHandler(creator, cc, pfunc))
}

func finderr(args ...interface{}) (string, errs.IError, bool) {
	for i := 0; i < len(args); i += 2 {
		name := args[i].(string)
		err := args[i+1].(errs.IError)
		if errs.IsErrOK(err) {
			continue
		}
		return name, err, true
	}
	return "", nil, false
}

func wrapHandler(h server.IHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var statuscode int
		var derr errs.IError
		var perr errs.IError
		var eerr errs.IError
		var rsp interface{}

		defer func() {
			derr = wrapErr(derr)
			perr = wrapErr(perr)
			eerr = wrapErr(eerr)
			step, err, exist := finderr("decode", derr, "proc", perr, "encode", eerr)
			msg := fmt.Sprintf("serve request finish, path:%s, statuscode:%d", ctx.Request.URL.Path, statuscode)
			writer := log.Infof
			if exist {
				msg += fmt.Sprintf(", err:[step:%s, code:%d, msg:%s, detail:%s]", step, err.Code(), err.Message(), err.Error())
				writer = log.Errorf
			}
			writer(msg)
		}()
		req := h.Request()
		c := h.Codec()
		if req != nil {
			derr = c.Decode(ctx, req)
		}
		if !errs.IsErrOK(derr) {
			writeJson(ctx, wrapErr(derr))
			return
		}
		statuscode, perr, rsp = h.Process(ctx, req)
		eerr = c.Encode(ctx, statuscode, wrapErr(perr), rsp)
	}
}

func writeJson(ctx *gin.Context, err errs.IError) {
	m := make(map[string]interface{})
	m["code"] = err.Code()
	m["message"] = err.Message()
	ctx.AbortWithStatusJSON(http.StatusOK, m)
}

func wrapErr(err errs.IError) errs.IError {
	if err == nil {
		return errs.ErrOK
	}
	return err
}
