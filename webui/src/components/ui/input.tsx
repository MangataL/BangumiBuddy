import * as React from "react";

import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  ({ className, type, readOnly, onChange, ...props }, ref) => {
    const handleChange = React.useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        if (type === "number" && e.target.value) {
          // 去除前导0，但保留单个0和小数点前的0
          const value = e.target.value.replace(/^0+(?=\d)/, "");
          e.target.value = value;
        }
        onChange?.(e);
      },
      [type, onChange]
    );

    return (
      <input
        type={type}
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-base ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-gray-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
          readOnly && "bg-gray-100 text-gray-400",
          className
        )}
        readOnly={readOnly}
        ref={ref}
        onChange={handleChange}
        {...props}
      />
    );
  }
);
Input.displayName = "Input";

export { Input };
