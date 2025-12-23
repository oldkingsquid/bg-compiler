package processor

import (
	"fmt"

	"github.com/oldkingsquid/bg-compiler/flags"
	"github.com/sirupsen/logrus"
)

var JobChan = make(chan *Job, flags.JobChannelLength())

type Worker struct {
	ID       string
	Compiles int64
	Errors   int64
}

func InitWorkers() {
	workers := flags.WorkerCount()
	logrus.Infof("Initializing %v workers", workers)

	for i := 0; i < workers; i++ {
		go NewWorker(fmt.Sprintf("%v", i)).Start()
	}
}

func NewWorker(id string) *Worker {
	return &Worker{
		ID:       id,
		Compiles: 0,
	}
}

func (w *Worker) Start() {
	for job := range JobChan {
		if err := w.RunJob(job); err != nil {
			logrus.WithError(err).Errorf("Error running job.")
			w.Errors++
		}
		w.Compiles++
	}
}

func (w *Worker) RunJob(job *Job) error {
	// TODO: There needs to be better error handling here.
	return job.Start()
}
