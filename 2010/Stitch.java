import java.util.ArrayList;

public class Stitch {
  private static final boolean DEBUG = true;

  public static void main(String[] args) throws Exception {
    if(args.length != 3) {
      System.err.println("ERROR: Invalid number of arguments (need 3)");
      System.err.println(
        "\t <inp-stream-file> <tgt-stream-file> <soln-out-file>");
      System.exit(1);
    }

    String inStr = Utils.fileToStr(args[0]);
    String tgtStr = Utils.fileToStr(args[1]);

    if(inStr.length() != tgtStr.length()) {
      System.err.println(
        "ERROR: Input stream not as long as target stream (" + inStr.length() + " v/s " + tgtStr.length() + ")");
      System.exit(1);
    }

    Lilo soln = solve(inStr, tgtStr);
  }

  private static Lilo solve(String inStr, String tgtStr) throws Exception {
    int[] inp = Utils.strToTrits(inStr);
    int[] tgt = Utils.strToTrits(tgtStr);

    Lilo base = new Lilo("0", null, inp, true);
    base.setLeftTgt(null, tgt, true);

    ArrayList<Lilo> nodes = new ArrayList<Lilo>(32);
    nodes.add(base);

    do {
      if(DEBUG)
        System.out.println( "INFO: " + nodes.size() + " nodes");

      /* Solve it... (Ha!) */

    } while (false);

    return base;
  }
}
