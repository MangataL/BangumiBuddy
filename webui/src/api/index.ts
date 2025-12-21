import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
  type AxiosRequestConfig,
} from "axios";
import AuthService from "@/utils/auth-service";

export interface ErrorResponse {
  error: string;
}

type RequestConfig = AxiosRequestConfig & { skipAuth?: boolean };

declare module "axios" {
  interface AxiosRequestConfig {
    skipAuth?: boolean;
  }
}

type HttpInstance = Omit<
  AxiosInstance,
  "get" | "post" | "put" | "delete" | "patch"
> & {
  post<T = any, R = T>(
    url: string,
    data?: any,
    config?: RequestConfig
  ): Promise<R>;
  get<T = any, R = T>(url: string, config?: RequestConfig): Promise<R>;
  put<T = any, R = T>(
    url: string,
    data?: any,
    config?: RequestConfig
  ): Promise<R>;
  delete<T = any, R = T>(url: string, config?: RequestConfig): Promise<R>;
  patch<T = any, R = T>(
    url: string,
    data?: any,
    config?: RequestConfig
  ): Promise<R>;
};

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

export const http: HttpInstance = axios.create({
  baseURL: "/apis/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.request.use(authRequestInterceptor);
http.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  errorInterceptor
);
