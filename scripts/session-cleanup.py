#!/usr/bin/env python3
"""
Reflex session cleanup hook.

Deletes the current session's state file from ~/.config/reflex/state/ on
session start (startup, clear, compact) and session end, so each session
starts with a fresh injection history.
"""

import json
import sys
from pathlib import Path

CLEAN_SOURCES = {"startup", "clear", "compact"}


def main():
    try:
        input_data = json.load(sys.stdin)
    except (json.JSONDecodeError, EOFError):
        sys.exit(0)

    event = input_data.get("hook_event_name", "")

    if event == "SessionStart":
        source = input_data.get("source", "")
        if source not in CLEAN_SOURCES:
            sys.exit(0)
    elif event != "SessionEnd":
        sys.exit(0)

    state_dir = Path.home() / ".config" / "reflex" / "state"

    # Derive session key the same way the main hook does
    transcript_path = input_data.get("transcript_path", "")
    session_key = Path(transcript_path).stem if transcript_path else input_data.get("session_id", "")

    if session_key:
        session_file = state_dir / f"{session_key}.json"
        if session_file.exists():
            try:
                session_file.unlink()
            except OSError:
                pass

    sys.exit(0)


if __name__ == "__main__":
    main()
