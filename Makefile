.PHONY: up down build logs seed lint test proto clean

up:
	docker-compose up -d

down:
	docker-compose down

build:
	docker-compose build

logs:
	docker-compose logs -f

seed:
	@echo "Seed test data here"

lint:
	cd server && golangci-lint run ./...

test:
	cd server && go test -v -race ./...

test-cover:
	cd server && go test -race -coverprofile=cover_all.out $$(go list ./... | grep -v /mocks/ | grep -v /proto/ | grep -v /cmd/) && grep -v "mocks\." cover_all.out > cover.out && go tool cover -func=cover.out | grep total

test-integration:
	cd server && go test -v -race -tags=integration ./...

proto:
	mkdir -p server/pkg/proto/account server/pkg/proto/transaction
	protoc --go_out=server/pkg/proto/account --go_opt=paths=source_relative \
		--go-grpc_out=server/pkg/proto/account --go-grpc_opt=paths=source_relative \
		server/proto/account.proto
	protoc --go_out=server/pkg/proto/transaction --go_opt=paths=source_relative \
		--go-grpc_out=server/pkg/proto/transaction --go-grpc_opt=paths=source_relative \
		server/proto/transaction.proto

clean:
	docker-compose down -v --remove-orphans
