package narou

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type NarouConfig struct {
	NarouURL  string
	UserAgent string
}

type NarouClient struct {
	narouURL string
	header   http.Header
	client   *http.Client
}

func NewNarouClient(config NarouConfig) *NarouClient {
	header := http.Header{}
	header.Add("User-Agent", config.UserAgent)

	return &NarouClient{
		narouURL: config.NarouURL,
		header:   header,
		client:   &http.Client{},
	}
}

func (c *NarouClient) GetNovel(ncode string) (*Novel, error) {
	url := &url.Values{}
	url.Add("ncode", ncode)
	url.Add("out", "json")

	req, _ := http.NewRequest(http.MethodGet, c.narouURL+"?"+url.Encode(), nil)
	req.Header = c.header

	httpResp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return nil, err
	}

	defer httpResp.Body.Close()

	novel := &Novel{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	if err := json.Unmarshal(buf.Bytes(), novel); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	novel.UnmarshalJSON(buf.Bytes())
	return novel, nil
}
