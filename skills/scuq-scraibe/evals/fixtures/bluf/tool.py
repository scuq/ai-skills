#!/usr/bin/env python3
"""tool - list, count, and export the entries of a directory."""
import argparse
import csv
import os
import sys


def main(argv):
    parser = argparse.ArgumentParser(prog="tool")
    sub = parser.add_subparsers(dest="command", required=True)
    lister = sub.add_parser("list", help="list the entries of DIR")
    lister.add_argument("dir")
    lister.add_argument("--limit", type=int, default=0, help="print at most N entries")
    counter = sub.add_parser("count", help="print the number of entries of DIR")
    counter.add_argument("dir")
    exporter = sub.add_parser("export", help="write the entries of DIR to the CSV file FILE")
    exporter.add_argument("dir")
    exporter.add_argument("file")
    args = parser.parse_args(argv)
    try:
        entries = sorted(os.listdir(args.dir))
    except OSError as err:
        print(f"tool: {err}", file=sys.stderr)
        return 1
    if args.command == "count":
        print(len(entries))
    elif args.command == "export":
        with open(args.file, "w", newline="") as handle:
            writer = csv.writer(handle)
            writer.writerow(["name", "size"])
            for name in entries:
                writer.writerow([name, os.path.getsize(os.path.join(args.dir, name))])
    else:
        if args.limit:
            entries = entries[:args.limit]
        print("\n".join(entries))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
