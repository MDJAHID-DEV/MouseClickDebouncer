Mouse Click Debouncer v3

A lightweight Windows utility that filters accidental repeated left-clicks.

Features:
- Global left-click debounce filter
- Adjustable 50 / 100 / 150 / 200 / 300 / 500 ms interval
- System tray operation
- Minimal modern settings UI with custom-drawn controls
- Single-instance protection
- Low-overhead timing path using GetTickCount64

Developer: Miah Muhammad Zahid
Copyright (c) 2026 Miah Muhammad Zahid. All rights reserved.

Build in VS Code / Windows:
  1. Install Go.
  2. Install go-winres once:
       go install github.com/tc-hib/go-winres@v0.3.3
  3. Run build_windows.bat

The winres/winres.json file sets CompanyName, ProductName, FileDescription,
version, copyright, and a Windows 10+ per-monitor-v2 manifest.

Important:
The Windows UAC "Publisher" field is trust/signature based. A real trusted
publisher name requires a code-signing certificate for Miah Muhammad Zahid.
The version metadata above sets the CompanyName shown in file Properties -> Details.
