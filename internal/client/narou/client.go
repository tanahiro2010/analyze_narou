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

func (c *NarouClient) getRequest(url string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header = c.header

	httpResp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return nil, err
	}

	return httpResp, nil
}

func (c *NarouClient) GetNovel(ncode string) (*Novel, error) {
	param := &url.Values{}
	param.Add("ncode", ncode)
	param.Add("out", "json")

	apiUrl := c.narouURL + "novelapi/api/?" + param.Encode()
	httpResp, err := c.getRequest(apiUrl)
	if err != nil {
		fmt.Printf("Error making request in GetNovel: %v\n", err)
		return nil, err
	}

	defer httpResp.Body.Close()

	novel := &Novel{}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(httpResp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	if err := json.Unmarshal(buf.Bytes(), novel); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	err = novel.UnmarshalJSON(buf.Bytes())
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return nil, err
	}

	return novel, nil
}

func (c *NarouClient) GetRanking(bigGenre BigGenre) (*RankingResult, error) {
	param := &url.Values{}
	param.Add("biggenre", fmt.Sprintf("%d", bigGenre))
	param.Add("out", "json")

	apiUrl := c.narouURL + "rank/rankget/?" + param.Encode()
	httpResp, err := c.getRequest(apiUrl)
	if err != nil {
		fmt.Printf("Error making request in GetRanking: %v\n", err)
		return nil, err
	}

	defer httpResp.Body.Close()

	rankingResult := &RankingResult{}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(httpResp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	if err := json.Unmarshal(buf.Bytes(), rankingResult); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	err = rankingResult.UnmarshalJSON(buf.Bytes())
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return nil, err
	}

	return rankingResult, nil
}
