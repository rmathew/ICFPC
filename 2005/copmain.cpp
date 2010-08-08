#include <cstring>

#include "cop.hpp"

int main (int argc, char* argv[])
{
  cop* a_cop;

  const char* name = "codermal";

  if (argc > 1)
  {
    if (::strcmp (argv[1], "-noop") == 0)
      a_cop = new cop (name);
    else if (::strcmp (argv[1], "-random") == 0)
      a_cop = new random_cop (name);
    else
      a_cop = new random_cop (name);
  }
  else
    a_cop = new random_cop (name);

  int status = a_cop->start ();

  return status;
}
