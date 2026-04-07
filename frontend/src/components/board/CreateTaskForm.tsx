import { useState } from "react";
import { Task } from "../../types";
import api from "../../lib/api";
import { Loader2 } from "lucide-react";

interface Props {
  boardId: string;
  columnId: string;
  position: number;
  onCreated: (task: Task) => void;
  onCancel: () => void;
}

export function CreateTaskForm({ boardId, columnId, position, onCreated, onCancel }: Props) {
  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState("medium");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    setLoading(true);
    const res = await api.post(`/boards/${boardId}/tasks`, {
      column_id: columnId,
      title,
      priority,
      position,
    });
    onCreated(res.data);
    setLoading(false);
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Task title"
        className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-violet-500"
        autoFocus
        required
      />
      <select
        value={priority}
        onChange={(e) => setPriority(e.target.value)}
        className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none"
      >
        <option value="low">Low priority</option>
        <option value="medium">Medium priority</option>
        <option value="high">High priority</option>
      </select>
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={loading}
          className="bg-violet-600 text-white px-3 py-1.5 rounded-lg text-xs font-medium hover:bg-violet-700 disabled:opacity-50 flex items-center gap-1"
        >
          {loading && <Loader2 className="h-3 w-3 animate-spin" />}
          Add
        </button>
        <button type="button" onClick={onCancel} className="text-xs text-gray-400 hover:text-white px-3 py-1.5">
          Cancel
        </button>
      </div>
    </form>
  );
}
