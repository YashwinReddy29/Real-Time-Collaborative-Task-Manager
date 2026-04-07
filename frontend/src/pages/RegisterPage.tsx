import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Zap, Loader2 } from "lucide-react";

export default function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ name: "", email: "", password: "" });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await register(form.name, form.email, form.password);
      navigate("/");
    } catch {
      setError("Email already in use");
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
        <h1 className="text-2xl font-bold text-white mb-2">Create account</h1>
        <p className="text-gray-400 text-sm mb-6">Start collaborating today</p>
        <form onSubmit={handleSubmit} className="space-y-4">
          {[
            { key: "name", label: "Full name", type: "text", placeholder: "Jane Smith" },
            { key: "email", label: "Email", type: "email", placeholder: "you@example.com" },
            { key: "password", label: "Password", type: "password", placeholder: "Min 6 characters" },
          ].map(({ key, label, type, placeholder }) => (
            <div key={key}>
              <label className="text-sm font-medium text-gray-300 block mb-1">{label}</label>
              <input
                type={type}
                value={form[key as keyof typeof form]}
                onChange={(e) => setForm((f) => ({ ...f, [key]: e.target.value }))}
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                placeholder={placeholder}
                required
              />
            </div>
          ))}
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-violet-600 text-white py-2 rounded-lg font-medium hover:bg-violet-700 transition disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Create account
          </button>
        </form>
        <p className="text-center text-sm text-gray-500 mt-6">
          Have an account?{" "}
          <Link to="/login" className="text-violet-400 hover:underline">Sign in</Link>
        </p>
      </div>
    </div>
  );
}
