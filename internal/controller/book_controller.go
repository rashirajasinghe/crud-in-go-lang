package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"crud-in-go-lang/internal/models"
	"crud-in-go-lang/internal/service"
	"crud-in-go-lang/pkg/utils"

	"github.com/gorilla/mux"
)


type BookController struct {
	service *service.BookService
}


func NewBookController(service *service.BookService) *BookController {
	return &BookController{
		service: service,
	}
}


func (c *BookController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/books", c.GetAll).Methods("GET")
	router.HandleFunc("/books", c.Create).Methods("POST")
	router.HandleFunc("/books/search", c.Search).Methods("GET")
	router.HandleFunc("/books/{id}", c.GetByID).Methods("GET")
	router.HandleFunc("/books/{id}", c.Update).Methods("PUT")
	router.HandleFunc("/books/{id}", c.Delete).Methods("DELETE")
}


func (c *BookController) GetAll(w http.ResponseWriter, r *http.Request) {

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 
	offset := 0 

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	books, err := c.service.GetAll(limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving books")
		return
	}

	count, err := c.service.Count()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error counting books")
		return
	}

	response := map[string]interface{}{
		"books":       books,
		"total_count": count,
		"limit":       limit,
		"offset":      offset,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}


func (c *BookController) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	book, err := c.service.GetByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Book not found: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, book)
}


func (c *BookController) Create(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&book); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	createdBook, err := c.service.Create(book)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating book: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, createdBook)
}


func (c *BookController) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var book models.Book
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&book); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	updatedBook, err := c.service.Update(id, book)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Error updating book: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedBook)
}


func (c *BookController) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := c.service.Delete(id); err != nil {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Error deleting book: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusNoContent, nil)
}


func (c *BookController) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	
	result, err := c.service.Search(query)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error searching books: %v", err))
		return
	}
	

	w.Header().Set("X-Search-Time-Ms", fmt.Sprintf("%d", result.SearchTime))
	
	utils.RespondWithJSON(w, http.StatusOK, result)
}