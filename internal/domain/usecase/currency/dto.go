package currency

import (
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
	"strconv"
)

type GetDTO struct {
	Name string
}

type EstablishDTO struct {
	Name  string
	Price string
}

func (e EstablishDTO) ConvertToEntity() (entity.Currency, error) {
	val, err := strconv.ParseFloat(e.Price, 64)
	if err != nil {
		return entity.Currency{}, err
	}

	return entity.Currency{
		Name:  e.Name,
		Price: val,
	}, nil
}
