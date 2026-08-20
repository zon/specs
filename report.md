Build testutil fixtures from the source-owned layout

WriteSkill, WriteAgent, WriteAgentBody, WriteDoc, and WriteDocBody now
build their paths from source.RelPath instead of re-encoding each kind's
layout. The existing fixture tests already pin the written paths and
contents, so no new tests were needed.

Ralph item 2 completed