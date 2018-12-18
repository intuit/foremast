package ai.foremast.metrics.demo.queue;

import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.Metrics;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;

import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedDeque;

/**
 * A simple queue
 * It just a sample who has metric
 */
@Component
public class SimpleQueue implements Runnable {


    private Queue<Object> queue = new ConcurrentLinkedDeque<>();

    private Thread thread;

    private boolean running = false;

    @PostConstruct
    public void init() {
        //Create a gauge to show the queue size
        Gauge.builder("k8s-metrics-demo.queue_size", queue, queue -> queue.size())
                .description("An sample of application metric")
                .register(Metrics.globalRegistry);

        running = true;
        thread = new Thread(this);
        thread.start();
    }


    public void push(int count) {
        for(int i = 0; i < count; i ++) {
            queue.offer(new Object());
        }
    }

    public int getSize() {
        return queue.size();
    }

    @Override
    public void run() {
        //Just poll the object out of queue
        while(running) {
            try {
                Thread.sleep(20000);
                int half = queue.size() / 4;
                for(int i = 0; i < half; i ++) {
                    queue.poll();
                }
            }
            catch(Exception ex) {
                //Ignore
            }
        }
    }


    @PreDestroy
    public void destroy() {
        running = false;
    }
}
