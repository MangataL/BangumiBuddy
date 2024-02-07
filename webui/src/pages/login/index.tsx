"use client";

import type React from "react";
import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { LogIn, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useToast } from "@/hooks/useToast";
import { useAuth } from "@/contexts/auth";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  PageTransition,
  PAGE_TRANSITION_DURATION,
} from "@/components/transition/page-transition";
import { PasswordInput } from "@/components/common/password-input";

export default function LoginPage() {
  const { login, isLoading, error, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const { toast } = useToast();
  const location = useLocation();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(false);
  const [fadeOut, setFadeOut] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);

  useEffect(() => {
    if (isAuthenticated) {
      const searchParams = new URLSearchParams(location.search);
      const redirectUrl = searchParams.get("redirect") || "/";
      setFadeOut(true);
      const timer = setTimeout(() => {
        navigate(redirectUrl);
      }, PAGE_TRANSITION_DURATION);
      return () => clearTimeout(timer);
    }
  }, [isAuthenticated, navigate, location]);

  useEffect(() => {
    if (error) {
      const authError = error.response?.data;
      setLoginError(authError?.error_description || "登录时发生错误，请重试");
      setFadeOut(false);
    }
  }, [error]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginError(null);
    try {
      await login(username, password, rememberMe);
      toast({
        title: "登录成功",
        description: "欢迎回来，开启你的番剧之旅吧！",
      });
    } catch (err) {
      // 错误已经在 useAuth 中处理
    }
  };

  return (
    <PageTransition fadeOut={fadeOut}>
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center mb-4">
              <div className="rounded-xl bg-primary p-2 animate-pulse-glow">
                <img src="/logo.png" alt="Logo" className="h-16 w-16" />
              </div>
            </div>
            <h1 className="text-3xl font-bold anime-gradient-text">
              BangumiBuddy
            </h1>
            <p className="text-muted-foreground mt-2">
              您的专属番剧订阅管理助手
            </p>
          </div>

          <Card className="border-primary/10 rounded-xl overflow-hidden">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5 pb-6">
              <CardTitle className="text-xl anime-gradient-text flex items-center gap-2">
                <LogIn className="h-5 w-5" />
                用户登录
              </CardTitle>
            </CardHeader>
            {loginError && (
              <div className="px-6 pt-2">
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>登录失败</AlertTitle>
                  <AlertDescription>{loginError}</AlertDescription>
                </Alert>
              </div>
            )}
            <form onSubmit={handleLogin}>
              <CardContent className="space-y-4 pt-6">
                <div className="space-y-2">
                  <Label htmlFor="username">用户名</Label>
                  <Input
                    id="username"
                    placeholder="请输入用户名"
                    value={username}
                    onChange={(e) => {
                      setUsername(e.target.value);
                      setLoginError(null);
                    }}
                    required
                    className="rounded-xl"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password">密码</Label>
                  <PasswordInput
                    id="password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      setLoginError(null);
                    }}
                    required
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="remember"
                    checked={rememberMe}
                    onCheckedChange={(checked: boolean | "indeterminate") =>
                      setRememberMe(checked === true)
                    }
                  />
                  <Label htmlFor="remember" className="text-sm font-normal">
                    记住我
                  </Label>
                </div>
              </CardContent>
              <CardFooter className="flex flex-col gap-4">
                <Button
                  type="submit"
                  className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                  disabled={isLoading}
                >
                  {isLoading ? "登录中..." : "登录"}
                </Button>
              </CardFooter>
            </form>
          </Card>
        </div>
      </div>
    </PageTransition>
  );
}
