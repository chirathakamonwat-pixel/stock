package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
)

type productsService struct {
	repository repositories.IProductsRepository
}

type IProductsService interface {
	GetAll() (*[]entities.Product, error)
	GetByID(id string) (*entities.Product, error)
	Create(data map[string]interface{}) error
	Update(id string, data map[string]interface{}) error
	Delete(id string) error
}

func NewProductsService(repo repositories.IProductsRepository) IProductsService {
	return &productsService{repository: repo}
}

func (sv *productsService) GetAll() (*[]entities.Product, error) {
	return sv.repository.FindAll()
}

func (sv *productsService) GetByID(id string) (*entities.Product, error) {
	return sv.repository.FindByID(id)
}

func (sv *productsService) Create(data map[string]interface{}) error {
	return sv.repository.Insert(data)
}

func (sv *productsService) Update(id string, data map[string]interface{}) error {
	return sv.repository.Update(id, data)
}

func (sv *productsService) Delete(id string) error {
	return sv.repository.Delete(id)
}
