package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"crud-in-go-lang/internal/controller"
	"crud-in-go-lang/internal/models"
	"crud-in-go-lang/internal/repository"
	"crud-in-go-lang/internal/service"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestEnvironment(t *testing.T) (*repository.FileRepository, *service.BookService, *controller.BookController, func()) {

	tmpFile, err := os.CreateTemp("", "test-books-*.json")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	
	tmpFile.Write([]byte("[]"))
	tmpFile.Close()
	

	repo, err := repository.NewFileRepository(tmpFile.Name())
	if err != nil {
		t.Fatalf("Could not create repository: %v", err)
	}
	
	svc := service.NewBookService(repo)
	ctrl := controller.NewBookController(svc)
	

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}
	
	return repo, svc, ctrl, cleanup
}


func createTestBooks(t *testing.T, ctrl *controller.BookController, count int) []models.Book {
	books := make([]models.Book, 0, count)
	
	for i := 0; i < count; i++ {
		book := models.Book{
			Title:           fmt.Sprintf("Test Book %d", i+1),
			AuthorID:        fmt.Sprintf("author-%d", i+1),
			PublisherID:     fmt.Sprintf("publisher-%d", i+1),
			PublicationDate: "2023-01-01",
			ISBN:            fmt.Sprintf("ISBN-%d", i+1),
			Pages:           100 + i*10,
			Genre:           "Test Genre",
			Description:     fmt.Sprintf("Description for test book %d", i+1),
			Price:           10.99 + float64(i),
			Quantity:        5 + i,
		}
		

		jsonBook, _ := json.Marshal(book)
		req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(jsonBook))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		
		handlerFunc := http.HandlerFunc(ctrl.Create)
		handlerFunc.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusCreated, rr.Code)
		
		var createdBook models.Book
		json.Unmarshal(rr.Body.Bytes(), &createdBook)
		books = append(books, createdBook)
	}
	
	return books
}

func TestGetAllBooks(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	_ = createTestBooks(t, ctrl, 5)
	

	req, _ := http.NewRequest("GET", "/books", nil)
	rr := httptest.NewRecorder()
	

	router := mux.NewRouter()
	router.HandleFunc("/books", ctrl.GetAll).Methods("GET")
	router.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	
	booksArray, ok := response["books"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 5, len(booksArray))
	
	totalCount, ok := response["total_count"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(5), totalCount)
}

func TestGetAllBooksPagination(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	createTestBooks(t, ctrl, 10)
	

	testCases := []struct {
		limit          int
		offset         int
		expectedCount  int
		expectedStatus int
	}{
		{3, 0, 3, http.StatusOK},       
		{3, 3, 3, http.StatusOK},       
		{3, 9, 1, http.StatusOK},       
		{5, 10, 0, http.StatusOK},      
		{0, 0, 10, http.StatusOK},     
		{-1, 0, 10, http.StatusOK},     
		{5, -1, 5, http.StatusOK},      
	}
	
	for _, tc := range testCases {

		req, _ := http.NewRequest("GET", 
			fmt.Sprintf("/books?limit=%d&offset=%d", tc.limit, tc.offset), 
			nil)
		rr := httptest.NewRecorder()
		

		router := mux.NewRouter()
		router.HandleFunc("/books", ctrl.GetAll).Methods("GET")
		router.ServeHTTP(rr, req)
		

		assert.Equal(t, tc.expectedStatus, rr.Code)
		

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		

		booksArray, ok := response["books"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, tc.expectedCount, len(booksArray))
	}
}

func TestGetBookByID(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	books := createTestBooks(t, ctrl, 1)
	bookID := books[0].BookID
	

	req, _ := http.NewRequest("GET", fmt.Sprintf("/books/%s", bookID), nil)
	req = mux.SetURLVars(req, map[string]string{"id": bookID})
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.GetByID)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var responseBook models.Book
	json.Unmarshal(rr.Body.Bytes(), &responseBook)
	
	assert.Equal(t, bookID, responseBook.BookID)
	assert.Equal(t, books[0].Title, responseBook.Title)
}

func TestGetBookByIDNotFound(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	req, _ := http.NewRequest("GET", "/books/non-existent-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "non-existent-id"})
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.GetByID)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreateBook(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	book := models.Book{
		Title:           "New Test Book",
		AuthorID:        "test-author",
		PublisherID:     "test-publisher",
		PublicationDate: "2023-05-15",
		ISBN:            "1234567890",
		Pages:           200,
		Genre:           "Fiction",
		Description:     "A test book description",
		Price:           19.99,
		Quantity:        10,
	}
	

	jsonBook, _ := json.Marshal(book)
	req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(jsonBook))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.Create)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusCreated, rr.Code)
	
	var responseBook models.Book
	json.Unmarshal(rr.Body.Bytes(), &responseBook)
	
	assert.NotEmpty(t, responseBook.BookID)
	assert.Equal(t, book.Title, responseBook.Title)
	assert.Equal(t, book.Description, responseBook.Description)
}

func TestCreateBookInvalidJSON(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	req, _ := http.NewRequest("POST", "/books", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.Create)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateBook(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	books := createTestBooks(t, ctrl, 1)
	existingBook := books[0]
	

	updatedBook := existingBook
	updatedBook.Title = "Updated Title"
	updatedBook.Description = "Updated description"
	updatedBook.Price = 29.99
	

	jsonBook, _ := json.Marshal(updatedBook)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/books/%s", existingBook.BookID), bytes.NewBuffer(jsonBook))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": existingBook.BookID})
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.Update)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var responseBook models.Book
	json.Unmarshal(rr.Body.Bytes(), &responseBook)
	
	assert.Equal(t, existingBook.BookID, responseBook.BookID)
	assert.Equal(t, "Updated Title", responseBook.Title)
	assert.Equal(t, "Updated description", responseBook.Description)
	assert.Equal(t, 29.99, responseBook.Price)
}

func TestUpdateBookNotFound(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	book := models.Book{
		Title:           "Updated Book",
		AuthorID:        "test-author",
		PublisherID:     "test-publisher",
		PublicationDate: "2023-05-15",
		ISBN:            "1234567890",
		Pages:           200,
		Genre:           "Fiction",
		Description:     "An updated description",
		Price:           19.99,
		Quantity:        10,
	}
	

	jsonBook, _ := json.Marshal(book)
	req, _ := http.NewRequest("PUT", "/books/non-existent-id", bytes.NewBuffer(jsonBook))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "non-existent-id"})
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.Update)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeleteBook(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	books := createTestBooks(t, ctrl, 1)
	bookID := books[0].BookID
	

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/books/%s", bookID), nil)
	req = mux.SetURLVars(req, map[string]string{"id": bookID})
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.Delete)
	handler.ServeHTTP(rr, req)
	
	// Check response
	assert.Equal(t, http.StatusNoContent, rr.Code)
	

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/books/%s", bookID), nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"id": bookID})
	getRr := httptest.NewRecorder()
	
	getHandler := http.HandlerFunc(ctrl.GetByID)
	getHandler.ServeHTTP(getRr, getReq)
	
	assert.Equal(t, http.StatusNotFound, getRr.Code)
}

func TestDeleteBookNotFound(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	req, _ := http.NewRequest("DELETE", "/books/non-existent-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "non-existent-id"})
	rr := httptest.NewRecorder()
	

	handler := http.HandlerFunc(ctrl.Delete)
	handler.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSearchBooks(t *testing.T) {
	_, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	books := []models.Book{
		{
			Title:           "Harry Potter",
			Description:     "A book about wizards",
			AuthorID:        "author1",
			PublisherID:     "publisher1",
			PublicationDate: "2001-01-01",
			ISBN:            "1234567890",
			Pages:           300,
			Genre:           "Fantasy",
			Price:           14.99,
			Quantity:        10,
		},
		{
			Title:           "The Hobbit",
			Description:     "A book about a hobbit on an adventure with wizards",
			AuthorID:        "author2",
			PublisherID:     "publisher2",
			PublicationDate: "1937-09-21",
			ISBN:            "0987654321",
			Pages:           310,
			Genre:           "Fantasy",
			Price:           12.99,
			Quantity:        5,
		},
		{
			Title:           "Pride and Prejudice",
			Description:     "A classic novel about relationships",
			AuthorID:        "author3",
			PublisherID:     "publisher3",
			PublicationDate: "1813-01-28",
			ISBN:            "1122334455",
			Pages:           279,
			Genre:           "Romance",
			Price:           9.99,
			Quantity:        8,
		},
	}
	
	for _, book := range books {
		jsonBook, _ := json.Marshal(book)
		req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(jsonBook))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		
		handler := http.HandlerFunc(ctrl.Create)
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusCreated, rr.Code)
	}
	
	// Test cases
	testCases := []struct {
		query    string
		expected int
	}{
		{"wizard", 2},      
		{"hobbit", 1},     
		{"classic", 1},     
		{"nonexistent", 0}, 
		{"", 3},            
	}
	
	for _, tc := range testCases {

		req, _ := http.NewRequest("GET", fmt.Sprintf("/books/search?q=%s", tc.query), nil)
		rr := httptest.NewRecorder()
		

		handler := http.HandlerFunc(ctrl.Search)
		handler.ServeHTTP(rr, req)
		

		assert.Equal(t, http.StatusOK, rr.Code)
		
		var result models.SearchResult
		json.Unmarshal(rr.Body.Bytes(), &result)
		
		assert.Equal(t, tc.expected, len(result.Books), "Search for '%s' should return %d books", tc.query, tc.expected)
		assert.Equal(t, tc.query, result.QueryString)
		assert.NotZero(t, result.SearchTime)
	}
}

func TestConcurrentOperations(t *testing.T) {
	repo, _, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()
	

	initialBooks := createTestBooks(t, ctrl, 5)
	

	done := make(chan bool)
	errChan := make(chan error, 3)
	

	go func() {
		for i := 0; i < 3; i++ {
			book := models.Book{
				Title:           fmt.Sprintf("Concurrent Book %d", i+1),
				AuthorID:        "concurrent-author",
				PublisherID:     "concurrent-publisher",
				PublicationDate: "2023-01-01",
				ISBN:            fmt.Sprintf("CONCURRENT-ISBN-%d", i+1),
				Pages:           200,
				Genre:           "Test Genre",
				Description:     "Created in concurrent test",
				Price:           15.99,
				Quantity:        3,
			}
			
			_, err := repo.Create(book)
			if err != nil {
				errChan <- err
				return
			}
		}
		done <- true
	}()
	

	go func() {
		for _, book := range initialBooks {
			updatedBook := book
			updatedBook.Title = fmt.Sprintf("Updated %s", book.Title)
			updatedBook.Price = book.Price * 1.1
			
			_, err := repo.Update(book.BookID, updatedBook)
			if err != nil {
				errChan <- err
				return
			}
		}
		done <- true
	}()
	

	go func() {
		_, err := repo.GetAll(models.PaginationParams{Limit: 100, Offset: 0})
		if err != nil {
			errChan <- err
			return
		}
		done <- true
	}()
	

	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			t.Fatalf("Error in concurrent operation: %v", err)
		case <-done:

		}
	}
	

	allBooks, err := repo.GetAll(models.PaginationParams{Limit: 100, Offset: 0})
	assert.NoError(t, err)
	assert.Equal(t, 8, len(allBooks)) 
	

	for _, book := range allBooks {
		if book.Description == "Created in concurrent test" {
			assert.Contains(t, book.Title, "Concurrent Book")
		} else {
			assert.Contains(t, book.Title, "Updated")
		}
	}
}