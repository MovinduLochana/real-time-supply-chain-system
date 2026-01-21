package movindu.authservice.controller;


import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import jakarta.validation.Valid;
import movindu.authservice.dto.*;
import movindu.authservice.service.AuthService;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/auth")
public class AuthController {

    private final AuthService authService;
    private final Tracer tracer;

    public AuthController(AuthService authService, Tracer tracer) {
        this.authService = authService;
        this.tracer = tracer;
    }

    @PostMapping("/login")
    public ResponseEntity<TokenResponse> login(@Valid @RequestBody LoginRequest request) {
        var span = tracer.spanBuilder("auth.login").startSpan();
        try {
//            log.info("Login attempt for user: {}", request.getUsername());
            TokenResponse response = authService.login(request);
            span.setAttribute("user.name", request.username());
            return ResponseEntity.ok(response);
        } finally {
            span.end();
        }
    }

    @PostMapping("/register")
    public ResponseEntity<UserResponse> register(@Valid @RequestBody RegisterRequest request) {
        Span span = tracer.spanBuilder("auth.register").startSpan();
        try {
//            log.info("Registration attempt for user: {}", request.getEmail());
            UserResponse response = authService.register(request);
            span.setAttribute("user.email", request.email());
            return ResponseEntity.ok(response);
        } finally {
            span.end();
        }
    }

    @PostMapping("/refresh")
    public ResponseEntity<TokenResponse> refresh(@Valid @RequestBody RefreshTokenRequest request) {
//        log.info("Token refresh request");
        TokenResponse response = authService.refreshToken(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/logout")
    public ResponseEntity<Void> logout(@AuthenticationPrincipal Jwt jwt) {
//        log.info("Logout request for user: {}", jwt.getSubject());
        authService.logout(jwt.getSubject());
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/me")
    public ResponseEntity<UserResponse> getCurrentUser(@AuthenticationPrincipal Jwt jwt) {
//        log.debug("Getting current user info for: {}", jwt.getSubject());
        UserResponse response = authService.getUserInfo(jwt.getSubject());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/validate")
    public ResponseEntity<TokenValidationResponse> validateToken(@AuthenticationPrincipal Jwt jwt) {
        return ResponseEntity.ok(new TokenValidationResponse(
                true,
                jwt.getSubject(),
                jwt.getExpiresAt()
        ));
    }

}
