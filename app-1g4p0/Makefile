.PHONY: all dev clean build env-up env-down run

all: clean build env-up run

dev: build run

##### BUILD
build:
	@echo "Build ..."
	@go build
	@echo "Build done"


##### RUN
run:
	@echo "Start app ..."
	@./app-1g4p0

##### CLEAN
clean: env-down
	@echo "Clean up ..."
	@rm -rf /tmp/app-1g4p0-* app-1g4p0
	@docker rm -f -v `docker ps -a --no-trunc | grep "app-1g4p0" | cut -d ' ' -f 1` 2>/dev/null || true
	@docker rmi `docker images --no-trunc | grep "app-1g4p0" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Clean up done"
