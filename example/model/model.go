package model

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

type User struct {
	Id   uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"size:255;not null" json:"name"`
	Info Info   `gorm:"not null" json:"info"`
}
type Info struct {
	Id    uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Color string `gorm:"size:255;not null" json:"color"`
}
type UserModel struct {
	*model.Model[*User]
	db *db.DB
	c  *core.Context
}

func (userModel *UserModel) Init(db *db.DB, c *core.Context) error {
	userModel.db = db
	userModel.c = c
	userModel.Model = model.NewModel[*User](db, "t_User")
	return nil
}

type InfoModel struct {
	*model.Model[*Info]
	db *db.DB
	c  *core.Context
}

func (infoModel *InfoModel) Init(db *db.DB, c *core.Context) error {
	infoModel.db = db
	infoModel.c = c
	infoModel.Model = model.NewModel[*Info](db, "t_Info")
	return nil
}

func main() {
	//&UserModel{}.Query()

}
