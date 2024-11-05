package crud

import (
	"net/http"

	"github.com/gouniverse/bs"
	"github.com/gouniverse/form"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/utils"
	"github.com/samber/lo"
)

type entityCreateController struct {
	crud *Crud
}

func (crud *Crud) newEntityCreateController() *entityCreateController {
	return &entityCreateController{
		crud: crud,
	}
}

func (controller *entityCreateController) modalShow(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(controller.modal().ToHTML()))
}

func (controller *entityCreateController) modalSave(w http.ResponseWriter, r *http.Request) {
	names := controller.crud.listCreateNames()

	posts := map[string]string{}
	for _, name := range names {
		posts[name] = utils.Req(r, name, "")
	}

	// Check required fields
	for _, field := range controller.crud.createFields {
		if !field.GetRequired() {
			continue
		}

		if _, exists := posts[field.GetName()]; !exists {
			errorMessage := field.GetLabel() + " is required field"
			response := hb.Swal(hb.SwalOptions{Icon: "error", Text: errorMessage}).ToHTML()
			w.Write([]byte(response))
			return
		}

		if lo.IsEmpty(posts[field.GetName()]) {
			errorMessage := field.GetLabel() + " is required field"
			response := hb.Swal(hb.SwalOptions{Icon: "error", Text: errorMessage}).ToHTML()
			w.Write([]byte(response))
			return
		}
	}

	entityID, err := controller.crud.funcCreate(posts)

	if err != nil {
		errorMessage := "Save failed: " + err.Error()
		response := hb.Swal(hb.SwalOptions{Icon: "error", Text: errorMessage}).ToHTML()
		w.Write([]byte(response))
		return
	}

	redirectURL := controller.crud.UrlEntityUpdate() + "?entity_id=" + entityID
	successMessage := "Saved successfully"
	response := hb.Wrap().
		Child(hb.Swal(hb.SwalOptions{
			Icon: "success",
			Text: successMessage,
		})).
		Child(hb.Script("setTimeout(() => {window.location.href = '" + redirectURL + "'}, 2000)")).
		ToHTML()

	w.Write([]byte(response))
}

func (controller *entityCreateController) modal() hb.TagInterface {
	form := form.NewForm(form.FormOptions{
		Fields: controller.crud.createFields,
	}).Build()

	//controller.crud.form(controller.crud.createFields)

	submitUrl := controller.crud.UrlEntityCreateAjax()

	modalID := "ModalEntityCreate"
	modalBackdropClass := "ModalBackdrop"

	modalCloseScript := `closeModal` + modalID + `();`

	modalHeading := hb.Heading5().
		Text("New " + controller.crud.entityNameSingular).
		Style(`margin:0px;`)

	modalClose := hb.Button().Type("button").
		Class("btn-close").
		Data("bs-dismiss", "modal").
		OnClick(modalCloseScript)

	jsCloseFn := `function closeModal` + modalID + `() {document.getElementById('ModalEntityCreate').remove();[...document.getElementsByClassName('` + modalBackdropClass + `')].forEach(el => el.remove());}`

	buttonSubmit := hb.Button().
		Child(hb.I().Class("bi bi-check me-2")).
		HTML("Create & Edit").
		Class("btn btn-primary float-end").
		HxInclude("#" + modalID).
		HxPost(submitUrl).
		HxSelectOob("#ModalproductCreate").
		HxTarget("body").
		HxSwap("beforeend")

	buttonCancel := hb.Button().
		Child(hb.I().Class("bi bi-chevron-left me-2")).
		HTML("Close").
		Class("btn btn-secondary float-start").
		Data("bs-dismiss", "modal").
		OnClick(modalCloseScript)

	modal := bs.Modal().
		ID(modalID).
		Class("fade show").
		Style(`display:block;position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:1051;`).
		Child(hb.Script(jsCloseFn)).
		Child(bs.ModalDialog().
			Child(bs.ModalContent().
				Child(
					bs.ModalHeader().
						Child(modalHeading).
						Child(modalClose)).
				Child(
					bs.ModalBody().
						Child(form)).
				Child(bs.ModalFooter().
					Style(`display:flex;justify-content:space-between;`).
					Child(buttonCancel).
					Child(buttonSubmit)),
			))

	backdrop := hb.Div().Class(modalBackdropClass).
		Class("modal-backdrop fade show").
		Style("display:block;z-index:1000;")

	return hb.Wrap().Children([]hb.TagInterface{
		modal,
		backdrop,
	})
}

// func (controller *entityCreateController) prepareDataAndValidate(r *http.Request) (data entityCreateControllerData, errorMessage string) {
// 	// authUser := helpers.GetAuthUser(r)

// 	// if authUser == nil {
// 	// 	return data, "You are not logged in. Please login to continue."
// 	// }

// 	// data.formTitle = strings.TrimSpace(utils.Req(r, "product_title", ""))

// 	// if r.Method != http.MethodPost {
// 	// 	return data, ""
// 	// }

// 	// if data.formTitle == "" {
// 	// 	return data, "product title is required"
// 	// }

// 	// product := shopstore.NewProduct()
// 	// product.SetTitle(data.formTitle)

// 	// err := config.ShopStore.ProductCreate(product)

// 	// if err != nil {
// 	// 	config.LogStore.ErrorWithContext("Error. At productCreateController > prepareDataAndValidate", err.Error())
// 	// 	return data, "Creating product failed. Please contact an administrator."
// 	// }

// 	data.successMessage = "product created successfully."

// 	return data, ""

// }
