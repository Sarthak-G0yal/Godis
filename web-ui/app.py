from __future__ import annotations

from datetime import datetime, timezone
import os

import streamlit as st

from modules.client import GodisClient


def _has_streamlit_context() -> bool:
    try:
        from streamlit import runtime
    except Exception:
        return False
    return runtime.exists()


def _default_host() -> str:
    return os.getenv("GODIS_HOST", "godis")


def _default_port() -> int:
    raw = os.getenv("GODIS_PORT", "6379")
    try:
        return int(raw)
    except ValueError:
        return 6379


def _default_timeout() -> float:
    raw = os.getenv("GODIS_TIMEOUT", "3.0")
    try:
        return float(raw)
    except ValueError:
        return 3.0


def _init_state() -> None:
    if "client" not in st.session_state:
        st.session_state.client = None
    if "history" not in st.session_state:
        st.session_state.history = []
    if "last_result" not in st.session_state:
        st.session_state.last_result = None


def _render_result(result: dict) -> None:
    kind = result.get("kind", "unknown")
    if kind == "simple":
        st.success(result.get("value", ""))
    elif kind == "error":
        st.error(result.get("message", "unknown error"))
    elif kind == "integer":
        st.info(f"Integer: {result.get('value')}")
    elif kind == "bulk":
        st.info(f"Bulk string: {result.get('value', '')}")
    elif kind == "nil":
        st.info("Nil bulk string")
    else:
        st.warning(result.get("value", "unknown response"))

    st.code(result.get("raw", ""), language="text")


def _history_summary(result: dict) -> str:
    kind = result.get("kind", "unknown")
    if kind == "simple":
        return result.get("value", "")
    if kind == "error":
        return result.get("message", "")
    if kind == "integer":
        return str(result.get("value", ""))
    if kind == "bulk":
        return result.get("value", "")
    if kind == "nil":
        return "nil"
    return result.get("value", "")


def _record(command: str, result: dict) -> None:
    entry = {
        "time": datetime.now(timezone.utc).strftime("%H:%M:%S"),
        "command": command,
        "type": result.get("kind", "unknown"),
        "summary": _history_summary(result),
        "raw": result.get("raw", ""),
    }
    st.session_state.history.insert(0, entry)
    st.session_state.history = st.session_state.history[:50]
    st.session_state.last_result = result


def _execute(command: str) -> None:
    try:
        client = _get_client(autoconnect=True)
        result = client.execute(command)
        _record(command, result)
        _render_result(result)
    except Exception as err:
        st.error(str(err))


def _get_client(autoconnect: bool) -> GodisClient:
    host = _default_host()
    port = _default_port()
    timeout = _default_timeout()

    client = st.session_state.client
    if client is None:
        client = GodisClient(host=host, port=port, timeout=timeout)
        st.session_state.client = client

    if autoconnect and not client.is_connected():
        client.connect()

    return client


def main() -> None:
    st.set_page_config(page_title="Godis UI", page_icon="db", layout="wide")
    st.title("Godis Web UI")
    st.caption("Command Runner + key/value forms backed by the Godis TCP server")

    _init_state()

    with st.sidebar:
        st.header("Server")
        st.write(f"Host: {_default_host()}")
        st.write(f"Port: {_default_port()}")
        st.write(f"Timeout: {_default_timeout()}s")

        col1, col2 = st.columns(2)

        if col1.button("Reconnect", use_container_width=True):
            try:
                if st.session_state.client is not None:
                    st.session_state.client.disconnect()
                st.session_state.client = None
                _get_client(autoconnect=True)
                st.success("Connected")
            except Exception as err:
                st.error(f"Connect failed: {err}")

        if col2.button("Disconnect", use_container_width=True):
            if st.session_state.client is not None:
                st.session_state.client.disconnect()
                st.session_state.client = None
            st.info("Disconnected")

        connected = st.session_state.client is not None and st.session_state.client.is_connected()
        st.write(f"Status: {'Connected' if connected else 'Disconnected'}")

    tab_command, tab_kv, tab_history = st.tabs(["Command Runner", "Key/Value", "History"])

    with tab_command:
        with st.form("raw-command-form"):
            raw_command = st.text_input("Command", value="PING", help="Example: SET language godis")
            send_raw = st.form_submit_button("Send Command")

        if send_raw:
            _execute(raw_command)

        if st.session_state.last_result is not None:
            st.subheader("Last Response")
            _render_result(st.session_state.last_result)

    with tab_kv:
        col_left, col_right = st.columns(2)

        with col_left:
            with st.form("set-form"):
                st.markdown("### SET")
                set_key = st.text_input("Key", key="set-key")
                set_value = st.text_input("Value", key="set-value")
                submit_set = st.form_submit_button("Run SET")

            if submit_set:
                if not set_key:
                    st.error("Key is required")
                else:
                    _execute(f"SET {set_key} {set_value}")

            with st.form("get-form"):
                st.markdown("### GET")
                get_key = st.text_input("Key", key="get-key")
                submit_get = st.form_submit_button("Run GET")

            if submit_get:
                if not get_key:
                    st.error("Key is required")
                else:
                    _execute(f"GET {get_key}")

        with col_right:
            with st.form("exists-form"):
                st.markdown("### EXISTS")
                exists_key = st.text_input("Key", key="exists-key")
                submit_exists = st.form_submit_button("Run EXISTS")

            if submit_exists:
                if not exists_key:
                    st.error("Key is required")
                else:
                    _execute(f"EXISTS {exists_key}")

            with st.form("del-form"):
                st.markdown("### DEL")
                del_key = st.text_input("Key", key="del-key")
                submit_del = st.form_submit_button("Run DEL")

            if submit_del:
                if not del_key:
                    st.error("Key is required")
                else:
                    _execute(f"DEL {del_key}")

    with tab_history:
        st.subheader("Recent Commands")
        if st.button("Clear History"):
            st.session_state.history = []

        if not st.session_state.history:
            st.info("No commands yet")
        else:
            rows = [
                {
                    "time": item["time"],
                    "command": item["command"],
                    "type": item["type"],
                    "summary": item["summary"],
                }
                for item in st.session_state.history
            ]
            st.dataframe(rows, use_container_width=True)

            with st.expander("Latest raw response"):
                st.code(st.session_state.history[0]["raw"], language="text")


if __name__ == "__main__":
    if not _has_streamlit_context():
        print("Run this UI with: uv run streamlit run app.py")
    else:
        main()
