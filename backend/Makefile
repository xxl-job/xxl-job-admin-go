.PHONY: tidy build run proto lint vet test up down clean

## 依赖：整理 go.mod
tidy:
	go mod tidy

## 构建：编译二进制到 bin/
build:
	go build -o bin/xxl-job-admin-go ./cmd/server

## 运行：本地启动 admin
run:
	go run ./cmd/server

## Proto：buf lint + 生成代码
proto:
	buf lint
	buf generate

## Lint：golangci-lint（与 CI 一致）
lint:
	golangci-lint run

## Vet：go vet 快速检查
vet:
	go vet ./...

## 测试
test:
	go test ./... -race -count=1

## 容器：启动 mysql + admin（compose 文件在 deploy/）
up:
	docker compose -f ../deploy/docker-compose.yml up -d

## 容器：停止
down:
	docker compose -f ../deploy/docker-compose.yml down

## 清理
clean:
	rm -rf bin/ dist/
