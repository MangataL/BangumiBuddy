# BangumiBuddy

基于 mikan 计划的管理追番软件

# 部署说明

## 通过 docker-compose 进行部署

```yaml
version: "3"

services:
  bangumi-buddy:
    image: mangatal/bangumi-buddy:latest
    container_name: bangumi-buddy
    restart: unless-stopped
    ports:
      - "6937:6937" # 前者替换为你自定义的端口或自行改为host模式
    volumes:
      - ./config:/config # ./config替换为你本地的路径
      - ./data:/data # ./data 替换为存放配置的本地路径
      - ~/bangumi/download:/download # 以下的可以自行挂载决定下载和媒体库存放在哪，可以分别设置也可以设置共同的父目录。注意需要与QBittorrent下载器挂载的路径相同，不然媒体库转移会出错
      - ~/bangumi/media:/video
```
