package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"xxl-job-admin-go/pkg/executor"
)

type mysqlQueryParam struct {
	DSN string `json:"dsn"`
	SQL string `json:"sql"`
}

// MySQLQuery 执行 MySQL 查询，回传结果行（JSON 数组）。
func MySQLQuery(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
	var p mysqlQueryParam
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return 500, "参数解析失败: " + err.Error()
	}
	if p.DSN == "" || p.SQL == "" {
		return 500, "dsn 和 sql 必填"
	}
	db, err := sql.Open("mysql", p.DSN)
	if err != nil {
		return 500, "打开连接失败: " + err.Error()
	}
	defer func() { _ = db.Close() }()
	rows, err := db.QueryContext(ctx, p.SQL)
	if err != nil {
		return 500, "查询失败: " + err.Error()
	}
	defer func() { _ = rows.Close() }()
	cols, _ := rows.Columns()
	var result []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return 500, "扫描行失败: " + err.Error()
		}
		m := map[string]any{}
		for i, c := range cols {
			m[c] = vals[i]
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return 500, "迭代失败: " + err.Error()
	}
	b, _ := json.Marshal(result)
	jc.Log("查询返回 " + strconv.Itoa(len(result)) + " 行")
	return 200, string(b)
}
