.PHONY: build
build:
	docker build -t poncheska/sa-data-collector -f builds/Dockerfile .
	docker push poncheska/sa-data-collector

.PHONY: run
run:
	go run ./main.go
