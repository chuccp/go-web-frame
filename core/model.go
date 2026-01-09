package core

import "github.com/chuccp/go-web-frame/db"

type ModelGroup struct {
	models []IModel
	db     *db.DB
}

func (m *ModelGroup) AddModel(model ...IModel) {
	m.models = append(m.models, model...)
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

func NewModelGroup(db *db.DB) *ModelGroup {
	return &ModelGroup{
		db: db,
	}
}
func EmptyModelGroup() *ModelGroup {
	return &ModelGroup{
		db: nil,
	}
}
