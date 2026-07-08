import { useCallback, useRef, useState } from "react";

import { BangumiDetail, discoveryAPI } from "@/api/discovery";
import { extractErrorMessage } from "@/utils/error";

export function useBangumiDetail() {
  const [selectedID, setSelectedID] = useState<string>();
  const [detail, setDetail] = useState<BangumiDetail>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const requestIDRef = useRef(0);

  const open = useCallback(async (id: string) => {
    const requestID = requestIDRef.current + 1;
    requestIDRef.current = requestID;
    setSelectedID(id);
    setDetail(undefined);
    setError("");
    setLoading(true);
    try {
      const nextDetail = await discoveryAPI.getBangumi(id);
      if (requestID !== requestIDRef.current) return;
      setDetail(nextDetail);
    } catch (loadError) {
      if (requestID !== requestIDRef.current) return;
      setError(extractErrorMessage(loadError));
    } finally {
      if (requestID === requestIDRef.current) {
        setLoading(false);
      }
    }
  }, []);

  const close = useCallback(() => {
    requestIDRef.current += 1;
    setSelectedID(undefined);
    setDetail(undefined);
    setError("");
    setLoading(false);
  }, []);

  return {
    selectedID,
    detail,
    loading,
    error,
    open,
    close,
    retry: () => selectedID && open(selectedID),
  };
}
