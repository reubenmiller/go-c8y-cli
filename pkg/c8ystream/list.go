package c8ystream

import (
	"context"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// ListCall assembles the standard collection Call from a single paginating
// list method, so every list command shares one consistent implementation:
//
//   - raw      -> one whole-envelope document per page via all().Pages()
//   - default  -> the extracted items via all().Items()
//
// Dry-run needs no branch here: the transport renders the first prepared
// request and returns an empty 204, so the iterator stops after one page and
// yields nothing. The printed request is the real paginating request the
// command would send, not a separate base request.
//
// all is the service's paginating list method (e.g. client.Devices.ListAll);
// opt carries the already-resolved per-item options, including pagination. T is
// the item model, inferred from the passed method value.
func ListCall[T jsondoc.Unwrapper, O any](
	raw bool,
	opt O,
	all func(context.Context, O) *pagination.Iterator[T],
) Call {
	return func(ctx context.Context) output.Seq {
		it := all(ctx, opt)
		if raw {
			return output.FromIterator(it.Pages())
		}
		return output.FromIterator(it.Items())
	}
}
