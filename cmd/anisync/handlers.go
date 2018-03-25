package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"
	"github.com/nstratos/go-kitsu/kitsu"
	"github.com/nstratos/go-myanimelist/mal"
)

const assetsFolder = "ui/"

func init() {

	// Preparing ui
	uiHandler := http.FileServer(http.Dir(assetsFolder))
	http.Handle("/static/", http.StripPrefix("/static", uiHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uiHandler.ServeHTTP(w, r)
	})

	app := &App{
		httpClient: http.DefaultClient,
	}

	// API handlers
	http.Handle("/api/check", appHandler(app.handleCheck))
	http.Handle("/api/sync", appHandler(app.handleSync))
	http.Handle("/api/mal-verify", appHandler(app.handleMALVerify))
	http.Handle("/api/mock/check", appHandler(app.handleTestCheck))
	http.Handle("/api/mock/sync", appHandler(app.handleTestSync))
	http.Handle("/api/mock/mal-verify", appHandler(app.handleTestMALVerify))
}

type App struct {
	httpClient *http.Client
}

type appErr struct {
	err        error
	Message    string
	Cause      string
	StatusCode int
}

func (e *appErr) Error() string {
	return fmt.Sprintf("%d %v: %v", e.StatusCode, e.Message, e.err.Error())
}

func NewAppError(err error, message string, statusCode int) *appErr {
	return &appErr{err: err, Message: message, Cause: err.Error(), StatusCode: statusCode}
}

type remoteErr struct {
	*appErr
	RemoteResponse *remoteResponse
}

type remoteResponse struct {
	StatusCode int
	Request    remoteRequest
}

type remoteRequest struct {
	Method string
	URL    string
}

func (e *remoteErr) Error() string {
	if e.RemoteResponse == nil {
		return fmt.Sprintf("%d %v: %v (No remote response)", e.StatusCode, e.Message, e.err.Error())
	}
	return fmt.Sprintf("%d %v: %v (Remote response: %d %v %v)",
		e.StatusCode,
		e.Message,
		e.err.Error(),
		e.RemoteResponse.StatusCode,
		e.RemoteResponse.Request.Method,
		e.RemoteResponse.Request.URL)
}

func newRemoteError(resp *http.Response, err error, message string, statusCode int) error {
	rerr := &remoteErr{appErr: NewAppError(err, message, statusCode)}
	if resp == nil {
		return rerr
	}
	rerr.RemoteResponse = &remoteResponse{
		StatusCode: resp.StatusCode,
		Request: remoteRequest{
			Method: resp.Request.Method,
			URL:    resp.Request.URL.String(),
		},
	}
	return rerr
}

type malErr remoteErr

func NewMALError(resp *http.Response, err error, message string, status int) error {
	return newRemoteError(resp, err, message, status)
}

type kitsuErr remoteErr

func NewKitsuError(resp *http.Response, err error, message string, status int) error {
	return newRemoteError(resp, err, message, status)
}

type appHandler func(http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if err := fn(w, r); err != nil {
		switch e := err.(type) {
		case *appErr:
			w.WriteHeader(e.StatusCode)
			writeJSON(w, e)
		case *remoteErr:
			w.WriteHeader(e.StatusCode)
			writeJSON(w, e)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func getDiff(c *anisync.Client, malUsername, kitsuEmail string) (*anisync.Diff, error) {
	malist, resp, err := c.GetMyAnimeList(malUsername)
	if err != nil {
		return nil, NewMALError(resp, err, "Could not get MyAnimeList to compare.", http.StatusConflict)
	}

	kitsuList, kitsuResp, err := c.GetKitsuAnimeList(kitsuEmail)
	if err != nil {
		return nil, NewKitsuError(kitsuResp.Response, err, "Could not get Kitsu list to compare.", http.StatusConflict)
	}
	_kitsuList := make([]anisync.Anime, 0)
	for _, a := range kitsuList {
		_kitsuList = append(_kitsuList, *a)
	}
	diff := anisync.Compare(malist, _kitsuList)

	return diff, err
}

func (app *App) handleCheck(w http.ResponseWriter, r *http.Request) error {
	// preparing anisync client
	httpcl := httpClientFromRequest(r)
	malClient := mal.NewClient(
		mal.HTTPClient(httpcl),
	)
	kitsuClient := kitsu.NewClient(httpcl)
	resources := anisync.NewResources(malClient, kitsuClient)
	c := anisync.NewClient(resources)

	malUsername := r.FormValue("malUsername")
	kitsuUserID := r.FormValue("kitsuUserID")
	diff, err := getDiff(c, malUsername, kitsuUserID)
	if err != nil {
		return err
	}

	// Including MyAnimeList account username in response.
	resp := struct {
		MalUsername string
		*anisync.Diff
	}{
		malUsername,
		diff,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		return NewAppError(err, "Check: Could not encode list difference.", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return nil
}

func (app *App) handleSync(w http.ResponseWriter, r *http.Request) error {
	// Receiving json from POST body.
	t := struct {
		KitsuUserID string `json:"kitsuUserID"`
		MALUsername string `json:"malUsername"`
		MALPassword string `json:"malPassword"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return NewAppError(err, "Sync: Could not decode request.", http.StatusBadRequest)
	}

	// preparing anisync client
	httpcl := httpClientFromRequest(r)
	malClient := mal.NewClient(
		mal.HTTPClient(httpcl),
		mal.Auth(t.MALUsername, t.MALPassword),
	)
	kitsuClient := kitsu.NewClient(httpcl)
	resources := anisync.NewResources(malClient, kitsuClient)
	c := anisync.NewClient(resources)

	diff, err := getDiff(c, t.MALUsername, t.KitsuUserID)
	if err != nil {
		return err
	}

	//if _, resp, err := c.VerifyMALCredentials(t.MALUsername, t.MALPassword); err != nil {
	//	return NewMALError(resp, err, "Sync: Could not verify MAL credentials.", http.StatusUnauthorized)
	//}

	syncResp := c.SyncMALAnime(*diff)

	diff, err = getDiff(c, t.MALUsername, t.KitsuUserID)
	if err != nil {
		return err
	}

	// Including MyAnimeList account username in response.
	resp := struct {
		MalUsername string
		Sync        *anisync.SyncResult
		*anisync.Diff
	}{
		t.MALUsername,
		syncResp,
		diff,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		return NewAppError(err, "Sync: Could not encode response.", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)

	return nil
}

/*
handleMALVerify is a handler that asks the MAL API for username and password
verification. It expects a request body that looks like this:

	{
		"malPassword": "some-password",
		"malUsername": "some-username"
	}

It returns a JSON response that is required by ngRemoteValidate, which should
look as follows:

	{
		isValid: bool, 		// Is the value received valid.
		value: 'myPassword!' 	// value received from server.
	}

In our case, we send two values to the server. As we can only return back one
value, we choose that to be the username.
*/
func (app *App) handleMALVerify(w http.ResponseWriter, r *http.Request) error {
	// Receiving json from POST body.
	t := struct {
		MALUsername string `json:"malUsername"`
		MALPassword string `json:"malPassword"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return NewAppError(nil, "Verify: Could not decode request.", http.StatusBadRequest)
	}

	// Asking MAL for verification of username and password and returning
	// a json response with the result.
	res := struct {
		IsValid bool   `json:"isValid"`
		Value   string `json:"value"` // We use username as the returned value.
	}{
		false,
		t.MALUsername,
	}

	// preparing anisync client
	httpcl := httpClientFromRequest(r)
	malClient := mal.NewClient(
		mal.HTTPClient(httpcl),
		mal.Auth(t.MALUsername, t.MALPassword),
	)
	kitsuClient := kitsu.NewClient(httpcl)
	resources := anisync.NewResources(malClient, kitsuClient)
	c := anisync.NewClient(resources)

	_, _, err = c.VerifyMALCredentials(t.MALUsername, t.MALPassword)
	if err == nil {
		res.IsValid = true
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return NewAppError(err, "Verify: Could not encode response.", http.StatusInternalServerError)
	}

	return nil
}

func (app *App) handleTestMALVerify(w http.ResponseWriter, r *http.Request) error {
	res := struct {
		IsValid bool `json:"isValid"`
	}{
		true,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return NewAppError(err, "Verify: Could not encode response.", http.StatusInternalServerError)
	}
	return nil
}
