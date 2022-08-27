package c8ybinary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

const BarFiller = "━━━  "

func CreateBinaryWithProgress(ctx context.Context, client *c8y.Client, path string, filename string, properties interface{}, progress *mpb.Progress) (*c8y.Response, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	values := map[string]io.Reader{
		"file": file,
	}

	if properties != nil {
		metadataBytes, err := json.Marshal(properties)
		if err != nil {
			return nil, fmt.Errorf("failed to convert binary properties to json. %w", err)
		}
		values["object"] = bytes.NewReader(metadataBytes)
	}

	return client.SendRequest(ctx, c8y.RequestOptions{
		Method:   "POST",
		Accept:   "application/json",
		Path:     path,
		FormData: values,
		PrepareRequest: func(r *http.Request) (*http.Request, error) {
			if r.Body == nil || progress == nil {
				return r, nil
			}

			var size int64
			fileInfo, err := os.Stat(filename)
			if err != nil {
				return nil, err
			}
			basename := filepath.Base(filename)
			size = fileInfo.Size()
			bar := progress.Add(size,
				mpb.NewBarFiller(BarFiller),
				mpb.PrependDecorators(
					decor.Name("elapsed", decor.WC{W: len("elapsed") + 1, C: decor.DidentRight}),
					decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
					decor.Name(basename, decor.WC{W: len(basename) + 1, C: decor.DidentRight}),
				),
				mpb.AppendDecorators(
					decor.Percentage(decor.WC{W: 6, C: decor.DidentRight}),
					decor.CountersKibiByte("% .2f / % .2f"),
				),
			)

			proxyReader := bar.ProxyReader(r.Body)
			r.Body = proxyReader
			defer proxyReader.Close()

			return r, nil
		},
	})
}

func AddProgress(cmd *cobra.Command, fileFlag string, progress *mpb.Progress) func(r *http.Request) (*http.Request, error) {
	return func(r *http.Request) (*http.Request, error) {
		if r.Body == nil || progress == nil {
			return r, nil
		}
		filename, err := cmd.Flags().GetString(fileFlag)
		if err != nil {
			// Don't return error here, as if the user does not provide a file, then ignore the progress
			return r, nil
		}

		var size int64
		fileInfo, err := os.Stat(filename)
		if err != nil {
			return r, err
		}

		basename := filepath.Base(filename)
		size = fileInfo.Size()
		bar := progress.Add(size,
			mpb.NewBarFiller(BarFiller),
			mpb.PrependDecorators(
				decor.Name("elapsed", decor.WC{W: len("elapsed") + 1, C: decor.DidentRight}),
				decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
				decor.Name(basename, decor.WC{W: len(basename) + 1, C: decor.DidentRight}),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 6, C: decor.DidentRight}),
				decor.CountersKibiByte("% .2f / % .2f"),
			),
		)

		proxyReader := bar.ProxyReader(r.Body)
		r.Body = proxyReader

		return r, nil
	}
}

func CreateProxyReader(progress *mpb.Progress) func(response *http.Response) io.Reader {
	return func(r *http.Response) io.Reader {
		size := int64(r.ContentLength)
		basename := "download"

		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Disposition"))
		if err == nil {
			if filename, ok := params["filename"]; ok {
				basename = filename
			}
		}

		bar := progress.Add(size,
			mpb.NewBarFiller(BarFiller),
			mpb.PrependDecorators(
				decor.Name("elapsed", decor.WC{W: len("elapsed") + 1, C: decor.DidentRight}),
				decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
				decor.Name(basename, decor.WC{W: len(basename) + 1, C: decor.DidentRight}),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 6, C: decor.DidentRight}),
				decor.CountersKibiByte("% .2f / % .2f"),
			),
		)

		proxyReader := bar.ProxyReader(r.Body)
		r.Body = proxyReader
		return proxyReader
	}
}
