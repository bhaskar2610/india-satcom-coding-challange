package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bhaskar2610/india-satcom-coding-challange/internal/handler"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/repository"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/service"
)

func main() {
	// Wire the dependency graph using constructor injection.
	repo := repository.NewInMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":8083"
	fmt.Printf("Task Management API running on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
