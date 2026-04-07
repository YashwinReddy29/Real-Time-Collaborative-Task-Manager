import { useState, useEffect, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import api from "../lib/api";
import { BoardData, WSEvent } from "../types";
import { useWebSocket } from "../hooks/useWebSocket";
import { KanbanBoard } from "../components/board/KanbanBoard";
import { ActivityFeed } from "../components/board/ActivityFeed";
import { Zap, ArrowLeft, Activity, Loader2, UserPlus } from "lucide-react";
import { InviteModal } from "../components/board/InviteModal";

export default function BoardPage() {
  const { boardId } = useParams<{ boardId: string }>();
  const navigate = useNavigate();
  const [data, setData] = useState<BoardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [showActivity, setShowActivity] = useState(false);
  const [activities, setActivities] = useState<any[]>([]);
  const [onlineUserIds, setOnlineUserIds] = useState<string[]>([]);
  const [showInvite, setShowInvite] = useState(false);

  useEffect(() => {
    if (!boardId) return;
    api.get(`/boards/${boardId}`)
      .then((res) => setData(res.data))
      .finally(() => setLoading(false));
    api.get(`/boards/${boardId}/activity`)
      .then((res) => setActivities(res.data));
  }, [boardId]);

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === "activity") {
      setActivities((prev) => [event.payload, ...prev].slice(0, 50));
      return;
    }

    if (event.type === "presence.update") {
      setOnlineUserIds((event.payload as string[]) ?? []);
      return;
    }

    setData((prev) => {
      if (!prev) return prev;
      switch (event.type) {
        case "task.created":
          if (prev.tasks.some((t) => t.id === event.payload.id)) return prev;
          return { ...prev, tasks: [...prev.tasks, event.payload] };
        case "task.updated":
        case "task.moved":
          return { ...prev, tasks: prev.tasks.map((t) => t.id === event.payload.id ? event.payload : t) };
        case "task.deleted":
          return { ...prev, tasks: prev.tasks.filter((t) => t.id !== event.payload.task_id) };
        case "column.created":
          if (prev.columns.some((c) => c.id === event.payload.id)) return prev;
          return { ...prev, columns: [...prev.columns, event.payload] };
        case "column.deleted":
          return { ...prev, columns: prev.columns.filter((c) => c.id !== event.payload.column_id) };
        default:
          return prev;
      }
    });
  }, []);

  useWebSocket(boardId ?? "", handleWSEvent);

  if (loading) return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center">
      <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
    </div>
  );

  if (!data) return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center text-gray-400">
      Board not found
    </div>
  );

  const onlineMembers = data.members.filter((m) => (onlineUserIds ?? []).includes(m.id));

  return (
    <div className="min-h-screen bg-gray-950 flex flex-col">
      <nav className="border-b border-gray-800 px-6 py-3 flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate("/")} className="text-gray-400 hover:text-white transition">
            <ArrowLeft className="h-5 w-5" />
          </button>
          <div className="flex items-center gap-2 font-bold text-white">
            <Zap className="h-4 w-4 text-violet-500" />
            {data.board.name}
          </div>
        </div>
        <div className="flex items-center gap-3">
          {/* Online presence */}
          <div className="flex items-center gap-2">
            {onlineMembers.length > 0 && (
              <span className="text-xs text-gray-500">
                {onlineMembers.length} online
              </span>
            )}
            <div className="flex -space-x-2">
              {data.members.map((m) => {
                const isOnline = (onlineUserIds ?? []).includes(m.id);
                return (
                  <div
                    key={m.id}
                    title={`${m.name}${isOnline ? " (online)" : ""}`}
                    className={`relative w-7 h-7 rounded-full border-2 border-gray-950 flex items-center justify-center text-xs font-medium transition-all ${
                      isOnline ? "bg-violet-600 text-white" : "bg-gray-700 text-gray-400"
                    }`}
                  >
                    {m.name[0]}
                    {isOnline && (
                      <span className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 rounded-full border-2 border-gray-950" />
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          <button
            onClick={() => setShowInvite(true)}
            className="flex items-center gap-2 text-sm px-3 py-1.5 rounded-lg text-gray-400 hover:text-white transition"
          >
            <UserPlus className="h-4 w-4" />
            Invite
          </button>
          <button
            onClick={() => setShowActivity(!showActivity)}
            className={`flex items-center gap-2 text-sm px-3 py-1.5 rounded-lg transition ${
              showActivity ? "bg-violet-600 text-white" : "text-gray-400 hover:text-white"
            }`}
          >
            <Activity className="h-4 w-4" />
            Activity
          </button>
        </div>
      </nav>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 overflow-hidden">
          <KanbanBoard
            boardId={boardId!}
            columns={data.columns}
            tasks={data.tasks}
            onDataChange={(updater) =>
              setData((prev) => prev ? { ...prev, ...updater(prev) } : prev)
            }
          />
        </div>
        {showActivity && (
          <div className="w-72 border-l border-gray-800 flex-shrink-0">
            <ActivityFeed activities={activities} />
          </div>
        )}
      </div>
    {showInvite && (
        <InviteModal
          boardId={boardId!}
          onClose={() => setShowInvite(false)}
          onInvited={(user) => {
            setData((prev) => prev && !prev.members.some(m => m.id === user.id)
              ? { ...prev, members: [...prev.members, user] }
              : prev
            );
            setShowInvite(false);
          }}
        />
      )}
    </div>
  );
}
