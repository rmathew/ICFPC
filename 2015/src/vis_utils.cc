#include "vis_utils.h"

#include <string.h>

namespace {

constexpr double kSqrt3ByTwo = 0.8660254037844386;
constexpr double kTwoBySqrt3 = 1.1547005383792515;
constexpr double kTwiceSqrt3 = 3.4641016151377544;

void DrawHorizLine(SDL_Surface* screen, Uint16 x, Uint16 y, Uint16 length,
  Uint32 color) {
    if (length == 0) {
        return;
    }

    const Uint8 bytes_per_pixel = screen->format->BytesPerPixel;
    Uint8* buff = static_cast<Uint8*>(screen->pixels) + y * screen->pitch +
      x * bytes_per_pixel;

    switch (bytes_per_pixel) {
      case 1: {
        memset(buff, color, length);
        break;
      }

      case 2: {
        Uint16* buff_2 = reinterpret_cast<Uint16*>(buff);
        for (int i = 0; i < length; ++i) {
            *buff_2 = color;
            ++buff_2;
        }
        break;
      }

      case 3: {
        Uint8 p0, p1, p2;
        if (SDL_BYTEORDER == SDL_BIG_ENDIAN) {
            p0 = (color >> 16) & 0xFF;
            p1 = (color >> 8) & 0xFF;
            p2 = color & 0xFF;
        } else {
            p0 = color & 0xFF;
            p1 = (color >> 8) & 0xFF;
            p2 = (color >> 16) & 0xFF;
        }
        for (int i = 0; i < length; ++i) {
            buff[0] = p0;
            buff[1] = p1;
            buff[2] = p2;
            buff += 3;
        }
        break;
      }

      case 4: {
        Uint32* buff_4 = reinterpret_cast<Uint32*>(buff);
        for (int i = 0; i < length; ++i) {
            *buff_4 = color;
            ++buff_4;
        }
        break;
      }
    }
}

}  // namespace

Uint16 GetHexHeight(Uint16 hex_width) {
    return static_cast<Uint16>(static_cast<double>(hex_width) * kTwoBySqrt3);
}

Uint16 GetHexWidth(Uint16 hex_height) {
    return static_cast<Uint16>(static_cast<double>(hex_height) * kSqrt3ByTwo);
}

void DrawHex(SDL_Surface* screen, Uint16 x, Uint16 y, Uint16 hex_width,
  Uint32 color) {
    const Uint16 cap_height = GetHexHeight(hex_width) / 4;
    for (Uint16 i = 0; i < cap_height; ++i) {
        Uint16 line_width =
          static_cast<Uint16>(static_cast<double>(i) * kTwiceSqrt3);
        DrawHorizLine(screen, (x + (hex_width - line_width) / 2), y, line_width,
            color);
        ++y;
    }

    const Uint16 body_height = 2 * cap_height;
    for (Uint16 i = 0; i < body_height; ++i) {
        DrawHorizLine(screen, x, y, hex_width, color);
        ++y;
    }

    for (Uint16 i = 0; i < cap_height; ++i) {
        Uint16 line_width = static_cast<Uint16>(
          static_cast<double>(cap_height - i) * kTwiceSqrt3);
        DrawHorizLine(screen, (x + (hex_width - line_width) / 2), y, line_width,
            color);
        ++y;
    }
}
