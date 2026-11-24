package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
)

type categoriesService struct {
	repository repositories.ICategoriesRepository
}

type ICategoriesService interface {
	GetAll() (*[]entities.Category, error)
	GetByID(id string) (*entities.Category, error)
	Create(data map[string]interface{}) error
	Update(id string, data map[string]interface{}) error
	Delete(id string) error
}

func NewCategoriesService(repo repositories.ICategoriesRepository) ICategoriesService {
	return &categoriesService{repository: repo}
}

func (sv *categoriesService) GetAll() (*[]entities.Category, error) {
	return sv.repository.FindAll()
}

func (sv *categoriesService) GetByID(id string) (*entities.Category, error) {
	return sv.repository.FindByID(id)
}

func (sv *categoriesService) Create(data map[string]interface{}) error {
	return sv.repository.Insert(data)
}

func (sv *categoriesService) Update(id string, data map[string]interface{}) error {
	return sv.repository.Update(id, data)
}

func (sv *categoriesService) Delete(id string) error {
	return sv.repository.Delete(id)
}
