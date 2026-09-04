import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { StatusBadge } from "@/components/StatusBadge";
import { AttentionCard } from "@/components/AttentionCard";
import { AppCard } from "@/components/AppCard";
import { DataGuard } from "@/components/LoadingState";
import { mockApps, mockAttentionItems } from "@/api/mock-data";

describe("StatusBadge", () => {
  it("renders with healthy status", () => {
    render(<StatusBadge status="healthy" />);
    expect(screen.getByRole("status")).toHaveTextContent("Healthy");
  });

  it("renders with error status", () => {
    render(<StatusBadge status="error" />);
    expect(screen.getByRole("status")).toHaveTextContent("Error");
  });

  it("renders with armed protection status", () => {
    render(<StatusBadge status="armed" />);
    expect(screen.getByRole("status")).toHaveTextContent("Armed");
  });
});

describe("AttentionCard", () => {
  it("renders attention item with correct severity", () => {
    render(
      <MemoryRouter>
        <AttentionCard item={mockAttentionItems[0]} />
      </MemoryRouter>,
    );
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText(mockAttentionItems[0].title)).toBeInTheDocument();
  });

  it("renders action link when provided", () => {
    render(
      <MemoryRouter>
        <AttentionCard item={mockAttentionItems[0]} />
      </MemoryRouter>,
    );
    const link = screen.getByText("View Details");
    expect(link).toBeInTheDocument();
    expect(link.closest("a")).toHaveAttribute("href", "/apps/nextcloud");
  });
});

describe("AppCard", () => {
  it("renders app name and version", () => {
    const app = mockApps[0];
    const onSelect = () => {};
    render(<AppCard app={app} onClick={onSelect} />);
    expect(screen.getByText(app.name)).toBeInTheDocument();
    expect(screen.getByText(`v${app.version}`)).toBeInTheDocument();
  });

  it("shows update indicator when update available", () => {
    const app = mockApps[0];
    const onSelect = () => {};
    render(<AppCard app={app} onClick={onSelect} />);
    expect(screen.getByLabelText("Update available")).toBeInTheDocument();
  });

  it("has accessible button role", () => {
    const app = mockApps[0];
    const onSelect = () => {};
    render(<AppCard app={app} onClick={onSelect} />);
    expect(screen.getByRole("button")).toBeInTheDocument();
  });
});

describe("DataGuard", () => {
  it("shows loading state", () => {
    render(
      <DataGuard state="loading" error={null}>
        <div>Content</div>
      </DataGuard>,
    );
    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText("Content")).not.toBeInTheDocument();
  });

  it("shows error state", () => {
    render(
      <DataGuard state="error" error="Network failure">
        <div>Content</div>
      </DataGuard>,
    );
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText("Network failure")).toBeInTheDocument();
  });

  it("shows children on success", () => {
    render(
      <DataGuard state="success" error={null}>
        <div>Content</div>
      </DataGuard>,
    );
    expect(screen.getByText("Content")).toBeInTheDocument();
  });
});
