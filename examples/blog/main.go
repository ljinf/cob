package main

import (
	"fmt"
	"github.com/ljinfu/cob"
	error2 "github.com/ljinfu/cob/error"
	"github.com/ljinfu/cob/log"
	"net/http"
)

func main() {

	//http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
	//	fmt.Fprintf(writer, "%s 欢迎来到从零实现一个微服务框架！", "ljf.com")
	//})
	//
	//if err := http.ListenAndServe(":8080", nil); err != nil {
	//	log.Fatal(err)
	//}

	start()
}

type User struct {
	Name string `xml:"name" json:"name" cob:"required"`
	Age  int    `xml:"age" json:"age" cob:"required"`
}

func start() {
	engine := cob.New()
	engine.RegistryErrHandler(func(err error) (int, interface{}) {
		switch e := err.(type) {
		case *cobResponse:
			return http.StatusOK, e.Response()

		default:
			return http.StatusInternalServerError, "500 error"
		}
	})
	user := engine.Group("user")
	user.Use(cob.Logging, cob.Recovery)
	//user.Add("/hello", func(writer http.ResponseWriter, request *http.Request) {
	//	fmt.Fprintf(writer, "%s 欢迎来到从零实现一个微服务框架！", "ljf.com")
	//})

	user.Use(func(next cob.HandleFunc) cob.HandleFunc {
		return func(ctx *cob.Context) {
			fmt.Println("通用 middleware handler")
			next(ctx)
		}
	})

	//user.Get("/*/hello", func(ctx *cob.Context) {
	//	fmt.Fprintf(ctx.Writer, "%s get /*/hello", "ljf.com")
	//})

	user.Get("/hello/ab", func(ctx *cob.Context) {
		fmt.Fprintf(ctx.Writer, "%s get /hello/ab", "ljf.com")
	})

	user.Get("/info", func(ctx *cob.Context) {
		fmt.Fprintf(ctx.Writer, "%s get info", "ljf.com")
	}, func(next cob.HandleFunc) cob.HandleFunc {
		return func(ctx *cob.Context) {
			fmt.Println("路由 middleware handler")
			next(ctx)
		}
	})

	user.Post("/info", func(ctx *cob.Context) {
		fmt.Fprintf(ctx.Writer, "%s post info", "ljf.com")
	})

	user.Post("/name", func(ctx *cob.Context) {
		fmt.Fprintf(ctx.Writer, "%s post name", "ljf.com")
	})

	user.Get("/html", func(ctx *cob.Context) {
		ctx.HTML(http.StatusOK, "<h1>html test</h1>")
	})

	user.Get("/htmltemplate", func(ctx *cob.Context) {

		u := struct {
			Name string
		}{
			Name: "abc",
		}

		err := ctx.HTMLTemplate("index.html", u, "tpl/index.html")
		if err != nil {
			fmt.Println(err)
		}
	})

	user.Get("/login/htmltemplate", func(ctx *cob.Context) {
		err := ctx.HTMLTemplate("login.html", "", "tpl/login.html", "tpl/header.html")
		if err != nil {
			fmt.Println(err)
		}
	})

	user.Get("/login/glob", func(ctx *cob.Context) {
		err := ctx.HTMLTemplateGlob("login.html", "", "tpl/*.html")
		if err != nil {
			fmt.Println(err)
		}
	})

	//engine.LoadTemplate("tpl/*.html")
	user.Get("/template", func(ctx *cob.Context) {
		u := struct {
			Name string
		}{
			Name: "abc",
		}
		err := ctx.Template("login.html", u)
		if err != nil {
			fmt.Println(err)
		}
	})

	user.Get("/json", func(ctx *cob.Context) {
		u := struct {
			Name string
		}{
			Name: "abc",
		}
		ctx.JSON(http.StatusOK, u)
	})

	user.Get("/xml", func(ctx *cob.Context) {
		u := &User{
			Name: "abc",
			Age:  22,
		}
		ctx.XML(http.StatusOK, u)
	})

	user.Get("/fs", func(ctx *cob.Context) {
		ctx.FileFormFS("test.txt", http.Dir("tpl"))
	})

	user.Get("/redirect", func(ctx *cob.Context) {
		ctx.Redirect(http.StatusFound, "/user/template")
		//ctx.Redirect(http.StatusOK, "/user/template")
	})

	user.Get("/query", func(ctx *cob.Context) {
		id := ctx.GetQuery("id")
		name := ctx.GetQuery("name")
		fmt.Printf("%s  %s", id, name)
	})

	user.Get("/form", func(ctx *cob.Context) {
		id := ctx.PostForm("id")
		name := ctx.PostForm("name")
		fmt.Printf("%s  %s", id, name)
	})

	user.Post("/json", func(ctx *cob.Context) {
		u := &User{}
		err := ctx.BindJson(u)
		if err != nil {
			fmt.Println(err)
		}
		ctx.JSON(http.StatusOK, u)
	})

	logger := log.Default()
	logger.Level = log.LevelDebug
	logger.SetLogPath("./log")

	user.Post("/log", func(ctx *cob.Context) {
		logger.WithFields(log.Fields{
			"name": "abc",
			"id":   1000,
		}).Debug("我是debug日志")
		logger.Debug("我是debug日志")
		logger.Info("我是info日志")
		logger.Error("我是error日志")

		lError := error2.Default()
		lError.Result(func(e *error2.LError) {
			fmt.Print(e.Error())
		})

		ctx.JSON(http.StatusOK, "")
	})

	user.Get("/res", func(ctx *cob.Context) {
		err := login()
		ctx.HandleWithError(http.StatusOK, "", err)
	})

	engine.Run()
}

type cobResponse struct {
	Code int
	Msg  string
	Data interface{}
}

func (lb *cobResponse) Error() string {
	return lb.Msg
}

func (lb *cobResponse) Response() interface{} {
	return lb
}

func login() *cobResponse {
	return &cobResponse{
		Code: 500,
		Msg:  "",
		Data: nil,
	}
}
