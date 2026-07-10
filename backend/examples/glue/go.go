//go:build ignore
// +build ignore

// GLUE-Go 任务示例（yaegi 在线解释执行）：GLUE 类型选「GLUE（Go 在线代码）」。
// 必须定义 Run 函数：func Run(param string) (int, string)
// build tag ignore：本文件仅为示例文本，不参与编译（无 main 函数）。
package main

import (
	"fmt"
	"time"
)

func Run(param string) (int, string) {
	fmt.Printf("[%s] GLUE Go 开始，参数: %s\n", time.Now().Format("15:04:05"), param)
	for i := 1; i <= 3; i++ {
		fmt.Printf("  step %d/3\n", i)
	}
	return 200, "glue-go ok: " + param
}
