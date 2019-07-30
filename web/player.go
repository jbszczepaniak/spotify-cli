package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type tokenToInsert struct {
	Token string
}

func PlayerHandle(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		msg := fmt.Sprintf("token is not provided")
		http.Error(w, msg, http.StatusBadRequest)
		log.Print(msg)
		return
	}

	parsedTemplate, err := template.ParseFiles("index_tmpl.html")
	if err != nil {
		msg := fmt.Sprintf("could not parse template, error: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Print(msg)
		return
	}

	err = parsedTemplate.Execute(w, tokenToInsert{token})
	if err != nil {
		msg := fmt.Sprintf("could not put insert token to template, error: %v", err)
		http.Error(w, msg, http.StatusNotFound)
		log.Print(msg)
		return
	}
}
