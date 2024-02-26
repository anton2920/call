package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func GetIDFromURL(u *url.URL, urlPrefix string) (int, error) {
	urlString := u.String()
	idString := urlString[len(urlPrefix):]
	id, err := strconv.Atoi(idString)
	if err != nil {
		return 0, NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid ID for '%s'", urlPrefix))
	}
	return id, nil
}
