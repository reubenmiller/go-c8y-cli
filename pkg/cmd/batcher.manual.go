package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type RequestTemplater interface {
	GetBody(i string) (map[string]interface{}, error)
	GetPath(i string) (string, error)
}

type TemplateGenerator struct {
	cmd  *cobra.Command
	Path string
	Body map[string]interface{}

	GetCustomBody func(string) (map[string]interface{}, error)
}

func newTemplateGenerator(cmd *cobra.Command, path string, body map[string]interface{}) *TemplateGenerator {
	return &TemplateGenerator{
		cmd:  cmd,
		Path: path,
		Body: body,
	}
}

func (n *TemplateGenerator) GetBody(i string) (interface{}, error) {
	body := mapbuilder.NewMapBuilder()

	if n.Body != nil {
		body.SetMap(n.Body)
	}

	if err := setDataTemplateFromFlags(n.cmd, body); err != nil {
		return nil, newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return nil, newUserError("Body validation error. ", err)
	}
	return body, nil
}

func (n *TemplateGenerator) GetPath(i string) (string, error) {
	pathParameters := make(map[string]string)
	pathParameters["id"] = i
	return replacePathParameters(n.Path, pathParameters), nil
}

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

func runBatched(requestIterator *RequestIterator, commonOptions CommonCommandOptions, batchOptions BatchOptions) error {
	// Two channels - to send them work and to collect their results.
	// buffer size does not really matter, it just needs to be high
	// enough not to block the workers

	// TODO: how to detect when request iterator is finished when using the body iterator (total number of requests?)
	jobs := make(chan batchArgument, 3)
	results := make(chan error, 100)
	workers := sync.WaitGroup{}

	if batchOptions.TotalWorkers < 1 {
		batchOptions.TotalWorkers = 1
	}

	for w := 1; w <= batchOptions.TotalWorkers; w++ {
		Logger.Infof("Starting worker: %d", w)
		workers.Add(1)
		go batchWorker(w, jobs, results, &workers)
	}

	jobID := int64(0)

	// add jobs async
	go func() {
		defer close(jobs)
		for {
			jobID++
			Logger.Infof("Checking job iterator: %d", jobID)

			request, err := requestIterator.GetNext()

			if err != nil {
				if !errors.Is(err, io.EOF) {
					Logger.Warningf("request iterator: unexpected error. err=%s", err)
				}
				break
			}
			Logger.Infof("Adding job: %d", jobID)

			jobs <- batchArgument{
				id:            jobID,
				batchOptions:  batchOptions,
				request:       *request,
				commonOptions: commonOptions,
			}
		}

		Logger.Info("Finished adding jobs")
	}()

	// collect all the results of the work.
	totalErrors := make([]error, 0)

	// close the results when the works are finished, but don't block reading the results
	go func() {
		workers.Wait()
		close(results)
	}()

	for err := range results {
		Logger.Infof("Reading job result: %s", err)
		if err != nil && err != io.EOF {
			totalErrors = append(totalErrors, err)
		}
		// exit early
		if batchOptions.AbortOnErrorCount != 0 && len(totalErrors) >= batchOptions.AbortOnErrorCount {
			close(results)
			return newUserErrorWithExitCode(103, fmt.Sprintf("aborted batch as error count has been exceeded. totalErrors=%d", batchOptions.AbortOnErrorCount))
		}
	}

	if total := len(totalErrors); total > 0 {
		if total > 1 {
			return newUserErrorWithExitCode(104, fmt.Sprintf("batch completed with %d errors", total))
		}
		return newUserErrorWithExitCode(104, fmt.Sprintf("batch completed with %d error", total))
	}
	return nil
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

// These workers will receive work on the `jobs` channel and send the corresponding
// results on `results`
func batchWorker(id int, jobs <-chan batchArgument, results chan<- error, wg *sync.WaitGroup) {

	defer wg.Done()
	for job := range jobs {
		Logger.Infof("worker %d: started job %d", id, job.id)
		startTime := time.Now().UnixNano()

		err := processRequestAndResponse([]c8y.RequestOptions{job.request}, job.commonOptions)
		elapsedMS := (time.Now().UnixNano() - startTime) / 1000.0 / 1000.0

		Logger.Infof("worker %d: finished job %d in %dms", id, job.id, elapsedMS)

		// return result before delay, so errors can be handled before the sleep
		results <- err

		// Skip delay if end of work
		if !errors.Is(err, io.EOF) && job.batchOptions.Delay > 0 {
			Logger.Infof("worker %d: sleeping %dms before fetching next job", id, job.batchOptions.Delay)
			time.Sleep(time.Duration(job.batchOptions.Delay) * time.Millisecond)
		}
	}
}

func addBatchFlags(cmd *cobra.Command, acceptInputFile bool) {
	if acceptInputFile {
		cmd.Flags().String("inputFile", "", "Input file of ids to add to processed (required)")
		// cmd.MarkFlagRequired("inputFile")
	}
	cmd.Flags().Int("count", 5, "Total number of objects")
	cmd.Flags().Int("startIndex", 1, "Start index value")
	// cmd.Flags().Int("abortOnErrors", 10, "Abort batch when reaching specified number of errors")
	// cmd.Flags().Int("delay", 200, "delay in milliseconds after each request")
	// cmd.Flags().Int("workers", 2, "Number of workers")
}

func runTemplateOnList(cmd *cobra.Command, requestIterator *RequestIterator) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	batchOptions, err := getBatchOptions(cmd)
	if err != nil {
		return err
	}

	return runBatched(requestIterator, commonOptions, *batchOptions)
}

func readFile(filepath string) ([]string, error) {
	csvfile, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not open csv file. %w", err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)
	items := make([]string, 0)

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return items, err
		}
		value := strings.TrimSpace(record[0])

		if _, err := strconv.Atoi(value); err != nil {
			continue
		}
		items = append(items, value)
	}
	return items, nil
}

func processRequestAndResponseWithWorkers(cmd *cobra.Command, r *c8y.RequestOptions, name string) error {

	var pathIter iterator.Iterator

	if name != "" {
		// create input iterator
		items, err := NewFlagPipeEnabledStringSlice(cmd, name)
		if err != nil {
			return err
		}
		pathIter = iterator.NewCompositeStringIterator(items, strings.ReplaceAll(r.Path, "{"+name+"}", "%s"))
	} else {
		// fixed path (no iteration)
		pathIter = iterator.NewRepeatIterator(r.Path, 1)
	}

	body := make(map[string]interface{})
	requestIter := NewRequestIterator(*r, pathIter, body)

	// for {
	// 	next, err := requestIter.GetNext()

	// 	if err == io.EOF {
	// 		break
	// 	}

	// 	Logger.Infof("request: %v, %s %s/%s?%s, %v", next.Header, next.Method, next.Host, next.Path, next.Query, next.Body)
	// 	break
	// }
	return runTemplateOnList(cmd, requestIter)
}
