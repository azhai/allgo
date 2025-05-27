package main

import (
	"fmt"
)

// GenModelCmd 生成Model文件
type GenModelCmd struct {
}

// Run 生成Model文件
func (c *GenModelCmd) Run() {
	fmt.Println("gen model")
}
