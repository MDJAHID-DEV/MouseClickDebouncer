package main

import (
    "fmt"
    "runtime"
    "sync/atomic"
    "syscall"
    "time"
    "unsafe"
)

// Mouse Click Debouncer - native Win32 tray + settings window.
// Coded by Miah Muhammad Zahid.
// Copyright © 2026 Miah Muhammad Zahid. All rights reserved.

const (
    WH_MOUSE_LL      = 14
    WM_MOUSEMOVE     = 0x0200
    WM_LBUTTONDOWN   = 0x0201
    WM_LBUTTONUP     = 0x0202
    WM_RBUTTONUP     = 0x0205
    WM_COMMAND       = 0x0111
    WM_CLOSE         = 0x0010
    WM_DESTROY       = 0x0002
    WM_PAINT         = 0x000F
    WM_ERASEBKGND    = 0x0014
    WM_SETFONT       = 0x0030
    WM_CTLCOLORSTATIC = 0x0138
    WM_APP           = 0x8000
    WM_TRAYICON      = WM_APP + 1
    WM_SHOWWINDOW    = 0x0018
    WM_HOTKEY        = 0x0312

    NIM_ADD          = 0
    NIM_MODIFY       = 1
    NIM_DELETE       = 2
    NIF_MESSAGE      = 1
    NIF_ICON         = 2
    NIF_TIP          = 4

    MF_STRING        = 0x0000
    MF_SEPARATOR     = 0x0800
    MF_CHECKED       = 0x0008
    MF_UNCHECKED     = 0x0000
    TPM_LEFTALIGN    = 0x0000
    TPM_BOTTOMALIGN  = 0x0020
    TPM_RETURNCMD    = 0x0100
    TPM_NONOTIFY     = 0x0080

    IDI_APPLICATION  = 32512
    ID_ENABLE        = 1001
    ID_50            = 1050
    ID_100           = 1100
    ID_150           = 1150
    ID_200           = 1200
    ID_300           = 1300
    ID_500           = 1500
    ID_SHOW          = 1800
    ID_ABOUT         = 1900
    ID_EXIT          = 1999

    IDC_ENABLE       = 2001
    IDC_COMBO        = 2002
    IDC_MINIMIZE     = 2003
    IDC_EXITBTN      = 2004

    WS_OVERLAPPEDWINDOW = 0x00CF0000
    WS_VISIBLE          = 0x10000000
    WS_CHILD            = 0x40000000
    WS_TABSTOP          = 0x00010000
    WS_EX_TOOLWINDOW    = 0x00000080
    BS_AUTOCHECKBOX     = 0x00000003
    BS_PUSHBUTTON       = 0x00000000
    CBS_DROPDOWNLIST    = 0x0003
    ES_AUTOHSCROLL      = 0x0080
    SS_LEFT             = 0x00000000
    SS_CENTER           = 0x00000001
    SW_HIDE             = 0
    SW_SHOW             = 5
    SW_MINIMIZE         = 6
    SW_RESTORE          = 9
    COLOR_WINDOW        = 5
    COLOR_BTNFACE       = 15
    DEFAULT_GUI_FONT    = 17
    MB_OK               = 0x00000000
    MB_ICONINFORMATION  = 0x00000040
    GWLP_USERDATA       = -21
    CW_USEDEFAULT       = 0x80000000
)

type point struct { x, y int32 }
type msg struct { hwnd uintptr; message uint32; wParam, lParam uintptr; time uint32; pt point; lPrivate uint32 }
type mouseHookStruct struct { pt point; mouseData, flags, time uint32; dwExtraInfo uintptr }
type wndClass struct { style uint32; lpfnWndProc uintptr; cbClsExtra, cbWndExtra int32; hInstance, hIcon, hCursor, hbrBackground uintptr; lpszMenuName, lpszClassName *uint16 }
type notifyIconData struct { cbSize uint32; hWnd uintptr; uID uint32; uFlags, uCallbackMessage uint32; hIcon uintptr; szTip [128]uint16; dwState, dwStateMask uint32; szInfo [256]uint16; uTimeoutOrVersion uint32; szInfoTitle [64]uint16; dwInfoFlags uint32; guid [16]byte; hBalloonIcon uintptr }
type menuItemInfo struct { cbSize uint32; fMask uint32; fType uint32; fState uint32; wID uint32; hSubMenu, hbmpChecked, hbmpUnchecked, dwItemData uintptr; dwTypeData *uint16; cch uint32; hbmpItem uintptr }
type paintStruct struct { hdc uintptr; erase uint32; rcPaint rect; restore uint32; incUpdate uint32; rgbReserved [32]byte }
type rect struct { left, top, right, bottom int32 }
type textMetric struct { height, ascent, descent, internalLeading, externalLeading, avgWidth, maxWidth, weight, overhang, digitizedAspectX, digitizedAspectY int32; firstChar, lastChar, defaultChar, breakChar uint16; italic, underlined, struckOut, pitchAndFamily, charSet byte }

type windowState struct { visible bool }

var (
    user32 = syscall.NewLazyDLL("user32.dll")
    kernel32 = syscall.NewLazyDLL("kernel32.dll")
    shell32 = syscall.NewLazyDLL("shell32.dll")
    gdi32 = syscall.NewLazyDLL("gdi32.dll")

    pSetWindowsHookEx = user32.NewProc("SetWindowsHookExW")
    pCallNextHookEx = user32.NewProc("CallNextHookEx")
    pUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
    pGetMessage = user32.NewProc("GetMessageW")
    pTranslateMessage = user32.NewProc("TranslateMessage")
    pDispatchMessage = user32.NewProc("DispatchMessageW")
    pPostQuitMessage = user32.NewProc("PostQuitMessage")
    pPostMessage = user32.NewProc("PostMessageW")
    pRegisterClass = user32.NewProc("RegisterClassW")
    pCreateWindowEx = user32.NewProc("CreateWindowExW")
    pDefWindowProc = user32.NewProc("DefWindowProcW")
    pDestroyWindow = user32.NewProc("DestroyWindow")
    pShowWindow = user32.NewProc("ShowWindow")
    pUpdateWindow = user32.NewProc("UpdateWindow")
    pLoadIcon = user32.NewProc("LoadIconW")
    pGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
    pCreateMutex = kernel32.NewProc("CreateMutexW")
    pGetLastError = kernel32.NewProc("GetLastError")
    pCloseHandle = kernel32.NewProc("CloseHandle")
    pGetTickCount64 = kernel32.NewProc("GetTickCount64")
    pShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")
    pCreatePopupMenu = user32.NewProc("CreatePopupMenu")
    pAppendMenu = user32.NewProc("AppendMenuW")
    pTrackPopupMenu = user32.NewProc("TrackPopupMenu")
    pDestroyMenu = user32.NewProc("DestroyMenu")
    pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
    pGetCursorPos = user32.NewProc("GetCursorPos")
    pMessageBox = user32.NewProc("MessageBoxW")
    pSetWindowText = user32.NewProc("SetWindowTextW")
    pGetDlgItem = user32.NewProc("GetDlgItem")
    pSendMessage = user32.NewProc("SendMessageW")
    pGetClientRect = user32.NewProc("GetClientRect")
    pBeginPaint = user32.NewProc("BeginPaint")
    pEndPaint = user32.NewProc("EndPaint")
    pFillRect = user32.NewProc("FillRect")
    pCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
    pDeleteObject = gdi32.NewProc("DeleteObject")
    pCreateFont = gdi32.NewProc("CreateFontW")
    pSelectObject = gdi32.NewProc("SelectObject")
    pSetTextColor = gdi32.NewProc("SetTextColor")
    pSetBkMode = gdi32.NewProc("SetBkMode")
    pRoundRect = gdi32.NewProc("RoundRect")
    pEllipse = gdi32.NewProc("Ellipse")
    pMoveToEx = gdi32.NewProc("MoveToEx")
    pLineTo = gdi32.NewProc("LineTo")
    pGetStockObject = gdi32.NewProc("GetStockObject")
    pSetBkColor = gdi32.NewProc("SetBkColor")
)

var (
    enabled atomic.Bool
    debounceMs atomic.Int64
    lastAccepted atomic.Uint64
    suppressing atomic.Bool
    hookHandle uintptr
    hwnd uintptr
    instance uintptr
    className = syscall.StringToUTF16Ptr("MouseClickDebouncerMainWindow")
    trayClassName = syscall.StringToUTF16Ptr("MouseClickDebouncerTrayWindow")
    mainFont uintptr
    titleFont uintptr
    appMutex uintptr
    combo uintptr
    check uintptr
)

func utf16(s string) uintptr { return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s))) }

func setWindowText(h uintptr, s string) { pSetWindowText.Call(h, utf16(s)) }

func createFont(height int32, weight int32) uintptr {
    r, _, _ := pCreateFont.Call(uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 0, 0, 0, 0, 0, utf16("Segoe UI"))
    return r
}

func sendFont(h uintptr, font uintptr) { if h != 0 { pSendMessage.Call(h, WM_SETFONT, font, 1) } }

func wndProc(hWnd, uMsg, wParam, lParam uintptr) uintptr {
    switch uMsg {
    case WM_TRAYICON:
        if lParam == WM_RBUTTONUP { showMenu(); return 0 }
        if lParam == WM_LBUTTONUP { showWindow(); return 0 }
    case WM_LBUTTONUP:
        // Custom-drawn controls keep the UI lightweight and consistent across Windows themes.
        x := int32(int16(lParam & 0xffff))
        y := int32(int16((lParam >> 16) & 0xffff))
        switch {
        case x >= 270 && x <= 450 && y >= 252 && y <= 286:
            enabled.Store(!enabled.Load()); syncUI()
        case x >= 270 && x <= 445 && y >= 300 && y <= 332:
            showDebounceMenu()
        case x >= 38 && x <= 188 && y >= 365 && y <= 405:
            hideWindow()
        case x >= 198 && x <= 288 && y >= 365 && y <= 405:
            cleanup(); pPostQuitMessage.Call(0)
        }
        return 0
    case WM_COMMAND:
        cmd := uint32(wParam & 0xffff)
        switch cmd {
        case ID_ENABLE:
            enabled.Store(!enabled.Load()); syncUI()
        case ID_50: setDebounce(50)
        case ID_100: setDebounce(100)
        case ID_150: setDebounce(150)
        case ID_200: setDebounce(200)
        case ID_300: setDebounce(300)
        case ID_500: setDebounce(500)
        case ID_SHOW:
            showWindow()
        case ID_ABOUT:
            showAbout()
        case ID_EXIT:
            cleanup(); pPostQuitMessage.Call(0)
        case IDC_ENABLE:
            if (wParam>>16) == 0 { enabled.Store(isChecked(check)); syncUI() }
        case IDC_COMBO:
            if (wParam>>16) == 1 { setDebounce(comboSelection()) }
        case IDC_MINIMIZE:
            hideWindow()
        case IDC_EXITBTN:
            cleanup(); pPostQuitMessage.Call(0)
        }
        return 0
    case WM_CLOSE:
        hideWindow(); return 0
    case WM_PAINT:
        return paintMain(hWnd)
    case WM_ERASEBKGND:
        return 1
    case WM_DESTROY:
        return 0
    case WM_CTLCOLORSTATIC:
        hdc := wParam
        pSetBkMode.Call(hdc, 1)
        pSetTextColor.Call(hdc, 0x00333333)
        brush, _, _ := pCreateSolidBrush.Call(0x00F7F8FA)
        return brush
    }
    r, _, _ := pDefWindowProc.Call(hWnd, uMsg, wParam, lParam)
    return r
}

func paintMain(hWnd uintptr) uintptr {
    var ps paintStruct
    hdc, _, _ := pBeginPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
    if hdc == 0 { return 0 }
    defer pEndPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))

    var rc rect
    pGetClientRect.Call(hWnd, uintptr(unsafe.Pointer(&rc)))

    // Clean, soft neutral background.
    bg, _, _ := pCreateSolidBrush.Call(0x00F7F8FC)
    pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bg)
    pDeleteObject.Call(bg)

    // Header card.
    cardBrush, _, _ := pCreateSolidBrush.Call(0x00FFFFFF)
    header := rect{20, 18, rc.right-20, 108}
    pFillRect.Call(hdc, uintptr(unsafe.Pointer(&header)), cardBrush)
    pDeleteObject.Call(cardBrush)

    // Small orange mouse logo.
    accent, _, _ := pCreateSolidBrush.Call(0x00457AFF)
    old, _, _ := pSelectObject.Call(hdc, accent)
    pRoundRect.Call(hdc, 38, 35, 78, 87, 18, 18)
    pSelectObject.Call(hdc, old)
    pDeleteObject.Call(accent)
    white, _, _ := pCreateSolidBrush.Call(0x00FFFFFF)
    old, _, _ = pSelectObject.Call(hdc, white)
    pRoundRect.Call(hdc, 44, 41, 72, 80, 13, 13)
    pSelectObject.Call(hdc, old)
    pDeleteObject.Call(white)
    wheel, _, _ := pCreateSolidBrush.Call(0x00457AFF)
    old, _, _ = pSelectObject.Call(hdc, wheel)
    pRoundRect.Call(hdc, 55, 45, 61, 59, 3, 3)
    pSelectObject.Call(hdc, old)
    pDeleteObject.Call(wheel)

    drawText(hdc, 94, 30, "Mouse Click Debouncer", titleFont, 0x001D2430)
    drawText(hdc, 94, 62, "Quietly blocks accidental left-click double presses", mainFont, 0x006B7280)

    // Status card.
    statusCard := rect{20, 122, rc.right-20, 210}
    statusBrush, _, _ := pCreateSolidBrush.Call(0x00FFFFFF)
    pFillRect.Call(hdc, uintptr(unsafe.Pointer(&statusCard)), statusBrush)
    pDeleteObject.Call(statusBrush)
    drawText(hdc, 38, 136, "FILTER STATUS", mainFont, 0x00727B8A)

    status := "Filtering is OFF"
    if enabled.Load() { status = "Filtering is ON" }
    statusColor := uintptr(0x005B626B)
    if enabled.Load() { statusColor = 0x002E9B66 }
    drawText(hdc, 38, 161, status, titleFont, statusColor)
    drawText(hdc, 250, 166, fmt.Sprintf("%d ms", debounceMs.Load()), mainFont, 0x00457AFF)
    drawText(hdc, 38, 190, "The first left click is accepted; rapid repeats are ignored.", mainFont, 0x006B7280)

    // Settings card.
    settings := rect{20, 224, rc.right-20, 350}
    settingsBrush, _, _ := pCreateSolidBrush.Call(0x00FFFFFF)
    pFillRect.Call(hdc, uintptr(unsafe.Pointer(&settings)), settingsBrush)
    pDeleteObject.Call(settingsBrush)
    drawText(hdc, 38, 238, "SETTINGS", mainFont, 0x00727B8A)
    drawText(hdc, 38, 263, "Left-click filtering", mainFont, 0x001D2430)
    // Custom checkbox.
    cbBrush, _, _ := pCreateSolidBrush.Call(0x00D8DEE8)
    oldCB, _, _ := pSelectObject.Call(hdc, cbBrush)
    pRoundRect.Call(hdc, 270, 252, 292, 274, 6, 6)
    pSelectObject.Call(hdc, oldCB)
    pDeleteObject.Call(cbBrush)
    if enabled.Load() {
        a, _, _ := pCreateSolidBrush.Call(0x00457AFF); oldA, _, _ := pSelectObject.Call(hdc, a)
        pRoundRect.Call(hdc, 270, 252, 292, 274, 6, 6); pSelectObject.Call(hdc, oldA); pDeleteObject.Call(a)
        drawText(hdc, 274, 251, "✓", mainFont, 0x00FFFFFF)
    } else {
        inner, _, _ := pCreateSolidBrush.Call(0x00FFFFFF); oldI, _, _ := pSelectObject.Call(hdc, inner)
        pRoundRect.Call(hdc, 272, 254, 290, 272, 5, 5); pSelectObject.Call(hdc, oldI); pDeleteObject.Call(inner)
    }
    drawText(hdc, 302, 252, "Enable filtering", mainFont, 0x001D2430)

    drawText(hdc, 38, 304, "Debounce interval", mainFont, 0x001D2430)
    // Custom dropdown.
    drop, _, _ := pCreateSolidBrush.Call(0x00F8F9FC); oldD, _, _ := pSelectObject.Call(hdc, drop)
    pRoundRect.Call(hdc, 270, 298, 445, 332, 7, 7); pSelectObject.Call(hdc, oldD); pDeleteObject.Call(drop)
    drawText(hdc, 284, 304, fmt.Sprintf("%d ms", debounceMs.Load()), mainFont, 0x001D2430)
    drawText(hdc, 417, 303, "▾", mainFont, 0x006B7280)
    drawText(hdc, 38, 329, "150 ms is a good starting point for most double-click faults.", mainFont, 0x006B7280)

    // Minimal custom action buttons.
    minBrush, _, _ := pCreateSolidBrush.Call(0x00EEF1F6); oldM, _, _ := pSelectObject.Call(hdc, minBrush)
    pRoundRect.Call(hdc, 38, 365, 188, 405, 8, 8); pSelectObject.Call(hdc, oldM); pDeleteObject.Call(minBrush)
    drawText(hdc, 62, 372, "Minimize to tray", mainFont, 0x001D2430)
    exitBrush, _, _ := pCreateSolidBrush.Call(0x00457AFF); oldE, _, _ := pSelectObject.Call(hdc, exitBrush)
    pRoundRect.Call(hdc, 198, 365, 288, 405, 8, 8); pSelectObject.Call(hdc, oldE); pDeleteObject.Call(exitBrush)
    drawText(hdc, 226, 372, "Exit", mainFont, 0x00FFFFFF)

    // Footer.
    drawText(hdc, 20, rc.bottom-27, "Coded by Miah Muhammad Zahid  •  © 2026", mainFont, 0x00727B8A)
    return 0
}

func drawText(hdc uintptr, x, y int32, s string, font uintptr, color uintptr) {
    old, _, _ := pSelectObject.Call(hdc, font)
    pSetTextColor.Call(hdc, color)
    pSetBkMode.Call(hdc, 1)
    proc := gdi32.NewProc("TextOutW")
    u := syscall.StringToUTF16(s)
    proc.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1))
    pSelectObject.Call(hdc, old)
}

func mouseProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
    if nCode >= 0 && enabled.Load() {
        switch wParam {
        case WM_LBUTTONDOWN:
            now := uint64(getTick())
            last := lastAccepted.Load()
            interval := uint64(debounceMs.Load())
            if last != 0 && now >= last && now-last < interval {
                suppressing.Store(true)
                return 1
            }
            lastAccepted.Store(now)
            suppressing.Store(false)
        case WM_LBUTTONUP:
            if suppressing.Load() {
                suppressing.Store(false)
                return 1
            }
        }
    }
    r, _, _ := pCallNextHookEx.Call(hookHandle, uintptr(nCode), wParam, lParam)
    return r
}

func getTick() uint64 { r, _, _ := pGetTickCount64.Call(); return uint64(r) }

func setDebounce(ms int64) {
    debounceMs.Store(ms)
    lastAccepted.Store(0)
    syncUI()
}

func syncUI() {
    if check != 0 { setChecked(check, enabled.Load()) }
    if combo != 0 { setComboSelection(combo, debounceMs.Load()) }
    updateTrayTip()
    invalidate(hwnd)
}

func updateTrayTip() {
    state := "OFF"
    if enabled.Load() { state = "ON" }
    addTrayIcon(fmt.Sprintf("Mouse Click Debouncer • %s • %d ms", state, debounceMs.Load()), NIM_MODIFY)
}

func addTrayIcon(tip string, action uintptr) {
    var data notifyIconData
    data.cbSize = uint32(unsafe.Sizeof(data)); data.hWnd = hwnd; data.uID = 1
    data.uFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP; data.uCallbackMessage = WM_TRAYICON
    icon, _, _ := pLoadIcon.Call(0, uintptr(IDI_APPLICATION)); data.hIcon = icon
    copy(data.szTip[:], syscall.StringToUTF16(tip))
    pShellNotifyIcon.Call(action, uintptr(unsafe.Pointer(&data)))
}

func removeTrayIcon() {
    var data notifyIconData
    data.cbSize = uint32(unsafe.Sizeof(data)); data.hWnd = hwnd; data.uID = 1
    pShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&data)))
}

func showMenu() {
    menu, _, _ := pCreatePopupMenu.Call(); if menu == 0 { return }
    state := uintptr(MF_UNCHECKED); if enabled.Load() { state = MF_CHECKED }
    appendMenu(menu, MF_STRING|state, ID_ENABLE, "Enable click filtering")
    appendMenu(menu, MF_SEPARATOR, 0, "")
    appendMenu(menu, MF_STRING, ID_50, "50 ms")
    appendMenu(menu, MF_STRING, ID_100, "100 ms")
    appendMenu(menu, MF_STRING, ID_150, "150 ms (recommended)")
    appendMenu(menu, MF_STRING, ID_200, "200 ms")
    appendMenu(menu, MF_STRING, ID_300, "300 ms")
    appendMenu(menu, MF_STRING, ID_500, "500 ms")
    appendMenu(menu, MF_SEPARATOR, 0, "")
    appendMenu(menu, MF_STRING, ID_SHOW, "Open settings")
    appendMenu(menu, MF_STRING, ID_ABOUT, "About")
    appendMenu(menu, MF_STRING, ID_EXIT, "Exit")
    var pt point; pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
    pSetForegroundWindow.Call(hwnd)
    cmd, _, _ := pTrackPopupMenu.Call(menu, TPM_LEFTALIGN|TPM_BOTTOMALIGN|TPM_RETURNCMD|TPM_NONOTIFY, uintptr(pt.x), uintptr(pt.y), 0, hwnd, 0)
    pDestroyMenu.Call(menu)
    if cmd != 0 { pPostMessage.Call(hwnd, WM_COMMAND, cmd, 0) }
}

func appendMenu(menu uintptr, flags uintptr, id int, text string) {
    if flags == MF_SEPARATOR { pAppendMenu.Call(menu, MF_SEPARATOR, 0, 0); return }
    pAppendMenu.Call(menu, flags, uintptr(id), utf16(text))
}

func showAbout() {
    title := "Mouse Click Debouncer"
    text := "Mouse Click Debouncer\n\nA lightweight left-click filter for mice with accidental double-clicks.\n\nCoded by Miah Muhammad Zahid\nCopyright © 2026 Miah Muhammad Zahid. All rights reserved."
    pMessageBox.Call(hwnd, utf16(text), utf16(title), MB_OK|MB_ICONINFORMATION)
}

func showWindow() {
    pShowWindow.Call(hwnd, SW_RESTORE); pShowWindow.Call(hwnd, SW_SHOW); pUpdateWindow.Call(hwnd)
}
func hideWindow() { pShowWindow.Call(hwnd, SW_HIDE) }
func invalidate(h uintptr) { if h != 0 { user32.NewProc("InvalidateRect").Call(h, 0, 1) } }

func isChecked(h uintptr) bool { r, _, _ := pSendMessage.Call(h, 0x00F0, 0, 0); return r != 0 }
func setChecked(h uintptr, v bool) { val := uintptr(0); if v { val = 1 }; pSendMessage.Call(h, 0x00F1, val, 0) }
func comboSelection() int64 {
    idx, _, _ := pSendMessage.Call(combo, 0x0147, 0, 0) // CB_GETCURSEL
    vals := []int64{50,100,150,200,300,500}; if idx < uintptr(len(vals)) { return vals[idx] }; return 150
}
func setComboSelection(h uintptr, ms int64) {
    vals := []int64{50,100,150,200,300,500}; for i,v := range vals { if v == ms { pSendMessage.Call(h, 0x014E, uintptr(i), 0); return } }
}

func makeControl(class, text string, style, x,y,w,h,id int, parent uintptr) uintptr {
    r,_,_ := pCreateWindowEx.Call(0, utf16(class), utf16(text), uintptr(style)|WS_CHILD|WS_VISIBLE, uintptr(x),uintptr(y),uintptr(w),uintptr(h),parent,uintptr(id),instance,0)
    sendFont(r, mainFont); return r
}

func createMainWindow() uintptr {
    proc := syscall.NewCallback(wndProc)
    wc := wndClass{lpfnWndProc: proc, hInstance: instance, hIcon: 0, lpszClassName: className}
    pRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
    title := "Mouse Click Debouncer"
    h,_,_ := pCreateWindowEx.Call(WS_EX_TOOLWINDOW, utf16("MouseClickDebouncerMainWindow"), utf16(title), WS_OVERLAPPEDWINDOW|WS_VISIBLE, CW_USEDEFAULT,CW_USEDEFAULT,560,450,0,0,instance,0)
    if h == 0 { return 0 }
    return h
}

func showDebounceMenu() {
    menu, _, _ := pCreatePopupMenu.Call(); if menu == 0 { return }
    appendMenu(menu, MF_STRING, ID_50, "50 ms")
    appendMenu(menu, MF_STRING, ID_100, "100 ms")
    appendMenu(menu, MF_STRING, ID_150, "150 ms (recommended)")
    appendMenu(menu, MF_STRING, ID_200, "200 ms")
    appendMenu(menu, MF_STRING, ID_300, "300 ms")
    appendMenu(menu, MF_STRING, ID_500, "500 ms")
    var pt point; pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
    cmd, _, _ := pTrackPopupMenu.Call(menu, TPM_LEFTALIGN|TPM_BOTTOMALIGN|TPM_RETURNCMD|TPM_NONOTIFY, uintptr(pt.x), uintptr(pt.y), 0, hwnd, 0)
    pDestroyMenu.Call(menu)
    if cmd != 0 { pPostMessage.Call(hwnd, WM_COMMAND, cmd, 0) }
}

func cleanup() {
    if hookHandle != 0 { pUnhookWindowsHookEx.Call(hookHandle); hookHandle = 0 }
    removeTrayIcon()
    if hwnd != 0 { pDestroyWindow.Call(hwnd); hwnd = 0 }
    if mainFont != 0 { pDeleteObject.Call(mainFont) }
    if titleFont != 0 { pDeleteObject.Call(titleFont) }
    if appMutex != 0 { pCloseHandle.Call(appMutex); appMutex = 0 }
}

func main() {
    runtime.LockOSThread()

    // Prevent multiple copies from installing multiple global mouse hooks.
    appMutex, _, _ = pCreateMutex.Call(0, 1, utf16("MiahMuhammadZahid.MouseClickDebouncer.Singleton"))
    errCode, _, _ := pGetLastError.Call()
    if errCode == 183 {
        if appMutex != 0 { pCloseHandle.Call(appMutex) }
        return
    }

    enabled.Store(true); debounceMs.Store(150)
    instance,_,_ = pGetModuleHandle.Call(0)
    mainFont = createFont(-16, 400)
    titleFont = createFont(-22, 600)
    hwnd = createMainWindow()
    if hwnd == 0 { return }
    addTrayIcon("Mouse Click Debouncer • ON • 150 ms", NIM_ADD)
    cb := syscall.NewCallback(mouseProc)
    hookHandle,_,_ = pSetWindowsHookEx.Call(WH_MOUSE_LL, cb, instance, 0)
    if hookHandle == 0 { showAbout(); cleanup(); return }
    var m msg
    for {
        r,_,_ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)),0,0,0)
        if int32(r) <= 0 { break }
        pTranslateMessage.Call(uintptr(unsafe.Pointer(&m))); pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
    }
    cleanup()
    time.Sleep(20 * time.Millisecond)
}
