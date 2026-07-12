// Package handlers 执行器内置通用 handler：admin 建 BEAN 任务填 handler 名 + JSON 参数即调度。
package handlers

import "xxl-job-admin-go/pkg/executor"

// Register 注册所有内置通用 handler。
func Register(ex *executor.Executor) {
	ex.Register("httpCall", HTTPCall)
	ex.Register("shellExec", ShellExec)
	ex.Register("mysqlQuery", MySQLQuery)
	ex.Register("sendWebhook", SendWebhook)
}
