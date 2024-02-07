import { createContext, useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { authApi, type AuthErrorResponse } from "@/api/auth";
import { TokenService } from "@/utils/token";
import type { AxiosError } from "axios";
import AuthService from "@/utils/auth-service";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (
    username: string,
    password: string,
    remember: boolean
  ) => Promise<void>;
  logout: () => void;
  error: AxiosError<AuthErrorResponse> | null;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<AxiosError<AuthErrorResponse> | null>(
    null
  );
  const [isAuthenticated, setIsAuthenticated] = useState(() => {
    const token = TokenService.getAccessToken();
    return !!token && !TokenService.isTokenExpired(token);
  });

  useEffect(() => {
    // 监听 token 过期
    const checkTokenExpiry = () => {
      const token = TokenService.getAccessToken();
      if (!token || TokenService.isTokenExpired(token)) {
        const refreshToken = TokenService.getRefreshToken();
        if (refreshToken && !TokenService.isTokenExpired(refreshToken)) {
          refreshAccessToken();
        } else {
          setIsAuthenticated(false);
          TokenService.clearTokens();
          navigate("/login");
        }
      }
    };

    const interval = setInterval(checkTokenExpiry, 60000); // 每分钟检查一次
    return () => clearInterval(interval);
  }, [navigate]);

  useEffect(() => {
    // 注册401未授权处理回调
    const handleUnauthorized = () => {
      setIsAuthenticated(false);
    };

    AuthService.registerLogoutCallback(handleUnauthorized);

    return () => {
      AuthService.unregisterLogoutCallback();
    };
  }, []);

  const refreshAccessToken = async () => {
    try {
      const refreshToken = TokenService.getRefreshToken();
      if (!refreshToken) return;

      const response = await authApi.refreshToken(refreshToken);
      TokenService.storeTokens(
        {
          accessToken: response.access_token,
          refreshToken: response.refresh_token,
        },
        true
      );
      setIsAuthenticated(true);
    } catch (err) {
      setIsAuthenticated(false);
      navigate("/login");
    }
  };

  const login = async (
    username: string,
    password: string,
    remember: boolean
  ) => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await authApi.login(username, password);
      TokenService.storeTokens(
        {
          accessToken: response.access_token,
          refreshToken: response.refresh_token,
        },
        remember
      );
      setIsAuthenticated(true);
    } catch (err) {
      setError(err as AxiosError<AuthErrorResponse>);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    TokenService.clearTokens();
    setIsAuthenticated(false);
    navigate("/login");
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoading,
        login,
        logout,
        error,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
