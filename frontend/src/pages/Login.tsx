import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { oauthURL } from "../api/client";
import { useAuth } from "../contexts/AuthContext";

const PROVIDERS = [
  { id: "google",   label: "Continue with Google",   bg: "bg-white border border-gray-300 text-gray-700 hover:bg-gray-50" },
  { id: "github",   label: "Continue with GitHub",   bg: "bg-gray-900 text-white hover:bg-gray-800" },
  { id: "slack",    label: "Continue with Slack",    bg: "bg-[#4A154B] text-white hover:bg-[#3e1240]" },
  { id: "linkedin", label: "Continue with LinkedIn", bg: "bg-[#0077B5] text-white hover:bg-[#006396]" },
];

export default function Login() {
  const { user, login } = useAuth();
  const nav = useNavigate();

  // Handle JWT delivered via URL fragment after OAuth callback redirect.
  useEffect(() => {
    const hash = window.location.hash;
    const match = hash.match(/token=([^&]+)/);
    if (match) {
      window.history.replaceState(null, "", window.location.pathname);
      login(match[1]).then(() => nav("/dashboard"));
    }
  }, [login, nav]);

  useEffect(() => {
    if (user) nav("/dashboard");
  }, [user, nav]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 to-purple-50">
      <div className="bg-white rounded-2xl shadow-lg p-10 w-full max-w-sm">
        <h1 className="text-2xl font-bold text-center text-gray-900 mb-2">Welcome</h1>
        <p className="text-sm text-center text-gray-500 mb-8">Sign in to continue</p>

        <div className="flex flex-col gap-3">
          {PROVIDERS.map((p) => (
            <a
              key={p.id}
              href={oauthURL(p.id)}
              className={`flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors ${p.bg}`}
            >
              {p.label}
            </a>
          ))}
        </div>
      </div>
    </div>
  );
}
