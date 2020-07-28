package it.dev.appsody.starter;

import static org.junit.Assert.assertEquals;

import javax.json.JsonObject;
import javax.ws.rs.client.Client;
import javax.ws.rs.client.ClientBuilder;
import javax.ws.rs.core.Response;

import org.apache.cxf.jaxrs.provider.jsrjsonp.JsrJsonpProvider;
import org.junit.After;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;

public class HealthEndpointTest {
    
    private static String baseUrl;
    private static final String LIVENESS_ENDPOINT = "/health/live";
    private static final String READINESS_ENDPOINT = "/health/ready";
    private Client client;
    private Response response;
    
    @BeforeClass
    public static void oneTimeSetup() {
        String port = System.getProperty("liberty.test.port");
        baseUrl = "http://localhost:" + port;
    }
    
    @Before
    public void setup() {
        response = null;
        client = ClientBuilder.newClient();
        client.register(JsrJsonpProvider.class);
    }
    
    @After
    public void teardown() {
        response.close();
        client.close();
    }

    @Test
    public void testLivenessEndpoint() {
        checkHealthEndpoint(LIVENESS_ENDPOINT, "alive");

    }
    
    @Test
    public void testReadinessEndpoint() {
        checkHealthEndpoint(READINESS_ENDPOINT, "ready");

    }

    private void checkHealthEndpoint(String endpoint, String state) {
        String healthURL = baseUrl + endpoint;
        response = this.getResponse(healthURL);
        this.assertResponse(healthURL, response);
        
        JsonObject healthJson = response.readEntity(JsonObject.class);
        
        String expectedOutcome = "UP";
        String actualOutcome = healthJson.getString("status");
        assertEquals("Application should be " + state, expectedOutcome, actualOutcome);
        
        actualOutcome = healthJson.getJsonArray("checks").getJsonObject(0).getString("status");
        assertEquals("First array element was expected to be SystemResource and it wasn't healthy", expectedOutcome, actualOutcome);
    
    }
    
    /**
     * <p>
     * Returns response information from the specified URL.
     * </p>
     *
     * @param url
     *          - target URL.
     * @return Response object with the response from the specified URL.
     */
    private Response getResponse(String url) {
        return client.target(url).request().get();
    }

    /**
     * <p>
     * Asserts that the given URL has the correct response code of 200.
     * </p>
     *
     * @param url
     *          - target URL.
     * @param response
     *          - response received from the target URL.
     */
    private void assertResponse(String url, Response response) {
        assertEquals("Incorrect response code from " + url, 200, response.getStatus());
    }

}
