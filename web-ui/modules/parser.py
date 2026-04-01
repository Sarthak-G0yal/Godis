from __future__ import annotations


def parse_response(raw: bytes) -> dict:
    if not raw:
        return {"kind": "error", "message": "empty response", "raw": ""}

    prefix = raw[:1]
    wire = raw.decode("utf-8", errors="replace")

    if prefix == b"+":
        value = _extract_line(raw[1:])
        return {"kind": "simple", "value": value, "raw": wire}

    if prefix == b"-":
        message = _extract_line(raw[1:])
        if message.startswith("ERR "):
            message = message[4:]
        return {"kind": "error", "message": message, "raw": wire}

    if prefix == b":":
        value = _extract_line(raw[1:])
        try:
            number = int(value)
        except ValueError:
            return {"kind": "error", "message": f"invalid integer response: {value}", "raw": wire}
        return {"kind": "integer", "value": number, "raw": wire}

    if prefix == b"$":
        if raw == b"$-1\r\n":
            return {"kind": "nil", "value": None, "raw": wire}

        header_end = raw.find(b"\r\n")
        if header_end == -1:
            return {"kind": "error", "message": "malformed bulk response header", "raw": wire}

        length_str = raw[1:header_end].decode("utf-8", errors="replace")
        try:
            length = int(length_str)
        except ValueError:
            return {"kind": "error", "message": f"invalid bulk length: {length_str}", "raw": wire}

        payload_start = header_end + 2
        payload_end = payload_start + length

        if length < 0:
            return {"kind": "error", "message": f"invalid bulk length: {length}", "raw": wire}

        if len(raw) < payload_end + 2:
            return {"kind": "error", "message": "incomplete bulk response payload", "raw": wire}

        payload = raw[payload_start:payload_end]
        trailer = raw[payload_end : payload_end + 2]
        if trailer != b"\r\n":
            return {"kind": "error", "message": "bulk response missing CRLF trailer", "raw": wire}

        return {"kind": "bulk", "value": payload.decode("utf-8", errors="replace"), "raw": wire}

    return {"kind": "unknown", "value": wire, "raw": wire}


def _extract_line(data: bytes) -> str:
    if data.endswith(b"\r\n"):
        data = data[:-2]
    return data.decode("utf-8", errors="replace")
