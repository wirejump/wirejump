package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const RequestTimeout = 10

// Will be called on startup to populate what's available to a user
func LoadAvailableProviders() ProvidersState {
	pstate := ProvidersState{}

	pstate.Available = map[string]WireguardProviderInitializer{
		"mullvad": MullvadInit,
		// other providers go here
	}

	for name := range pstate.Available {
		pstate.Names = append(pstate.Names, name)
	}

	return pstate
}

func FormatError(base string) func(error) {
	return func(e error) {
		log.Fatal(base, ": ", e)
	}
}

func FormatURL(base string, addTrailing bool) func(...interface{}) string {
	baselen := len(base)

	// remove trailing slash from base url if needed
	if base[baselen-1] == '/' {
		base = base[:baselen-1]
	}

	return func(parts ...interface{}) string {
		out := base

		for _, part := range parts {
			out = fmt.Sprintf("%s/%s", out, part.(string))
		}

		if addTrailing && out[len(out)-1] != '/' {
			out = out + "/"
		}

		return out
	}
}

// Generic API Request method. Should be wrapped in order for API errors to be
// decoded properly. Returns bool for API error and error for generic errors
func RequestAPI(
	HTTPMethod string,
	URL string,
	Headers http.Header,
	Data interface{},
	Dest interface{},
	APIError interface{},
) (bool, error) {
	var requestData []byte = nil

	if HTTPMethod == "POST" && Data != nil {
		payload, err := json.Marshal(Data)

		if err != nil {
			return false, errors.New("cannot encode request payload: " + err.Error())
		}

		requestData = payload
	}

	req, err := http.NewRequest(HTTPMethod, URL, bytes.NewBuffer(requestData))

	if err != nil {
		return false, errors.New("error reading request: " + err.Error())
	}

	// req.Header.Set("User-Agent", "wirejump/0.0")

	if HTTPMethod != "OPTIONS" {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range Headers {
		req.Header.Add(k, v[0])
	}

	client := &http.Client{Timeout: time.Second * RequestTimeout}
	resp, err := client.Do(req)

	if err != nil {
		return false, errors.New("error reading response: " + err.Error())
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return false, errors.New("error reading body: " + err.Error())
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if Dest == nil {
			err = nil
		} else {
			err = json.Unmarshal(body, &Dest)
		}
		if err != nil {
			return false, errors.New("failed to parse response JSON: " + err.Error())
		}
	} else {
		err = json.Unmarshal(body, APIError)

		if err != nil {
			return false, errors.New("failed to parse response JSON: " + err.Error())
		} else {
			return true, errors.New("API error")
		}
	}

	return false, nil
}
