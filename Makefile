IMG ?= bigkevmcd/peanut-helmpipelines

proto-format:
	@buf format -w ./api

generate:
	@buf generate

test:
	@go test -v ./...

docker-build:
	docker build -t ${IMG} .

docker-push:
	docker push ${IMG}
