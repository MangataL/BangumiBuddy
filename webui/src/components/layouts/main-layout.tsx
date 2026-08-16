import type React from "react";

import { useState } from "react";
import { useLocation, Link } from "react-router-dom";
import {
  Rss,
  Settings,
  FileText,
  Menu,
  User,
  LogOut,
  Moon,
  Sun,
  Magnet,
  Compass,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { useMobile } from "@/hooks/useMobile";
import { useToast } from "@/hooks/useToast";
import { useAuth } from "@/contexts/auth";
import { useTheme } from "@/components/theme-provider";
import { ChangeAccountDialog } from "@/components/account/change-account-dialog";
import {
  PageTransition,
  PAGE_TRANSITION_DURATION,
} from "@/components/transition/page-transition";

interface MainLayoutProps {
  children: React.ReactNode;
}

export default function MainLayout({ children }: MainLayoutProps) {
  const location = useLocation();
  const isMobile = useMobile();
  const { logout } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const [fadeOut, setFadeOut] = useState(false);
  const { theme, setTheme } = useTheme();
  const { toast } = useToast();
  const navigation = [
    { name: "订阅管理", href: "/", icon: Rss },
    { name: "发现", href: "/discover", icon: Compass },
    { name: "磁力下载", href: "/download", icon: Magnet },
    { name: "设置中心", href: "/settings", icon: Settings },
    { name: "日志", href: "/logs", icon: FileText },
  ];

  const handleLogout = (needToast: boolean = true) => {
    setFadeOut(true);
    if (needToast) {
      toast({
        title: "已退出登录",
        description: "您已成功退出系统",
      });
    }
    setTimeout(() => {
      logout();
    }, PAGE_TRANSITION_DURATION);
  };

  const NavLinks = () => (
    <>
      <div className="flex h-20 shrink-0 items-center px-6">
        <Link to="/" className="flex items-center gap-2">
          <div className="rounded-xl bg-primary p-1 animate-pulse-glow">
            <img src="/logo.png" alt="Logo" className="h-8 w-8" />
          </div>
          <span className="text-xl font-bold anime-gradient-text">
            BangumiBuddy
          </span>
        </Link>
      </div>

      <nav className="flex flex-1 flex-col">
        <ul className="flex flex-1 flex-col gap-2 px-4">
          {navigation.map((item) => {
            const isActive = location.pathname === item.href;
            return (
              <li key={item.name}>
                <Link
                  to={item.href}
                  className={cn(
                    "group flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium transition-all duration-200 hover:bg-accent",
                    isActive ? "bg-accent anime-border" : "transparent"
                  )}
                  onClick={() => setIsOpen(false)}
                >
                  <item.icon
                    className={cn(
                      "h-5 w-5 transition-transform duration-300 group-hover:scale-110",
                      isActive ? "text-primary" : "text-muted-foreground"
                    )}
                  />
                  {item.name}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
    </>
  );

  return (
    <PageTransition fadeOut={fadeOut}>
      <div className="flex h-screen w-full overflow-hidden supports-[height:100dvh]:h-[100dvh]">
        {isMobile ? (
          <Sheet open={isOpen} onOpenChange={setIsOpen}>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="absolute left-4 top-4 z-50 md:hidden"
              >
                <Menu className="h-5 w-5" />
                <span className="sr-only">Toggle menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent
              side="left"
              className="w-64 p-0 bg-sidebar text-sidebar-foreground"
            >
              <NavLinks />
            </SheetContent>
          </Sheet>
        ) : (
          <div className="fixed z-10 hidden h-screen w-64 flex-col border-r bg-sidebar text-sidebar-foreground supports-[height:100dvh]:h-[100dvh] md:flex">
            <NavLinks />
          </div>
        )}

        <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden md:ml-64">
          <header className="sticky top-0 z-10 flex h-16 flex-none items-center justify-end border-b bg-background/80 px-6 backdrop-blur-md">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
              className="mr-2 rounded-full anime-button"
            >
              {theme === "dark" ? (
                <Sun className="h-5 w-5" />
              ) : (
                <Moon className="h-5 w-5" />
              )}
              <span className="sr-only">切换主题</span>
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="rounded-full anime-button"
                >
                  <User className="h-5 w-5" />
                  <span className="sr-only">User menu</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="end"
                className="rounded-xl border-primary/20 bg-card/95 backdrop-blur-md"
              >
                <ChangeAccountDialog onUserChange={() => handleLogout(false)} />
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="rounded-lg focus:bg-primary/10"
                  onClick={() => handleLogout()}
                >
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>注销登录</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </header>

          <main className="min-h-0 flex-1 overflow-auto p-4 pb-[calc(1rem+env(safe-area-inset-bottom))] sm:p-6 sm:pb-[calc(1.5rem+env(safe-area-inset-bottom))]">
            {children}
          </main>
        </div>
      </div>
    </PageTransition>
  );
}
