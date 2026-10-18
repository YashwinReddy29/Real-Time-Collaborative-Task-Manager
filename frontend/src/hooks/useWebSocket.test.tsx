// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render } from "@testing-library/react";
import { useWebSocket } from "./useWebSocket";
import api from "../lib/api";
import { AxiosError } from "axios";

vi.mock("../lib/api", () => ({
  default: {
    get: vi.fn(),
  },
}));

type SocketHandler = (() => void) | null;

class MockWebSocket {
  static instances: MockWebSocket[] = [];

  onopen: SocketHandler = null;
  onclose: SocketHandler = null;
  onerror: SocketHandler = null;
  onmessage: ((event: MessageEvent<string>) => void) | null = null;

  close = vi.fn(() => {
    this.onclose?.();
  });

  constructor(public url: string) {
    MockWebSocket.instances.push(this);
  }

  open() {
    this.onopen?.();
  }

  disconnect() {
    this.onclose?.();
  }
}

function TestHook({ onReconnect = () => {} }: {
  onReconnect?: () => void;
}) {
  useWebSocket("board-123", () => {}, onReconnect);
  return null;
}

describe("useWebSocket", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal("WebSocket", MockWebSocket);
    localStorage.setItem("token", "test-token");

    MockWebSocket.instances = [];
    vi.mocked(api.get).mockReset();
    vi.mocked(api.get).mockResolvedValue({ data: {} });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it("connects to the current origin", () => {
    const { unmount } = render(<TestHook />);

    expect(MockWebSocket.instances).toHaveLength(1);
    expect(MockWebSocket.instances[0].url).toContain(
      "/ws/board-123?token=test-token"
    );

    unmount();
  });

  it("reconnects after a temporary failure", async () => {
    const { unmount } = render(<TestHook />);

    await act(async () => {
      MockWebSocket.instances[0].disconnect();
      await Promise.resolve();
    });

    expect(api.get).toHaveBeenCalledWith(
      "/boards/board-123",
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    );

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1100);
    });

    expect(MockWebSocket.instances).toHaveLength(2);

    unmount();
  });

  it("stops reconnecting after board access is denied", async () => {
    vi.mocked(api.get).mockRejectedValue(
      new AxiosError(
        "Access denied",
        "ERR_BAD_REQUEST",
        undefined,
        undefined,
        {
          data: { error: "access denied" },
          status: 403,
          statusText: "Forbidden",
          headers: {},
          config: { headers: {} } as never,
        }
      )
    );

    const { unmount } = render(<TestHook />);

    await act(async () => {
      MockWebSocket.instances[0].disconnect();
      await Promise.resolve();
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(35000);
    });

    expect(MockWebSocket.instances).toHaveLength(1);

    unmount();
  });

  it("refreshes board state after a successful reconnect", async () => {
    const onReconnect = vi.fn();
    const { unmount } = render(
      <TestHook onReconnect={onReconnect} />
    );

    act(() => MockWebSocket.instances[0].open());

    await act(async () => {
      MockWebSocket.instances[0].disconnect();
      await Promise.resolve();
      await vi.advanceTimersByTimeAsync(1100);
    });

    act(() => MockWebSocket.instances[1].open());

    expect(onReconnect).toHaveBeenCalledTimes(1);

    unmount();
  });

  it("cancels reconnection after unmount", async () => {
    const { unmount } = render(<TestHook />);

    await act(async () => {
      MockWebSocket.instances[0].disconnect();
      await Promise.resolve();
    });

    unmount();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(35000);
    });

    expect(MockWebSocket.instances).toHaveLength(1);
  });
});
