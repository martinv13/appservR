package controllers

import (
	"errors"
	"html/template"
	"net/http/httptest"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/config"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// invoke runs a handler and flushes any pending status/header write.
// gin's real engine always calls c.Writer.WriteHeaderNow() once the handler
// chain finishes (see (*Engine).handleHTTPRequest), which is what makes a
// bare c.Redirect() on a POST actually reach the ResponseWriter (net/http's
// Redirect only writes a body - triggering the flush as a side effect - for
// GET/HEAD requests). Calling handlers directly, outside that engine loop,
// skips this, so tests must do it explicitly.
func invoke(c *gin.Context, h gin.HandlerFunc) {
	h(c)
	c.Writer.WriteHeaderNow()
}

// newTestContext builds a gin context wired with stub templates for every
// view these controllers render, so c.HTML() calls don't panic on a missing
// template while we're only asserting on status codes / data, not markup.
func newTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)

	tmpl := template.New("")
	for _, name := range []string{
		"apps.html", "app.html",
		"users.html", "user.html",
		"groups.html", "group.html",
		"login.html", "signup.html", "signupsuccess.html",
	} {
		template.Must(tmpl.New(name).Parse("{{.errorMessage}}{{.successMessage}}"))
	}
	engine.SetHTMLTemplate(tmpl)

	return c, w
}

type fakeConfig struct {
	logger config.Logger
}

func (c *fakeConfig) ExecutableFolder() string { return "." }
func (c *fakeConfig) GetString(string) string  { return "" }
func (c *fakeConfig) Logger() *config.Logger   { return &c.logger }

func newFakeConfig() *fakeConfig {
	return &fakeConfig{logger: config.NewLogger(3)}
}

// fakeAppModel is a minimal in-memory models.AppModel for controller tests.
type fakeAppModel struct {
	apps    map[string]models.App
	saveErr error
	delErr  error
}

func newFakeAppModel(apps ...models.App) *fakeAppModel {
	m := &fakeAppModel{apps: map[string]models.App{}}
	for _, a := range apps {
		m.apps[a.Name] = a
	}
	return m
}

func (m *fakeAppModel) All() ([]models.App, error) {
	var res []models.App
	for _, a := range m.apps {
		res = append(res, a)
	}
	return res, nil
}
func (m *fakeAppModel) Find(name string) (models.App, error) {
	a, ok := m.apps[name]
	if !ok {
		return models.App{}, errors.New("app not found")
	}
	return a, nil
}
func (m *fakeAppModel) Save(app models.App, oldName string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	delete(m.apps, oldName)
	m.apps[app.Name] = app
	return nil
}
func (m *fakeAppModel) Delete(name string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.apps, name)
	return nil
}
func (m *fakeAppModel) AsMap(app models.App) (map[string]interface{}, error) {
	return map[string]interface{}{"Name": app.Name}, nil
}
func (m *fakeAppModel) AsMapSlice(apps []models.App) ([]map[string]interface{}, error) {
	res := make([]map[string]interface{}, len(apps))
	for i, a := range apps {
		res[i] = map[string]interface{}{"Name": a.Name}
	}
	return res, nil
}

// fakeUserModel is a minimal in-memory models.UserModel for controller tests.
type fakeUserModel struct {
	users     map[string]models.User
	saveErr   error
	adminErr  error
	delErr    error
	loginErr  error
	loginUser models.User
}

func newFakeUserModel(users ...models.User) *fakeUserModel {
	m := &fakeUserModel{users: map[string]models.User{}}
	for _, u := range users {
		m.users[u.Username] = u
	}
	return m
}

func (m *fakeUserModel) All() ([]models.User, error) {
	var res []models.User
	for _, u := range m.users {
		res = append(res, u)
	}
	return res, nil
}
func (m *fakeUserModel) Find(username string) (models.User, error) {
	u, ok := m.users[username]
	if !ok {
		return models.User{}, errors.New("user not found")
	}
	return u, nil
}
func (m *fakeUserModel) Save(user models.User, oldUsername string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	delete(m.users, oldUsername)
	m.users[user.Username] = user
	return nil
}
func (m *fakeUserModel) AdminSave(user models.User, oldUsername string) error {
	if m.adminErr != nil {
		return m.adminErr
	}
	delete(m.users, oldUsername)
	m.users[user.Username] = user
	return nil
}
func (m *fakeUserModel) Delete(username string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.users, username)
	return nil
}
func (m *fakeUserModel) AsMap(user models.User) (map[string]interface{}, error) {
	return map[string]interface{}{"Username": user.Username}, nil
}
func (m *fakeUserModel) AsMapSlice(users []models.User) ([]map[string]interface{}, error) {
	res := make([]map[string]interface{}, len(users))
	for i, u := range users {
		res[i] = map[string]interface{}{"Username": u.Username}
	}
	return res, nil
}
func (m *fakeUserModel) Login(user models.User) (models.User, error) {
	if m.loginErr != nil {
		return models.User{}, m.loginErr
	}
	return m.loginUser, nil
}

// fakeGroupModel is a minimal in-memory models.GroupModel for controller tests.
type fakeGroupModel struct {
	groups   map[string]models.Group
	saveErr  error
	delErr   error
	addCalls []string
	remCalls []string
}

func newFakeGroupModel(groups ...models.Group) *fakeGroupModel {
	m := &fakeGroupModel{groups: map[string]models.Group{}}
	for _, g := range groups {
		m.groups[g.Name] = g
	}
	return m
}

func (m *fakeGroupModel) AllNames() ([]string, error) {
	var res []string
	for n := range m.groups {
		res = append(res, n)
	}
	return res, nil
}
func (m *fakeGroupModel) AsMapSlice() ([]map[string]interface{}, error) {
	var res []map[string]interface{}
	for n := range m.groups {
		res = append(res, map[string]interface{}{"GroupName": n})
	}
	return res, nil
}
func (m *fakeGroupModel) Find(name string) (models.Group, error) {
	g, ok := m.groups[name]
	if !ok {
		return models.Group{}, errors.New("group not found")
	}
	return g, nil
}
func (m *fakeGroupModel) Save(group models.Group, oldGroupName string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	delete(m.groups, oldGroupName)
	m.groups[group.Name] = group
	return nil
}
func (m *fakeGroupModel) Delete(groupName string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.groups, groupName)
	return nil
}
func (m *fakeGroupModel) AddMember(groupName string, username string) error {
	m.addCalls = append(m.addCalls, groupName+"/"+username)
	return nil
}
func (m *fakeGroupModel) RemoveMember(groupName string, username string) error {
	m.remCalls = append(m.remCalls, groupName+"/"+username)
	return nil
}
