package crud

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gouniverse/api"
	"github.com/gouniverse/bs"
	"github.com/gouniverse/cdn"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/icons"
	"github.com/gouniverse/utils"
)

const pathEntityCreateAjax = "entity-create-ajax"
const pathEntityManager = "entity-manager"
const pathEntityUpdate = "entity-update"
const pathEntityUpdateAjax = "entity-update-ajax"
const pathEntityTrashAjax = "entity-trash-ajax"

const FORM_GROUP_TYPE_STRING = "string"
const FORM_GROUP_TYPE_TEXTAREA = "textarea"
const FORM_GROUP_TYPE_SELECT = "select"

type CrudConfig struct {
	CreateFields        []FormField
	Endpoint            string
	HomeURL             string
	FileManagerURL      string
	EntityNamePlural    string
	EntityNameSingular  string
	ColumnNames         []string
	FuncCreate          func(data map[string]string) (userID string, err error)
	FuncRows            func() (rows []Row, err error)
	UpdateFields        []FormField
	FuncFetchUpdateData func(entityID string) (map[string]string, error)
	FuncTrash           func(entityID string) error
	FuncUpdate          func(entityID string, data map[string]string) error
	FuncLayout          func(w http.ResponseWriter, r *http.Request, title string, content string, styleFiles []string, style string, jsFiles []string, js string) string
}

type Crud struct {
	createFields        []FormField
	endpoint            string
	homeURL             string
	fileManagerURL      string
	entityNamePlural    string
	entityNameSingular  string
	columnNames         []string
	funcCreate          func(data map[string]string) (userID string, err error)
	funcLayout          func(w http.ResponseWriter, r *http.Request, title string, content string, styleFiles []string, style string, jsFiles []string, js string) string
	funcRows            func() (rows []Row, err error)
	funcFetchUpdateData func(entityID string) (map[string]string, error)
	funcTrash           func(entityID string) error
	funcUpdate          func(entityID string, data map[string]string) error
	updateFields        []FormField
}

type Breadcrumb struct {
	Name string
	URL  string
}
type FormFieldOption struct {
	Key   string
	Value string
}

type FormField struct {
	Type     string
	Name     string
	Label    string
	Help     string
	Options  []FormFieldOption
	OptionsF func() []FormFieldOption
}

type Row struct {
	ID   string
	Data []string
}

func NewCrud(config CrudConfig) (crud Crud, err error) {
	if config.FuncRows == nil {
		return Crud{}, errors.New("FuncRows function is required")
	}

	if config.UpdateFields == nil {
		return Crud{}, errors.New("UpdateFields is required")
	}

	isUpdateEnabled := config.FuncUpdate != nil && config.FuncFetchUpdateData != nil && len(config.UpdateFields) > 0

	if isUpdateEnabled && config.FuncUpdate == nil {
		return Crud{}, errors.New("FuncUpdate function is required")
	}

	crud = Crud{}
	crud.columnNames = config.ColumnNames
	crud.createFields = config.CreateFields
	crud.funcCreate = config.FuncCreate
	crud.funcLayout = config.FuncLayout
	crud.funcRows = config.FuncRows
	crud.funcFetchUpdateData = config.FuncFetchUpdateData
	crud.funcTrash = config.FuncTrash
	crud.funcUpdate = config.FuncUpdate
	crud.endpoint = config.Endpoint
	crud.entityNamePlural = config.EntityNamePlural
	crud.entityNameSingular = config.EntityNameSingular
	crud.homeURL = config.HomeURL
	crud.fileManagerURL = config.FileManagerURL
	crud.updateFields = config.UpdateFields

	// crud.pathEntitiesEntityCreateAjax = "entities/entity-create-ajax"
	// crud.pathEntityManager = "entities/entity-manager"
	// crud.pathEntityUpdate = "entities/entity-update"
	// crud.pathEntityUpdateAjax = "entities/entity-update-ajax"
	// crud.pathEntityTrashAjax = "entities/entity-update-ajax"

	return crud, err
}

func (crud Crud) Handler(w http.ResponseWriter, r *http.Request) {
	path := utils.Req(r, "path", "home")

	if path == "" {
		path = "home"
	}

	ctx := context.WithValue(r.Context(), "", r.URL.Path)

	routeFunc := crud.getRoute(path)
	routeFunc(w, r.WithContext(ctx))
}

func (crud Crud) getRoute(route string) func(w http.ResponseWriter, r *http.Request) {
	routes := map[string]func(w http.ResponseWriter, r *http.Request){
		"home": crud.pageEntityManager,
		// START: Custom Entities
		pathEntityCreateAjax: crud.pageEntityCreateAjax,
		pathEntityManager:    crud.pageEntityManager,
		pathEntityUpdate:     crud.pageEntityUpdate,
		pathEntityUpdateAjax: crud.pageEntityUpdateAjax,
		pathEntityTrashAjax:  crud.pageEntityTrashAjax,
		// END: Custom Entities

	}
	// log.Println(route)
	if val, ok := routes[route]; ok {
		return val
	}

	return routes["home"]
}

func (crud Crud) pageEntityCreateAjax(w http.ResponseWriter, r *http.Request) {
	names := crud._listCreateNames()
	posts := map[string]string{}
	for _, name := range names {
		posts[name] = utils.Req(r, name, "")
	}

	entityID, err := crud.funcCreate(posts)

	if err != nil {
		api.Respond(w, r, api.Error("Save failed: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("Saved successfully", map[string]interface{}{"entity_id": entityID}))
}

func (crud Crud) pageEntityManager(w http.ResponseWriter, r *http.Request) {
	// header := cms.cmsHeader(endpoint)
	breadcrums := crud._breadcrumbs([]Breadcrumb{
		{
			Name: "Home",
			URL:  crud.urlHome(),
		},
		{
			Name: crud.entityNameSingular + " Manager",
			URL:  crud.UrlEntityManager(),
		},
	})

	container := hb.NewDiv().Class("container").ID("entity-manager")
	heading := hb.NewHeading1().HTML(crud.entityNameSingular + " Manager")
	button := hb.NewButton().
		Class("btn btn-success float-end").
		Attr("v-on:click", "showEntityCreateModal").
		AddChild(icons.Icon("bi-plus-circle", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("New " + crud.entityNameSingular)
	heading.AddChild(button)

	// container.AddChild(hb.NewHTML(header))
	container.AddChild(heading)
	container.AddChild(hb.NewHTML(breadcrums))

	container.AddChild(crud.pageEntitiesEntityCreateModal())
	container.AddChild(crud.pageEntitiesEntityTrashModal())

	table := hb.NewTable().ID("TableEntities").Class("table table-responsive table-striped mt-3")
	thead := hb.NewThead()
	tbody := hb.NewTbody()
	table.AddChild(thead).AddChild(tbody)

	tr := hb.NewTR()
	for _, columnName := range crud.columnNames {
		tr.AddChild(hb.NewTH().HTML(columnName))
	}
	tr.AddChild(hb.NewTD().HTML("Actions").Attr("style", "width:120px;"))
	thead.AddChild(tr)

	rows, errRows := crud.funcRows()

	if errRows == nil {
		for _, row := range rows {
			buttonEdit := hb.NewButton().
				Class("btn btn-primary btn-sm").
				Child(icons.Icon("bi-pencil-square", 12, 12, "white").
					Style("margin-top:-4px;margin-right:4px;")).
				HTML("Edit").
				Attr("type", "button").
				Attr("v-on:click", "entityEdit('"+row.ID+"')").
				Style("margin-right:5px")

			buttonTrash := hb.NewButton().
				Class("btn btn-danger btn-sm").
				Child(icons.Icon("bi-trash", 12, 12, "white").Style("margin-top:-4px;margin-right:4px;")).
				HTML("Trash").
				Attr("type", "button").
				Attr("v-on:click", "showEntityTrashModal('"+row.ID+"')")

			tr := hb.NewTR()
			for _, cell := range row.Data {
				tr.AddChild(hb.NewTH().HTML(cell))
			}
			tr.AddChild(hb.NewTD().AddChild(buttonEdit).AddChild(buttonTrash))
			tbody.AddChild(tr)
		}
		container.Child(table)
	} else {
		alert := hb.NewDiv().
			Class("alert alert-danger").
			HTML("There was an error retrieving the data. Please try again later")
		container.Child(alert)
	}

	content := container.ToHTML()

	urlEntityCreateAjax, _ := utils.ToJSON(crud.UrlEntityCreateAjax())
	urlEntityTrashAjax, _ := utils.ToJSON(crud.UrlEntityTrashAjax())
	urlEntityUpdate, _ := utils.ToJSON(crud.UrlEntityUpdate())
	inlineScript := `
var entityCreateUrl = ` + urlEntityCreateAjax + `;
var entityTrashUrl = ` + urlEntityTrashAjax + `;
var entityUpdateUrl = ` + urlEntityUpdate + `;
const EntityManager = {
	data() {
		return {
		  entityCreateModel:{
		  },
		  entityTrashModel:{
			entityId:null,
		  }
		}
	},
	created(){
		//setTimeout(() => {
		//	console.log("Initing data table...");
			this.initDataTable();
		//}, 1000);
	},
	methods: {
		initDataTable(){
			$(() => {
				$('#TableEntities').DataTable({
					"order": [[ 0, "asc" ]] // 1st column
				});
			});
		},
        showEntityCreateModal(){
			var modalEntityCreate = new bootstrap.Modal(document.getElementById('ModalEntityCreate'));
			modalEntityCreate.show();
		},
		showEntityTrashModal(entityId){
			this.entityTrashModel.entityId = entityId;
			var modalEntityDelete = new bootstrap.Modal(document.getElementById('ModalEntityTrash'));
			modalEntityDelete.show();
		},
		entityCreate(){
		    $.post(entityCreateUrl, this.entityCreateModel).done((result)=>{
				if (result.status==="success"){
					var modalEntityCreate = new bootstrap.Modal(document.getElementById('ModalEntityCreate'));
			        modalEntityCreate.hide();
					return location.href = entityUpdateUrl+ "&entity_id=" + result.data.entity_id;
				}
				
				return Swal.fire({icon: 'error', title: 'Oops...', text: result.message});
			}).fail((result)=>{
				return Swal.fire({icon: 'error', title: 'Oops...', text: result});
			});
		},

		entityEdit(entityId){
			return location.href = entityUpdateUrl+ "&entity_id=" + entityId;
		},

		entityTrash(){
			var entityId = this.entityTrashModel.entityId;

			$.post(entityTrashUrl, {
				entity_id:entityId
			}).done((response)=>{
				if (response.status !== "success") {
					return Swal.fire({icon: 'error', title: 'Oops...', text: result.message});
				}

				setTimeout(()=>{return location.href = location.href;}, 3000)

				return Swal.fire({icon: 'success', title: 'Entity trashed'});
			}).fail((result)=>{
				console.log(result);
				return Swal.fire({icon: 'error', title: 'Oops...', text: result});
			});
		}
	}
};
Vue.createApp(EntityManager).mount('#entity-manager')
	`
	title := crud.entityNameSingular + " Manager"
	html := crud._layout(w, r, title, content, []string{
		"//cdn.datatables.net/1.13.4/css/jquery.dataTables.min.css",
	}, "html{width:100%;}", []string{
		"//cdn.datatables.net/1.13.4/js/jquery.dataTables.min.js",
	}, inlineScript)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (crud Crud) pageEntityUpdate(w http.ResponseWriter, r *http.Request) {
	// endpoint := r.Context().Value(keyEndpoint).(string)
	// log.Println(endpoint)

	entityID := utils.Req(r, "entity_id", "")
	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
		return
	}

	// header := cms.cmsHeader(r.Context().Value(keyEndpoint).(string))
	breadcrums := crud._breadcrumbs([]Breadcrumb{
		{
			Name: "Home",
			URL:  crud.urlHome(),
		},
		{
			Name: crud.entityNameSingular + " Manager",
			URL:  crud.UrlEntityManager(),
		},
		{
			Name: "Edit " + crud.entityNameSingular,
			URL:  crud.UrlEntityUpdate() + "&entity_id=" + entityID,
		},
	})

	buttonSave := hb.NewButton().Class("btn btn-success float-end").Attr("v-on:click", "entitySave(true)").
		AddChild(icons.Icon("bi-check-all", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("Save")
	buttonApply := hb.NewButton().Class("btn btn-success float-end").Attr("v-on:click", "entitySave").
		Style("margin-right:10px;").
		AddChild(icons.Icon("bi-check", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("Apply")
	heading := hb.NewHeading1().HTML("Edit " + crud.entityNameSingular).
		AddChild(buttonSave).
		AddChild(buttonApply)

	// container.AddChild(hb.NewHTML(header))
	container := hb.NewDiv().Attr("class", "container").Attr("id", "entity-update").
		AddChild(heading).
		AddChild(hb.NewHTML(breadcrums))

	customAttrValues, errData := crud.funcFetchUpdateData(entityID)

	if errData != nil {
		api.Respond(w, r, api.Error("Fetch data failed"))
		return
	}

	container.AddChildren(crud._form(crud.updateFields))

	content := container.ToHTML()

	jsonCustomValues, _ := utils.ToJSON(customAttrValues)

	urlHome, _ := utils.ToJSON(crud.endpoint)
	urlEntityTrashAjax, _ := utils.ToJSON(crud.UrlEntityTrashAjax())
	urlEntityUpdateAjax, _ := utils.ToJSON(crud.UrlEntityUpdateAjax())

	inlineScript := `
	var entityManagerUrl = ` + urlHome + `;
	var entityUpdateUrl = ` + urlEntityUpdateAjax + `;
	var entityTrashUrl = ` + urlEntityTrashAjax + `;
	var entityId = "` + entityID + `";
	var customValues = ` + jsonCustomValues + `;
	const EntityUpdate = {
		data() {
			return {
				entityModel:{
					entityId,
					...customValues
			    },
				trumbowigConfig: {
					btns: [
						['undo', 'redo'], 
						['formatting'], 
						['strong', 'em', 'del', 'superscript', 'subscript'], 
						['link','justifyLeft','justifyRight','justifyCenter','justifyFull'], 
						['unorderedList', 'orderedList'], 
						['horizontalRule'], 
						['removeformat'], 
						['fullscreen']
					],	
					autogrow: true,
					removeformatPasted: true,
					tagsToRemove: ['script', 'link', 'embed', 'iframe', 'input'],
					tagsToKeep: ['hr', 'img', 'i'],
					autogrowOnEnter: true,
					linkTargets: ['_blank'],
				},
			}
		},
		methods: {
			entitySave(redirect){
				var entityId = this.entityModel.entityId;
				var data = JSON.parse(JSON.stringify(this.entityModel));
				data["entity_id"] = data["entityId"];
				delete data["entityId"];

				$.post(entityUpdateUrl, data).done((response)=>{
					if (response.status !== "success") {
						return Swal.fire({icon: 'error', title: 'Oops...', text: response.message});
					}

					if (redirect===true) {
						setTimeout(()=>{
							window.location.href=entityManagerUrl;
						}, 3000)
					}

					return Swal.fire({icon: 'success',title: 'Entity saved'});
				}).fail((result)=>{
					console.log(result);
					return Swal.fire({icon: 'error', title: 'Oops...', text: result});
				});
			}
		}
	};
	Vue.createApp(EntityUpdate).use(ElementPlus).component('Trumbowyg', VueTrumbowyg.default).mount('#entity-update')
		`

	// webpage := crud.webpage("Edit "+crud.entityNameSingular, h)
	// webpage.AddScript(inlineScript)

	title := "Edit " + crud.entityNameSingular
	html := crud._layout(w, r, title, content, []string{
		"https://cdnjs.cloudflare.com/ajax/libs/datatables/1.10.21/css/jquery.dataTables.css",
		// "https://cdnjs.cloudflare.com/ajax/libs/Trumbowyg/2.19.1/ui/trumbowyg.min.css",
		cdn.TrumbowygCss_2_27_3(),
	}, "", []string{
		"https://cdnjs.cloudflare.com/ajax/libs/datatables/1.10.21/js/jquery.dataTables.js",
		// "https://cdnjs.cloudflare.com/ajax/libs/Trumbowyg/2.19.1/trumbowyg.min.js",
		// "https://unpkg.com/vue@3/dist/vue.global.js",
		cdn.TrumbowygJs_2_27_3(),
		"https://cdn.jsdelivr.net/npm/vue-trumbowyg@4",
		"https://cdn.jsdelivr.net/npm/element-plus",
	}, inlineScript)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (crud Crud) pageEntityUpdateAjax(w http.ResponseWriter, r *http.Request) {
	entityID := strings.Trim(utils.Req(r, "entity_id", ""), " ")

	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
		return
	}

	names := crud._listUpdateNames()
	posts := map[string]string{}
	for _, name := range names {
		posts[name] = utils.Req(r, name, "")
	}

	err := crud.funcUpdate(entityID, posts)

	if err != nil {
		api.Respond(w, r, api.Error("Save failed: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("Saved successfully", map[string]interface{}{"entity_id": entityID}))
}

func (crud Crud) pageEntityTrashAjax(w http.ResponseWriter, r *http.Request) {
	entityID := strings.Trim(utils.Req(r, "entity_id", ""), " ")

	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
		return
	}

	err := crud.funcTrash(entityID)

	if err != nil {
		api.Respond(w, r, api.Error("Entity failed to be trashed: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("Entity trashed successfully", map[string]interface{}{"entity_id": entityID}))
}

func (crud Crud) pageEntitiesEntityTrashModal() *hb.Tag {
	modal := hb.NewDiv().ID("ModalEntityTrash").Class("modal fade")
	modalDialog := hb.NewDiv().Attr("class", "modal-dialog")
	modalContent := hb.NewDiv().Attr("class", "modal-content")
	modalHeader := hb.NewDiv().Attr("class", "modal-header").AddChild(hb.NewHeading5().HTML("Trash Entity"))
	modalBody := hb.NewDiv().Attr("class", "modal-body")
	modalBody.AddChild(hb.NewParagraph().HTML("Are you sure you want to move this entity to trash bin?"))
	modalFooter := hb.NewDiv().Attr("class", "modal-footer")
	modalFooter.AddChild(hb.NewButton().HTML("Close").Attr("class", "btn btn-secondary").Attr("data-bs-dismiss", "modal"))
	modalFooter.AddChild(hb.NewButton().HTML("Move to trash bin").Attr("class", "btn btn-danger").Attr("v-on:click", "entityTrash"))
	modalContent.AddChild(modalHeader).AddChild(modalBody).AddChild(modalFooter)
	modalDialog.AddChild(modalContent)
	modal.AddChild(modalDialog)
	return modal
}

func (crud Crud) pageEntitiesEntityCreateModal() *hb.Tag {
	fields := []*hb.Tag{}
	for _, field := range crud.createFields {
		attrName := field.Name
		attrFormControlLabel := field.Label
		if attrFormControlLabel == "" {
			attrFormControlLabel = attrName
		}
		formGroupAttr := hb.NewDiv().Attr("class", "form-group mt-3")
		formGroupAttrLabel := hb.NewLabel().HTML(attrFormControlLabel).Attr("class", "form-label")
		formGroupAttrInput := hb.NewInput().Attr("class", "form-control").Attr("v-model", "entityCreateModel."+attrName)
		if field.Type == "textarea" {
			formGroupAttrInput = hb.NewTextArea().Attr("class", "form-control").Attr("v-model", "entityCreateModel."+attrName)
		}
		formGroupAttr.AddChild(formGroupAttrLabel)
		formGroupAttr.AddChild(formGroupAttrInput)

		// Add help
		if field.Help != "" {
			formGroupAttrHelp := hb.NewParagraph().Attr("class", "text-info").HTML(field.Help)
			formGroupAttr.AddChild(formGroupAttrHelp)
		}

		fields = append(fields, formGroupAttr)
	}

	modalHeader := hb.NewDiv().Class("modal-header").
		AddChild(hb.NewHeading5().HTML("New " + crud.entityNameSingular))

	modalBody := hb.NewDiv().Class("modal-body").AddChildren(fields)

	modalFooter := hb.NewDiv().Class("modal-footer").
		AddChild(hb.NewButton().HTML("Close").Class("btn btn-prsecondary").Attr("data-bs-dismiss", "modal")).
		AddChild(hb.NewButton().HTML("Create & Continue").Class("btn btn-primary").Attr("v-on:click", "entityCreate"))

	modal := hb.NewDiv().ID("ModalEntityCreate").Class("modal fade").AddChildren([]*hb.Tag{
		hb.NewDiv().Class("modal-dialog").AddChildren([]*hb.Tag{
			hb.NewDiv().Class("modal-content").AddChildren([]*hb.Tag{
				modalHeader,
				modalBody,
				modalFooter,
			}),
		}),
	})

	return modal
}

func (crud Crud) urlHome() string {
	url := crud.homeURL
	return url
}

func (crud Crud) UrlEntityManager() string {
	url := crud.endpoint + "?path=" + pathEntityManager
	return url
}

func (crud Crud) UrlEntityCreateAjax() string {
	url := crud.endpoint + "?path=" + pathEntityCreateAjax
	return url
}

func (crud Crud) UrlEntityTrashAjax() string {
	url := crud.endpoint + "?path=" + pathEntityTrashAjax
	return url
}

func (crud Crud) UrlEntityUpdate() string {
	url := crud.endpoint + "?path=" + pathEntityUpdate
	return url
}

func (crud Crud) UrlEntityUpdateAjax() string {
	url := crud.endpoint + "?path=" + pathEntityUpdateAjax
	return url
}

// Webpage returns the webpage template for the website
func (crud Crud) webpage(title, content string) *hb.Webpage {
	faviconImgCms := `data:image/x-icon;base64,AAABAAEAEBAQAAEABAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAmzKzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABEQEAAQERAAEAAQABAAEAAQABAQEBEQABAAEREQEAAAERARARAREAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAAi6MAALu7AAC6owAAuC8AAIkjAAD//wAA//8AAP//AAD//wAA`
	app := ""
	webpage := hb.NewWebpage()
	webpage.SetTitle(title)
	webpage.SetFavicon(faviconImgCms)

	webpage.AddStyleURLs([]string{
		"https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css",
	})
	webpage.AddScriptURLs([]string{
		"https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/js/bootstrap.bundle.min.js",
		"https://code.jquery.com/jquery-3.6.0.min.js",
		"https://unpkg.com/vue@3/dist/vue.global.js",
		"https://cdn.jsdelivr.net/npm/sweetalert2@9",
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
	webpage.AddChild(hb.NewHTML(content))
	return webpage
}

func (crud Crud) _breadcrumbs(breadcrumbs []Breadcrumb) string {
	nav := hb.NewNav().Attr("aria-label", "breadcrumb")
	ol := hb.NewOL().Attr("class", "breadcrumb")
	for _, breadcrumb := range breadcrumbs {
		li := hb.NewLI().Attr("class", "breadcrumb-item")
		link := hb.NewHyperlink().HTML(breadcrumb.Name).Attr("href", breadcrumb.URL)

		li.AddChild(link)

		ol.AddChild(li)
	}
	nav.AddChild(ol)
	return nav.ToHTML()
}

func (crud Crud) _layout(w http.ResponseWriter, r *http.Request, title string, content string, styleFiles []string, style string, jsFiles []string, js string) string {
	html := ""

	if crud.funcLayout != nil {
		// jsFiles = append([]string{"//unpkg.com/naive-ui"}, jsFiles...)
		jsFiles = append([]string{"//cdn.jsdelivr.net/npm/element-plus"}, jsFiles...)
		jsFiles = append([]string{"https://unpkg.com/vue@3/dist/vue.global.js"}, jsFiles...)
		jsFiles = append([]string{"https://cdn.jsdelivr.net/npm/sweetalert2@9"}, jsFiles...)
		styleFiles = append([]string{"https://unpkg.com/element-plus/dist/index.css"}, styleFiles...)
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

func (crud Crud) _form([]FormField) []*hb.Tag {
	tags := []*hb.Tag{}
	for _, field := range crud.updateFields {
		fieldName := field.Name
		fieldLabel := field.Label
		if fieldLabel == "" {
			fieldLabel = fieldName
		}
		formGroup := hb.NewDiv().Class("form-group mt-3")
		formGroupLabel := hb.NewLabel().HTML(fieldLabel).Attr("class", "form-label")
		formGroupInput := hb.NewInput().Class("form-control").Attr("v-model", "entityModel."+fieldName)
		if field.Type == "image" {
			formGroupInput = hb.NewDiv().Children([]*hb.Tag{
				hb.NewImage().
					Attr(`v-bind:src`, `entityModel.`+fieldName+`||'https://www.freeiconspng.com/uploads/no-image-icon-11.PNG'`).
					Style(`width:200px;`),
				bs.InputGroup().Children([]*hb.Tag{
					hb.NewInput().Type(hb.TYPE_URL).Class("form-control").Attr("v-model", "entityModel."+fieldName),
					hb.If(crud.fileManagerURL != "", bs.InputGroupText().Children([]*hb.Tag{
						hb.NewHyperlink().HTML("Browse").Href(crud.fileManagerURL).Target("_blank"),
					})),
				}),
			})
		}
		if field.Type == "datetime" {
			// formGroupInput = hb.NewInput().Type(hb.TYPE_DATETIME).Class("form-control").Attr("v-model", "entityModel."+fieldName)
			formGroupInput = hb.NewTag(`el-date-picker`).Attr("type", "datetime").Attr("v-model", "entityModel."+fieldName)
			// formGroupInput = hb.NewTag(`n-date-picker`).Attr("type", "datetime").Class("form-control").Attr("v-model", "entityModel."+fieldName)
		}
		if field.Type == "textarea" {
			formGroupInput = hb.NewTextArea().Class("form-control").Attr("v-model", "entityModel."+fieldName)
		}
		if field.Type == "htmlarea" {
			formGroupInput = hb.NewTag("trumbowyg").Attr("v-model", "entityModel."+fieldName).Attr(":config", "trumbowigConfig").Class("form-control")
		}
		if field.Type == "select" {
			formGroupInput = hb.NewSelect().Class("form-select").Attr("v-model", "entityModel."+fieldName)
			for _, opt := range field.Options {
				option := hb.NewOption().Value(opt.Key).HTML(opt.Value)
				formGroupInput.AddChild(option)
			}
			if field.OptionsF != nil {
				for _, opt := range field.OptionsF() {
					option := hb.NewOption().Value(opt.Key).HTML(opt.Value)
					formGroupInput.AddChild(option)
				}
			}
		}
		formGroup.AddChild(formGroupLabel)
		formGroup.AddChild(formGroupInput)

		// Add help
		if field.Help != "" {
			formGroupHelp := hb.NewParagraph().Class("text-info").HTML(field.Help)
			formGroup.AddChild(formGroupHelp)
		}

		tags = append(tags, formGroup)
	}

	return tags
}

func (crud Crud) _listCreateNames() []string {
	names := []string{}
	for _, field := range crud.createFields {
		if field.Name == "" {
			continue
		}
		names = append(names, field.Name)
	}
	return names
}

func (crud Crud) _listUpdateNames() []string {
	names := []string{}
	for _, field := range crud.updateFields {
		if field.Name == "" {
			continue
		}
		names = append(names, field.Name)
	}
	return names
}
