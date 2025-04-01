package services

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetLongURL(req *http.Request) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", errors.New("error reading body")
	}
	longURLStr := string(body)
	if longURLStr == "" {
		return "", errors.New("didnt get url")
	}
	_, err = url.ParseRequestURI(longURLStr)
	if err != nil {
		errorText := fmt.Sprintf("got incorrect url to shorten: url=%v, err=%v", longURLStr, err.Error())
		return "", errors.New(errorText)
	}
	return longURLStr, nil
}
