package utils

import "fmt"

var (
	chineseDigitMap = map[rune]int{
		'零': 0,
		'〇': 0,
		'一': 1,
		'二': 2,
		'两': 2,
		'三': 3,
		'四': 4,
		'五': 5,
		'六': 6,
		'七': 7,
		'八': 8,
		'九': 9,
	}
	chineseUnitMap = map[rune]int{
		'十': 10,
		'百': 100,
		'千': 1000,
		'万': 10000,
		'亿': 100000000,
	}
)

// ChineseNumberToInt 将纯中文数字转换为整数，例如 "十二万三千四百五十六" -> 123456。
func ChineseNumberToInt(input string) (int, error) {
	if input == "" {
		return 0, fmt.Errorf("中文数字不能为空")
	}

	total := 0
	section := 0
	number := 0

	for _, r := range input {
		if digit, ok := chineseDigitMap[r]; ok {
			number = digit
			continue
		}

		unit, ok := chineseUnitMap[r]
		if !ok {
			return 0, fmt.Errorf("无法识别的中文数字字符: %q", r)
		}

		switch unit {
		case 10, 100, 1000:
			if number == 0 {
				number = 1
			}
			section += number * unit
			number = 0
		case 10000, 100000000:
			if number == 0 && section == 0 {
				section = 1
			}
			section += number
			total += section * unit
			section = 0
			number = 0
		}
	}

	return total + section + number, nil
}
