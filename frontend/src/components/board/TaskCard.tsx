import { useState } from "react";
import { Task, Column } from "../../types";
import { Flag } from "lucide-react";
import { TaskModal } from "./TaskModal";

interface Props {
  task: Task;
  columns: Column[];
  boardId: string;
  onMove: (taskId: string, columnId: string, position: number) => void;
  onDelete: (taskId: string) => void;
  onUpdate: (task: Task) => void;
}

const priorityColors = {
  low: "text-blue-400 bg-blue-900/30",
  medium: "text-yellow-400 bg-yellow-900/30",
  high: "text-red-400 bg-red-900/30",
};

export function TaskCard({ task, columns, boardId, onMove, onDelete, onUpdate }: Props) {
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <div
        onClick={() => setShowModal(true)}
        className="bg-gray-800 rounded-lg p-3 border border-gray-700 hover:border-violet-600/50 transition cursor-pointer group"
      >
        <p className="text-sm text-white font-medium leading-snug">{task.title}</p>
        {task.description && (
          <p className="text-xs text-gray-500 mt-1 line-clamp-2">{task.description}</p>
        )}
        <div className="flex items-center justify-between mt-2">
          <span className={`text-xs px-1.5 py-0.5 rounded flex items-center gap-1 ${priorityColors[task.priority]}`}>
            <Flag className="h-2.5 w-2.5" />
            {task.priority}
          </span>
        </div>
      </div>

      {showModal && (
        <TaskModal
          task={task}
          columns={columns}
          boardId={boardId}
          onClose={() => setShowModal(false)}
          onUpdate={(updated) => {
            onUpdate(updated);
            setShowModal(false);
          }}
          onDelete={(taskId) => {
            onDelete(taskId);
            setShowModal(false);
          }}
        />
      )}
    </>
  );
}
