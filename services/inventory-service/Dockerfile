# Build stage
FROM gradle:8.5-jdk21 AS builder

WORKDIR /app

COPY build.gradle.kts .
COPY settings.gradle.kts .
COPY src src/

RUN gradle build -x test --no-daemon

# Runtime stage
FROM eclipse-temurin:21-jre-alpine

RUN apk add --no-cache curl

WORKDIR /app

COPY --from=builder /app/build/libs/*.jar app.jar

EXPOSE 50051 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/actuator/health || exit 1

ENTRYPOINT ["java", \
    "-Dspring.profiles.active=production", \
    "-Xms256m", \
    "-Xmx512m", \
    "-jar", \
    "app.jar"]
