import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { OverviewPage } from "@/features/OverviewPage";
import { AppsPage } from "@/features/AppsPage";
import { ProtectionPage } from "@/features/ProtectionPage";
import { ActivityPage } from "@/features/ActivityPage";

function renderWithRouter(ui: React.ReactElement, path = "/") {
  return render(<MemoryRouter initialEntries={[path]}>{ui}</MemoryRouter>);
}

describe("OverviewPage", () => {
  it("renders page title", () => {
    renderWithRouter(<OverviewPage />);
    expect(screen.getByText("Overview")).toBeInTheDocument();
  });

  it("shows summary cards after loading", async () => {
    renderWithRouter(<OverviewPage />);
    await waitFor(() => {
      expect(screen.getByText("Apps")).toBeInTheDocument();
    });
    expect(screen.getByText("Protection")).toBeInTheDocument();
    expect(screen.getByText("Storage")).toBeInTheDocument();
  });

  it("shows attention items after loading", async () => {
    renderWithRouter(<OverviewPage />);
    await waitFor(() => {
      expect(screen.getByText("Needs Attention")).toBeInTheDocument();
    });
  });
});

describe("AppsPage", () => {
  it("renders page title", () => {
    renderWithRouter(<AppsPage />);
    expect(screen.getByText("Apps")).toBeInTheDocument();
  });

  it("shows app cards after loading", async () => {
    renderWithRouter(<AppsPage />);
    await waitFor(() => {
      expect(screen.getByText(/Nextcloud/)).toBeInTheDocument();
    });
    expect(screen.getByText(/Immich/)).toBeInTheDocument();
  });
});

describe("ProtectionPage", () => {
  it("renders page title", () => {
    renderWithRouter(<ProtectionPage />);
    expect(screen.getByText("Protection")).toBeInTheDocument();
  });

  it("shows protection rules after loading", async () => {
    renderWithRouter(<ProtectionPage />);
    await waitFor(() => {
      expect(screen.getByText("Security Rules")).toBeInTheDocument();
    });
    expect(screen.getByText("Inbound Firewall")).toBeInTheDocument();
  });
});

describe("ActivityPage", () => {
  it("renders page title", () => {
    renderWithRouter(<ActivityPage />);
    expect(screen.getByText("Activity")).toBeInTheDocument();
  });

  it("shows activity events after loading", async () => {
    renderWithRouter(<ActivityPage />);
    await waitFor(() => {
      expect(screen.getByText("Nextcloud update failed")).toBeInTheDocument();
    });
  });
});
