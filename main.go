package main

import (
	"fmt"
	"weixin-backend/wechat/wx"
	"net/http"
	"time"
	"weixin-backend/mail"
	"weixin-backend/task"
)

const port = 80
const token = "xxxxxxx"

/*
	通过聊天窗口发送的信息会被wx以post请求的方式转发到后端go
	消息体是XML格式的数据
	ToUserName: 公众号的原始ID
	FromUserName: 用户的OpenId
	CreateTime: 时间戳
 	MsgType: text
	Content: 发送的内容
	MsgId： 消息的ID
*/
func post(w http.ResponseWriter, r *http.Request) {
	// 将r，w，token传递给sdk按照微信文档的要求进行按章
	client, err := wx.NewClient(r, w, token)
	if err != nil {
		fmt.Printf("error : %v", err)
		w.WriteHeader(403)
		return
	}
	client.Run()
	return
}

/*
	参照：https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Access_Overview.html
	程序的整个生命周期过程中，只会收到一次get请求，就是在我们点击保存 基本配置 时的请求
*/
func get(w http.ResponseWriter, r *http.Request) {
	// 将r，w，token传递给sdk按照微信文档的要求进行验证
	// 参照URL中的第二步：https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Access_Overview.html
	client, err := wx.NewClient(r, w, token)
	if err != nil {
		fmt.Printf("error : %v", err)
		w.WriteHeader(403)
		return
	}
	// 如果上面的验证通过了，按照微信的要求，将echostr原封不动的返回
	if len(client.Query.Echostr) > 0 {
		// wx要求我们收到echostr后将其写回
		w.Write([]byte(client.Query.Echostr))
		return
	}
	w.WriteHeader(403)
	return
}

func StartTask() {
	c := time.Tick(1 * time.Hour)
	for now := range c {
		task.CornTask()
		if now.Hour() == 12 {
			mail.DoSendMail("646450308@qq.com", "健康状态", time.Now().String()+"正常")
		}
	}
}

func main() {
	// 开启定时任务
	fmt.Println("准备开启定时任务")
	go StartTask()
	fmt.Println("定时任务已开启")
	// 开启服务器
	handler := &httpHandler{}
	server := http.Server{
		// 注意这个： 不要丢
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 0,
	}

	fmt.Println("listen : ", port)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
}
