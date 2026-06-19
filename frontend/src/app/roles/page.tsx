"use client";
import { useEffect, useState } from "react";
import Nav from "@/components/nav";
import { api } from "@/lib/api";

export default function RolesPage() {
  const [roles, setRoles] = useState<any[]>([]);
  const [perms, setPerms] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [desc, setDesc] = useState("");
  const [attachState, setAttachState] = useState<Record<string, string>>({});

  const load = () => {
    api.listRoles().then((r) => setRoles(Array.isArray(r) ? r : []));
    api.listPermissions().then((p) => setPerms(Array.isArray(p) ? p : []));
  };

  useEffect(load, []);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createRole(name, desc);
      setName("");
      setDesc("");
      load();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const del = async (id: string) => {
    if (!confirm("Delete this role?")) return;
    await api.deleteRole(id);
    load();
  };

  const attach = async (roleId: string) => {
    const permId = attachState[roleId];
    if (!permId) return;
    try {
      await api.attachPermission(roleId, permId);
      alert("Permission attached");
    } catch (err: any) {
      alert(err.message);
    }
  };

  return (
    <>
      <Nav />
      <main className="p-6 max-w-5xl mx-auto">
        <h1 className="text-2xl font-bold mb-4">Roles</h1>

        <form onSubmit={create} className="bg-white rounded-lg shadow p-4 mb-6 flex gap-3 items-end">
          <div>
            <label className="text-xs text-gray-500">Name</label>
            <input value={name} onChange={(e) => setName(e.target.value)} required className="block border rounded px-3 py-1.5 text-sm" />
          </div>
          <div>
            <label className="text-xs text-gray-500">Description</label>
            <input value={desc} onChange={(e) => setDesc(e.target.value)} className="block border rounded px-3 py-1.5 text-sm" />
          </div>
          <button type="submit" className="px-4 py-1.5 bg-green-600 text-white rounded text-sm hover:bg-green-700">Create</button>
        </form>

        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-left">
              <tr>
                <th className="px-4 py-3">Role</th>
                <th className="px-4 py-3">Description</th>
                <th className="px-4 py-3">Attach Permission</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {roles.map((r) => (
                <tr key={r.ID} className="border-t">
                  <td className="px-4 py-3 font-medium">{r.Name}</td>
                  <td className="px-4 py-3 text-gray-500">{r.Description}</td>
                  <td className="px-4 py-3 flex gap-2">
                    <select
                      value={attachState[r.ID] || ""}
                      onChange={(e) => setAttachState({ ...attachState, [r.ID]: e.target.value })}
                      className="border rounded px-2 py-1 text-xs"
                    >
                      <option value="">Select permission</option>
                      {perms.map((p: any) => (
                        <option key={p.ID} value={p.ID}>{p.Name}</option>
                      ))}
                    </select>
                    <button onClick={() => attach(r.ID)} className="px-2 py-1 bg-blue-600 text-white rounded text-xs">Attach</button>
                  </td>
                  <td className="px-4 py-3">
                    <button onClick={() => del(r.ID)} className="text-red-500 hover:underline text-xs">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </main>
    </>
  );
}
