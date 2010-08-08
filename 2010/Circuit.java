public class Circuit {
  public static final int INVALID_GATE = -1;
  public static final int EXTERNAL_GATE = Integer.MAX_VALUE;

  private static final boolean DEBUG = false;

  /* NOTE: Derived decision-tables (and input-stream) for gate-type 0 by
   * reverse-engineering after looking at the output in error messages for
   * various simple circuits. Look at "gout.txt" for the trail.
   *
   * The gate-logic seems to be:
   *   LO = LI - RI
   *   RO = LI*RI + 2
   * where LO and RO are left and right output and are stored modulo 3.
   *
   * There doesn't (mercifully) seem to be any other gate-type.
   */

  /* NOTE: The nice symmetric distribution of values in this table
   * is critical to using the left output of this gate-type to
   * reverse-engineer the input-stream and to determine a circuit that
   * takes a given input to a given output.
   */
  private static int[][] leftOutput = {
    { 0, 2, 1},
    { 1, 0, 2},
    { 2, 1, 0},
  };

  private static int[][] rightOutput = {
    { 2, 2, 2},
    { 2, 0, 1},
    { 2, 1, 0},
  };

  private boolean circuitWired = false;

  private int numGates = 0;
  private int[] inpConns = null;
  private int[] outConns = null;
  private int[] gateInputs = null;

  public int getNumGates() {
    return numGates;
  }

  public int getInputConn(int gateNum, boolean right) {
    int delta = right ? 1 : 0;
    return inpConns[2*gateNum + delta];
  }

  public int getOutputConn(int gateNum, boolean right) {
    int delta = right ? 1 : 0;
    return outConns[2*gateNum + delta];
  }

  public int getExtInputPos() {
    for(int i = 0; i < 2*numGates; i++) {
      if(inpConns[i] == EXTERNAL_GATE)
        return i;
    }
    return -1;
  }

  public int getExtOutputPos() {
    for(int i = 0; i < 2*numGates; i++) {
      if(outConns[i] == EXTERNAL_GATE)
        return i;
    }
    return -1;
  }

  public int[] run(int[] extInput) {
    if(!circuitWired) {
      System.err.println("ERROR: Circuit is not yet wired up");
      System.exit(1);
    }

    if(DEBUG) {
      System.out.println("INFO: Input stream");
      System.out.print("\t");
      for(int i = 0; i < extInput.length; i++)
        System.out.print(extInput[i]);
      System.out.println("");
    }

    int[] retVal = new int[extInput.length];

    for(int inpNum = 0; inpNum < extInput.length; inpNum++) {
      if(DEBUG)
        System.out.print("@" + inpNum + ": " + extInput[inpNum] + " + [");

      retVal[inpNum] = execCycle(extInput[inpNum]);

      if(DEBUG)
        System.out.println("]");
    }

    return retVal;
  }

  private int execCycle(int extIn) {
    int retVal = -1;

    int numSlots = inpConns.length;
    for(int leftIdx = 0; leftIdx < numSlots; leftIdx += 2) {
      int rightIdx = leftIdx + 1;

      if(inpConns[leftIdx] == EXTERNAL_GATE)
        gateInputs[leftIdx] = extIn;

      if(inpConns[rightIdx] == EXTERNAL_GATE)
        gateInputs[rightIdx] = extIn;

      int leftIn = gateInputs[leftIdx];
      int rightIn = gateInputs[rightIdx];

      if(DEBUG)
        System.out.print("(" + leftIn + "," + rightIn + ")");

      /* NOTE: Replacing table-lookup by the equivalent arithmetic on the
       * inputs shaves off some running time (95K -> 70K ns).
       */
      /*
      int leftOut = leftOutput[leftIn][rightIn];
      int rightOut = rightOutput[leftIn][rightIn];
      */

      int leftOut = (leftIn - rightIn + 3) % 3;
      int rightOut = (leftIn * rightIn + 2) % 3;

      int leftOutTgt = outConns[leftIdx];
      if(leftOutTgt == EXTERNAL_GATE) {
        retVal = leftOut;
      } else {
        gateInputs[leftOutTgt] = leftOut;
      }

      int rightOutTgt = outConns[rightIdx];
      if(rightOutTgt == EXTERNAL_GATE) {
        retVal = rightOut;
      } else {
        gateInputs[rightOutTgt] = rightOut;
      }
    }
    
    return retVal;
  }

  public void parseFactory(String facStr) throws Exception {
    String[] secs = facStr.split(":");
    if(secs.length != 3) {
      System.err.println(
          "ERROR: Unexpected number (" + secs.length + ") of sections");
      System.exit(1);
    }

    String[] gatesDes = secs[1].split(",");
    numGates = gatesDes.length;
    inpConns = new int[2*numGates];
    outConns = new int[2*numGates];
    gateInputs = new int[2*numGates];
    for(int i = 0; i < 2*numGates; i++) {
      inpConns[i] = outConns[i] = INVALID_GATE;
      gateInputs[i] = 0;
    }

    int[] tmpGates = new int[2];
    for(int i = 0; i < numGates; i++) {
      if(DEBUG)
        System.out.println(
          "INFO: Parsing gate desc #" + i + " \"" + gatesDes[i] + "\"");

      String[] comps = gatesDes[i].split("#");

      if(comps.length != 2) {
        System.err.println(
          "ERROR: Gate description #" + i + " \"" + gatesDes[i]
          + "\" is invalid");
        System.exit(1);
      }

      int leftIdx = 2*i;
      int rightIdx = leftIdx + 1;

      String inSpec = comps[0];
      parseConnSpec(inSpec, tmpGates);
      inpConns[leftIdx] = tmpGates[0];
      inpConns[rightIdx] = tmpGates[1];

      String outSpec = comps[1];
      parseConnSpec(outSpec, tmpGates);
      outConns[leftIdx] = tmpGates[0];
      outConns[rightIdx] = tmpGates[1];
    }

    /* Sanity check. */
    boolean snafu = false;
    for(int i = 0; i < 2*numGates; i++) {
      if(inpConns[i] == INVALID_GATE) {
        System.err.println("ERROR: INVALID_GATE at input pos " + posToStr(i));
        snafu = true;
      }

      if(outConns[i] == INVALID_GATE) {
        System.err.println("ERROR: INVALID_GATE at output pos " + posToStr(i));
        snafu = true;
      }
    }
    if(snafu)
      System.exit(1);

    if(DEBUG) {
      System.out.println("INFO: Connections");
      for(int i = 0; i < numGates; i++) {
        int idx = 2*i;
        String inStr1 = posToStr(inpConns[idx]);
        String inStr2 = posToStr(inpConns[idx + 1]);
        String outStr1 = posToStr(outConns[idx]);
        String outStr2 = posToStr(outConns[idx + 1]);

        System.out.println(
          "\t" + i + ": " + inStr1 + " " + inStr2 + " -> " + outStr1 + " "
          + outStr2);
      }
    }

    circuitWired = true;
  }

  private void parseConnSpec(String spec, int[] out) throws Exception {
    if(DEBUG)
      System.out.println("INFO: \tParsing conn spec \"" + spec + "\"");

    char[] specChars = spec.toCharArray();

    int pos = 0;
    int specComp = 0;
    while(specComp < 2) {
      if(specChars[pos] == 'X') {
        out[specComp] = EXTERNAL_GATE;
      } else if(Character.isDigit(specChars[pos])) {
        int num = 0;
        do {
          num = num*10 + Character.digit(specChars[pos], 10);
          pos++;
        } while(Character.isDigit(specChars[pos]));

        int delta = 0;
        switch(specChars[pos]) {
          case 'L':
            delta = 0;
            break;
          case 'R':
            delta = 1;
            break;
          default:
            System.err.println(
              "ERROR: Unexpected spec character '" + specChars[pos]
              + "' (expecting 'L' or 'R')");
            System.exit(1);
            break;
        }
        out[specComp] = 2*num + delta;
      } else {
        System.err.println(
          "ERROR: Unexpected spec character '" + specChars[pos] + "'");
        System.exit(1);
      }

      if(DEBUG)
        System.out.println(
          "INFO: \t\tConnection to " + posToStr(out[specComp]) + " in "
          + ((specComp == 0) ? "L" : "R"));

      pos++;
      specComp++;
    }
  }

  private String posToStr(int pos) {
    if(pos == EXTERNAL_GATE)
      return "X";
    else
      return Integer.toString(pos/2) + (((pos % 2) == 0) ? "L" : "R");
  }
}
