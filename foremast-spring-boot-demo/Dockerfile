FROM openjdk:8-jdk-alpine
  
WORKDIR /app

EXPOSE 8080
COPY target/foremast-spring-boot-demo-0.1.0-SNAPSHOT.jar /app


ENTRYPOINT exec java $JAVA_OPTS -jar ./foremast-spring-boot-demo-0.1.0-SNAPSHOT.jar
