#ifndef VIS_UTILS_H_INCLUDED
#define VIS_UTILS_H_INCLUDED

#include <SDL.h>

Uint16 GetHexHeight(Uint16 hex_width);

Uint16 GetHexWidth(Uint16 hex_height);

void DrawHex(SDL_Surface* screen, Uint16 x, Uint16 y, Uint16 hex_width,
  Uint32 color);

void DrawPivot(SDL_Surface* screen, Uint16 x, Uint16 y, Uint16 pivot_width,
  Uint32 color);

#endif /* VIS_UTILS_H_INCLUDED */
