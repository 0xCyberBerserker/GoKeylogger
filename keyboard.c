#include <windows.h>
#include <stdio.h>

HHOOK keyboardHook;

// Función de callback para el hook de teclado
LRESULT CALLBACK LowLevelKeyboardProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode == HC_ACTION) {
        KBDLLHOOKSTRUCT *kbdStruct = (KBDLLHOOKSTRUCT *)lParam;
        if (wParam == WM_KEYDOWN) {
            if (kbdStruct->vkCode == VK_RMENU) { // AltGr
                printf("Tecla presionada: AltGr\n");
            }
            if (kbdStruct->vkCode == VK_LMENU) { // AltLeft
                printf("Tecla presionada: AltLeft\n");
            }
            // Aquí puedes manejar otras teclas
        }
    }
    return CallNextHookEx(keyboardHook, nCode, wParam, lParam);
}

// Función para instalar el hook de teclado
void InstallKeyboardHook() {
    keyboardHook = SetWindowsHookEx(WH_KEYBOARD_LL, LowLevelKeyboardProc, NULL, 0);
}
