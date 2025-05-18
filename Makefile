ifneq (,$(wildcard ./.env))
    include .env
    export
endif

target/blog-srv: cmd/blog-srv/main.go $(wildcard pkg/dto/*.go)
	@echo "Building blog-srv..."
	@mkdir -p target
	@go build -o target/blog-srv cmd/blog-srv/main.go
	@echo "Build complete."

target/blog-cli: cmd/blog-cli/main.go $(wildcard pkg/dto/*.go)
	@echo "Building blog-cli..."
	@mkdir -p target
	@go build -o target/blog-cli cmd/blog-cli/main.go
	@echo "Build complete."

upload-data: .data/articles.json
	@echo "Uploading data to server..."
	@ssh $(SERVER_NAME) "mkdir -p $(BLOG_PATH)/.data"
	@scp $^ $(SERVER_NAME):$(BLOG_PATH)/.data
	@echo "Upload data complete."

download-data:
	@echo "Downloading data from server..."
	@ssh $(SERVER_NAME) "cd $(BLOG_PATH) && target/blog-cli"
	@scp $(SERVER_NAME):$(BLOG_PATH)/.data/article-clicks.json .data
	@cp .data/article-clicks.json $(BLOG_FRONTEND_PATH)/assets/article-clicks.json
	@echo "Downloading data complete."

upload: target/blog-srv target/blog-cli
	@echo "Uploading to server..."
	@ssh $(SERVER_NAME) "mkdir -p $(BLOG_PATH)/target"
	@scp $^ $(SERVER_NAME):$(BLOG_PATH)/target/
	@echo "Upload complete."

login:
	@ssh $(SERVER_NAME) -t "cd $(BLOG_PATH) && bash" || true

.PHONY: all clean upload-data upload login
