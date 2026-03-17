package main

import (
	"context"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

// UserController is a complete REST controller example
type UserController struct {
	// Embed core.IService interface
	core.IService
}

// Init initializes the controller and registers routes
func (u *UserController) Init(ctx *core.Context) error {
	// Register routes through context
	ctx.Get("/users", u.GetUsers)
	ctx.Get("/users/:id", u.GetUser)
	ctx.Post("/users", u.CreateUser)
	ctx.Put("/users/:id", u.UpdateUser)
	ctx.Delete("/users/:id", u.DeleteUser)

	return nil
}

// GetUsers handles getting all users with pagination
func (u *UserController) GetUsers(c *web.Request) (any, error) {
	// Access query parameters
	page := c.Query("page")
	limit := c.Query("limit")

	return map[string]any{
		"users": []string{"alice", "bob"},
		"page":  page,
		"limit": limit,
	}, nil
}

// GetUser handles getting a single user by ID
func (u *UserController) GetUser(c *web.Request) (any, error) {
	id := c.Param("id")
	return map[string]any{
		"id":   id,
		"name": "alice",
	}, nil
}

// CreateUser handles creating a new user
func (u *UserController) CreateUser(c *web.Request) (any, error) {
	var user struct {
		Name string `json:"name"`
	}
	if err := c.BindJSON(&user); err != nil {
		return nil, err
	}

	// In a real app, you would save to database here
	return map[string]any{
		"id":   1,
		"name": user.Name,
	}, nil
}

// UpdateUser handles updating an existing user
func (u *UserController) UpdateUser(c *web.Request) (any, error) {
	id := c.ParamInt("id")

	var user struct {
		Name string `json:"name"`
	}
	if err := c.BindJSON(&user); err != nil {
		return nil, err
	}

	// In a real app, you would update in database here
	return map[string]any{
		"id":   id,
		"name": user.Name,
	}, nil
}

// DeleteUser handles deleting a user
func (u *UserController) DeleteUser(c *web.Request) (any, error) {
	id := c.ParamInt("id")

	// In a real app, you would delete from database here
	return map[string]any{
		"id":      id,
		"deleted": true,
	}, nil
}

func main() {
	app := wf.NewWithAutoConfig()
	app.AddRest(&UserController{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}
}
