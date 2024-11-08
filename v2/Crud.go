package crud

import (
	"context"
	"net/http"
	"strings"

	"github.com/gouniverse/bs"
	"github.com/gouniverse/cdn"
	"github.com/gouniverse/form"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/utils"
	"github.com/samber/lo"
)

type Crud struct {
	columnNames         []string
	createFields        []form.FieldInterface
	endpoint            string
	entityNamePlural    string
	entityNameSingular  string
	fileManagerURL      string
	funcCreate          func(data map[string]string) (userID string, err error)
	funcReadExtras      func(entityID string) []hb.TagInterface
	funcFetchReadData   func(entityID string) ([][2]string, error)
	funcFetchUpdateData func(entityID string) (map[string]string, error)
	funcLayout          func(w http.ResponseWriter, r *http.Request, title string, content string, styleFiles []string, style string, jsFiles []string, js string) string
	funcRows            func() (rows []Row, err error)
	funcTrash           func(entityID string) error
	funcUpdate          func(entityID string, data map[string]string) error
	homeURL             string
	readFields          []form.FieldInterface
	updateFields        []form.FieldInterface
}

func (crud Crud) Handler(w http.ResponseWriter, r *http.Request) {
	path := utils.Req(r, "path", pathHome)

	if path == "" {
		path = pathHome
	}

	ctx := context.WithValue(r.Context(), "", r.URL.Path)

	routeFunc := crud.getRoute(path)
	routeFunc(w, r.WithContext(ctx))
}

func (crud *Crud) getRoute(route string) func(w http.ResponseWriter, r *http.Request) {
	routes := map[string]func(w http.ResponseWriter, r *http.Request){
		pathHome:              crud.newEntityManagerController().page,
		pathEntityCreateAjax:  crud.newEntityCreateController().modalSave,
		pathEntityCreateModal: crud.newEntityCreateController().modalShow,
		pathEntityManager:     crud.newEntityManagerController().page,
		pathEntityRead:        crud.newEntityReadController().page,
		pathEntityUpdate:      crud.newEntityUpdateController().page,
		pathEntityUpdateAjax:  crud.newEntityUpdateController().pageSave,
		pathEntityTrashAjax:   crud.newEntityTrashController().pageEntityTrashAjax,
	}
	// log.Println(route)
	if val, ok := routes[route]; ok {
		return val
	}

	return routes[pathHome]
}

// func (crud *Crud) pageEntitiesEntityCreateModal() hb.TagInterface {
// 	form := crud.form(crud.createFields)

// 	modalHeader := hb.Div().Class("modal-header").
// 		AddChild(hb.Heading5().Text("New " + crud.entityNameSingular))

// 	modalBody := hb.Div().Class("modal-body").AddChildren(form)

// 	modalFooter := hb.Div().Class("modal-footer").
// 		AddChild(hb.Button().Text("Close").Class("btn btn-secondary").Attr("data-bs-dismiss", "modal")).
// 		AddChild(hb.Button().Text("Create & Continue").Class("btn btn-primary").Attr("v-on:click", "entityCreate"))

// 	modal := hb.Div().ID("ModalEntityCreate").Class("modal fade").AddChildren([]hb.TagInterface{
// 		hb.Div().Class("modal-dialog").AddChildren([]hb.TagInterface{
// 			hb.Div().Class("modal-content").AddChildren([]hb.TagInterface{
// 				modalHeader,
// 				modalBody,
// 				modalFooter,
// 			}),
// 		}),
// 	})

// 	return modal
// }

func (crud *Crud) urlHome() string {
	url := crud.homeURL
	return url
}

func (crud *Crud) UrlEntityManager() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityManager
	return url
}

func (crud *Crud) UrlEntityCreateModal() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityCreateModal
	return url
}

func (crud *Crud) UrlEntityCreateAjax() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityCreateAjax
	return url
}

func (crud *Crud) UrlEntityTrashAjax() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityTrashAjax
	return url
}

func (crud *Crud) UrlEntityRead() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityRead
	return url
}

func (crud *Crud) UrlEntityUpdate() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityUpdate
	return url
}

func (crud *Crud) UrlEntityUpdateAjax() string {
	q := lo.Ternary(strings.Contains(crud.endpoint, "?"), "&", "?")
	url := crud.endpoint + q + "path=" + pathEntityUpdateAjax
	return url
}

// Webpage returns the webpage template for the website
func (crud *Crud) webpage(title, content string) *hb.HtmlWebpage {
	faviconImgCms := `data:image/x-icon;base64,AAABAAEAEBAQAAEABAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAmzKzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABEQEAAQERAAEAAQABAAEAAQABAQEBEQABAAEREQEAAAERARARAREAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAAi6MAALu7AAC6owAAuC8AAIkjAAD//wAA//8AAP//AAD//wAA`
	app := ""
	webpage := hb.Webpage()
	webpage.SetTitle(title)
	webpage.SetFavicon(faviconImgCms)

	webpage.AddStyleURLs([]string{
		cdn.BootstrapCss_5_3_3(),
	})
	webpage.AddScriptURLs([]string{
		cdn.BootstrapJs_5_3_3(),
		cdn.Jquery_3_7_1(),
		cdn.VueJs_3(),
		cdn.Sweetalert2_11(),
	})
	webpage.AddScripts([]string{
		app,
	})
	webpage.AddStyle(`html,body{height:100%;font-family: Ubuntu, sans-serif;}`)
	webpage.AddStyle(`body {
		font-family: "Nunito", sans-serif;
		font-size: 0.9rem;
		font-weight: 400;
		line-height: 1.6;
		color: #212529;
		text-align: left;
		background-color: #f8fafc;
	}
	.form-select {
		display: block;
		width: 100%;
		padding: .375rem 2.25rem .375rem .75rem;
		font-size: 1rem;
		font-weight: 400;
		line-height: 1.5;
		color: #212529;
		background-color: #fff;
		background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3e%3cpath fill='none' stroke='%23343a40' stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M2 5l6 6 6-6'/%3e%3c/svg%3e");
		background-repeat: no-repeat;
		background-position: right .75rem center;
		background-size: 16px 12px;
		border: 1px solid #ced4da;
		border-radius: .25rem;
		-webkit-appearance: none;
		-moz-appearance: none;
		appearance: none;
	}`)
	webpage.AddChild(hb.Raw(content))
	return webpage
}

func (crud *Crud) _breadcrumbs(breadcrumbs []Breadcrumb) string {
	nav := hb.Nav().Attr("aria-label", "breadcrumb")
	ol := hb.OL().Attr("class", "breadcrumb")

	for _, breadcrumb := range breadcrumbs {
		li := hb.LI().Attr("class", "breadcrumb-item")
		link := hb.Hyperlink().Text(breadcrumb.Name).Attr("href", breadcrumb.URL)

		li.AddChild(link)

		ol.AddChild(li)
	}

	nav.AddChild(ol)

	return nav.ToHTML()
}

// layout is a function that generates an HTML layout for a web page.
//
// Parameters:
// - w: an http.ResponseWriter object for writing the HTTP response.
// - r: a pointer to an http.Request object representing the HTTP request.
// - title: a string containing the title of the web page.
// - content: a string containing the content of the web page.
// - styleFiles: a slice of strings representing the URLs of the style files to be included in the web page.
// - style: a string containing the CSS style to be applied to the web page.
// - jsFiles: a slice of strings representing the URLs of the JavaScript files to be included in the web page.
// - js: a string containing the JavaScript code to be executed in the web page.
//
// Returns:
// - string - a string representing the generated HTML layout.
func (crud *Crud) layout(w http.ResponseWriter, r *http.Request, title string, content string, styleFiles []string, style string, jsFiles []string, js string) string {
	html := ""

	if crud.funcLayout != nil {
		// jsFiles = append([]string{"//unpkg.com/naive-ui"}, jsFiles...)
		jsFiles = append([]string{cdn.VueElementPlusJs_2_3_8()}, jsFiles...)
		jsFiles = append([]string{cdn.VueJs_3()}, jsFiles...)
		jsFiles = append([]string{cdn.Sweetalert2_11()}, jsFiles...)
		jsFiles = append([]string{cdn.Htmx_2_0_0()}, jsFiles...)
		styleFiles = append([]string{cdn.VueElementPlusCss_2_3_8()}, styleFiles...)
		html = crud.funcLayout(w, r, title, content, styleFiles, style, jsFiles, js)
	} else {
		webpage := crud.webpage(title, content)
		webpage.AddStyleURLs(styleFiles)
		webpage.AddStyle(style)
		webpage.AddScriptURLs(jsFiles)
		webpage.AddScript(js)
		html = webpage.ToHTML()
	}

	return html
}

// form generates a form with entries for each form field.
//
// Parameters:
// - fields: a slice of FormField structs representing the fields in the form.
//
// Returns:
// - a slice of hb.Tags representing the form.
func (crud *Crud) form(fields []form.FieldInterface) []hb.TagInterface {
	tags := []hb.TagInterface{}
	for _, field := range fields {
		fieldID := field.GetID()
		if fieldID == "" {
			fieldID = "id_" + utils.StrRandomFromGamma(32, "abcdefghijklmnopqrstuvwxyz1234567890")
		}
		fieldName := field.GetName()
		fieldValue := field.GetValue()
		fieldLabel := field.GetLabel()
		if fieldLabel == "" {
			fieldLabel = fieldName
		}

		formGroup := hb.Div().Class("form-group mb-3")

		formGroupLabel := hb.Label().
			Text(fieldLabel).
			Class("form-label").
			ChildIf(
				field.GetRequired(),
				hb.Sup().Text("*").Class("text-danger ml-1"),
			)

		formGroupInput := hb.Input().
			Class("form-control").
			Attr("v-model", "entityModel."+fieldName)

		if field.GetType() == FORM_FIELD_TYPE_IMAGE {
			formGroupInput = hb.Div().Children([]hb.TagInterface{
				hb.Image("").
					Attr(`v-bind:src`, `entityModel.`+fieldName+`||'https://www.freeiconspng.com/uploads/no-image-icon-11.PNG'`).
					Style(`width:200px;`),
				bs.InputGroup().Children([]hb.TagInterface{
					hb.Input().Type(hb.TYPE_URL).Class("form-control").Attr("v-model", "entityModel."+fieldName),
					hb.If(crud.fileManagerURL != "", bs.InputGroupText().Children([]hb.TagInterface{
						hb.Hyperlink().Text("Browse").Href(crud.fileManagerURL).Target("_blank"),
					})),
				}),
			})
		}

		if field.GetType() == FORM_FIELD_TYPE_IMAGE_INLINE {
			formGroupInput = hb.Div().
				Children([]hb.TagInterface{
					hb.Image("").
						Attr(`v-bind:src`, `entityModel.`+fieldName+`||'https://www.freeiconspng.com/uploads/no-image-icon-11.PNG'`).
						Style(`width:200px;`),
					hb.Input().
						Type(hb.TYPE_FILE).
						Attr("v-on:change", "uploadImage($event, '"+fieldName+"')").
						Attr("accept", "image/*"),
					hb.Button().
						HTML("See Image Data").
						Attr("v-on:click", "tmp.show_url_"+fieldName+" = !tmp.show_url_"+fieldName),
					hb.TextArea().
						Type(hb.TYPE_URL).
						Class("form-control").
						Attr("v-if", "tmp.show_url_"+fieldName).
						Attr("v-model", "entityModel."+fieldName),
				})
		}

		if field.GetType() == FORM_FIELD_TYPE_DATETIME {
			// formGroupInput = hb.Input().Type(hb.TYPE_DATETIME).Class("form-control").Attr("v-model", "entityModel."+fieldName)
			formGroupInput = hb.NewTag(`el-date-picker`).Attr("type", "datetime").Attr("v-model", "entityModel."+fieldName)
			// formGroupInput = hb.Tag(`n-date-picker`).Attr("type", "datetime").Class("form-control").Attr("v-model", "entityModel."+fieldName)
		}

		if field.GetType() == FORM_FIELD_TYPE_HTMLAREA {
			formGroupInput = hb.NewTag("trumbowyg").Attr("v-model", "entityModel."+fieldName).Attr(":config", "trumbowigConfig").Class("form-control")
		}

		if field.GetType() == FORM_FIELD_TYPE_NUMBER {
			formGroupInput.Type(hb.TYPE_NUMBER)
		}

		if field.GetType() == FORM_FIELD_TYPE_PASSWORD {
			formGroupInput.Type(hb.TYPE_PASSWORD)
		}

		if field.GetType() == FORM_FIELD_TYPE_SELECT {
			formGroupInput = hb.Select().Class("form-select").Attr("v-model", "entityModel."+fieldName)
			for _, opt := range field.GetOptions() {
				option := hb.Option().Value(opt.Key).Text(opt.Value)
				formGroupInput.AddChild(option)
			}
			if field.GetOptionsF() != nil {
				options := field.GetOptionsF()()
				for _, opt := range options {
					option := hb.Option().Value(opt.Key).Text(opt.Value)
					formGroupInput.AddChild(option)
				}
			}
		}

		if field.GetType() == FORM_FIELD_TYPE_TEXTAREA {
			formGroupInput = hb.TextArea().Class("form-control").Attr("v-model", "entityModel."+fieldName)
		}

		if field.GetType() == FORM_FIELD_TYPE_BLOCKAREA {
			formGroupInput = hb.TextArea().Class("form-control").Attr("v-model", "entityModel."+fieldName)
		}

		if field.GetType() == FORM_FIELD_TYPE_RAW {
			formGroupInput = hb.Raw(fieldValue)
		}

		formGroupInput.ID(fieldID)
		if field.GetType() != FORM_FIELD_TYPE_RAW {
			formGroup.AddChild(formGroupLabel)
		}
		formGroup.AddChild(formGroupInput)

		// Add help
		if field.GetHelp() != "" {
			formGroupHelp := hb.Paragraph().Class("text-info").HTML(field.GetHelp())
			formGroup.AddChild(formGroupHelp)
		}

		tags = append(tags, formGroup)

		if field.GetType() == FORM_FIELD_TYPE_BLOCKAREA {
			script := hb.NewTag(`component`).
				Attr(`:is`, `'script'`).
				HTML(`setTimeout(() => {
				const blockArea = new BlockArea('` + fieldID + `');
				blockArea.registerBlock(BlockAreaHeading);
				blockArea.registerBlock(BlockAreaText);
				blockArea.registerBlock(BlockAreaImage);
				blockArea.registerBlock(BlockAreaCode);
				blockArea.registerBlock(BlockAreaRawHtml);
				blockArea.init();
			}, 2000)`)
			tags = append(tags, script)
		}
	}

	return tags
}

// listCreateNames returns a list of names from the createFields
// slice in the Crud struct.
//
// Parameters:
//   - None
//
// Returns:
//   - []string - a list of field names
func (crud *Crud) listCreateNames() []string {
	names := []string{}

	for _, field := range crud.createFields {
		if field.GetName() == "" {
			continue
		}
		names = append(names, field.GetName())
	}

	return names
}

// listUpdateNames returns a list of names from the updateFields
// slice in the Crud struct.
//
// Parameters:
//   - None
//
// Returns:
//   - []string - a list of field names
func (crud *Crud) listUpdateNames() []string {
	names := []string{}

	for _, field := range crud.updateFields {
		if field.GetName() == "" {
			continue
		}
		names = append(names, field.GetName())
	}

	return names
}
