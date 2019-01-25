package ai.foremast.metrics.demo.error;

import javax.annotation.PostConstruct;
import java.io.InputStream;
import java.io.FileInputStream;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.IOException;
import java.net.URL;
import java.lang.Math;
import java.util.ArrayList;


// read whole file into memory at once
// calculate file

public class FileErrorGenerator implements Runnable {


    private Thread thread;

    private int frequency = 5;

    private String url = null;

    private int sleepTime = 800;

    private String filename = null;

    private double lastValue = -100;

    private int lastLine = 0;

    private double threshold = 0;

    private ArrayList<Double> valuelist = new ArrayList<Double>();
    private ArrayList<String> datelist = new ArrayList<String>();

    public FileErrorGenerator(int frequency, String errorType, String filename, String threshold) {
        this.frequency = frequency;
        this.filename = filename;
        if (threshold != null) {
            this.threshold = Double.parseDouble(threshold);
        }
        else {
            this.threshold = 0.5;
        }
        if ("4xx".equalsIgnoreCase(errorType)) {
            url = "http://localhost:8080/not_existed?t=";
        }
        else if ("5xx".equalsIgnoreCase(errorType)) {
            url = "http://localhost:8080/error5xx?t=";
        }

        sleepTime = 1000/frequency;
    }

    @PostConstruct
    public void init() {
        if (url != null) {
            thread = new Thread(this, "Error_Generator");
            thread.start();

            FileInputStream fstream = null;
            try {
              fstream = new FileInputStream(filename);
              BufferedReader br = new BufferedReader(new InputStreamReader(fstream));

              String strLine;
              int curLine = 0;
              while ((strLine = br.readLine()) != null) {
                String[] vals = strLine.split(",", 2);
                double value = Double.parseDouble(vals[1]);
                String date = vals[0];
                this.valuelist.add(value);
                this.datelist.add(date);

              }
              fstream.close();
            } catch (IOException ex) {
                ex.printStackTrace();
            } finally {
                if (fstream != null) {
                  try {
                    fstream.close();
                  } catch (Exception ex) {
                    ex.printStackTrace();
                  }

                }
            }
        }
    }

    @Override
    public void run() {
        if (url != null) {
            while (true) {
              for (int i = 0; i < this.valuelist.size(); i++) {
                InputStream input = null;

                try {
                    double v = this.valuelist.get(i);
                    long sleepTime = v > 0 ? (long)(1000.00/v) : 500;
                    if (sleepTime > 5000) {
                        sleepTime = 5000;
                    }
                  Thread.sleep(sleepTime);
                    URL u = new URL(url);
                    input = u.openStream();
                } catch (Exception ex) {
                    ex.printStackTrace();
                } finally {
                    if (input != null) {
                      try {
                        input.close();
                      } catch (Exception ex) {
                        ex.printStackTrace();
                      }
                    }
                }

              }

            }
        }
    }

    public int getFrequency() {
        return frequency;
    }
}
