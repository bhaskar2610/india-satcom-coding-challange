package main

import (
	"fmt"
	"log"
	"net/http"

	"task-management-api/internal/handler"
	"task-management-api/internal/repository"
	"task-management-api/internal/service"
)

func main() {
	// Wire the dependency graph using constructor injection.
	repo := repository.NewInMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":8080"
	fmt.Printf("Task Management API running on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
