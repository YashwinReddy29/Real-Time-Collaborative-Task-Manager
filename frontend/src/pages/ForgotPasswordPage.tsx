import { useState } from "react";
import { Link } from "react-router-dom";
import { Zap, Loader2, ArrowLeft, CheckCircle } from "lucide-react";
import api from "../lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await api.post("/auth/forgot-password", { email });
      setSent(true);
    } catch {
      setError("Something went wrong. Please try again.");
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

        {sent ? (
          <div className="text-center py-4">
            <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
            <h2 className="text-white font-semibold text-lg mb-2">Check your terminal</h2>
            <p className="text-gray-400 text-sm mb-6">
              A reset link has been printed to the server terminal.
            </p>
            <Link to="/login" className="text-violet-400 hover:underline text-sm">
              Back to sign in
            </Link>
          </div>
        ) : (
          <>
            <h1 className="text-2xl font-bold text-white mb-2">Forgot password</h1>
            <p className="text-gray-400 text-sm mb-6">
              Enter your email and we'll send a reset link.
            </p>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-300 block mb-1">Email</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                  placeholder="you@example.com"
                  required
                />
              </div>
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <button
                type="submit"
                disabled={loading}
                className="w-full bg-violet-600 text-white py-2 rounded-lg font-medium hover:bg-violet-700 transition disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                Send reset link
              </button>
            </form>
            <Link
              to="/login"
              className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-300 mt-6 transition"
            >
              <ArrowLeft className="h-3 w-3" /> Back to sign in
            </Link>
          </>
        )}
      </div>
    </div>
  );
}
