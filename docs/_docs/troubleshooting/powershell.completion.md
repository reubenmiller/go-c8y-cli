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
```
