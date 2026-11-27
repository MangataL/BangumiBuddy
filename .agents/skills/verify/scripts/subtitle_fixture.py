"""验收用字幕、过期字体附件与外部服务响应。只写入本次运行的临时目录。"""

import base64
import hashlib
import json
import shutil
import textwrap
from http.server import BaseHTTPRequestHandler
from io import BytesIO
from urllib.parse import urlsplit

from fontTools import subset
from fontTools.pens.ttGlyphPen import TTGlyphPen
from fontTools.ttLib import TTFont

FONTS = ("方正隶变_GBK.ttf", "multilingual.ttf", "verify-bold-italic.ttf")
CASES = (
    ("chinese-style", 500, "春日山川"),
    ("japanese-override", 1500, "AAA"),
    ("bold-italic", 2500, "AAA"),
)
VIDEO = "Show.01.mkv"
SIDECAR = "Show.01.zh-CN.forced.ass"
SECOND = "Show.01.commentary.ass"
TITLE = "Verify Show"
HASH = hashlib.sha1(b"BangumiBuddy isolated subtitle fixture").hexdigest()
SCRIPT = r"""[Script Info]
ScriptType: v4.00+
PlayResX: 640
PlayResY: 160
[V4+ Styles]
Format: Name,Fontname,Fontsize,PrimaryColour,SecondaryColour,OutlineColour,BackColour,Bold,Italic,Underline,StrikeOut,ScaleX,ScaleY,Spacing,Angle,BorderStyle,Outline,Shadow,Alignment,MarginL,MarginR,MarginV,Encoding
Style: Default,方正隶变_GBK,64,&H00FFFFFF,&H00FFFFFF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,0,0,5,0,0,0,1
[Events]
Format: Layer,Start,End,Style,Name,MarginL,MarginR,MarginV,Effect,Text
Dialogue: 0,0:00:00.00,0:00:01.00,Default,,0,0,0,,春日山川
Dialogue: 0,0:00:01.00,0:00:02.00,Default,,0,0,0,,{\fn多言語テスト}AAA
Dialogue: 0,0:00:02.00,0:00:03.00,Default,,0,0,0,,{\fn多言語テスト\b1\i1}AAA
"""


def control_font(source, chars, points):
    """保留字体名字，用不同轮廓让「沿用旧附件」和「回退」可见。"""
    font = TTFont(source)
    options = subset.Options()
    options.name_IDs = ["*"]
    options.name_languages = ["*"]
    options.name_legacy = True
    keep = subset.Subsetter(options)
    keep.populate(text=chars)
    keep.subset(font)
    em = font["head"].unitsPerEm
    for glyph in set(font.getBestCmap().values()):
        pen = TTGlyphPen(None)
        outline = [(round(x * em), round(y * em)) for x, y in points]
        pen.moveTo(outline[0])
        for point in outline[1:]:
            pen.lineTo(point)
        pen.closePath()
        font["glyf"][glyph] = pen.glyph()
    data = BytesIO()
    font.save(data)
    return data.getvalue()


def ass_attachment(name, data):
    # ASS 附件是无填充的 6 位编码，字符范围 33..96。
    alphabet = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
    encoded = (
        base64.b64encode(data).rstrip(b"=").translate(bytes.maketrans(alphabet, bytes(range(33, 97))))
    )
    return (
        "fontname: "
        + name
        + "\n"
        + "\n".join(textwrap.wrap(encoded.decode(), 80, break_on_hyphens=False))
    )


def install_fonts(source, dest):
    for name in FONTS[:2]:
        shutil.copyfile(source / name, dest / name)
    # 同一家族的物理粗斜体有独立轮廓，常规体加合成倾斜也无法通过对照。
    data = control_font(source / FONTS[1], "A", [(0.2, 0.1), (0.6, 0.1), (0.9, 0.9), (0.5, 0.9)])
    font = TTFont(BytesIO(data))
    for record in font["name"].names:
        value = {2: "Bold Italic", 4: "Test Multi Bold Italic", 6: "TestMulti-BoldItalic"}.get(
            record.nameID
        )
        if value:
            record.string = value.encode(record.getEncoding())
    font["OS/2"].usWeightClass = 700
    font["OS/2"].fsSelection = 0x21
    font["head"].macStyle = 3
    font["post"].italicAngle = -12
    font.save(dest / FONTS[2])


def create_subtitles(font_dir, scratch):
    old = []
    for i, ((_, _, chars), name) in enumerate(zip(CASES, FONTS)):
        data = control_font(font_dir / name, chars, [(0.1, 0.1), (0.8, 0.5), (0.1, 0.9)])
        old.append(ass_attachment(f"stale{i}.ttf", data))
    (scratch / "fallback.ttf").write_bytes(
        control_font(
            font_dir / FONTS[0], "春日山川A", [(0.1, 0.1), (0.8, 0.1), (0.8, 0.35), (0.1, 0.35)]
        )
    )
    (scratch / "download" / SIDECAR).write_text(
        SCRIPT + "\n[Fonts]\n" + "\n".join(old) + "\n", encoding="utf-8"
    )


class Boundaries(BaseHTTPRequestHandler):
    """仅模拟真实 HTTP 边界。文件解析、任务状态、字体库和转移由应用处理。"""

    def log_message(self, fmt, *args):
        pass

    def reply(self, value, status=200, mime=None):
        raw = value.encode() if isinstance(value, str) else json.dumps(value).encode()
        self.send_response(status)
        self.send_header(
            "Content-Type",
            mime or ("text/plain; charset=utf-8" if isinstance(value, str) else "application/json"),
        )
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def do_GET(self):
        path = urlsplit(self.path).path
        if path in ("/rss/1", "/rss/2"):
            season = int(path[-1])
            date = ("2020-01-02", "2022-07-02")[season - 1]
            self.reply(
                f"""<rss version="2.0"><channel><title>Mikan Project - Verify Show</title>
                <item><guid isPermaLink="false">[Verify] Verify Show - 01</guid>
                <enclosure url="http://127.0.0.1/fixture.torrent" type="application/x-bittorrent" length="100"/>
                <torrent><pubDate>{date}T00:15:00</pubDate></torrent></item></channel></rss>""",
                mime="application/rss+xml",
            )
            return
        torrent = {
            "hash": HASH,
            "name": "[Verify] Verify Show - 01",
            "size": 100,
            "amount_left": 0 if self.server.started else 100,
            "state": "uploading" if self.server.started else "stoppedDL",
        }
        resources = {
            "/api/v2/torrents/tags": ["BangumiBuddy"],
            "/api/v2/torrents/info": [torrent] if self.server.added else [],
            "/api/v2/torrents/properties": torrent,
            "/api/v2/torrents/files": [
                {"index": i, "name": n, "size": 50, "priority": 1}
                for i, n in enumerate((VIDEO, SIDECAR, SECOND))
            ],
            "/api/v2/app/webapiVersion": "2.11.0",
            "/api/v2/app/version": "v5.0.5",
            "/3/search/tv": {"page": 1, "results": [{"id": 11, "name": TITLE, "genre_ids": [16]}]},
            "/3/tv/11": {
                "id": 11,
                "name": TITLE,
                "first_air_date": "2020-01-01",
                "seasons": [
                    {"season_number": 1, "air_date": "2020-01-01", "episode_count": 12},
                    {"season_number": 2, "air_date": "2022-07-01", "episode_count": 10},
                ],
            },
        }
        if path in resources:
            self.reply(resources[path])
        else:
            self.reply(path, 404)

    def do_POST(self):
        path = urlsplit(self.path).path
        self.rfile.read(int(self.headers.get("Content-Length", 0)))
        if path == "/api/v2/torrents/add":
            self.server.added = True
        elif path in ("/api/v2/torrents/resume", "/api/v2/torrents/start"):
            self.server.started = True
        elif path != "/api/v2/torrents/filePrio":
            self.reply(path, 404)
            return
        self.reply("Ok.")
