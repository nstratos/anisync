package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"
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
		malAgent:   malAgent, // malAgent is produced by go generate.
	}

	// API handlers
	http.Handle("/api/check", appHandler(app.handleCheck))
	http.Handle("/api/sync", appHandler(app.handleSync))
	http.Handle("/api/test/check", appHandler(app.handleTestCheck))
	http.Handle("/api/test/sync", appHandler(app.handleTestSync))
	http.Handle("/api/mal/verify", appHandler(app.handleMALVerify))
}

type App struct {
	httpClient *http.Client
	malAgent   string
}

type appErr struct {
	err     error
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *appErr) Error() string { return fmt.Sprintf("%d %v: %v", e.Code, e.Message, e.err.Error()) }

type appHandler func(http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if err := fn(w, r); err != nil {
		if e, ok := err.(*appErr); ok {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(e.Code)
			if err := json.NewEncoder(w).Encode(e); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

func getDiff(c *anisync.Client, malUsername, hbUsername string) (*anisync.Diff, error) {
	malist, resp, err := c.GetMyAnimeList(malUsername)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, &appErr{err, fmt.Sprintf("could not get MyAnimeList for user %v", malUsername), http.StatusNotFound}
		}
		return nil, &appErr{err, "could not get MyAnimeList", resp.StatusCode}
	}

	hblist, resp, err := c.GetHBAnimeList(hbUsername)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, &appErr{err, fmt.Sprintf("could not get Hummingbird list for user %v", hbUsername), http.StatusNotFound}
		}
		return nil, &appErr{err, "could not get Hummingbird list", resp.StatusCode}
	}
	diff := anisync.Compare(malist, hblist)

	return diff, err
}

func (app *App) handleCheck(w http.ResponseWriter, r *http.Request) error {
	hbUsername := r.FormValue("hbUsername")
	malUsername := r.FormValue("malUsername")

	c := newAnisyncClient(app.httpClient, app.malAgent, r)

	diff, err := getDiff(c, malUsername, hbUsername)
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
		return &appErr{err, "could not marshal diff", http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return nil
}

func (app *App) handleSync(w http.ResponseWriter, r *http.Request) error {
	// Receiving json from POST body.
	t := struct {
		HBUsername  string `json:"hbUsername"`
		MALUsername string `json:"malUsername"`
		MALPassword string `json:"malPassword"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return &appErr{nil, "sync: could not decode body", http.StatusBadRequest}
	}

	c := newAnisyncClient(app.httpClient, app.malAgent, r)

	diff, err := getDiff(c, t.MALUsername, t.HBUsername)
	if err != nil {
		return err
	}

	err = c.VerifyMALCredentials(t.MALUsername, t.MALPassword)
	if err != nil {
		return &appErr{err, "could not verify MAL credentials", http.StatusUnauthorized}
	}

	syncResp := c.SyncMALAnime(*diff)

	bytes, err := json.Marshal(syncResp)
	if err != nil {
		return &appErr{err, "could not marshal sync response", http.StatusInternalServerError}
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

		return &appErr{nil, "verify: could not decode body", http.StatusInternalServerError}
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

	c := newAnisyncClient(app.httpClient, app.malAgent, r)

	err = c.VerifyMALCredentials(t.MALUsername, t.MALPassword)
	if err == nil {
		res.IsValid = true
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return &appErr{nil, "verify: could not encode response", http.StatusInternalServerError}
	}

	return nil
}
