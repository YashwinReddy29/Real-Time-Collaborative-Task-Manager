import { useState } from "react";
import {
  DndContext,
  DragOverlay,
  closestCorners,
  PointerSensor,
  useSensor,
  useSensors,
  DragStartEvent,
  DragEndEvent,
  DragOverEvent,
  useDroppable,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Column, Task, BoardData } from "../../types";
import { TaskCard } from "./TaskCard";
import { CreateTaskForm } from "./CreateTaskForm";
import api from "../../lib/api";
import { Plus, Trash2, Loader2, GripVertical } from "lucide-react";

interface Props {
  boardId: string;
  columns: Column[];
  tasks: Task[];
  onDataChange: (updater: (prev: BoardData) => Partial<BoardData>) => void;
}

function DroppableColumn({ col, children }: { col: Column; children: React.ReactNode }) {
  const { setNodeRef, isOver } = useDroppable({ id: col.id });
  return (
    <div
      ref={setNodeRef}
      className={`flex-1 overflow-y-auto p-3 space-y-2 min-h-[80px] rounded-lg transition-colors ${
        isOver ? "bg-violet-950/30 border border-dashed border-violet-700" : ""
      }`}
    >
      {children}
    </div>
  );
}

function SortableTaskCard({
  task,
  columns,
  boardId,
  onMove,
  onDelete,
  onUpdate,
}: {
  task: Task;
  columns: Column[];
  boardId: string;
  onMove: (taskId: string, columnId: string, position: number) => void;
  onDelete: (taskId: string) => void;
  onUpdate: (task: Task) => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: task.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  };

  return (
    <div ref={setNodeRef} style={style}>
      <div className="relative group/sortable">
        <div
          {...attributes}
          {...listeners}
          className="absolute left-1 top-1/2 -translate-y-1/2 z-10 opacity-0 group-hover/sortable:opacity-100 cursor-grab active:cursor-grabbing text-gray-600 hover:text-gray-400 transition"
        >
          <GripVertical className="h-3.5 w-3.5" />
        </div>
        <div className="pl-4">
          <TaskCard
            task={task}
            columns={columns}
            boardId={boardId}
            onMove={onMove}
            onDelete={onDelete}
            onUpdate={onUpdate}
          />
        </div>
      </div>
    </div>
  );
}

export function KanbanBoard({ boardId, columns, tasks, onDataChange }: Props) {
  const [addingColumn, setAddingColumn] = useState(false);
  const [newColName, setNewColName] = useState("");
  const [creatingCol, setCreatingCol] = useState(false);
  const [addingTaskCol, setAddingTaskCol] = useState<string | null>(null);
  const [activeTask, setActiveTask] = useState<Task | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  );

  const getTasksForColumn = (columnId: string) =>
    tasks.filter((t) => t.column_id === columnId).sort((a, b) => a.position - b.position);

  function handleDragStart(event: DragStartEvent) {
    const task = tasks.find((t) => t.id === event.active.id);
    if (task) setActiveTask(task);
  }

  function handleDragOver(event: DragOverEvent) {
    const { active, over } = event;
    if (!over) return;

    const draggedTask = tasks.find((t) => t.id === active.id);
    if (!draggedTask) return;

    // Over a column droppable zone
    const overColumn = columns.find((c) => c.id === over.id);
    if (overColumn && draggedTask.column_id !== overColumn.id) {
      onDataChange((prev) => ({
        tasks: prev.tasks.map((t) =>
          t.id === draggedTask.id ? { ...t, column_id: overColumn.id } : t
        ),
      }));
      return;
    }

    // Over another task
    const overTask = tasks.find((t) => t.id === over.id);
    if (overTask && overTask.column_id !== draggedTask.column_id) {
      onDataChange((prev) => ({
        tasks: prev.tasks.map((t) =>
          t.id === draggedTask.id ? { ...t, column_id: overTask.column_id } : t
        ),
      }));
    }
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    setActiveTask(null);
    if (!over) return;

    const task = tasks.find((t) => t.id === active.id);
    if (!task) return;

    let targetColumnId = task.column_id;
    const overColumn = columns.find((c) => c.id === over.id);
    const overTask = tasks.find((t) => t.id === over.id);

    if (overColumn) targetColumnId = overColumn.id;
    else if (overTask) targetColumnId = overTask.column_id;

    const columnTasks = getTasksForColumn(targetColumnId);
    const newPosition = overTask
      ? columnTasks.findIndex((t) => t.id === overTask.id)
      : columnTasks.length;

    try {
      const res = await api.patch(`/boards/${boardId}/tasks/${task.id}/move`, {
        column_id: targetColumnId,
        position: Math.max(0, newPosition),
      });
      onDataChange((prev) => ({
        tasks: prev.tasks.map((t) => t.id === task.id ? res.data : t),
      }));
    } catch (e) {
      console.error("Move failed", e);
    }
  }

  async function createColumn(e: React.FormEvent) {
    e.preventDefault();
    if (!newColName.trim()) return;
    setCreatingCol(true);
    const res = await api.post(`/boards/${boardId}/columns`, {
      name: newColName,
      position: columns.length,
    });
    onDataChange((prev) => ({
      columns: prev.columns.some((c) => c.id === res.data.id)
        ? prev.columns
        : [...prev.columns, res.data],
    }));
    setNewColName("");
    setAddingColumn(false);
    setCreatingCol(false);
  }

  async function deleteColumn(columnId: string) {
    await api.delete(`/boards/${boardId}/columns/${columnId}`);
    onDataChange((prev) => ({
      columns: prev.columns.filter((c) => c.id !== columnId),
      tasks: prev.tasks.filter((t) => t.column_id !== columnId),
    }));
  }

  async function moveTask(taskId: string, newColumnId: string, newPosition: number) {
    const res = await api.patch(`/boards/${boardId}/tasks/${taskId}/move`, {
      column_id: newColumnId,
      position: newPosition,
    });
    onDataChange((prev) => ({
      tasks: prev.tasks.map((t) => t.id === taskId ? res.data : t),
    }));
  }

  async function deleteTask(taskId: string) {
    await api.delete(`/boards/${boardId}/tasks/${taskId}`);
    onDataChange((prev) => ({
      tasks: prev.tasks.filter((t) => t.id !== taskId),
    }));
  }

  function handleTaskCreated(task: Task) {
    onDataChange((prev) => ({
      tasks: prev.tasks.some((t) => t.id === task.id)
        ? prev.tasks
        : [...prev.tasks, task],
    }));
    setAddingTaskCol(null);
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 p-6 overflow-x-auto h-full items-start">
        {columns.map((col) => {
          const colTasks = getTasksForColumn(col.id);
          return (
            <div
              key={col.id}
              className="flex-shrink-0 w-72 bg-gray-900 rounded-xl border border-gray-800 flex flex-col max-h-full"
            >
              <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800">
                <h3 className="font-semibold text-white text-sm">{col.name}</h3>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-gray-500 bg-gray-800 px-2 py-0.5 rounded-full">
                    {colTasks.length}
                  </span>
                  <button
                    onClick={() => deleteColumn(col.id)}
                    className="text-gray-600 hover:text-red-400 transition"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>

              <SortableContext
                items={colTasks.map((t) => t.id)}
                strategy={verticalListSortingStrategy}
              >
                <DroppableColumn col={col}>
                  {colTasks.length === 0 ? (
                    <div className="flex items-center justify-center h-16 text-xs text-gray-600">
                      Drop tasks here
                    </div>
                  ) : (
                    colTasks.map((task) => (
                      <SortableTaskCard
                        key={task.id}
                        task={task}
                        columns={columns}
                        boardId={boardId}
                        onMove={moveTask}
                        onDelete={deleteTask}
                        onUpdate={(updated) =>
                          onDataChange((prev) => ({
                            tasks: prev.tasks.map((t) =>
                              t.id === updated.id ? updated : t
                            ),
                          }))
                        }
                      />
                    ))
                  )}
                </DroppableColumn>
              </SortableContext>

              <div className="p-3 border-t border-gray-800">
                {addingTaskCol === col.id ? (
                  <CreateTaskForm
                    boardId={boardId}
                    columnId={col.id}
                    position={colTasks.length}
                    onCreated={handleTaskCreated}
                    onCancel={() => setAddingTaskCol(null)}
                  />
                ) : (
                  <button
                    onClick={() => setAddingTaskCol(col.id)}
                    className="w-full flex items-center gap-2 text-sm text-gray-500 hover:text-gray-300 transition py-1"
                  >
                    <Plus className="h-4 w-4" /> Add task
                  </button>
                )}
              </div>
            </div>
          );
        })}

        <div className="flex-shrink-0 w-72">
          {addingColumn ? (
            <form
              onSubmit={createColumn}
              className="bg-gray-900 border border-gray-800 rounded-xl p-4 space-y-3"
            >
              <input
                value={newColName}
                onChange={(e) => setNewColName(e.target.value)}
                placeholder="Column name"
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                autoFocus
                required
              />
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={creatingCol}
                  className="bg-violet-600 text-white px-3 py-1.5 rounded-lg text-sm font-medium hover:bg-violet-700 disabled:opacity-50 flex items-center gap-1"
                >
                  {creatingCol && <Loader2 className="h-3 w-3 animate-spin" />}
                  Add
                </button>
                <button
                  type="button"
                  onClick={() => setAddingColumn(false)}
                  className="text-sm text-gray-400 hover:text-white px-3 py-1.5"
                >
                  Cancel
                </button>
              </div>
            </form>
          ) : (
            <button
              onClick={() => setAddingColumn(true)}
              className="w-full flex items-center gap-2 text-sm text-gray-500 hover:text-gray-300 transition bg-gray-900 border border-dashed border-gray-700 rounded-xl px-4 py-3 hover:border-gray-500"
            >
              <Plus className="h-4 w-4" /> Add column
            </button>
          )}
        </div>
      </div>

      <DragOverlay>
        {activeTask && (
          <div className="rotate-2 opacity-90 w-72">
            <TaskCard
              task={activeTask}
              columns={columns}
              boardId={boardId}
              onMove={() => {}}
              onDelete={() => {}}
              onUpdate={() => {}}
            />
          </div>
        )}
      </DragOverlay>
    </DndContext>
  );
}
