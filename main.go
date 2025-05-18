package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/cors"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/fetch"
	"io"
	"net/http"
	"strings"
	"errors"
)

var vs = fetch.NewClient()
var ui = cloudflare.Getenv("FRONTEND")
var discord = cloudflare.Getenv("DISCORD")

type dis struct {
	Id     string `json:"request_id,omitempty"`
	Ip     string `json:"ip_address,omitempty"`
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
	Status string `json:"status,omitempty"`
	Time   string `json:"time,omitempty"`
	Info   string `json:"info,omitempty"`
}

func serialize(d dis) (io.Reader, error) {
	p := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       "📄 New Log Entry",
				"description": fmt.Sprintf("Status: %s", d.Status),
				"color":       5814783, // color
				"fields": []map[string]string{
					{"name": "Request ID", "value": d.Id, "inline": "false"},
					{"name": "IP Address", "value": d.Ip, "inline": "false"},
					{"name": "Method", "value": d.Method, "inline": "false"},
					{"name": "Path", "value": d.Path, "inline": "false"},
					{"name": "Time", "value": d.Time, "inline": "false"},
					{"name": "Info", "value": d.Info, "inline": "false"},
				},
			},
		},
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(p); err != nil {
		return nil, err
	}

	return buf, nil
}

func emit(body io.Reader) error {
	res, err := vs.HTTPClient("").Post(discord, "application/json", body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}

	byts, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	
	defer func(b io.ReadCloser) { err = b.Close() }(res.Body)
	err = errors.New(string(byts))
	return err
}

func add(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}

func main() {
	han := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		b := dis{
			Id:     uuid.NewString(),
			Method: r.Method,
			Path:   r.URL.Path,
			Ip:     add(r),
			Status: "LOG",
		}

		if name == "" {
			s := "name not provided"
			fmt.Println(s)
			b.Info = s
			b.Status = "ERROR"
			i, err := serialize(b)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				if err = emit(i); err != nil {
					fmt.Println(err.Error())
				}
			}
			w.WriteHeader(400)
			return
		}

		if i, err := serialize(b); err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(500)
		} else {
			b.Info = name + " is visiting"
			if err = emit(i); err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(204)
		}
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{ui},
		AllowedMethods:   []string{http.MethodPost},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
		ExposedHeaders:   []string{"Content-Length"},
		// AllowCredentials: true,
	})

	fmt.Println("server listening on default port 9900")
	workers.Serve(c.Handler(han))
}
