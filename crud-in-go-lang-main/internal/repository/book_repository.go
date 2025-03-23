package repository

import (
	"crud-in-go-lang/internal/models"
)


type BookRepository interface {

	GetAll(params models.PaginationParams) ([]models.Book, error)
	

	GetByID(id string) (*models.Book, error)
	

	Create(book models.Book) (*models.Book, error)
	

	Update(id string, book models.Book) (*models.Book, error)
	

	Delete(id string) error
	

	Search(query string) ([]models.Book, error)
	

	Count() (int, error)
}