package jsonapi

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/nstratos/jsonapi"
)

// Offset represents the pagination offset returned in the API response when
// there's a list of resources.
type Offset struct {
	First, Last, Prev, Next int
}

func parseOffset(links jsonapi.Links) (Offset, error) {
	m := map[string]int{"first": 0, "last": 0, "prev": 0, "next": 0}
	var err error

	for name := range m {
		val, ok := links[name]
		if ok {
			str, isString := val.(string)
			if !isString {
				return Offset{}, fmt.Errorf("%q link is not a string", name)
			}
			m[name], err = parseOffsetFromLink(str)
			if err != nil {
				return Offset{}, fmt.Errorf("failed to parse %q link: %v", name, err)
			}
		}

	}

	o := Offset{
		First: m["first"],
		Last:  m["last"],
		Prev:  m["prev"],
		Next:  m["next"],
	}

	return o, nil
}

func parseOffsetFromLink(link string) (int, error) {
	var offset int
	u, err := url.Parse(link)
	if err != nil {
		return offset, err
	}
	v := u.Query()
	s := v.Get("page[offset]")
	if s == "" {
		return offset, nil
	}
	return strconv.Atoi(s)
}
