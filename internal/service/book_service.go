package service

import (
	"time"

	"crud-in-go-lang/internal/models"
	"crud-in-go-lang/internal/repository"
)


type BookService struct {
	repo repository.BookRepository
}


func NewBookService(repo repository.BookRepository) *BookService {
	return &BookService{
		repo: repo,
	}
}


func (s *BookService) GetAll(limit, offset int) ([]models.Book, error) {

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	params := models.PaginationParams{
		Limit:  limit,
		Offset: offset,
	}

	return s.repo.GetAll(params)
}


func (s *BookService) GetByID(id string) (*models.Book, error) {
	return s.repo.GetByID(id)
}


func (s *BookService) Create(book models.Book) (*models.Book, error) {

	
	return s.repo.Create(book)
}


func (s *BookService) Update(id string, book models.Book) (*models.Book, error) {

	
	return s.repo.Update(id, book)
}


func (s *BookService) Delete(id string) error {
	return s.repo.Delete(id)
}


func (s *BookService) Search(query string) (models.SearchResult, error) {
	startTime := time.Now()
	
	books, err := s.repo.Search(query)
	if err != nil {
		return models.SearchResult{}, err
	}
	
	searchTime := time.Since(startTime).Milliseconds()
	
	return models.SearchResult{
		Books:       books,
		SearchTime:  searchTime,
		TotalCount:  len(books),
		QueryString: query,
	}, nil
}


func (s *BookService) Count() (int, error) {
	return s.repo.Count()
}