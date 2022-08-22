package idgen

import (
	gen "github.com/yitter/idgenerator-go/idgen"
)

func Init(wrkid uint16) error {
	opt := gen.NewIdGeneratorOptions(wrkid)
	gen.SetIdGenerator(opt)
	return nil
}

func NextId() uint64 {
	return uint64(gen.NextId())
}
