# BangumiBuddy

<div align="center">

<img src="webui/public/logo.png" alt="Logo" width="200">

**基于 Mikan 计划的智能追番管理系统**

一站式番剧订阅、下载、整理、通知解决方案

[功能特性](#核心功能) | [快速开始](#快速开始) | [详细功能说明](https://github.com/MangataL/BangumiBuddy/wiki/BangumiBuddy功能说明)

</div>

---

## 项目简介

BangumiBuddy 是一个自动化的番剧管理工具，通过 Mikan Project RSS 订阅实现番剧的自动追踪、下载和媒体库整理。支持与 qBittorrent 下载器集成，自动从 **TMDB** 获取元数据，并通过多种通知渠道推送更新状态。

**请不要在 B 站小红书等中国境内社交平台宣传该项目**

## 核心功能

### 📺 番剧订阅自动化

- 基于 Mikan Project RSS 的自动订阅
- 支持正则表达式过滤规则（包含/排除），并支持手动勾选 rss 订阅项处理/未处理
- 支持订阅多字幕组并设置媒体库优先级覆盖
- 自动集数识别与偏移设置
- 订阅日历视图，一目了然播出时间
- 自动停止订阅功能（完结后自动停止）

### 📥 下载器集成

- 支持集成 qBittorrent 下载器

### 📁 媒体库自动整理

- 支持硬链接和软链接两种转移模式实现保种
- 可自定义媒体库命名格式（番剧/剧场版）
- 自动转移字幕和音频文件
- 字幕文件扩展重命名（简繁体命名适配不同媒体软件）
- ASS 字幕字体自动子集化
- 媒体库 nfo 元数据支持自动补全信息——用于媒体文件入库时 TMDB 元数据不全从而后续需要手动刷新的场景

### 🎯 磁力任务管理

- 手动添加磁力链接下载，用于 BD 资源/剧场版下载
- 灵活的文件选择和入库管理
- 支持一键转移并重命名其他地方下载的字幕到资源保存目录
- 资源本身自带字体时会自动识别并用作当次子集化的字体库

### 🎬 元数据集成

- 集成 TMDB（The Movie Database）
- 自动获取海报、背景图、简介等信息
- 支持电视剧和电影元数据搜索

### 🔔 多平台通知

支持三种通知方式：

- **Telegram Bot** - 即时推送到 Telegram
- **Email** - 邮件通知
- **Bark** - iOS 推送通知

可配置的通知节点：

- 订阅更新（新资源发布）
- 下载完成
- 媒体库转移完成
- 错误提醒

### 🌐 现代化 Web 界面

- 基于 React 19 + TypeScript 构建
- 响应式设计，支持移动端访问
- 暗色/亮色主题自动切换
- 直观的可视化配置界面

## 快速开始

### 使用 Docker Compose 部署

1. **创建 `docker-compose.yml` 文件**

```yaml
version: "3"

services:
  bangumi-buddy:
    image: mangatal/bangumi-buddy:latest
    container_name: bangumi-buddy
    restart: unless-stopped
    environment:
      - TZ=Asia/Shanghai # 设置时区，可根据需要修改
    ports:
      - "6937:6937" # 映射端口，可自定义前面的端口号
    volumes:
      - ./config:/config # 配置文件目录，容器内挂在路径不可变
      - ./data:/data # 数据库和日志目录，容器内挂在路径不可变
      - /path/to/download:/download # 下载目录，可自定义（需与 qBittorrent 一致）
      - /path/to/media:/media # 媒体库目录，可自定义
      - ./xxx/fonts:/data/fonts # 字体库目录，用于字幕的子集化，可选配置，也可以直接复用/data的挂载点
```

2. **启动容器**

```bash
docker-compose up -d
```

3. **访问 Web 界面**

打开浏览器访问 `http://localhost:6937`

默认账户：

- 用户名：`admin`
- 密码：`admin123`

⚠️ **重要**: 首次登录后请立即修改密码！

### 路径挂载说明

| 容器路径      | 说明                                                 | 必需 |
| ------------- | ---------------------------------------------------- | ---- |
| `/config`     | 配置文件目录（config.yaml）                          | ✅   |
| `/data`       | 数据库和日志存储                                     | ✅   |
| `/download`   | 下载目录（需与 qBittorrent 一致）                    | ✅   |
| `/media`      | 媒体库输出目录(可选，按需可与下载目录挂载同一父路径) | ⬜   |
| `/data/fonts` | 字幕字体库目录（可选）                               | ⬜   |

**关键注意事项**：

- `download` 路径必须与 qBittorrent 容器挂载的下载路径**完全一致**
- 如果使用硬链接模式，下载目录和媒体库目录必须在同一文件系统上

### 配置 qBittorrent

在 Web 界面的「设置」→「下载器配置」中设置：

- **qBittorrent 地址**: `http://your-qbittorrent-host:8080`
- **用户名**: qBittorrent 的用户名
- **密码**: qBittorrent 的密码

更多配置说明请查看 [部署指南](docs/wiki/部署指南.md)

## 使用流程

1. **配置下载器** - 连接 qBittorrent
2. **配置 TMDB** - 设置 TMDB API Token（用于获取元数据）
3. **配置通知**（可选）- 设置 Telegram/Email/Bark
4. **配置转移设置** - 设置媒体库路径和命名格式
5. **添加订阅** - 从 蜜柑计划 获取 RSS 链接
6. **开始使用** - 系统将自动下载和整理番剧

## 许可证

本项目采用 [GNU General Public License v3.0](LICENSE) 许可证。

<div align="center">

如果这个项目对你有帮助，请给一个 ⭐ Star 支持一下！

</div>
