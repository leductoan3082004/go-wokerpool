package main

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	workerpool "go-workerpool/worker-pool"
	"time"
)

func main() {
	wp := workerpool.NewWorkerPool(5)

	err := wp.Start(context.Background())

	handler := func(index int) workerpool.Job {
		return func(ctx context.Context) error {
			if index%2 == 0 {
				return errors.New("index is even")
			} else {
				return errors.New("index is odd")
			}
		}
	}

	job := make([]workerpool.Job, 10)
	for i := 0; i < 10; i++ {
		job[i] = handler(i)
	}

	go wp.Dispatch(job...)

	//defer func() {
	//	time.Sleep(5 * time.Second)
	//	wp.Stop()
	//}()

	go func() {
		for err := range err {
			if err != nil {
				logrus.Errorln(err)
			}
		}
	}()

	defer func() {
		time.Sleep(5 * time.Second)
		wp.Stop()
	}()
}
