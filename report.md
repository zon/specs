Accept --source and --target flags on the update command

The update command accepts a --source flag for a local source directory and
a --target flag naming the claude or opencode target (default: opencode).
Flags take the "--flag value" and "--flag=value" forms and interleave with
the scope word. Unknown flags, missing values, and unknown targets print
usage to stderr and exit non-zero.

Tests cover target and option parsing, command recognition with flags at the
run level, and built-binary runs that exercise both flags.

Ralph item 18 completed
