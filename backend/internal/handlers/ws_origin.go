package handlers

import (
	"net/http"
	"net/url"
)

func isAllowedWebSocketOrigin(
	r *http.Request,
	frontendURL string,
) bool {
	allowed, err := url.Parse(frontendURL)
	if err != nil || allowed.Scheme == "" || allowed.Host == "" {
		return false
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	actual, err := url.Parse(origin)
	if err != nil {
		return false
	}

	return actual.Scheme == allowed.Scheme &&
		actual.Host == allowed.Host &&
		actual.User == nil &&
		actual.Path == "" &&
		actual.RawQuery == "" &&
		actual.Fragment == ""
}
