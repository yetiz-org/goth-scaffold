package queryfilter

import "github.com/yetiz-org/gone/ghttp"

// FromRequest parses the q= (RSQL filter) and s= (sort) query parameters
// from an HTTP request.
//
// Returns (nil, nil, nil) when both parameters are absent or empty.
func FromRequest(req *ghttp.Request) (Node, []SortField, error) {
	node, err := ParseFilter(req.FormValue("q"))
	if err != nil {
		return nil, nil, err
	}

	return node, ParseSort(req.FormValue("s")), nil
}
