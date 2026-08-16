Read agent definitions from agents/<name>.md

The system reads an agent definition only from its canonical path under
a source. The source reader lists each agent keyed by the file name
that holds it, and reports the path it read from.

Added tests assert the reader finds an agent at the canonical path and
reports that path.

Ralph item 3 completed
