// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"fullmetal-api/data"
	"fullmetal-api/internal/handler"
	"fullmetal-api/internal/repository"
	"fullmetal-api/internal/router"
	"fullmetal-api/internal/service"
)

func main() {
	// 依存される側から順に組み立てる：repository → service → handler
	repo, err := repository.NewNarrationRepository(data.NarrationsJSON)
	if err != nil {
		log.Fatalf("failed to load narrations: %v", err)
	}
	svc := service.NewNarrationService(repo)
	h := handler.NewNarrationHandler(svc)

	mux := router.New(h)

	// ホスティング先がPORTを指定する場合に備え、環境変数を優先する
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
