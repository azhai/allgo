package config

import (
	"fmt"
	"net/http"
)

// NewWithUrl is short for Request GET a specified URL.
func NewWithUrl(url string, auth ...string) *Environ {
	req, err := http.NewRequest("GET", url, nil)
	if err == nil && len(auth) == 2 {
		// Set username and password if provided
		req.SetBasicAuth(auth[0], auth[1])
	}
	return NewWithRequest(req)
}

// NewWithRequest creates a new Environ instance and loads environment variables from the specified request.
func NewWithRequest(req *http.Request) *Environ {
	env := &Environ{storage: make(map[string]Entry)}
	res, err := http.DefaultClient.Do(req)
	if res != nil && res.StatusCode == http.StatusOK {
		err = env.Load(res.Body, err)
	} else {
		url := req.URL.String()
		err = fmt.Errorf("env: failed to open %s: %v", url, err)
	}
	if err != nil {
		panic(err)
	}
	return env
}
