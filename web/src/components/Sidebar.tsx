import { NavLink } from "react-router";
import { useAuth } from "../queries/auth";
import { useTheme } from "../hooks/useTheme";

const navItems = [
  { to: "/", label: "Dashboard" },
  { to: "/media", label: "Media" },
  { to: "/rules", label: "Rules" },
  { to: "/actions", label: "Actions" },
  { to: "/connections", label: "Connections" },
  { to: "/settings", label: "Settings" },
];

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { logout, user } = useAuth();
  const { theme, toggleTheme } = useTheme();

  return (
    <>
      {isOpen && (
        <div className="fixed inset-0 z-20 bg-black/50 lg:hidden" onClick={onClose} />
      )}

      <aside
        className={`fixed left-0 top-0 z-30 flex h-full w-64 flex-col bg-white shadow-lg transition-transform dark:bg-gray-800 lg:static lg:translate-x-0 ${
          isOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex h-16 items-center border-b border-gray-200 px-4 dark:border-gray-700">
          <span className="text-lg font-bold text-gray-900 dark:text-white">Media Reaper</span>
        </div>

        <nav className="flex-1 space-y-1 p-4">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              onClick={onClose}
              className={({ isActive }) =>
                `block rounded px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200"
                    : "text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-gray-200 p-4 dark:border-gray-700">
          <button
            onClick={toggleTheme}
            className="mb-2 w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
          >
            {theme === "dark" ? "Light mode" : "Dark mode"}
          </button>
          <div className="mb-2 truncate px-3 text-sm text-gray-500 dark:text-gray-400">
            {user?.username}
          </div>
          <button
            onClick={() => logout()}
            className="w-full rounded px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
          >
            Sign out
          </button>
        </div>
      </aside>
    </>
  );
}
