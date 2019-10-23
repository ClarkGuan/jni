package main

import (
	"flag"
	"fmt"

	"github.com/ClarkGuan/jni/tool"
)

func main() {
	var pkg string
	flag.StringVar(&pkg, "p", "jni", "指定 Go package 名称")
	flag.Parse()

	fmt.Print(tool.GenerateCode(pkg))
}
