import { http } from "./index";

// 后端返回的日志格式
interface BackendLogEntry {
  level: string;
  ts: number;
  msg: string;
}

// 前端使用的日志格式
export interface LogEntry {
  level: string;
  ts: number;
  message: string;
}

export const LogService = {
  /**
   * 获取系统日志
   * @param level 日志级别，不传则获取全部
   * @param keyword 关键字，不传则不进行关键字过滤
   */
  getLogs: async (level?: string, keyword?: string): Promise<LogEntry[]> => {
    const response: BackendLogEntry[] = await http.get("/logs", {
      params: { level: level && level !== "all" ? level : undefined, keyword },
    });
    // 把后端返回的 msg 字段转换成前端使用的 message 字段
    return response.map((entry) => ({
      level: entry.level,
      ts: entry.ts,
      message: entry.msg,
    }));
  },
};
