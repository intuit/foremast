/**
 * Licensed to the Foremast under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package ai.foremast.metrics.demo.load;

import javax.annotation.PostConstruct;
import java.io.*;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;

/**
 * Load generator
 *
 * @author Sheldon Shao
 * @version 1.0
 */
public class LoadGenerator implements Runnable {


    private Thread[] threads;

    private Runner[] runners;

    private String url = "http://localhost:8080/load";

    private Thread mainThread;


    public LoadGenerator() {

    }

    public LoadGenerator(String url) {
        this.url = url;
    }

    private static class DataPoint {
        int traffic;
        float latency;
        float error;

        public DataPoint(String line) {
            parse(line);
        }

        public void parse(String line) {
            String[] vals = line.split(",");
            if (vals.length == 3) {
                traffic = Integer.parseInt(vals[0]);
                latency = Float.parseFloat(vals[1]);
                error = Float.parseFloat(vals[2]);
            }
        }
    }
    private List<DataPoint> datelist = new ArrayList<>();

    private class Runner implements Runnable {

        private boolean enabled = false;

        private DataPoint point;

        /**
         * When an object implementing interface <code>Runnable</code> is used
         * to create a thread, starting the thread causes the object's
         * <code>run</code> method to be called in that separately executing
         * thread.
         * <p>
         * The general contract of the method <code>run</code> is that it may
         * take any action whatsoever.
         *
         * @see Thread#run()
         */
        @Override
        public void run() {
            while(true) {
                if (enabled) {
                    doRun();
                }
                else {
                    synchronized (this) {
                        try {
                            this.wait(50);
                        } catch (InterruptedException e) {
                        }
                    }
                }
            }
        }

        protected void doRun() {
            try {
                DataPoint point = this.point;
                int expectedTime = 1000;
                URL u = new URL(url + "?latency=" + point.latency + "&errorRate=" + point.error);
                long start = System.currentTimeMillis();
                try (InputStream input = u.openStream()) {
                } catch (Exception ex) {
                }
                long time = System.currentTimeMillis() - start;
                if (time < expectedTime) {
                    Thread.sleep(expectedTime-time);
                }
            } catch (Exception ex) {
            }
        }


        public boolean isEnabled() {
            return enabled;
        }

        public void setEnabled(boolean enabled) {
            this.enabled = enabled;
            if (enabled) {
                synchronized (this) {
                    this.notifyAll();
                }
            }
        }

        public DataPoint getPoint() {
            return point;
        }

        public void setPoint(DataPoint point) {
            this.point = point;
        }
    }

    @PostConstruct
    public void init() {
        if (url != null) {
            int maxThreads = 10;
            try (InputStream input = Thread.currentThread().getContextClassLoader().getResourceAsStream("load.txt")) {
                BufferedReader br = new BufferedReader(new InputStreamReader(input));

                String strLine = br.readLine(); //Header
                while ((strLine = br.readLine()) != null) {
                    DataPoint point = new DataPoint(strLine);
                    this.datelist.add(point);

                    maxThreads = Math.max(maxThreads, point.traffic);
                }
            }
            catch (IOException ex) {
                ex.printStackTrace();
            }

            threads = new Thread[maxThreads];
            runners = new Runner[maxThreads];
            for(int i = 0; i < maxThreads; i ++) {
                runners[i] = new Runner();
                runners[i].setEnabled(false);
                threads[i] = new Thread(runners[i]);
                threads[i].start();
            }

            mainThread = new Thread(this);
            mainThread.start();
        }
    }

    @Override
    public void run() {
        while(true) {
            for(DataPoint point: datelist) {
                int currentThreads = point.traffic;

                for(int i = 0; i < runners.length; i ++) {
                    if (i < currentThreads) {
                        runners[i].setPoint(point);
                        runners[i].setEnabled(true);
                    }
                    else {
                        runners[i].setEnabled(false);
                    }
                }

                //Switch to next simulation
                try {
                    Thread.sleep(30000);
                }
                catch(Exception ex) {
                }
            }
        }
    }

    public static void main(String[] args) {
        if (args.length == 0) {
            System.out.println("Usage: java LoadGenerator <base_url>");
            System.exit(0);
        }
        LoadGenerator loadGenerator = new LoadGenerator(args[0]);
        loadGenerator.init();
    }
}
