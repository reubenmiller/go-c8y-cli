## 1. Views

- [x] 1.1 Add `autoMatch` to view definitions and skip views with `autoMatch: false` in automatic detection
- [x] 1.2 Add `SetActiveView` and `GetDefinitionByName` to the data view
- [x] 1.3 Unit tests for `autoMatch`, the active view, and namespaced lookup

## 2. API commands

- [x] 2.1 Add `view` (`name`, `columns`) to the API command model, and the extension name of the command
- [x] 2.2 Apply the command's view when the view option is `auto`, resolving names in the command's extension first
- [x] 2.3 Set the extension name when loading extension API commands
- [x] 2.4 Unit tests for precedence, name resolution, unknown names, and parsing

## 3. Schemas and docs

- [x] 3.1 Add `view` to `extensionCommands.json` and `autoMatch` to `views.json`
- [x] 3.2 Document the command view on the API command groups page and `autoMatch` on the views page
