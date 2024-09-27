package crud

import (
	"net/http"
	"strings"

	"github.com/gouniverse/cdn"
	"github.com/gouniverse/form"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/icons"
	"github.com/gouniverse/utils"
	"github.com/samber/lo"
)

type entityManagerController struct {
	crud *Crud
}

func (crud *Crud) newEntityManagerController() *entityManagerController {
	return &entityManagerController{
		crud: crud,
	}
}

func (controller *entityManagerController) page(w http.ResponseWriter, r *http.Request) {
	// header := cms.cmsHeader(endpoint)
	breadcrumbs := controller.crud._breadcrumbs([]Breadcrumb{
		{
			Name: "Home",
			URL:  controller.crud.urlHome(),
		},
		{
			Name: controller.crud.entityNameSingular + " Manager",
			URL:  controller.crud.UrlEntityManager(),
		},
	})

	// buttonCreate := hb.NewButton().
	// 	Class("btn btn-success float-end").
	// 	Attr("v-on:click", "showEntityCreateModal").
	// 	AddChild(icons.Icon("bi-plus-circle", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
	// 	HTML("New " + crud.entityNameSingular)

	buttonCreate := hb.NewButton().
		Class("btn btn-success float-end").
		// Attr("v-on:click", "showEntityCreateModal").
		AddChild(icons.Icon("bi-plus-circle", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("New " + controller.crud.entityNameSingular).
		HxGet(controller.crud.UrlEntityCreateModal()).
		HxTarget("body").
		HxSwap("beforeend")

	heading := hb.NewHeading1().
		HTML(controller.crud.entityNameSingular + " Manager").
		Child(buttonCreate)

	rows, errRows := controller.crud.funcRows()

	tableContent := lo.IfF(errRows != nil, func() hb.TagInterface {
		alert := hb.NewDiv().
			Class("alert alert-danger").
			HTML("There was an error retrieving the data. Please try again later")

		return alert
	}).ElseF(func() hb.TagInterface {
		table := hb.NewTable().
			ID("TableEntities").
			Class("table table-responsive table-striped mt-3").
			Child(
				hb.NewThead().
					Children([]hb.TagInterface{
						hb.NewTR().
							Children(lo.Map(controller.crud.columnNames, func(columnName string, _ int) hb.TagInterface {
								columnName = strings.ReplaceAll(columnName, "{!!", "")
								columnName = strings.ReplaceAll(columnName, "!!}", "")
								return hb.NewTH().Text(columnName)
							})).
							Child(hb.NewTD().
								HTML("Actions").
								Style("width:120px;")),
					})).
			Child(
				hb.NewTbody().
					Children(lo.Map(rows, func(row Row, _ int) hb.TagInterface {
						buttonView := hb.NewHyperlink().
							Class("btn btn-sm btn-outline-info").
							Child(icons.Icon("bi-eye", 18, 18, "#333").
								Style("margin-top:-4px;")).
							Attr("title", "Show").
							Href(controller.crud.UrlEntityRead() + "&entity_id=" + row.ID).
							Style("margin-right:5px")

						buttonEdit := hb.NewHyperlink().
							Class("btn btn-sm btn-outline-warning").
							Child(icons.Icon("bi-pencil-square", 18, 18, "#333").
								Style("margin-top:-4px;")).
							Attr("title", "Edit").
							Attr("type", "button").
							Href(controller.crud.UrlEntityUpdate() + "&entity_id=" + row.ID).
							Style("margin-right:5px")

						buttonTrash := hb.NewButton().
							Class("btn btn-sm btn-outline-danger").
							Child(icons.Icon("bi-trash", 18, 18, "#333").
								Style("margin-top:-4px;")).
							Attr("title", "Trash").
							Attr("type", "button").
							Attr("v-on:click", "showEntityTrashModal('"+row.ID+"')")

						tr := hb.NewTR().
							Children(lo.Map(row.Data, func(cell string, index int) hb.TagInterface {
								name := controller.crud.columnNames[index]
								isRaw := strings.HasPrefix(name, "{!!") && strings.HasSuffix(name, "!!}")
								cell = strings.ReplaceAll(cell, "{!!", "")
								cell = strings.ReplaceAll(cell, "!!}", "")
								cell = strings.TrimSpace(cell)
								return hb.NewTD().TextIf(!isRaw, cell).HTMLIf(isRaw, cell)
							})).
							Child(
								hb.NewTD().
									Style(`white-space:nowrap;`).
									ChildIf(controller.crud.funcFetchReadData != nil, buttonView).
									ChildIf(controller.crud.funcFetchUpdateData != nil, buttonEdit).
									ChildIf(controller.crud.funcTrash != nil, buttonTrash),
							)
						return tr
					})))

		return table
	})

	container := hb.NewDiv().
		ID("entity-manager").
		Class("container").
		Child(heading).
		Child(hb.NewHTML(breadcrumbs)).
		// Child(crud.pageEntitiesEntityCreateModal()).
		Child(controller.crud.newEntityTrashController().pageEntitiesEntityTrashModal()).
		Child(tableContent)

	content := container.ToHTML()

	urlEntityCreateAjax, _ := utils.ToJSON(controller.crud.UrlEntityCreateAjax())
	urlEntityTrashAjax, _ := utils.ToJSON(controller.crud.UrlEntityTrashAjax())
	urlEntityUpdate, _ := utils.ToJSON(controller.crud.UrlEntityUpdate())

	customAttrValues := map[string]string{}
	lo.ForEach(controller.crud.createFields, func(field form.Field, index int) {
		customAttrValues[field.Name] = field.Value
	})
	jsonCustomValues, _ := utils.ToJSON(customAttrValues)

	inlineScript := `
const entityCreateUrl = ` + urlEntityCreateAjax + `;
const entityUpdateUrl = ` + urlEntityUpdate + `;
const entityTrashUrl = ` + urlEntityTrashAjax + `;
const customValues = ` + jsonCustomValues + `;
const EntityManager = {
	data() {
		return {
		  entityModel:{
			...customValues
		  },
		  entityTrashModel:{
			entityId:null,
		  }
		}
	},
	created(){
		//setTimeout(() => {
		//	console.log("Init data table...");
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
			const modalEntityCreate = new bootstrap.Modal(document.getElementById('ModalEntityCreate'));
			modalEntityCreate.show();
		},
		showEntityTrashModal(entityId){
			this.entityTrashModel.entityId = entityId;
			const modalEntityDelete = new bootstrap.Modal(document.getElementById('ModalEntityTrash'));
			modalEntityDelete.show();
		},
		entityCreate(){
		    $.post(entityCreateUrl, this.entityModel).done((result)=>{
				if (result.status==="success"){
					const modalEntityCreate = new bootstrap.Modal(document.getElementById('ModalEntityCreate'));
			        modalEntityCreate.hide();
					return location.href = entityUpdateUrl+ "&entity_id=" + result.data.entity_id;
				}
				
				return Swal.fire({icon: 'error', title: 'Oops...', text: result.message});
			}).fail((result)=>{
				return Swal.fire({icon: 'error', title: 'Oops...', text: result});
			});
		},

		entityTrash(){
			const entityId = this.entityTrashModel.entityId;

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
	title := controller.crud.entityNameSingular + " Manager"
	html := controller.crud.layout(w, r, title, content, []string{
		cdn.JqueryDataTablesCss_1_13_4(),
	}, "html{width:100%;}", []string{
		cdn.JqueryDataTablesJs_1_13_4(),
	}, inlineScript)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
