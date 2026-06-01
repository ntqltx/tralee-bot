package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	KV_API_BASE      = "https://api.cloudflare.com/client/v4"
	KV_SUB_PREFIX    = "sub:"
	KV_SEEN_KEY      = "seen"
	KV_HTTP_TIMEOUT  = 30 * time.Second
	KV_LIST_PAGESIZE = 1000
)

var kvClient = &http.Client{Timeout: KV_HTTP_TIMEOUT}

type kvKey struct {
	Name string `json:"name"`
}

type kvListResponse struct {
	Result     []kvKey `json:"result"`
	ResultInfo struct {
		Cursor string `json:"cursor"`
	} `json:"result_info"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func kvEnv() (account, namespace, token string, err error) {
	account = os.Getenv("CF_ACCOUNT_ID")
	namespace = os.Getenv("CF_KV_NAMESPACE_ID")
	token = os.Getenv("CF_API_TOKEN")

	if account == "" || namespace == "" || token == "" {
		err = fmt.Errorf("CF_ACCOUNT_ID, CF_KV_NAMESPACE_ID, CF_API_TOKEN must be set")
	}
	return
}

func kvRequest(method, path string, body io.Reader) (*http.Response, error) {
	account, namespace, token, err := kvEnv()
	if err != nil {
		return nil, err
	}

	full := fmt.Sprintf("%s/accounts/%s/storage/kv/namespaces/%s%s", KV_API_BASE, account, namespace, path)
	req, err := http.NewRequest(method, full, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return kvClient.Do(req)
}

func kvListSubscribers() ([]int64, error) {
	out := []int64{}
	cursor := ""
	for {
		q := url.Values{}
		q.Set("prefix", KV_SUB_PREFIX)
		q.Set("limit", strconv.Itoa(KV_LIST_PAGESIZE))

		if cursor != "" {
			q.Set("cursor", cursor)
		}

		resp, err := kvRequest(http.MethodGet, "/keys?"+q.Encode(), nil)
		if err != nil {
			return nil, fmt.Errorf("kv list: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("kv list status %d: %s", resp.StatusCode, string(body))
		}

		var parsed kvListResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("kv list parse: %w", err)
		}
		if !parsed.Success {
			return nil, fmt.Errorf("kv list errors: %+v", parsed.Errors)
		}

		for _, k := range parsed.Result {
			idStr := strings.TrimPrefix(k.Name, KV_SUB_PREFIX)
			id, err := strconv.ParseInt(idStr, 10, 64)

			if err != nil {
				continue
			}
			out = append(out, id)
		}

		if parsed.ResultInfo.Cursor == "" {
			break
		}
		cursor = parsed.ResultInfo.Cursor
	}
	return out, nil
}

func kvGetSeen() (map[int64]bool, error) {
	seen := map[int64]bool{}
	resp, err := kvRequest(http.MethodGet, "/values/"+KV_SEEN_KEY, nil)

	if err != nil {
		return nil, fmt.Errorf("kv get seen: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return seen, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kv get seen status %d: %s", resp.StatusCode, string(body))
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return seen, nil
	}

	var ids []int64
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, fmt.Errorf("kv parse seen: %w", err)
	}
	for _, id := range ids {
		seen[id] = true
	}
	return seen, nil
}

func kvPutSeen(seen map[int64]bool) error {
	ids := make([]int64, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	resp, err := kvRequest(http.MethodPut, "/values/"+KV_SEEN_KEY, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("kv put seen: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("kv put seen status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
