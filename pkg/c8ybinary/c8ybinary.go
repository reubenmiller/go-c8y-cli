// Package c8ybinary shows upload and download progress bars for binaries, and
// uploads a binary with the go-c8y v2 client.
package c8ybinary

import (
	"context"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

const BarFiller = "[━━ ]"

// const BarFiller = "[██-]"

// UploadBinary uploads a file as a global inventory binary with the go-c8y v2
// client, and shows a progress bar (when the progress indicator is enabled).
// The file is not uploaded when ctx is a dry run (the SDK renders the request).
func UploadBinary(ctx context.Context, client *apiv2.Client, progress *iostreams.Progress, filename string) (jsonmodels.Binary, error) {
	if err := fileguard.Check(filename); err != nil {
		return jsonmodels.Binary{}, err
	}
	opt := core.UploadFileOptions{
		FilePath:   filename,
		Properties: map[string]any{"c8y_Global": map[string]any{}},
	}

	finish := func() {}
	if progress.IsEnabled() {
		fn, done, err := NewUploadProgress(progress, filename)
		if err != nil {
			return jsonmodels.Binary{}, err
		}
		opt.Progress = fn
		finish = done
	}

	result := client.Binaries.Create(ctx, opt)
	finish()
	progress.Wait()
	if result.Err != nil {
		return jsonmodels.Binary{}, result.Err
	}
	return result.Data, nil
}

// NewUploadProgress returns a go-c8y progress function which advances a
// progress bar sized to the file, and a function which completes the bar (call
// it when the upload ends, whether it succeeded or not, so waiting for the
// progress indicator never blocks).
func NewUploadProgress(progress progressBarAdder, filename string) (core.ProgressFunc, func(), error) {
	bar, err := newFileProgressBar(progress, filename)
	if err != nil {
		return nil, nil, err
	}
	if bar == nil {
		return func(int64, int64) {}, func() {}, nil
	}
	return func(read, _ int64) { bar.SetCurrent(read) }, func() { bar.Abort(false) }, nil
}

// AddProgress returns a request hook (for apiv2.RequestOptions.PrepareRequest)
// which shows the progress of the request body (the upload of the file given
// by the fileFlag flag) in the progress indicator. It does nothing when no file
// is given.
func AddProgress(cmd *cobra.Command, fileFlag string, progress *mpb.Progress) func(r *http.Request) error {
	return func(r *http.Request) error {
		if r.Body == nil || r.Body == http.NoBody || progress == nil {
			return nil
		}
		filename, err := cmd.Flags().GetString(fileFlag)
		if err != nil {
			// Don't return error here, as if the user does not provide a file, then ignore the progress
			return nil
		}

		if filename == "" {
			return nil
		}

		fn, _, err := NewUploadProgress(progress, filename)
		if err != nil {
			return err
		}
		r.Body = core.NewProgressReader(r.Body, r.ContentLength, fn)
		return nil
	}
}

// DownloadProgress returns a response body wrapper (see
// api.WithResponseBodyWrapper) which shows a progress bar for the download of a
// binary (a response with a Content-Disposition header and a known length).
// Any other response body is left as is.
func DownloadProgress(progress *iostreams.Progress) func(resp *http.Response) io.ReadCloser {
	return func(r *http.Response) io.ReadCloser {
		if !progress.IsEnabled() || r.Header.Get("Content-Disposition") == "" || r.ContentLength <= 0 {
			return nil
		}
		return newDownloadReader(progress, r)
	}
}

// CreateProxyReader returns a function which wraps the body of a response so
// that the download is shown in a progress bar
func CreateProxyReader(progress *iostreams.Progress) func(response *http.Response) io.Reader {
	return func(r *http.Response) io.Reader {
		reader := newDownloadReader(progress, r)
		if reader == nil {
			return r.Body
		}
		r.Body = reader
		return reader
	}
}

// newDownloadReader returns the body of a response which advances a progress bar
// named after the downloaded file, or nil when progress bars are disabled
func newDownloadReader(progress *iostreams.Progress, r *http.Response) io.ReadCloser {
	basename := "download"

	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Disposition"))
	if err == nil {
		if filename, ok := params["filename"]; ok {
			basename = filename
		}
	}

	bar := progress.Add(r.ContentLength, mpb.NewBarFiller(BarFiller), barDecorators(basename)...)
	if bar == nil {
		return nil
	}
	return bar.ProxyReader(r.Body)
}

// progressBarAdder adds a new bar to a progress indicator
type progressBarAdder interface {
	Add(total int64, filler mpb.BarFiller, options ...mpb.BarOption) *mpb.Bar
}

func NewProgressBar(progress *iostreams.Progress, filename string) (*mpb.Bar, error) {
	if !progress.IsEnabled() {
		return nil, nil
	}
	return newFileProgressBar(progress, filename)
}

func newFileProgressBar(progress progressBarAdder, filename string) (*mpb.Bar, error) {
	file, err := os.Stat(filename)

	if err != nil {
		return nil, err
	}

	bar := progress.Add(file.Size(), mpb.NewBarFiller(BarFiller), barDecorators(filepath.Base(filename))...)
	return bar, nil
}

func barDecorators(basename string) []mpb.BarOption {
	return []mpb.BarOption{
		mpb.PrependDecorators(
			decor.Name("elapsed", decor.WC{W: len("elapsed") + 1, C: decor.DidentRight}),
			decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
			decor.Name(basename, decor.WC{W: len(basename) + 1, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WC{W: 6, C: decor.DidentRight}),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
	}
}
