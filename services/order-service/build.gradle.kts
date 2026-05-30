plugins {
    id 'java'
    id 'org.springframework.boot' version '3.2.0'
    id 'io.spring.dependency-management' version '1.1.4'
    id 'com.google.protobuf' version '0.9.4'
}

group = 'com.rtscs'
version = '0.1.0-alpha'
sourceCompatibility = '21'

repositories {
    mavenCentral()
}

ext {
    set('springCloudVersion', "2023.0.0")
    grpcVersion = '1.59.0'
    protocVersion = '3.24.0'
}

dependencies {
    // Spring Boot
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa'
    implementation 'org.springframework.boot:spring-boot-starter-actuator'

    // Database
    implementation 'org.postgresql:postgresql:42.7.0'
    implementation 'org.flywaydb:flyway-core:9.22.3'

    // gRPC
    implementation "io.grpc:grpc-netty-shaded:${grpcVersion}"
    implementation "io.grpc:grpc-protobuf:${grpcVersion}"
    implementation "io.grpc:grpc-stub:${grpcVersion}"
    implementation "io.grpc:grpc-services:${grpcVersion}"
    implementation 'io.github.lognet:grpc-spring-boot-starter:2.15.0'

    // gRPC Client
    implementation 'io.github.lognet:grpc-client-spring-boot-starter:2.15.0'

    // Kafka
    implementation 'org.springframework.kafka:spring-kafka'

    // OpenTelemetry
    implementation 'io.opentelemetry:opentelemetry-api:1.32.0'
    implementation 'io.opentelemetry:opentelemetry-sdk:1.32.0'
    implementation 'io.opentelemetry:opentelemetry-exporter-jaeger:1.32.0'
    implementation 'io.opentelemetry.instrumentation:opentelemetry-grpc:1.31.0-alpha'
    implementation 'io.opentelemetry.instrumentation:opentelemetry-spring-boot-starter:1.31.0-alpha'

    // Logging
    implementation 'ch.qos.logback:logback-classic'
    implementation 'net.logstash.logback:logstash-logback-encoder:7.4'

    // Protobuf
    implementation "com.google.protobuf:protobuf-java:${protocVersion}"
    implementation "com.google.protobuf:protobuf-java-util:${protocVersion}"

    // Metrics
    implementation 'io.micrometer:micrometer-registry-prometheus'

    // Testing
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
    testImplementation 'org.testcontainers:testcontainers:1.19.3'
    testImplementation 'org.testcontainers:postgresql:1.19.3'
    testImplementation "io.grpc:grpc-testing:${grpcVersion}"
    testImplementation 'io.grpc:grpc-inprocess:${grpcVersion}'

    // Lombok
    compileOnly 'org.projectlombok:lombok'
    annotationProcessor 'org.projectlombok:lombok'
}

dependencyManagement {
    imports {
        mavenBom "org.springframework.cloud:spring-cloud-dependencies:${springCloudVersion}"
    }
}

protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:${protocVersion}"
    }
    plugins {
        grpc {
            artifact = "io.grpc:protoc-gen-grpc-java:${grpcVersion}"
        }
    }
    generateProtoTasks {
        all()*.plugins {
            grpc {}
        }
    }
}

tasks.named('test') {
    useJUnitPlatform()
}

// Copy proto files from monorepo
task copyProtos {
    doLast {
        copy {
            from '../../proto/v1'
            into 'src/main/proto/rtscs'
        }
    }
}

compileJava.dependsOn copyProtos
