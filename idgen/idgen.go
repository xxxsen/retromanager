package idgen

import (
	gen "github.com/yitter/idgenerator-go"
)

func Init(wrkid uint16) error {
	gen.SetOptions(wrkid)
	return nil
}

func NextId() uint64 {
	return uint64(gen.NextId())
}
