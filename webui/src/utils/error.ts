import { AxiosError } from "axios";

/**
 * 从各种错误类型中提取错误信息
 * 优先从Axios错误的response.data.error获取，其次从Error.message获取，最后返回默认信息
 * @param error 捕获的错误对象
 * @param defaultMessage 默认错误信息
 * @returns 提取的错误信息
 */
export function extractErrorMessage(
  error: unknown,
  defaultMessage: string = "未知原因失败，请重试"
): string {
  return (
    (error as AxiosError<{ error: string }>)?.response?.data?.error ||
    (error as Error)?.message ||
    defaultMessage
  );
}
