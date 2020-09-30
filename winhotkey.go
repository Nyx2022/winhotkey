package winhotkey

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

const (
	// ModAlt either ALT key must be held down.
	ModAlt = 0x0001 << iota
	// ModCtrl either CTRL key must be held down.
	ModCtrl
	// ModShift either SHIFT key must be held down.
	ModShift
	// ModWin either WINDOWS key was held down.
	//These keys are labeled with the Windows logo. Keyboard shortcuts that involve the WINDOWS key are reserved for use by the operating system.
	ModWin
	// ModNoRepeat Changes the hotkey behavior so that the keyboard auto-repeat does not yield multiple hotkey notifications.
	ModNoRepeat = 0x4000 // Windows Vista:  This flag is not supported.
)

// Hotkey defines a key configuration
type Hotkey struct {
	ID           int16          // Unique id
	Modifiers    int            // Mask of modifiers
	KeyCode      int            // Key code, e.g. 'A'
	CallbackChan chan<- *Hotkey // The callback channel
	CallbackFunc func(*Hotkey)  // The callback function
}

// MSG Contains message information from a thread's message queue.
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-msg
type MSG struct {
	HWND   uintptr
	UINT   uintptr
	WPARAM int16
	LPARAM int64
	DWORD  int32
	POINT  struct{ X, Y int64 }
}

// String returns a human-friendly display name of the hotkey
// such as "Hotkey[Id: 1, Alt+Ctrl+O]"
func (h *Hotkey) String() string {
	var mod strings.Builder
	if h.Modifiers&ModNoRepeat != 0 {
		mod.WriteString("NoRepeat ")
	}
	if h.Modifiers&ModCtrl != 0 {
		mod.WriteString("Ctrl+")
	}
	if h.Modifiers&ModAlt != 0 {
		mod.WriteString("Alt+")
	}
	if h.Modifiers&ModShift != 0 {
		mod.WriteString("Shift+")
	}
	if h.Modifiers&ModWin != 0 {
		mod.WriteString("Win+")
	}
	return fmt.Sprintf("Hotkey[ID: %d, %s%c]", h.ID, mod.String(), h.KeyCode)
}

var (
	getmsg    ProcInterface
	reghotkey ProcInterface
	user32    *syscall.DLL
	keys      map[int16]*Hotkey
)

// ProcInterface wrapper for syscall.Proc.Call to enable mocking in tests
type ProcInterface interface {
	Call(a ...uintptr) (r1 uintptr, r2 uintptr, lastErr error)
}

func init() {
	runtime.LockOSThread()
	keys = make(map[int16]*Hotkey)
	user32 = syscall.MustLoadDLL("user32")
	// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registerhotkey
	reghotkey = user32.MustFindProc("RegisterHotKey")
	// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getmessagew
	getmsg = user32.MustFindProc("GetMessageW")
}

// Run the Hotkey handler
func Run() error {
	for {
		var msg = &MSG{}
		r1, _, err := getmsg.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
		if err != nil && r1 != 1 {
			return err
		}
		if id := msg.WPARAM; id != 0 {
			h, found := keys[id]
			if found {
				if h.CallbackFunc != nil {
					go h.CallbackFunc(h)
				}
				if h.CallbackChan != nil {
					select {
					case h.CallbackChan <- h:
					default:
					}
				}
			}
		}
	}
}

// RegisterHotkey registers a Hotkey
func RegisterHotkey(h *Hotkey) error {
	for _, hk := range keys {
		if hk.ID == h.ID {
			return fmt.Errorf("ID %d already assigned to %s", h.ID, hk)
		}
	}
	r1, _, err := reghotkey.Call(0, uintptr(h.ID), uintptr(h.Modifiers), uintptr(h.KeyCode))
	if r1 == 0 {
		return fmt.Errorf("Failed to register %s, error: %s", h, err)
	}
	if err != nil && r1 != 1 {
		return err
	}
	keys[h.ID] = h
	return nil
}
