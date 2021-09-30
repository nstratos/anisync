package hb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL       = "http://hummingbird.me/"
	defaultBaseSecureURL = "https://hummingbird.me/"
)

// Client manages communication with the Hummingbird API.
type Client struct {
	client *http.Client

	BaseURL *url.URL

	User    *UserService
	Anime   *AnimeService
	Library *LibraryService
}

// NewClient returns a new Hummingbird API client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseSecureURL)

	c := &Client{client: httpClient, BaseURL: baseURL}

	c.User = &UserService{client: c}
	c.Anime = &AnimeService{client: c}
	c.Library = &LibraryService{client: c}
	return c
}

// NewClientHTTP returns a new Hummingbird API client that uses HTTP instead of
// HTTPS. This is intended for the rare cases that Hummingbird API cannot be
// accessed through HTTPS.
//
// See App Engine bug:
// https://code.google.com/p/googleappengine/issues/detail?id=12588
//
// You should probably always use NewClient instead.
func NewClientHTTP(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL}

	c.User = &UserService{client: c}
	c.Anime = &AnimeService{client: c}
	c.Library = &LibraryService{client: c}
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If body
// is passed as an argument, then it will be encoded to JSON and used as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(url)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("cannot encode body: %v", err)
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

// Do sends an API request and returns the API response. If an API error has
// occurred both the response and the error will be returned in case the caller
// wishes to further inspect the response. If v is passed as an argument, then
// the API response is JSON decoded and stored to v.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}

// checkResponse checks the API response for errors. A response is considered an
// error if it has status code outside the 200 range. API error responses are
// expected to have a JSON response body that maps to ErrorResponse. Other
// response bodies are ignored.
func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	body, err := ioutil.ReadAll(r.Body)
	if err == nil && body != nil {
		json.Unmarshal(body, errorResponse)
	}
	return errorResponse
}

// ErrorResponse represents a Hummingbird API error response.
type ErrorResponse struct {
	Response *http.Response
	Message  string `json:"error"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %v %v", r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}
