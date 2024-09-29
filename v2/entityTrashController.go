package crud

import (
	"net/http"
	"strings"

	"github.com/gouniverse/api"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/utils"
)

type entityTrashController struct {
	crud *Crud
}

func (crud *Crud) newEntityTrashController() *entityTrashController {
	return &entityTrashController{
		crud: crud,
	}
}

func (controller *entityTrashController) pageEntityTrashAjax(w http.ResponseWriter, r *http.Request) {
	entityID := strings.Trim(utils.Req(r, "entity_id", ""), " ")

	if entityID == "" {
		api.Respond(w, r, api.Error("Entity ID is required"))
		return
	}

	err := controller.crud.funcTrash(entityID)

	if err != nil {
		api.Respond(w, r, api.Error("Entity failed to be trashed: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("Entity trashed successfully", map[string]interface{}{"entity_id": entityID}))
}

func (controller *entityTrashController) pageEntitiesEntityTrashModal() hb.TagInterface {
	modal := hb.Div().ID("ModalEntityTrash").Class("modal fade")
	modalDialog := hb.Div().Attr("class", "modal-dialog")
	modalContent := hb.Div().Attr("class", "modal-content")
	modalHeader := hb.Div().Attr("class", "modal-header").AddChild(hb.Heading5().Text("Trash Entity"))
	modalBody := hb.Div().Attr("class", "modal-body")
	modalBody.AddChild(hb.Paragraph().Text("Are you sure you want to move this entity to trash bin?"))
	modalFooter := hb.Div().Attr("class", "modal-footer")
	modalFooter.AddChild(hb.Button().Text("Close").Attr("class", "btn btn-secondary").Attr("data-bs-dismiss", "modal"))
	modalFooter.AddChild(hb.Button().Text("Move to trash bin").Attr("class", "btn btn-danger").Attr("v-on:click", "entityTrash"))
	modalContent.AddChild(modalHeader).AddChild(modalBody).AddChild(modalFooter)
	modalDialog.AddChild(modalContent)
	modal.AddChild(modalDialog)
	return modal
}
