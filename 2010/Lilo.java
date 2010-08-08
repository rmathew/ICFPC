public class Lilo {
  private String id = null;

  private Lilo leftSrc = null;
  private Lilo rightSrc = null;

  private Lilo leftTgt = null;
  private Lilo rightTgt = null;

  private int[] leftIn = null;
  private int[] rightIn = null;

  private int[] leftOut = null;
  private int[] rightOut = null;

  private boolean leftOfLeftSrc = true;
  private boolean leftOfRightSrc = true;

  private boolean leftOfLeftTgt = true;
  private boolean leftOfRightTgt = true;


  public Lilo(String theId, Lilo lSrc, int[] lIn, boolean fromLeftOutput) {
    id = theId;

    leftSrc = lSrc;
    leftIn = lIn;
    leftOfLeftSrc = fromLeftOutput;
  }

  public String getId() {
    return id;
  }

  public int[] getLeftIn() {
    return leftIn;
  }

  public int[] getRightIn() {
    return rightIn;
  }

  public int[] getLeftOut() {
    return leftOut;
  }

  public int[] getRightOut() {
    return rightOut;
  }

  public void setLeftTgt(Lilo lTgt, int[] lOut, boolean toLeftInput) {
    leftTgt = lTgt;
    leftOut = lOut;
    leftOfLeftTgt = toLeftInput;

    rightIn = new int[leftIn.length];
    rightOut = new int[leftIn.length];
    for(int i = 0; i < leftIn.length; i++) {
      rightIn[i] = (leftIn[i] - leftOut[i] + 3) % 3;
      rightOut[i] = (leftIn[i] * rightIn[i] + 2) % 3;
    }
  }

  public void setRightSrc(Lilo rSrc, boolean fromLeftOutput) {
    rightSrc = rSrc;
    leftOfRightSrc = fromLeftOutput;
  }

  public void setRightTgt(Lilo rTgt, boolean toLeftInput) {
    rightTgt = rTgt;
    leftOfRightTgt = toLeftInput;
  }

  public String getNodeDesc() {
    StringBuilder sb = new StringBuilder(16);

    if(leftSrc == null)
      sb.append('X');
    else
      sb.append(leftSrc.getId() + (leftOfLeftSrc ? "L" : "R"));

    if(rightSrc == null)
      sb.append('X');
    else
      sb.append(rightSrc.getId() + (leftOfRightSrc ? "L" : "R"));

    sb.append("0#");

    if(leftTgt == null)
      sb.append('X');
    else
      sb.append(leftTgt.getId() + (leftOfLeftTgt ? "L" : "R"));

    if(rightTgt == null)
      sb.append('X');
    else
      sb.append(rightTgt.getId() + (leftOfRightTgt ? "L" : "R"));

    return sb.toString();
  }
}
