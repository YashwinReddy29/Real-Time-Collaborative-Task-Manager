import { useEffect, useRef, useCallback } from "react";
import { WSEvent } from "../types";

export function useWebSocket(boardId: string, onEvent: (event: WSEvent) => void) {
  const ws = useRef<WebSocket | null>(null);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  const connect = useCallback(() => {
    const token = localStorage.getItem("token");
    if (!token || !boardId) return;

    const wsUrl = `ws://localhost:8080/ws/${boardId}?token=${token}`;
    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => {
      console.log("[WS] connected to board", boardId);
    };

    ws.current.onmessage = (e) => {
      try {
        const event: WSEvent = JSON.parse(e.data);
        onEventRef.current(event);
      } catch {}
    };

    ws.current.onclose = () => {
      console.log("[WS] disconnected, reconnecting in 3s...");
      setTimeout(connect, 3000);
    };

    ws.current.onerror = () => {
      ws.current?.close();
    };
  }, [boardId]);

  useEffect(() => {
    connect();
    return () => {
      ws.current?.close();
    };
  }, [connect]);
}
