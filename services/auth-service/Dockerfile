FROM eclipse-temurin:25-jdk-alpine AS build
WORKDIR /auth-service

COPY gradle gradle
COPY gradlew build.gradle.kts settings.gradle.kts ./
RUN ./gradlew dependencies --no-daemon

COPY src src
RUN ./gradlew bootJar --no-daemon -x test

# Runtime Only
FROM eclipse-temurin:25-jre-alpine
WORKDIR /app

# No root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=build /app/build/libs/auth-service-0.0.1-SNAPSHOT.jar ./auth-service.jar
EXPOSE 8081
ENTRYPOINT ["java", "-jar", "auth-service.jar"]

