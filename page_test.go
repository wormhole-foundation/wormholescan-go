package wormholescan

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginateWalksTwoPages(t *testing.T) {
	t.Parallel()

	const pageSize = 2
	var fetches atomic.Int32
	fetch := func(_ context.Context, opts PageOptions) (Page[int], error) {
		fetches.Add(1)
		switch opts.Page {
		case 0:
			return Page[int]{Items: []int{1, 2}, Page: 0, PageSize: pageSize}, nil
		case 1:
			return Page[int]{Items: []int{3}, Page: 1, PageSize: pageSize}, nil
		default:
			return Page[int]{}, errors.New("unexpected page")
		}
	}
	seq := paginate(t.Context(), PageOptions{PageSize: pageSize}, fetch)

	var got []int
	for item, err := range seq {
		require.NoError(t, err)
		got = append(got, item)
	}
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, int32(2), fetches.Load())
}

func TestPaginatePinsPageSizeFromFirstPage(t *testing.T) {
	t.Parallel()

	var fetches atomic.Int32
	fetch := func(_ context.Context, opts PageOptions) (Page[int], error) {
		n := fetches.Add(1)
		switch n {
		case 1:
			assert.Equal(t, 0, opts.PageSize)
			return Page[int]{Items: []int{1, 2}, Page: 0, PageSize: opts.PageSize}, nil
		case 2:
			assert.Equal(t, 2, opts.PageSize)
			return Page[int]{Items: []int{3}, Page: 1, PageSize: opts.PageSize}, nil
		default:
			return Page[int]{}, errors.New("unexpected extra fetch")
		}
	}
	seq := paginate(t.Context(), PageOptions{}, fetch)

	var got []int
	for item, err := range seq {
		require.NoError(t, err)
		got = append(got, item)
	}
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, int32(2), fetches.Load())
}

func TestPaginateYieldsFetchErrorOnce(t *testing.T) {
	t.Parallel()

	want := errors.New("fetch failed")
	var fetches atomic.Int32
	seq := paginate(t.Context(), PageOptions{PageSize: 2}, func(context.Context, PageOptions) (Page[int], error) {
		fetches.Add(1)
		return Page[int]{}, want
	})

	var gotErr error
	count := 0
	for _, err := range seq {
		count++
		gotErr = err
	}
	assert.Equal(t, 1, count)
	require.ErrorIs(t, gotErr, want)
	assert.Equal(t, int32(1), fetches.Load())
}

func TestPageLast(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		page Page[int]
		want bool
	}{
		{name: "short page", page: Page[int]{Items: []int{1}, PageSize: 2}, want: true},
		{name: "full page", page: Page[int]{Items: []int{1, 2}, PageSize: 2}, want: false},
		{name: "unknown size empty", page: Page[int]{}, want: true},
		{name: "unknown size with items", page: Page[int]{Items: []int{1}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.page.Last())
		})
	}
}
