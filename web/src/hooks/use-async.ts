import { useState, useEffect, useCallback } from "react";
import type { ApiResult, LoadingState } from "@/api/types";

export function useAsync<T>(
  fetcher: () => Promise<T>,
  deps: unknown[] = [],
): ApiResult<T> & { refetch: () => void } {
  const [data, setData] = useState<T | null>(null);
  const [state, setState] = useState<LoadingState>("idle");
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(() => {
    setState("loading");
    setError(null);
    fetcher()
      .then((result) => {
        setData(result);
        setState("success");
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Unknown error");
        setState("error");
      });
  }, deps);

  useEffect(() => {
    execute();
  }, [execute]);

  return { data, state, error, refetch: execute };
}
