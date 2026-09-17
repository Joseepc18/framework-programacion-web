import { Link } from "react-router-dom";
import { Home, Package, Network } from "lucide-react";

interface SidebarProps {
  isCollapsed: boolean;
  onNavigate: () => void;
}

const Sidebar = ({ isCollapsed, onNavigate }: SidebarProps) => {
  const menuItems = [
    { to: "/", label: "Dashboard", Icon: Home },
    { to: "/catalogo", label: "Catálogo", Icon: Package },
    { to: "/mi-red", label: "Mi Red", Icon: Network },
  ];

  return (
    <aside
      className={`fixed inset-y-0 left-0 z-40 w-64 bg-green-500 text-white flex flex-col transform transition-all duration-300
      ${isCollapsed ? "translate-x-0" : "-translate-x-full"}
      md:static md:z-auto md:translate-x-0 ${isCollapsed ? "md:w-20" : "md:w-64"}`}
    >
      <div className="p-6 text-2xl font-bold border-b border-slate-700 text-center whitespace-nowrap overflow-hidden">
        <span className="md:hidden">MultiCatálogo</span>
        <span className="hidden md:inline">{isCollapsed ? "MC" : "MultiCatálogo"}</span>
      </div>

      <nav className="flex-1 p-4 space-y-2">
        {menuItems.map((item) => (
          <Link
            key={item.to}
            to={item.to}
            title={item.label}
            onClick={onNavigate}
            className={`flex items-center gap-3 p-3 rounded hover:bg-slate-800 transition ${
              isCollapsed ? "md:justify-center md:gap-0" : ""
            }`}
          >
            <item.Icon className="w-6 h-6 shrink-0" />
            <span className={isCollapsed ? "md:hidden" : ""}>{item.label}</span>
          </Link>
        ))}
      </nav>
    </aside>
  );
};

export default Sidebar;
