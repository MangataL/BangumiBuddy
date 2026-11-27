// @vitest-environment jsdom
import { StrictMode } from "react";
import { act, cleanup, render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, it, expect, vi } from "vitest";
import { configAPI } from "@/api/config";
import MainLayout from "./main-layout";

const { toast } = vi.hoisted(() => ({ toast: vi.fn() }));
vi.mock("@/hooks/useToast", () => ({ useToast: () => ({ toast }) }));
vi.mock("@/contexts/auth", () => ({ useAuth: () => ({ logout: vi.fn() }) }));

afterEach(() => {
  cleanup();
  localStorage.clear();
  toast.mockClear();
  vi.restoreAllMocks();
});

async function openApp() {
  await act(async () => {
    render(<StrictMode><MemoryRouter><MainLayout>订阅</MainLayout></MemoryRouter></StrictMode>);
  });
}

it("仅在旧字体库首次打开时提示，同一浏览器刷新后仍记住", async () => {
  const stats = vi.spyOn(configAPI, "getSubtitleFontMetaSetStats")
    .mockResolvedValue({ total: 3, initDone: true, needsRebuild: true });
  await openApp();
  expect(toast).toHaveBeenCalledTimes(1);
  expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: "字体库需要重建" }));
  cleanup();
  stats.mockClear();
  await openApp();
  expect(toast).toHaveBeenCalledTimes(1);
  expect(stats).not.toHaveBeenCalled();
});

it.each([0, 3])("新库或已重建的字体库不提示（%i 款）", async (total) => {
  vi.spyOn(configAPI, "getSubtitleFontMetaSetStats")
    .mockResolvedValue({ total, initDone: total > 0, needsRebuild: false });
  await openApp();
  expect(toast).not.toHaveBeenCalled();
});
