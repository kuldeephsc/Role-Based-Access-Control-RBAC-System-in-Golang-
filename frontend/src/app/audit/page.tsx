"use client";
import { useEffect, useState } from "react";
import Nav from "@/components/nav";
import { api } from "@/lib/api";

export default function AuditPage() {
  const [logs, setLogs] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const limit = 50;

  useEffect(() => {
    api.listAudit(limit, page * limit).then((d) => {
      setLogs(d.logs || []);
      setTotal(d.total);
    }).catch(() => {});
  }, [page]);

  return (
    <>
      <Nav />
      <main className="p-6 max-w-5xl mx-auto">
        <h1 className="text-2xl font-bold mb-4">Audit Log ({total})</h1>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-left">
              <tr>
                <th className="px-4 py-3">Time</th>
                <th className="px-4 py-3">Action</th>
                <th className="px-4 py-3">Actor</th>
                <th className="px-4 py-3">Details</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((l) => (
                <tr key={l.ID} className="border-t">
                  <td className="px-4 py-3 text-xs text-gray-500 whitespace-nowrap">
                    {new Date(l.CreatedAt).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs">{l.Action}</td>
                  <td className="px-4 py-3 font-mono text-xs">{l.ActorID ? l.ActorID.slice(0, 8) + "..." : "system"}</td>
                  <td className="px-4 py-3 text-xs text-gray-500 max-w-xs truncate">
                    {l.Metadata ? JSON.stringify(l.Metadata) : ""}
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
