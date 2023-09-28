default: init docker

init:
	go mod init terrasync || true
	go mod tidy

docker:
	docker build --tag terrasync .
	docker run -v $$(pwd)/terraform:/terraform -d terrasync

clean:
	docker stop $$(docker ps -aq) 2> /dev/null || true
	docker rm $$(docker ps -aq) 2> /dev/null || true
