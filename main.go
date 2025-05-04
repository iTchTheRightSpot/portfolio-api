package main

import (
	"context"
	"github.com/iTchTheRightSpot/utility/middleware"
	"github.com/iTchTheRightSpot/utility/utils"
	"github.com/rs/cors"
	"github.com/syumai/workers/cloudflare"
	"net/http"
	"time"

	"github.com/syumai/workers"
)

func main() {
	ui := cloudflare.Getenv("FRONTEND")
	discord := cloudflare.Getenv("DISCORD")

	lg := utils.ProdLogger(time.RFC3339, "UTC", discord)
	m := middleware.Middleware{Logger: lg}

	han := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			utils.ErrorResponse(w, &utils.BadRequestError{})
			return
		}
		lg.Log(r.Context(), name+" is visiting your portfolio", "frontend "+ui)
		w.WriteHeader(204)
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{ui},
		AllowedMethods:   []string{http.MethodPost},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
	})

	lg.Log(context.Background(), "server listening on default port 9900")
	workers.Serve(m.Log(c.Handler(m.Panic(han))))
}
