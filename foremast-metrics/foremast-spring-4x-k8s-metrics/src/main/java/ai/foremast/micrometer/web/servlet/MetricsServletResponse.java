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
package ai.foremast.micrometer.web.servlet;

import javax.servlet.http.HttpServletResponse;
import javax.servlet.http.HttpServletResponseWrapper;
import java.io.IOException;

/**
 * Wrapper for servlet response to get the status in servlet 2.5
 *
 * @author Sheldon Shao
 * @version 1.0
 */
public class MetricsServletResponse extends HttpServletResponseWrapper {

    private int status = 200;

    public MetricsServletResponse(HttpServletResponse response) {
        super(response);
    }

    public void sendError(int sc, String msg) throws IOException {
        this.status = sc;
        super.sendError(sc, msg);
    }

    public void sendError(int sc) throws IOException {
        this.status = sc;
        super.sendError(sc);
    }

    /**
     * Get the HTTP status code for this Response.
     *
     * @return The HTTP status code for this Response
     *
     * @since Servlet 3.0
     */
    public int getStatus(){
        return status;
    }

    public void setStatus(int sc) {
        this.status = sc;
        super.setStatus(sc);
    }

    public void setStatus(int sc, String sm) {
        this.status = sc;
        super.setStatus(sc, sm);
    }
}
