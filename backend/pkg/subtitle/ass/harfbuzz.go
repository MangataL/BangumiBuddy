package ass

/*
#cgo pkg-config: harfbuzz harfbuzz-subset
#include <hb-subset.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func CreateSubfontData(fontData []byte, fontIndex int, codePoints []rune) ([]byte, error) {
	if len(fontData) == 0 {
		return nil, fmt.Errorf("字体数据为空")
	}
	if len(codePoints) == 0 {
		return nil, fmt.Errorf("码点列表为空")
	}

	// 1. 创建字体 blob
	cData := C.CBytes(fontData)
	defer C.free(cData)

	blob := C.hb_blob_create((*C.char)(cData), C.uint(len(fontData)), C.HB_MEMORY_MODE_READONLY, nil, nil)
	defer C.hb_blob_destroy(blob)

	// 2. 创建字体 face
	face := C.hb_face_create(blob, C.uint(fontIndex))
	defer C.hb_face_destroy(face)

	// 3. 创建 codepoint set
	cpSet := C.hb_set_create()
	defer C.hb_set_destroy(cpSet)
	for _, cp := range codePoints {
		C.hb_set_add(cpSet, C.uint(cp))
	}

	// 4. 创建 subset input
	input := C.hb_subset_input_create_or_fail()
	if input == nil {
		return nil, fmt.Errorf("创建子集输入失败")
	}
	defer C.hb_subset_input_destroy(input)

	inputCodepoints := C.hb_subset_input_set(input, C.HB_SUBSET_SETS_UNICODE)
	C.hb_set_union(inputCodepoints, cpSet)

	// 5. 子集化
	subsetFace := C.hb_subset_or_fail(face, input)
	if subsetFace == nil {
		return nil, fmt.Errorf("字体子集化失败")
	}
	defer C.hb_face_destroy(subsetFace)

	// 6. 获取子集数据
	subsetBlob := C.hb_face_reference_blob(subsetFace)
	defer C.hb_blob_destroy(subsetBlob)

	var length C.uint
	subsetData := C.hb_blob_get_data(subsetBlob, &length)
	if subsetData == nil || length == 0 {
		return nil, fmt.Errorf("获取子集化字体数据失败")
	}

	return C.GoBytes(unsafe.Pointer(subsetData), C.int(length)), nil
}
