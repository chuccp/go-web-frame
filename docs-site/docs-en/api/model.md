# Model API

Generic ORM API.

## Model[T]

Base generic model.

```go
type UserModel struct {
    *model.Model[*User]
}

// Initialize
m.Model = model.NewModel[*User](db, "t_user")
```

## Query Operations

```go
// All records
users, err := m.Query().All()

// Single record
user, err := m.Query().Where("id = ?", 1).One()

// Count
count, err := m.Query().Count()

// With conditions
users, err := m.Query().Where("status = ?", 1).Order("id desc").All()

// Pagination
users, total, err := m.Query().Page(page)

// Limit
users, err := m.Query().Limit(10).All()

// Preload (GORM)
users, err := m.Query().Preload("Profile").All()
```

## Update Operations

```go
err := m.Update().Where("id = ?", 1).UpdateColumn("name", "bob")

err := m.Update().Where("id = ?", 1).UpdateForMap(map[string]any{
    "name": "bob",
    "status": 1,
})
```

## Delete Operations

```go
err := m.Delete().Where("id = ?", 1).Delete()
```

## Save

```go
user := &User{Name: "alice"}
err := m.Save(user)
// user.Id is populated after Save
```

## EntryModel[T]

Extended model for entities with Id, CreateTime, UpdateTime.

```go
type UserModel struct {
    *model.EntryModel[*User]
}

// Additional methods
user, err := m.FindById(1)
users, err := m.FindAll()
err := m.DeleteById(1)
err := m.UpdateById(user)
pageAble, err := m.Query().PageForWeb(page)
err := m.UpdateColumn(1, "name", "bob")
```