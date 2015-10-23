package onepic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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
				URL string
			}
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&re)
	if err != nil {
		return "", err
	}
	fmt.Printf("%+v\n", re)
	return re.ResponseData.Results[0].URL, nil
}
