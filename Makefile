.PHONY: all build-frontend build-backend clean

all: build-frontend build-backend
	@echo "构建完成！"
	@echo "可执行文件在: ./backend/BangumiBuddy"

build-frontend:
	@echo "构建前端..."
	cd webui && pnpm install && pnpm build
	@mkdir -p backend/web
	@echo "复制前端资源到后端..."
	@cp -r webui/dist/* backend/web/

build-backend:
	@echo "构建后端..."
	cd backend && go build -o BangumiBuddy

clean:
	@echo "清理构建产物..."
	@rm -rf backend/web
	@rm -f backend/BangumiBuddy 

docker:
	@echo "构建Docker镜像..."
	docker-compose up -d --build
	