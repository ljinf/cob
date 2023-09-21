package cob

import "net/http"

const ANY = "ANY"

//路由组
type RouterGroup struct {
	name              string
	handleFuncMap     map[string]map[string]HandleFunc //第一层为router url  ,第二层为post/get等method
	handlerMethodMap  map[string][]string
	middlewareFuncMap map[string]map[string][]MiddlewareFunc

	treeNode    *treeNode
	middlewares []MiddlewareFunc
}

func (g *RouterGroup) Use(middlewareFunc ...MiddlewareFunc) {
	g.middlewares = append(g.middlewares, middlewareFunc...)
}

func (g *RouterGroup) MethodHandle(routerName, method string, ctx *Context, handleFunc HandleFunc) {
	//组中间件
	if g.middlewares != nil {
		for _, middle := range g.middlewares {
			handleFunc = middle(handleFunc)
		}
	}

	//路由中间件
	funcs := g.middlewareFuncMap[routerName][method]
	if funcs != nil {
		for _, f := range funcs {
			handleFunc = f(handleFunc)
		}
	}
	handleFunc(ctx)
}

//func (g *RouterGroup) Add(pattern string, handler HandleFunc) {
//	g.handleFuncMap[pattern] = handler
//}

func (g *RouterGroup) handle(pattern, method string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	_, ok := g.handleFuncMap[pattern]
	if !ok {
		g.handleFuncMap[pattern] = make(map[string]HandleFunc)
		g.middlewareFuncMap[pattern] = make(map[string][]MiddlewareFunc)
	}
	_, ok = g.handleFuncMap[pattern][method]
	if ok {
		panic("有重复的路由")
	}
	g.handleFuncMap[pattern][method] = handler

	g.middlewareFuncMap[pattern][method] = append(g.middlewareFuncMap[pattern][method], middlewareFunc...)

	g.treeNode.Put(pattern)
}

func (g *RouterGroup) Any(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, ANY, handler, middlewareFunc...)
}

func (g *RouterGroup) Get(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodGet, handler, middlewareFunc...)
}

func (g *RouterGroup) Post(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodPost, handler, middlewareFunc...)
}

func (g *RouterGroup) Put(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodPut, handler, middlewareFunc...)
}

func (g *RouterGroup) Delete(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodDelete, handler, middlewareFunc...)
}

func (g *RouterGroup) Patch(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodPatch, handler, middlewareFunc...)
}

func (g *RouterGroup) Options(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodOptions, handler, middlewareFunc...)
}

func (g *RouterGroup) Head(pattern string, handler HandleFunc, middlewareFunc ...MiddlewareFunc) {
	g.handle(pattern, http.MethodHead, handler, middlewareFunc...)
}
