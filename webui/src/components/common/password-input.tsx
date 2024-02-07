import { useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface PasswordInputProps
  extends Omit<React.ComponentProps<"input">, "type"> {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onBlur?: () => void;
  onFocus?: () => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  placeholder?: string;
  id?: string;
  className?: string;
}

export function PasswordInput({
  value,
  onChange,
  onBlur,
  onFocus,
  onKeyDown,
  placeholder = "请输入密码",
  id = "password",
  className = "",
  ...props
}: PasswordInputProps) {
  const [showPassword, setShowPassword] = useState(false);

  return (
    <div className="relative">
      <Input
        id={id}
        type={showPassword ? "text" : "password"}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        onBlur={onBlur}
        onFocus={onFocus}
        onKeyDown={onKeyDown}
        className={`rounded-xl pr-10 ${className}`}
        {...props}
      />
      <Button
        type="button"
        variant="ghost"
        size="icon"
        tabIndex={-1}
        className="absolute right-0 top-0 h-full px-3 py-2"
        onClick={() => setShowPassword(!showPassword)}
      >
        {showPassword ? (
          <EyeOff className="h-4 w-4 text-muted-foreground" />
        ) : (
          <Eye className="h-4 w-4 text-muted-foreground" />
        )}
      </Button>
    </div>
  );
}
