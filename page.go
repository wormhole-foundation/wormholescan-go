package wormholescan

import (
	"context"
	"iter"
)

// SortOrder is a list sort direction accepted by endpoints that support it.
type SortOrder string

const (
	// SortAsc requests ascending order.
	SortAsc SortOrder = "ASC"
	// SortDesc requests descending order.
	SortDesc SortOrder = "DESC"
)

// PageOptions selects a page of a list endpoint.
type PageOptions struct {
	// Page is the 0-based page index.
	Page int
	// PageSize is the number of items to request. 0 means the server default.
	PageSize int
	// Sort is the sort direction. The empty value means the server default.
	Sort SortOrder
}

// Page is one page of results from a list endpoint.
type Page[T any] struct {
	// Items is the page payload.
	Items []T
	// Page is the 0-based page index that produced Items.
	Page int
	// PageSize is the requested page size used to detect the end of the list.
	// Fetchers must set this to opts.PageSize.
	PageSize int
}

// Last reports whether this page ends the list (fewer items than PageSize,
// or PageSize unknown and Items empty).
func (p Page[T]) Last() bool {
	if p.PageSize <= 0 {
		return len(p.Items) == 0
	}
	return len(p.Items) < p.PageSize
}

// Paginate walks pages starting at first until [Page.Last], yielding items in
// order. The first error from fetch is yielded once and iteration stops.
func Paginate[T any](
	ctx context.Context,
	first PageOptions,
	fetch func(context.Context, PageOptions) (Page[T], error),
) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		opts := first
		for {
			if err := ctx.Err(); err != nil {
				var zero T
				yield(zero, err)
				return
			}
			page, err := fetch(ctx, opts)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
			if page.Last() {
				return
			}
			opts = nextPageOptions(opts, len(page.Items))
		}
	}
}

// paginate walks pages starting at first until [Page.Last], yielding items in
// order. The first error from fetch is yielded once and iteration stops.
func paginate[T any](
	ctx context.Context,
	first PageOptions,
	fetch func(context.Context, PageOptions) (Page[T], error),
) iter.Seq2[T, error] {
	return Paginate(ctx, first, fetch)
}

// nextPageOptions advances to the next page, pinning PageSize from n when unset.
func nextPageOptions(opts PageOptions, n int) PageOptions {
	if opts.PageSize == 0 {
		opts.PageSize = n
	}
	opts.Page++
	return opts
}
