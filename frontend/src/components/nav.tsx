"use client";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

const links = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/users", label: "Users" },
  { href: "/roles", label: "Roles" },
  { href: "/permissions", label: "Permissions" },
  { href: "/audit", label: "Audit Log" },
];

export default function Nav() {
  const path = usePathname();
  const router = useRouter();

  const logout = () => {
    const rt = localStorage.getItem("refresh_token");
    fetch("/api/v1/auth/logout", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("access_token")}`,
      },
      body: JSON.stringify({ refresh_token: rt }),
    }).finally(() => {
      localStorage.clear();
      router.push("/login");
    });
  };

  return (
    <nav className="bg-white border-b px-6 py-3 flex items-center justify-between">
      <div className="flex items-center gap-1">
        <span className="font-bold text-lg mr-6">RBAC</span>
        {links.map((l) => (
          <Link
            key={l.href}
            href={l.href}
            className={`px-3 py-1.5 rounded text-sm ${
              path === l.href
                ? "bg-blue-100 text-blue-700 font-medium"
                : "text-gray-600 hover:bg-gray-100"
            }`}
          >
            {l.label}
          </Link>
        ))}
      </div>
      <button
        onClick={logout}
        className="text-sm text-gray-500 hover:text-red-600"
      >
        Logout
      </button>
    </nav>
  );
}
