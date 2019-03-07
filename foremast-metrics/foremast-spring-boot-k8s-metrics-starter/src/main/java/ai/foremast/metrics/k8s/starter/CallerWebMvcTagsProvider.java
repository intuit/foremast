package ai.foremast.metrics.k8s.starter;

import io.micrometer.core.instrument.Tag;
import io.micrometer.core.instrument.Tags;
import org.springframework.boot.actuate.metrics.web.servlet.WebMvcTags;
import org.springframework.boot.actuate.metrics.web.servlet.WebMvcTagsProvider;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;


public class CallerWebMvcTagsProvider implements WebMvcTagsProvider {

    private static final Tag CALLER_UNKNOWN = Tag.of("caller", "UNKNOWN");

    private String headerName;

    public CallerWebMvcTagsProvider(String headerName) {
        this.headerName = headerName;
    }

    public Tag caller(HttpServletRequest request) {
        if (request != null) {
            String header = request.getHeader(headerName);
            return header != null ? Tag.of("caller", header.trim()) : CALLER_UNKNOWN;
        }
        return CALLER_UNKNOWN;
    }

    @Override
    public Iterable<Tag> getTags(HttpServletRequest request, HttpServletResponse response, Object handler, Throwable exception) {
        return Tags.of(new Tag[]{WebMvcTags.method(request), WebMvcTags.uri(request, response), WebMvcTags.exception(exception), WebMvcTags.status(response), caller(request)});
    }

    @Override
    public Iterable<Tag> getLongRequestTags(HttpServletRequest request, Object handler) {
        return Tags.of(new Tag[]{WebMvcTags.method(request), WebMvcTags.uri(request, (HttpServletResponse)null), caller(request)});
    }
}
