#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

static void do_cat(const char *path);
static void do_cat_fd(int, const char*);
static void die(const char *s);

int
main(int argc, char *argv[])
{
  int i;

  if (argc < 2) {
    do_cat_fd(STDIN_FILENO, "stdin");
    exit(0);
  }

  for (i = 1; i < argc; i++) {
    do_cat(argv[i]);
  }

  exit(0);
}

static void
do_cat(const char *path)
{
  int fd;

  fd = open(path, O_RDONLY);
  if (fd < 0) die(path);

  do_cat_fd(fd, path);

  if (close(fd) < 0) die(path);
}

#define BUFFER_SIZE 2048

static void
do_cat_fd(int fd, const char *path)
{
  unsigned char buf[BUFFER_SIZE];
  int n;

  for (;;) {
    n = read(fd, buf, sizeof buf);
    if (n < 0) die(path);
    if (n == 0) break;
    if (write(STDOUT_FILENO, buf, n) < 0) die(path);
  }
}

static void
die(const char *s)
{
  perror(s);
  exit(1);
}
