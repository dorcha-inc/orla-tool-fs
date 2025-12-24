#!/usr/bin/env python3
"""
File read tool for orla
Reads and returns the contents of a file
"""

import argparse
import sys
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description="Read file contents")
    parser.add_argument("--path", required=True, help="Path to the file to read")

    args = parser.parse_args()

    try:
        file_path = Path(args.path)

        # Check if file exists
        if not file_path.exists():
            print(f"Error: File not found: {args.path}", file=sys.stderr)
            sys.exit(1)

        # Check if it's a file (not a directory)
        if not file_path.is_file():
            print(f"Error: Path is not a file: {args.path}", file=sys.stderr)
            sys.exit(1)

        # Read file contents
        try:
            content = file_path.read_text(encoding="utf-8")
            print(content, end="")
            sys.exit(0)
        except UnicodeDecodeError:
            print(
                f"Error: File contains binary data or invalid UTF-8: {args.path}",
                file=sys.stderr,
            )
            sys.exit(1)
        except PermissionError:
            print(
                f"Error: Permission denied reading file: {args.path}", file=sys.stderr
            )
            sys.exit(1)
        except Exception as e:
            print(f"Error: Failed to read file: {e}", file=sys.stderr)
            sys.exit(1)

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()

