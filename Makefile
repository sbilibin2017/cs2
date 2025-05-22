build-parser:
	go build -o ./cmd/parser/parser ./cmd/parser

run-parser:
	./cmd/parser/parser -d "clickhouse://user:password@localhost:9000/db"

mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./internal/... -cover	

migrate:
	goose -dir ./migrations clickhouse "clickhouse://user:password@localhost:9000/db" up

docker-run:
	docker run --name metrics-clickhouse \
		-p 8123:8123 \
		-p 9000:9000 \
		-e CLICKHOUSE_USER=user \
		-e CLICKHOUSE_PASSWORD=password \
		-e CLICKHOUSE_DB=db \
		-d clickhouse/clickhouse-server:latest