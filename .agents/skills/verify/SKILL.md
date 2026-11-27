---
name: verify
description: 在隔离的 BangumiBuddy 实例上验收 Web 操作、订阅、磁力转移和字体库。涉及 ASS 字幕子集化、字体匹配或媒体库命名时，用 libass 实际渲染入库字幕并保留像素对照。
---

# BangumiBuddy 验收

## Launch

从仓库根目录执行。需要项目的 Go、CGo、FreeType、HarfBuzz、Node 和 pnpm，以及 Python 3.11+、FFmpeg 和 libass 动态库。Python 验收依赖放在独立环境。下面的 shell 变量供本次终端使用。

```sh
make build-frontend
verify_env=$(mktemp -d)
python3 -m venv "$verify_env/venv"
. "$verify_env/venv/bin/activate"
python -m pip install fonttools pillow
export PYTHONDONTWRITEBYTECODE=1
driver=.agents/skills/verify/scripts/verify.py
proof="artifacts/verify/$(date +%Y%m%d-%H%M%S)-$$"
python "$driver" subtitles "$proof"
```

`ffmpeg` 默认从 PATH 查找，亦可用 `FFMPEG` 指定可执行文件。libass 可用系统包或独立环境中的版本；若加载器找不到，设置 `VERIFY_LIBASS` 为 `libass.dylib` 或 `libass.so` 的绝对路径。macOS 可用 Homebrew 的 `libass` 与 `ffmpeg`，也可通过临时环境提供同名依赖。FFmpeg 生成三秒的真实视频，字幕画面由 libass 渲染。

脚本编译当前 Go 源码、使用当前前端构建，从空 SQLite 创建测试实例。生产监听地址固定为 6937，验收构建仅通过 Go overlay 改为独立的回环端口，通过链接参数把 TMDB SDK 的地址指向本地样例服务。qBittorrent 也通过应用的正式配置指向样例 Web API。下载器与 TMDB 的响应由样例提供，鉴权、数据库、种子监控、字体库、子集化及硬链接转移均使用应用实现。各实例拥有自己的端口、下载目录、字体目录和媒体库。

要手动操作，在独占终端运行 `python "$driver" serve "$proof-manual"`。出现 `READY` 和 `SESSION` 表示前端可读、测试账号能登录、受保护 API 可读。只在打印的 URL 使用临时实例默认账号 `admin / admin123`。

## Doctor

将启动输出中的 `SESSION` 文件路径赋给 `session`。在同一 Python 环境的另一个终端执行以下命令。它只读检查该 PID、登录态、首页、字体库以及本次媒体库路径，并显示运行构建的 SHA-256 和提交号。

```sh
python .agents/skills/verify/scripts/verify.py doctor "$session"
```

## Drive

先读 [功能地图](features/README.md)。自动字幕验收从创建磁力任务进入；用详情 API 交叉核对最终入库路径。下载监控约每分钟同步一次，脚本等到文件列表实际落库再触发转移。交互验收可以在打印的 URL 操作 WebUI，也可用相同公开 API。下面读取字体库状态。

```sh
python .agents/skills/verify/scripts/verify.py api "$session" GET /subtitle/meta-sets/stats
```

## Evidence

- 记录用户动作、API 响应、最终文件和用户可见结果。涉及按钮、布局或动效时，保留前后截图或连续录屏；单靠 API 不能证明页面交互。
- 字幕验收以**媒体库里的最终 ASS** 为准。解码其实际附件，核对内部 PostScript 名、使用字符和子集大小。以独立完整字体为正对照，分别用去附件字幕、同名错字形的旧附件为负对照。关闭 libass 系统字体源，每次渲染使用新的字体环境。任一指定时刻画面错误，都应报失败。
- 每个时间点实际帧必须可见，与参考逐像素相同，并与两个负对照各有至少 100 个不同像素。这个下限只用于本仓库的固定样例。真实故障字幕要选择能区分预期字体和回退字体的台词和时间点。
- 每次使用新的证据目录。`actions.jsonl` 保留请求和状态，`library-files.json` 记录入库文件，`libass.log` 记录播放器选中的字体，`render.json` 记录附件和像素结果，`comparison.png` 和单帧 PNG 供目视复查。保留原字幕、入库字幕、可播放的三秒 MKV 和运行日志。将使用的样例边界与实际覆盖的入口写进验收结论。
- 若要复核已保存的文件，可用同一版本的仓库和 libass 独立重跑渲染。把重放结果、负对照和录屏都存入本次证据目录，供下次验收统一清理。

```sh
python .agents/skills/verify/scripts/render_proof.py \
  "$proof/S01E01.subset.zh-CN.forced.ass" "$proof/source-with-stale-fonts.ass" \
  backend/pkg/subtitle/ass/testdata "$proof/fallback.ttf" "$proof/replay" \
  --plain-final "$proof/S01E01.subset.commentary.ass"
```

要检验验收脚本本身，把第一项输入换成同名旧字体的源字幕。下面这次调用应以非零状态退出，且 `render.json` 中三帧对参考的像素差都大于零。若只是命令或文件缺失，同样属于验收失败。

```sh
python .agents/skills/verify/scripts/render_proof.py \
  "$proof/source-with-stale-fonts.ass" "$proof/source-with-stale-fonts.ass" \
  backend/pkg/subtitle/ass/testdata "$proof/fallback.ttf" "$proof/negative"
```

需要向人展示实际播放效果时，在证据目录中把同一份视频和字幕转成三秒的带字幕预览。此命令需 FFmpeg 支持 `ass` 滤镜和 `libx264`。视频供目视复核，像素判据仍使用关闭系统字体源的渲染。

```sh
(cd "$proof" && "${FFMPEG:-ffmpeg}" -nostdin -hide_banner -i S01E01.mkv \
  -vf 'ass=S01E01.subset.zh-CN.forced.ass' -an -c:v libx264 -pix_fmt yuv420p \
  -movflags +faststart playback.mp4 > playback.log 2>&1)
```

## Cleanup

自动验收在成功、断言失败或中断后，都会结束所启动的应用和样例 HTTP 服务并清除临时 SQLite、会话令牌及下载目录。手动实例就绪后用前台终端的 Ctrl-C 正常结束。退出后确认本次证据仍可读，并核对证据中的 PID 已退出。

Git 忽略仓库根目录的 `artifacts/`。每次运行 `subtitles` 或 `serve` 都在 `artifacts/verify/` 下新建一个直接子目录。成功结束后，脚本删除此前已结束的验收目录，只保留本次结果。失败或中断会保留上次成功结果和最新一次失败现场。并行运行按结束顺序清理。需要长期保存的证据可复制到 `artifacts/verify/` 之外。

## Helpers

`verify.py subtitles` 做完整验收，`serve` 保持独立实例，`doctor` 只读诊断，`api` 调用该实例的正式路由。`render_proof.py` 可重放已归档的字幕并返回进程状态。`subtitle_fixture.py` 是两个入口共用的样例模块。
