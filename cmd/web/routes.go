package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	// procedury obsługi żądań http
	r.Get("/", app.showFacts)
	r.Get("/cytaty", app.showQuotes)
	r.Get("/ksiazki", app.showBooks)
	r.Get("/informacje", app.showInformation)
	r.Get("/dzien/{month}/{day}", app.showFactsByDay)
	// api
	r.Get("/api/dzien/{month}/{day}", app.apiFactsByDay)
	r.Get("/api/today", app.apiFactsToday)
	r.Get("/api/short", app.apiFactsShort) // zwraca skrócony opis dla Twittera

	// obsługa plików statycznych, w katalogu i podkatalogach pusty plik index.html
	FileServer(r, "/static/", http.Dir("./ui/static/"))

	return r
}

// FileServer - obsługa plików statycznych
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
