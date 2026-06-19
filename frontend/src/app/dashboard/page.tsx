"use client";
import { useEffect, useState } from "react";
import Nav from "@/components/nav";
import { api } from "@/lib/api";

export default function DashboardPage() {
  const [stats, setStats] = useState({ users: 0, roles: 0, permissions: 0, auditLogs: 0 });

  useEffect(() => {
    Promise.all([
      api.listUsers(1, 0).catch(() => ({ total: 0 })),
      api.listRoles().catch(() => []),
      api.listPermissions().catch(() => []),
      api.listAudit(1, 0).catch(() => ({ total: 0 })),
    ]).then(([u, r, p, a]) => {
      setStats({
        users: (u as any).total || 0,
        roles: Array.isArray(r) ? r.length : 0,
        permissions: Array.isArray(p) ? p.length : 0,
        auditLogs: (a as any).total || 0,
      });
    });
  }, []);

  const cards = [
    { label: "Users", value: stats.users, color: "bg-blue-500" },
    { label: "Roles", value: stats.roles, color: "bg-green-500" },
    { label: "Permissions", value: stats.permissions, color: "bg-purple-500" },
    { label: "Audit Events", value: stats.auditLogs, color: "bg-orange-500" },
  ];

  return (
    <>
      <Nav />
      <main className="p-6 max-w-5xl mx-auto">
        <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {cards.map((c) => (
            <div key={c.label} className="bg-white rounded-lg shadow p-5">
              <div className={`w-10 h-10 rounded ${c.color} mb-3`} />
              <p className="text-2xl font-bold">{c.value}</p>
              <p className="text-sm text-gray-500">{c.label}</p>
            </div>
          ))}
        </div>
      </main>
    </>
  );
}
