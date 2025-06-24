plugins {
    id("java")
}

group = "net.stef"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    implementation("com.google.code.gson:gson:2.13.1")
    implementation("com.github.luben:zstd-jni:1.5.7-3")

    // gRPC dependencies
    runtimeOnly("io.grpc:grpc-netty-shaded:1.73.0")
    implementation("io.grpc:grpc-protobuf:1.73.0")
    implementation ("io.grpc:grpc-stub:1.73.0")
    implementation("com.google.protobuf:protobuf-java:4.31.1")
    compileOnly("org.apache.tomcat:annotations-api:6.0.53")


    // JMH dependencies for benchmarking
    implementation("org.openjdk.jmh:jmh-core:1.37")
    annotationProcessor("org.openjdk.jmh:jmh-generator-annprocess:1.37")
}

tasks.test {
    useJUnitPlatform()
}

// JMH benchmark task
val jmh by tasks.registering(JavaExec::class) {
    group = "benchmark"
    description = "Run JMH benchmarks."
    classpath = sourceSets["main"].runtimeClasspath
    mainClass.set("net.stef.benchmarks.AllBenchmarks")
}

