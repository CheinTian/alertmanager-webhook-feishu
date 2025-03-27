fmt:
	go fmt ./...
run:fmt
	go run main.go server -c config.yml -v
test:
	go test -v ./...
build:
	goreleaser release --snapshot
docker_build:
	docker build -t lolspider/alertmanager-webhook-feishu .
docker_push:docker_build
	docker push lolspider/alertmanager-webhook-feishu
