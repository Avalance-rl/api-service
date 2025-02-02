package model

import (
	"cryptobot/api-service/internal/domain/entity"
)

type Currency struct {
}

func (c *Currency) ConvertFromEntity(currencyEntity entity.Currency) {

}

func (c *Currency) ConvertToEntity() entity.Currency {
	return entity.Currency{}
}
