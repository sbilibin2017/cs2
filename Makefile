mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./... -cover	

migrate:
	goose -dir ./migrations clickhouse "clickhouse://user:password@localhost:9000/db?secure=false" up

docker-run:
	docker run --name metrics-clickhouse \
		-e CLICKHOUSE_USER=user \
		-e CLICKHOUSE_PASSWORD=password \
		-e CLICKHOUSE_DB=db \
		-p 9000:9000 \
		-d clickhouse/clickhouse-server:latest
