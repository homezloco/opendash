import { NavLink } from "react-router-dom";

interface NavItemProps {
  to: string;
  label: string;
  icon: React.ReactNode;
}

export function NavItem({ to, label, icon }: NavItemProps) {
  return (
    <NavLink
      to={to}
      className="nav-item"
      style={({ isActive }) => ({
        display: "flex",
        alignItems: "center",
        gap: "var(--space-3)",
        padding: "var(--space-2) var(--space-4)",
        borderRadius: "var(--radius-md)",
        fontSize: "var(--text-sm)",
        fontWeight: isActive
          ? ("var(--weight-semibold)" as unknown as number)
          : ("var(--weight-medium)" as unknown as number),
        color: isActive
          ? "var(--color-accent-primary)"
          : "var(--color-text-secondary)",
        backgroundColor: isActive
          ? "var(--color-accent-primary-subtle)"
          : "transparent",
        transition: "background-color var(--transition-fast), color var(--transition-fast)",
        textDecoration: "none",
      })}
      aria-current={undefined}
    >
      <span
        style={{
          width: 20,
          height: 20,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          flexShrink: 0,
        }}
      >
        {icon}
      </span>
      <span>{label}</span>
    </NavLink>
  );
}
