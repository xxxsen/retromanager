package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func Md5Raw(raw []byte) []byte {
	m := md5.New()
	_, _ = m.Write(raw)
	return m.Sum(nil)
}

func CalcMd5(raw []byte) string {
	return hex.EncodeToString(Md5Raw(raw))
}

func CalcMd5Base64(raw []byte) string {
	return base64.StdEncoding.EncodeToString(Md5Raw(raw))
}

func Sha256Raw(raw []byte) []byte {
	m := sha256.New()
	_, _ = m.Write(raw)
	return m.Sum(nil)
}

func CalcSha256Hex(raw []byte) string {
	return hex.EncodeToString(Sha256Raw(raw))
}
