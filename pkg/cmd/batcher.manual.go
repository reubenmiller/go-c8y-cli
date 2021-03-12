package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/progressbar"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type BatchOptions struct {
	StartIndex        int
	NumJobs           int
	TotalWorkers      int
	Delay             int
	AbortOnErrorCount int

	InputData []string

	inputIndex int
}

func (b *BatchOptions) GetItem() (string, error) {
	defer func() {
		b.inputIndex++
	}()

	if b.useInputData() {
		if b.inputIndex < len(b.InputData) {

			return b.InputData[b.inputIndex], nil
		}
		return "", fmt.Errorf("end of input data")
	}

	if b.inputIndex >= b.NumJobs {
		return "", fmt.Errorf("end of input data")
	}
	return fmt.Sprintf("%d", b.inputIndex+b.StartIndex), nil

}

func (b *BatchOptions) useInputData() bool {
	return b.InputData != nil && len(b.InputData) > 0
}

func getBatchOptions(cmd *cobra.Command) (*BatchOptions, error) {
	options := &BatchOptions{
		AbortOnErrorCount: 10,
		TotalWorkers:      1,
		Delay:             1000,
	}

	if v, err := cmd.Flags().GetInt("count"); err == nil {
		options.NumJobs = v
	}

	if v, err := cmd.Flags().GetInt("startIndex"); err == nil {
		options.StartIndex = v
	}

	if v, err := cmd.Flags().GetInt("delay"); err == nil {
		options.Delay = v
	}

	if v, err := cmd.Flags().GetInt("workers"); err == nil {
		if v > globalFlagBatchMaxWorkers {
			return nil, fmt.Errorf("number of workers exceeds the maximum workers limit of %d", globalFlagBatchMaxWorkers)
		}
		options.TotalWorkers = v
	}

	if v, err := cmd.Flags().GetInt("abortOnErrors"); err == nil {
		options.AbortOnErrorCount = v
	}

	return options, nil
}

type batchArgument struct {
	id            int64
	request       c8y.RequestOptions
	commonOptions CommonCommandOptions
	batchOptions  BatchOptions
}

func processRequestAndResponseWithWorkers(cmd *cobra.Command, r *c8y.RequestOptions, inputIterators *flags.RequestInputIterators) error {
	var err error
	var pathIter iterator.Iterator

	if inputIterators != nil && inputIterators.Total > 0 {
		if inputIterators.Path != nil {
			pathIter = inputIterators.Path
		} else {
			// use continuous path repeater so that it does not stop the other interators
			pathIter = iterator.NewRepeatIterator(r.Path, 0)
		}
		if inputIterators.Body != nil {
			r.Body = inputIterators.Body
		}
	}
	if pathIter == nil {
		pathIter = iterator.NewRepeatIterator(r.Path, 1)
	}
	// Note: Body accepts iterator types, so no need for special handling here
	requestIter := NewRequestIterator(*r, pathIter, inputIterators.Query, r.Body)

	// get common options and batch settings
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	batchOptions, err := getBatchOptions(cmd)
	if err != nil {
		return err
	}

	return runBatched(requestIter, commonOptions, *batchOptions)
}

func runBatched(requestIterator *RequestIterator, commonOptions CommonCommandOptions, batchOptions BatchOptions) error {
	// Two channels - to send them work and to collect their results.
	// buffer size does not really matter, it just needs to be high
	// enough not to block the workers

	// TODO: how to detect when request iterator is finished when using the body iterator (total number of requests?)
	if batchOptions.TotalWorkers < 1 {
		batchOptions.TotalWorkers = 1
	}
	jobs := make(chan batchArgument, batchOptions.TotalWorkers-1)
	results := make(chan error, batchOptions.TotalWorkers-1)
	workers := sync.WaitGroup{}

	progbar := progressbar.NewMultiProgressBar(rootCmd.ErrOrStderr(), 1, batchOptions.TotalWorkers, "requests", globalFlagProgressBar)
	progbar.Start(float64(batchOptions.Delay * 2 / 1000))

	for w := 1; w <= batchOptions.TotalWorkers; w++ {
		Logger.Debugf("starting worker: %d", w)
		workers.Add(1)
		go batchWorker(w, jobs, results, progbar, &workers)
	}

	jobID := int64(0)
	skipConfirm := false
	promptCount := int32(0)
	promptWG := sync.WaitGroup{}

	// add jobs async
	go func() {
		defer close(jobs)
		jobInputErrors := int64(0)
		for {
			jobID++
			Logger.Debugf("checking job iterator: %d", jobID)

			if jobID > globalFlagBatchMaxJobs {
				Logger.Warningf("maximum jobs reached: limit=%d", globalFlagBatchMaxJobs)
				break
			}

			request, input, err := requestIterator.GetNext()

			if err != nil {
				if errors.Is(err, io.EOF) {
					// no more requests
					break
				}
				jobInputErrors++

				rootCauseErr := err
				if errors.Is(err, ErrNoMatchesFound) {
					rootCauseErr = err
				} else if parentErr := errors.Unwrap(err); parentErr != nil {
					rootCauseErr = parentErr
				}

				LogErrorF(rootCauseErr, "skipping job: %d. %s", jobID, rootCauseErr)
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
			Logger.Debugf("adding job: %d", jobID)

			// confirm action
			if request != nil && !skipConfirm && shouldConfirm(request.Method) {
				// wait for any other previous prompted jobs to finish
				promptWG.Wait()

				operation := "Execute command"
				if commonOptions.ConfirmText != "" {
					operation = commonOptions.ConfirmText
				} else if len(os.Args[1:]) > 1 {
					// build confirm text from cmd structure
					operation = fmt.Sprintf("%s %s", os.Args[2], strings.TrimRight(os.Args[1], "s"))
				}
				promptMessage, _ := getConfirmationMessage(operation, request, input)
				confirmResult, err := prompt.Confirm(fmt.Sprintf("(job: %d)", jobID), promptMessage, "tenant "+client.TenantName, prompt.ConfirmYes.String(), false)

				switch confirmResult {
				case prompt.ConfirmYesToAll:
					skipConfirm = true
				case prompt.ConfirmYes:
					// confirmed
				case prompt.ConfirmNo:
					Logger.Warningf("skipping job: %d. %s", jobID, err)
					activityLogger.LogCustom(err.Error() + ". " + request.Path)
					results <- err
					continue
				case prompt.ConfirmNoToAll:
					Logger.Warningf("skipping job: %d. %s", jobID, err)
					activityLogger.LogCustom(err.Error() + ". " + request.Path)
					Logger.Warning("cancelling all remaining jobs")
					results <- err
				}
				if confirmResult == prompt.ConfirmNoToAll {
					break
				}

				promptWG.Add(1)
				atomic.AddInt32(&promptCount, 1)
			}

			jobs <- batchArgument{
				id:            jobID,
				batchOptions:  batchOptions,
				request:       *request,
				commonOptions: commonOptions,
			}
		}

		Logger.Debug("finished adding jobs")
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
			Logger.Debugf("job successful")
		} else {
			Logger.Infof("job error. %s", err)
		}

		if err != nil && err != io.EOF {
			totalErrors = append(totalErrors, err)
		}
		// exit early
		if batchOptions.AbortOnErrorCount != 0 && len(totalErrors) >= batchOptions.AbortOnErrorCount {
			if atomic.AddInt32(&wasCancelled, 1) == 1 {
				close(results)
			}
			return cmderrors.NewUserErrorWithExitCode(103, fmt.Sprintf("aborted batch as error count has been exceeded. totalErrors=%d", batchOptions.AbortOnErrorCount))
		}

		// communicate that the prompt has received a result
		pendingPrompts := atomic.AddInt32(&promptCount, -1)
		if pendingPrompts+1 > 0 {
			promptWG.Done()
		}
	}
	if progbar.IsEnabled() && jobID > 1 {
		// wait for progress bar to update last increment
		time.Sleep(progbar.RefreshRate())
	}

	maxJobsReached := jobID > globalFlagBatchMaxJobs
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
		return cmderrors.NewUserErrorWithExitCode(104, message)
	}
	if maxJobsReached {
		return cmderrors.NewUserErrorWithExitCode(105, fmt.Sprintf("max job limit exceeded. limit=%d", globalFlagBatchMaxJobs))
	}
	return nil
}

// These workers will receive work on the `jobs` channel and send the corresponding
// results on `results`
func batchWorker(id int, jobs <-chan batchArgument, results chan<- error, prog *progressbar.ProgressBar, wg *sync.WaitGroup) {
	var err error
	onStartup := true

	var total int64

	defer wg.Done()
	for job := range jobs {
		total++
		workerStart := prog.StartJob(id, total)
		if !onStartup {
			if !errors.Is(err, io.EOF) && job.batchOptions.Delay > 0 {
				Logger.Infof("worker %d: sleeping %dms before fetching next job", id, job.batchOptions.Delay)
				time.Sleep(time.Duration(job.batchOptions.Delay) * time.Millisecond)
			}
		} else {
			jitter := rand.Int31n(50)
			time.Sleep(time.Duration(jitter) * time.Millisecond)
		}
		onStartup = false

		Logger.Infof("worker %d: started job %d", id, job.id)
		startTime := time.Now().UnixNano()

		err = processRequestAndResponse([]c8y.RequestOptions{job.request}, job.commonOptions)
		elapsedMS := (time.Now().UnixNano() - startTime) / 1000.0 / 1000.0

		Logger.Infof("worker %d: finished job %d in %dms", id, job.id, elapsedMS)
		prog.FinishedJob(id, workerStart)

		// return result before delay, so errors can be handled before the sleep
		results <- err
	}
	prog.WorkerCompleted(id)
}

func getConfirmationMessage(prefix string, request *c8y.RequestOptions, input interface{}) (string, error) {
	name := ""
	id := ""
	if input != nil {
		Logger.Infof("input: %s", input)
		Logger.Infof("input type: %s", reflect.TypeOf(input))

		switch v := input.(type) {
		case []byte:
			id = fmt.Sprintf("%s", v)
			if gjson.ValidBytes(v) {
				jsonobj := gjson.ParseBytes(v)

				name = jsonobj.Get("name").Str
				if idField := jsonobj.Get("id"); idField.Exists() {
					id = jsonobj.Get("id").Str
				}
			}

		case string:
			if !strings.HasPrefix(v, "{") && !strings.HasPrefix(v, "[") {
				id = v
			} else {
				if gjson.Valid(v) {
					jsonobj := gjson.Parse(v)

					name = jsonobj.Get("name").Str
					if idField := jsonobj.Get("id"); idField.Exists() {
						id = jsonobj.Get("id").Str
					}
				}
			}
		}
	}

	if request != nil {
		switch v := request.Body.(type) {
		case map[string]interface{}:
			jsonText, err := json.Marshal(v)

			if err == nil {
				devicePaths := []string{"source.id", "deviceId", "id"}
				for _, path := range devicePaths {
					if device := gjson.ParseBytes(jsonText).Get(path); device.Exists() {
						Logger.Infof("device: %s", device.Str)
						id = device.Str
						break
					}
				}
			} else {
				Logger.Debugf("json error: %s", err)
			}
		}
	}

	target := ""
	if id != "" && isID(id) {
		target += "id=" + id
	}

	if name != "" {
		target += ", name=" + name
	}

	Logger.Infof("target: [%s]", target)

	if target != "" {
		return fmt.Sprintf("%s [%s]", prefix, target), nil
	}

	return prefix, nil
}

func isID(v string) bool {
	isNotDigit := func(c rune) bool { return c < '0' || c > '9' }
	value := strings.TrimSpace(v)
	return strings.IndexFunc(value, isNotDigit) <= -1
}
