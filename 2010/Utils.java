import java.io.Reader;
import java.io.FileReader;

public class Utils {
  public static final int TERNARY_BASE = 3;

  public static String fileToStr(String fileName) throws Exception {
    FileReader fr = new FileReader(fileName);
    StringBuilder sb = new StringBuilder(32);
    int aChar;
    while((aChar = Utils.nextChar(fr)) != -1) {
      char theChar = Character.toUpperCase((char)aChar);
      sb.append(theChar);
    }
    fr.close();

    return sb.toString();
  }

  public static int[] strToTrits(String str) throws IllegalArgumentException {
    if(str == null || str.trim().equals(""))
      throw new IllegalArgumentException("Empty string");

    str = str.trim();
    char[] strChars = str.toCharArray();
    int[] retVal = new int[strChars.length];
    for(int i = 0; i < strChars.length; i++) {
      retVal[i] = Character.digit(strChars[i], TERNARY_BASE);
      if(retVal[i] < 0)
        throw new IllegalArgumentException(
            "Invalid character '" + strChars[i] + "'");
    }
    return retVal;
  }

  public static String tritsToStr(int[] t) throws IllegalArgumentException {
    StringBuilder sb = new StringBuilder(t.length);
    for(int i = 0; i < t.length; i++) {
      char c = Character.forDigit(t[i], 3);
      if(c != '\u0000')
        sb.append(c);
      else
        throw new IllegalArgumentException(
          "Invalid ternary digit " + t[i] + " at position " + i);
    }
    return sb.toString();
  }

  public static int nextChar(Reader r) throws Exception {
    int retVal;
    do {
      retVal = r.read();
    } while((retVal != -1) && Character.isWhitespace(retVal));
    return retVal;
  }
}
