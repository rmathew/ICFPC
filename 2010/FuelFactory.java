public class FuelFactory {
  /* NOTE: Derived input-stream by first figuring out gate-logic and
   * then using the simple circuit in "circuits/L2L_1.txt" to figure
   * out the input trits based on the output.
   */
  private static final String SRV_INPUT_PREFIX = "01202101210201202";

  /**
   * Sample input-stream for the sample circuit ("circuits/tasksam.txt")
   * from the task description.
   *
   * IMPORTANT: This gives the key-prefix according to the task description
   * that then needs to be prefixed to any ternary stream that encodes a
   * solution.
   */
  private static final String TASK_INPUT_PREFIX = "02222220210110011";

  private static final boolean DEBUG = false;

  public static void main(String[] args) throws Exception {
    if(args.length != 2) {
      System.err.println(
        "ERROR: Need input-stream and factory description as arguments");
      System.exit(1);
    }

    String inpFile = args[0];
    String factoryFile = args[1];

    String facStr = Utils.fileToStr(factoryFile);
    Circuit c = new Circuit();
    c.parseFactory(facStr);

    String inpStr = Utils.fileToStr(inpFile);
    int[] inp = Utils.strToTrits(inpStr);

    long t0 = System.nanoTime();
    int[] out = c.run(inp);
    long t1 = System.nanoTime();

    if(DEBUG)
      System.out.println("INFO: Run took " + (t1 - t0) + " ns");

    System.out.println("Output: " + Utils.tritsToStr(out));
  }
}
