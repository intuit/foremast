package ai.foremast.metrics.demo.controller;

import ai.foremast.metrics.demo.queue.SimpleQueue;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

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
}
