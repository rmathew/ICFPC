// ICFP 2007
// 20 Jul 2007
//
// build.c: RNA -> PPM
//
// The Great Indian Rope Trick (formerly Kalianpur Bakaits)
// - Ranjit Mathew, rmathew@gmail.com
// - Manoj Plakal, plakal@gmail.com

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/mman.h>

static unsigned long long rna_com_black = 0x0A43494949504950ULL;
static unsigned long long rna_com_red = 0x0A50494949504950ULL;
static unsigned long long rna_com_green = 0x0A43434949504950ULL;
static unsigned long long rna_com_yellow = 0x0A46434949504950ULL;
static unsigned long long rna_com_blue = 0x0A50434949504950ULL;
static unsigned long long rna_com_magenta = 0x0A43464949504950ULL;
static unsigned long long rna_com_cyan = 0x0A46464949504950ULL;
static unsigned long long rna_com_white = 0x0A43504949504950ULL;
static unsigned long long rna_com_transp = 0x0A46504949504950ULL;
static unsigned long long rna_com_opaq = 0x0A50504949504950ULL;
static unsigned long long rna_com_clear = 0x0A50434950494950ULL;
static unsigned long long rna_com_move = 0x0A50494949494950ULL;
static unsigned long long rna_com_turncc = 0x0A50434343434350ULL;
static unsigned long long rna_com_turnc = 0x0A50464646464650ULL;
static unsigned long long rna_com_mark = 0x0A50464649434350ULL;
static unsigned long long rna_com_line = 0x0A50434349464650ULL;
static unsigned long long rna_com_fill = 0x0A50494950494950ULL;
static unsigned long long rna_com_push = 0x0A50464650434350ULL;
static unsigned long long rna_com_compose = 0x0A50434350464650ULL;
static unsigned long long rna_com_clip = 0x0A46434349464650ULL;

static unsigned char bitmaps[10][600][600][4];
static int bm_top = 0;

static int bucket_r = 0, bucket_g = 0, bucket_b = 0, bucket_cols = 0;
static int bucket_opaqs = 0, bucket_alphas = 0;

typedef enum { NORTH = 0, EAST = 1, SOUTH = 2, WEST = 3 } Dir;
static Dir dir = EAST;
static int dx = 1, dy = 0;

static int pos_x = 0, pos_y = 0;
static int mark_x = 0, mark_y = 0;

static inline void black() {
  bucket_cols++;
}


static inline void red() {
  bucket_r++;
  bucket_cols++;
}


static inline void green() {
  bucket_g++;
  bucket_cols++;
}


static inline void yellow() {
  bucket_r++;
  bucket_g++;
  bucket_cols++;
}


static inline void blue() {
  bucket_b++;
  bucket_cols++;
}


static inline void magenta() {
  bucket_r++;
  bucket_b++;
  bucket_cols++;
}


static inline void cyan() {
  bucket_g++;
  bucket_b++;
  bucket_cols++;
}


static inline void white() {
  bucket_r++;
  bucket_g++;
  bucket_b++;
  bucket_cols++;
}


static inline void transp() {
  bucket_alphas++;
}


static inline void opaq() {
  bucket_opaqs++;
  bucket_alphas++;
}


static inline void clear() {
  bucket_r = bucket_g = bucket_b = bucket_cols = bucket_opaqs =  bucket_alphas = 0;
}

// Use a macro for his?
// could use a cache if bucket remains unchanged across calls
// or use current_pixel() once and then use set_pixel_with for
// repeated calls in line/fill etc.
static inline void current_pixel(unsigned char* r, unsigned char* g,
                          unsigned char* b, unsigned char* a) {
  if (bucket_cols == 0) { *r = *g = *b = 0; }
  else {
    *r = (bucket_r * 255) / bucket_cols;
    *g = (bucket_g * 255) / bucket_cols;
    *b = (bucket_b * 255) / bucket_cols;
  }

  if (bucket_alphas == 0) { *a = 255; }
  else { *a = (bucket_opaqs * 255) / bucket_alphas; }

  *r = (*r * *a) / 255;
  *g = (*g * *a) / 255;
  *b = (*b * *a) / 255;
}

static inline void get_bitmap_pixel(int bm, int x, int y,
                             unsigned char* r, unsigned char* g,
                             unsigned char* b, unsigned char*a ) {
  *r = bitmaps[bm][y][x][0];
  *g = bitmaps[bm][y][x][1];
  *b = bitmaps[bm][y][x][2];
  *a = bitmaps[bm][y][x][3];
}

static inline void get_pixel(int x, int y,
                      unsigned char* r, unsigned char* g,
                      unsigned char* b, unsigned char*a ) {
  get_bitmap_pixel(bm_top, x, y, r, g, b, a);
}

static inline void set_bitmap_pixel_with(int bm, int x, int y,
                                  unsigned char r, unsigned char g,
                                  unsigned char b, unsigned char a) {
  bitmaps[bm][y][x][0] = r;
  bitmaps[bm][y][x][1] = g;
  bitmaps[bm][y][x][2] = b;
  bitmaps[bm][y][x][3] = a;
}

static inline void set_pixel_with(int x, int y,
                           unsigned char r, unsigned char g,
                           unsigned char b, unsigned char a) {
  set_bitmap_pixel_with(bm_top, x, y, r, g, b, a);
}

static inline void set_pixel(int x, int y) {
  unsigned char r, g, b, a;
  current_pixel(&r, &g, &b, &a);
  set_pixel_with(x, y, r, g, b, a);
}

static inline void move() {
  // Look ma, no branches!
  pos_x = (pos_x + 600 + dx) % 600;
  pos_y = (pos_y + 600 + dy) % 600;
}

static inline void set_deltas() {
  // Compute these only when changing direction.
  // Look ma, no branches!
  dx = (dir % 2) * (2 - dir);
  dy = ((dir + 1) % 2) * (dir - 1);
}

static inline void turncc() {
  // Look ma, no branches!
  dir = (dir + 4 - 1) % 4;
  set_deltas();
}


static inline void turnc() {
  // Look ma, no branches!
  dir = (dir + 1) % 4;
  set_deltas();
}


static inline void mark() {
  mark_x = pos_x;
  mark_y = pos_y;
}

#define ABS(x) ((x) < 0 ? -(x): (x))
#define MAX(x,y) ((x) < (y) ? (y) : (x))

static inline void line() {
  int deltax = mark_x - pos_x;
  int deltay = mark_y - pos_y;
  int d = MAX(ABS(deltax), ABS(deltay));
  int c = (deltax * deltay <= 0) ? 1 : 0;
  int x = pos_x * d + ((d - c) / 2);
  int y = pos_y * d + ((d - c) / 2);
  unsigned char r, g, b, a;
  int i;

  current_pixel(&r, &g, &b, &a);
  for (i = 0; i < d; i++) {
    set_pixel_with(x/d, y/d, r, g, b, a);
    x += deltax;
    y += deltay;
  }  
  set_pixel_with(mark_x, mark_y, r, g, b, a);
}


static inline void fill() {
  unsigned char old_r, old_g, old_b, old_a;
  unsigned char new_r, new_g, new_b, new_a;

  get_pixel(pos_x, pos_y, &old_r, &old_g, &old_b, &old_a);
  current_pixel(&new_r, &new_g, &new_b, &new_a);
  if ((old_r != new_r) || (old_g != new_g) || (old_b != new_b) || (old_a != new_a)) {
    // starting with a simple non-recursive version of 4-way recursive
    // flood-fill using an explicit stack/queue. could replace this with
    // a faster scanline fill algorithm from foley/van dam.
    // http:/en.wikipedia.org/wiki/Flood_fill
#define MAX_QUEUE_SIZE (1 * 1024 * 1024)
#define PUSH(x, y)\
  if ((q_h - q_t) < MAX_QUEUE_SIZE) {\
     x_q[(q_h++) % MAX_QUEUE_SIZE] = (x);\
     y_q[(q_h++) % MAX_QUEUE_SIZE] = (y);\
  }\
  else { fprintf(stderr, "queue overflow in fill! skipping.\n"); return; }
#define POP(x, y)\
  if (q_t < q_h) {\
     x = x_q[(q_t++) % MAX_QUEUE_SIZE];\
     y = y_q[(q_t++) % MAX_QUEUE_SIZE];\
  }\
  else { fprintf(stderr, "queue underflow in fill! skipping.\n"); return; }

    static int x_q[MAX_QUEUE_SIZE], y_q[MAX_QUEUE_SIZE];
    unsigned long q_h = 0, q_t = 0;  // assuming total # pushes <= ULONG_MAX
    int x, y;
    unsigned char r, g, b, a;

    PUSH(pos_x, pos_y);
    while (q_h > q_t) {
      POP(x, y);
      get_pixel(x, y, &r, &g, &b, &a);
      if ((r == old_r) && (g == old_g) && (b == old_b) && (a == old_a)) {
        set_pixel_with(x, y, new_r, new_g, new_b, new_a);
        if (x > 0)   { PUSH(x-1, y);   }
        if (x < 599) { PUSH(x+1, y);   }
        if (y > 0)   { PUSH(x,   y-1); }
        if (y < 599) { PUSH(x,   y+1); }
      }
    }
  }
}


static inline void push() {
  if (bm_top < 9) {
    bm_top++;
    memset(&bitmaps[bm_top][0][0][0], 0, 600 * 600 * 4);
  } 
}


static inline void compose() {
  if (bm_top > 0) {
    int x, y;
    for (x = 0; x < 600; x++) {
      for (y = 0; y < 600; y++) {
        unsigned char r0, g0, b0, a0;      
        unsigned char r1, g1, b1, a1;      
        get_bitmap_pixel(bm_top, x, y, &r0, &g0, &b0, &a0);
        get_bitmap_pixel(bm_top - 1, x, y, &r1, &g1, &b1, &a1);
        set_bitmap_pixel_with(bm_top - 1, x, y,
                              r0 + ((r1 * (255 - a0)) / 255),
                              g0 + ((g1 * (255 - a0)) / 255),
                              b0 + ((b1 * (255 - a0)) / 255),
                              a0 + ((a1 * (255 - a0)) / 255)
                              );
      }
    }
    bm_top--;
  }
}


static inline void clip() {
  if (bm_top > 0) {
    int x, y;
    for (x = 0; x < 600; x++) {
      for (y = 0; y < 600; y++) {
        unsigned char r0, g0, b0, a0;      
        unsigned char r1, g1, b1, a1;      
        get_bitmap_pixel(bm_top, x, y, &r0, &g0, &b0, &a0);
        get_bitmap_pixel(bm_top - 1, x, y, &r1, &g1, &b1, &a1);
        set_bitmap_pixel_with(bm_top - 1, x, y,
                              ((r1 * a0) / 255),
                              ((g1 * a0) / 255),
                              ((b1 * a0) / 255),
                              ((a1 * a0) / 255)
                              );
      }
    }
    bm_top--;
  }
}

static void dump_pbm() {
  int i, j;
  printf("P3\n600 600\n255\n");
  for (i = 0; i < 600; i++) {
    for (j = 0; j < 600; j++) {
      printf("%u %u %u\n",
             (unsigned int) bitmaps[bm_top][i][j][0], 
             (unsigned int) bitmaps[bm_top][i][j][1], 
             (unsigned int) bitmaps[bm_top][i][j][2]);
    }
  }
}

static void render(char* rna, int rna_size) {
  int total_rna = 0;
  int skipped_rna = 0;
  char* rna_p = rna;
  while (rna_p <= rna + rna_size - 8) {
    unsigned long long rna_com = *(unsigned long long *)rna_p;       
    if (rna_com == rna_com_black) { black(); }
    else if (rna_com == rna_com_red) { red(); }
    else if (rna_com == rna_com_green) { green(); }
    else if (rna_com == rna_com_yellow) { yellow(); }
    else if (rna_com == rna_com_blue) { blue(); }
    else if (rna_com == rna_com_magenta) { magenta(); }
    else if (rna_com == rna_com_cyan) { cyan(); }
    else if (rna_com == rna_com_white) { white(); }
    else if (rna_com == rna_com_transp) { transp(); }
    else if (rna_com == rna_com_opaq) { opaq(); }
    else if (rna_com == rna_com_clear) { clear(); }
    else if (rna_com == rna_com_move) { move(); }
    else if (rna_com == rna_com_turncc) { turncc(); }
    else if (rna_com == rna_com_turnc) { turnc(); }
    else if (rna_com == rna_com_mark) { mark(); }
    else if (rna_com == rna_com_line) { line(); }
    else if (rna_com == rna_com_fill) { fill(); }
    else if (rna_com == rna_com_push) { push(); }
    else if (rna_com == rna_com_compose) { compose(); }
    else if (rna_com == rna_com_clip) { clip(); }
    else {
#ifdef DEBUG
     fprintf(stderr, "skipped %llX\n", rna_com);
#endif
     skipped_rna++;
    }
  
    rna_p += 8;
    total_rna++;
    if ((total_rna % 10000) == 0) {
      fprintf(stderr, "munged %d rnas, skipped %d\n", total_rna, skipped_rna);
      fflush(stderr);
    }
  }

  fprintf(stderr, "total rna: %d, skipped %d\n", total_rna, skipped_rna);
  dump_pbm();
}

int main(int argc, char* argv[]) {
  int fd = open(argv[1], O_RDONLY);
  if (fd < 0) { fprintf(stderr, "pass RNA file as one and only argument!\n"); abort(); }
  int rna_size = (int) lseek(fd, 0, SEEK_END);
  char* rna = (char*) mmap(NULL, rna_size, PROT_READ, MAP_PRIVATE, fd, 0);
  if (rna == NULL) { fprintf(stderr, "could not mmap() RNA file!\n"); abort(); }
  render(rna, rna_size);
  munmap(rna, rna_size);
  return 0;
}
