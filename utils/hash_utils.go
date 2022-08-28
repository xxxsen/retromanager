package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
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

func CalcMd5Reader(r io.Reader) (string, error) {
	hash := md5.New()
	_, err := io.Copy(hash, r)

	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func CalcMd5File(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return CalcMd5Reader(f)
}
