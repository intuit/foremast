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
package ai.foremast.micrometer.autoconfigure.export.simple;

import io.micrometer.core.instrument.simple.CountingMode;
import io.micrometer.core.instrument.simple.SimpleConfig;
import ai.foremast.micrometer.autoconfigure.export.properties.PropertiesConfigAdapter;

import java.time.Duration;

/**
 * Adapter to convert {@link SimpleProperties} to a {@link SimpleConfig}.
 *
 * @author Jon Schneider
 */
public class SimplePropertiesConfigAdapter extends PropertiesConfigAdapter<SimpleProperties> implements SimpleConfig {

    public SimplePropertiesConfigAdapter(SimpleProperties properties) {
        super(properties);
    }

    @Override
    public String get(String key) {
        return null;
    }

    @Override
    public Duration step() {
        return get(SimpleProperties::getStep, SimpleConfig.super::step);
    }

    @Override
    public CountingMode mode() {
        return get(SimpleProperties::getMode, SimpleConfig.super::mode);
    }

}
