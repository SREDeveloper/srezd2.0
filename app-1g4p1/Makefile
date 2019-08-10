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
	@./app-1g4p1

##### CLEAN
clean: env-down
	@echo "Clean up ..."
	@rm -rf /tmp/app-1g4p1-* app-1g4p1
	@docker rm -f -v `docker ps -a --no-trunc | grep "app-1g4p1" | cut -d ' ' -f 1` 2>/dev/null || true
	@docker rmi `docker images --no-trunc | grep "app-1g4p1" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Clean up done"
