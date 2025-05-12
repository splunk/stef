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
}

tasks.test {
    useJUnitPlatform()
}