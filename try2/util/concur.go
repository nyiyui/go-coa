package util

import (
	"context"
	"fmt"
	"runtime/pprof"
)

func Go(labels map[string]interface{}, f func()) {
	slice := make([]string, 0)
	for key, val := range labels {
		slice = append(slice, key, fmt.Sprint(val))
	}
	pprof.Do(context.Background(), pprof.Labels(slice...), func(_ context.Context) {
		f()
	})
}
