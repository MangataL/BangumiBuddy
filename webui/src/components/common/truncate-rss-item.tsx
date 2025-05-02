import { useRef, useEffect, useState, useMemo } from "react";

// 定义字幕信息的关键字
const subtitleKeywords = [
  "简体",
  "繁体",
  "简日",
  "简繁日",
  "繁日",
  "简繁",
  "CHS",
  "CHT",
];

export function TruncatedText({ text }: { text: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  // 状态：截断后的文本
  const [truncatedText, setTruncatedText] = useState(text);

  // 创建并复用 Canvas 上下文以测量文本宽度
  const context = useMemo(() => {
    const canvas = document.createElement("canvas");
    return canvas.getContext("2d");
  }, []);

  function extractPositions(text: string) {
    // 提取集数，如 [01] 或 - 01
    const episodeMatch = text.match(/(\[\d+\]|-\s*\d+)/);
    const episodeIndex = episodeMatch ? episodeMatch.index : -1;
    const episodeText = episodeMatch ? episodeMatch[0] : "";

    // 从后向前查找字幕信息
    const brackets = text.match(/\[[^\]]+\]/g) || [];
    let subtitleIndex = -1;
    let subtitleText = "";
    for (let i = brackets.length - 1; i >= 0; i--) {
      const content = brackets[i].slice(1, -1);
      for (const keyword of subtitleKeywords) {
        if (content.includes(keyword)) {
          subtitleText = keyword;
          subtitleIndex = text.lastIndexOf(brackets[i]);
          break;
        }
      }
      if (subtitleText) break;
    }

    return { episodeIndex, episodeText, subtitleIndex, subtitleText };
  }
  function truncateText(text: string) {
    if (!containerRef.current) return;
    const containerWidth = containerRef.current.offsetWidth;
    const font = window.getComputedStyle(containerRef.current).font;
    if (!context) return;
    context.font = font;

    // 测量文本宽度的函数
    const measureTextWidth = (text: string) => context.measureText(text).width;

    // 如果完整文本宽度小于容器宽度，直接展示完整文本
    const fullWidth = measureTextWidth(text);
    if (fullWidth <= containerWidth) {
      setTruncatedText(text);
      return;
    }

    // 提取集数和字幕信息的位置
    const { episodeIndex, episodeText, subtitleIndex, subtitleText } =
      extractPositions(text);
    if (
      !episodeIndex ||
      !subtitleIndex ||
      episodeIndex === -1 ||
      subtitleIndex === -1
    ) {
      setTruncatedText(text); // 无法解析时展示原始文本
      return;
    }

    // 构建必须展示的部分
    const middleText = text
      .substring(episodeIndex + episodeText.length, subtitleIndex)
      .trim();
    const suffixText = text
      .substring(subtitleIndex + subtitleText.length + 2)
      .trim();
    let mustDisplay = `${episodeText}`;
    if (middleText.length > 0) {
      mustDisplay = `${mustDisplay}...`;
    }
    mustDisplay = `${mustDisplay} ${subtitleText}`;
    if (suffixText.length > 0) {
      mustDisplay = `${mustDisplay}...`;
    }
    const mustDisplayWidth = measureTextWidth(mustDisplay);

    // 计算可用于前缀的剩余宽度
    const availableWidth = containerWidth - mustDisplayWidth;

    // 提取前缀部分
    const prefixText = text.substring(0, episodeIndex).trim();
    const prefixWidth = measureTextWidth(prefixText);

    if (prefixWidth <= availableWidth) {
      // 前缀宽度足够，直接拼接完整展示
      setTruncatedText(prefixText + mustDisplay);
    } else {
      // 前缀需要截断，使用二分查找确定截断位置
      let low = 0;
      let high = prefixText.length;
      while (low < high) {
        const mid = Math.floor((low + high + 1) / 2);
        const prefix = prefixText.substring(0, mid);
        const truncated = prefix + "...";
        if (measureTextWidth(truncated) <= availableWidth) {
          low = mid;
        } else {
          high = mid - 1;
        }
      }
      const prefix = prefixText.substring(0, low);
      const truncated = prefix + "...";
      setTruncatedText(truncated + mustDisplay);
    }
  }
  useEffect(() => {
    // 初始渲染时执行一次截断
    truncateText(text);

    if (!containerRef.current) return;
    const el = containerRef.current;
    const resizeObserver = new ResizeObserver(() => {
      truncateText(text);
    });
    resizeObserver.observe(el);
    return () => resizeObserver.disconnect();
  }, [text, truncateText, context]);

  return (
    <div ref={containerRef}>
      {truncatedText}
    </div>
  );
}
