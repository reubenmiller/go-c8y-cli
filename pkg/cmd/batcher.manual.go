package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	Body string

	GetCustomBody func(string) (map[string]interface{}, error)
}

func newTemplateGenerator(cmd *cobra.Command, path string, body string) *TemplateGenerator {
	return &TemplateGenerator{
		cmd:  cmd,
		Path: path,
		Body: body,
	}
}

func (n *TemplateGenerator) GetBody(i string) (interface{}, error) {
	// if n.Body != "" {
	// 	body, err := n.GetStringBody(i)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	bodyMap := make(map[string]interface{})
	// 	if err := json.Unmarshal([]byte(body), &bodyMap); err != nil {
	// 		return "", err
	// 	}
	// 	return bodyMap, nil
	// }
	return n.GetTemplateBody(i)
}

func (n *TemplateGenerator) GetStringBody(i string) (string, error) {

	pathParameters := make(map[string]string)
	pathParameters["id"] = i
	return replacePathParameters(n.Body, pathParameters), nil
}

func (n *TemplateGenerator) GetTemplateBody(i string) (map[string]interface{}, error) {
	// body
	body := mapbuilder.NewMapBuilder()

	body.TemplateIndex = i
	body.SetMap(getDataFlag(n.cmd))

	// TODO: merge a fixed body
	if n.Body != "" {
		fixedBody, err := n.GetStringBody(i)
		if err != nil {
			return nil, err
		}
		if err := body.MergeJsonnet(fixedBody, false); err != nil {
			return nil, err
		}
	}

	if err := setDataTemplateFromFlags(n.cmd, body); err != nil {
		return nil, newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return nil, newUserError("Body validation error. ", err)
	}
	return body.GetMap(), nil
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

func (b *BatchOptions) GetJobsTotal() int {
	if b.InputData != nil && len(b.InputData) > 0 {
		return len(b.InputData)
	}
	return b.NumJobs
}

func runBatched(templater *TemplateGenerator, req c8y.RequestOptions, commonOptions CommonCommandOptions, batchOptions BatchOptions) error {
	// Two channels - to send them work and to collect their results.
	totalJobs := batchOptions.GetJobsTotal()
	totalWorkers := batchOptions.TotalWorkers

	jobs := make(chan batchArgument, totalJobs)
	results := make(chan error, totalJobs)

	// This starts up 3 workers, initially blocked because there are no jobs yet.
	for w := 1; w <= totalWorkers; w++ {
		go batchWorker(w, templater, jobs, results)
	}

	for {
		value, err := batchOptions.GetItem()
		if err != nil {
			break
		}
		jobs <- batchArgument{
			id:            value,
			batchOptions:  batchOptions,
			request:       req,
			commonOptions: commonOptions,
		}
	}
	close(jobs)

	// collect all the results of the work.
	totalErrors := make([]error, 0)

	for a := 1; a <= totalJobs; a++ {
		if err := <-results; err != nil {
			totalErrors = append(totalErrors, err)
		}
		if batchOptions.AbortOnErrorCount != 0 && len(totalErrors) >= batchOptions.AbortOnErrorCount {
			return newUserErrorWithExitCode(103, fmt.Sprintf("aborted batch as error count has been exceeded. totalErrors=%d", batchOptions.AbortOnErrorCount))
		}
	}
	if total := len(totalErrors); total > 0 {
		return newUserErrorWithExitCode(104, fmt.Sprintf("batch completed with %d errors", total))
	}
	return nil
}

func getBatchOptions(cmd *cobra.Command) (*BatchOptions, error) {
	options := &BatchOptions{
		AbortOnErrorCount: 10,
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
	id            string
	request       c8y.RequestOptions
	commonOptions CommonCommandOptions
	batchOptions  BatchOptions
}

// These workers will receive work on the `jobs` channel and send the corresponding
// results on `results`
func batchWorker(id int, template *TemplateGenerator, jobs <-chan batchArgument, results chan<- error) {
	for job := range jobs {
		Logger.Infof("worker %d: started job %s", id, job.id)
		startTime := time.Now().UnixNano()

		currentRequest := job.request

		if body, err := template.GetBody(job.id); err == nil {
			currentRequest.Body = body
		} else {
			Logger.Errorf("error parsing template. %s", err)
			results <- err
			continue
		}

		if path, err := template.GetPath(job.id); err == nil {
			currentRequest.Path = path
		}

		err := processRequestAndResponse([]c8y.RequestOptions{currentRequest}, job.commonOptions)

		elapsedMS := (time.Now().UnixNano() - startTime) / 1000.0 / 1000.0

		Logger.Infof("worker %d: finished job %s in %dms", id, job.id, elapsedMS)
		if job.batchOptions.Delay > 0 {
			Logger.Infof("worker %d: sleeping %dms before fetching next job", id, job.batchOptions.Delay)
			time.Sleep(time.Duration(job.batchOptions.Delay) * time.Millisecond)
		}
		results <- err
	}
}

func addBatchFlags(cmd *cobra.Command, acceptInputFile bool) {
	if acceptInputFile {
		cmd.Flags().String("inputFile", "", "Input file of ids to add to processed (required)")
		cmd.MarkFlagRequired("inputFile")
	}
	cmd.Flags().Int("abortOnErrors", 10, "Abort batch when reaching specified number of errors")
	cmd.Flags().Int("count", 5, "Total number of objects")
	cmd.Flags().Int("startIndex", 1, "Start index value")
	cmd.Flags().Int("delay", 200, "delay in milliseconds after each request")
	cmd.Flags().Int("workers", 2, "Number of workers")
}

func runTemplateOnList(cmd *cobra.Command, method, path string, body string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// headers
	headers := http.Header{}
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	req := c8y.RequestOptions{
		Method:       method,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	templator := newTemplateGenerator(cmd, path, body)
	batchOptions, err := getBatchOptions(cmd)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("inputFile") {

		if v, err := cmd.Flags().GetString("inputFile"); err == nil && v != "" {
			items, err := readFile(v)
			if err != nil {
				return err
			}
			batchOptions.InputData = items
		}
	}

	return runBatched(templator, req, commonOptions, *batchOptions)
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
