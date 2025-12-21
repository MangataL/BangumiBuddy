import type { ReactNode } from "react";
import { toast as sonnerToast } from "sonner";

type ToastVariant = "default" | "destructive";

type ToastOptions = {
  title?: ReactNode;
  description?: ReactNode;
  variant?: ToastVariant;
} & Omit<Parameters<typeof sonnerToast>[1], "description">;

const toast = ({
  title,
  description,
  variant = "default",
  ...options
}: ToastOptions) => {
  const message = title ?? description ?? "";
  const sonnerOptions =
    description && title !== description
      ? { ...options, description }
      : { ...options };

  if (variant === "destructive") {
    return sonnerToast.error(message, sonnerOptions);
  }

  return sonnerToast(message, sonnerOptions);
};

function useToast() {
  return {
    toast,
    dismiss: sonnerToast.dismiss,
  };
}

export { useToast, toast };
