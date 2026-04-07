import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import api from "../lib/api";
import { Board } from "../types";
import { Plus, Zap, LogOut, Loader2, Layout } from "lucide-react";

export default function BoardsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [boards, setBoards] = useState<Board[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: "", description: "" });
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    api.get("/boards")
      .then((res) => setBoards(res.data))
      .finally(() => setLoading(false));
  }, []);

  async function createBoard(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    const res = await api.post("/boards", form);
    setBoards((b) => [res.data, ...b]);
    setShowForm(false);
    setForm({ name: "", description: "" });
    setCreating(false);
  }

  return (
    <div className="min-h-screen bg-gray-950">
      <nav className="border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2 font-bold text-xl text-white">
          <Zap className="h-5 w-5 text-violet-500" />
          TaskFlow
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-400">{user?.name}</span>
          <button onClick={() => { logout(); navigate("/login"); }} className="text-gray-400 hover:text-white transition">
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </nav>

      <main className="max-w-6xl mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold text-white">Your Boards</h1>
          <button
            onClick={() => setShowForm(true)}
            className="bg-violet-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-violet-700 transition flex items-center gap-2"
          >
            <Plus className="h-4 w-4" /> New Board
          </button>
        </div>

        {showForm && (
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 mb-6">
            <h3 className="font-semibold text-white mb-4">Create Board</h3>
            <form onSubmit={createBoard} className="space-y-4">
              <input
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                placeholder="Board name"
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                required
              />
              <input
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                placeholder="Description (optional)"
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
              />
              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={creating}
                  className="bg-violet-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-violet-700 disabled:opacity-50 flex items-center gap-2"
                >
                  {creating && <Loader2 className="h-4 w-4 animate-spin" />}
                  Create
                </button>
                <button type="button" onClick={() => setShowForm(false)} className="px-4 py-2 text-sm text-gray-400 hover:text-white">
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {loading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
          </div>
        ) : boards.length === 0 ? (
          <div className="text-center py-24">
            <Layout className="h-12 w-12 text-gray-700 mx-auto mb-4" />
            <p className="text-gray-400">No boards yet. Create your first one.</p>
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-4">
            {boards.map((board) => (
              <div
                key={board.id}
                onClick={() => navigate(`/board/${board.id}`)}
                className="bg-gray-900 border border-gray-800 rounded-xl p-5 hover:border-violet-500 cursor-pointer transition group"
              >
                <div className="w-8 h-8 bg-violet-600 rounded-lg flex items-center justify-center mb-4 text-white font-bold text-sm">
                  {board.name[0]}
                </div>
                <h3 className="font-semibold text-white mb-1 group-hover:text-violet-400 transition">{board.name}</h3>
                {board.description && (
                  <p className="text-sm text-gray-500 line-clamp-2">{board.description}</p>
                )}
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
