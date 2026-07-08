import { useCallback, useRef, useState } from "react";

import {
  DISCOVERY_SEARCH_PAGE_SIZE,
  discoveryAPI,
  SearchDiscoveryResp,
} from "@/api/discovery";
import { extractErrorMessage } from "@/utils/error";

export type DiscoverySearchState = SearchDiscoveryResp;

const emptySearchState: DiscoverySearchState = {
  bangumis: [],
  resources: [],
  bangumiTotal: 0,
  resourceTotal: 0,
  resourcePage: 1,
  resourcePageSize: DISCOVERY_SEARCH_PAGE_SIZE,
};

export function useDiscoverySearch() {
  const [query, setQuery] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [data, setData] = useState<DiscoverySearchState>(emptySearchState);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const requestIDRef = useRef(0);

  const runSearch = useCallback(
    async (value: string, page = 1) => {
      const normalizedQuery = value.trim();
      if (!normalizedQuery) {
        requestIDRef.current += 1;
        setSubmittedQuery("");
        setData(emptySearchState);
        setError("");
        setLoading(false);
        return;
      }

      const requestID = requestIDRef.current + 1;
      requestIDRef.current = requestID;
      const isNewQuery = normalizedQuery !== submittedQuery;
      setSubmittedQuery(normalizedQuery);
      setLoading(true);
      setError("");
      if (isNewQuery) {
        setData(emptySearchState);
      }
      try {
        const resp = await discoveryAPI.search(
          normalizedQuery,
          page,
          DISCOVERY_SEARCH_PAGE_SIZE
        );
        if (requestID !== requestIDRef.current) return;
        setData((current) => ({
          bangumis: page === 1 ? resp.bangumis : current.bangumis,
          resources: resp.resources,
          bangumiTotal: resp.bangumiTotal,
          resourceTotal: resp.resourceTotal,
          resourcePage: resp.resourcePage,
          resourcePageSize: resp.resourcePageSize,
        }));
      } catch (searchError) {
        if (requestID !== requestIDRef.current) return;
        setError(extractErrorMessage(searchError));
      } finally {
        if (requestID === requestIDRef.current) {
          setLoading(false);
        }
      }
    },
    [submittedQuery]
  );

  const submit = useCallback(() => runSearch(query, 1), [query, runSearch]);

  const clear = useCallback(() => {
    requestIDRef.current += 1;
    setQuery("");
    setSubmittedQuery("");
    setData(emptySearchState);
    setError("");
    setLoading(false);
  }, []);

  return {
    query,
    setQuery,
    submittedQuery,
    data,
    loading,
    error,
    submit,
    searchPage: (page: number) => runSearch(submittedQuery, page),
    clear,
  };
}
