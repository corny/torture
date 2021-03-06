package main

import (
	"github.com/flosch/pongo2"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
)

type FrontendConfig struct {
	HttpListen    string
	ElasticServer string
	LogOutput     io.Writer
	PerPage       int
}

type Frontend struct {
	cfg           FrontendConfig
	elasticSearch *ElasticSearch
	templates     *pongo2.TemplateSet

	Log *log.Logger
}

func CreateFrontend(cfg FrontendConfig) (frontend *Frontend, err error) {
	frontend = &Frontend{cfg: cfg}

	// Create logger
	frontend.Log = log.New(frontend.cfg.LogOutput, "frontend: ", log.Ldate|log.Lshortfile)

	// Create an ElasticSearch connection
	frontend.elasticSearch, err = CreateElasticSearch(frontend.cfg.ElasticServer)
	if err != nil {
		return
	}

	// Create a pongo2 template set
	frontend.templates = pongo2.NewSet("torture")
	frontend.templates.SetBaseDirectory("templates")

	// Sub-Apps
	search, err := CreateSearch(SearchConfig{
		Frontend: frontend,
	})
	if err != nil {
		return
	}

	mux := httprouter.New()
	mux.Handle("GET", "/s", search.Handler)
	mux.Handler("GET", "/", http.RedirectHandler("/s", 301))
	mux.ServeFiles("/static/*filepath", http.Dir("static"))

	log.Fatal(http.ListenAndServe(frontend.cfg.HttpListen, mux))

	return
}
