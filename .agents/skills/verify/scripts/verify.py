#!/usr/bin/env python3
"""BangumiBuddy 验收入口。运行期间可通过 doctor 或 api 访问本次隔离实例。"""

import argparse
import hashlib
import json
import os
import shutil
import signal
import socket
import sqlite3
import subprocess
import sys
import tempfile
import threading
import time
from contextlib import ExitStack, closing
from http.server import ThreadingHTTPServer
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import ProxyHandler, Request, build_opener

from subtitle_fixture import (
    FONTS,
    HASH,
    SCRIPT,
    SECOND,
    SIDECAR,
    VIDEO,
    Boundaries,
    create_subtitles,
    install_fonts,
)

SCRIPTS = Path(__file__).resolve().parent
ROOT = SCRIPTS.parents[3]
FONTS_DIR = ROOT / "backend/pkg/subtitle/ass/testdata"
HTTP = build_opener(ProxyHandler({}))


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def request(url, method="GET", data=None, token="", form=False):
    headers = {"Authorization": "Bearer " + token} if token else {}
    body = None
    if data is not None:
        body = urlencode(data).encode() if form else json.dumps(data).encode()
        headers["Content-Type"] = "application/x-www-form-urlencoded" if form else "application/json"
    with HTTP.open(Request(url, body, headers, method=method), timeout=30) as response:
        raw = response.read()
        return response.status, json.loads(raw) if raw else None


def checked(ok, message):
    if not ok:
        raise ValueError(message)


class Session:
    def __init__(self, state):
        self.state = state
        self.url = state["url"] + "/apis/v1"
        self.evidence = Path(state["evidence"])

    def api(self, method, route, data=None, *, record=True):
        checked(route.startswith("/") and not route.startswith("//"), "请输入相对 API 路由")
        event = {"method": method, "route": route}
        if data is not None:
            event["request"] = data
        try:
            status, result = request(self.url + route, method, data, self.state["token"])
            event.update(status=status, response=result)
            return result
        except HTTPError as exc:
            event.update(status=exc.code, response=exc.read().decode())
            raise RuntimeError(f"{method} {route}: {event['response']}") from exc
        finally:
            if record:
                with (self.evidence / "actions.jsonl").open("a") as f:
                    f.write(json.dumps(event, ensure_ascii=False) + "\n")

    def doctor(self):
        os.kill(self.state["pid"], 0)
        config = self.api("GET", "/config/transfer", record=False)
        checked(config["tvPath"] == self.state["library"], "运行目录与服务配置不符")
        stats = self.api("GET", "/subtitle/meta-sets/stats", record=False)
        with HTTP.open(self.state["url"] + "/", timeout=3) as response:
            checked(b'id="root"' in response.read(), "前端资源未就绪")
        return {k: self.state[k] for k in ("pid", "url", "fixtures", "revision", "binarySha256")} | {
            "fonts": stats
        }


def start(scratch, evidence, peers, stack):
    for folder in ("fonts", "download", "library"):
        (scratch / folder).mkdir()
    install_fonts(FONTS_DIR, scratch / "fonts")
    backend = ROOT / "backend"
    checked((backend / "web/index.html").is_file(), "请先运行 make build-frontend")
    # 只覆写固定监听地址；业务代码和工作树都使用原版本。
    with socket.socket() as port_picker:
        port_picker.bind(("127.0.0.1", 0))
        port = port_picker.getsockname()[1]
    url = f"http://127.0.0.1:{port}"
    main = backend / "main.go"
    code = main.read_text()
    literal = 'r.Run("[::]:6937")'
    checked(code.count(literal) == 1, "监听入口已变化，请核对启动方式")
    patched = scratch / "listen.go"
    patched.write_text(code.replace(literal, f'r.Run("127.0.0.1:{port}")'))
    overlay = scratch / "overlay.json"
    write_json(overlay, {"Replace": {str(main): str(patched)}})
    tmdb = "-X github.com/cyruzin/golang-tmdb.baseURL=" + peers + "/3"
    binary = scratch / "BangumiBuddy"
    subprocess.run(
        ["go", "build", "-overlay", str(overlay), "-ldflags", tmdb, "-o", binary, "."],
        cwd=backend,
        check=True,
    )
    config = {
        "tmdb": {"tmdb_token": "local-fixture"},
        "download": {
            "downloader": {"download_type": "qbittorrent", "qbittorrent": {"host": peers}},
            "manager": {"tv_save_path": str(scratch / "download")},
        },
        "transfer": {
            "tv_path": str(scratch / "library"),
            "tv_format": "{name}/{season_episode}",
            "transfer_type": "hardlink",
            "enable_subtitle_subset": True,
            "interval": 1440,
        },
        "subtitle": {
            "generate_new_file": True,
            "cover_exist_sub_font": True,
            "use_similar_font": True,
            "check_glyphs": True,
            "use_system_fonts_dir": False,
        },
        "scraper": {"enable": False, "check_interval": 24},
        "subscriber": {"rss_check_interval": 30},
        "notice": {"enabled": False},
    }
    config_path = scratch / "config.yaml"
    write_json(config_path, config)  # JSON 是有效 YAML，路径由编码器转义。
    env = dict(
        os.environ,
        DB_PATH=str(scratch / "app.db"),
        CONFIG_FILE_PATH=str(config_path),
        FONT_DIR=str(scratch / "fonts"),
        LOG_FILE_PATH=str(evidence / "app.log"),
        DEV="false",
    )
    log = stack.enter_context((evidence / "server.log").open("w"))
    proc = subprocess.Popen([binary], cwd=scratch, env=env, stdout=log, stderr=subprocess.STDOUT)
    stack.callback(stop, proc)
    login = {"grant_type": "password", "username": "admin", "password": "admin123"}
    for _ in range(125):
        checked(proc.poll() is None, f"服务退出，请检查 {log.name}")
        try:
            _, credentials = request(url + "/apis/v1/token", "POST", login, form=True)
            token = credentials["access_token"]
            break
        except (URLError, OSError):
            time.sleep(0.2)
    else:
        raise TimeoutError("隔离实例登录未就绪")
    with binary.open("rb") as built:
        digest = hashlib.file_digest(built, "sha256").hexdigest()
    state = {
        "url": url,
        "fixtures": peers,
        "pid": proc.pid,
        "library": str(scratch / "library"),
        "token": token,
        "evidence": str(evidence),
        "binarySha256": digest,
        "revision": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip(),
    }
    statefile = scratch / "session.json"
    write_json(statefile, state)
    session = Session(state)
    write_json(evidence / "doctor.json", session.doctor())
    print(f"READY {url}\nSESSION {statefile}", flush=True)
    return session


def stop(proc):
    if proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=7)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait()


def prepare_download(scratch):
    download = scratch / "download"
    create_subtitles(scratch / "fonts", scratch)
    (download / SECOND).write_text(SCRIPT.split("Dialogue: 0,0:00:01.00")[0], encoding="utf-8")
    ffmpeg = os.environ.get("FFMPEG") or "ffmpeg"
    subprocess.run(
        [
            ffmpeg,
            "-hide_banner",
            "-loglevel",
            "error",
            "-f",
            "lavfi",
            "-i",
            "color=black:s=640x160:d=3",
            "-c:v",
            "ffv1",
            str(download / VIDEO),
        ],
        check=True,
    )


def subtitles(session, scratch):
    evidence = session.evidence
    download = scratch / "download"
    original = download / SIDECAR
    shutil.copyfile(original, evidence / "source-with-stale-fonts.ass")
    task = session.api("POST", "/magnets", {"type": "tv", "magnetLink": "magnet:?xt=urn:btih:" + HASH})
    detail = "/magnets/" + task["taskID"]
    torrent = task["torrent"]
    for f in torrent["files"]:
        if f["media"]:
            f.update(season=1, episode=1)
    session.api(
        "PUT",
        detail,
        {"tmdbID": 11, "releaseGroup": "Verify", "torrent": torrent, "continueDownload": True},
    )
    # 详情 API 会先显示下载器的实时状态，转移实际读取已持久化的文件列表。
    with closing(sqlite3.connect((scratch / "app.db").as_uri() + "?mode=ro", uri=True)) as db:
        for _ in range(80):
            row = db.execute(
                "SELECT status, file_names FROM torrents WHERE hash = ?", (HASH,)
            ).fetchone()
            if row and row[0] == "downloaded" and VIDEO in row[1]:
                break
            time.sleep(1)
        else:
            raise TimeoutError(f"下载监控尚未持久化种子文件列表: {row}")
    session.api("GET", detail)
    transfer = "/torrents/" + HASH + "/transfer"
    library = scratch / "library"
    dest = library / "Verify Show/S01E01"
    subset_file = dest.with_suffix(".subset.zh-CN.forced.ass")
    missing = (
        "缺少 3 款字体",
        f"方正隶变_GBK：{SECOND}、{SIDECAR}",
        f"多言語テスト：{SIDECAR}",
        f"多言語テスト（粗斜体）：{SIDECAR}",
    )
    try:
        session.api("POST", transfer)
    except RuntimeError:
        pass
    state = session.api("GET", detail)
    checked(
        state["downloadStatus"] == "transferredError"
        and all(s in state["downloadStatusDetail"] for s in missing),
        f"缺字体失败状态或汇总不完整: {state['downloadStatusDetail']}",
    )
    checked(not any(library.rglob("*")), "严格模式已写入媒体库")
    config = session.api("GET", "/config/transfer")
    config["ignoreSubsetError"] = True
    session.api("PUT", "/config/transfer", config)
    session.api("POST", transfer)
    state = session.api("GET", detail)
    checked(
        state["downloadStatus"] == "transferred"
        and all(s in state["downloadStatusDetail"] for s in missing),
        "忽略缺字体时未保留完整警告",
    )
    checked(
        dest.with_suffix(".zh-CN.forced.ass").read_bytes() == original.read_bytes()
        and not subset_file.exists(),
        "忽略缺字体时原字幕或子集副本异常",
    )
    session.api("POST", "/subtitle/meta-sets")
    stats = session.api("GET", "/subtitle/meta-sets/stats")
    checked(stats["total"] == len(FONTS) and not stats["needsRebuild"], "字体库重建结果异常")
    config["ignoreSubsetError"] = False
    session.api("PUT", "/config/transfer", config)
    session.api("POST", transfer)
    state = session.api("GET", detail)
    checked(
        state["downloadStatus"] == "transferred" and not state["downloadStatusDetail"], "重试仍有警告"
    )
    checked(
        any(
            f["linkFile"] == str(dest.with_suffix(".mkv"))
            for f in state["torrent"]["files"]
            if f["media"]
        ),
        "任务记录未指向预期入库文件",
    )
    checked(dest.with_suffix(".mkv").samefile(download / VIDEO), "视频未按硬链接入库")
    checked(
        subset_file.samefile(download / "Show.01.subset.zh-CN.forced.ass"), "入库字幕来源或后缀错误"
    )
    checked(
        original.read_bytes() == (evidence / "source-with-stale-fonts.ass").read_bytes(), "原字幕被覆盖"
    )
    plain_subset = dest.with_suffix(".subset.commentary.ass")
    checked(
        plain_subset.samefile(download / "Show.01.subset.commentary.ass"), "无附件字幕的入库路径错误"
    )
    shutil.copyfile(plain_subset, evidence / plain_subset.name)
    shutil.copyfile(dest.with_suffix(".mkv"), evidence / "S01E01.mkv")
    shutil.copyfile(subset_file, evidence / subset_file.name)
    shutil.copyfile(scratch / "fallback.ttf", evidence / "fallback.ttf")
    shutil.copyfile(scratch / "fonts" / FONTS[2], evidence / FONTS[2])
    write_json(
        evidence / "library-files.json",
        sorted(str(p.relative_to(library)) for p in library.rglob("*") if p.is_file()),
    )
    with (evidence / "libass.log").open("w") as log:
        subprocess.run(
            [
                sys.executable,
                SCRIPTS / "render_proof.py",
                subset_file,
                original,
                scratch / "fonts",
                scratch / "fallback.ttf",
                evidence,
                "--plain-final",
                plain_subset,
            ],
            stderr=log,
            check=True,
        )


def run(mode, path):
    evidence = path.resolve()
    evidence.mkdir(parents=True, exist_ok=False)
    with tempfile.TemporaryDirectory(prefix="bangumibuddy-verify-") as tmp, ExitStack() as stack:
        scratch = Path(tmp).resolve()
        server = stack.enter_context(ThreadingHTTPServer(("127.0.0.1", 0), Boundaries))
        server.added = server.started = False
        threading.Thread(target=server.serve_forever, daemon=True).start()
        stack.callback(server.shutdown)
        session = start(scratch, evidence, f"http://127.0.0.1:{server.server_port}", stack)
        prepare_download(scratch)
        if mode == "subtitles":
            subtitles(session, scratch)
        else:
            print("Ctrl-C 结束本次实例", flush=True)
            threading.Event().wait()
    print(f"PASS {evidence}", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="mode", required=True)
    for cmd in ("subtitles", "serve"):
        commands.add_parser(cmd).add_argument("evidence", type=Path)
    commands.add_parser("doctor").add_argument("session", type=Path)
    api = commands.add_parser("api")
    api.add_argument("session", type=Path)
    api.add_argument("method", choices=("GET", "POST", "PUT", "DELETE"))
    api.add_argument("route")
    api.add_argument("json", nargs="?")
    args = parser.parse_args()
    if args.mode in ("subtitles", "serve"):
        signal.signal(signal.SIGTERM, lambda *_: sys.exit(143))
        try:
            run(args.mode, args.evidence)
        except KeyboardInterrupt:
            sys.exit(130)
    else:
        session = Session(json.loads(args.session.read_text()))
        status = session.doctor()
        result = (
            status
            if args.mode == "doctor"
            else session.api(args.method, args.route, json.loads(args.json) if args.json else None)
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))
