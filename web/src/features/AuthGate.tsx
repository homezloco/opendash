import {
  createContext,
  type FormEvent,
  type ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";
import { api, APIError, useMockApi } from "@/api/client";
import type { BootstrapStatus } from "@/api/types";

const AuthContext = createContext<{
  authEnabled: boolean;
  logout: () => Promise<void>;
}>({
  authEnabled: false,
  logout: async () => undefined,
});

export function useAuth() {
  return useContext(AuthContext);
}

export function AuthGate({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<BootstrapStatus | null>(
    useMockApi ? { authEnabled: false, bootstrapRequired: false } : null,
  );
  const [authenticated, setAuthenticated] = useState(useMockApi);
  const [error, setError] = useState("");

  useEffect(() => {
    if (useMockApi) return;
    let active = true;
    async function load() {
      try {
        const next = await api.getBootstrapStatus();
        if (!active) return;
        setStatus(next);
        if (!next.authEnabled) {
          setAuthenticated(true);
        } else if (!next.bootstrapRequired) {
          try {
            await api.getSession();
            if (active) setAuthenticated(true);
          } catch (cause) {
            if (!(cause instanceof APIError && cause.status === 401)) throw cause;
          }
        }
      } catch (cause) {
        if (active) {
          setError(cause instanceof Error ? cause.message : "Unable to load authentication status");
        }
      }
    }
    void load();
    return () => {
      active = false;
    };
  }, []);

  async function logout() {
    await api.logout();
    setAuthenticated(false);
    setStatus((current) =>
      current ? { ...current, bootstrapRequired: false } : current,
    );
  }

  if (authenticated) {
    return (
      <AuthContext.Provider
        value={{ authEnabled: status?.authEnabled ?? false, logout }}
      >
        {children}
      </AuthContext.Provider>
    );
  }
  if (!status) {
    return (
      <main aria-live="polite" style={{ padding: "3rem", textAlign: "center" }}>
        {error || "Loading OpenDash…"}
      </main>
    );
  }
  return (
    <AuthForm
      bootstrap={status.bootstrapRequired}
      onSuccess={() => setAuthenticated(true)}
      onBootstrapConflict={() =>
        setStatus({ authEnabled: true, bootstrapRequired: false })
      }
    />
  );
}

function AuthForm({
  bootstrap,
  onSuccess,
  onBootstrapConflict,
}: {
  bootstrap: boolean;
  onSuccess: () => void;
  onBootstrapConflict: () => void;
}) {
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const values = new FormData(event.currentTarget);
    try {
      await (bootstrap ? api.bootstrap : api.login)(
        String(values.get("username")),
        String(values.get("password")),
      );
      onSuccess();
    } catch (cause) {
      if (bootstrap && cause instanceof APIError && cause.status === 409) {
        onBootstrapConflict();
        setError("An administrator already exists. Sign in instead.");
      } else {
        setError(cause instanceof Error ? cause.message : "Authentication failed");
      }
    } finally {
      setBusy(false);
    }
  }

  const heading = bootstrap ? "Create administrator" : "Sign in to OpenDash";
  return (
    <main style={{ maxWidth: 420, margin: "10vh auto", padding: "2rem" }}>
      <h1>{heading}</h1>
      <p>
        {bootstrap
          ? "Create the one OpenDash administrator account. Use at least 12 password characters."
          : "Enter your administrator credentials."}
      </p>
      <form onSubmit={submit} style={{ display: "grid", gap: "1rem" }}>
        <label>
          Username
          <input
            name="username"
            minLength={3}
            maxLength={64}
            required
            autoComplete="username"
            autoFocus
          />
        </label>
        <label>
          Password
          <input
            name="password"
            type="password"
            minLength={12}
            maxLength={1024}
            required
            autoComplete={bootstrap ? "new-password" : "current-password"}
          />
        </label>
        {error && (
          <p role="alert" aria-live="assertive">
            {error}
          </p>
        )}
        <button disabled={busy} type="submit">
          {busy ? "Please wait…" : heading}
        </button>
      </form>
    </main>
  );
}
