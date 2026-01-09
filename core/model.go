package core

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"go.uber.org/zap"
)

type ModelGroup struct {
	models []IModel
	db     *db.DB
	name   string
}

func (m *ModelGroup) AddModel(model ...IModel) {
	m.models = append(m.models, model...)
}
func (m *ModelGroup) GetModel() []IModel {
	return m.models
}
func (m *ModelGroup) GetTransaction() *model.Transaction {
	if m.db == nil {
		log.Panic("db is nil", zap.String("name", m.name))
	}
	return model.NewTransaction(m.db)
}
func (m *ModelGroup) SwitchDB(db *db.DB, context *Context) error {
	m.db = db
	for _, model := range m.models {
		err := model.Init(m.db, context)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	return nil
}

func (m *ModelGroup) Name() string {
	return m.name
}
func (m *ModelGroup) Init(context *Context) error {
	if m.db != nil {
		for _, model := range m.models {
			err := model.Init(m.db, context)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const DefaultName = "DefaultName"

func NewModelGroup(db *db.DB, name string) *ModelGroup {
	return &ModelGroup{
		db:     db,
		name:   name,
		models: make([]IModel, 0),
	}
}
func DefaultModelGroup() *ModelGroup {
	return &ModelGroup{
		db:     nil,
		name:   DefaultName,
		models: make([]IModel, 0),
	}
}
func EmptyModelGroup(name string) *ModelGroup {
	return &ModelGroup{
		db:     nil,
		name:   name,
		models: make([]IModel, 0),
	}
}
