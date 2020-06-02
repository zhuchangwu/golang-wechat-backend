package main

import (
	"io"
	"net/http"
	"regexp"
)

// webController， 映射控制器和pattern的关系
type WebController struct{
	Function func(w http.ResponseWriter, r *http.Request)
	Method string
	Pattern string
}

var mux []WebController

func init() {
	mux = append(mux, WebController{post,"POST","^/"})
	mux = append(mux, WebController{get,"GET","^/"})
}
type httpHandler struct {

}

// 让处理继承 http.Handler ， 重写ServeHttp， 给main函数使用
// 然后在这里面使用我们自己定义的WebController
func (*httpHandler)ServeHTTP(w http.ResponseWriter,r *http.Request){
	for _,webController := range mux {
		if ok,_:=regexp.MatchString(webController.Pattern,r.URL.Path);ok{
			if r.Method == webController.Method{
				webController.Function(w,r)
				return
			}
		}
	}
	io.WriteString(w,"")
	return
}










