#!/usr/bin/env python3
"""
Reflex session cleanup hook.

Clears .reflex/.state/ on session start (startup, clear, compact) and
session end, so each session starts with a fresh injection history.
"""

import json
import os
import shutil
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
        # Clean on startup and clear; leave state intact on compact
        source = input_data.get("source", "")
        if source not in CLEAN_SOURCES:
            sys.exit(0)
    elif event != "SessionEnd":
        sys.exit(0)

    project_dir = Path(input_data.get("cwd") or os.environ.get("CLAUDE_PROJECT_DIR", "."))
    state_dir = project_dir / ".reflex" / ".state"

    if state_dir.exists():
        shutil.rmtree(state_dir, ignore_errors=True)

    sys.exit(0)


if __name__ == "__main__":
    main()
