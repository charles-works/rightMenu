# RightMenu / DEV调试

`rightmenu.exe` is a small Windows x64 utility that adds a personal `DEV调试` file right-click menu. Menu items are configured in JSON. Choosing a child item launches the configured program with the configured DEV argument contract for the selected file.

## First-version scope

Included:

- Single-file Explorer context menu named `DEV调试`.
- Configurable child commands, for example `AA`.
- `设置` menu item that opens the config file.
- User-scope install/uninstall under `HKCU`; no administrator rights intended.
- GitHub Actions Windows x64 build artifact.

Not included in v1:

- No settings GUI.
- No MSI/team installer/auto-update.
- No COM/IExplorerCommand shell extension.
- No multi-file selection support.
- No replacement for Windows default file associations or `打开方式`.

## Download artifact

The release workflow packages:

```text
rightmenu-windows-amd64.zip
  rightmenu.exe
  examples/config.json
  scripts/install.ps1
  scripts/uninstall.ps1
  README.md
```

Extract the zip to a stable directory before installing.

## Canonical paths

`rightmenu install` pins/copies the executable and creates config if needed:

- Pinned executable: `%LOCALAPPDATA%\RightMenu\rightmenu.exe`
- Config file: `%APPDATA%\RightMenu\config.json`
- Log file: `%APPDATA%\RightMenu\rightmenu.log` by default
- Registry menu subtree: `HKCU\Software\Classes\*\shell\DEVDebug`

Moving the original extracted zip directory after install is okay because Explorer commands use the pinned executable path.

## Configuration

Edit `%APPDATA%\RightMenu\config.json`:

```json
{
  "menuTitle": "DEV调试",
  "logging": {
    "enabled": true,
    "path": ""
  },
  "items": [
    {
      "id": "aa",
      "title": "AA",
      "program": "C:\\Tools\\AA.exe",
      "specifiedFolder": "C:\\DEV"
    }
  ]
}
```

Rules:

- `logging` is a global optional object:
  - `enabled` defaults to `true`; set it to `false` to disable run logging.
  - `path` is optional. Empty or omitted uses `%APPDATA%\RightMenu\rightmenu.log`.
  - Each run appends one JSON line containing the run time, selected file path, selected filename, target program, argument array, and formatted full command.
- `id` must be unique and match `^[A-Za-z0-9._-]+$`.
- `title` is shown in the submenu.
- `program` is the target executable path.
- `specifiedFolder` is the per-item folder path passed as argument 3 when using the default DEV argument contract.
- `args` is an optional advanced override. When `args` is omitted or empty, `specifiedFolder` is required and the target receives exactly six arguments:
  1. `" "` (single space)
  2. `"IF_A_000N"`
  3. `specifiedFolder`
  4. selected file's containing folder path
  5. selected file's filename only
  6. `"Q"`
- If `args` is present, it replaces the default six-argument contract for compatibility.
- In custom `args`, `{file}` expands to the selected full file path.

### Log example

With default logging enabled, each `run` appends one JSON line to `%APPDATA%\RightMenu\rightmenu.log` unless `logging.path` overrides it:

```json
{"time":"2026-05-15T01:02:03Z","selectedFile":"C:\\Temp\\path with spaces\\file.txt","selectedFileName":"file.txt","program":"C:\\Tools\\AA.exe","args":[" ","IF_A_000N","C:\\DEV","C:\\Temp\\path with spaces","file.txt","Q"],"command":"\"C:\\Tools\\AA.exe\" \" \" IF_A_000N C:\\DEV \"C:\\Temp\\path with spaces\" file.txt Q"}
```

Use a custom log path or disable logging globally:

```json
{
  "logging": {
    "enabled": true,
    "path": "C:\\Logs\\rightmenu.log"
  }
}
```

### Custom args example

If a target program needs the old/simple selected-file argument, set `args` explicitly. This bypasses the default six-argument contract for that item:

```json
{
  "id": "aa-legacy",
  "title": "AA Legacy",
  "program": "C:\\Tools\\AA.exe",
  "args": ["--open", "{file}", "--mode", "debug"]
}
```

For selected file `C:\Temp\path with spaces\file.txt`, the target receives:

```text
--open
C:\Temp\path with spaces\file.txt
--mode
debug
```

After editing config, run:

```powershell
rightmenu.exe refresh
```

`refresh` rewrites the static registry submenu entries from the current config.

## Commands

```powershell
rightmenu.exe install      # pin exe, create config if missing, register menu
rightmenu.exe refresh      # rebuild menu entries from config
rightmenu.exe uninstall    # remove owned DEV调试 menu registry subtree
rightmenu.exe config       # open config file
rightmenu.exe paths        # print canonical paths
```

PowerShell wrappers are also included:

```powershell
.\scripts\install.ps1
.\scripts\uninstall.ps1
```

## Registry design

The v1 menu uses a static HKCU registry cascade:

```text
HKCU\Software\Classes\*\shell\DEVDebug
  MUIVerb = DEV调试
  Icon = %LOCALAPPDATA%\RightMenu\rightmenu.exe
  MultiSelectModel = Single
  ExtendedSubCommandsKey
    Shell
      <item-id>
        MUIVerb = <configured title>
        command
          (Default) = "%LOCALAPPDATA%\RightMenu\rightmenu.exe" run "<item-id>" "%1"
      settings
        MUIVerb = 设置
        command
          (Default) = "%LOCALAPPDATA%\RightMenu\rightmenu.exe" config
```

Uninstall removes only `HKCU\Software\Classes\*\shell\DEVDebug`; it leaves your config file in place.

## Windows 11 context menu caveat

Static registry verbs may appear under **Show more options** / the legacy context menu on some Windows 11 systems. A primary modern Windows 11 menu extension would require a heavier COM/IExplorerCommand-style implementation, which is intentionally out of scope for this first personal-tool version.

## Smoke test

Use a harmless target program/script in config, then run:

```powershell
.\rightmenu.exe install
reg query "HKCU\Software\Classes\*\shell\DEVDebug" /s
.\rightmenu.exe refresh
.\rightmenu.exe run aa "C:\Temp\path with spaces\sample file.txt"
.\rightmenu.exe config
.\rightmenu.exe uninstall
reg query "HKCU\Software\Classes\*\shell\DEVDebug" /s
```

Expected results:

- The first `reg query` shows `DEV调试`, `MultiSelectModel = Single`, child commands, and `设置`.
- The `run aa ...` command launches your configured target with the default six DEV arguments, including the selected file's folder and filename.
- The final `reg query` fails because uninstall removed the owned menu subtree.

For full Explorer verification, right-click a file whose path contains spaces, choose `DEV调试 -> AA`, and confirm the target receives the exact six arguments. If you configured custom `args`, confirm that override receives the selected path through `{file}` as expected.

## Development

```bash
go test ./...
GOOS=windows GOARCH=amd64 go build -o dist/rightmenu.exe ./cmd/rightmenu
```

Linux/macOS builds can run portable tests, but registry install/uninstall is Windows-only.
