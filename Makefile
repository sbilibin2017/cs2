run-parser:
	go run ./cmd/parser/main.go 

run-traintestsplitter:
	go run ./cmd/traintestsplitter/main.go 

run-lableencoders:
	go run ./cmd/lableencoders/main.go 

run-datasetmaker:
	go run ./cmd/datasetmaker/main.go 	

mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./internal/... -cover	

lint:
	staticcheck ./internal/... 

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