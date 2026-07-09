package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockAuthentication is a test double for Authentication[U].
type mockAuthentication struct {
	user    *mockUser
	signErr error
}

type mockUser struct {
	ID   int
	Name string
}

func (m *mockAuthentication) Init(ctx *core.Context) error { return nil }
func (m *mockAuthentication) SignIn(value any, request *web.Request) (any, error) {
	return &mockUser{ID: 1, Name: "signed-in"}, nil
}
func (m *mockAuthentication) SignOut(request *web.Request) (any, error) {
	m.user = nil
	return "signed-out", nil
}
func (m *mockAuthentication) User(request *web.Request) (*mockUser, error) {
	if m.user == nil {
		return nil, NoLogin
	}
	return m.user, nil
}

func newTestContext() *core.Context {
	cfg := config2.NewConfig()
	servers := web.NewServers()
	server, _ := servers.CreateServer(web.DefaultServerConfig())
	return core.NewContext(server, cfg, context.Background())
}

func newTestRequest(method, path string, meta *web.HandlerMeta) *web.Request {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, path, nil)
	mockResp := &mockResponse{ResponseWriter: c.Writer, ctx: c}
	return web.NewRequestForTest(c, mockResp, meta)
}

func TestAuth_NoLoginError(t *testing.T) {
	err := NoLogin
	assert.Error(t, err)
	assert.Equal(t, "Not logged in", err.Error())
}

func TestAuth_WithLogin(t *testing.T) {
	opt := WithLogin()
	assert.NotNil(t, opt)

	// WithLogin creates a MetaOption that sets LoginKey.
	// Apply it via a route's WithMeta, then verify through Handle behavior
	// (covered by TestAuth_Handle_RequiresLogin_* tests which use LoginKey directly).
}

func TestAuth_NewAuthenticationFilter(t *testing.T) {
	mockAuth := &mockAuthentication{}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	assert.NotNil(t, f)
	assert.NotNil(t, f.authentication)
}

func TestAuth_Init_WithAuthentication(t *testing.T) {
	mockAuth := &mockAuthentication{}
	f := NewAuthenticationFilter[*mockUser](mockAuth)

	err := f.Init(newTestContext())
	assert.NoError(t, err)
	assert.NotNil(t, f.ctx)
}

func TestAuth_Init_WithNilAuthentication(t *testing.T) {
	f := &AuthenticationFilter[*mockUser]{}
	err := f.Init(newTestContext())
	assert.NoError(t, err)
}

func TestAuth_Init_AuthInitFails(t *testing.T) {
	mockAuth := &failingAuth{}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	err := f.Init(newTestContext())
	assert.Error(t, err)
}

func TestAuth_SignIn(t *testing.T) {
	mockAuth := &mockAuthentication{}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	_ = f.Init(newTestContext())

	req := newTestRequest(http.MethodPost, "/signin", nil)
	result, err := f.SignIn(nil, req)
	assert.NoError(t, err)
	user, ok := result.(*mockUser)
	assert.True(t, ok)
	assert.Equal(t, "signed-in", user.Name)
}

func TestAuth_SignOut(t *testing.T) {
	mockAuth := &mockAuthentication{user: &mockUser{ID: 1, Name: "alice"}}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	_ = f.Init(newTestContext())

	req := newTestRequest(http.MethodPost, "/signout", nil)
	result, err := f.SignOut(req)
	assert.NoError(t, err)
	assert.Equal(t, "signed-out", result)
}

func TestAuth_Handle_RequiresLogin_NotAuthenticated(t *testing.T) {
	mockAuth := &mockAuthentication{user: nil}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	_ = f.Init(newTestContext())

	meta := web.NewHandlerMeta()
	meta.Add(LoginKey, true)
	req := newTestRequest(http.MethodGet, "/protected", meta)

	fc := &mockFilterChain{nextFunc: func() (any, error) {
		return "should-not-reach", nil
	}}

	result, err := f.Handle(fc, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, NoLogin))
}

func TestAuth_Handle_RequiresLogin_Authenticated(t *testing.T) {
	mockAuth := &mockAuthentication{user: &mockUser{ID: 1, Name: "alice"}}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	_ = f.Init(newTestContext())

	meta := web.NewHandlerMeta()
	meta.Add(LoginKey, true)
	req := newTestRequest(http.MethodGet, "/protected", meta)

	called := false
	fc := &mockFilterChain{nextFunc: func() (any, error) {
		called = true
		return "ok", nil
	}}

	result, err := f.Handle(fc, req)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "ok", result)
}

func TestAuth_Handle_NoLoginRequired(t *testing.T) {
	mockAuth := &mockAuthentication{user: nil}
	f := NewAuthenticationFilter[*mockUser](mockAuth)
	_ = f.Init(newTestContext())

	// No LoginKey in meta → should pass through without auth check
	meta := web.NewHandlerMeta()
	req := newTestRequest(http.MethodGet, "/public", meta)

	called := false
	fc := &mockFilterChain{nextFunc: func() (any, error) {
		called = true
		return "ok", nil
	}}

	result, err := f.Handle(fc, req)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "ok", result)
}

func TestAuth_User(t *testing.T) {
	mockAuth := &mockAuthentication{user: &mockUser{ID: 1, Name: "alice"}}
	f := NewAuthenticationFilter[*mockUser](mockAuth)

	req := newTestRequest(http.MethodGet, "/me", nil)
	user, err := User(f, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "alice", user.Name)
}

func TestAuth_User_NotLoggedIn(t *testing.T) {
	mockAuth := &mockAuthentication{user: nil}
	f := NewAuthenticationFilter[*mockUser](mockAuth)

	req := newTestRequest(http.MethodGet, "/me", nil)
	user, err := User(f, req)
	assert.Error(t, err)
	assert.Nil(t, user)
}

// --- helpers ---

type failingAuth struct{}

func (f *failingAuth) Init(ctx *core.Context) error { return errors.New("init failed") }
func (f *failingAuth) SignIn(value any, request *web.Request) (any, error) {
	return nil, nil
}
func (f *failingAuth) SignOut(request *web.Request) (any, error) { return nil, nil }
func (f *failingAuth) User(request *web.Request) (*mockUser, error) {
	return nil, nil
}

type mockResponse struct {
	gin.ResponseWriter
	ctx *gin.Context
}

func (r *mockResponse) SetAttachmentFileName(fileName string)   {}
func (r *mockResponse) JSON(code int, value any)                {}
func (r *mockResponse) Abort()                                  {}
func (r *mockResponse) Redirect(code int, location string)      {}
func (r *mockResponse) FileAttachment(path string, name string) {}
func (r *mockResponse) WriteStatus(code int)                    {}
func (r *mockResponse) Message(t *web.Message)                  {}
func (r *mockResponse) AbortWithMessage(t *web.Message)         {}
func (r *mockResponse) AbortWithStatusJSON(i int, value any)    {}
func (r *mockResponse) AbortWithError(err error) error {
	return r.ctx.AbortWithError(http.StatusInternalServerError, err)
}

type mockFilterChain struct {
	nextFunc func() (any, error)
}

func (m *mockFilterChain) Next() (any, error) {
	return m.nextFunc()
}
