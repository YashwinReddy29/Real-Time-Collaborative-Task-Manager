import { type Activity } from "../../types";

export function ActivityFeed({ activities }: { activities: Activity[] }) {
  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-gray-800">
        <h3 className="font-semibold text-white text-sm">Activity</h3>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {activities.length === 0 ? (
          <p className="text-sm text-gray-500 text-center pt-8">No activity yet</p>
        ) : (
          activities.map((a) => (
            <div key={a.id} className="flex gap-3">
              <div className="w-6 h-6 rounded-full bg-violet-700 flex items-center justify-center text-xs text-white font-medium flex-shrink-0 mt-0.5">
                {a.user_name[0]}
              </div>
              <div>
                <p className="text-xs text-gray-300">
                  <span className="font-medium text-white">{a.user_name}</span>{" "}
                  {a.action}
                </p>
                <p className="text-xs text-gray-600 mt-0.5">
                  {new Date(a.created_at).toLocaleTimeString()}
                </p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
