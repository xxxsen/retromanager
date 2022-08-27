package retry

import (
	"context"
	"time"
)

type RetryDoFunc func(ctx context.Context) error

func RetryDo(ctx context.Context, repeat uint32, delimis time.Duration, fn RetryDoFunc) error {
	repeat += 1
	var err error
	for i := 0; i < int(repeat); i++ {
		if err = fn(ctx); err != nil {
			time.Sleep(delimis)
			continue
		}
		break
	}
	if err != nil {
		return err
	}
	return nil
}
