from __future__ import annotations

import socket
from typing import Optional

from modules.parser import parse_response


class GodisClient:
    def __init__(self, host: str, port: int, timeout: float = 3.0) -> None:
        self.host = host
        self.port = port
        self.timeout = timeout
        self._conn: Optional[socket.socket] = None
        self._reader = None

    def connect(self) -> None:
        if self.is_connected():
            return

        self._conn = socket.create_connection((self.host, self.port), timeout=self.timeout)
        self._conn.settimeout(self.timeout)
        self._reader = self._conn.makefile("rb")

    def disconnect(self) -> None:
        if self._reader is not None:
            self._reader.close()
            self._reader = None

        if self._conn is not None:
            self._conn.close()
            self._conn = None

    def is_connected(self) -> bool:
        if self._conn is None:
            return False
        try:
            self._conn.getpeername()
            return True
        except OSError:
            return False

    def execute(self, command: str) -> dict:
        command = command.strip()
        if not command:
            raise ValueError("command cannot be empty")

        if not self.is_connected():
            self.connect()

        if self._conn is None:
            raise RuntimeError("connection not available")

        try:
            self._conn.sendall((command + "\n").encode("utf-8"))
            raw = self._read_raw_response()
            return parse_response(raw)
        except (socket.timeout, OSError) as err:
            self.disconnect()
            raise RuntimeError(f"communication failed: {err}") from err

    def _read_raw_response(self) -> bytes:
        prefix = self._read_exact(1)

        if prefix in (b"+", b"-", b":"):
            return prefix + self._readline()

        if prefix == b"$":
            header = self._readline()
            raw = prefix + header

            try:
                length = int(header[:-2].decode("utf-8", errors="replace"))
            except ValueError as err:
                raise RuntimeError("invalid bulk header from server") from err

            if length == -1:
                return raw

            if length < -1:
                raise RuntimeError("invalid bulk length from server")

            payload_with_trailer = self._read_exact(length + 2)
            return raw + payload_with_trailer

        return prefix + self._readline()

    def _readline(self) -> bytes:
        if self._reader is None:
            raise RuntimeError("reader is not initialized")

        line = self._reader.readline()
        if line == b"":
            raise RuntimeError("server closed connection")
        return line

    def _read_exact(self, size: int) -> bytes:
        if self._reader is None:
            raise RuntimeError("reader is not initialized")

        data = self._reader.read(size)
        if data is None or len(data) != size:
            raise RuntimeError("server closed connection before full response was read")
        return data
