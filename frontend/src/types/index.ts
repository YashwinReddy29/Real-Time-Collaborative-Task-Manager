export type User = {
  id: string;
  name: string;
  email: string;
  created_at: string;
}

export type Board = {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  created_at: string;
}

export type Column = {
  id: string;
  board_id: string;
  name: string;
  position: number;
}

export type Task = {
  id: string;
  column_id: string;
  board_id: string;
  title: string;
  description: string;
  position: number;
  assignee_id: string | null;
  priority: "low" | "medium" | "high";
  created_at: string;
}

export type Activity = {
  id: string;
  board_id: string;
  user_id: string;
  user_name: string;
  action: string;
  created_at: string;
}

export type WSEvent = {
  type: string;
  payload: any;
  board_id: string;
  user_id?: string;
}

export type BoardData = {
  board: Board;
  columns: Column[];
  tasks: Task[];
  members: User[];
}
