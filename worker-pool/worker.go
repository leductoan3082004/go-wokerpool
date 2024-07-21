package workerpool

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

type Job func(ctx context.Context) error

type WorkerPool interface {
	Start(ctx context.Context) <-chan error
	Dispatch(jobs ...Job)
	Stop()
}

const defaultNumWorker = 10

type workerPool struct {
	wg        sync.WaitGroup
	jobChan   chan Job
	numWorker int
	errChan   chan error
	quit      chan struct{}
}

func NewWorkerPool(numWorker int) WorkerPool {
	if numWorker <= 0 {
		numWorker = defaultNumWorker
	}
	wp := &workerPool{
		jobChan:   make(chan Job),
		numWorker: numWorker,
		errChan:   make(chan error),
		wg:        sync.WaitGroup{},
		quit:      make(chan struct{}),
	}

	return wp
}

func (wp *workerPool) Dispatch(jobs ...Job) {
	go func() {
		for _, j := range jobs {
			select {
			case <-wp.quit:
				return
			case wp.jobChan <- j:
			}
		}
	}()
}

func (wp *workerPool) recover() {
	go func() {
		for {
			if err := recover(); err != nil {
				logrus.Errorln(err)
			}
		}
	}()
}

func (wp *workerPool) work(ctx context.Context, id int) {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.quit:
			fmt.Println("Worker", id, "is done")
			return
		case job, ok := <-wp.jobChan:
			if ok {
				logrus.Infoln("Worker", id, "is processing job")
				if err := job(ctx); err != nil {
					wp.errChan <- err
				}
			} else {
				return
			}
		}
	}
}
func (wp *workerPool) Start(ctx context.Context) <-chan error {
	wp.recover()
	wp.wg.Add(wp.numWorker)
	logrus.Infoln("Starting worker pool")

	for i := 0; i < wp.numWorker; i++ {
		go wp.work(ctx, i)
	}

	logrus.Infoln("Worker pool started")
	return wp.errChan
}

func (wp *workerPool) Stop() {
	logrus.Infoln("Stopping worker pool")
	close(wp.quit)
	wp.wg.Wait()
	close(wp.jobChan)
	close(wp.errChan)
}
