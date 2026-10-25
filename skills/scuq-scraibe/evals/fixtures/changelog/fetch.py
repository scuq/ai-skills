#!/usr/bin/env python3
"""fetch - download one URL to a file."""
import argparse
import sys
import time
import urllib.request

# 10 seconds is the longest wait that the operators accepted in the review.
DEFAULT_TIMEOUT = 10


def main(argv):
    parser = argparse.ArgumentParser(prog="fetch")
    parser.add_argument("url")
    parser.add_argument("out")
    parser.add_argument("--timeout", type=float, default=DEFAULT_TIMEOUT)
    args = parser.parse_args(argv)
    start = time.time()
    try:
        with urllib.request.urlopen(args.url, timeout=args.timeout) as resp, open(args.out, "wb") as f:
            f.write(resp.read())
    except TimeoutError:
        print(f"fetch: {args.url}: timeout after {args.timeout}s", file=sys.stderr)
        return 3
    except OSError as err:
        print(f"fetch: {args.url}: {err}", file=sys.stderr)
        return 1
    print(f"{args.out} {time.time() - start:.1f}s")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
