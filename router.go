package cob

type HandleFunc func(ctx *Context)

type MiddlewareFunc func(handleFunc HandleFunc) HandleFunc

type Router struct {
	groups []*RouterGroup
	engin  *Engine
}

func (r *Router) Group(name string) *RouterGroup {
	group := &RouterGroup{
		name:              name,
		handleFuncMap:     make(map[string]map[string]HandleFunc),
		handlerMethodMap:  make(map[string][]string),
		middlewareFuncMap: make(map[string]map[string][]MiddlewareFunc),
		treeNode:          &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	group.Use(r.engin.Middles...)
	r.groups = append(r.groups, group)
	return group
}
