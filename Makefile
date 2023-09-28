default: init build run

init:
	go mod init terrasync || true
	go mod tidy

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./terrasync

run:
	TERRASYNC_ROOT_WORKING_DIR=./terraform TERRASYNC_SYNC_TIME_SECONDS=7 ./terrasync

docker:
	docker buildx build --tag terrasync .
	docker run -v $$(pwd)/terraform:/terraform -p 8080:8080 -d terrasync
