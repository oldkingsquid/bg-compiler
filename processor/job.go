package processor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/oldkingsquid/bg-compiler/docker"
	"github.com/oldkingsquid/bg-compiler/flags"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Job is the internal representation of a job that
// needs to be run by a worker. The output of this job will be
// entry.JobOutput.
type Job struct {
	ID  string // incremented based on the number of jobs in the definition.
	Def *definition

	/** Compiler specific fields **/
	logger      *logrus.Entry
	containerID string

	/** Output fields **/
	StdOut   *LogWriter
	StdErr   *LogWriter
	TimedOut bool
	Output   *JobOutput // Output field for the Job.

	done chan struct{} // signal that the job is done and cleaned.
}

type JobOutput struct {
	Job *Job `json:"-"`

	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
	Duration int64  `json:"duration_ms"`
	TimedOut bool   `json:"timed_out"`
}

func (job *Job) Start() error {
	defer job.Def.wg.Done()

	ctx, cancel := context.WithDeadline(job.Def.ctx,
		time.Now().Add(flags.ContainerMaxDuration()))

	logs, err := docker.Client.StartContainer(ctx, job.containerID)
	if err != nil {
		cancel()
		return errors.Wrap(err, "Failed to start container")
	}
	start := time.Now()

	if err := job.MaybeFeedStdIn(); err != nil {
		job.logger.WithError(err).Errorf("Error feeding StdIn")
	}

	go job.WaitForContainerExit(ctx)

	job.StdOut, job.StdErr, err = ReadLogOutputs(logs)
	if err != nil {
		cancel()
		return errors.Wrap(err, "Failed to read log outputs")
	}
	cancel()
	<-job.done

	job.Output = &JobOutput{
		Job:      job,
		StdOut:   job.StdOut.String(),
		StdErr:   job.StdErr.String(),
		Duration: time.Since(start).Milliseconds(),
		TimedOut: job.TimedOut,
	}

	return nil
}

func (job *Job) CreateContainer(ctx context.Context) error {
	fullCmd := fmt.Sprintf(
		"%s %s",
		job.Def.Submission.Cmd,
		filepath.Base(job.Def.hostSourceCodeFile))

	var err error
	job.containerID, err = docker.Client.CreateContainer(ctx,
		&docker.CreateContainerInput{
			ID:          fmt.Sprintf("bg_%s_%s", job.Def.id, job.ID),
			FullCommand: fullCmd,
			Mounts:      job.Def.mounts,
			Image:       job.Def.Submission.Image,
		})

	return errors.Wrap(err, "Error creating container")
}

func (job *Job) WaitForContainerExit(ctx context.Context) {
	<-ctx.Done()
	timedout, err := docker.Client.KillContainer(context.Background(),
		job.containerID)
	if err != nil && !strings.Contains(err.Error(), "No such container") {
		job.logger.WithError(err).Errorf("Failed to kill container")
	} else if timedout {
		job.TimedOut = true
	}
	job.done <- struct{}{}
}

func (job *Job) MaybeFeedStdIn() error {
	if job.Def.Submission.StdIn == nil {
		return nil
	}
	return docker.Client.FeedStdIn(job.Def.ctx, job.containerID, *job.Def.Submission.StdIn)
}
