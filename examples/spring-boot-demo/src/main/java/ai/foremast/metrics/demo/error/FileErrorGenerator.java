package ai.foremast.metrics.demo.error;

import javax.annotation.PostConstruct;
import java.io.InputStream;
import java.io.FileInputStream;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.IOException;
import java.net.URL;
import java.util.ArrayList;

public class FileErrorGenerator implements Runnable {


    private Thread thread;

    private int frequency = 5;

    private String url = null;

    private String filename = null;

    private ArrayList<Double> valuelist = new ArrayList<Double>();
    private ArrayList<String> datelist = new ArrayList<String>();

    public FileErrorGenerator(int frequency, String errorType, String filename) {
        this.frequency = frequency;
        this.filename = filename;
        if ("4xx".equalsIgnoreCase(errorType)) {
            url = "http://localhost:8080/not_existed?t=";
        }
        else if ("5xx".equalsIgnoreCase(errorType)) {
            url = "http://localhost:8080/error5xx?t=";
        }
    }

    @PostConstruct
    public void init() {
        if (url != null) {
            thread = new Thread(this, "Error_Generator");
            thread.start();

            try (FileInputStream fstream = new FileInputStream(filename)){
              BufferedReader br = new BufferedReader(new InputStreamReader(fstream));

              String strLine;
              while ((strLine = br.readLine()) != null) {
                String[] vals = strLine.split(",", 2);
                double value = Double.parseDouble(vals[1]);
                String date = vals[0];
                this.valuelist.add(value);
                this.datelist.add(date);
              }
            }
            catch (IOException ex) {
                ex.printStackTrace();
            }
        }
    }

    @Override
    public void run() {
        if (url != null) {
            while (true) {
              for (int i = 0; i < this.valuelist.size(); i++) {
                try {
                    double v = this.valuelist.get(i);
                    if (v < 0.001) {
                        continue;
                    }
                    long sleepTime = (long)(1000.00/v);
                    if (v > 1) {
                        for(int j = 0; j < v; j ++) {
                            Thread.sleep(sleepTime);
                            URL u = new URL(url);
                            try (InputStream input = u.openStream()) {
                            }
                        }
                    }
                    else {
                        Thread.sleep(sleepTime);
                        URL u = new URL(url);
                        try (InputStream input = u.openStream()) {
                        }
                    }
                } catch (Exception ex) {
                }

              }

            }
        }
    }

    public int getFrequency() {
        return frequency;
    }
}
