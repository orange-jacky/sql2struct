package main

import (
	"io/ioutil"
	"os"
)

func main() {
	filename := os.Args[1]
	//
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var content []byte
	content, err = ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	//
	Sql2Struct(string(content))
}
