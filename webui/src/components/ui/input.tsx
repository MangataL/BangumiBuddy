import * as React from "react";

import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  (
    { className, type, readOnly, onChange, onBlur, onFocus, value, ...props },
    ref
  ) => {
    const isControlledNumber = type === "number" && value !== undefined;
    const [draftValue, setDraftValue] = React.useState<string | null>(null);
    const [isFocused, setIsFocused] = React.useState(false);

    React.useEffect(() => {
      if (!isFocused) {
        setDraftValue(null);
      }
    }, [value, isFocused]);

    const handleChange = React.useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        if (type === "number" && e.target.value) {
          const rawValue = e.target.value;
          if (isControlledNumber && rawValue === "-") {
            setDraftValue(rawValue);
            return;
          }
          // 去除前导0，但保留单个0和小数点前的0
          const value = rawValue.replace(/^(-?)0+(?=\d)/, "$1");
          if (value !== rawValue) {
            e.target.value = value;
          }
          if (isControlledNumber) {
            setDraftValue(value);
          }
          onChange?.(e);
          return;
        }
        if (isControlledNumber) {
          setDraftValue(e.target.value);
        }
        onChange?.(e);
      },
      [type, onChange, isControlledNumber]
    );

    const handleFocus = React.useCallback(
      (e: React.FocusEvent<HTMLInputElement>) => {
        if (isControlledNumber) {
          setIsFocused(true);
        }
        onFocus?.(e);
      },
      [isControlledNumber, onFocus]
    );

    const handleBlur = React.useCallback(
      (e: React.FocusEvent<HTMLInputElement>) => {
        if (isControlledNumber) {
          setIsFocused(false);
          setDraftValue(null);
        }
        onBlur?.(e);
      },
      [isControlledNumber, onBlur]
    );

    const inputValue =
      isControlledNumber && draftValue !== null ? draftValue : value;

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
        onFocus={handleFocus}
        onBlur={handleBlur}
        value={inputValue}
        {...props}
      />
    );
  }
);
Input.displayName = "Input";

export { Input };
