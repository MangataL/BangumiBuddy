import { useState } from "react";
import { Lock, AlertCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/useToast";
import { DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { authApi, AuthErrorResponse } from "@/api/auth";
import { AxiosError } from "axios";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { PasswordInput } from "@/components/common/password-input";

interface ChangeAccountDialogProps {
  onUserChange: () => void;
}

export function ChangeAccountDialog({
  onUserChange: onPasswordChange,
}: ChangeAccountDialogProps) {
  const [open, setOpen] = useState(false);
  const [newUsername, setNewUsername] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [usernameError, setUsernameError] = useState("");
  const [usernameTouched, setUsernameTouched] = useState(false);
  const [passwordTouched, setPasswordTouched] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();
  const [error, setError] = useState<string | null>(null);

  const validateUsername = (username: string, force = false) => {
    if (!force && !usernameTouched) return true;
    if (!username.trim()) {
      setUsernameError("用户名不能为空");
      return false;
    }
    setUsernameError("");
    return true;
  };

  const validatePasswords = (force = false) => {
    if (!force && !passwordTouched) return true;
    if (!newPassword) {
      setPasswordError("请输入新密码");
      return false;
    }
    if (confirmPassword != "" && newPassword !== confirmPassword) {
      setPasswordError("两次输入的密码不一致");
      return false;
    }
    setPasswordError("");
    return true;
  };

  const handleAccountChange = async () => {
    const isUsernameValid = validateUsername(newUsername, true);
    const isPasswordValid = validatePasswords(true);

    if (!isUsernameValid || !isPasswordValid) {
      return;
    }

    setIsLoading(true);
    try {
      await authApi.updateUser(newUsername, newPassword);
      handleOpenChange(false);
      toast({
        title: "账号修改成功",
        description: "您的用户名和密码已成功更新，将自动退出登录",
      });
      setTimeout(() => {}, 200);
      onPasswordChange();
    } catch (error) {
      setError(
        (error as AxiosError<AuthErrorResponse>).response?.data
          ?.error_description || "修改账户时发生错误，请重试"
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      cleanup();
    }
    setOpen(open);
  };

  const cleanup = () => {
    setNewUsername("");
    setNewPassword("");
    setConfirmPassword("");
    setPasswordError("");
    setUsernameError("");
    setUsernameTouched(false);
    setPasswordTouched(false);
    setError(null);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <DropdownMenuItem
          className="rounded-lg focus:bg-primary/10"
          onSelect={(e) => {
            e.preventDefault();
          }}
        >
          <Lock className="mr-2 h-4 w-4" />
          <span>修改账号</span>
        </DropdownMenuItem>
      </DialogTrigger>
      <DialogContent className="rounded-xl border-primary/20 bg-card/95 backdrop-blur-md sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-xl anime-gradient-text flex items-center gap-2">
            <Lock className="h-5 w-5" />
            修改账号
          </DialogTitle>
          <DialogDescription>请输入您的新用户名和新密码</DialogDescription>
        </DialogHeader>
        {error && (
          <div className="px-2">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>修改失败</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </div>
        )}
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="new-username">新用户名</Label>
            <Input
              id="new-username"
              type="text"
              value={newUsername}
              onChange={(e) => {
                setNewUsername(e.target.value);
                setUsernameTouched(true);
              }}
              onBlur={() => validateUsername(newUsername)}
              onFocus={() => setError(null)}
              className="rounded-xl"
            />
          </div>
          {usernameError && (
            <div className="text-sm text-destructive">{usernameError}</div>
          )}
          <div className="grid gap-2">
            <Label htmlFor="new-password">新密码</Label>
            <PasswordInput
              id="new-password"
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value);
                setPasswordTouched(true);
              }}
              onBlur={() => validatePasswords()}
              onFocus={() => setError(null)}
              placeholder="请输入新密码"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="confirm-password">确认新密码</Label>
            <PasswordInput
              id="confirm-password"
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value);
              }}
              onBlur={() => validatePasswords()}
              onFocus={() => setError(null)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  handleAccountChange();
                }
              }}
              placeholder="请再次输入新密码"
            />
          </div>
          {passwordError && (
            <div className="text-sm text-destructive">{passwordError}</div>
          )}
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            className="rounded-xl"
            disabled={isLoading}
          >
            取消
          </Button>
          <Button
            onClick={handleAccountChange}
            className="rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
            disabled={isLoading}
          >
            {isLoading ? "修改中..." : "确认修改"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
