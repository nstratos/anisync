package mal

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Status specifies a status for anime and manga entries.
type Status int

// Anime and manga entries have a status such as completed, on hold and
// dropped.
//
// Current is for entries marked as currently watching or reading.
//
// Planned is for entries marked as plan to watch or read.
const (
	Current   Status = 1
	Completed        = 2
	OnHold           = 3
	Dropped          = 4
	Planned          = 6
)

const (
	defaultBaseURL             = "https://myanimelist.net/"
	defaultListEndpoint        = "malappinfo.php"
	defaultAccountEndpoint     = "api/account/verify_credentials.xml"
	defaultAnimeAddEndpoint    = "api/animelist/add/"
	defaultAnimeUpdateEndpoint = "api/animelist/update/"
	defaultAnimeDeleteEndpoint = "api/animelist/delete/"
	defaultAnimeSearchEndpoint = "api/anime/search.xml"
	defaultMangaAddEndpoint    = "api/mangalist/add/"
	defaultMangaUpdateEndpoint = "api/mangalist/update/"
	defaultMangaDeleteEndpoint = "api/mangalist/delete/"
	defaultMangaSearchEndpoint = "api/manga/search.xml"
)

// Client manages communication with the MyAnimeList API.
type Client struct {
	client *http.Client

	username string
	password string

	// Base URL for MyAnimeList API requests.
	BaseURL *url.URL

	Account *AccountService
	Anime   *AnimeService
	Manga   *MangaService
}

// Auth is an option that can be passed to NewClient. It allows to specify the
// username and password to be used for authenticating with the MyAnimeList
// API. When this option is used, the client will use basic authentication on
// the requests than need them.
//
// Most API methods require authentication so it is typical to pass this option
// when creating a new client.
func Auth(username, password string) func(*Client) {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// HTTPClient is an option that can be passed to NewClient. It allows to
// specify the HTTP client that will be used to create the requests. If this
// option is not set, a default HTTP client (http.DefaultClient) will be used
// which is usually sufficient.
//
// This option can be set for less trivial cases, when more control over the
// created HTTP requests is required. One such example is providing a timeout
// to cancel requests that exceed it.
func HTTPClient(httpClient *http.Client) func(*Client) {
	return func(c *Client) {
		c.client = httpClient
	}
}

// NewClient returns a new MyAnimeList API client.
func NewClient(options ...func(*Client)) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	listEndpoint, _ := url.Parse(defaultListEndpoint)
	accountEndpoint, _ := url.Parse(defaultAccountEndpoint)
	animeAddEndpoint, _ := url.Parse(defaultAnimeAddEndpoint)
	animeUpdateEndpoint, _ := url.Parse(defaultAnimeUpdateEndpoint)
	animeDeleteEndpoint, _ := url.Parse(defaultAnimeDeleteEndpoint)
	animeSearchEndpoint, _ := url.Parse(defaultAnimeSearchEndpoint)
	mangaAddEndpoint, _ := url.Parse(defaultMangaAddEndpoint)
	mangaUpdateEndpoint, _ := url.Parse(defaultMangaUpdateEndpoint)
	mangaDeleteEndpoint, _ := url.Parse(defaultMangaDeleteEndpoint)
	mangaSearchEndpoint, _ := url.Parse(defaultMangaSearchEndpoint)

	c := &Client{
		BaseURL: baseURL,
	}

	c.Account = &AccountService{
		client:   c,
		Endpoint: accountEndpoint,
	}

	c.Anime = &AnimeService{
		client:         c,
		ListEndpoint:   listEndpoint,
		AddEndpoint:    animeAddEndpoint,
		UpdateEndpoint: animeUpdateEndpoint,
		DeleteEndpoint: animeDeleteEndpoint,
		SearchEndpoint: animeSearchEndpoint,
	}

	c.Manga = &MangaService{
		client:         c,
		ListEndpoint:   listEndpoint,
		AddEndpoint:    mangaAddEndpoint,
		UpdateEndpoint: mangaUpdateEndpoint,
		DeleteEndpoint: mangaDeleteEndpoint,
		SearchEndpoint: mangaSearchEndpoint,
	}

	for _, option := range options {
		if option != nil {
			option(c)
		}
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	return c
}

// Response wraps http.Response and is returned in all the library functions
// that communicate with the MyAnimeList API. Even if an error occurs the
// response will always be returned along with the actual error so that the
// caller can further inspect it if needed. For the same reason it also keeps
// a copy of the http.Response.Body that was read when the response was first
// received.
type Response struct {
	*http.Response
	Body []byte
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.  If data
// is passed as an argument then it will first be encoded in XML and then added
// to the request body as URL encoded value data=<xml>...
// This is how the MyAnimeList requires to receive the data when adding or
// updating entries.
//
// MyAnimeList API docs: http://myanimelist.net/modules.php?go=api
func (c *Client) NewRequest(method, urlStr string, data interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var body io.Reader
	if data != nil {
		d, merr := xml.Marshal(data)
		if merr != nil {
			return nil, merr
		}
		v := url.Values{}
		v.Set("data", string(d))
		body = strings.NewReader(v.Encode())
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	return req, nil

}

// Do sends an API request and returns the API response. The API response is
// XML decoded and stored in the value pointed to by v. If XML was unable to get
// decoded, it will be returned in Response.Body along with the error so that
// the caller can further inspect it if needed.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := readResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		b := response.Body
		// enconding/xml cannot handle entity &bull;
		b = bytes.Replace(b, []byte("&bull;"), []byte("<![CDATA[&bull;]]>"), -1)
		err := xml.Unmarshal(b, v)
		if err != nil {
			return response, fmt.Errorf("cannot decode: %v", err)
		}
	}

	return response, nil
}

// ErrNoContent is returned when a MyAnimeList API method returns error 204.
var ErrNoContent = errors.New("no content")

func readResponse(r *http.Response) (*Response, error) {
	resp := &Response{Response: r}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return resp, fmt.Errorf("cannot read response body: %v", err)
	}
	resp.Body = data

	if r.StatusCode == http.StatusNoContent {
		return resp, ErrNoContent
	}

	if r.StatusCode < 200 || r.StatusCode > 299 {
		return resp, fmt.Errorf("%v %v: %d %s",
			r.Request.Method, r.Request.URL,
			r.StatusCode, string(data))
	}

	return resp, nil
}

// post sends a POST API request used by Add and Update.
func (c *Client) post(endpoint string, id int, entry interface{}, useAuth bool) (*Response, error) {
	req, err := c.NewRequest("POST", fmt.Sprintf("%s%d.xml", endpoint, id), entry)
	if err != nil {
		return nil, err
	}
	if useAuth {
		req.SetBasicAuth(c.username, c.password)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return c.Do(req, nil)
}

// delete sends a DELETE API request used by Delete.
func (c *Client) delete(endpoint string, id int, useAuth bool) (*Response, error) {
	req, err := c.NewRequest("DELETE", fmt.Sprintf("%s%d.xml", endpoint, id), nil)
	if err != nil {
		return nil, err
	}
	if useAuth {
		req.SetBasicAuth(c.username, c.password)
	}

	return c.Do(req, nil)
}

// get sends a GET API request used by List and Search.
func (c *Client) get(url string, result interface{}, useAuth bool) (*Response, error) {
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if useAuth {
		req.SetBasicAuth(c.username, c.password)
	}

	return c.Do(req, result)
}
