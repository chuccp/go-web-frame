package core

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
)

type ModelGroup struct {
	models          []IModel
	db              *db.DB
	name            string
	autoCreateTable bool
}

func (m *ModelGroup) AddModel(model ...IModel) {
	m.models = append(m.models, model...)
}
func (m *ModelGroup) GetModel() []IModel {
	return m.models
}

func (m *ModelGroup) AutoCreateTable(autoCreateTable bool) {
	m.autoCreateTable = autoCreateTable
}
func (m *ModelGroup) GetTransaction() *model.Transaction {
	if m.db == nil {
		log.Panic("db is nil", zap.String("name", m.name))
	}
	return model.NewTransaction(m.db)
}

func (m *ModelGroup) SetDefaultDB(db *db.DB) {
	log.Info("set db", zap.String("name", m.name))
	m.db = db
}

func (m *ModelGroup) SwitchDB(db *db.DB, context *Context) error {
	m.db = db
	for _, iModel := range m.models {
		log.Debug("Init", zap.String("model", util.GetStructFullQualifiedName(iModel)))
		err := iModel.Init(m.db, context)
		if err != nil {
			return errors.WithStackIf(err)
		}
		if m.autoCreateTable {
			err = iModel.CreateTable()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *ModelGroup) Name() string {
	return m.name
}
func (m *ModelGroup) Init(context *Context) error {
	if m.db != nil {
		for _, iModel := range m.models {
			log.Debug("Init", zap.String("model", util.GetStructFullQualifiedName(iModel)))
			err := iModel.Init(m.db, context)
			if err != nil {
				return err
			}
			if m.autoCreateTable {
				err = iModel.CreateTable()
				if err != nil {
					return err
				}
			}
		}
	} else {
		log.Warn("db is nil", zap.String("name", m.name))
	}
	return nil
}

const ModelDefaultName = "ModelDefaultName"

func newModelGroup(db *db.DB, name string) *ModelGroup {
	return &ModelGroup{
		db:     db,
		name:   name,
		models: make([]IModel, 0),
	}
}

//	func DefaultModelGroup() *ModelGroup {
//		return &ModelGroup{
//			db:     nil,
//			name:   DefaultName,
//			models: make([]IModel, 0),
//		}
//	}
func EmptyModelGroup(name string) *ModelGroup {
	return &ModelGroup{
		db:     nil,
		name:   name,
		models: make([]IModel, 0),
	}
}

type ModelGroupBuilder struct {
	db   *db.DB
	name string
	models []IModel
	autoCreateTable bool
}

func NewModelGroupBuilder() *ModelGroupBuilder {
	return &ModelGroupBuilder{
		db:   nil,
		name: ModelDefaultName,
		models: make([]IModel, 0),
	}
}
func (m *ModelGroupBuilder) DB(db *db.DB) *ModelGroupBuilder {
	m.db = db
	return m
}
func (m *ModelGroupBuilder) Name(name string) *ModelGroupBuilder {
	m.name = name
	return m
}
func (m *ModelGroupBuilder) Model(model ...IModel) *ModelGroupBuilder {
	m.models = append(m.models, model...)
	return m
}
func (m *ModelGroupBuilder) AutoCreateTable(auto bool) *ModelGroupBuilder {
	m.autoCreateTable = auto
	return m
}
func (m *ModelGroupBuilder) Build() *ModelGroup {
	group := newModelGroup(m.db, m.name)
	group.AddModel(m.models...)
	group.AutoCreateTable(m.autoCreateTable)
	return group
}
