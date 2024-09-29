package crud

import (
	"net/http"
	"strings"

	"github.com/gouniverse/api"
	"github.com/gouniverse/cdn"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/icons"
	"github.com/gouniverse/utils"
	"github.com/samber/lo"
)

type entityUpdateController struct {
	crud *Crud
}

func (crud *Crud) newEntityUpdateController() *entityUpdateController {
	return &entityUpdateController{
		crud: crud,
	}
}

func (controller *entityUpdateController) page(w http.ResponseWriter, r *http.Request) {
	entityID := utils.Req(r, "entity_id", "")
	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
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
			Name: "Edit " + controller.crud.entityNameSingular,
			URL:  controller.crud.UrlEntityUpdate() + "&entity_id=" + entityID,
		},
	})

	buttonSave := hb.Button().Class("btn btn-success float-end").Attr("v-on:click", "entitySave(true)").
		AddChild(icons.Icon("bi-check-all", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("Save")
	buttonApply := hb.Button().Class("btn btn-success float-end").Attr("v-on:click", "entitySave").
		Style("margin-right:10px;").
		AddChild(icons.Icon("bi-check", 16, 16, "white").Style("margin-top:-4px;margin-right:8px;")).
		HTML("Apply")
	heading := hb.Heading1().Text("Edit " + controller.crud.entityNameSingular).
		AddChild(buttonSave).
		AddChild(buttonApply)

	// container.AddChild(hb.HTML(header))
	container := hb.Div().Attr("class", "container").Attr("id", "entity-update").
		AddChild(heading).
		AddChild(hb.Raw(breadcrumbs))

	customAttrValues, errData := controller.crud.funcFetchUpdateData(entityID)

	if errData != nil {
		api.Respond(w, r, api.Error("Fetch data failed"))
		return
	}

	container.AddChildren(controller.crud.form(controller.crud.updateFields))

	content := container.ToHTML()

	jsonCustomValues, _ := utils.ToJSON(customAttrValues)

	urlHome, _ := utils.ToJSON(controller.crud.endpoint)
	urlEntityTrashAjax, _ := utils.ToJSON(controller.crud.UrlEntityTrashAjax())
	urlEntityUpdateAjax, _ := utils.ToJSON(controller.crud.UrlEntityUpdateAjax())

	inlineScript := `
	const entityManagerUrl = ` + urlHome + `;
	const entityUpdateUrl = ` + urlEntityUpdateAjax + `;
	const entityTrashUrl = ` + urlEntityTrashAjax + `;
	const entityId = "` + entityID + `";
	const customValues = ` + jsonCustomValues + `;
	const EntityUpdate = {
		data() {
			return {
				entityModel:{
					entityId,
					...customValues
			    },
				tmp:{},
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
				const entityId = this.entityModel.entityId;
				let data = JSON.parse(JSON.stringify(this.entityModel));
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
			},
			uploadImage(event, fieldName) {
				const self = this;
				if ( event.target.files && event.target.files[0] ) {
					var FR= new FileReader();
					FR.onload = function(e) {
						self.entityModel[fieldName] = e.target.result;
						event.target.value = "";
					};       
					FR.readAsDataURL( event.target.files[0] );
				}
			}
		}
	};
	Vue.createApp(EntityUpdate).use(ElementPlus).component('Trumbowyg', VueTrumbowyg.default).mount('#entity-update')
		`

	// webpage := crud.webpage("Edit "+crud.entityNameSingular, h)
	// webpage.AddScript(inlineScript)

	title := "Edit " + controller.crud.entityNameSingular
	html := controller.crud.layout(w, r, title, content, []string{
		cdn.JqueryDataTablesCss_1_13_4(),
		cdn.TrumbowygCss_2_27_3(),
	}, "", []string{
		cdn.JqueryDataTablesJs_1_13_4(),
		cdn.TrumbowygJs_2_27_3(),
		"https://cdn.jsdelivr.net/npm/vue-trumbowyg@4",
		"https://cdn.jsdelivr.net/npm/element-plus",
	}, inlineScript)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (controller *entityUpdateController) pageSave(w http.ResponseWriter, r *http.Request) {
	entityID := strings.Trim(utils.Req(r, "entity_id", ""), " ")

	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
		return
	}

	names := controller.crud.listUpdateNames()
	posts := map[string]string{}
	for _, name := range names {
		posts[name] = utils.Req(r, name, "")
	}

	// Check required fields
	for _, field := range controller.crud.updateFields {
		if !field.Required {
			continue
		}

		if _, exists := posts[field.Name]; !exists {
			api.Respond(w, r, api.Error(field.Label+" is required field"))
			return
		}

		if lo.IsEmpty(posts[field.Name]) {
			api.Respond(w, r, api.Error(field.Label+" is required field"))
			return
		}
	}

	err := controller.crud.funcUpdate(entityID, posts)

	if err != nil {
		api.Respond(w, r, api.Error("Save failed: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("Saved successfully", map[string]interface{}{"entity_id": entityID}))
}
