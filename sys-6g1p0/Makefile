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
	@./sys-6g1p0

##### CLEAN
clean: env-down
	@echo "Clean up ..."
	@rm -rf /tmp/sys-6g1p0-* sys-6g1p0
	@docker rm -f -v `docker ps -a --no-trunc | grep "sys-6g1p0" | cut -d ' ' -f 1` 2>/dev/null || true
	@docker rmi `docker images --no-trunc | grep "sys-6g1p0" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Clean up done"
