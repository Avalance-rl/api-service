package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
)

type ExchangeProvider struct {
	client *http.Client
	apiKey string
}

func NewExchangeProvider(apiKey string) *ExchangeProvider {
	return &ExchangeProvider{
		client: &http.Client{},
		apiKey: apiKey,
	}
}

type CoinAPIResponse struct {
	Time         string  `json:"time"`
	AssetIDBase  string  `json:"asset_id_base"`
	AssetIDQuote string  `json:"asset_id_quote"`
	Rate         float64 `json:"rate"`
}

func (p *ExchangeProvider) FetchPrice(ctx context.Context, name string) (*entity.Currency, error) {
	url := fmt.Sprintf("https://rest.coinapi.io/v1/exchangerate/%s/USD", name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("X-CoinAPI-Key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	var data CoinAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &entity.Currency{
		Name:  name,
		Price: data.Rate,
	}, nil
}
