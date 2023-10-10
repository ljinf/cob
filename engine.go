package cob

import (
	"fmt"
	coblog "github.com/ljinfu/cob/log"
	"github.com/ljinfu/cob/render"
	"html/template"
	"net/http"
	"sync"
)

type ErrorHandler func(err error) (int, interface{})

type Engine struct {
	Router
	funcMap    template.FuncMap
	HTMLRender *render.HTMLRender
	pool       sync.Pool

	Logger *coblog.Logger

	Middles    []MiddlewareFunc
	errHandler ErrorHandler
}

func New() *Engine {
	engine := &Engine{
		Router: Router{},
	}
	engine.Router.engin = engine
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Router.engin = engine
	engine.Logger = coblog.Default()
	engine.Use(Logging, Recovery)
	return engine
}

func (e *Engine) allocateContext() interface{} {
	return &Context{engine: e}
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = r
	ctx.Logger = e.Logger
	//初始化query参数
	ctx.initQueryCache()
	ctx.initFormCache()

	e.httpRequestHandle(ctx, w, r)
	e.pool.Put(ctx)
}

func (e *Engine) httpRequestHandle(ctx *Context, w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, group := range e.groups {
		routerName := SubStringLast(r.URL.Path, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil && node.isEnd {
			handler, ok := group.handleFuncMap[node.routerName][ANY]
			if ok {
				group.MethodHandle(node.routerName, ANY, ctx, handler)
				return
			}
			//method 匹配
			handler, ok = group.handleFuncMap[node.routerName][method]
			if ok {
				group.MethodHandle(node.routerName, method, ctx, handler)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s %s not allowed \n ", r.RequestURI, method)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s %s not found \n ", r.RequestURI, method)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHtmlTemplate(t)
}

func (e *Engine) SetHtmlTemplate(t *template.Template) {
	e.HTMLRender = &render.HTMLRender{Template: t}
}

func (e *Engine) Use(handleFunc ...MiddlewareFunc) {
	e.Middles = append(e.Middles, handleFunc...)
}

func (e *Engine) RegistryErrHandler(handler ErrorHandler) {
	e.errHandler = handler
}
