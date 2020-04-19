package web

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-tarival/index"
)

// HandleObject object for send index or db in ServerStart
type HandleObject struct {
	Index index.ReverseIndex
	DB    *sql.DB
}

// ServerStart is start the server at handle address, handle functions and index params
func ServerStart(listen string, timeout time.Duration, handle HandleObject) error {
	mux := http.NewServeMux()

	handler := logMiddleware(mux)

	server := http.Server{
		Addr:         listen,
		Handler:      handler,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	mux.HandleFunc("/", handleSearch)
	mux.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		handleResult(w, r, handle)
	})

	log.Info().
		Str("Interface", listen).
		Msg("Server started to listen at interface ")

	return server.ListenAndServe()
}

func handleResult(w http.ResponseWriter, r *http.Request, handle HandleObject) {
	tmp, err := template.ParseFiles("web/templates/result.html")
	if err != nil {
		log.Error().Err(err).Msg("Parse template err")
	}
	query := html.EscapeString(r.FormValue("query"))

	log.Info().Str("Get search phrase", query).Msg("Get query")

	var results string

	tmpData := struct {
		Results string
		Query   string
	}{
		Results: "",
		Query:   query,
	}

	var searchResult []string
	if handle.Index != nil {
		searchResult, err = handle.Index.Searching(query)
	}
	if handle.DB != nil {
		searchResult, err = index.SearchingDB(handle.DB, query)
	}
	if err != nil {
		log.Error().Err(err).Msg("Searching err")
		err = tmp.Execute(w, tmpData)
		if err != nil {
			log.Error().Err(err).Msg("Execute html template err")
		}
		return
	}

	if len(searchResult) == 0 {
		results = "Not found any result with your request"
	} else {
		for i, result := range searchResult {
			results += fmt.Sprintf("<p>%v) %v</p>\n", i+1, result)
		}
	}

	tmpData.Results = results

	err = tmp.Execute(w, tmpData)
	if err != nil {
		log.Error().Err(err).Msg("Execute html template err")
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		log.Error().Err(err).Msg("Parse template err")
	}

	query := html.EscapeString(r.FormValue("query"))

	if len(query) == 0 {
		err := tmp.Execute(w, struct{}{})
		if err != nil {
			log.Error().Err(err).Msg("Execute html template err")
			return
		}
	} else {
		http.Redirect(w, r, "/result?query="+query, http.StatusFound)
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Debug().
			Str("method", r.Method).
			Str("remote", r.RemoteAddr).
			Msgf("Called url %s", r.URL.Path)
		log.Info().
			Str("query", r.FormValue("query")).
			Msg("Get request")
	})
}
