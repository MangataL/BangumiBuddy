package freetype

import (
	"encoding/binary"
	"unicode/utf16"
)

const (
	// Platform IDs
	PlatformUnicode   = 0 // OpenType 标准
	PlatformMacintosh = 1
	PlatformWindows   = 3

	// Encoding IDs for Unicode Platform
	EncodingUnicode10   = 0 // Unicode 1.0
	EncodingUnicode11   = 1 // Unicode 1.1
	EncodingUnicode20   = 3 // Unicode 2.0 BMP
	EncodingUnicodeFull = 4 // Unicode 2.0 full

	// Encoding IDs for Windows Platform
	EncodingWindowsUnicodeBMP = 1

	// Language IDs for Windows Platform
	LanguageEnglishUS          = 0x0409 // 美国英语
	LanguageChineseSimplified  = 0x0804 // 简体中文
	LanguageChineseTraditional = 0x0404 // 繁体中文

	// Name IDs
	NameIDFamily     = 1
	NameIDFullName   = 4
	NameIDPostScript = 6
)

// FontNames 字体名称集合
type FontNames struct {
	FamilyName            string
	FullName              string
	PostScriptName        string
	ChineseFamilyName     string
	ChineseFullName       string
	ChinesePostScriptName string
}

// ExtractFontNames 从字体Face提取所有名称（英文和中文）
// 兼容 OpenType 标准，支持 Unicode Platform (0) 和 Windows Platform (3)
func ExtractFontNames(face *Face) FontNames {
	names := FontNames{}

	count := face.GetSfntNameCount()

	// 用于存储中文名称的临时变量（简体和繁体）
	chineseNames := make(map[uint16]map[uint16]string) // nameID -> languageID -> value

	for i := 0; i < count; i++ {
		record, err := face.GetSfntName(i)
		if err != nil {
			continue
		}

		var decoded string

		// 处理 Unicode 平台（OpenType 标准）
		// Unicode 平台是语言中性的，仅用于补充英文名称
		if record.PlatformID == PlatformUnicode {
			// Unicode 平台支持多种编码
			if record.EncodingID == EncodingUnicode10 ||
				record.EncodingID == EncodingUnicode11 ||
				record.EncodingID == EncodingUnicode20 ||
				record.EncodingID == EncodingUnicodeFull {
				decoded = decodeUTF16BE([]byte(record.Value))
			} else {
				continue
			}

			// Unicode 平台本身不区分语言，仅用于补充英文名称字段
			switch record.NameID {
			case NameIDFamily:
				if names.FamilyName == "" {
					names.FamilyName = decoded
				}
			case NameIDFullName:
				if names.FullName == "" {
					names.FullName = decoded
				}
			case NameIDPostScript:
				if names.PostScriptName == "" {
					names.PostScriptName = decoded
				}
			}
		} else if record.PlatformID == PlatformWindows && record.EncodingID == EncodingWindowsUnicodeBMP {
			// 处理 Windows 平台的 Unicode BMP 编码
			decoded = decodeUTF16BE([]byte(record.Value))

			// 处理英文名称（Language ID 0x0409 = 美国英语）
			if record.LanguageID == LanguageEnglishUS {
				switch record.NameID {
				case NameIDFamily:
					if names.FamilyName == "" {
						names.FamilyName = decoded
					}
				case NameIDFullName:
					if names.FullName == "" {
						names.FullName = decoded
					}
				case NameIDPostScript:
					if names.PostScriptName == "" {
						names.PostScriptName = decoded
					}
				}
			}

			// 处理中文名称（简体和繁体）
			if record.LanguageID == LanguageChineseSimplified || record.LanguageID == LanguageChineseTraditional {
				if chineseNames[record.NameID] == nil {
					chineseNames[record.NameID] = make(map[uint16]string)
				}
				chineseNames[record.NameID][record.LanguageID] = decoded
			}
		}
	}

	// 选择中文名称：优先简体中文，其次繁体中文（仅限 Windows 平台）
	names.ChineseFamilyName = selectChineseName(chineseNames[NameIDFamily])
	names.ChineseFullName = selectChineseName(chineseNames[NameIDFullName])
	names.ChinesePostScriptName = selectChineseName(chineseNames[NameIDPostScript])

	return names
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
