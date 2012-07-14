/*
  The interface to the text output routines for SDL.
*/

#ifndef SDLTXT_H_INCLUDED
#define SDLTXT_H_INCLUDED

extern int sdltxt_init(const SDL_PixelFormat *fmt);

extern int sdltxt_metrics(Uint16 *font_width, Uint16 *font_height);

extern int sdltxt_write(const char *txt, Uint16 max, SDL_Surface *dest,
  Uint16 x, Uint16 y);

extern void sdltxt_quit(void);

#endif /* SDLTXT_H_INCLUDED */
