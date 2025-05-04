package main

import (
	"cmp"
	"context"
	"github.com/iTchTheRightSpot/utility/middleware"
	"github.com/iTchTheRightSpot/utility/utils"
	"github.com/rs/cors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/syumai/workers"
)

func main() {
	ui := cmp.Or(os.Getenv("FRONTEND"), "http://localhost:4200")
	discord := cmp.Or(os.Getenv("DISCORD"), "private-channel-webhook")

	lg := utils.DevLogger("UTC")
	if discord != "private-channel-webhook" {
		if strings.Contains(ui, "localhost") {
			lg.Fatal("please set frontend url")
		}
		lg = utils.ProdLogger(time.RFC3339, "UTC", discord)
	}

	m := middleware.Middleware{Logger: lg}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			utils.ErrorResponse(w, &utils.BadRequestError{})
			return
		}
		lg.Log(r.Context(), name+" is visiting your portfolio")
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
	workers.Serve(m.Log(c.Handler(m.Panic(mux))))
}
