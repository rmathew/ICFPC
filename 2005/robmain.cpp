#include <cstring>

#include "robber.hpp"

int main (int argc, char* argv[])
{
  robber* a_robber;

  const char* name = "codermal";

  if (argc > 1)
  {
    if (::strcmp (argv[1], "-noop") == 0)
      a_robber = new robber (name);
    else if (::strcmp (argv[1], "-justbank") == 0)
      a_robber = new just_bank (name);
    else if (::strcmp (argv[1], "-random") == 0)
      a_robber = new random_robber (name);
    else
      a_robber = new just_bank (name);
  }
  else
    a_robber = new just_bank (name);

  int status = a_robber->start ();

  return status;
}
