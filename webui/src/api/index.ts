import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
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

  const accessToken = await AuthService.getAccessToken();
  if (!accessToken) {
    AuthService.handleUnauthorized();
    return Promise.reject(new Error("登陆信息已过期，请重新登陆"));
  }
  config.headers.Authorization = `Bearer ${accessToken}`;
  return config;
};

const errorInterceptor = (error: AxiosError<unknown>) => {
  if (error.response?.status === 401) {
    AuthService.handleUnauthorized();
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
