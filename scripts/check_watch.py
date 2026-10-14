"""Debounce edits and recheck affected projects, including additions and deletions."""

import time

from source_files import ROOT, affected_services, inventory


def snapshot():
    result = {}
    for name in inventory():
        path = ROOT / name
        try:
            stat = path.stat()
        except FileNotFoundError:
            continue
        result[name] = (stat.st_mtime_ns, stat.st_size)
    return result


def difference(before, after):
    return {name for name in before.keys() | after.keys() if before.get(name) != after.get(name)}


def watch(services):
    from check import check
    from source_files import SERVICES

    before = snapshot()
    check(list(services or SERVICES), full=not services)
    print("Watching source files; Ctrl-C stops checks", flush=True)
    pending = None
    while True:
        time.sleep(0.3)
        after = snapshot()
        if after == before:
            pending = None
            continue
        if after != pending:
            pending = after
            continue
        names = difference(before, after)
        before = after
        selected = affected_services(names)
        if services:
            selected = [service for service in selected if service in services]
        if selected or not services:
            shared = any(not name.startswith("services/") for name in names)
            check(selected, full=shared and not services)
