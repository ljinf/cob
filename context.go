package cob

import (
	"errors"
	"fmt"
	"github.com/ljinfu/cob/binding"
	"github.com/ljinfu/cob/log"
	"github.com/ljinfu/cob/render"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const defaultMaxMemory = 32 << 20 //32m

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	engine     *Engine
	queryCache url.Values
	formCache  url.Values

	DisallowUnknownFields bool
	IsInvalid             bool
	StatusCode            int

	Logger *log.Logger

	Keys     map[string]interface{}
	mu       sync.RWMutex
	sameSite http.SameSite
}

func (c *Context) initQueryCache() {
	if c.Request != nil {
		c.queryCache = c.Request.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) initFormCache() {
	if c.Request != nil {
		if err := c.Request.ParseMultipartForm(defaultMaxMemory); err != nil {
			if errors.Is(err, http.ErrNotMultipart) {
				fmt.Println(err)
			}
		}
		c.formCache = c.Request.PostForm
	} else {
		c.formCache = url.Values{}
	}
}

func (c *Context) GetQuery(key string) string {
	return c.queryCache.Get(key)
}

func (c *Context) GetDefaultQuery(key, defaultVal string) string {
	array, ok := c.GetQueryArray(key)
	if !ok {
		return defaultVal
	}
	return array[0]
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	strings, ok := c.queryCache[key]
	return strings, ok
}

func (c *Context) QueryArray(key string) []string {
	return c.queryCache[key]
}

//http://xxx:xx/naa?user[id]=1&user[name]=a
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	return c.get(c.queryCache, key)
}

func (c *Context) get(cache map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range cache {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}

	}
	return dicts, exist
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if array, ok := c.GetPostFormArray(key); ok {
		return array[0], ok
	}
	return "", false
}

func (c *Context) PostForm(key string) string {
	val, _ := c.GetPostForm(key)
	return val
}

func (c *Context) PostFormArray(key string) []string {
	array, _ := c.GetPostFormArray(key)
	return array
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	vals, ok := c.formCache[key]
	return vals, ok
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	return c.get(c.formCache, key)
}

func (c *Context) PostFormMap(key string) map[string]string {
	formMap, _ := c.GetPostFormMap(key)
	return formMap
}

func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	file, header, err := c.Request.FormFile(key)
	defer file.Close()
	return header, err
}

func (c *Context) FormFiles(key string) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	return form.File[key], err
}

//将file 保存到dst
func (c *Context) TransferTo(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(defaultMaxMemory)
	return c.Request.MultipartForm, err
}

func (c *Context) BindJson(obj interface{}) error {
	json := binding.JSON
	json.DisallowUnknownFields = true
	json.IsInvalid = true
	return c.MustBindWith(obj, &json)
}

func (c *Context) BindXML(obj interface{}) error {
	return c.MustBindWith(obj, &binding.XML)
}

func (c *Context) HTML(status int, html string) error {
	return c.Render(status, &render.HTML{Data: html, IsTemplate: false})
}

func (c *Context) HTMLTemplate(name string, data interface{}, filenames ...string) error {
	c.Writer.Header().Set("Content-Type", "text/html;charset=utf-8")
	t := template.New(name)
	t, err := t.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	err = t.Execute(c.Writer, data)
	return err
}

func (c *Context) HTMLTemplateGlob(name string, data interface{}, pattern string) error {
	c.Writer.Header().Set("Content-Type", "text/html;charset=utf-8")
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(c.Writer, data)
	return err
}

func (c *Context) Template(name string, data interface{}) error {
	return c.Render(http.StatusOK, &render.HTML{
		Data:       data,
		Name:       name,
		Template:   c.engine.HTMLRender.Template,
		IsTemplate: true,
	})
}

func (c *Context) JSON(status int, value interface{}) error {
	return c.Render(status, &render.Json{Data: value})
}

func (c *Context) XML(status int, value interface{}) error {
	return c.Render(status, &render.Xml{Data: value})
}

func (c *Context) File(filename string) {
	http.ServeFile(c.Writer, c.Request, filename)
}

func (c *Context) FileAttachment(filePath, filename string) {
	if isASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", `attachment;filename="`+filename+`"`)
	} else {
		c.Writer.Header().Set("Content-Disposition", `attachment;filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.Writer, c.Request, filePath)
}

//filepath 是相对于文件系统的路径
func (c *Context) FileFormFS(filePath string, fs http.FileSystem) {
	defer func(old string) {
		c.Request.URL.Path = old
	}(c.Request.URL.Path)

	c.Request.URL.Path = filePath

	http.FileServer(fs).ServeHTTP(c.Writer, c.Request)
}

//重定向
func (c *Context) Redirect(status int, url string) error {
	return c.Render(status, &render.Redirect{
		Code:     status,
		Request:  c.Request,
		Location: url,
	})
}

func (c *Context) String(status int, format string, val ...interface{}) error {
	return c.Render(status, &render.String{Format: format, Data: val})
}

func (c *Context) Render(status int, r render.Render) error {
	err := r.Render(c.Writer, status)
	c.StatusCode = status
	return err
}

func (c *Context) MustBindWith(obj interface{}, bind binding.Binding) error {
	err := c.ShouldBind(obj, bind)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
	}
	return err
}

func (c *Context) ShouldBind(obj interface{}, bind binding.Binding) error {
	return bind.Bind(c.Request, obj)
}

func (c *Context) Fail(code int, msg string) {
	c.String(code, msg)
}

func (c *Context) HandleWithError(statusCode int, obj interface{}, err error) {
	if err != nil {
		code, data := c.engine.errHandler(err)
		c.JSON(code, data)
		return
	}
	c.JSON(statusCode, obj)
}

func (c *Context) SetBasicAuth(username, password string) {
	c.Request.SetBasicAuth(username, password)
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
	c.mu.Unlock()
}

func (c *Context) Get(key string) (value interface{}, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.Keys[key]
	return
}

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: c.sameSite,
	})
}

func (c *Context) SetSameSite(s http.SameSite) {
	c.sameSite = s
}
