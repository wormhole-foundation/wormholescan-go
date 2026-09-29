package wormholescan

import (
	"context"
	"fmt"
	"time"
)

// buildDateLayout is the layout of the server's build_date field, e.g.
// "20260902053048".
const buildDateLayout = "20060102150405"

// Version identifies the server build answering requests. Wormholescan has no
// public release line: Version is an internal counter such as "0.0.91-rc1",
// Build is the source commit, and BuildDate is when it was built (UTC).
type Version struct {
	// Version is the server's own version string.
	Version string
	// Build is the git commit the server was built from.
	Build string
	// BuildDate is when the server was built, in UTC. Zero when the server
	// sent no build date.
	BuildDate time.Time
}

// GetVersion reports the build of the server behind the client's base URL.
// Use it to record which server build a result came from; the API contract is
// not versioned beyond the /api/v1 path prefix.
func (c *Client) GetVersion(ctx context.Context) (Version, error) {
	rsp, err := c.api.GetVersionWithResponse(ctx)
	if err != nil {
		return Version{}, fmt.Errorf("%sget version: %w", errPrefix, err)
	}
	if err := checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return Version{}, err
	}
	if rsp.JSON200 == nil {
		return Version{}, fmt.Errorf("%sdecode get version: no JSON body", errPrefix)
	}
	v := Version{
		Version: deref(rsp.JSON200.Version),
		Build:   deref(rsp.JSON200.Build),
	}
	if raw := deref(rsp.JSON200.BuildDate); raw != "" {
		t, err := time.Parse(buildDateLayout, raw)
		if err != nil {
			return Version{}, fmt.Errorf("%sdecode get version: build_date %q: %w", errPrefix, raw, err)
		}
		v.BuildDate = t.UTC()
	}
	return v, nil
}
