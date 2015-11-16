// Package onepic uses Google image search API to bring back a single picture
// based on a query.
package onepic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Search uses Google image search API to search for a small image based on a
// query and returns the URL of that image.
func Search(q string) (string, error) {
	apiURL := "https://ajax.googleapis.com/ajax/services/search/images?v=1.0&imgsz=small&rsz=1&q=%v"

	resp, err := http.Get(fmt.Sprintf(apiURL, url.QueryEscape(q)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var re struct {
		ResponseData struct {
			Results []struct {
				UnescapedURL string
				URL          string
			}
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&re)
	if err != nil {
		return "", err
	}
	// We use the UnescapedURL instead of URL so that it can be directly used.
	// URL may contain escaped characters which will fail to load.
	return re.ResponseData.Results[0].UnescapedURL, nil
}
