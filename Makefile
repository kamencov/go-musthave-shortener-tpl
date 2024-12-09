.PHONY: cover
cover:
	go test -short -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

.PHONY: test
test:
	go test ./... -coverprofile=cover.out
	grep -Ev "pb.go|_mock.go|docs.go|main.go" cover.out > cover.filtered.out || true
	go tool cover -func=cover.filtered.out
	rm cover.out cover.filtered.out

.PHONY: covernomock
covernomock:
	go test -coverprofile=coverage.out ./...
	grep -v "mock.go|main.go|docs.go|alias.go|pb.go|noexit.go|revive.go" coverage.out > coverage.filtered.out
	go tool cover -html=coverage.filtered.out
	rm coverage.out coverage.filtered.out

.PHONY: protorun
protorun:
	protoc --go_out=./internal/proto --go_opt=paths=source_relative \
	--go-grpc_out=./internal/proto --go-grpc_opt=paths=source_relative \
	./proto/shortener.proto


.PHONY: git_push
git_push:
	git add .
	git commit -m "update"
	git push origin iter25