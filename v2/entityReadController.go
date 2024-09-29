package crud

import (
	"net/http"
	"strings"

	"github.com/gouniverse/api"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/icons"
	"github.com/gouniverse/utils"
	"github.com/samber/lo"
)

type entityReadController struct {
	crud *Crud
}

func (crud *Crud) newEntityReadController() *entityReadController {
	return &entityReadController{
		crud: crud,
	}
}

func (controller *entityReadController) page(w http.ResponseWriter, r *http.Request) {
	entityID := utils.Req(r, "entity_id", "")
	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
		return
	}

	if controller.crud.funcFetchReadData == nil {
		api.Respond(w, r, api.Error("FuncFetchReadData is required"))
		return
	}

	breadcrumbs := controller.crud._breadcrumbs([]Breadcrumb{
		{
			Name: "Home",
			URL:  controller.crud.urlHome(),
		},
		{
			Name: controller.crud.entityNameSingular + " Manager",
			URL:  controller.crud.UrlEntityManager(),
		},
		{
			Name: "View " + controller.crud.entityNameSingular,
			URL:  controller.crud.UrlEntityUpdate() + "&entity_id=" + entityID,
		},
	})

	buttonEdit := hb.Hyperlink().
		Class("btn btn-primary ml-2 float-end").
		Child(icons.Icon("bi-pencil-square", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("Edit").
		Href(controller.crud.UrlEntityUpdate() + "&entity_id=" + entityID)

	buttonCancel := hb.Hyperlink().
		Class("btn btn-secondary ml-2 float-end").
		Child(icons.Icon("bi-chevron-left", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("Back").
		Href(controller.crud.UrlEntityManager())

	heading := hb.Heading1().
		HTML("View " + controller.crud.entityNameSingular).
		Child(buttonEdit).
		Child(buttonCancel)

	container := hb.Div().
		ID("entity-read").
		Class("container").
		Child(heading).
		Child(hb.Raw(breadcrumbs))

	data, err := controller.crud.funcFetchReadData(entityID)

	table := lo.IfF(err != nil, func() hb.TagInterface {
		alert := hb.Div().
			Class("alert alert-danger").
			HTML("There was an error retrieving the data. Please try again later")

		return alert
	}).ElseF(func() hb.TagInterface {
		table := hb.Table().
			Class("table table-hover table-striped").
			Child(hb.Thead().Child(hb.TR())).
			Child(hb.Tbody().Children(lo.Map(data, func(row [2]string, _ int) hb.TagInterface {
				key := row[0]
				value := row[1]
				isRawKey := strings.HasPrefix(key, "{!!") && strings.HasSuffix(key, "!!}")
				isRawValue := strings.HasPrefix(value, "{!!") && strings.HasSuffix(value, "!!}")

				key = strings.ReplaceAll(key, "{!!", "")
				key = strings.ReplaceAll(key, "!!}", "")
				key = strings.TrimSpace(key)

				value = strings.ReplaceAll(value, "{!!", "")
				value = strings.ReplaceAll(value, "!!}", "")
				value = strings.TrimSpace(value)

				return hb.TR().Children([]hb.TagInterface{
					hb.TH().TextIf(!isRawKey, key).HTMLIf(isRawKey, key),
					hb.TD().TextIf(!isRawValue, value).HTMLIf(isRawValue, value),
				})
			})))

		return table
	})

	card := hb.Div().
		Class("card").
		Child(
			hb.Div().
				Class("card-header").
				Style(`display:flex;justify-content:space-between;align-items:center;`).
				Child(hb.Heading4().
					HTML(controller.crud.entityNameSingular + " Details").
					Style("margin-bottom:0;display:inline-block;")).
				Child(buttonEdit),
		).
		Child(
			hb.Div().
				Class("card-body").
				Child(table))

	container.Child(card)
	if controller.crud.funcReadExtras != nil {
		container.Children(controller.crud.funcReadExtras(entityID))
	}
	content := container.ToHTML()
	title := "View " + controller.crud.entityNameSingular
	html := controller.crud.layout(w, r, title, content, []string{}, "", []string{}, "")

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
