package core

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
)

// ModelGroup manages a collection of related models sharing a database connection.
// It handles initialization, table creation, and transaction management.
type ModelGroup struct {
	models          []IModel
	db              *db.DB
	name            string
	autoCreateTable bool
}

// AddModel adds one or more models to this group.
func (m *ModelGroup) AddModel(model ...IModel) {
	m.models = append(m.models, model...)
}
// GetModel returns all models in this group.
func (m *ModelGroup) GetModel() []IModel {
	return m.models
}

// AutoCreateTable sets whether tables should be auto-created during initialization.
func (m *ModelGroup) AutoCreateTable(autoCreateTable bool) {
	m.autoCreateTable = autoCreateTable
}
// GetTransaction returns a new transaction manager for this group's database.
// Panics if no database is configured.
func (m *ModelGroup) GetTransaction() *model.Transaction {
	if m.db == nil {
		log.Panic("db is nil", zap.String("name", m.name))
	}
	return model.NewTransaction(m.db)
}

// SetDefaultDB sets the default database connection for this model group.
func (m *ModelGroup) SetDefaultDB(db *db.DB) {
	log.Info("set db", zap.String("name", m.name))
	m.db = db
}

// SwitchDB replaces the database connection and reinitializes all models.
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

// Name returns the name of this model group.
func (m *ModelGroup) Name() string {
	return m.name
}
// Init initializes all models in the group, optionally creating tables.
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

// ModelDefaultName is the default name for the primary model group.
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
// EmptyModelGroup creates a model group with no database, useful for testing or placeholder purposes.
func EmptyModelGroup(name string) *ModelGroup {
	return &ModelGroup{
		db:     nil,
		name:   name,
		models: make([]IModel, 0),
	}
}

// ModelGroupBuilder provides a fluent API for constructing ModelGroup configurations.
type ModelGroupBuilder struct {
	db   *db.DB
	name string
	models []IModel
	autoCreateTable bool
}

// NewModelGroupBuilder creates a new ModelGroupBuilder for fluent model group construction.
func NewModelGroupBuilder() *ModelGroupBuilder {
	return &ModelGroupBuilder{
		db:   nil,
		name: ModelDefaultName,
		models: make([]IModel, 0),
	}
}
// DB sets the database connection for the model group.
func (m *ModelGroupBuilder) DB(db *db.DB) *ModelGroupBuilder {
	m.db = db
	return m
}

// Name sets the name of the model group.
func (m *ModelGroupBuilder) Name(name string) *ModelGroupBuilder {
	m.name = name
	return m
}

// Model adds one or more models to the group.
func (m *ModelGroupBuilder) Model(model ...IModel) *ModelGroupBuilder {
	m.models = append(m.models, model...)
	return m
}
// AutoCreateTable enables or disables automatic table creation during initialization.
func (m *ModelGroupBuilder) AutoCreateTable(auto bool) *ModelGroupBuilder {
	m.autoCreateTable = auto
	return m
}

// Build creates a ModelGroup from the builder configuration.
func (m *ModelGroupBuilder) Build() *ModelGroup {
	group := newModelGroup(m.db, m.name)
	group.AddModel(m.models...)
	group.AutoCreateTable(m.autoCreateTable)
	return group
}
