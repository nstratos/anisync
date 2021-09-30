package kitsu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nstratos/go-kitsu/kitsu/internal/jsonapi"
)

const (
	defaultBaseURL    = "https://kitsu.io/"
	defaultAPIVersion = "api/edge/"

	defaultMediaType = "application/vnd.api+json"
)

// Client manages communication with the kitsu.io API.
type Client struct {
	client *http.Client

	BaseURL *url.URL

	common service

	Anime   *AnimeService
	User    *UserService
	Library *LibraryService
}

type service struct {
	client *Client
}

// NewClient returns a new kitsu.io API client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL}

	c.common.client = c

	c.Anime = (*AnimeService)(&c.common)
	c.User = (*UserService)(&c.common)
	c.Library = (*LibraryService)(&c.common)

	return c
}

// URLOption allows to specify URL parameters to the Kitsu API to change the
// data that will be retrieved.
type URLOption func(v *url.Values)

// Pagination allows to choose how many pages of a resource to receive by
// specifying pagination parameters limit and offset. Resources are paginated
// by default.
func Pagination(limit, offset int) URLOption {
	return func(v *url.Values) {
		v.Set("page[limit]", strconv.Itoa(limit))
		v.Set("page[offset]", strconv.Itoa(offset))
	}
}

// Limit allows to control the number of results that will be retrieved. It can
// be used together with Offset to control the pagination results. Results have
// a default limit.
func Limit(limit int) URLOption {
	return func(v *url.Values) {
		v.Set("page[limit]", strconv.Itoa(limit))
	}
}

// Offset is meant to be used together with Limit and allows to control the
// offset of the pagination.
func Offset(offset int) URLOption {
	return func(v *url.Values) {
		v.Set("page[offset]", strconv.Itoa(offset))
	}
}

// Filter allows to query data that contains certain matching attributes or
// relationships. For example, to retrieve all the anime of the Action genre,
// "genres" can be passed as the attribute and "action" as one of the values
// like so:
//
//     Filter("genres", "action").
//
// More than one values can be provided to be filtered:
//
//     Filter("genres", "action", "drama").
//
// Some resources support additional filters.
//
// Anime: text, season, streamers
//
// Manga: text
//
// Drama: text
//
// LibraryEntry: userId
func Filter(attribute string, values ...string) URLOption {
	return func(v *url.Values) {
		v.Set(fmt.Sprintf("filter[%s]", attribute), strings.Join(values, ","))
	}
}

// Search can be passed as an option and allows to search for media based on
// query text.
//
// Search can only be used for media such as Anime and Manga. Passing the
// search option to one of the User methods will return an error. Alternatively
// the Filter option with the "name" attribute could be used instead.
func Search(query string) URLOption {
	return func(v *url.Values) {
		v.Set("filter[text]", query)
	}
}

// Sort can be specified to provide sorting for one or more attributes. By default, sorts are applied in ascending order.
// For descending order a - can be prepended to the sort parameter (e.g.
// -averageRating for Anime).
//
// For example to sort by the attribute "averageRating" of Anime:
//
//    Sort("averageRating")
//
// And for descending order:
//
//    Sort("-averageRating")
//
// Many sort parameters can be specified:
//
//    Sort("followersCount", "-followingCount")
//
func Sort(attributes ...string) URLOption {
	return func(v *url.Values) {
		v.Set("sort", strings.Join(attributes, ","))
	}
}

// Include allows to include one or more related resources by specifying the
// relationships and successive relationships using a dot notation. For example
// for Anime to also include Casting:
//
//    Include("castings")
//
// If Casting is needed to also include Person and Character:
//
//    Include("castings.character", "castings.person"),
//
func Include(relationships ...string) URLOption {
	return func(v *url.Values) {
		v.Set("include", strings.Join(relationships, ","))
	}
}

// NewRequest creates an API request. If a relative URL is provided in urlStr,
// it will be resolved relative to the BaseURL of the Client. Relative URLs
// should always be specified without a preceding slash. If body is specified,
// it will be encoded to JSON and used as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}, opts ...URLOption) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	v := rel.Query()
	for _, opt := range opts {
		if opt != nil { // Avoid panic in case the user passes a nil option.
			opt(&v)
		}
	}
	rel.RawQuery = v.Encode()

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err = jsonapi.Encode(buf, body); err != nil {
			return nil, err
		}
	}

	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-type", defaultMediaType)
	}
	req.Header.Set("Accept", defaultMediaType)

	return req, nil
}

// Response is a Kitsu API response. It wraps the standard http.Response
// returned from the request and provides access to pagination offsets for
// responses that return many results.
type Response struct {
	*http.Response

	Offset PageOffset
}

// PageOffset holds the offset values for each pagination link that is returned
// in the JSON API document. It is contained in the Response that is returned
// from each method of the API. A common usage is to use the Next value
// together with the Pagination option to access the next page of results.
type PageOffset struct {
	Next, Prev, First, Last int
}

func makePageOffset(o jsonapi.Offset) PageOffset {
	return PageOffset{
		First: o.First,
		Next:  o.Next,
		Last:  o.Last,
		Prev:  o.Prev,
	}
}

func newResponse(r *http.Response) *Response {
	return &Response{Response: r}
}

// Do sends an API request and returns the API response. If an API error has
// occurred both the response and the error will be returned in case the caller
// wishes to inspect the response further.
//
// If v is passed as an argument, then the API response (assuming it is a valid
// JSON API document) is decoded and stored to v.
//
// Decoding requires v to be a pointer to a struct. For example:
//
//   a := new(Anime)
//   c.client.Do(req, a)
//
// Alternatively you may pass the address of a slice of pointers to structs:
//
//   var anime []*Anime
//   c.client.Do(req, &anime)
//
// Do closes the response body on return.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {

	// Do HTTP request.
	dumpRequest(req, true) // only when built with -tags=debug
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dumpResponse(resp, true) // only when built with -tags=debug

	// Check response for errors.
	if err = checkResponse(resp); err != nil {
		// Despite the error, the response is still returned in case the caller
		// wishes to inspect it further.
		return newResponse(resp), err
	}

	// No v passed, nothing to do.
	if v == nil {
		return newResponse(resp), nil
	}

	// Decode response body to v.
	o, err := jsonapi.Decode(resp.Body, v)
	if err != nil {
		return newResponse(resp), err
	}
	response := &Response{
		Response: resp,
		Offset:   makePageOffset(o),
	}

	return response, nil
}

// ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Errors   []Error        `json:"errors"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %+v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Errors)
}

// Error holds the details of each invidivual error in an ErrorResponse.
//
// JSON API docs: http://jsonapi.org/format/#error-objects
type Error struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
	Status string `json:"status"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v: error %v: %v(%v)",
		e.Status, e.Code, e.Title, e.Detail)
}

// checkResponse checks the API response for errors and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range.
//
// API error responses are expected to have either no response body, or a JSON
// response body that maps to ErrorResponse. Any other response body will be
// silently ignored.
func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	body, err := ioutil.ReadAll(r.Body)
	if err == nil && body != nil {
		_ = json.Unmarshal(body, errorResponse)
	}
	return errorResponse
}
