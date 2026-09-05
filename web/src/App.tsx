import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { OverviewPage } from "@/features/OverviewPage";
import { AppsPage } from "@/features/AppsPage";
import { CatalogPage } from "@/features/CatalogPage";
import { ProtectionPage } from "@/features/ProtectionPage";
import { ActivityPage } from "@/features/ActivityPage";
import { SettingsPage } from "@/features/SettingsPage";
import { SourcesPage } from "@/features/SourcesPage";
import { AuthGate } from "@/features/AuthGate";
import { BackupPage } from "@/features/BackupPage";
import { RecoveryPage } from "@/features/RecoveryPage";

export function App() {
  return (
    <AuthGate>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route index element={<OverviewPage />} />
            <Route path="apps" element={<AppsPage />} />
            <Route path="apps/:id" element={<AppsPage />} />
            <Route path="catalog" element={<CatalogPage />} />
            <Route path="sources" element={<SourcesPage />} />
            <Route path="protection" element={<ProtectionPage />} />
            <Route path="activity" element={<ActivityPage />} />
            <Route path="backups" element={<BackupPage />} />
            <Route path="recovery" element={<RecoveryPage />} />
            <Route path="settings" element={<SettingsPage />} />
            <Route
              path="*"
              element={
                <div
                  style={{
                    padding: "var(--space-12)",
                    textAlign: "center",
                    color: "var(--color-text-tertiary)",
                  }}
                >
                  <h1
                    style={{
                      fontSize: "var(--text-3xl)",
                      fontWeight: "var(--weight-bold)" as unknown as number,
                      marginBottom: "var(--space-2)",
                    }}
                  >
                    404
                  </h1>
                  <p>Page not found</p>
                </div>
              }
            />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthGate>
  );
}
