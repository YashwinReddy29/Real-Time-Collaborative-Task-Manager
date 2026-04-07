import { useState } from "react";
import { X, UserPlus, Loader2, CheckCircle } from "lucide-react";
import api from "../../lib/api";

interface Props {
  boardId: string;
  onClose: () => void;
  onInvited: (user: any) => void;
}

export function InviteModal({ boardId, onClose, onInvited }: Props) {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function handleInvite(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    setSuccess("");

    try {
      const res = await api.post(`/boards/${boardId}/invite`, { email });
      setSuccess(`${res.data.user.name} added to the board!`);
      onInvited(res.data.user);
      setEmail("");
    } catch (err: any) {
      setError(err.response?.data?.error ?? "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="bg-gray-900 border border-gray-800 rounded-2xl w-full max-w-md shadow-2xl">
        <div className="flex items-center justify-between p-5 border-b border-gray-800">
          <div className="flex items-center gap-2">
            <UserPlus className="h-5 w-5 text-violet-400" />
            <h2 className="font-semibold text-white">Invite to board</h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="p-5">
          <p className="text-sm text-gray-400 mb-4">
            Invite someone by their registered email address.
          </p>

          <form onSubmit={handleInvite} className="space-y-3">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="colleague@example.com"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
              required
              autoFocus
            />

            {error && (
              <p className="text-red-400 text-sm flex items-center gap-2">
                ⚠ {error}
              </p>
            )}

            {success && (
              <p className="text-green-400 text-sm flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                {success}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-violet-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-violet-700 disabled:opacity-50 flex items-center justify-center gap-2 transition"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              Send Invite
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
