package worker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/progressbar"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type Runner func(Job) (any, error)

func NewGenericWorker(log *logger.Logger, cfg *config.Config, iostream *iostreams.IOStreams, client *c8y.Client, activityLog *activitylogger.ActivityLogger, runFunc Runner, checkError func(error) error) (*GenericWorker, error) {
	return &GenericWorker{
		Config:         cfg,
		Logger:         log,
		IO:             iostream,
		ActivityLogger: activityLog,
		Client:         client,
		Execute:        runFunc,
		CheckError:     checkError,
	}, nil
}

type GenericWorker struct {
	Config         *config.Config
	IO             *iostreams.IOStreams
	Logger         *logger.Logger
	Client         *c8y.Client
	ActivityLogger *activitylogger.ActivityLogger
	CheckError     func(error) error
	Execute        Runner
}

// GetMaxWorkers maximum number of workers
func (w *GenericWorker) GetMaxWorkers() int {
	if w.Config == nil {
		return 5
	}
	return w.Config.GetMaxWorkers()
}

// GetMaxJob maximum number of jobs allowed
func (w *GenericWorker) GetMaxJobs() int64 {
	if w.Config == nil {
		return 100
	}
	return w.Config.GetMaxJobs()
}

func (w *GenericWorker) GetBatchOptions(cmd *cobra.Command) (*BatchOptions, error) {
	options := &BatchOptions{
		AbortOnErrorCount: w.Config.AbortOnErrorCount(),
		TotalWorkers:      w.Config.GetWorkers(),
		Delay:             w.Config.WorkerDelay(),
		DelayBefore:       w.Config.WorkerDelayBefore(),
		SemanticMethod:    flags.GetSemanticMethodFromAnnotation(cmd),
	}

	if v, err := cmd.Flags().GetInt("count"); err == nil {
		options.NumJobs = v
	}

	if v, err := cmd.Flags().GetInt("startIndex"); err == nil {
		options.StartIndex = v
	}

	return options, nil
}

type Job struct {
	ID            int64
	Value         any
	CommonOptions config.CommonCommandOptions
	Input         any
	Options       BatchOptions
}

func (w *GenericWorker) RunSequentially(cmd *cobra.Command, iter iterator.Iterator, inputIterators *flags.RequestInputIterators) error {
	// TODO: How does an unbound iterator get caught here?
	if inputIterators == nil {
		return fmt.Errorf("missing input iterators")
	}

	// get common options and batch settings
	commonOptions, err := w.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	batchOptions, err := w.GetBatchOptions(cmd)
	if err != nil {
		return err
	}

	// Configure post actions
	batchOptions.PostActions = inputIterators.PipeOptions.PostActions

	out, input, err := iter.GetNext()
	if err != nil {
		return err
	}
	job := Job{
		ID:            1,
		Value:         out,
		CommonOptions: commonOptions,
		Options:       *batchOptions,
		Input:         input,
	}
	_, executeErr := w.Execute(job)

	return executeErr
}

func (w *GenericWorker) Run(cmd *cobra.Command, iter iterator.Iterator, inputIterators *flags.RequestInputIterators) error {
	// TODO: How does an unbound iterator get caught here?
	if inputIterators == nil {
		return fmt.Errorf("missing input iterators")
	}

	// get common options and batch settings
	commonOptions, err := w.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	batchOptions, err := w.GetBatchOptions(cmd)
	if err != nil {
		return err
	}

	// Configure post actions
	batchOptions.PostActions = inputIterators.PipeOptions.PostActions

	return w.run(iter, commonOptions, *batchOptions)
}

func (w *GenericWorker) run(iter iterator.Iterator, commonOptions config.CommonCommandOptions, batchOptions BatchOptions) error {
	// Two channels - to send them work and to collect their results.
	// buffer size does not really matter, it just needs to be high
	// enough not to block the workers

	// TODO: how to detect when request iterator is finished when using the body iterator (total number of requests?)
	if batchOptions.TotalWorkers < 1 {
		batchOptions.TotalWorkers = 1
	}
	jobs := make(chan Job, batchOptions.TotalWorkers-1)
	results := make(chan error, batchOptions.TotalWorkers-1)
	workers := sync.WaitGroup{}

	// don't start the progress bar until all confirmations are done
	progbar := progressbar.NewMultiProgressBar(w.IO.ErrOut, 1, batchOptions.TotalWorkers, "requests", w.Config.ShowProgress())

	for iWork := 1; iWork <= batchOptions.TotalWorkers; iWork++ {
		w.Logger.Debugf("starting worker: %d", iWork)
		workers.Add(1)
		go w.StartWorker(iWork, jobs, results, progbar, &workers)
	}

	jobID := int64(0)
	skipConfirm := false
	shouldConfirm := false
	promptCount := int32(0)
	promptWG := sync.WaitGroup{}

	maxJobs := w.GetMaxJobs()
	tenantName := ""
	if w.Client != nil {
		tenantName = w.Client.TenantName
	}
	w.Logger.Infof("Max jobs: %d", maxJobs)

	// add jobs async
	go func() {
		defer close(jobs)
		jobInputErrors := int64(0)
		for {
			jobID++
			w.Logger.Debugf("checking job iterator: %d", jobID)

			// check if iterator is exhausted
			value, input, err := iter.GetNext()

			if errors.Is(err, io.EOF) {
				// no more requests, decrement job id as the job was not started
				jobID--
				break
			}

			if maxJobs != 0 && jobID > maxJobs {
				w.Logger.Infof("maximum jobs reached: limit=%d", maxJobs)
				break
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					// no more requests
					break
				}
				jobInputErrors++

				rootCauseErr := err
				if errors.Is(err, cmderrors.ErrNoMatchesFound) {
					rootCauseErr = err
				} else if parentErr := errors.Unwrap(err); parentErr != nil {
					rootCauseErr = parentErr
				}

				w.Config.LogErrorF(rootCauseErr, "skipping job: %d. %s", jobID, rootCauseErr)
				results <- err

				// Note: stop adding jobs if total errors are exceeded
				// This is necessary as the worker still needs time to process
				// the current job, so there can be a delay before the results are read.
				if jobInputErrors >= int64(batchOptions.AbortOnErrorCount) {
					break
				}

				// move to next job
				continue
			}
			w.Logger.Debugf("adding job: %d", jobID)

			if value != nil {
				if batchOptions.SemanticMethod != "" {
					// Use a custom method which controls how the request should be handled but is not the actual request
					shouldConfirm = w.Config.ShouldConfirm(batchOptions.SemanticMethod)
				} else {
					// TODO: Allow a job to control if something should be confirmed or not
					// shouldConfirm = w.Config.ShouldConfirm(request.Method)
				}
			}

			// confirm action
			if !skipConfirm && shouldConfirm {
				// wait for any other previous prompted jobs to finish
				promptWG.Wait()

				operation := "Execute command"
				if commonOptions.ConfirmText != "" {
					operation = commonOptions.ConfirmText
				} else if len(os.Args[1:]) > 1 {
					// build confirm text from cmd structure
					operation = fmt.Sprintf("%s %s", os.Args[2], strings.TrimRight(os.Args[1], "s"))
				}
				promptMessage, _ := batchOptions.GetConfirmationMessage(operation, value, input)
				confirmResult, err := prompt.Confirm(fmt.Sprintf("(job: %d)", jobID), promptMessage, "tenant "+tenantName, prompt.ConfirmYes.String(), false)

				switch confirmResult {
				case prompt.ConfirmYesToAll:
					skipConfirm = true
				case prompt.ConfirmYes:
					// confirmed
				case prompt.ConfirmNo:
					w.Logger.Warningf("skipping job: %d. %s", jobID, err)
					if w.ActivityLogger != nil {
						// TODO: Let batching control custom log message
						// w.ActivityLogger.LogCustom(err.Error() + ". " + request.Path)
					}
					results <- err
					continue
				case prompt.ConfirmNoToAll:
					w.Logger.Infof("skipping job: %d. %s", jobID, err)
					if w.ActivityLogger != nil {
						// TODO: Let batching control custom log message
						// w.ActivityLogger.LogCustom(err.Error() + ". " + request.Path)
					}
					w.Logger.Infof("cancelling all remaining jobs")
					results <- err
				}
				if confirmResult == prompt.ConfirmNoToAll {
					break
				}

				promptWG.Add(1)
				atomic.AddInt32(&promptCount, 1)
			}

			if skipConfirm || !shouldConfirm {
				progbar.Start(float64(batchOptions.Delay * 2 / time.Millisecond))
			}

			jobs <- Job{
				ID:            jobID,
				Options:       batchOptions,
				Input:         input,
				Value:         value,
				CommonOptions: commonOptions,
			}
		}

		w.Logger.Debugf("finished adding jobs. lastJobID=%d", jobID)
	}()

	// collect all the results of the work.
	totalErrors := make([]error, 0)

	// close the results when the works are finished, but don't block reading the results
	wasCancelled := int32(0)
	go func() {
		workers.Wait()
		time.Sleep(200 * time.Microsecond)

		// prevent closing channel twice
		if atomic.AddInt32(&wasCancelled, 1) == 1 {
			close(results)
		}
	}()

	for err := range results {
		if err == nil {
			w.Logger.Debugf("job successful")
		} else {
			w.Logger.Infof("job error. %s", err)
		}

		if err != nil && err != io.EOF {

			// overwrite error
			err = w.CheckError(err)
			if err != nil {
				totalErrors = append(totalErrors, err)
			}
		}
		// exit early
		if batchOptions.AbortOnErrorCount != 0 && len(totalErrors) >= batchOptions.AbortOnErrorCount {
			if atomic.AddInt32(&wasCancelled, 1) == 1 {
				close(results)
			}
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, fmt.Sprintf("aborted batch as error count has been exceeded. totalErrors=%d", batchOptions.AbortOnErrorCount))
		}

		// communicate that the prompt has received a result
		pendingPrompts := atomic.AddInt32(&promptCount, -1)
		if pendingPrompts+1 > 0 {
			promptWG.Done()
		}
	}
	if progbar.IsEnabled() && progbar.IsRunning() && jobID > 1 {
		// wait for progress bar to update last increment
		time.Sleep(progbar.RefreshRate())
	}

	maxJobsReached := maxJobs != 0 && jobID > maxJobs
	if total := len(totalErrors); total > 0 {
		if total == 1 {
			// return only error
			return totalErrors[0]
		}
		// aggregate error
		message := fmt.Sprintf("jobs completed with %d errors", total)
		if maxJobsReached {
			message += fmt.Sprintf(". job limit exceeded=%v", maxJobsReached)
		}
		return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCompletedWithErrors, message)
	}
	if maxJobsReached {
		return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitJobLimitExceeded, fmt.Sprintf("max job limit exceeded. limit=%d", maxJobs))
	}
	return nil
}

// These workers will receive work on the `jobs` channel and send the corresponding
// results on `results`
func (w *GenericWorker) StartWorker(id int, jobs <-chan Job, results chan<- error, prog *progressbar.ProgressBar, wg *sync.WaitGroup) {
	var err error
	onStartup := true

	var total int64

	defer wg.Done()
	for job := range jobs {
		total++
		workerStart := prog.StartJob(id, total)

		if job.Options.DelayBefore > 0 {
			w.Logger.Infof("worker %d: sleeping %s before starting job", id, job.Options.DelayBefore)
			time.Sleep(job.Options.DelayBefore)
		}

		if !onStartup {
			if !errors.Is(err, io.EOF) && job.Options.Delay > 0 {
				w.Logger.Infof("worker %d: sleeping %s before fetching next job", id, job.Options.Delay)
				time.Sleep(job.Options.Delay)
			}
		}
		onStartup = false

		w.Logger.Infof("worker %d: started job %d", id, job.ID)
		startTime := time.Now().UnixNano()

		result, resultErr := w.Execute(job)
		// Handle post request actions (only if original response was ok)
		// and stop actions if an error is encountered
		if resultErr == nil {
			for i, action := range job.Options.PostActions {
				w.Logger.Debugf("Executing action: %d", i)
				runOutput, runErr := action.Run(result)
				if runErr != nil {
					resultErr = runErr
					w.Logger.Warningf("Action failed. output=%#v, err=%s", runOutput, runErr)
					break
				}
			}
		}

		elapsedMS := (time.Now().UnixNano() - startTime) / 1000.0 / 1000.0

		w.Logger.Infof("worker %d: finished job %d in %dms", id, job.ID, elapsedMS)
		prog.FinishedJob(id, workerStart)

		// return result before delay, so errors can be handled before the sleep
		results <- resultErr
	}
	prog.WorkerCompleted(id)
}
