import { useAuth } from "../contexts/AuthContext";
import Layout from "../components/Layout";

export default function Dashboard() {
  const { user } = useAuth();

  return (
    <Layout>
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <StatCard label="Email"    value={user?.email ?? "—"} />
          <StatCard label="Provider" value={user?.provider ?? "—"} />
          <StatCard label="Status"   value={user?.paid_at ? "Paid" : "Free"} highlight={!!user?.paid_at} />
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold mb-4">Your Profile</h2>
          <div className="flex items-center gap-4">
            {user?.avatar_url && (
              <img src={user.avatar_url} alt="" className="h-16 w-16 rounded-full" />
            )}
            <div>
              <p className="font-medium text-gray-900">{user?.name || "(no name)"}</p>
              <p className="text-sm text-gray-500">{user?.email}</p>
              <span className={`inline-block mt-1 text-xs px-2 py-0.5 rounded-full font-medium ${
                user?.role === "admin"
                  ? "bg-purple-100 text-purple-700"
                  : "bg-gray-100 text-gray-600"
              }`}>
                {user?.role}
              </span>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}

function StatCard({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-5">
      <p className="text-xs text-gray-500 uppercase tracking-wide mb-1">{label}</p>
      <p className={`text-lg font-semibold ${highlight ? "text-green-600" : "text-gray-900"}`}>
        {value}
      </p>
    </div>
  );
}
