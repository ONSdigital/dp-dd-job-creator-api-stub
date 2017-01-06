build:
	go build -o build/dp-dd-job-creator-api-stub

debug: build
	HUMAN_LOG=1 ./build/dp-dd-job-creator-api-stub

.PHONY: build debug
