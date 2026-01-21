package movindu.authservice.dto;

import java.time.Instant;

public record TokenValidationResponse(
    Boolean valid,
    String subject,
    Instant expiresAt
){
}
