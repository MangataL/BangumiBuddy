import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  BangumiCandidate,
  discoveryAPI,
  ReleaseGroupSummary,
  Season,
} from "@/api/discovery";
import { extractErrorMessage } from "@/utils/error";

import { getCurrentSeason } from "../discovery-options";
import { getCurrentWeekday } from "../weekday-schedule";

export function useSeasonDiscovery() {
  const currentWeekday = useMemo(() => getCurrentWeekday(), []);
  const [year, setYear] = useState(new Date().getFullYear());
  const [season, setSeason] = useState<Season>(getCurrentSeason());
  const [activeWeekday, setActiveWeekday] = useState(currentWeekday);
  const [bangumis, setBangumis] = useState<BangumiCandidate[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [summaries, setSummaries] = useState<Record<string, ReleaseGroupSummary>>({});
  const [summaryLoadingIDs, setSummaryLoadingIDs] = useState<string[]>([]);
  const [summaryErrors, setSummaryErrors] = useState<Record<string, string>>({});
  const generationRef = useRef(0);
  const summariesRef = useRef<Record<string, ReleaseGroupSummary>>({});
  const attemptedIDsRef = useRef(new Set<string>());
  const inFlightIDsRef = useRef(new Set<string>());

  const resetSummaryState = useCallback(() => {
    summariesRef.current = {};
    attemptedIDsRef.current = new Set();
    inFlightIDsRef.current = new Set();
    setSummaries({});
    setSummaryLoadingIDs([]);
    setSummaryErrors({});
  }, []);

  const loadSeason = useCallback(
    async (resetActiveDay: boolean, clearContent: boolean) => {
      const generation = generationRef.current + 1;
      generationRef.current = generation;
      setLoading(true);
      setError("");
      if (clearContent) {
        setBangumis([]);
      }
      try {
        const nextBangumis = await discoveryAPI.listBangumis({ year, season });
        if (generation !== generationRef.current) return;
        resetSummaryState();
        setBangumis(nextBangumis);
        if (resetActiveDay) {
          setActiveWeekday(currentWeekday);
        }
      } catch (loadError) {
        if (generation !== generationRef.current) return;
        setError(extractErrorMessage(loadError));
      } finally {
        if (generation === generationRef.current) {
          setLoading(false);
        }
      }
    },
    [currentWeekday, resetSummaryState, season, year]
  );

  useEffect(() => {
    void loadSeason(true, true);
  }, [loadSeason]);

  const loadReleaseGroups = useCallback(async (ids: string[], force = false) => {
    const generation = generationRef.current;
    const requestIDs = Array.from(new Set(ids.map((id) => id.trim()).filter(Boolean))).filter(
      (id) =>
        !summariesRef.current[id] &&
        !inFlightIDsRef.current.has(id) &&
        (force || !attemptedIDsRef.current.has(id))
    );
    if (requestIDs.length === 0) return;

    requestIDs.forEach((id) => {
      attemptedIDsRef.current.add(id);
      inFlightIDsRef.current.add(id);
    });
    setSummaryLoadingIDs(Array.from(inFlightIDsRef.current));
    setSummaryErrors((current) => {
      const next = { ...current };
      requestIDs.forEach((id) => delete next[id]);
      return next;
    });

    try {
      const resp = await discoveryAPI.batchReleaseGroups(requestIDs);
      if (generation !== generationRef.current) return;
      const nextSummaries = { ...summariesRef.current };
      resp.items.forEach((summary) => {
        nextSummaries[summary.mikanBangumiID] = summary;
      });
      summariesRef.current = nextSummaries;
      setSummaries(nextSummaries);
      if (resp.failures.length > 0) {
        setSummaryErrors((current) => {
          const next = { ...current };
          resp.failures.forEach((failure) => {
            next[failure.mikanBangumiID] = failure.message;
          });
          return next;
        });
      }
    } catch (loadError) {
      if (generation !== generationRef.current) return;
      const message = extractErrorMessage(loadError);
      setSummaryErrors((current) => {
        const next = { ...current };
        requestIDs.forEach((id) => {
          next[id] = message;
        });
        return next;
      });
    } finally {
      if (generation === generationRef.current) {
        requestIDs.forEach((id) => inFlightIDsRef.current.delete(id));
        setSummaryLoadingIDs(Array.from(inFlightIDsRef.current));
      }
    }
  }, []);

  const activeBangumis = useMemo(
    () => bangumis.filter((item) => item.weekday === activeWeekday),
    [activeWeekday, bangumis]
  );

  useEffect(() => {
    const missingIDs = activeBangumis
      .filter(
        (item) =>
          item.releaseGroupsKnownEmpty !== true &&
          !summaries[item.mikanBangumiID]
      )
      .map((item) => item.mikanBangumiID);
    void loadReleaseGroups(missingIDs);
  }, [activeBangumis, loadReleaseGroups, summaries]);

  const retryReleaseGroups = useCallback(
    (id: string) => {
      attemptedIDsRef.current.delete(id);
      return loadReleaseGroups([id], true);
    },
    [loadReleaseGroups]
  );

  const refreshReleaseGroups = useCallback(
    async (id: string) => {
      const generation = generationRef.current;
      while (inFlightIDsRef.current.has(id)) {
        await new Promise((resolve) => window.setTimeout(resolve, 50));
        if (generation !== generationRef.current) return;
      }
      if (generation !== generationRef.current) return;
      const nextSummaries = { ...summariesRef.current };
      delete nextSummaries[id];
      summariesRef.current = nextSummaries;
      setSummaries(nextSummaries);
      attemptedIDsRef.current.delete(id);
      await loadReleaseGroups([id], true);
    },
    [loadReleaseGroups]
  );

  return {
    year,
    setYear,
    season,
    setSeason,
    activeWeekday,
    setActiveWeekday,
    currentWeekday,
    bangumis,
    activeBangumis,
    loading,
    error,
    summaries,
    summaryLoadingIDs,
    summaryErrors,
    refresh: () => loadSeason(false, false),
    retryReleaseGroups,
    refreshReleaseGroups,
  };
}
