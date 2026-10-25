#!/usr/bin/env python3
"""tool - list and count the entries of a directory."""
import argparse
import json
import os
import sys


def main(argv):
    parser = argparse.ArgumentParser(prog="tool")
    sub = parser.add_subparsers(dest="command", required=True)
    lister = sub.add_parser("list", help="list the entries of DIR")
    lister.add_argument("dir")
    lister.add_argument("--json", action="store_true", help="print one JSON list")
    counter = sub.add_parser("count", help="print the number of entries of DIR")
    counter.add_argument("dir")
    args = parser.parse_args(argv)
    try:
        entries = sorted(os.listdir(args.dir))
    except OSError as err:
        print(f"tool: {err}", file=sys.stderr)
        return 1
    if args.command == "count":
        print(len(entries))
    elif args.json:
        print(json.dumps(entries))
    else:
        print("\n".join(entries))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
