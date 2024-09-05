package gchat

import (
	"fmt"
	"html/template"
	"io"
	"log"
)

type templateType struct {
	name string
	path string
}

var (
	rootTemplate templateType = templateType{
		name: "root",
		path: "templates",
	}
	pageTemplate templateType = templateType{
		name: "page",
		path: "templates/pages",
	}
	componentTemplate templateType = templateType{
		name: "component",
		path: "templates/components",
	}
)

type appTemplate struct {
	templateName      string
	templateType      templateType
	requiredTemplates []appTemplate
}

var (
	baseTemplate appTemplate = generateAppTemplate("base", rootTemplate)

	indexTemplate    appTemplate = generateAppTemplate("index", pageTemplate, baseTemplate, navbarTemplate, playerCardTemplate)
	settingsTemplate appTemplate = generateAppTemplate("settings", pageTemplate, baseTemplate, navbarTemplate, messageTemplate)

	colourPickerTemplate   appTemplate = generateAppTemplate("colourpicker", componentTemplate)
	messageTemplate        appTemplate = generateAppTemplate("message", componentTemplate)
	navbarTemplate         appTemplate = generateAppTemplate("navbar", componentTemplate)
	notificationTemplate   appTemplate = generateAppTemplate("notification", componentTemplate)
	otherEnterRoomTemplate appTemplate = generateAppTemplate("otherenterroom", componentTemplate, notificationTemplate)
	otherLeaveRoomTemplate appTemplate = generateAppTemplate("otherleaveroom", componentTemplate, notificationTemplate)
	playerCardTemplate     appTemplate = generateAppTemplate("playercard", componentTemplate)
	playerLineTemplate     appTemplate = generateAppTemplate("playerline", componentTemplate)
)

func generateAppTemplate(templateName string, templateType templateType, requiredTemplates ...appTemplate) appTemplate {
	return appTemplate{
		templateName:      templateName,
		templateType:      templateType,
		requiredTemplates: requiredTemplates,
	}
}

func serveTemplate(w io.Writer, t appTemplate, data any) {

	var appTemplates []appTemplate = append([]appTemplate{}, t.requiredTemplates...)
	appTemplates = append(appTemplates, t)

	var templatePaths []string

	for _, appTemplate := range appTemplates {
		templatePaths = append(
			templatePaths,
			fmt.Sprintf("%s/%s.html", appTemplate.templateType.path, appTemplate.templateName),
		)
	}

	parsedTemplate := template.Must(
		template.ParseFiles(templatePaths...),
	)

	err := parsedTemplate.Execute(w, data)
	if err != nil {
		log.Printf("error rendering page template: %s: %v\n", t.templateName, err)
	}
}
