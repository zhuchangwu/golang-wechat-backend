[TOC]



### 一、Golang模拟用户登陆，突破教务系统



#### 1.1 请求登陆页面

整个流程中的第一步是获取登陆页面，就像下图这样人为的通过浏览器访问服务端，服务端返回反馈返回登陆页面

![image-20200602062729513](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200602100219961-120669290.png)

访问登陆页面的目的上图中标注出来了，为了获取到Cookie，给真正发起登陆到请求方法使用。

下面的golang发送http到get请求，获取登陆页面的代码：

 ```go
// 访问登陆也，获取cookie
func GetCookieFromLoginhtml(url string) (cookie string, e error) {
	res, err := http.Get(url)
	if err != nil {
		e = err
	}
	// 获取cookie
	cookie = res.Header.Get("Set-Cookie")
	cookie = util.GetOneValueByPrefixAndSurfix("JSESSIONID=", "; Path=/", cookie)
	res.Body.Close()
	return
}
 ```



#### 1.2 抓包分析登陆请求

输入账号账号密码后点击登陆，将向后端发送登陆请求，如下图：

![image-20200602063307341](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200602100220631-432004266.png)





分析向后端发送到登陆请求都携带了哪些请求参数，携带了哪些请求头信息，以及需要通过Content-Type判断，该如何处理form表单中的数据发送到后台。后台才能正常响应。

![image-20200602064317230](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200602100221242-2049638668.png)



在浏览器的控制台中我们可以去看下登陆页面源码

![image-20200602065641448](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200602100221716-1041379403.png)

登陆页面对应的js源码

![image-20200602070318841](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200602100222213-846199918.png)





#### 1.3 golang使用js引擎合成salt

这一步也是必须的，所谓获取salt，其实就是通过golang使用js引擎执行`encodeInp(xxx)`, 这样我们才能得到经过加密后的username和password，进一步获取到encoded

```go
import (
	"github.com/robertkrimen/otto"
	"io/ioutil"
)
func EncodeInp(input string)(result string,e error)  {
	jsfile := "js/encodeUriJs.js"
	bytes, err := ioutil.ReadFile(jsfile)
	if err != nil {
		e = err
	}
	vm := otto.New()
	_, err = vm.Run(string(bytes))
	if err != nil {
		e = err
	}
	enc,err :=vm.Call("encodeInp",nil,input)
	if err != nil {
		e = err
	}
	result = enc.String()
	return
}
```

js部分的代码就不往外贴了，可以去下面的github地址中获取



#### 1.4 模拟表单提交，完成登陆

使用golang模拟登陆请求

```go
// 模拟登陆
func login(salt, cookie string) (html string) {
	
	req, err := http.NewRequest("POST", LoginUrl, strings.NewReader("encoded="+salt))
  
	// 添加请求头
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	//发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("无密版qlu教务系统登陆请求失败 : %v", err)
		return
	}
	// todo 根据状态码判断下一步如何操作，如果状态码是302，表示操作成功
	fmt.Println("resp.Status:  ", resp.Status)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	// 返回个人主页的html
	html = string(b)
	// 手动关闭
	resp.Body.Close()
	return
}

```

这一步中值得注意的地方：

**第一：我们发送的请求的类型是POST请求**



**第二：我们应该如何处理form表单中的数据后，再发送给后端，后端才能正常处理呢？**

具体处理成什么样，是需要根据请求头中的Content-Type决定的。

不知道大家知不知道常见的Content-Type的几种类型：在form 表单中有一个属性叫做 entype可以间接将数据处理成Content-Type指定数据格式， 比如我们可以像这样设置：

* `enctype = text/plain` 那么form表单最终提交的格式就是： 用纯文本的形式发送。

* `enctype = application/x-www-form-urlencoded` 
  * 表单中的enctype值如果不设置，则默认是application/x-www-form-urlencoded，它会将表单中的数据变为键值对的形式。
  * 如果action为get，则将表单数据编码为(name1=value1&name2=value2…)，然后把这个字符串加到url后面，中间用?分隔。
  * 如果action为post，浏览器把form数据封装到http body中，然后发送到服务器。

* `enctype = mutipart/form-data` 
  * 上传的是非文本内容，比如是个图片，文件，mp3。



根据这个知识点，结合我们当前的情况，method=post，Content-Type = application/x-www-form-urlencoded

所以，在选择golang的api时，我们选择下图这个api使用

![image-20200602073500224](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200602100222683-1330209311.png)



#### 1.5 进入成绩查询页，解析用户成绩

如果不出意外，经过上面的处理，我们已经完成登陆，并且获取到后台页面的html源码了。

再之后我们就直奔成绩查询模块，还是使用如何的分析思路

```go
func getAllScore(stuIdentify, cookie string) ([]mtStruct.Score, error) {
	// 发送查询成绩的请求
	u := "http://jwxt.qlu.edu.cn/jsxsd/kscj/cjcx_list"
	req, err := http.NewRequest("POST", u, strings.NewReader("kksj=&kcxz=&kcmc=&xsfs=all"))
	if err != nil {
		fmt.Printf("error : %v", err)
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	req.Header.Add("Referer", "http://jwxt.qlu.edu.cn/jsxsd/kscj/cjcx_query?Ves632DSdyV=NEW_XSD_XJCJ")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	client := &http.Client{}
	resp, err := client.Do(req)
  ...
  
}
```

代码详情可以去github上查看。



### 二、植入微信公共号后台

上面的功能实现后再结合Golang开发微信公众号就能实现一款好玩的应用。

让用户通过微信公共号平台和后端进行数据的交互，我们获取到用户的信息，拿着用户的信息帮用户监听教务系统的成绩单的状态。一旦有成绩第一时间推送给用户。

[点击查查看公众号端设计思路](https://mp.weixin.qq.com/cgi-bin/appmsg?t=media/appmsg_edit&action=edit&type=10&appmsgid=100000011&token=11620974&lang=zh_CN)


