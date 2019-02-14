package ai.foremast.metrics.demo.controller;

import ai.foremast.metrics.demo.queue.SimpleQueue;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.RequestParam;

import java.util.Random;

@RestController
public class QueueController {


    @Autowired
    SimpleQueue simpleQueue;

    private Random random = new Random();

    /**
     * A way to add more items in the queue
     */
    @GetMapping("/pushSome")
    public String pushSome() {
        simpleQueue.push(1 + random.nextInt(500));
        return "Done:" + simpleQueue.getSize();
    }


    @GetMapping("/error5xx")
    public void error5xx() {
        throw new IllegalStateException("5xx");
    }


    @GetMapping("/load")
    public void load(@RequestParam("latency") float latency, @RequestParam("errorRate") float errorRate) {

        float error = errorRate * 100;
        if (error > 0.1) {
            int random = random.nextInt(1000);
            if (error * 10 < random) {
                throw new IllegalStateException("5xx");
            }
        }

        if (latency <= 0) {
            latency = 10;
        }
        try {
            Thread.sleep((int)latency);
        }
        catch(Exception ex) {
        }
    }
}
