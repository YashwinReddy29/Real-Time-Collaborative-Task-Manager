import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import { Zap, Loader2, CheckCircle, XCircle } from "lucide-react";
import api from "../lib/api";

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get("token");

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) navigate("/login");
  }, [token, navigate]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (password !== confirm) {
      setError("Passwords do not match");
      return;
    }
    setLoading(true);
    setError("");
    try {
      await api.post("/auth/reset-password", { token, password });
      setSuccess(true);
      setTimeout(() => navigate("/login"), 2000);
    } catch (err: any) {
      setError(err.response?.data?.error ?? "Invalid or expired token");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
      <div className="bg-gray-900 border border-gray-800 rounded-2xl p-8 w-full max-w-sm">
        <div className="flex items-center gap-2 font-bold text-xl text-white mb-8">
          <Zap className="h-5 w-5 text-violet-500" />
          TaskFlow
        </div>

        {success ? (
          <div className="text-center py-4">
            <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
            <h2 className="text-white font-semibold text-lg mb-2">Password updated!</h2>
            <p className="text-gray-400 text-sm">Redirecting to sign in...</p>
          </div>
        ) : (
          <>
            <h1 className="text-2xl font-bold text-white mb-2">Reset password</h1>
            <p className="text-gray-400 text-sm mb-6">Enter your new password below.</p>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-300 block mb-1">New password</label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                  placeholder="Min 6 characters"
                  required
                  minLength={6}
                />
              </div>
              <div>
                <label className="text-sm font-medium text-gray-300 block mb-1">Confirm password</label>
                <input
                  type="password"
                  value={confirm}
                  onChange={(e) => setConfirm(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                  placeholder="Repeat password"
                  required
                />
              </div>
              {error && (
                <p className="text-red-400 text-sm flex items-center gap-2">
                  <XCircle className="h-4 w-4" /> {error}
                </p>
              )}
              <button
                type="submit"
                disabled={loading}
                className="w-full bg-violet-600 text-white py-2 rounded-lg font-medium hover:bg-violet-700 transition disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                Update password
              </button>
            </form>
            <Link to="/login" className="block text-center text-sm text-gray-500 hover:text-gray-300 mt-6">
              Back to sign in
            </Link>
          </>
        )}
      </div>
    </div>
  );
}
