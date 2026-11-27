# 字体库

## Sub-features

空库、初始化、多语言名称、同一家族按样式选字型；升级旧库时对受影响用户的一次性提示。

## How to get to it (user POV)

进入 `/settings` → 工具设置 → 字体库，点「初始化」。已有旧索引的用户登录后会收到重建提示。

## Driving it with verify.py

在手动实例读取空库状态，建库，再从正式接口重新查询。三个字型来自本次运行的私有字体目录。

```sh
python "$driver" api "$session" GET /subtitle/meta-sets/stats
python "$driver" api "$session" POST /subtitle/meta-sets
python "$driver" api "$session" GET '/subtitle/meta-sets/fonts?page=1&page_size=10'
python "$driver" api "$session" GET /subtitle/meta-sets/stats
```

确认总数从 0 变成 3、初始化完成，且无需重建。日文别名与粗斜体是否可用，继续跑[实际渲染](subtitles.md)。

## Gotchas

空库无法验证升级提醒。该旅程需要一份由旧版本建立的非空索引、可重新扫描的原字体和独立浏览器配置。保留首次提示、同一浏览器再次打开页面的结果、重建后的统计；另用空库验证不会误提示。
