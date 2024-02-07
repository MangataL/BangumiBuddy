import { ThemeProvider } from "./components/theme-provider";
import { Toaster } from "@/components/ui/toaster";
import { useNavigate } from "react-router-dom";
import { useEffect } from "react";
import NavigationService from "@/utils/navigation";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { AuthProvider } from "@/contexts/auth";

// 创建一个QueryClient实例
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1, // 失败重试次数
      refetchOnWindowFocus: false, // 窗口聚焦时不重新请求
    },
  },
});

export default function App({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();

  useEffect(() => {
    NavigationService.init(navigate);
  }, [navigate]);

  return (
    <>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="light"
            enableSystem
            disableTransitionOnChange
          >
            {children}
          </ThemeProvider>
        </AuthProvider>
        <ReactQueryDevtools initialIsOpen={false} /> {/* 开发工具 */}
      </QueryClientProvider>
      <Toaster />
    </>
  );
}
