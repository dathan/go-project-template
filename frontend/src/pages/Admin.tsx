import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { listUsers, type User } from "../api/client";
import { useAuth } from "../contexts/AuthContext";
import Layout from "../components/Layout";

export default function Admin() {
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { assume } = useAuth();
  const nav = useNavigate();

  useEffect(() => {
    listUsers()
      .then(({ users, total }) => { setUsers(users); setTotal(total); })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const handleAssume = async (id: string) => {
    try {
      await assume(id);
      nav("/dashboard");
    } catch (e: unknown) {
      alert((e as Error).message);
    }
  };

  return (
    <Layout>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Admin — Users</h1>
          <span className="text-sm text-gray-500">{total} total</span>
        </div>

        {loading && <p className="text-gray-500">Loading…</p>}
        {error && <p className="text-red-500">{error}</p>}

        {!loading && !error && (
          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  {["Name", "Email", "Provider", "Role", "Paid", "Joined", "Actions"].map((h) => (
                    <th key={h} className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {users.map((u) => (
                  <tr key={u.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 flex items-center gap-2">
                      {u.avatar_url && <img src={u.avatar_url} alt="" className="h-6 w-6 rounded-full" />}
                      <span className="font-medium text-gray-900">{u.name || "—"}</span>
                    </td>
                    <td className="px-4 py-3 text-gray-600">{u.email}</td>
                    <td className="px-4 py-3 text-gray-600 capitalize">{u.provider}</td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                        u.role === "admin" ? "bg-purple-100 text-purple-700" : "bg-gray-100 text-gray-600"
                      }`}>
                        {u.role}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      {u.paid_at
                        ? <span className="text-green-600 font-medium">Paid</span>
                        : <span className="text-gray-400">Free</span>}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleAssume(u.id)}
                        className="text-xs text-indigo-600 hover:text-indigo-800 font-medium underline"
                      >
                        Assume role
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </Layout>
  );
}
