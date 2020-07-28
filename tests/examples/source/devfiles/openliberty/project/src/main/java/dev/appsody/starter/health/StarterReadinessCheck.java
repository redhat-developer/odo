package dev.appsody.starter.health;

import javax.enterprise.context.ApplicationScoped;

import org.eclipse.microprofile.health.HealthCheck;
import org.eclipse.microprofile.health.HealthCheckResponse;
import org.eclipse.microprofile.health.Readiness;

@Readiness
@ApplicationScoped
public class StarterReadinessCheck implements HealthCheck {

    private boolean isReady() {
        // perform readiness checks, e.g. database connection, etc.

        return true;
    }
	
    @Override
    public HealthCheckResponse call() {
        boolean up = isReady();
        return HealthCheckResponse.named(this.getClass().getSimpleName()).state(up).build();
    }
    
}
