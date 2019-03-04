/**
 * Copyright 2017 Pivotal Software, Inc.
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package ai.foremast.micrometer.autoconfigure.export;

import io.micrometer.core.instrument.util.StringUtils;
import io.micrometer.core.lang.Nullable;
import org.springframework.boot.context.properties.ConfigurationPropertiesBinding;
import org.springframework.core.annotation.Order;
import org.springframework.core.convert.converter.Converter;

import java.time.Duration;
import java.time.format.DateTimeParseException;
import java.util.Optional;
import java.util.function.Function;

/**
 * A {@link Converter} to create a {@link Duration} from a {@link String}.
 *
 * @author Jon Schneider
 * @author Andy Wilkinson
 */
@ConfigurationPropertiesBinding
@Order(0)
public class StringToDurationConverter implements Converter<String, Duration> {

    @Nullable
    private static Duration simpleParse(@Nullable String rawTime) {
        if (StringUtils.isEmpty(rawTime))
            return null;
        if (!Character.isDigit(rawTime.charAt(0)))
            return null;

        String time = rawTime.toLowerCase();
        return tryParse(time, "ns", Duration::ofNanos)
            .orElseGet(() -> tryParse(time, "ms", Duration::ofMillis)
                .orElseGet(() -> tryParse(time, "s", Duration::ofSeconds)
                    .orElseGet(() -> tryParse(time, "m", Duration::ofMinutes)
                        .orElseGet(() -> tryParse(time, "h", Duration::ofHours)
                            .orElseGet(() -> tryParse(time, "d", Duration::ofDays)
                                .orElse(null))))));
    }

    private static Optional<Duration> tryParse(String time, String unit, Function<Long, Duration> toDuration) {
        if (time.endsWith(unit)) {
            String trim = time.substring(0, time.lastIndexOf(unit)).trim();
            try {
                return Optional.of(toDuration.apply(Long.parseLong(trim)));
            } catch (NumberFormatException ignore) {
                return Optional.empty();
            }
        }
        return Optional.empty();
    }

    @Override
    public Duration convert(@Nullable String source) {
        Duration duration = simpleParse(source);
        try {
            return duration == null ? Duration.parse(source) : duration;
        } catch (DateTimeParseException e) {
            throw new IllegalArgumentException("Cannot convert '" + source + "' to Duration", e);
        }
    }
}
