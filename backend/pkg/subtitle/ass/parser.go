package ass

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFontSet(ctx context.Context, filePath string, options ...ParseFontSetOption) (FontSet, bool, error) {
	opts := &parseFontSetOptions{
		IgnoreFontExist: false,
	}
	for _, opt := range options {
		opt(opts)
	}

	assSections, err := p.parserSection(filePath)
	if err != nil {
		return nil, false, err
	}

	if opts.IgnoreFontExist && len(assSections.Fonts) != 0 {
		return nil, true, nil
	}

	styles, err := p.parseStyles(assSections.Styles)
	if err != nil {
		return nil, false, err
	}
	dialogues, err := p.parseDialogues(assSections.Events)
	if err != nil {
		return nil, false, err
	}

	fontSet := p.buildFontSet(styles, dialogues)
	return fontSet, false, nil
}

type assSections struct {
	ScriptInfo []LineContent
	Styles     []LineContent
	Events     []LineContent
	Fonts      []LineContent
	Graphics   []LineContent
}

type LineContent struct {
	LineNum    int
	RawContent string
}

func (p *Parser) parserSection(filePath string) (assSections, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return assSections{}, err
	}
	defer file.Close()

	sections := assSections{
		ScriptInfo: make([]LineContent, 0),
		Styles:     make([]LineContent, 0),
		Events:     make([]LineContent, 0),
		Fonts:      make([]LineContent, 0),
		Graphics:   make([]LineContent, 0),
	}

	scanner := newBufferScanner(file)
	lineNum := 0
	currentSection := ""

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 检查是否是节标题
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// 根据当前节分类内容
		lineContent := LineContent{
			LineNum:    lineNum,
			RawContent: line,
		}

		switch currentSection {
		case "script info":
			sections.ScriptInfo = append(sections.ScriptInfo, lineContent)
		case "v4+ styles", "v4 styles":
			sections.Styles = append(sections.Styles, lineContent)
		case "events":
			sections.Events = append(sections.Events, lineContent)
		case "fonts":
			sections.Fonts = append(sections.Fonts, lineContent)
		case "graphics":
			sections.Graphics = append(sections.Graphics, lineContent)
		}
	}

	if err := scanner.Err(); err != nil {
		return assSections{}, err
	}

	return sections, nil
}

type Style struct {
	Name     string
	FontName string
	Bold     int
	Italic   bool
}

type Field struct {
	FieldName string
	Index     int
}

type Styles struct {
	styles       map[string]Style
	defaultStyle *Style
}

func (s *Styles) GetStyle(name string) (Style, bool) {
	style, ok := s.styles[name]
	if !ok {
		if s.defaultStyle != nil {
			return *s.defaultStyle, true
		}
		return Style{}, false
	}
	return style, true
}

func (p *Parser) parseStyles(lines []LineContent) (Styles, error) {
	styles := Styles{
		styles:       make(map[string]Style),
		defaultStyle: nil,
	}

	formatFields, index, err := p.parseFormatLine(lines, "Styles")
	if err != nil {
		return styles, err
	}

	// 解析每个 Style 行
	for _, line := range lines[index:] {
		if strings.HasPrefix(strings.ToLower(line.RawContent), "style:") {
			values, err := p.parseLineValues(line, "Styles", "Style", 0)
			if err != nil {
				return styles, err
			}
			getFieldValue := p.createFieldGetter(formatFields, values, line.LineNum, "Styles")
			name, err := getFieldValue("Name")
			if err != nil {
				return styles, err
			}
			fontName, err := getFieldValue("Fontname")
			if err != nil {
				return styles, err
			}
			fontName = strings.TrimPrefix(fontName, "@")
			bold, err := getFieldValue("Bold")
			if err != nil {
				return styles, err
			}
			italic, err := getFieldValue("Italic")
			if err != nil {
				return styles, err
			}
			style := Style{
				Name:     name,
				FontName: fontName,
				Bold:     p.parseBold(bold),
				Italic:   p.parseItalic(italic),
			}
			styles.styles[name] = style
			if strings.ToLower(name) == "default" {
				styles.defaultStyle = &style
			}
		}
	}

	return styles, nil
}

func (p *Parser) parseFormatLine(lines []LineContent, section string) (map[string]Field, int, error) {
	var (
		formatFields = make(map[string]Field)
		index        int
	)
	for i, line := range lines {
		if startWithIgnoreCase(line.RawContent, "format:") {
			fieldNames, err := p.parseLineValues(line, section, "Format", 0)
			if err != nil {
				return nil, 0, err
			}
			for i, fileName := range fieldNames {
				formatFields[strings.ToLower(fileName)] = Field{
					FieldName: fileName,
					Index:     i,
				}
			}
			index = i
			break
		}
	}
	return formatFields, index, nil
}

func startWithIgnoreCase(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), prefix)
}

func (p *Parser) parseLineValues(lc LineContent, section, key string, valuesMaxNum int) ([]string, error) {
	parts := strings.SplitN(lc.RawContent, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%s Section中第%d行的%s行格式不正确", section, lc.LineNum, key)
	}
	var values []string
	if valuesMaxNum > 0 {
		values = strings.SplitN(parts[1], ",", valuesMaxNum)
	} else {
		values = strings.Split(parts[1], ",")
	}
	for i, value := range values {
		values[i] = strings.TrimSpace(value)
	}
	return values, nil
}

// parseBold 解析粗体值，ASS 格式中 -1 表示 true，0 表示 false
func (p *Parser) parseBold(value string) int {
	weight, err := strconv.Atoi(value)
	if err != nil {
		return WeightNormal
	}
	return p.parseBoldWeight(weight)
}

func (p *Parser) parseBoldWeight(weight int) int {
	if weight == -1 || weight == 1 {
		return WeightBold
	}
	if weight == 0 {
		return WeightNormal
	}
	return weight
}

// parseItalic 解析斜体值，ASS 格式中 -1 和 1 表示 true，0 表示 false
func (p *Parser) parseItalic(value string) bool {
	return value == "-1" || value == "1"
}

// createFieldGetter 创建字段获取器闭包
func (p *Parser) createFieldGetter(formatFields map[string]Field, values []string, lineNum int, section string) func(fieldName string) (string, error) {
	return func(fieldName string) (string, error) {
		field, ok := formatFields[strings.ToLower(fieldName)]
		if !ok || field.Index >= len(values) {
			return "", fmt.Errorf("%s Section中第%d行的值格式错误，找不到%s字段", section, lineNum, fieldName)
		}
		return values[field.Index], nil
	}
}

func (p *Parser) parseDialogues(lines []LineContent) ([]Dialogue, error) {
	events := make([]Dialogue, 0)

	formatFields, index, err := p.parseFormatLine(lines, "Events")
	if err != nil {
		return nil, err
	}

	// 解析每个 Dialogue 行
	for _, line := range lines[index:] {
		if strings.HasPrefix(strings.ToLower(line.RawContent), "dialogue:") {
			values, err := p.parseLineValues(line, "Events", "Dialogue", len(formatFields))
			if err != nil {
				return nil, err
			}

			getFieldValue := p.createFieldGetter(formatFields, values, line.LineNum, "Events")

			style, err := getFieldValue("Style")
			if err != nil {
				return nil, err
			}

			text, err := getFieldValue("Text")
			if err != nil {
				return nil, err
			}

			event := Dialogue{
				Style: style,
				Text:  text,
			}
			events = append(events, event)
		}
	}

	return events, nil
}

func (p *Parser) buildFontSet(styles Styles, dialogues []Dialogue) FontSet {
	fontSet := make(FontSet)

	for _, dialogue := range dialogues {
		// 获取对话的样式
		style, ok := styles.GetStyle(dialogue.Style)
		if !ok {
			continue
		}

		// 初始化当前的字体属性
		currentFont := Font{
			FontName:   style.FontName,
			BoldWeight: style.Bold,
			Italic:     style.Italic,
		}

		// 保存样式的初始值，用于恢复
		styleBold := style.Bold
		styleItalic := style.Italic
		styleFontName := style.FontName

		// 遍历对话文本
		text := dialogue.Text
		runes := []rune(text) // 转换为 rune 切片以正确处理 UTF-8
		i := 0
		for i < len(runes) {
			// 检查是否是覆写代码块 {}
			if runes[i] == '{' {
				// 查找对应的右括号
				end := i + 1
				for end < len(runes) && runes[end] != '}' {
					end++
				}
				if end < len(runes) {
					// 解析覆写代码块
					overrideBlock := string(runes[i+1 : end])
					p.parseOverrideCodes(overrideBlock, &currentFont, overrideValue{
						BoldWeight: styleBold,
						Italic:     styleItalic,
						FontName:   styleFontName,
					})
					i = end + 1
					continue
				}
			}

			// 检查是否是不带大括号的覆写代码 \h, \n, \N
			if runes[i] == '\\' && i+1 < len(runes) {
				next := runes[i+1]
				if next == 'h' || next == 'n' || next == 'N' {
					// 跳过这些特殊字符
					i += 2
					continue
				}
			}

			// 普通字符，添加到字体集合中
			r := runes[i]

			// 将字符添加到对应的字体集合中
			if _, exists := fontSet[currentFont]; !exists {
				fontSet[currentFont] = make(CodePoints)
			}
			fontSet[currentFont][r] = struct{}{}
			i++
		}
	}

	return fontSet
}

type overrideValue struct {
	BoldWeight int
	Italic     bool
	FontName   string
}

// parseOverrideCodes 解析覆写代码块中的粗体、斜体和字体名称属性
func (p *Parser) parseOverrideCodes(overrideBlock string, currentFont *Font, defaultStyle overrideValue) {
	i := 0
	for i < len(overrideBlock) {
		if overrideBlock[i] == '\\' {
			// 提取完整的覆写代码（tag + param 作为一个整体）
			overrideTag, nextPos := p.extractOverrideTag(overrideBlock, i)
			i = nextPos

			// 判断是否符合字体覆盖语法
			if fontName, ok := p.isFontNameOverrideCommand(overrideTag); ok {
				p.applyFontNameOverride(currentFont, fontName, defaultStyle.FontName)
				continue
			}

			// 判断是否符合粗体覆盖语法
			if weight, cancel, ok := p.isBoldOverrideCommand(overrideTag); ok {
				p.applyBoldOverride(currentFont, cancel, weight, defaultStyle.BoldWeight)
				continue
			}

			// 判断是否符合斜体覆盖语法
			if italicValue, cancel, ok := p.isItalicOverrideCommand(overrideTag); ok {
				p.applyItalicOverride(currentFont, cancel, italicValue, defaultStyle.Italic)
				continue
			}
			// 其他覆写代码忽略
		} else {
			i++
		}
	}
}

// extractOverrideTag 提取完整的覆写代码（包括标签和参数作为一个整体）
// 返回: 完整的覆写代码, 下一个位置
func (p *Parser) extractOverrideTag(overrideBlock string, startPos int) (overrideTag string, nextPos int) {
	if startPos >= len(overrideBlock) || overrideBlock[startPos] != '\\' {
		return "", startPos + 1
	}

	i := startPos + 1 // 跳过 \
	tagStart := i

	// 提取整个覆写代码（直到遇到下一个 \ 或到达末尾）
	for i < len(overrideBlock) && overrideBlock[i] != '\\' {
		i++
	}

	overrideTag = overrideBlock[tagStart:i]
	return overrideTag, i
}

// isFontNameOverrideCommand 检查是否是字体名称覆盖指令
// 如果是 fn 开头，返回字体名和 true
func (p *Parser) isFontNameOverrideCommand(overrideTag string) (string, bool) {
	if strings.HasPrefix(overrideTag, "fn") {
		return overrideTag[2:], true
	}
	return "", false
}

// isBoldOverrideCommand 检查是否是粗体覆盖指令
// b开头且后面跟随数字的是粗体覆写指令
// 返回数字值，是否取消覆写，是否有效指令
func (p *Parser) isBoldOverrideCommand(overrideTag string) (weight int, cancel bool, valid bool) {
	if !strings.HasPrefix(overrideTag, "b") {
		return 0, false, false
	}

	param := overrideTag[1:]
	if param == "" {
		return 0, true, true // 只有 \b，恢复样式设置
	}
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, false, false
	}

	return value, false, true
}

// isItalicOverrideCommand 检查是否是斜体覆盖指令
// 如果是 i 开头且后面跟随数字，数字只有-1, 1和0三种值
// 返回数字值，是否取消覆写，是否有效指令
func (p *Parser) isItalicOverrideCommand(overrideTag string) (italic bool, cancel bool, valid bool) {
	if !strings.HasPrefix(overrideTag, "i") {
		return false, false, false
	}

	param := overrideTag[1:]
	if param == "" {
		return false, true, true // 只有 \i，恢复样式设置
	}
	switch param {
	case "-1", "1":
		return true, false, true // 设置为斜体
	case "0":
		return false, false, true // 取消斜体
	default:
		return false, false, false
	}
}

// applyFontNameOverride 应用字体名称覆写
func (p *Parser) applyFontNameOverride(currentFont *Font, fontName string, styleFontName string) {
	fontName = strings.TrimSpace(fontName)
	if fontName == "" {
		// 只有 \fn，恢复样式设置
		currentFont.FontName = styleFontName
	} else {
		// 移除可能的 @ 前缀（垂直文字标记）
		currentFont.FontName = strings.TrimPrefix(fontName, "@")
	}
}

// applyBoldOverride 应用粗体覆写
func (p *Parser) applyBoldOverride(currentFont *Font, cancel bool, weight int, styleBold int) {
	if cancel {
		currentFont.BoldWeight = styleBold
		return
	}
	currentFont.BoldWeight = p.parseBoldWeight(weight)
}

// applyItalicOverride 应用斜体覆写
func (p *Parser) applyItalicOverride(currentFont *Font, cancel bool, italicValue, styleItalic bool) {
	if cancel {
		currentFont.Italic = styleItalic
		return
	}
	currentFont.Italic = italicValue
}
