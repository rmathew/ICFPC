/*
 * A simple controller for the Mars Rover in ICFP contest 2008.
 *
 * Author: Ranjit Mathew <rmathew@gmail.com>
 *
 * Based on the enhanced "Vector Field Histogram" (VFH+) method as described
 * in "VFH+: Reliable Obstacle Avoidance for Fast Mobile Robots" by Iwan
 * Ulrich and Johann Borenstein:
 *
 *   http://www-personal.umich.edu/~johannb/Papers/paper73.pdf
 *
 */
import java.io.File;
import java.io.InputStream;
import java.io.OutputStream;

import java.util.Set;
import java.util.HashSet;
import java.util.List;
import java.util.ArrayList;
import java.util.Arrays;

import java.net.Socket;

import java.awt.Color;
import java.awt.Graphics2D;

import java.awt.image.BufferedImage;

import javax.imageio.ImageIO;


public class Rover
{
  private static final int RUN_STATE_CRATER_FALL = 0;
  private static final int RUN_STATE_MARTIAN_KILL = 1;
  private static final int RUN_STATE_TIMED_OUT = 2;
  private static final int RUN_STATE_RUNNING = 3;
  private static final int RUN_STATE_SUCCESS = 4;

  private static final int ACC_STATE_BRAKING = 0;
  private static final int ACC_STATE_ROLLING = 1;
  private static final int ACC_STATE_ACCELERATING = 2;

  private static final int TURN_STATE_HARD_LEFT = 0;
  private static final int TURN_STATE_LEFT = 1;
  private static final int TURN_STATE_STRAIGHT = 2;
  private static final int TURN_STATE_RIGHT = 3;
  private static final int TURN_STATE_HARD_RIGHT = 4;

  private static final int STEER_LEFT = 0;
  private static final int STEER_FORWARD = 1;
  private static final int STEER_RIGHT = 2;
  private static final int STEER_BACKWARD = 3;

  private static final double ROVER_RAD = 0.5;
  private static final double MARTIAN_RAD = 0.4;
  private static final double ROVER_SAFE_DIST = 0.2;

  private static final double SAMPLE_INTERVAL = 0.1;

  private static final int WINDOW_SIZE = 129;
  private static final double CELL_SIZE = 0.1;

  /* NOTE: Must be a small integral factor of 360. */
  private static final int SECTOR_SPAN = 3;
  private static final int NUM_SECTORS = 360 / SECTOR_SPAN;
  private static final int WIDE_OPENING = 16;

  private static final int PRESENT_MASK = 0x00000001;


  private double[] polarHist = new double[NUM_SECTORS];
  private boolean[] biPolarHist = new boolean[NUM_SECTORS];
  private boolean[] maskPolarHist = new boolean[NUM_SECTORS];

  private double wSpan;
  private double aForMag, bForMag;

  private double lowMagTh;
  private double highMagTh;

  private double[] preCompMags;
  private double[] preCompAngles;
  private int[] preCompSectors;

  private Set<String> seenStaticObjs = new HashSet<String>( 1000);

  private BufferedImage worldMap;
  private Graphics2D wmGfx;

  private List<String> visMarXs = new ArrayList<String>( 10);
  private List<String> visMarYs = new ArrayList<String>( 10);

  private int[] candOpBegs = new int[NUM_SECTORS];
  private int[] candOpEnds = new int[NUM_SECTORS];

  private Socket sk;

  private InputStream is;
  private OutputStream os;

  private int numRuns = 0;

  private double limX, limY;
  private double minSensor, maxSensor;
  private double maxSpeed, maxTurn, maxHardTurn;
  private double safeSpeed;
  private double curX, curY;
  private double curSpeed, curDir;
  private int curAccState, curTurnState, curRunState;

  private long curMsgTs, expMinMsgTs;

  private char[][] steerDecTable
    = {
                       /* LEFT, FORWARD, RIGHT, BACKWARD */
      /* HARD_LEFT */  {   '-',   'r',     'r',   '-'},
      /* LEFT */       {   '-',   'r',     'r',   'l'},
      /* STRAIGHT */   {   'l',   '-',     'r',   'r'},
      /* RIGHT */      {   'l',   'l',     '-',   'r'},
      /* HARD_RIGHT */ {   'l',   'l',     '-',   '-'},
    };

  private StringBuilder sb = new StringBuilder( );
  private byte[] outMsg = new byte[3];


  private Rover( String host, int port) throws Exception
  {
    sk = new Socket( host, port);
    sk.setTcpNoDelay( true);

    is = sk.getInputStream( );
    os = sk.getOutputStream( );
  }

  private void startRoving( )
    throws Exception
  {
    initialise( );

    getInitialData( );

    while( readMsg( ))
    {
      writeMsg( );
    }

    os.close( );
    is.close( );

    sk.close( );
  }

  private void initialise( ) throws Exception
  {
    initRunState( );
    preComputeData( );
  }

  private void initRunState( ) throws Exception
  {
    curAccState = ACC_STATE_ROLLING;
    curTurnState = TURN_STATE_STRAIGHT;
    curRunState = RUN_STATE_RUNNING;

    Arrays.fill( biPolarHist, false);

    curMsgTs = -1L;
    expMinMsgTs = 0L;
  }

  private void getInitialData( ) throws Exception
  {
    String[] initMsg = getMsg( );

    double dX = Double.valueOf( initMsg[1]);
    double dY = Double.valueOf( initMsg[2]);

    limX = dX / 2.0;
    limY = dY / 2.0;

    minSensor = Double.valueOf( initMsg[4]);
    maxSensor = Double.valueOf( initMsg[5]);

    maxSpeed = Double.valueOf( initMsg[6]);

    maxTurn = Double.valueOf( initMsg[7]);
    maxHardTurn = Double.valueOf( initMsg[8]);

    int numCellsX = (int )(dX / CELL_SIZE) + 1;
    int numCellsY = (int )(dY / CELL_SIZE) + 1;

    worldMap
      = new BufferedImage(
          numCellsX, numCellsY, BufferedImage.TYPE_BYTE_BINARY);
   
    wmGfx = worldMap.createGraphics( );
    wmGfx.setPaintMode( );
    wmGfx.setColor( Color.WHITE);
    wmGfx.fillRect( 0, 0, numCellsX, numCellsY);

    wmGfx.setColor( Color.BLACK);
    wmGfx.drawRect( 0, 0, numCellsX - 1, numCellsY - 1);

    System.out.println( "Map Size: " + dX + " x " + dY);
    System.out.println( "Time-Limit: " + initMsg[3] + " ms");
  }

  private void preComputeData( ) throws Exception
  {
    wSpan = WINDOW_SIZE * CELL_SIZE;

    safeSpeed = (wSpan / 2.0) / SAMPLE_INTERVAL;

    /* In the lookout window, "maxDist = wSpan / sqrt(2)". */
    double maxDistSqrd = wSpan * wSpan / 2.0;

    /* Select "a" and "b" such that "a - b*maxDistSqrd = 1". */
    aForMag = 1.0 + wSpan * wSpan;
    bForMag = 2.0;

    /* Low threshold. */
    lowMagTh = 1.0;

    /* High threshold. */
    highMagTh = aForMag - bForMag * maxDistSqrd / 4.0;
    

    int cellsInWindow = WINDOW_SIZE * WINDOW_SIZE;

    preCompMags = new double[cellsInWindow];
    preCompAngles = new double[cellsInWindow];
    preCompSectors = new int[cellsInWindow];

    double centreX = CELL_SIZE * ((double )WINDOW_SIZE / 2.0  + 0.5);
    double centreY = centreX;

    for( int i = 0; i < cellsInWindow; i++)
    {
      int row = i / WINDOW_SIZE;
      int col = (i % WINDOW_SIZE);

      double x = ((double )col + 0.5) * CELL_SIZE;
      double y = ((double )row + 0.5) * CELL_SIZE;

      double gapX = (x - centreX);
      double gapY = (y - centreY);

      double distSqrd = gapX * gapX + gapY * gapY;

      double mag = aForMag - bForMag * distSqrd;

      double beta = angleForPoint( gapX, gapY);

      int secNum = (int )Math.floor( beta / (double )SECTOR_SPAN);

      preCompMags[i] = mag;
      preCompAngles[i] = beta;
      preCompSectors[i] = secNum;
    }
  }

  private boolean readMsg( ) throws Exception
  {
    boolean retVal = true;

    String[] msg = getMsg( );

    if( (msg != null) && (msg.length > 0) && (msg[0] != null)
      && (msg[0].length( ) > 0))
    {
      char msgType = msg[0].charAt( 0);

      switch( msgType)
      {
      case 'T':
        /* Telemetry data from the sensors. */
        curMsgTs = Long.valueOf( msg[1]);

        /* Only bother if it's not stale telemetry data. */
        if( curMsgTs >= expMinMsgTs)
        {
          parseVehicleCtl( msg[2]);

          curX = Double.valueOf( msg[3]);
          curY = Double.valueOf( msg[4]);

          curDir = normAngle( Double.valueOf( msg[5]));

          curSpeed = Double.valueOf( msg[6]);

          visMarXs.clear( );
          visMarYs.clear( );

          if( msg.length > 7)
          {
            /* There are visible boulders, craters, martians or home base. */
            int idx = 7;
            while( idx < msg.length)
            {
              if( msg[idx].equals( "b") || msg[idx].equals( "c"))
              {
                /* Boulder or Crater */
                maybeAddStaticObj(
                  msg[idx], msg[idx+1], msg[idx+2], msg[idx+3]);

                idx += 4;
              }
              else if( msg[idx].equals( "m"))
              {
                /* Martian */
                visMarXs.add( msg[idx+1]);
                visMarYs.add( msg[idx+2]);

                idx += 5;
              }
              else
              {
                /* Home base */
                idx += 4;
                continue;
              }
            }
          }
        }
        else
        {
          System.err.println(
            "WARNING: Stale data (" + curMsgTs + " ms v/s minimum "
            + expMinMsgTs + " ms)");
        }
        break;

      case 'B':
        /* Hit a boulder. */
        break;

      case 'C':
        /* Fell into a crater. */
        curRunState = RUN_STATE_CRATER_FALL;
        break;

      case 'K':
        /* Killed by a martian. */
        curRunState = RUN_STATE_MARTIAN_KILL;
        break;

      case 'S':
        /* Reached home base. */
        curRunState = RUN_STATE_SUCCESS;
        break;

      case 'E':
        /* Run ended. */
        numRuns++;
        printRunInfo( msg[1], msg[2]);
        initRunState( );
        break;
      }
    }
    else
      retVal = false;

    return retVal;
  }

  private void parseVehicleCtl( String s) throws Exception
  {
    char accState = s.charAt( 0);
    switch( accState)
    {
    case 'a':
      curAccState = ACC_STATE_ACCELERATING;
      break;

    case 'b':
      curAccState = ACC_STATE_BRAKING;
      break;

    case '-':
      curAccState = ACC_STATE_ROLLING;
      break;
    }

    char turnState = s.charAt( 1);
    switch( turnState)
    {
    case 'L':
      curTurnState = TURN_STATE_HARD_LEFT;
      break;

    case 'l':
      curTurnState = TURN_STATE_LEFT;
      break;

    case '-':
      curTurnState = TURN_STATE_STRAIGHT;
      break;

    case 'r':
      curTurnState = TURN_STATE_RIGHT;
      break;

    case 'R':
      curTurnState = TURN_STATE_HARD_RIGHT;
      break;
    }
  }

  private void printRunInfo( String time, String score)
  {
    System.out.print( "Run " + numRuns + ": " + score + " (");
    String runState = "<unknown>";
    switch( curRunState)
    {
      case RUN_STATE_CRATER_FALL:
        runState = "fell into a crater";
        break;

      case RUN_STATE_MARTIAN_KILL:
        runState = "killed by a martian";
        break;

      case RUN_STATE_RUNNING:
      case RUN_STATE_TIMED_OUT:
        runState = "timed out";
        break;

      case RUN_STATE_SUCCESS:
        runState = "reached home";
        break;
    }
    System.out.println( runState + "), " + time + " ms.");
  }

  private void writeMsg( ) throws Exception
  {
    /* Only bother if it's not stale telemetry data. */
    if( curMsgTs >= expMinMsgTs)
    {
      long t0 = System.currentTimeMillis( );

      calcPolarHist( );

      double steerDir = chooseSteerDir( );

      sendSteerCommand( steerDir);

      long t = System.currentTimeMillis( ) - t0;

      expMinMsgTs = curMsgTs + t;
    }
  }

  private void sendSteerCommand( double steerDir) throws Exception
  {
    int n = 0;

    int steer = STEER_FORWARD;

    double dif = absAngDif( curDir, steerDir);

    if( dif > 90.0)
      steer = STEER_BACKWARD;
    else if( dif > 5.0)
    {
      if( isToTheLeft( steerDir, curDir))
        steer = STEER_LEFT;
      else
        steer = STEER_RIGHT;
    }
 
    if( (steer == STEER_BACKWARD) || (curSpeed > safeSpeed))
      outMsg[n++] = 'b';
    else
      outMsg[n++] = 'a';

    char steerChar = steerDecTable[curTurnState][steer];
    if( steerChar != '-')
      outMsg[n++] = (byte )steerChar;

    outMsg[n++] = ';';

    os.write( outMsg, 0, n);
  }

  private double chooseSteerDir( )
  {
    double tgtDir = angleForPoint( -curX, -curY);
    double chosenDir = tgtDir;

    double minWt = Double.MAX_VALUE;
    for( int i = 0; i < NUM_SECTORS; i++)
    {
      if( maskPolarHist[i] == false)
      {
        double dir = i * SECTOR_SPAN;
        double wt = computeDirCost( dir, tgtDir);
        if( wt < minWt)
        {
          minWt = wt;
          chosenDir = dir;
        }
      }
    }

    if( minWt == Double.MAX_VALUE)
      chosenDir = normAngle( curDir + 180.0);

    return chosenDir;
  }

  private int findCandOps( )
  {
    /* Find openings in the masked polar histogram. */
    int numCandOps = 0;

    int beg = -1;
    int end = -1;
    int initEnd = -1;

    for( int i = 0; i < NUM_SECTORS; i++)
    {
      if( maskPolarHist[i] == false)
      {
        if( beg == -1)
          beg = i;

        end = i;
      }
      else
      {
        if( beg == 0)
          initEnd = end;
        else
        {
          candOpBegs[numCandOps] = beg;
          candOpEnds[numCandOps] = end;
          numCandOps += 1;
        }

        beg = -1;
        end = -1;
      }
    }


    if( maskPolarHist[NUM_SECTORS - 1] == false)
    {
      candOpBegs[numCandOps] = beg;

      if( maskPolarHist[0] == false)
      {
        candOpEnds[numCandOps] = initEnd;
      }
      else
      {
        candOpEnds[numCandOps] = end;
      }

      numCandOps += 1;
    }
    else if( initEnd != -1)
    {
      candOpBegs[numCandOps] = 0;
      candOpEnds[numCandOps] = initEnd;
      numCandOps += 1;
    }

    return numCandOps;
  }

  private double evalCandOps( int numCandOps)
  {
    double tgtDir = angleForPoint( -curX, -curY);
    double chosenDir = tgtDir;

    double minWt = Double.MAX_VALUE;
    double wt;

    for( int i = 0; i < numCandOps; i++)
    {
      int beg = candOpBegs[i];
      int end = candOpEnds[i];

      int span = end - beg + 1;

      if( span < 0)
        span += NUM_SECTORS;

      if( span > WIDE_OPENING)
      {
        double dirBeg = normAngle( (beg + (WIDE_OPENING / 2)) * SECTOR_SPAN);
        wt = computeDirCost( dirBeg, tgtDir);
        if( wt < minWt)
        {
          minWt = wt;
          chosenDir = dirBeg;
        }

        double dirEnd = normAngle( (end - (WIDE_OPENING / 2)) * SECTOR_SPAN);
        wt = computeDirCost( dirEnd, tgtDir);
        if( wt < minWt)
        {
          minWt = wt;
          chosenDir = dirEnd;
        }

        if( isWithin( tgtDir, dirBeg, dirEnd))
        {
          wt = computeDirCost( tgtDir, tgtDir);
          if( wt < minWt)
          {
            minWt = wt;
            chosenDir = tgtDir;
          }
        }
      }
      else
      {
        int avg = (beg + end) / 2;

        double dir = avg * SECTOR_SPAN;
        if( beg > end)
          dir = normAngle( dir + 180.0);

        wt = computeDirCost( dir, tgtDir);
        if( wt < minWt)
        {
          minWt = wt;
          chosenDir = dir;
        }
      }
    }

    return chosenDir;
  }

  double computeDirCost( double dir, double tgtDir)
  {
    double turnAngle = 0.0;

    switch( curTurnState)
    {
    case TURN_STATE_HARD_LEFT:
      turnAngle = 30.0;
      break;
    case TURN_STATE_LEFT:
      turnAngle = 15.0;
      break;
    case TURN_STATE_RIGHT:
      turnAngle = -15.0;
      break;
    case TURN_STATE_HARD_RIGHT:
      turnAngle = -30.0;
      break;
    }

    turnAngle = normAngle( turnAngle);

    double cost = 7.0 * absAngDif( dir, tgtDir);
    cost += 3.0 * absAngDif( dir, curDir);
    cost += 1.0 * absAngDif( dir, turnAngle);

    return cost;
  }

  double absAngDif( double a1, double a2)
  {
    double retVal = Math.abs( a1 - a2);
    retVal = Math.min( retVal, Math.abs( a1 - a2 - 360.0));
    retVal = Math.min( retVal, Math.abs( a1 - a2 + 360.0));
    return retVal;  
  }

  private void dumpWorldMap( ) throws Exception
  {
    ImageIO.write( worldMap, "png", new File( "worldmap.png"));
  }

  private void calcPolarHist( ) throws Exception
  {
    /* Initialise polar histogram to zero. */
    Arrays.fill( polarHist, 0.0);

    /* Determine the base of the histogram window on the map. */
    double origX0 = curX - (wSpan / 2.0);
    double origY0 = curY - (wSpan / 2.0);

    /* Find the bounds of the histogram window on the map. */
    double wX0 = origX0;
    double wY0 = origY0;

    double wX1 = wX0 + wSpan;
    double wY1 = wY0 + wSpan;

    /* Clip the histogram window if it spills over the map. */
    if( wX0 < -limX)
      wX0 = -limX;

    if( wY0 < -limY)
      wY0 = -limY;

    if( wX1 > limX)
      wX1 = limX;

    if( wY1 > limY)
      wY1 = limY;

    /* Determine the width and height of the histogram window on the map. */
    double wSpanX = wX1 - wX0;
    double wSpanY = wY1 - wY0;

    /* Determine the base of the histogram window in the histogram. */
    int mBegX = (int )Math.floor( (wX0 + limX) / CELL_SIZE);
    int mBegY = (int )Math.floor( (wY0 + limY) / CELL_SIZE);

    /* Determine the span of the histogram window in the histogram. */
    int mLimX = mBegX + (int )Math.floor( wSpanX / CELL_SIZE);
    int mLimY = mBegY + (int )Math.floor( wSpanY / CELL_SIZE);

    /* If the histogram window is no longer a square due to clipping,
     * determine the padding needed on each of the axes. */
    int wPadX = (int )Math.floor( (wX0 - origX0) / CELL_SIZE);
    int wPadY = (int )Math.floor( (wY0 - origY0) / CELL_SIZE);


    /* Determine the left and right minimum turn radius and the
     * centres of these circles. */
    double minTurnRad = (curSpeed * 180.0) / (maxTurn * Math.PI);
    double minTurnRadSqrd = minTurnRad * minTurnRad;

    double rSinTheta = minTurnRad * Math.sin( curDir * Math.PI / 180.0);
    double rCosTheta = minTurnRad * Math.cos( curDir * Math.PI / 180.0);

    double rX = ((double )WINDOW_SIZE / 2.0 + 0.5) * CELL_SIZE;
    double rY = rX;

    double lCenX = rX - rSinTheta;
    double lCenY = rY + rCosTheta;

    double rCenX = rX + rSinTheta;
    double rCenY = rY - rCosTheta;

    double phiBack = normAngle( curDir + 180.0);

    double phiLeft = normAngle( phiBack - 5.0);
    double phiRight = normAngle( phiBack + 5.0);

    for( int j = mBegY; j < mLimY; j++)
    {
      for( int i = mBegX; i < mLimX; i++)
      {
        boolean occupied = ((worldMap.getRGB( i, j) & PRESENT_MASK) == 0);
        if( occupied)
        {
          int xIdx = wPadX + i - mBegX;
          int yIdx = wPadY + j - mBegY;

          int flatIdx = xIdx + (yIdx * WINDOW_SIZE);

          int secNum = preCompSectors[flatIdx];

          polarHist[secNum] += preCompMags[flatIdx];

          double objX = ((double )xIdx + 0.5) * CELL_SIZE;
          double objY = ((double )yIdx + 0.5) * CELL_SIZE;

          double beta = preCompAngles[flatIdx];

          if( isToTheLeft( beta, curDir) && isToTheRight( beta, phiLeft))
          {
            double dLX = (objX - lCenX);
            double dLY = (objY - lCenY);

            double dLSqrd = dLX * dLX + dLY * dLY;

            if( dLSqrd < minTurnRadSqrd)
              phiLeft = beta;
          }
          else if( isToTheRight( beta, curDir) && isToTheLeft( beta, phiRight))
          {
            double dRX = (objX - rCenX);
            double dRY = (objY - rCenY);

            double dRSqrd = dRX * dRX + dRY * dRY;

            if( dRSqrd < minTurnRadSqrd)
              phiRight = beta;
          }
        }
      }
    }


    /* Crudely account for visible martians, if any. */
    if( !visMarXs.isEmpty( ))
    {
      double effMarDia = 2.0 * (MARTIAN_RAD + ROVER_RAD + ROVER_SAFE_DIST);
      double effMarArea = Math.PI * effMarDia * effMarDia / 4.0;
      double effMarCells = effMarArea / (CELL_SIZE * CELL_SIZE);

      int numMartians = visMarXs.size( );
      for( int i = 0; i < numMartians; i++)
      {
        double marX = Double.valueOf( visMarXs.get( i));
        double marY = Double.valueOf( visMarYs.get( i));

        double distX = marX - curX;
        double distY = marY - curY;

        double distSqrd = distX * distX + distY * distY;

        double marAngle = angleForPoint( distX, distY);

        int secNum = (int )Math.floor( marAngle / (double )SECTOR_SPAN);

        double mag = effMarCells * (aForMag - bForMag * distSqrd);

        polarHist[secNum] += mag;
      }
    }


    /* Update the binary polar histogram. */
    for( int i = 0; i < NUM_SECTORS; i++)
    {
      double mag = polarHist[i];

      if( mag > highMagTh)
        biPolarHist[i] = true;
      else if( mag < lowMagTh)
        biPolarHist[i] = false;
    }

    /* Update the masked polar histogram. */
    for( int i = 0; i < NUM_SECTORS; i++)
    {
      maskPolarHist[i] = biPolarHist[i];
      if( maskPolarHist[i] == false)
      {
        double dir = i * SECTOR_SPAN;

        if( isToTheLeft( dir, curDir))
          maskPolarHist[i] = isToTheLeft( dir, phiLeft);
        else
          maskPolarHist[i] = isToTheRight( dir, phiRight);
      }
    }
  }

  private boolean isWithin( double a, double a1, double a2)
  {
    return isToTheLeft( a, a1) && isToTheRight( a, a2);
  }

  private boolean isToTheLeft( double a1, double a2)
  {
    boolean retVal = false;

    double a2rev = normAngle( a2 + 180.0);

    if( a2 < a2rev)
      retVal = (a1 > a2) && (a1 <= a2rev);
    else
      retVal = (a1 > a2) || (a1 <= a2rev);

    return retVal;
  }

  private boolean isToTheRight( double a1, double a2)
  {
    boolean retVal = false;

    double a2rev = normAngle( a2 + 180.0);

    if( a2 < a2rev)
      retVal = (a1 <= a2) || (a1 > a2rev);
    else
      retVal = (a1 <= a2) && (a1 > a2rev);

    return retVal;
  }

  private String[] getMsg( )
  {
    sb.setLength( 0);

    try
    {
      int c;
      while( ((c = is.read( )) != -1) && (c != ';'))
      {
        sb.append( (char )c);
      }
    }
    catch( Exception e)
    {
      /* IGNORED. */
    }

    if( sb.length( ) > 0)
      return sb.toString( ).split( " ");
    else
      return null;
  }

  private void maybeAddStaticObj( String kind, String x, String y, String r)
  {
    String key = x + "," + y;

    if( !seenStaticObjs.contains( key))
    {
      double xVal = Double.valueOf( x);
      double yVal = Double.valueOf( y);
      double rVal = Double.valueOf( r);

      rVal += ROVER_SAFE_DIST;
      rVal += ROVER_RAD;

      int mPosX = (int )Math.floor( (xVal + limX - rVal) / CELL_SIZE);
      int mPosY = (int )Math.floor( (yVal + limY - rVal) / CELL_SIZE);

      int nCells = (int )Math.floor( 2.0 * rVal / CELL_SIZE);

      wmGfx.fillOval( mPosX, mPosY, nCells, nCells);

      seenStaticObjs.add( key);
    }
  }

  private double angleForPoint( double x, double y)
  {
    return normAngle( Math.atan2( y, x) * 180.0 / Math.PI);
  }

  private double normAngle( double a)
  {
    double retVal = a;

    if( retVal >= 360.0)
      retVal -= 360.0;
    else if( retVal < 0.0)
      retVal += 360.0;

    return retVal;
  }


  public static void main( String[] args) throws Exception
  {
    String host = "localhost";
    int port = 17676;

    if( args.length > 0)
    {
      host = args[0];
    }

    if( args.length > 1)
    {
      port = Integer.valueOf( args[1]);
    }

    Rover r = new Rover( host, port);

    r.startRoving( );
  }
}
