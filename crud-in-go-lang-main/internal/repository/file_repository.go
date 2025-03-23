package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"crud-in-go-lang/internal/models"

	"github.com/google/uuid"
)


type FileRepository struct {
	filename string
	mutex    sync.RWMutex
}


func NewFileRepository(filename string) (*FileRepository, error) {

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}


	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()


		_, err = file.Write([]byte("[]"))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file: %w", err)
		}
	}

	return &FileRepository{
		filename: filename,
	}, nil
}


func (r *FileRepository) GetAll(params models.PaginationParams) ([]models.Book, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	books, err := r.readBooks()
	if err != nil {
		return nil, err
	}


	start := params.Offset
	end := params.Offset + params.Limit

	if start >= len(books) {
		return []models.Book{}, nil
	}

	if end > len(books) {
		end = len(books)
	}

	return books[start:end], nil
}


func (r *FileRepository) GetByID(id string) (*models.Book, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	books, err := r.readBooks()
	if err != nil {
		return nil, err
	}

	for _, book := range books {
		if book.BookID == id {
			return &book, nil
		}
	}

	return nil, fmt.Errorf("book not found with ID: %s", id)
}


func (r *FileRepository) Create(book models.Book) (*models.Book, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Generate a UUID if not provided
	if book.BookID == "" {
		book.BookID = uuid.New().String()
	}

	books, err := r.readBooks()
	if err != nil {
		return nil, err
	}


	for _, existingBook := range books {
		if existingBook.BookID == book.BookID {
			return nil, fmt.Errorf("book with ID %s already exists", book.BookID)
		}
	}


	books = append(books, book)


	if err := r.writeBooks(books); err != nil {
		return nil, err
	}

	return &book, nil
}


func (r *FileRepository) Update(id string, book models.Book) (*models.Book, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	books, err := r.readBooks()
	if err != nil {
		return nil, err
	}


	found := false
	for i, existingBook := range books {
		if existingBook.BookID == id {
			book.BookID = id 
			books[i] = book
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("book not found with ID: %s", id)
	}


	if err := r.writeBooks(books); err != nil {
		return nil, err
	}

	return &book, nil
}


func (r *FileRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	books, err := r.readBooks()
	if err != nil {
		return err
	}


	foundIndex := -1
	for i, book := range books {
		if book.BookID == id {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("book not found with ID: %s", id)
	}


	books = append(books[:foundIndex], books[foundIndex+1:]...)


	return r.writeBooks(books)
}


func (r *FileRepository) Search(query string) ([]models.Book, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	books, err := r.readBooks()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return books, nil
	}

	query = strings.ToLower(query)


	type searchResult struct {
		books []models.Book
		err   error
	}


	titleChan := make(chan searchResult)
	descChan := make(chan searchResult)


	go func() {
		var results []models.Book
		for _, book := range books {
			if strings.Contains(strings.ToLower(book.Title), query) {
				results = append(results, book)
			}
		}
		titleChan <- searchResult{books: results}
	}()


	go func() {
		var results []models.Book
		for _, book := range books {
			if strings.Contains(strings.ToLower(book.Description), query) {
				results = append(results, book)
			}
		}
		descChan <- searchResult{books: results}
	}()


	titleResults := <-titleChan
	descResults := <-descChan

	if titleResults.err != nil {
		return nil, titleResults.err
	}
	if descResults.err != nil {
		return nil, descResults.err
	}


	uniqueBooks := make(map[string]models.Book)
	
	for _, book := range titleResults.books {
		uniqueBooks[book.BookID] = book
	}
	
	for _, book := range descResults.books {
		uniqueBooks[book.BookID] = book
	}


	var result []models.Book
	for _, book := range uniqueBooks {
		result = append(result, book)
	}

	return result, nil
}


func (r *FileRepository) Count() (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	books, err := r.readBooks()
	if err != nil {
		return 0, err
	}

	return len(books), nil
}


func (r *FileRepository) readBooks() ([]models.Book, error) {
	data, err := os.ReadFile(r.filename)
	if err != nil {
		return nil, fmt.Errorf("error reading book data: %w", err)
	}

	if len(data) == 0 {
		return []models.Book{}, nil
	}

	var books []models.Book
	err = json.Unmarshal(data, &books)
	if err != nil {
		return nil, fmt.Errorf("error parsing book data: %w", err)
	}

	return books, nil
}


func (r *FileRepository) writeBooks(books []models.Book) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing book data: %w", err)
	}

	err = os.WriteFile(r.filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing book data: %w", err)
	}

	return nil
}