package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	FLARE_DEFAULT_URL = "http://localhost:8191/v1"
	FLARE_TIMEOUT_MS  = 60000
	FLARE_HTTP_OK     = 200
	FLARE_STATUS_OK   = "ok"
)

type flareRequest struct {
	Cmd        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
}

type flareSolution struct {
	Status   int    `json:"status"`
	Response string `json:"response"`
}

type flareResponse struct {
	Status   string        `json:"status"`
	Message  string        `json:"message"`
	Solution flareSolution `json:"solution"`
}

var flareClient = &http.Client{Timeout: 90 * time.Second}

func postSolver(target string) ([]byte, error) {
	payload, err := json.Marshal(flareRequest{
		Cmd:        "request.get",
		URL:        target,
		MaxTimeout: FLARE_TIMEOUT_MS,
	})
	if err != nil {
		return nil, err
	}

	resp, err := flareClient.Post(
		FLARE_DEFAULT_URL,
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("flaresolverr unreachable: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func solve(target string) (string, error) {
	body, err := postSolver(target)
	if err != nil {
		return "", err
	}

	var fr flareResponse
	if err := json.Unmarshal(body, &fr); err != nil {
		return "", fmt.Errorf("parse flaresolverr response: %w", err)
	}

	if fr.Status != FLARE_STATUS_OK {
		return "", fmt.Errorf("flaresolverr status %q: %s", fr.Status, fr.Message)
	}
	if fr.Solution.Status != FLARE_HTTP_OK {
		return "", fmt.Errorf("upstream status %d via flaresolverr", fr.Solution.Status)
	}
	return fr.Solution.Response, nil
}
