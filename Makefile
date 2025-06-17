ifneq (,$(wildcard ./.env))
    include .env
    export
endif

BACKEND_DATA_PATH := $(BACKEND_PATH)/backend/data

prepare:
	@ssh $(SERVER_NAME) "mkdir -p $(BACKEND_PATH)/target"

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

target/blog-http: $(wildcard cmd/blog-http/*.go) $(shell find pkg/ -type f -name '*.go')
	CGO_ENABLED=0 go build -tags netgo -o target/blog-http ./cmd/blog-http

upload-http: target/blog-http
	rsync -vr target/blog-http $(FRONTEND_SERVER):~

jl-dev: $(wildcard cmd/blog-http/*.go) $(shell find pkg/ -type f -name '*.go')
	cd packages/jl && pnpm dev --port 11465

jl: $(wildcard cmd/blog-http/*.go) $(shell find pkg/ -type f -name '*.go')
	rsync -vr $(FRONTEND_SERVER):~/www/caddy/log/ packages/jl/.data/log/
	cd packages/jl && pnpm build && cd dist && python -m http.server -b 127.0.0.1 11465 

$(FRONTEND_PATH)/dist/articles.json: $(wildcard $(FRONTEND_PATH)/content/article/*.typ)
	cd $(FRONTEND_PATH) && pnpm build

.data/articles.json: $(FRONTEND_PATH)/dist/articles.json
	@echo "Copying articles.json..."
	@mkdir -p .data
	@cp $< $@
	@echo "Copy complete."

upload-data: .data/articles.json
	@echo "Uploading data to server..."
	@ssh $(SERVER_NAME) "mkdir -p $(BACKEND_DATA_PATH)/"
	@rsync -vr $^ $(SERVER_NAME):$(BACKEND_DATA_PATH)/
	@echo "Upload data complete."

download-data:
	@echo "Downloading data from server..."
	@rsync -vr $(SERVER_NAME):$(BACKEND_DATA_PATH)/article-stats.json $(SERVER_NAME):$(BACKEND_DATA_PATH)/article-comments.json $(SERVER_NAME):$(BACKEND_DATA_PATH)/article-email-comments.json .data
	@cp .data/article-stats.json $(FRONTEND_PATH)/content/snapshot/article-stats.json
	@cp .data/article-comments.json $(FRONTEND_PATH)/content/snapshot/article-comments.json
	@cp .data/article-email-comments.json packages/jl/.data/article-comments.json
	@echo "Downloading data complete."

sync: upload-data download-data
	@echo "Sync complete."

uploadA: scripts/reload-caddy.sh deployment/server/docker-compose.yaml
	rsync -vr deployment/server/docker-compose.yaml $(SERVER_NAME):$(BACKEND_PATH)/docker-compose.yaml
	rsync -vr scripts/reload-caddy.sh $(SERVER_NAME):$(BACKEND_PATH)/scripts/reload-caddy.sh

uploadB: target/blog-srv target/blog-cli
	rsync -vr $^ $(SERVER_NAME):$(BACKEND_PATH)/target/

upload: uploadA uploadB
	@echo "Upload complete."

reload-caddy:
	@echo "Reloading Caddy..."
	@ssh $(SERVER_NAME) "cd $(BACKEND_PATH) && ./scripts/reload-caddy.sh"
	@echo "Caddy reloaded."

deploy: upload
	@echo "Deploying to server..."
	@ssh $(SERVER_NAME) "cd $(BACKEND_PATH) && \
	  docker compose down blog-backend && docker compose up blog-backend -d"
	@echo "Deployment complete."

logs:
	@echo "Fetching logs from server..."
	@ssh $(SERVER_NAME) "cd $(BACKEND_PATH) && docker compose logs blog-backend -f" || true

login:
	@ssh $(SERVER_NAME) -t "cd $(BACKEND_PATH) && bash" || true

.PHONY: all clean sync download-data upload login deploy jl
