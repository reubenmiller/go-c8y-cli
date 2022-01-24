package main

import (
	"crypto/rand"
	"io"
	"io/ioutil"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

func main() {

	var total int64 = 1024 * 1024 * 500
	reader := io.LimitReader(rand.Reader, total)

	p := mpb.New()
	bar := p.AddBar(total,
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(reader)
	defer proxyReader.Close()

	// and copy from reader, ignoring errors
	_, _ = io.Copy(ioutil.Discard, proxyReader)

	p.Wait()
}
