package movindu.authservice.service;

import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Scope;
import jakarta.ws.rs.core.Response;
import movindu.authservice.config.KeycloakConfig;
import movindu.authservice.dto.*;
import org.keycloak.admin.client.Keycloak;
import org.keycloak.admin.client.resource.RealmResource;
import org.keycloak.admin.client.resource.UsersResource;
import org.keycloak.representations.idm.CredentialRepresentation;
import org.keycloak.representations.idm.RoleRepresentation;
import org.keycloak.representations.idm.UserRepresentation;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@Service
public class AuthService {

    private final Keycloak keycloakAdmin;
    private final KeycloakConfig keycloakConfig;
    private RestTemplate restTemplate;
    private final Tracer tracer;

    public AuthService(Keycloak keycloakAdmin, KeycloakConfig keycloakConfig, Tracer tracer) {
        this.keycloakAdmin = keycloakAdmin;
        this.keycloakConfig = keycloakConfig;
        this.tracer = tracer;
    }

    @Value("${keycloak.auth-server-url}")
    private String authServerUrl;

    @Value("${keycloak.realm}")
    private String realm;

    @Value("${keycloak.client-id}")
    private String clientId;

    @Value("${keycloak.client-secret}")
    private String clientSecret;

    public TokenResponse login(LoginRequest request) {
        var span = tracer.spanBuilder("keycloak.token").startSpan();
        try (var _ = span.makeCurrent()) {
            String tokenUrl = authServerUrl + "/realms/" + realm + "/protocol/openid-connect/token";

            var headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_FORM_URLENCODED);

            var body = new LinkedMultiValueMap<String, String>();
            body.add("grant_type", "password");
            body.add("client_id", clientId);
            body.add("client_secret", clientSecret);
            body.add("username", request.username());
            body.add("password", request.password());

            var entity = new HttpEntity<MultiValueMap<String, String>>(body, headers);

            try {
                var response = restTemplate.exchange(tokenUrl, HttpMethod.POST, entity, Map.class);
                var responseBody = response.getBody();

                span.setAttribute("auth.success", true);

                return new TokenResponse(
                        (String) responseBody.get("access_token"),
                        (String) responseBody.get("refresh_token"),
                        (Integer) responseBody.get("expires_in"),
                        (String) responseBody.get("token_type")
                );
            } catch (Exception e) {
                span.setAttribute("auth.success", false);
                span.recordException(e);
//                log.error("Login failed for user: {}", request.username(), e);
                throw new RuntimeException("Invalid credentials");
            }
        } finally {
            span.end();
        }
    }

    public UserResponse register(RegisterRequest request) {
        var span = tracer.spanBuilder("keycloak.createUser").startSpan();
        try (var _ = span.makeCurrent()) {
            var realmResource = keycloakAdmin.realm(keycloakConfig.getRealm());
            var usersResource = realmResource.users();

            // Check if user already exists
            var existingUsers = usersResource.search(request.email());
            if (!existingUsers.isEmpty()) {
                throw new RuntimeException("User with email " + request.email() + " already exists");
            }

            // Create user representation
            var user = new UserRepresentation();
            user.setEnabled(true);
            user.setUsername(request.email());
            user.setEmail(request.email());
            user.setFirstName(request.firstName());
            user.setLastName(request.lastName());
            user.setEmailVerified(false);

            // Set password
            var credential = new CredentialRepresentation();
            credential.setTemporary(false);
            credential.setType(CredentialRepresentation.PASSWORD);
            credential.setValue(request.password());
            user.setCredentials(Collections.singletonList(credential));

            // Create user
            var response = usersResource.create(user);
            if (response.getStatus() != 201) {
                throw new RuntimeException("Failed to create user: " + response.getStatusInfo());
            }

            // Get created user ID
            String userId = response.getLocation().getPath().replaceAll(".*/([^/]+)$", "$1");

            // Assign default role
            var customerRole = realmResource.roles().get("customer").toRepresentation();
            usersResource.get(userId).roles().realmLevel().add(Collections.singletonList(customerRole));

            span.setAttribute("user.id", userId);

            return new UserResponse(
                    userId,
                    request.email(),
                    request.firstName(),
                    request.lastName(),
                    Collections.singletonList("customer")
            );

        } finally {
            span.end();
        }
    }

    public TokenResponse refreshToken(RefreshTokenRequest request) {
        String tokenUrl = authServerUrl + "/realms/" + realm + "/protocol/openid-connect/token";

        var headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_FORM_URLENCODED);

        MultiValueMap<String, String> body = new LinkedMultiValueMap<>();
        body.add("grant_type", "refresh_token");
        body.add("client_id", clientId);
        body.add("client_secret", clientSecret);
        body.add("refresh_token", request.refreshToken());

        var entity = new HttpEntity<>(body, headers);

        try {
            var response = restTemplate.exchange(tokenUrl, HttpMethod.POST, entity, Map.class);
            var responseBody = response.getBody();

            return new TokenResponse(
                    (String) responseBody.get("access_token"),
                    (String) responseBody.get("refresh_token"),
                    (Integer) responseBody.get("expires_in"),
                    (String) responseBody.get("token_type")
            );
        } catch (Exception e) {
//            log.error("Token refresh failed", e);
            throw new RuntimeException("Invalid refresh token");
        }
    }

    public void logout(String userId) {
        try {
            keycloakAdmin.realm(keycloakConfig.getRealm()).users().get(userId).logout();
//            log.info("User {} logged out successfully", userId);
        } catch (Exception e) {
//            log.error("Logout failed for user: {}", userId, e);
        }
    }

    public UserResponse getUserInfo(String userId) {
        try {
            var user = keycloakAdmin.realm(keycloakConfig.getRealm())
                    .users().get(userId).toRepresentation();

            var roles = keycloakAdmin.realm(keycloakConfig.getRealm())
                    .users().get(userId).roles().realmLevel().listEffective()
                    .stream()
                    .map(RoleRepresentation::getName)
                    .toList();

            return new UserResponse(
                    user.getId(),
                    user.getEmail(),
                    "username",
                    user.getFirstName(),
                    user.getLastName(),
                    user.isEnabled(),
                    roles
            );

        } catch (Exception e) {
//            log.error("Failed to get user info for: {}", userId, e);
            throw new RuntimeException("User not found");
        }
    }
}
