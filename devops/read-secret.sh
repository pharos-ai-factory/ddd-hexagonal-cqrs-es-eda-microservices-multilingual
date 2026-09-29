#!/bin/sh
# NAME_FILE is the runtime secret interface; errors never include its contents.
read_secret() {
 name=$1
 direct=$(printenv "$name" || true)
 file=$(printenv "${name}_FILE" || true)
 if printenv "$name" >/dev/null && printenv "${name}_FILE" >/dev/null; then
  echo "$name has ambiguous secret sources" >&2; return 1
 fi
 if [ -n "$file" ]; then
  direct=$(cat "$file" 2>/dev/null) || { echo "$name cannot be read" >&2; return 1; }
 fi
 [ -n "$direct" ] || { echo "$name is required" >&2; return 1; }
 printf '%s' "$direct"
}
