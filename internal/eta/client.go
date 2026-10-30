package eta

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL string
	httpClient *http.Client
}

func New(url string) *Client {
	if url == "" {
		url = os.Getenv("PYTHON_ETA_URL")
		if url == "" {
			url = "http://localhost:8000"
		}
	}
	return &Client{baseURL: url, httpClient: &http.Client{Timeout: 2 * time.Second}}
}

func (c *Client) Predict(distKm, speedKmh float64) (float64, error) {
	if speedKmh < 1 {
		return 0, nil
	}
	if c.baseURL == "" {
		return distKm / speedKmh * 60, nil
	}
	body, _ := json.Marshal(map[string]float64{"distance": distKm, "speed": speedKmh})
	resp, err := c.httpClient.Post(c.baseURL+"/predict", "application/json", bytes.NewReader(body))
	if err != nil {
		return distKm / speedKmh * 60, nil
	}
	defer resp.Body.Close()
	var out struct{ EtaMinutes float64 `json:"eta_minutes"`; Eta float64 `json:"eta"` }
	json.NewDecoder(resp.Body).Decode(&out)
	if out.EtaMinutes != 0 {
		return out.EtaMinutes, nil
	}
	if out.Eta != 0 {
		return out.Eta, nil
	}
	return distKm / speedKmh * 60, nil
}