import { useEffect, useRef } from "react";
import axios from "axios";
import api from "../lib/api";
import type { WSEvent } from "../types";

const INITIAL_RETRY_MS = 1000;
const MAX_RETRY_MS = 30000;

export function useWebSocket(
  boardId: string,
  onEvent: (event: WSEvent) => void,
  onReconnect?: () => void
) {
  const onEventRef = useRef(onEvent);
  const onReconnectRef = useRef(onReconnect);

  useEffect(() => {
    onEventRef.current = onEvent;
    onReconnectRef.current = onReconnect;
  }, [onEvent, onReconnect]);

  useEffect(() => {
    if (!boardId) return;

    let disposed = false;
    let socket: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let retryCount = 0;
    let hasConnected = false;
    const controller = new AbortController();

    const checkAccess = async (): Promise<boolean> => {
      if (!localStorage.getItem("token")) return false;

      try {
        await api.get(`/boards/${boardId}`, {
          signal: controller.signal,
        });

        return true;
      } catch (error: unknown) {
        if (disposed || controller.signal.aborted) return false;

        if (axios.isAxiosError(error)) {
          const status = error.response?.status;

          if (status === 401 || status === 403 || status === 404) {
            console.warn(
              "WebSocket reconnection stopped: access denied"
            );
            return false;
          }
        }

        // Temporary network/server failures remain retryable.
        return true;
      }
    };

    const connect = () => {
      if (disposed) return;

      const token = localStorage.getItem("token");
      if (!token) return;

      const url = new URL(
        `/ws/${encodeURIComponent(boardId)}`,
        window.location.origin
      );

      url.protocol =
        window.location.protocol === "https:" ? "wss:" : "ws:";

      url.searchParams.set("token", token);

      const connection = new WebSocket(url.toString());
      socket = connection;

      connection.onopen = () => {
        if (disposed) {
          connection.close();
          return;
        }

        retryCount = 0;

        if (hasConnected) {
          onReconnectRef.current?.();
        }

        hasConnected = true;
      };

      connection.onmessage = (message: MessageEvent<string>) => {
        if (disposed) return;

        try {
          const event: WSEvent = JSON.parse(message.data);
          onEventRef.current(event);
        } catch {
          console.warn("Ignored malformed WebSocket event");
        }
      };

      connection.onerror = () => {
        connection.close();
      };

      connection.onclose = () => {
        if (disposed || !localStorage.getItem("token")) return;

        void (async () => {
          const allowed = await checkAccess();

          if (!allowed || disposed) return;

          const delay = Math.min(
            INITIAL_RETRY_MS * 2 ** Math.min(retryCount, 5),
            MAX_RETRY_MS
          );

          retryCount += 1;

          retryTimer = setTimeout(() => {
            retryTimer = null;
            connect();
          }, delay);
        })();
      };
    };

    connect();

    return () => {
      disposed = true;
      controller.abort();

      if (retryTimer !== null) {
        clearTimeout(retryTimer);
      }

      if (socket !== null) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onclose = null;
        socket.onerror = null;
        socket.close();
      }
    };
  }, [boardId]);
}
