package ai.foremast.metrics.demo.error;

import javax.annotation.PostConstruct;
import java.io.InputStream;
import java.net.URL;

public class ErrorGenerator implements Runnable {


    private Thread thread;

    private int frequency = 5;

    private String url = null;

    private int sleepTime = 800;

    public ErrorGenerator(int frequency, String errorType) {
        this.frequency = frequency;
        if ("4xx".equalsIgnoreCase(errorType)) {
            url = "http://localhost:8080/not_existed?t=";
        }
        else if ("5xx".equalsIgnoreCase(errorType)) {
            url = "http://localhost:8080/error5xx?t=";
        }

        sleepTime = 1000 / frequency;
    }

    @PostConstruct
    public void init() {
        if (url != null) {
            thread = new Thread(this, "Error_Generator");
            thread.start();
        }
    }

    @Override
    public void run() {
        if (url != null) {
            while (true) {
                try {
                    Thread.sleep(sleepTime);

                    InputStream input = null;
                    try {
                        URL u = new URL(url);
                        input = u.openStream();
                    } catch (Exception ex) {
                        ex.printStackTrace();
                    } finally {
                        if (input != null) {
                            input.close();
                        }
                    }
                } catch (Exception ex) {
                    ex.printStackTrace();
                }
            }
        }
    }

    public int getFrequency() {
        return frequency;
    }
}
