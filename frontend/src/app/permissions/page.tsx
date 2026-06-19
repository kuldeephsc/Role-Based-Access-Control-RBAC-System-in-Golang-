"use client";
import { useEffect, useState } from "react";
import Nav from "@/components/nav";
import { api } from "@/lib/api";

export default function PermissionsPage() {
  const [perms, setPerms] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [resource, setResource] = useState("");
  const [action, setAction] = useState("");
  const [desc, setDesc] = useState("");

  const load = () => {
    api.listPermissions().then((p) => setPerms(Array.isArray(p) ? p : []));
  };

  useEffect(load, []);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createPermission(name, resource, action, desc);
      setName(""); setResource(""); setAction(""); setDesc("");
      load();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const del = async (id: string) => {
    if (!confirm("Delete this permission?")) return;
    await api.deletePermission(id);
    load();
  };

  return (
    <>
      <Nav />
      <main className="p-6 max-w-5xl mx-auto">
        <h1 className="text-2xl font-bold mb-4">Permissions</h1>

        <form onSubmit={create} className="bg-white rounded-lg shadow p-4 mb-6 flex gap-3 items-end flex-wrap">
          <div>
            <label className="text-xs text-gray-500">Name</label>
            <input value={name} onChange={(e) => setName(e.target.value)} required className="block border rounded px-3 py-1.5 text-sm" placeholder="e.g. create_user" />
          </div>
          <div>
            <label className="text-xs text-gray-500">Resource</label>
            <input value={resource} onChange={(e) => setResource(e.target.value)} required className="block border rounded px-3 py-1.5 text-sm" placeholder="e.g. user" />
          </div>
          <div>
            <label className="text-xs text-gray-500">Action</label>
            <input value={action} onChange={(e) => setAction(e.target.value)} required className="block border rounded px-3 py-1.5 text-sm" placeholder="e.g. create" />
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
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3">Resource</th>
                <th className="px-4 py-3">Action</th>
                <th className="px-4 py-3">Description</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {perms.map((p) => (
                <tr key={p.ID} className="border-t">
                  <td className="px-4 py-3 font-mono text-xs">{p.Name}</td>
                  <td className="px-4 py-3">{p.Resource}</td>
                  <td className="px-4 py-3">{p.Action}</td>
                  <td className="px-4 py-3 text-gray-500">{p.Description}</td>
                  <td className="px-4 py-3">
                    <button onClick={() => del(p.ID)} className="text-red-500 hover:underline text-xs">Delete</button>
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
