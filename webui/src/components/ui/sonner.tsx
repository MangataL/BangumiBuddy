"use client";

import { useTheme } from "next-themes";
import { Toaster as Sonner } from "sonner";

type ToasterProps = React.ComponentProps<typeof Sonner>;

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme();

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group pointer-events-none"
      position="top-center"
      richColors
      toastOptions={{
        classNames: {
          toast:
            "pointer-events-auto select-text group toast shadow-lg data-[type=default]:bg-background data-[type=default]:text-foreground data-[type=default]:border-border",
          description: "group-[.toast]:text-muted-foreground",
          actionButton:
            "pointer-events-auto group-[.toast]:bg-primary group-[.toast]:text-primary-foreground",
          cancelButton:
            "pointer-events-auto group-[.toast]:bg-muted group-[.toast]:text-muted-foreground",
        },
      }}
      {...props}
    />
  );
};

export { Toaster };
