"use client";
import { useEffect, useState } from "react";
import Nav from "@/components/nav";
import { api } from "@/lib/api";

export default function UsersPage() {
  const [users, setUsers] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [roles, setRoles] = useState<any[]>([]);
  const [selectedRole, setSelectedRole] = useState<Record<string, string>>({});
  const limit = 20;

  const load = () => {
    api.listUsers(limit, page * limit).then((d) => {
      setUsers(d.users || []);
      setTotal(d.total);
    });
    api.listRoles().then((r) => setRoles(Array.isArray(r) ? r : []));
  };

  useEffect(load, [page]);

  const assign = async (userId: string) => {
    const roleId = selectedRole[userId];
    if (!roleId) return;
    try {
      await api.assignRole(userId, roleId);
      alert("Role assigned");
    } catch (e: any) {
      alert(e.message);
    }
  };

  return (
    <>
      <Nav />
      <main className="p-6 max-w-5xl mx-auto">
        <h1 className="text-2xl font-bold mb-4">Users ({total})</h1>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-left">
              <tr>
                <th className="px-4 py-3">Email</th>
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3">Active</th>
                <th className="px-4 py-3">Assign Role</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.ID} className="border-t">
                  <td className="px-4 py-3 font-mono text-xs">{u.Email}</td>
                  <td className="px-4 py-3">{u.FullName}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 rounded text-xs ${u.IsActive ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}>
                      {u.IsActive ? "Yes" : "No"}
                    </span>
                  </td>
                  <td className="px-4 py-3 flex gap-2">
                    <select
                      value={selectedRole[u.ID] || ""}
                      onChange={(e) => setSelectedRole({ ...selectedRole, [u.ID]: e.target.value })}
                      className="border rounded px-2 py-1 text-xs"
                    >
                      <option value="">Select role</option>
                      {roles.map((r: any) => (
                        <option key={r.ID} value={r.ID}>{r.Name}</option>
                      ))}
                    </select>
                    <button onClick={() => assign(u.ID)} className="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700">
                      Assign
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="flex gap-2 mt-4 justify-center">
          <button disabled={page === 0} onClick={() => setPage(page - 1)} className="px-3 py-1 border rounded disabled:opacity-30">Prev</button>
          <span className="px-3 py-1 text-sm">Page {page + 1}</span>
          <button disabled={(page + 1) * limit >= total} onClick={() => setPage(page + 1)} className="px-3 py-1 border rounded disabled:opacity-30">Next</button>
        </div>
      </main>
    </>
  );
}
