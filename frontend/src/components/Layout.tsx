import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export default function Layout({ children }: { children: React.ReactNode }) {
  const { user, impersonatedUser, isAdmin, logout, exitAssume } = useAuth();
  const nav = useNavigate();

  const handleLogout = () => {
    logout();
    nav("/");
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Impersonation banner */}
      {impersonatedUser && (
        <div className="bg-amber-400 text-amber-900 px-4 py-2 text-sm flex items-center justify-between">
          <span>
            Acting as <strong>{impersonatedUser.name || impersonatedUser.email}</strong>
          </span>
          <button onClick={exitAssume} className="underline font-medium">
            Exit impersonation
          </button>
        </div>
      )}

      <nav className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/dashboard" className="text-lg font-semibold text-indigo-600">
            AppTemplate
          </Link>
          <Link to="/dashboard" className="text-sm text-gray-600 hover:text-gray-900">
            Dashboard
          </Link>
          <Link to="/agent" className="text-sm text-gray-600 hover:text-gray-900">
            Agent
          </Link>
          <Link to="/payment" className="text-sm text-gray-600 hover:text-gray-900">
            Payment
          </Link>
          {isAdmin() && (
            <Link to="/admin" className="text-sm text-gray-600 hover:text-gray-900">
              Admin
            </Link>
          )}
        </div>
        <div className="flex items-center gap-3">
          {user?.avatar_url && (
            <img src={user.avatar_url} alt="" className="h-8 w-8 rounded-full" />
          )}
          <span className="text-sm text-gray-700">{user?.name || user?.email}</span>
          <button
            onClick={handleLogout}
            className="text-sm text-red-500 hover:text-red-700"
          >
            Logout
          </button>
        </div>
      </nav>

      <main className="flex-1 px-6 py-8 max-w-6xl mx-auto w-full">{children}</main>
    </div>
  );
}
