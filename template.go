package main

import (
	"fmt"
	"html/template"
	"io"
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

	t.Execute(w, data)
}

func serveComponentTemplate(w io.Writer, templateName string, data any) {
	t := template.Must(
		template.ParseFiles(
			fmt.Sprintf("templates/components/%s.html", templateName),
		),
	)

	t.Execute(w, data)
}
