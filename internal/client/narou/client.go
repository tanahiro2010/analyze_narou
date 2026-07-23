package narou

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type NarouConfig struct {
	NarouURL     string
	UserAgent    string
	RankingLimit int
}

type NarouClient struct {
	narouURL     string
	rankingLimit int
	header       http.Header
	client       *http.Client
}

type RankingMode string

const (
	RankingModeDaily     RankingMode = "d"
	RankingModeWeekly    RankingMode = "w"
	RankingModeMonthly   RankingMode = "m"
	RankingModeQuarterly RankingMode = "q"
	RankingModeYearly    RankingMode = "y"
)

func (m RankingMode) novelAPIOrder() (string, error) {
	switch m {
	case RankingModeDaily:
		return "dailypoint", nil
	case RankingModeWeekly:
		return "weeklypoint", nil
	case RankingModeMonthly:
		return "monthlypoint", nil
	case RankingModeQuarterly:
		return "quarterpoint", nil
	case RankingModeYearly:
		return "yearlypoint", nil
	default:
		return "", fmt.Errorf("unsupported ranking mode for novelapi: %s", m)
	}
}

func NewNarouClient(config NarouConfig) *NarouClient {
	header := http.Header{}
	header.Add("User-Agent", config.UserAgent)

	return &NarouClient{
		narouURL:     config.NarouURL,
		rankingLimit: config.RankingLimit,
		header:       header,
		client:       &http.Client{},
	}
}

func (c *NarouClient) getRequest(url string) (*http.Response, error) {
	fmt.Printf("Making GET request to: %s\n", url)
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

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(httpResp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	response := &Response{}
	if err := json.Unmarshal(buf.Bytes(), response); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	if len(response.Novels) == 0 {
		err := fmt.Errorf("novel not found: %s", ncode)
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	return &response.Novels[0], nil
}

func (c *NarouClient) GetRanking(bigGenre BigGenre, startDate string, mode RankingMode) (*RankingResult, error) {
	param := &url.Values{}
	param.Add("biggenre", fmt.Sprintf("%d", bigGenre))
	param.Add("out", "json")
	param.Add("rtype", startDate+"-"+string(mode))

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

	return rankingResult, nil
}

func (c *NarouClient) GetRankingWithNovelAPI(bigGenre BigGenre, mode RankingMode) (*RankingWithNovelAPIResult, error) {
	order, err := mode.novelAPIOrder()
	if err != nil {
		return nil, err
	}

	param := &url.Values{}
	param.Add("biggenre", fmt.Sprintf("%d", bigGenre))
	param.Add("lim", fmt.Sprintf("%d", c.rankingLimit))
	param.Add("order", order)
	param.Add("out", "json")

	apiUrl := c.narouURL + "novelapi/api/?" + param.Encode()
	httpResp, err := c.getRequest(apiUrl)
	if err != nil {
		fmt.Printf("Error making request in GetRankingWithNovelAPI: %v\n", err)
		return nil, err
	}

	defer httpResp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(httpResp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	rankingResult := &RankingWithNovelAPIResult{}
	if err := json.Unmarshal(buf.Bytes(), rankingResult); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	return rankingResult, nil
}
