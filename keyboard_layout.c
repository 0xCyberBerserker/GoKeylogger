#include <windows.h>

__declspec(dllexport) int GetKeyboardLayoutID() {
    // Obtiene el layout actual de teclado
    HKL layout = GetKeyboardLayout(0);

    // Convierte el puntero de layout a un valor entero para extraer el ID
    return (int)((uintptr_t)layout & 0xFFFF);
}
