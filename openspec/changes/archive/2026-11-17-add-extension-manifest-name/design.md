## Context

The name of an installed extension is the name of its folder in the extension directory without the `c8y-` prefix (`Extension.Name()` in `pkg/cmd/extension/extension.go`). The folder name comes from:

- `--name`, if given (`c8y-` is added if missing).
- Otherwise the repository name for git and binary installs, which must start with or contain `c8y-` (`checkValidExtension`), or the source folder name for local installs (a symlink, or a file with the path on Windows).

The name is checked against built-in commands and installed extensions before anything is downloaded, because it's known up front. Binary installs already extract into a hidden temporary folder (`.tmp-install-*`) and move it into place, but git installs clone straight into the final folder. Local installs aren't checked at all.

Other parts depend on the installed name but not on how it was chosen: the command tree (`c8y <name>`), session provider references (`<extension>/<provider>`), and `c8y extensions update|delete <name>`. Binary upgrades find the repository through `manifest.yml` (owner, repository name), and git upgrades run `git pull` in the folder, so neither depends on the folder name.

## Goals / Non-Goals

**Goals:**

- Authors can choose the command name independently of the repository name.
- Users get that name without passing `--name`.
- The same name rules apply to git, binary, and local installs.
- Extensions without `name` and installed extensions behave as before.

**Non-Goals:**

- Renaming installed extensions, on update or otherwise. Users can delete the extension and install it again.
- Deriving names from repository names with other rules (removing `c8y-cli-`, using the part after the last `-`).
- Changing the names of extensions in system folders: package managers choose the folder name, and they should use `c8y-<name>`.
- Several names (aliases) for one extension.

## Decisions

### `name` is the name without the `c8y-` prefix

`name: analytics` is installed as `c8y-analytics`, the same way as `--name analytics`. The value is the command name, which is what authors and users see. A value with the `c8y-` prefix is accepted and the prefix removed, so `name: c8y-analytics` gives the same result. `c8y extensions validate` reports this as a warning, since the prefix is redundant.

The name is converted to lowercase, so `name: Analytics` gives `analytics`. Folder names on macOS and Windows are case-insensitive by default, so names which only differ in case would clash there, and lowercase names keep the command the same on every platform. After that, the name must match `^[a-z0-9][a-z0-9-]*$`, so it is a valid command name and folder name on every platform, without spaces, underscores, or path separators. If the name is invalid, the install fails with an error, so a broken name never becomes a folder.

Alternative: reject uppercase letters. Rejected because the intended name is clear, so failing the install doesn't help the user.

Alternative: the full name with the prefix (`name: c8y-analytics`). Rejected because the prefix is an implementation detail of the extension folder.

### Order of precedence: `--name`, then `name`, then the repository or folder name

`--name` is the user's explicit choice, for example to avoid a clash with another extension, so it wins. The manifest name is the author's choice and replaces the derived name. Without either, the name is derived as today.

### Resolve the name after the download, from the temporary folder

The name in `extension.yaml` is only known once the extension is downloaded. So all remote installs first download into a hidden temporary folder in the extension directory (`.tmp-install-*`, ignored when listing), then:

1. Read `extension.yaml` from the temporary folder and resolve the name.
2. Run the existing checks with the resolved name: built-in command, and an installed extension with the same name (system extensions excepted, as today).
3. Rename the temporary folder to `c8y-<name>`. On failure, remove the temporary folder.

Binary installs already work this way; only the target folder name is now chosen after extraction. Git installs change from `git clone <url> <final folder>` to a clone into the temporary folder, followed by the checkout of a pinned commit and the rename. A clone can be renamed, because git stores no absolute paths in the repository.

If `--name` is given, c8y keeps checking the name before the download, so a clash is reported without downloading anything. The check after the download is repeated anyway, because it costs nothing.

Alternative: read `extension.yaml` through the GitHub contents API before the download. Rejected because it only works for GitHub, needs an extra API call (rate limits), and for binary releases the manifest in the archive can differ from the one in the repository.

### The repository prefix is only required when the name comes from the repository

Today `c8y extensions install owner/my-tool` fails because the repository name doesn't contain `c8y-`. With this change, the check moves after the download, and it only applies when neither `--name` nor `name` is set. The error then says that the repository name must start with `c8y-` or that `extension.yaml` must set `name`.

This means a repository without the prefix and without `name` is downloaded before it's rejected. That's acceptable, because the download goes to the temporary folder, which is removed.

### Local installs use the same resolution and checks

`c8y extensions install <dir>` reads `<dir>/extension.yaml` and uses `--name`, then `name`, then the folder name. Local installs get the built-in command and installed-extension checks too, so a local `name: devices` can't hide the built-in `devices` command. Installing the same folder again, with the same resolved name, keeps working: the existing link is replaced, so authors can re-run `c8y extensions install .` after changing the name.

The link points to the source folder, so a later change of `name` in the source folder doesn't rename the link. `c8y extensions validate` reports this case (see below).

### Updates keep the installed name

`c8y extensions update` doesn't rename extensions: the folder name is what users type in scripts and aliases, and renaming it would break them without warning. If the update changes the declared name, and the new name differs from the installed name, c8y prints a notice with the new name and the commands to reinstall:

```
! extension analytics now declares the name 'dtm-analytics'. To use the new name, run:
  c8y extensions delete analytics
  c8y extensions install owner/c8y-cli-extension-analytics
```

c8y reads the declared name before and after the update (for binary extensions, after the new version is extracted; for git extensions, after `git pull`) and compares them. Comparing the two versions, instead of the new name with the installed name, means extensions installed with `--name` don't get the notice on every update, and c8y doesn't need to record whether `--name` was used.

### Validation

`c8y extensions validate` adds a `name` check:

- Error: the name doesn't match the format.
- Error: the name is the name of a built-in command.
- Warning: the name has the `c8y-` prefix or uppercase letters, with the name that is used.
- Info: the extension is installed under a different name than it declares, either because of `--name` or because the declared name changed. This is not a warning, because `--name` is a valid user choice.

The check runs for folders (`c8y extensions validate .`) and installed extensions. For the built-in command check, it uses the root command, like the install check.

### The default binary path is unchanged

Without `binary.path`, the binary is `bin/<original name>`: the repository name from `manifest.yml` for binary installs, or the folder name for git and local installs. For a git install with a manifest name, this is now `c8y-<name>`, the same as with `--name` today. The scaffolding templates always set `binary.path`, and the docs recommend setting it, so the binary path doesn't depend on how the extension is installed.

### Scaffolding writes the name

`c8y extensions create c8y-foo` writes `name: foo` into `extension.yaml` for all templates. The value is the name passed to `create` without the prefix, so the generated extension behaves as before, and authors can see where to change the name.

## Risks / Trade-offs

- [An extension changes its name in a new version, and users are left with the old name] → Updates keep the name and print a notice with the reinstall commands. `validate` shows the difference.
- [Two extensions declare the same name] → The second install fails with the existing "already an installed extension" error. Users can install it with `--name`.
- [A repository without `c8y-` in its name is downloaded and then rejected] → The download goes to a temporary folder, which is removed. The error says how to fix it.
- [Dry run (`--dry`) can't show the declared name, because nothing is downloaded] → The dry run message says the name may change to the one declared in `extension.yaml`.
- [Moving a git clone fails across file systems] → The temporary folder is in the extension directory, so the rename is on the same file system, as for binary installs.

## Migration Plan

No migration is needed. Installed extensions keep their folders and names. Extensions without `name` are installed as before. Authors can add `name` to `extension.yaml` at any time, and it applies to new installs.

## Open Questions

- Should `c8y extensions list` show the declared name as well as the installed name when they differ? Not planned: `validate` reports it, and the list stays as it is.
