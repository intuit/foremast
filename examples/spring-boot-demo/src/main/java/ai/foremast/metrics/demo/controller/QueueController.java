package ai.foremast.metrics.demo.controller;

import ai.foremast.metrics.demo.queue.SimpleQueue;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.RequestParam;

import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
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
    public void error5xx(HttpServletResponse response) throws IOException {
        response.sendError(501, "Internal error");
    }


    @GetMapping("/load")
    public String load(@RequestParam("latency") float latency, @RequestParam("errorRate") float errorRate, HttpServletResponse response) throws IOException {
        float error = errorRate * 100;
        if (error > 0.1) {
            int r = random.nextInt(1000);
            if (r < error * 10) {
                response.sendError(501, "Internal error");
                return "Error";
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
        return "OK";
    }
}
