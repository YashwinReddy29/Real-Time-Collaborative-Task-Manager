import { useState } from "react";
import { Task, Column } from "../../types";
import api from "../../lib/api";
import { X, Flag, Loader2, Trash2 } from "lucide-react";

interface Props {
  task: Task;
  columns: Column[];
  boardId: string;
  onClose: () => void;
  onUpdate: (task: Task) => void;
  onDelete: (taskId: string) => void;
}

const priorityColors = {
  low: "bg-blue-900/40 text-blue-400 border-blue-800",
  medium: "bg-yellow-900/40 text-yellow-400 border-yellow-800",
  high: "bg-red-900/40 text-red-400 border-red-800",
};

export function TaskModal({ task, columns, boardId, onClose, onUpdate, onDelete }: Props) {
  const [form, setForm] = useState({
    title: task.title,
    description: task.description,
    priority: task.priority,
    column_id: task.column_id,
  });
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  async function save() {
    setSaving(true);
    try {
      const res = await api.put(`/boards/${boardId}/tasks/${task.id}`, {
        title: form.title,
        description: form.description,
        priority: form.priority,
      });

      // Move if column changed
      if (form.column_id !== task.column_id) {
        const moveRes = await api.patch(`/boards/${boardId}/tasks/${task.id}/move`, {
          column_id: form.column_id,
          position: 0,
        });
        onUpdate(moveRes.data);
      } else {
        onUpdate(res.data);
      }
      onClose();
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    setDeleting(true);
    await api.delete(`/boards/${boardId}/tasks/${task.id}`);
    onDelete(task.id);
    onClose();
  }

  return (
    <div
      className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="bg-gray-900 border border-gray-800 rounded-2xl w-full max-w-lg shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between p-5 border-b border-gray-800">
          <h2 className="font-semibold text-white">Edit Task</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition">
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-5 space-y-4">
          <div>
            <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-2">
              Title
            </label>
            <input
              value={form.title}
              onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-2">
              Description
            </label>
            <textarea
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              rows={4}
              placeholder="Add a description..."
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 resize-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-2">
                Priority
              </label>
              <div className="flex gap-2">
                {(["low", "medium", "high"] as const).map((p) => (
                  <button
                    key={p}
                    onClick={() => setForm((f) => ({ ...f, priority: p }))}
                    className={`flex-1 py-1.5 rounded-lg text-xs font-medium border transition capitalize ${
                      form.priority === p
                        ? priorityColors[p]
                        : "bg-gray-800 border-gray-700 text-gray-500 hover:text-gray-300"
                    }`}
                  >
                    <Flag className="h-3 w-3 inline mr-1" />
                    {p}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-2">
                Column
              </label>
              <select
                value={form.column_id}
                onChange={(e) => setForm((f) => ({ ...f, column_id: e.target.value }))}
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
              >
                {columns.map((col) => (
                  <option key={col.id} value={col.id}>{col.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="text-xs text-gray-600">
            Created {new Date(task.created_at).toLocaleDateString()}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-5 border-t border-gray-800">
          <button
            onClick={handleDelete}
            disabled={deleting}
            className="flex items-center gap-2 text-sm text-red-400 hover:text-red-300 transition disabled:opacity-50"
          >
            {deleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
            Delete task
          </button>
          <div className="flex gap-3">
            <button
              onClick={onClose}
              className="text-sm text-gray-400 hover:text-white px-4 py-2 transition"
            >
              Cancel
            </button>
            <button
              onClick={save}
              disabled={saving}
              className="bg-violet-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-violet-700 disabled:opacity-50 flex items-center gap-2"
            >
              {saving && <Loader2 className="h-4 w-4 animate-spin" />}
              Save changes
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
