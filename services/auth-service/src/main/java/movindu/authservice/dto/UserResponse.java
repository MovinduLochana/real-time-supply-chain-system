package movindu.authservice.dto;

public record UserResponse(
        String id,
        String email,
        String username,
        String firstName,
        String lastName,
        boolean enabled,
        java.util.List<String> roles
) {

    public UserResponse(String id, String email, String firstName, String lastName, java.util.List<String> roles) {
        this(id, email, "username", firstName, lastName, true, roles);
    }
}
