# 磁力任务与转移

## Sub-features

添加磁力任务、确认媒体和集号、下载状态同步、手动转移、失败后重试，以及任务详情中的文件路径与警告。

## How to get to it (user POV)

进入 `/download` 点「添加磁力任务」，确认下载文件；下载完成后可从任务列表或详情点击「转移文件」。

## Driving it with verify.py

`verify.py subtitles` 从 `POST /magnets` 创建任务，使用返回的任务 ID 调用 `PUT /magnets/{id}` 确认开始下载。下载器同步后，通过 `POST /torrents/{hash}/transfer` 转移，通过 `GET /magnets/{id}` 重读状态和实际链接。确认入库视频与源视频指向同一文件，原字幕和生成的轨道均有正确的文件名。重试成功后警告应消失。对照 `actions.jsonl` 和 `library-files.json` 核查。

## Gotchas

详情可以先显示下载器的实时完成状态，后台稍后才把完整文件列表写入 SQLite。驱动只读等待这个前置条件。样例下载器提供本地种子状态与文件列表。真实 qBittorrent 鉴权、BT 下载、通知服务送达或列表与详情按钮是否响应，需分别用相应的外部实例或浏览器验收。
