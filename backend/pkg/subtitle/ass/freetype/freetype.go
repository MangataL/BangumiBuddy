package freetype

/*
#cgo pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H
#include FT_SFNT_NAMES_H
#include FT_TRUETYPE_IDS_H
#include FT_TRUETYPE_TABLES_H
*/
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	StyleFlagItalic = 1 << 0
	StyleFlagBold   = 1 << 1
)

// Library FreeType库句柄
type Library struct {
	lib C.FT_Library
}

// Face 字体Face句柄
type Face struct {
	face C.FT_Face
}

// NewLibrary 初始化FreeType库
func NewLibrary() (*Library, error) {
	var lib C.FT_Library
	err := C.FT_Init_FreeType(&lib)
	if err != 0 {
		return nil, fmt.Errorf("初始化FreeType失败: %d", err)
	}
	return &Library{lib: lib}, nil
}

// Done 释放FreeType库
func (l *Library) Done() {
	if l.lib != nil {
		C.FT_Done_FreeType(l.lib)
		l.lib = nil
	}
}

// NewFaceFromFile 从文件加载字体
func (l *Library) NewFaceFromFile(path string, index int) (*Face, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var face C.FT_Face
	err := C.FT_New_Face(l.lib, cPath, C.FT_Long(index), &face)
	if err != 0 {
		return nil, fmt.Errorf("加载字体失败: %d", err)
	}

	return &Face{face: face}, nil
}

// NewFaceFromMemory 从内存加载字体
func (l *Library) NewFaceFromMemory(data []byte, index int) (*Face, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("字体数据为空")
	}

	var face C.FT_Face
	err := C.FT_New_Memory_Face(
		l.lib,
		(*C.FT_Byte)(unsafe.Pointer(&data[0])),
		C.FT_Long(len(data)),
		C.FT_Long(index),
		&face,
	)
	if err != 0 {
		return nil, fmt.Errorf("从内存加载字体失败: %d", err)
	}

	return &Face{face: face}, nil
}

// Done 释放字体Face
func (f *Face) Done() {
	if f.face != nil {
		C.FT_Done_Face(f.face)
		f.face = nil
	}
}

// NumFaces 获取字体集合中的字体数量
func (f *Face) NumFaces() int {
	return int(f.face.num_faces)
}

// FaceIndex 获取当前字体在集合中的索引
func (f *Face) FaceIndex() int {
	return int(f.face.face_index)
}

// StyleFlags 返回 FreeType style_flags 位掩码
func (f *Face) StyleFlags() uint {
	if f.face == nil {
		return 0
	}
	return uint(f.face.style_flags)
}

// GetCharIndex 获取字符的字形索引
func (f *Face) GetCharIndex(charCode rune) uint {
	return uint(C.FT_Get_Char_Index(f.face, C.FT_ULong(charCode)))
}

// NameRecord Name Table中的一条记录
type NameRecord struct {
	PlatformID uint16
	EncodingID uint16
	LanguageID uint16
	NameID     uint16
	Value      string
}

// GetSfntNameCount 获取Name Table记录数
func (f *Face) GetSfntNameCount() int {
	return int(C.FT_Get_Sfnt_Name_Count(f.face))
}

// GetSfntName 获取指定索引的Name Table记录
func (f *Face) GetSfntName(index int) (*NameRecord, error) {
	var aname C.FT_SfntName
	err := C.FT_Get_Sfnt_Name(f.face, C.FT_UInt(index), &aname)
	if err != 0 {
		return nil, fmt.Errorf("获取Name记录失败: %d", err)
	}

	// 提取字符串数据
	stringData := C.GoBytes(unsafe.Pointer(aname.string), C.int(aname.string_len))

	record := &NameRecord{
		PlatformID: uint16(aname.platform_id),
		EncodingID: uint16(aname.encoding_id),
		LanguageID: uint16(aname.language_id),
		NameID:     uint16(aname.name_id),
		Value:      string(stringData),
	}

	return record, nil
}

// GetOS2Table 获取OS/2表信息（字重和倾斜度）
func (f *Face) GetOS2Table() (weight int, italic bool, err error) {
	// 获取OS/2表
	os2Ptr := C.FT_Get_Sfnt_Table(f.face, C.FT_SFNT_OS2)
	if os2Ptr == nil {
		return 400, false, fmt.Errorf("无法获取OS/2表")
	}

	// 将 void* 转换为 TT_OS2* 类型
	os2 := (*C.TT_OS2)(os2Ptr)

	// 读取 usWeightClass（字重）
	weight = int(os2.usWeightClass)

	// 读取 fsSelection（样式标志）
	// fsSelection bit 0: ITALIC, bit 9: OBLIQUE
	fsSelection := uint16(os2.fsSelection)
	italic = (fsSelection&0x01 != 0) || (fsSelection&0x200 != 0)

	return weight, italic, nil
}
