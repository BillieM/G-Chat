package main

import (
	"fmt"
	"html/template"
	"io"
)

func servePageTemplate(w io.Writer, templateName string, data any) {

	t := template.Must(
		template.ParseFiles(
			"templates/base.html",
			"templates/components/nav.html",
			fmt.Sprintf("templates/pages/%s.html", templateName),
		),
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
