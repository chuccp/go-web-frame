# Database

Advanced database usage.

## Transactions

```go
err := db.Transaction(func(tx *db.DB) error {
    userModel := model.NewModel[*User](tx, "t_user")
    orderModel := model.NewModel[*Order](tx, "t_order")
    
    user := &User{Name: "alice"}
    if err := userModel.Save(user); err != nil {
        return err
    }
    
    order := &Order{UserId: user.Id}
    return orderModel.Save(order)
})
```

## Model Groups

Share database connection across models:

```go
group := builder.NewModelGroup(db, "user_group")
group.AddModel(&UserModel{}, &ProfileModel{})
```

## Multiple Databases

```go
db1, err := db.NewDB(config1)
db2, err := db.NewDB(config2)

builder.ModelGroup("users").WithDB(db1).Register(&UserModel{})
builder.ModelGroup("orders").WithDB(db2).Register(&OrderModel{})
```

## Connection Pool

```yaml
web:
  db:
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600  # seconds
```

## Raw SQL

```go
result, err := db.Raw("SELECT * FROM users WHERE id = ?", 1)
```