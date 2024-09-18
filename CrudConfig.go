package crud

import (
	"net/http"

	"github.com/gouniverse/hb"
)

type CrudConfig struct {
	ColumnNames         []string
	CreateFields        []FormField
	Endpoint            string
	EntityNamePlural    string
	EntityNameSingular  string
	FileManagerURL      string
	FuncCreate          func(data map[string]string) (userID string, err error)
	FuncFetchReadData   func(entityID string) ([][2]string, error)
	FuncFetchUpdateData func(entityID string) (map[string]string, error)
	FuncLayout          func(w http.ResponseWriter, r *http.Request, title string, content string, styleFiles []string, style string, jsFiles []string, js string) string
	FuncRows            func() (rows []Row, err error)
	FuncTrash           func(entityID string) error
	FuncUpdate          func(entityID string, data map[string]string) error
	HomeURL             string
	ReadFields          []FormField
	UpdateFields        []FormField
	FuncReadExtras      func(entityID string) []hb.TagInterface
}