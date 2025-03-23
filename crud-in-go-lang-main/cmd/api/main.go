package main

import (
	"log"
	"net/http"

	"crud-in-go-lang/internal/controller"
	"crud-in-go-lang/internal/repository"
	"crud-in-go-lang/internal/router"
	"crud-in-go-lang/internal/service"
)

func main() {

	repo, err := repository.NewFileRepository("data/books.json")
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}


	svc := service.NewBookService(repo)


	ctrl := controller.NewBookController(svc)


	r := router.SetupRouter(ctrl)


	port := "8080"
	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}