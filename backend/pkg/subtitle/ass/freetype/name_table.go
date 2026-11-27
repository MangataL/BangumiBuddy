package freetype

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

const (
	// Platform IDs
	PlatformUnicode   = 0 // OpenType 标准
	PlatformMacintosh = 1
	PlatformWindows   = 3

	// Language IDs for Windows Platform
	LanguageEnglishUS          = 0x0409 // 美国英语
	LanguageChineseSimplified  = 0x0804 // 简体中文
	LanguageChineseTraditional = 0x0404 // 繁体中文

	// Name IDs
	NameIDFamily     = 1
	NameIDFullName   = 4
	NameIDPostScript = 6
)

type FontName struct {
	Kind  uint16
	Value string
}

type FontNames struct {
	Aliases               []FontName
	FamilyName            string
	FullName              string
	PostScriptName        string
	ChineseFamilyName     string
	ChineseFullName       string
	ChinesePostScriptName string
}

// ExtractFontNames preserves localized names for matching alongside the display names.
func ExtractFontNames(face *Face) FontNames {
	var names FontNames
	seen := make(map[FontName]bool)
	defaults := make(map[uint16]string)
	chinese := make(map[uint16]map[uint16]string)
	fallback := make(map[uint16]string)
	for i, count := 0, face.GetSfntNameCount(); i < count; i++ {
		record, err := face.GetSfntName(i)
		if err != nil {
			continue
		}
		kind := record.NameID
		if kind != NameIDFamily && kind != NameIDFullName && kind != NameIDPostScript {
			continue
		}
		value := decodeName(record)
		if value == "" {
			continue
		}
		alias := FontName{Kind: kind, Value: value}
		if !seen[alias] {
			names.Aliases = append(names.Aliases, alias)
			seen[alias] = true
		}
		if fallback[kind] == "" {
			fallback[kind] = value
		}
		if defaults[kind] == "" && (record.PlatformID == PlatformUnicode || (record.PlatformID == PlatformWindows && record.LanguageID == LanguageEnglishUS)) {
			defaults[kind] = value
		}
		if record.PlatformID == PlatformWindows && (record.LanguageID == LanguageChineseSimplified || record.LanguageID == LanguageChineseTraditional) {
			if chinese[kind] == nil {
				chinese[kind] = make(map[uint16]string)
			}
			chinese[kind][record.LanguageID] = value
		}
	}
	for kind, value := range fallback {
		if defaults[kind] == "" {
			defaults[kind] = value
		}
	}
	names.FamilyName, names.FullName, names.PostScriptName = defaults[NameIDFamily], defaults[NameIDFullName], defaults[NameIDPostScript]
	names.ChineseFamilyName = selectChineseName(chinese[NameIDFamily])
	names.ChineseFullName = selectChineseName(chinese[NameIDFullName])
	names.ChinesePostScriptName = selectChineseName(chinese[NameIDPostScript])
	return names
}

func decodeName(r *NameRecord) string {
	var name string
	switch {
	case r.PlatformID == PlatformUnicode, r.PlatformID == PlatformWindows && (r.EncodingID == 0 || r.EncodingID == 1 || r.EncodingID == 10):
		name = decodeUTF16BE([]byte(r.Value))
	default:
		enc := legacyEncoding(r)
		if enc == nil {
			return ""
		}
		name, _ = enc.NewDecoder().String(r.Value)
	}
	if strings.ContainsRune(name, '�') {
		return ""
	}
	return strings.TrimSpace(strings.Trim(name, "\x00"))
}

// 一些旧字体只在非 Unicode 的记录里声明本地名称。
func legacyEncoding(r *NameRecord) encoding.Encoding {
	switch r.PlatformID {
	case PlatformWindows:
		switch r.EncodingID {
		case 2:
			return japanese.ShiftJIS
		case 3:
			return simplifiedchinese.GBK
		case 4:
			return traditionalchinese.Big5
		case 5:
			return korean.EUCKR
		}
	case PlatformMacintosh:
		switch r.EncodingID {
		case 0:
			// 这些语言的 Roman 记录使用其它 Mac 代码页。
			switch r.LanguageID {
			case 15, 17, 18, 24, 25, 26, 27, 28, 36, 37, 38, 39, 40:
				return nil
			}
			return charmap.Macintosh
		case 1:
			return japanese.ShiftJIS
		case 2:
			return traditionalchinese.Big5
		case 3:
			return korean.EUCKR
		case 7:
			return charmap.MacintoshCyrillic
		case 25:
			return simplifiedchinese.GBK
		}
	}
	return nil
}

// selectChineseName 选择中文名称（仅限 Windows 平台），优先简体中文，其次繁体中文
func selectChineseName(names map[uint16]string) string {
	if names == nil {
		return ""
	}

	// 优先返回简体中文（Windows 平台，Language ID 0x0804）
	if simplified, ok := names[LanguageChineseSimplified]; ok && simplified != "" {
		return simplified
	}

	// 其次返回繁体中文（Windows 平台，Language ID 0x0404）
	if traditional, ok := names[LanguageChineseTraditional]; ok && traditional != "" {
		return traditional
	}

	return ""
}

// decodeUTF16BE 解码 UTF-16 Big Endian 字符串
func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		// UTF-16 必须是偶数字节
		return ""
	}

	// 转换为 uint16 切片
	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}

	// 解码为 UTF-8 字符串
	runes := utf16.Decode(u16s)
	return string(runes)
}
