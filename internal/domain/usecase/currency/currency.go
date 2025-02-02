package currency

import (
	"context"
	"encoding/json"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
	"time"
)

type Service interface {
	GetPrice(ctx context.Context, name string) (float64, error)
	SavePrice(ctx context.Context, currency entity.Currency) error
}

type usecase struct {
	service    Service
	currencies []string
	//log           *logger.Logger
}

func New(service Service) *usecase {
	return &usecase{
		service: service,
	}
}

func (u *usecase) StartPriceUpdateWorker() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		u.updatePrices()
	}
}

func (u *usecase) updatePrices() {
	currencies := u.currencies // Список можно получать динамически
	for _, symbol := range currencies {
		currency, err := u.fetchPriceFromExchange(symbol)
		if err != nil {
			continue
		}

		// Сохраняем в кэш
		u.service.SavePrice(context.Background(), *currency)

		// Отправляем в RabbitMQ
		u.publishPriceUpdate(*currency)
	}
}

func (u *usecase) fetchPriceFromExchange(name string) (*entity.Currency, error) {
	// Здесь должна быть реализация запроса к бирже
	// Например, через Binance API
	return &entity.Currency{
		Name:  name,
		Price: 50000.0,
	}, nil
}

func (u *usecase) publishPriceUpdate(price entity.Currency) error {
	data, err := json.Marshal(price)
	if err != nil {
		return err
	}

	_ = data

	return nil

	//return s.rabbitCh.Publish(
	//	"",       // exchange
	//	"prices", // routing key
	//	false,    // mandatory
	//	false,    // immediate
	//	amqp091.Publishing{
	//		ContentType: "application/json",
	//		Body:        data,
	//	},
	//)
}
