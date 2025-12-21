import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

function getEventTarget(event: unknown): EventTarget | null {
  if (!event || typeof event !== "object") return null;
  const e = event as any;
  return e?.detail?.originalEvent?.target ?? e?.target ?? null;
}

export function isSonnerEventTarget(target: EventTarget | null) {
  if (!target) return false;

  // Radix/React 事件 target 可能是 Text 节点
  const node = target as any;
  const el =
    typeof Node !== "undefined" && node instanceof Node
      ? node.nodeType === Node.TEXT_NODE
        ? node.parentElement
        : node
      : null;

  if (!el || typeof (el as any).closest !== "function") return false;
  return Boolean(
    (el as Element).closest("[data-sonner-toaster],[data-sonner-toast]")
  );
}

export function preventRadixDismissOnSonnerInteract(event: unknown) {
  const target = getEventTarget(event);
  if (!isSonnerEventTarget(target)) return;
  const e = event as any;
  if (typeof e?.preventDefault === "function") e.preventDefault();
}
