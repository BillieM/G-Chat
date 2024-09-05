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

	indexTemplate    appTemplate = generateAppTemplate("index", pageTemplate, baseTemplate, navbarTemplate, playerCardTemplate, addToPlayerListTemplate)
	settingsTemplate appTemplate = generateAppTemplate("settings", pageTemplate, baseTemplate, navbarTemplate, messageTemplate)

	colourPickerTemplate         appTemplate = generateAppTemplate("colourpicker", componentTemplate)
	messageTemplate              appTemplate = generateAppTemplate("message", componentTemplate)
	navbarTemplate               appTemplate = generateAppTemplate("navbar", componentTemplate)
	notificationTemplate         appTemplate = generateAppTemplate("notification", componentTemplate)
	enterNotificationTemplate    appTemplate = generateAppTemplate("enternotification", componentTemplate, notificationTemplate)
	addToPlayerListTemplate      appTemplate = generateAppTemplate("addtoplayerlist", componentTemplate)
	clearPlayerListTemplate      appTemplate = generateAppTemplate("clearplayerlist", componentTemplate)
	removeFromPlayerListTemplate appTemplate = generateAppTemplate("removefromplayerlist", componentTemplate)
	exitNotificationTemplate     appTemplate = generateAppTemplate("exitnotification", componentTemplate, notificationTemplate)
	playerCardTemplate           appTemplate = generateAppTemplate("playercard", componentTemplate)
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
