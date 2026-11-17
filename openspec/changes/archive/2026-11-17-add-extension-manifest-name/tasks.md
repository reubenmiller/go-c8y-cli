## 1. Manifest

- [x] 1.1 Add `Name` to `extensions.Manifest` (`pkg/extensions/extension.go`) and a helper which removes a leading `c8y-`, converts the name to lowercase, and checks the format `^[a-z0-9][a-z0-9-]*$`, with unit tests (valid, prefix, uppercase converted, spaces, underscore, empty)
- [x] 1.2 Add a function which reads the declared name from `extension.yaml` in a folder (missing file or missing `name` returns an empty name, invalid YAML returns an error)
- [x] 1.3 Add the `name` property to `tools/schema/extension.json`

## 2. Name resolution on install

- [x] 2.1 Add a function which resolves the install name from `--name`, the declared name, and the repository or folder name, and returns whether the repository prefix rule applies, with unit tests for the order of precedence
- [x] 2.2 Split `checkValidExtension` so the built-in command and installed-extension checks can run with a resolved name after the download, and the repository prefix check only runs when the name comes from the repository
- [x] 2.3 Change `installGit` to clone into a hidden temporary folder in the extension directory, check out a pinned commit and write the pin file there, then resolve and check the name and rename the folder to `c8y-<name>`; remove the temporary folder on any failure
- [x] 2.4 Change `installBinRelease` to choose the target folder after extraction from the resolved name (keeping `manifest.yml` with the repository name, so updates and the default binary path still work)
- [x] 2.5 Change `InstallLocal` and the local branch of `install` to use the resolved name, run the same checks, and replace an existing link to the same folder
- [x] 2.6 Update the `install` command: keep the early check when `--name` is given, print the resolved name in the success message and the validation summary, and update the dry run message
- [x] 2.7 Update the `install` help text and the extension command's long description (repository prefix, `name` in `extension.yaml`)

## 3. Updates

- [x] 3.1 Read the declared name before and after updating git and binary extensions, and print the reinstall notice when it changed and differs from the installed name
- [x] 3.2 Unit tests: notice on a changed name, no notice for an unchanged name (also with `--name`), and the extension keeps its folder

## 4. Validation

- [x] 4.1 Add the `name` check to `c8y extensions validate` (invalid format, built-in command, `c8y-` prefix or uppercase letters as warnings, installed name differs as info) and include it in the install summary
- [x] 4.2 Unit tests for each finding, and that the info finding doesn't fail `--strict`

## 5. Scaffolding

- [x] 5.1 Write `name: <name>` into the `extension.yaml` templates for `--precompiled go` and `--precompiled other`
- [x] 5.2 Turn the script template's `extension.yaml` (`ext_tmpls/extension.yaml`) into a template with `name`, and fix its outdated schema URL (`feat/extensions-manager` branch)
- [x] 5.3 Extend the scaffold tests so the generated `extension.yaml` contains the name and validates without findings

## 6. Tests and docs

- [x] 6.1 Unit tests for installs with a fake git and a local release server: declared name, `--name` override, no name, invalid name, clash with a built-in command, clash with an installed extension, repository without the prefix (with and without `name`), pinned git install, and no leftover temporary folders
- [x] 6.2 Commander tests in `tests/manual/extensions`: local install with a declared name, reinstall of the same folder, and `validate` name findings
- [x] 6.3 Docs: naming of extensions (the `name` field, order of precedence, repository prefix, updates) on the extension authoring pages, and a note for package managers to use `c8y-<name>` as the folder name in system folders
- [x] 6.4 Regenerate the command docs for `c8y extensions install`, and run `openspec validate add-extension-manifest-name --strict`
