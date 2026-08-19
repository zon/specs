Make targetdir.Write return only an error

- Drop the written bool from Write and always record the path in owned
- Update the tests that asserted on the bool
- Remove the completed plan from refactoring.md

Ralph item 2 completed
