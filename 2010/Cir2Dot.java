public class Cir2Dot {
  public static void main(String[] args) throws Exception {
    if(args.length < 1) {
      System.err.println("ERROR: Missing circuit file");
      System.exit(1);
    }

    String facStr = Utils.fileToStr(args[0]);
    Circuit c = new Circuit();
    c.parseFactory(facStr);

    System.out.println("digraph {");

    System.out.println(
      "  ext [shape=Mrecord, fontname=\"monospace\", fontsize=10, style=filled, fillcolor=gray90, label=\"e\"];");

    for(int i = 0; i < c.getNumGates(); i++) {
      System.out.println(
        "  gate" + i + " [shape=record, fontname=\"monospace\", fontsize=10, label=\"{{ <li> li | <ri> ri } | g" + i + " | { <lo> lo | <ro> ro }}\"];");
    }

    int extIn = c.getExtInputPos();
    boolean inLeft = ((extIn % 2) == 0);
    System.out.println(
      "  ext:e -> gate" + extIn/2 + ":" + (inLeft ? "li" : "ri") + ":n;");

    int extOut = c.getExtOutputPos();
    boolean outLeft = ((extOut % 2) == 0);
    System.out.println(
      "  gate" + extOut/2 + ":" + (outLeft ? "lo" : "ro") + ":s -> ext:w;");

    for(int i = 0; i < c.getNumGates(); i++) {
      int leftOut = c.getOutputConn(i, false);
      boolean tgtLeft = ((leftOut % 2) == 0);
      if(leftOut != Circuit.EXTERNAL_GATE)
        System.out.println(
          "  gate" + i + ":lo:s -> gate" + leftOut/2 + ":" + (tgtLeft ? "li" : "ri") + ":n;");

      int rightOut = c.getOutputConn(i, true);
      tgtLeft = ((rightOut % 2) == 0);
      if(rightOut != Circuit.EXTERNAL_GATE)
        System.out.println(
          "  gate" + i + ":ro:s -> gate" + rightOut/2 + ":" + (tgtLeft ? "li" : "ri") + ":n;");
    }
    System.out.println("}");
  }
}
