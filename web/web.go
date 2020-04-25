package web

import (
	"fmt"
	"html"
	"net/http"
	"text/template"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-tarival/index"
)

// HandleObject object for send index or db in ServerStart
type HandleObject struct {
	Index index.ReverseIndex
	DB    *pg.DB
}

type handler struct {
	tmpIndex  *template.Template
	tmpResult *template.Template
	data      HandleObject
}

// ServerStart is start the server at handle address, handle functions and index params
func ServerStart(listen string, timeout time.Duration, handle HandleObject) error {
	mux := http.NewServeMux()

	mw := logMiddleware(mux)

	server := http.Server{
		Addr:         listen,
		Handler:      mw,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	tmpIndex, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		return err
	}

	tmpResult, err := template.ParseFiles("web/templates/result.html")
	if err != nil {
		return err
	}

	h := handler{
		tmpIndex:  tmpIndex,
		tmpResult: tmpResult,
		data:      handle,
	}

	mux.HandleFunc("/", h.handleSearch)
	mux.HandleFunc("/result", h.handleResult)

	log.Info().
		Str("Interface", listen).
		Msg("Server started to listen at interface ")

	return server.ListenAndServe()
}

func (handle handler) handleResult(w http.ResponseWriter, r *http.Request) {
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

	var err error
	var searchResult []string
	if handle.data.Index != nil {
		searchResult, err = handle.data.Index.Searching(query)
	}
	if handle.data.DB != nil {
		searchResult, err = index.SearchingDB(handle.data.DB, query)
	}
	if err != nil {
		log.Error().Err(err).Msg("Searching err")
		err = handle.tmpResult.Execute(w, tmpData)
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

	err = handle.tmpResult.Execute(w, tmpData)
	if err != nil {
		log.Error().Err(err).Msg("Execute html template err")
	}
}

func (handle handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := html.EscapeString(r.FormValue("query"))

	if len(query) == 0 {
		err := handle.tmpIndex.Execute(w, struct{}{})
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
