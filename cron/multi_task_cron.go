package cron

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xxxsen/errs"
)

func init() {
	//每5分钟, 刷新当前时间往前偏移30分钟的数据到es, 最多2000条
	cr := NewMultiTaskCron(
		newCleanRemovedRecordCron(),
		newRefreshESCron(),
	)
	Regist("*/5 * * * *", cr)
}

type MultiTaskCron struct {
	lst []ICron
}

func NewMultiTaskCron(cr ...ICron) *MultiTaskCron {
	return &MultiTaskCron{
		lst: cr,
	}
}

func (c *MultiTaskCron) Name() string {
	names := bytes.NewBuffer(nil)
	for _, item := range c.lst {
		names.WriteString(item.Name())
	}
	return fmt.Sprintf("multi_task:[%s]", names.String())
}

func (c *MultiTaskCron) Run(ctx context.Context) error {
	var retErr error
	var name string
	for _, tk := range c.lst {
		err := tk.Run(ctx)
		if err != nil {
			retErr = err
			name = tk.Name()
		}
	}
	if retErr != nil {
		return errs.Wrap(errs.ErrServiceInternal, "task exec fail", retErr).WithDebugMsg("taskname:%s", name)
	}
	return nil
}
