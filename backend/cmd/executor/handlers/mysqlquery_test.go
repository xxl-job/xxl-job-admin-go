package handlers

import (
	"context"
	"testing"

	"xxl-job-admin-go/pkg/executor"
)

func TestMySQLQueryParamValidation(t *testing.T) {
	jc := &executor.JobContext{LogID: 1}
	// 缺 dsn/sql
	code, _ := MySQLQuery(context.Background(), jc, `{}`)
	if code != 500 {
		t.Fatalf("缺 dsn/sql 应 500: got %d", code)
	}
	// 参数解析失败
	code2, _ := MySQLQuery(context.Background(), jc, `bad json`)
	if code2 != 500 {
		t.Fatalf("坏 json 应 500: got %d", code2)
	}
}
