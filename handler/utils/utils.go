package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"retromanager/constants"
	"retromanager/handler/config"

	"github.com/gin-gonic/gin"
)

func MustGetConfig(ctx *gin.Context) *config.Config {
	c, exist := ctx.Get(constants.KeyConfigAttach)
	if !exist {
		panic("config not exist")
	}
	return c.(*config.Config)
}

func ReadAndCloseFile(r io.ReadCloser) ([]byte, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}
	return raw, nil
}

func CalcMd5(raw []byte) string {
	m := md5.New()
	_, _ = m.Write(raw)
	return hex.EncodeToString(m.Sum(nil))
}
