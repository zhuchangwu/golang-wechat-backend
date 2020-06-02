package js

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



func GetEncoded(data ,username , password string) (salt string,e error){
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
	enc,err :=vm.Call("getEncode",nil,data,username,password)
	if err != nil {
		e = err
	}
	salt = enc.String()
	return
}

func EncodeUri(uri string)(result string,e error)  {
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

	enc,err :=vm.Call("encodeUri",nil,uri)
	if err != nil {
		e = err
	}

	result = enc.String()
	return
}
