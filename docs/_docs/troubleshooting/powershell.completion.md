---
layout: default
category: Troubleshooting
title: PowerShell TabCompletion
---

### Shows navigable menu of all options when hitting Tab

The PowerShell advanced tab completion is not enabled on Linux and MacOS, however it can be activated by changing the default tab-completion key to **tab**.

```powershell
Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete

# Autocompletion for arrow keys
Set-PSReadlineKeyHandler -Key UpArrow -Function HistorySearchBackward

Set-PSReadlineKeyHandler -Key DownArrow -Function HistorySearchForward

Set-PSReadLineOption -HistorySearchCursorMovesToEnd
```

**Note**

You can add custom auto completion settings your PowerShell profile by adding them

```powershell
$PROFILE
```

For all of the `PSReadLine` options, view the [online documentation](https://docs.microsoft.com/en-us/powershell/module/PSReadline/Set-PSReadlineOption?view=powershell-7&viewFallbackFrom=powershell-5.0)
