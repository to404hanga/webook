.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
	@./build.bat
	@docker rmi -f to404hanga/webook:v0.0.1
	@docker build -t to404hanga/webook:v0.0.1