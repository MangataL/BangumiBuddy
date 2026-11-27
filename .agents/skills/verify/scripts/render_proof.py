#!/usr/bin/env python3
"""直接渲染入库的 ASS，并保存参考、回退、旧附件三组对照。"""

import argparse
import ctypes as c
import ctypes.util
import hashlib
import io
import json
import os
import re
from pathlib import Path

from fontTools.ttLib import TTFont
from PIL import Image, ImageChops, ImageDraw
from subtitle_fixture import CASES, FONTS, SCRIPT


class Layer(c.Structure):
    pass


Layer._fields_ = [
    ("w", c.c_int),
    ("h", c.c_int),
    ("stride", c.c_int),
    ("bitmap", c.c_void_p),
    ("color", c.c_uint32),
    ("x", c.c_int),
    ("y", c.c_int),
    ("next", c.POINTER(Layer)),
]


def load_libass():
    lib = c.CDLL(os.environ.get("VERIFY_LIBASS") or ctypes.util.find_library("ass") or "libass.so")
    ptr, integer, path = c.c_void_p, c.c_int, c.c_char_p
    signatures = [
        ("library_version", integer, []),
        ("library_init", ptr, []),
        ("library_done", None, [ptr]),
        ("set_extract_fonts", None, [ptr, integer]),
        ("renderer_init", ptr, [ptr]),
        ("renderer_done", None, [ptr]),
        ("set_frame_size", None, [ptr, integer, integer]),
        ("set_storage_size", None, [ptr, integer, integer]),
        ("read_file", ptr, [ptr, path, path]),
        ("free_track", None, [ptr]),
        ("set_fonts", None, [ptr, path, path, integer, path, integer]),
        ("render_frame", c.POINTER(Layer), [ptr, ptr, c.c_longlong, c.POINTER(integer)]),
    ]
    for name, result, args in signatures:
        fn = getattr(lib, "ass_" + name)
        fn.restype, fn.argtypes = result, args
    return lib


def render(lib, subtitle, default_face, millis):
    # 每次用全新的 library，禁止访问系统字体（ASS_FONTPROVIDER_NONE = 0）。
    library = lib.ass_library_init()
    if not library:
        raise RuntimeError("libass 初始化失败")
    lib.ass_set_extract_fonts(library, 1)
    track = lib.ass_read_file(library, os.fsencode(subtitle), b"UTF-8")
    renderer = lib.ass_renderer_init(library)
    try:
        if not track or not renderer:
            raise RuntimeError(f"libass 无法加载 {subtitle}")
        width, height = 640, 160
        lib.ass_set_storage_size(renderer, width, height)
        lib.ass_set_frame_size(renderer, width, height)
        lib.ass_set_fonts(renderer, os.fsencode(default_face), None, 0, None, 0)
        layer = lib.ass_render_frame(renderer, track, millis, None)
        frame = Image.new("L", (width, height))
        while layer:
            node = layer.contents
            if node.w and node.h:
                # libass 最后一行只保证有 w 个字节。
                length = node.stride * (node.h - 1) + node.w
                mask = Image.frombytes(
                    "L", (node.w, node.h), c.string_at(node.bitmap, length), "raw", "L", node.stride, 1
                )
                opacity = 255 - (node.color & 255)
                if opacity != 255:
                    mask = mask.point([a * opacity // 255 for a in range(256)])
                # 样例只使用灰阶颜色。
                frame.paste(node.color >> 24, (node.x, node.y, node.x + node.w, node.y + node.h), mask)
            layer = node.next
        return frame
    finally:
        if track:
            lib.ass_free_track(track)
        if renderer:
            lib.ass_renderer_done(renderer)
        lib.ass_library_done(library)


def unpack_ass(text):
    """独立读取最终附件，不调用应用编码器。"""
    sections = re.split(r"(?m)(?=^\[[^\]\r\n]+\]\s*$)", text)
    attachments = {}
    kept = []
    for section in sections:
        if not section.lower().startswith("[fonts]"):
            kept.append(section)
            continue
        for record in section.split("fontname: ")[1:]:
            name, *lines = record.splitlines()
            values = [ord(ch) - 33 for ch in "".join(lines)]
            if not values or len(values) % 4 == 1 or any(v < 0 or v > 63 for v in values):
                raise ValueError(f"无效的 ASS 附件 {name}")
            data = bytearray()
            for i in range(0, len(values), 4):
                group = values[i : i + 4]
                bits = sum(v << (18 - j * 6) for j, v in enumerate(group))
                data.extend(bits.to_bytes(3, "big")[: len(group) - 1])
            if name in attachments:
                raise ValueError(f"附件名称重复 {name}")
            attachments[name] = bytes(data)
    return "".join(kept).strip(), attachments


def different_pixels(a, b):
    return a.width * a.height - ImageChops.difference(a, b).histogram()[0]


def prove(final, stale, fonts, fallback, evidence, plain_final=None):
    evidence.mkdir(parents=True, exist_ok=True)
    lib = load_libass()
    plain, attachments = unpack_ass(final.read_text(encoding="utf-8-sig"))
    if plain != SCRIPT.strip():
        raise ValueError("转移后字幕的样式、时间或对白发生变化")
    stripped = evidence / "without-attachments.ass"
    stripped.write_text(plain + "\n", encoding="utf-8")
    info = {}
    for name, data in attachments.items():
        with TTFont(io.BytesIO(data)) as face:
            ps = face["name"].getDebugName(6)
            if ps in info:
                raise ValueError(f"重复字体 {ps}")
            info[ps] = {
                "attachment": name,
                "bytes": len(data),
                "sha256": hashlib.sha256(data).hexdigest(),
                "codepoints": sorted(face.getBestCmap()),
            }
    report = {
        "libass": hex(lib.ass_library_version()),
        "systemFonts": False,
        "attachments": info,
        "frames": {},
    }
    panels = []
    expected_names = set()
    for (label, ms, chars), font_name in zip(CASES, FONTS):
        source = fonts / font_name
        if font_name == FONTS[2] and not source.exists():
            source = stale.parent / font_name
        with TTFont(source) as face:
            ps = face["name"].getDebugName(6)
        expected_names.add(ps)
        if ps not in info or not set(map(ord, chars)).issubset(info[ps]["codepoints"]):
            raise ValueError(f"最终附件不含预期字体或字形 {ps} {chars}")
        if label == "chinese-style" and info[ps]["bytes"] >= source.stat().st_size:
            raise ValueError("大字体未经子集化")
        frames = {}
        for case, script, default_face in (
            ("reference", stripped, source),
            ("actual", final, fallback),
            ("fallback", stripped, fallback),
            ("stale", stale, fallback),
        ):
            frame = render(lib, script, default_face, ms)
            frame.save(evidence / f"{label}-{case}.png")
            frames[case] = frame
            panels.append((f"{label} | {ms} ms | {case}", frame))
        actual = frames["actual"]
        stats = {
            case: different_pixels(actual, frame) for case, frame in frames.items() if case != "actual"
        }
        stats["visiblePixels"] = sum(actual.histogram()[1:])
        stats["fallbackVisible"] = sum(frames["fallback"].histogram()[1:])
        stats["staleVisible"] = sum(frames["stale"].histogram()[1:])
        report["frames"][label] = stats
    if plain_final is not None:
        plain_text, plain_fonts = unpack_ass(plain_final.read_text(encoding="utf-8-sig"))
        if plain_text != SCRIPT.split("Dialogue: 0,0:00:01.00")[0].strip() or len(plain_fonts) != 1:
            raise ValueError("从无附件字幕生成的子集内容不完整")
        sample = render(lib, plain_final, fallback, CASES[0][1])
        sample.save(evidence / "plain-actual.png")
        with Image.open(evidence / "chinese-style-reference.png") as reference:
            report["plainSubtitleDifferentPixels"] = different_pixels(sample, reference)
    report["finalAssSha256"] = hashlib.sha256(final.read_bytes()).hexdigest()
    output = Image.new("L", (2 * 640, len(panels) // 2 * 184), 42)
    draw = ImageDraw.Draw(output)
    for i, (label, frame) in enumerate(panels):
        x, y = (i % 2) * 640, (i // 2) * 184
        draw.text((x + 12, y + 5), label, fill=255)
        output.paste(frame, (x, y + 24))
    output.save(evidence / "comparison.png")
    (evidence / "render.json").write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n")
    if report.get("plainSubtitleDifferentPixels", 0) != 0:
        raise ValueError("无附件字幕生成后的渲染与指定字体不符")
    if set(info) != expected_names:
        raise ValueError("附件包含意外字体")
    for label, stats in report["frames"].items():
        if (
            stats["reference"] != 0
            or min(
                stats["visiblePixels"],
                stats["fallbackVisible"],
                stats["staleVisible"],
                stats["fallback"],
                stats["stale"],
            )
            < 100
        ):
            raise ValueError(f"{label} 渲染验收失败: {stats}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    for arg in ("final", "stale", "fonts", "fallback", "evidence"):
        parser.add_argument(arg, type=Path)
    parser.add_argument("--plain-final", type=Path, help="同时验收从无附件原字幕生成的第二条字幕轨")
    args = parser.parse_args()
    prove(**vars(args))
