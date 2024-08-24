package crud

import "errors"

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
	crud.endpoint = config.Endpoint
	crud.entityNamePlural = config.EntityNamePlural
	crud.entityNameSingular = config.EntityNameSingular
	crud.fileManagerURL = config.FileManagerURL
	crud.funcCreate = config.FuncCreate
	crud.funcReadExtras = config.FuncReadExtras
	crud.funcFetchReadData = config.FuncFetchReadData
	crud.funcFetchUpdateData = config.FuncFetchUpdateData
	crud.funcLayout = config.FuncLayout
	crud.funcRows = config.FuncRows
	crud.funcTrash = config.FuncTrash
	crud.funcUpdate = config.FuncUpdate
	crud.homeURL = config.HomeURL
	crud.readFields = config.ReadFields
	crud.updateFields = config.UpdateFields

	return crud, err
}
