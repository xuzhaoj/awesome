.PHONY: docker
docker:
	@rm -f webook || true
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f flycash/webook-live:v0.0.1 || true
	@docker build -t flycash/webook-live:v0.0.1 .
