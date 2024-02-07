import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/contexts/auth";

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      // 将当前路径作为redirect参数
      const redirectPath = `${location.pathname}${location.search}`;
      navigate(`/login?redirect=${encodeURIComponent(redirectPath)}`);
    }
  }, [isAuthenticated, isLoading, navigate, location]);

  // 如果正在加载认证状态，可以显示加载中状态
  if (isLoading) {
    return <div>加载中...</div>; // 你可以替换成更好看的加载组件
  }

  // 如果已认证，渲染子组件
  return isAuthenticated ? <>{children}</> : null;
} 