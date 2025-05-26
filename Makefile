ifneq (,$(wildcard ./.env))
    include .env
    export
endif


ServerDockerfilePath ?= deployment/server/Dockerfile
ServerServicePath ?= deployment/server/docker-compose.yaml

target/blog-srv: $(wildcard cmd/blog-srv/*.go) $(shell find pkg/ -type f -name '*.go')
	@echo "Building blog-srv..."
	@mkdir -p target
	@go build -buildvcs -o target/blog-srv ./cmd/blog-srv/
	@echo "Build complete."

target/blog-cli: $(wildcard cmd/blog-cli/*.go) $(shell find pkg/ -type f -name '*.go')
	@echo "Building blog-cli..."
	@mkdir -p target
	@go build -buildvcs -o target/blog-cli ./cmd/blog-cli/
	@echo "Build complete."

$(BLOG_FRONTEND_PATH)/dist/articles.json: $(wildcard $(BLOG_FRONTEND_PATH)/content/article/*.typ)
	cd $(BLOG_FRONTEND_PATH) && pnpm build

.data/articles.json: $(BLOG_FRONTEND_PATH)/dist/articles.json
	@echo "Copying articles.json..."
	@mkdir -p .data
	@cp $< $@
	@echo "Copy complete."

upload-data: .data/articles.json
	@echo "Uploading data to server..."
	@ssh $(SERVER_NAME) "mkdir -p $(BLOG_PATH)/.data"
	@scp $^ $(SERVER_NAME):$(BLOG_PATH)/.data
	@echo "Upload data complete."

download-data:
	@echo "Downloading data from server..."
	@scp $(SERVER_NAME):$(BLOG_PATH)/.data/article-stats.json $(SERVER_NAME):$(BLOG_PATH)/.data/article-comments.json .data
	@cp .data/article-stats.json $(BLOG_FRONTEND_PATH)/content/snapshot/article-stats.json
	@cp .data/article-comments.json $(BLOG_FRONTEND_PATH)/content/snapshot/article-comments.json
	@echo "Downloading data complete."

sync: upload-data download-data
	@echo "Sync complete."

upload: target/blog-srv target/blog-cli
	@echo "Uploading to server..."
	@ssh $(SERVER_NAME) "mkdir -p $(BLOG_PATH)/target"
	@scp $^ $(SERVER_NAME):$(BLOG_PATH)/target/
	@echo "Upload complete."

deploy: upload
	@echo "Deploying to server..."
	@scp $(ServerDockerfilePath) $(ServerServicePath) $(SERVER_NAME):$(BLOG_PATH) && \
	  ssh $(SERVER_NAME) "cd $(BLOG_PATH) && docker build -t blog-srv . && \
	  docker-compose down && docker-compose up -d"
	@echo "Deployment complete."

logs:
	@echo "Fetching logs from server..."
	@ssh $(SERVER_NAME) "cd $(BLOG_PATH) && docker-compose logs -f" || true

login:
	@ssh $(SERVER_NAME) -t "cd $(BLOG_PATH) && bash" || true

.PHONY: all clean sync download-data upload login deploy
