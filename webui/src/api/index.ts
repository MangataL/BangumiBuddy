import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { TokenService } from "@/utils/token";
import NavigationService from "@/utils/navigation";
import AuthService from "@/utils/auth-service";

export interface ErrorResponse {
  error: string;
}

interface RequestConfig extends InternalAxiosRequestConfig {
  skipAuth?: boolean;
}

declare module "axios" {
  interface AxiosRequestConfig {
    skipAuth?: boolean;
  }
}

const authRequestInterceptor = async (config: InternalAxiosRequestConfig) => {
  const requestConfig = config as RequestConfig;
  if (requestConfig.skipAuth) {
    return config;
  }

  const accessToken = TokenService.getAccessToken();
  if (!accessToken) {
    NavigationService.navigateToLogin(window.location.pathname);
    throw new Error("登陆过期");
  }

  config.headers.Authorization = `Bearer ${accessToken}`;
  return config;
};

const errorInterceptor = (error: AxiosError<unknown>) => {
  if (error.response?.status === 401) {
    TokenService.clearTokens();
    // 使用AuthService处理未授权状态
    AuthService.handleUnauthorized();
    NavigationService.navigateToLogin(window.location.pathname);
  }
  return Promise.reject(error);
};

export const http = axios.create({
  baseURL: "/apis/v1",
  headers: {
    "Content-Type": "application/json",
  },
}) as AxiosInstance & {
  post<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T>;
  get<T = any>(url: string, config?: RequestConfig): Promise<T>;
  put<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T>;
  delete<T = any>(url: string, config?: RequestConfig): Promise<T>;
  patch<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T>;
};

http.interceptors.request.use(authRequestInterceptor);
http.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  errorInterceptor
);
