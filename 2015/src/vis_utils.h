#ifndef VIS_UTILS_H_INCLUDED
#define VIS_UTILS_H_INCLUDED

#include <SDL.h>

Uint16 GetHexHeight(Uint16 hex_width);

Uint16 GetHexWidth(Uint16 hex_height);

Uint16 GetHexCapHeight(Uint16 hex_width);

void DrawHex(SDL_Surface* screen, Uint16 x, Uint16 y, Uint16 hex_width,
  Uint32 color);

#endif /* VIS_UTILS_H_INCLUDED */
