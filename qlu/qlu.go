package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"weixin-backend/js"
	"weixin-backend/util"
)

// 教务系统地址
var qluUrl = "http://jwxt.qlu.edu.cn"

// 发送预检请的地址
var urlEncode = "http://jwxt.qlu.edu.cn/Logon.do?method=logon&flag=sess"

// 真正发送登陆请求的url
var LoginUrl = "http://jwxt.qlu.edu.cn/Logon.do?method=logon"

/*根据url获取登陆页面html+cookie*/
func GetHtmlByLoginUrl(u string) (html string, cookie string, e error) {
	res, err := http.Get(u)
	if err != nil {
		e = err
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if e != nil {
		e = err
	}
	// 获取cookie
	cookie = res.Header.Get("Set-Cookie")
	cookie = util.GetOneValueByPrefixAndSurfix("JSESSIONID=", "; Path=/", cookie)
	// 获取html
	html = string(bytes)
	return
}

// 发送预检请求，获取用于计算slat的返回值
func getSaltResponse(cookie string) (dataStr string) {
	// 构建请求url
	req, err := http.NewRequest("POST", urlEncode, nil)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	// 创建客户端，发送请求
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	dataStr = string(b)
	return
}

// 模拟js，拼接出加密参数
func getSalt(cookie, username, password string) (salt string) {
	dataStr := getSaltResponse(cookie)
	scode := strings.Split(dataStr, "#")[0]
	sxh := strings.Split(dataStr, "#")[1]
	code := username + "%%%" + password
	var encode string
	for i := 0; i < len(code); i++ {
		if i < 20 {
			index, err := strconv.Atoi(sxh[i : i+1])
			if err != nil {
				fmt.Printf("error : %v", err)
				return
			}
			encode = encode + code[i:i+1] + scode[0:index]
			scode = scode[index:len(scode)]
		} else {
			encode = encode + code[i:len(code)]
			i = len(code)
		}
	}
	salt = encode
	return
}

// 通过js引擎获取加密串
func getSalt2(cookie, username, password string) (salt string) {
	dataStr := getSaltResponse(cookie)
	salt, e := js.GetEncoded(dataStr, username, password)
	if e != nil {
		fmt.Printf("error 通过js获取盐失败 : %v", e)
		return
	}
	return
}

func getImg(url string) (n int64, err error) {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	fmt.Println(name)
	out, err := os.Create(name + ".png")
	defer out.Close()
	resp, err := http.Get(url)
	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	n, err = io.Copy(out, bytes.NewReader(pix))
	return

}

func main() {
	// 向登陆页面发送请求
	html, cookie, err := GetHtmlByLoginUrl(qluUrl)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}

	// 获取验证码链接
	qluResult := util.GetOneValueByPrefixAndSurfix("<img  src=\"", "\" id=\"SafeCodeImg\"", html)
	targetCodeUrl := qluUrl + qluResult

	// fmt.Println("html:  ", html)
	fmt.Println("cookie: ", cookie)
	fmt.Println("验证码链接：", targetCodeUrl)
	go getImg(targetCodeUrl)
	// todo 通过grpc 发送请求，获取验证码的值
	// todo 通过键盘录入，模拟验证码的获取
	var code string
	fmt.Scan(&code)
	// todo 账号密码从mysql中查询
	username := "201708120034"
	password := "2424zcw.."

	// 发送post登陆请求：
	// 获取加密参数
	salt := getSalt2(cookie, username, password)

	// 使用js对uri进行编码
	salt, err = js.EncodeUri(salt)
	if err != nil {
		fmt.Printf("使用js引擎编码出错 : %v", err)
		return
	}

	// data := url.Values{}
	// data.Set("RANDOMCODE", code)
	// data.Set("encoded", salt)
	// data.Set("useDogCode", "")
	// data.Set("view", "1")

	 u, _ := url.ParseRequestURI(qluUrl)
	 u.Path = "/Logon.do?method=logon"
	 urlStr := u.String()

	// todo 姑且认为，请求已经正常发送了，但是验证码机制有问题
	// view=1&useDogCode=&encoded=2n7b0e31i27fM098G21r2K70M620t232E409%25Y%25Tw0%25Y2xB42qg24406ze5Lcw..&
	// queryStr:=""
	// queryStr:="view=1&useDogCode=&encoded="+salt+"&RANDOMCODE="+code
	queryStr := "view=1&useDogCode=&encoded=" + salt + "&RANDOMCODE=" + code
	// 构建请求,没有请求体，仅仅是post调表单
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(queryStr))
	// 添加请求头
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Referer", "http://jwxt.qlu.edu.cn/")

	//err = req.ParseForm()

	/*主管：陈存立*/
	if err != nil {
		fmt.Printf("err:=req.ParseForm() : %v", err)
		return
	}
	// 创建客户端，发送请求
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	fmt.Println("resp.Status , ", resp.Status)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println(string(b))
}
