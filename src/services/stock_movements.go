package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
)

type stockMovementsService struct {
	repository repositories.IStockMovementsRepository
}

type IStockMovementsService interface {
	GetAll() (*[]entities.StockMovement, error)
	Insert(data entities.StockMovement) error
}

func NewStockMovementsService(repo repositories.IStockMovementsRepository) IStockMovementsService {
	return &stockMovementsService{
		repository: repo,
	}
}

func (sv *stockMovementsService) GetAll() (*[]entities.StockMovement, error) {
	return sv.repository.FindAll()
}

func (sv *stockMovementsService) Insert(data entities.StockMovement) error {
	return sv.repository.Insert(data)
}
