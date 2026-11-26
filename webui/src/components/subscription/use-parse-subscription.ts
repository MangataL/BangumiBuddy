import { useEffect, useRef, useState } from "react";
import { subscriptionAPI, type ParseRSSResponse } from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { extractErrorMessage } from "@/utils/error";

export function useParseSubscription(
  onParsed: (result: ParseRSSResponse) => void
) {
  const { toast } = useToast();
  const active = useRef<AbortController | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => () => active.current?.abort(), []);

  function cancel() {
    active.current?.abort();
    setLoading(false);
  }

  async function parse(rssLink: string, tmdbID?: number) {
    active.current?.abort();
    const controller = new AbortController();
    active.current = controller;
    setLoading(true);
    try {
      const result = await subscriptionAPI.parseRSS(
        { rssLink, tmdbID },
        controller.signal
      );
      if (!controller.signal.aborted) onParsed(result);
    } catch (error) {
      if (!controller.signal.aborted) {
        toast({
          title: "解析失败",
          description: extractErrorMessage(error),
          variant: "destructive",
        });
      }
    } finally {
      if (!controller.signal.aborted) setLoading(false);
    }
  }

  return { parse, cancel, loading };
}
