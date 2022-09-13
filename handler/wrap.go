package handler

import (
	"net/http"
	"reflect"

	"github.com/xxxsen/common/errs"
	"github.com/xxxsen/common/logutil"
	"github.com/xxxsen/common/naivesvr"
	"github.com/xxxsen/common/naivesvr/codec"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func WrapHandler(ptr interface{}, cc codec.ICodec, pfunc naivesvr.ProcessFunc) gin.HandlerFunc {
	creator := func() interface{} {
		if ptr == nil {
			return nil
		}
		typ := reflect.TypeOf(ptr)
		val := reflect.New(typ.Elem())
		return val.Interface()
	}
	return wrapHandler(naivesvr.NewHandler(creator, cc, pfunc))
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

func wrapHandler(h naivesvr.IHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var statuscode int
		var derr error
		var perr error
		var eerr error
		var rsp interface{}

		defer func() {
			derr = wrapErr(derr)
			perr = wrapErr(perr)
			eerr = wrapErr(eerr)
			step, err, exist := finderr("decode", derr, "proc", perr, "encode", eerr)
			logger := logutil.GetLogger(ctx).With(
				zap.String("method", ctx.Request.Method),
				zap.Int("statuscode", statuscode),
				zap.String("path", ctx.Request.URL.Path),
			)
			writer := logger.Info
			if exist {
				logger = logger.With(
					zap.String("step", step),
					zap.Int("code", int(err.Code())),
					zap.String("msg", err.Message()),
					zap.Error(err),
				)
				writer = logger.Error
			}
			writer("process msg finish")
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

func wrapErr(err error) errs.IError {
	if err == nil {
		return errs.ErrOK
	}
	return errs.FromError(err)
}
