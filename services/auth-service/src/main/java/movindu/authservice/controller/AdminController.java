package movindu.authservice.controller;

import movindu.authservice.dto.UserResponse;
import movindu.authservice.service.UserAdminService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/v1/auth/admin")
@PreAuthorize("hasRole('admin')")
public class AdminController {

    private final UserAdminService userAdminService;
    private final Logger log = LoggerFactory.getLogger(AdminController.class);

    public AdminController(UserAdminService userAdminService) {
        this.userAdminService = userAdminService;
    }

    @GetMapping("/users")
    public ResponseEntity<List<UserResponse>> getAllUsers(@RequestParam int page, @RequestParam(defaultValue = "20") int size) {
        log.info("Admin fetching users - page: {}, size: {}", page, size);
        return ResponseEntity.ok(userAdminService.getAllUsers(page, size));
    }

    @GetMapping("/users/{userId}")
    public ResponseEntity<UserResponse> getUser(@PathVariable String userId) {
        log.info("Admin fetching user: {}", userId);
        return ResponseEntity.ok(userAdminService.getUser(userId));
    }

    @PutMapping("/users/{userId}/roles")
    public ResponseEntity<UserResponse> updateUserRoles(@PathVariable String userId, @RequestBody List<String> roles) {
        log.info("Admin updating roles for user: {} - roles: {}", userId, roles);
        return ResponseEntity.ok(userAdminService.updateUserRoles(userId, roles));
    }

    @DeleteMapping("/users/{userId}")
    public ResponseEntity<Void> deleteUser(@PathVariable String userId) {
        log.info("Admin deleting user: {}", userId);
        userAdminService.deleteUser(userId);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/users/{userId}/disable")
    public ResponseEntity<Void> disableUser(@PathVariable String userId) {
        log.info("Admin disabling user: {}", userId);
        userAdminService.setUserEnabled(userId, false);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/users/{userId}/enable")
    public ResponseEntity<Void> enableUser(@PathVariable String userId) {
        log.info("Admin enabling user: {}", userId);
        userAdminService.setUserEnabled(userId, true);
        return ResponseEntity.noContent().build();
    }
}
