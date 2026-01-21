package movindu.authservice.service;

import movindu.authservice.config.KeycloakConfig;
import movindu.authservice.dto.UserResponse;
import org.jspecify.annotations.Nullable;
import org.keycloak.admin.client.Keycloak;
import org.keycloak.representations.idm.RoleRepresentation;
import org.keycloak.representations.idm.UserRepresentation;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UserAdminService {

    private final Logger log = LoggerFactory.getLogger(UserAdminService.class);
    private final Keycloak keycloakAdmin;
    private final KeycloakConfig keycloakConfig;

    public UserAdminService(Keycloak keycloakAdmin, KeycloakConfig keycloakConfig) {
        this.keycloakAdmin = keycloakAdmin;
        this.keycloakConfig = keycloakConfig;
    }

    public List<UserResponse> getAllUsers(int page, int size) {
        var realmResource = keycloakAdmin.realm(keycloakConfig.getRealm());

        return realmResource.users().list(page * size, size).stream()
                .map(this::mapToUserResponse)
                .toList();
    }

    public UserResponse getUser(String userId) {
        var user = keycloakAdmin.realm(keycloakConfig.getRealm())
                .users().get(userId).toRepresentation();
        return mapToUserResponse(user);
    }

    public UserResponse updateUserRoles(String userId, List<String> roles) {

        var realmResource = keycloakAdmin.realm(keycloakConfig.getRealm());
        var userResource = realmResource.users().get(userId);
        var allRoles = realmResource.roles().list();
        userResource.roles().realmLevel().remove(allRoles);

        var rolesToAdd = allRoles.stream()
                .filter(role -> roles.contains(role.getName()))
                .toList();
        userResource.roles().realmLevel().add(rolesToAdd);

        return getUser(userId);

    }

    public void deleteUser(String userId) {
        keycloakAdmin.realm(keycloakConfig.getRealm())
                .users().delete(userId).close();
        log.info("User {} deleted", userId);
    }

    public void setUserEnabled(String userId, boolean b) {
        var userResource = keycloakAdmin.realm(keycloakConfig.getRealm())
                .users().get(userId);
        var user = userResource.toRepresentation();
        user.setEnabled(b);
        userResource.update(user);
        log.info("User {} enabled", userId);
    }

    private UserResponse mapToUserResponse(UserRepresentation user) {
        List<String> roles = keycloakAdmin.realm(keycloakConfig.getRealm())
                .users().get(user.getId()).roles().realmLevel().listEffective()
                .stream()
                .map(RoleRepresentation::getName)
                .filter(role -> !role.startsWith("default-roles"))
                .toList();

        return new UserResponse(
                user.getId(),
                user.getEmail(),
                user.getUsername(),
                user.getFirstName(),
                user.getLastName(),
                user.isEnabled(),
                roles
        );
    }
}
