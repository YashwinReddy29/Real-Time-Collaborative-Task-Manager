import json
import time
import uuid
from urllib import request, error, parse
from websocket import (
    create_connection,
    WebSocketTimeoutException,
    WebSocketBadStatusException,
)

BASE = "http://localhost:3000"
WS_BASE = "ws://localhost:3000"
SUFFIX = uuid.uuid4().hex[:12]
PASSWORD = "SmokeTest-StrongPass-2026!"

def api(method, path, token=None, body=None, expected=200):
    headers = {"Content-Type": "application/json"}

    if token:
        headers["Authorization"] = f"Bearer {token}"

    data = json.dumps(body).encode() if body is not None else None

    req = request.Request(
        BASE + path,
        data=data,
        headers=headers,
        method=method,
    )

    try:
        with request.urlopen(req, timeout=15) as response:
            status = response.status
            result = json.load(response)
    except error.HTTPError as exc:
        status = exc.code
        result = json.loads(exc.read())

    if status != expected:
        raise AssertionError(
            f"{method} {path}: HTTP {status}, expected {expected}: {result}"
        )

    return result

def next_event(ws, event_type, timeout=8):
    deadline = time.monotonic() + timeout

    while time.monotonic() < deadline:
        ws.settimeout(max(0.1, deadline - time.monotonic()))

        try:
            raw = ws.recv()
        except WebSocketTimeoutException:
            continue

        event = json.loads(raw)

        if event.get("type") == event_type:
            return event

    raise AssertionError(f"Never received WebSocket event: {event_type}")

def main():
    print("===== REGISTER TWO USERS =====")
    email_a = f"smoke-owner-{SUFFIX}@example.test"
    email_b = f"smoke-member-{SUFFIX}@example.test"

    owner = api("POST", "/api/auth/register", body={
        "name": "Smoke Owner",
        "email": email_a,
        "password": PASSWORD,
    }, expected=201)

    member = api("POST", "/api/auth/register", body={
        "name": "Smoke Member",
        "email": email_b,
        "password": PASSWORD,
    }, expected=201)

    owner_token = owner["token"]
    member_token = member["token"]
    print("PASS: Both users registered")

    print("===== CREATE BOARD =====")
    board = api("POST", "/api/boards", owner_token, {
        "name": f"Smoke Collaboration {SUFFIX}",
        "description": "Temporary multi-user smoke-test board",
    }, expected=201)

    board_id = board["id"]
    print("PASS: Board created")

    print("===== ACCESS CONTROL =====")
    api("GET", f"/api/boards/{board_id}", member_token, expected=403)
    print("PASS: Nonmember access rejected")

    print("===== REJECT UNAUTHORIZED WEBSOCKET =====")
    try:
        unauthorized_ws = create_connection(
            f"{WS_BASE}/ws/{board_id}?token={parse.quote(member_token)}",
            origin=BASE,
            timeout=5,
        )
    except WebSocketBadStatusException as exc:
        assert exc.status_code == 403, (
            f"Expected 403, got {exc.status_code}"
        )
        print("PASS: Nonmember WebSocket rejected")
    else:
        unauthorized_ws.close()
        raise AssertionError("Nonmember WebSocket was accepted")

    print("===== CREATE ISOLATION BOARD =====")
    isolated = api("POST", "/api/boards", owner_token, {
        "name": f"Isolation Board {SUFFIX}",
    }, expected=201)
    isolated_board_id = isolated["id"]
    print("PASS: Separate board created")

    print("===== INVITE SECOND USER =====")
    api("POST", f"/api/boards/{board_id}/invite", owner_token, {
        "email": email_b,
    })
    print("PASS: Second user invited")

    board_data = api("GET", f"/api/boards/{board_id}", member_token)
    column_id = board_data["columns"][0]["id"]
    print("PASS: Invited member can access board")

    print("===== OPEN TWO WEBSOCKETS =====")
    sockets = []

    try:
        ws_a = create_connection(
            f"{WS_BASE}/ws/{board_id}?token={parse.quote(owner_token)}",
            origin=BASE,
            timeout=8,
        )
        sockets.append(ws_a)

        ws_b = create_connection(
            f"{WS_BASE}/ws/{board_id}?token={parse.quote(member_token)}",
            origin=BASE,
            timeout=8,
        )
        sockets.append(ws_b)

        next_event(ws_b, "presence.update")
        print("PASS: Two authenticated WebSocket connections")

        ws_isolated = create_connection(
            f"{WS_BASE}/ws/{isolated_board_id}?token={parse.quote(owner_token)}",
            origin=BASE,
            timeout=8,
        )
        sockets.append(ws_isolated)

        next_event(ws_isolated, "presence.update")
        print("PASS: Separate board WebSocket connected")

        print("===== CREATE TASK =====")
        task = api("POST", f"/api/boards/{board_id}/tasks", owner_token, {
            "column_id": column_id,
            "title": f"Real-time smoke task {SUFFIX}",
            "priority": "high",
            "position": 0,
        }, expected=201)

        event = next_event(ws_b, "task.created")

        assert event["payload"]["id"] == task["id"]
        assert event["board_id"] == board_id
        print("PASS: Task creation broadcast received by second user")

        print("===== UPDATE TASK =====")
        updated = api(
            "PUT",
            f"/api/boards/{board_id}/tasks/{task['id']}",
            owner_token,
            {
                "title": f"Updated smoke task {SUFFIX}",
                "description": "Updated by collaboration test",
                "priority": "medium",
            },
        )

        event = next_event(ws_b, "task.updated")
        assert event["payload"]["id"] == updated["id"]
        assert event["payload"]["title"] == updated["title"]
        print("PASS: Task update broadcast received")

        print("===== MOVE TASK =====")
        target_column = board_data["columns"][1]["id"]

        moved = api(
            "PATCH",
            f"/api/boards/{board_id}/tasks/{task['id']}/move",
            owner_token,
            {
                "column_id": target_column,
                "position": 0,
            },
        )

        event = next_event(ws_b, "task.moved")
        assert event["payload"]["id"] == task["id"]
        assert event["payload"]["column_id"] == target_column
        assert moved["column_id"] == target_column
        print("PASS: Task move broadcast received")

        print("===== CROSS-BOARD WEBSOCKET ISOLATION =====")
        ws_isolated.settimeout(1.5)

        try:
            raw = ws_isolated.recv()
            unexpected = json.loads(raw)
            raise AssertionError(
                f"Unrelated board received an event: {unexpected['type']}"
            )
        except WebSocketTimeoutException:
            print("PASS: No cross-board events received")

        print("===== WEBSOCKET RECONNECTION =====")
        ws_b.close()
        sockets.remove(ws_b)

        ws_b = create_connection(
            f"{WS_BASE}/ws/{board_id}?token={parse.quote(member_token)}",
            origin=BASE,
            timeout=8,
        )
        sockets.append(ws_b)

        presence = next_event(ws_b, "presence.update")
        assert member["user"]["id"] in presence["payload"]
        print("PASS: Member reconnected and received presence")

        current = api("GET", f"/api/boards/{board_id}", member_token)
        persisted = next(
            t for t in current["tasks"] if t["id"] == task["id"]
        )
        assert persisted["column_id"] == target_column
        assert persisted["title"] == updated["title"]
        print("PASS: Reconnected member can recover current board state")

        print("===== DELETE TASK =====")
        api(
            "DELETE",
            f"/api/boards/{board_id}/tasks/{task['id']}",
            owner_token,
        )

        event = next_event(ws_b, "task.deleted")
        assert event["payload"]["task_id"] == task["id"]
        print("PASS: Task deletion broadcast received")

        print("===== DATABASE PERSISTENCE =====")
        current = api("GET", f"/api/boards/{board_id}", member_token)
        assert all(t["id"] != task["id"] for t in current["tasks"])
        print("PASS: Deleted task absent from persisted board data")

    finally:
        for ws in sockets:
            ws.close()

    print()
    print("ALL COLLABORATION SMOKE TESTS PASSED")

if __name__ == "__main__":
    main()
