import { useState, useCallback, useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { NavItem } from "./NavItem";
import {
  IconOverview,
  IconApps,
  IconCatalog,
  IconProtection,
  IconActivity,
  IconSettings,
  IconMenu,
  IconClose,
} from "./Icons";
import { useAuth } from "@/features/AuthGate";
import styles from "./Layout.module.css";

const NAV_ITEMS = [
  { to: "/", label: "Overview", icon: <IconOverview /> },
  { to: "/apps", label: "Apps", icon: <IconApps /> },
  { to: "/catalog", label: "Catalog", icon: <IconCatalog /> },
  { to: "/protection", label: "Protection", icon: <IconProtection /> },
  { to: "/activity", label: "Activity", icon: <IconActivity /> },
  { to: "/settings", label: "Settings", icon: <IconSettings /> },
] as const;

export function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [logoutError, setLogoutError] = useState("");
  const location = useLocation();
  const { authEnabled, logout } = useAuth();

  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname]);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen((prev) => !prev);
  }, []);

  const closeSidebar = useCallback(() => {
    setSidebarOpen(false);
  }, []);

  useEffect(() => {
    if (!sidebarOpen) return;
    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") closeSidebar();
    }
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [sidebarOpen, closeSidebar]);

  return (
    <div className={styles.layout}>
      <div
        className={styles.mobileOverlay}
        data-open={sidebarOpen}
        onClick={closeSidebar}
        aria-hidden="true"
      />
      <aside
        className={styles.sidebar}
        data-open={sidebarOpen}
        aria-label="Main navigation"
      >
        <div className={styles.logo}>
          <div className={styles.logoMark} aria-hidden="true">
            O
          </div>
          <span className={styles.logoText}>OpenDash</span>
        </div>
        <nav className={styles.nav} role="navigation">
          {NAV_ITEMS.map((item) => (
            <NavItem key={item.to} {...item} />
          ))}
        </nav>
        <div className={styles.sidebarFooter}>
          {logoutError && <span role="alert">{logoutError}</span>}
          {authEnabled && (
            <button
              type="button"
              onClick={() => {
                setLogoutError("");
                void logout().catch((cause: unknown) => {
                  setLogoutError(
                    cause instanceof Error ? cause.message : "Unable to sign out",
                  );
                });
              }}
            >
              Sign out
            </button>
          )}
          <span>OpenDash v0.1.0</span>
        </div>
      </aside>
      <div className={styles.main}>
        <header className={styles.mobileHeader}>
          <button
            className={styles.menuButton}
            onClick={toggleSidebar}
            aria-label={sidebarOpen ? "Close navigation" : "Open navigation"}
            aria-expanded={sidebarOpen}
          >
            {sidebarOpen ? <IconClose /> : <IconMenu />}
          </button>
          <span className={styles.logoText}>OpenDash</span>
        </header>
        <main className={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
