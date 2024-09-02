package gchat

import (
	"fmt"
	"html/template"
	"io"
	"log"
)

func servePageTemplate(w io.Writer, data any, pageName string, componentNames ...string) {

	templates := []string{
		"templates/base.html",
		"templates/components/navbar.html",
	}

	for _, componentName := range componentNames {
		templates = append(templates, fmt.Sprintf("templates/components/%s.html", componentName))
	}

	templates = append(templates, fmt.Sprintf("templates/pages/%s.html", pageName))

	t := template.Must(
		template.ParseFiles(templates...),
	)

	err := t.Execute(w, data)
	if err != nil {
		log.Printf("error rendering page template: %s: %v\n", pageName, err)
	}
}

func serveComponentTemplate(w io.Writer, templateName string, data any) {

	t := template.Must(
		template.ParseFiles(
			fmt.Sprintf("templates/components/%s.html", templateName),
		),
	)

	err := t.Execute(w, data)
	if err != nil {
		log.Printf("error rendering component template: %s: %v\n", templateName, err)
	}
}
