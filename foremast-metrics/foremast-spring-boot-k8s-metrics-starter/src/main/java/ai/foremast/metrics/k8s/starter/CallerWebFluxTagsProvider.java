package ai.foremast.metrics.k8s.starter;

import io.micrometer.core.instrument.Tag;
import io.micrometer.core.instrument.Tags;
import org.springframework.boot.actuate.metrics.web.reactive.server.WebFluxTagsProvider;
import org.springframework.boot.actuate.metrics.web.reactive.server.WebFluxTags;
import org.springframework.web.server.ServerWebExchange;


public class CallerWebFluxTagsProvider implements WebFluxTagsProvider {

    private static final Tag CALLER_UNKNOWN = Tag.of("caller", "UNKNOWN");

    private String headerName;

    public CallerWebFluxTagsProvider(String headerName) {
        this.headerName = headerName;
    }

    public Tag caller(ServerWebExchange exchange) {
        if (exchange != null) {
            String header = exchange.getRequest().getHeaders().getFirst((headerName));
            return header != null ? Tag.of("caller", header.trim()) : CALLER_UNKNOWN;
        }
        return CALLER_UNKNOWN;
    }
    @Override
    public Iterable<Tag> httpRequestTags(ServerWebExchange exchange, Throwable ex) {
        return Tags.of(new Tag[]{WebFluxTags.method(exchange), WebFluxTags.uri(exchange), WebFluxTags.exception(ex), WebFluxTags.status(exchange), caller(exchange)});

    }
}
