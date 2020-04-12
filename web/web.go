package web

import (
	"fmt"
	"html"
	"net/http"
	"text/template"
	"time"

	"github.com/polisgo2020/search-tarival/index"
)

// HandleObject is parameter for ServerStart
type HandleObject struct {
	Address   string
	Tmp       string
	WithIndex bool
	Index     index.ReverseIndex
}

// ServerStart is start the server at handleObjs addresses, handle functions and index params
func ServerStart(listen string, timeout time.Duration, handleObjs []HandleObject) error {
	mux := http.NewServeMux()

	server := http.Server{
		Addr:         listen,
		Handler:      mux,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	for _, handle := range handleObjs {
		tmp, err := template.ParseFiles(handle.Tmp)
		if err != nil {
			return err
		}
		if handle.WithIndex {
			mux.HandleFunc(handle.Address, func(w http.ResponseWriter, r *http.Request) {
				handleResult(w, r, tmp, handle.Index)
			})
		} else {
			mux.HandleFunc(handle.Address, func(w http.ResponseWriter, r *http.Request) {
				handleSearch(w, r, tmp)
			})
		}
	}

	fmt.Println("Server started to listen at interface ", listen)

	return server.ListenAndServe()
}

func handleResult(w http.ResponseWriter, r *http.Request, tmp *template.Template, Index index.ReverseIndex) {
	query := html.EscapeString(r.FormValue("query"))

	fmt.Printf("Get search phrase: %v\n", query)

	var results string

	searchResult, err := Index.Searching(query)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	if len(searchResult) == 0 {
		results = "Not found any result with your request"
	} else {
		for i, result := range searchResult {
			results += fmt.Sprintf("<p>%v) %v</p>\n", i+1, result)
		}
	}

	tmpData := struct {
		Results string
		Query   string
	}{
		Results: results,
		Query:   query,
	}

	err = tmp.Execute(w, tmpData)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err.Error())
		return
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request, tmp *template.Template) {
	query := html.EscapeString(r.FormValue("query"))

	if len(query) == 0 {
		err := tmp.Execute(w, struct{}{})
		if err != nil {
			fmt.Println(err)
			fmt.Fprintln(w, err.Error())
			return
		}
	} else {
		http.Redirect(w, r, "/result?query="+query, http.StatusFound)
	}
}
