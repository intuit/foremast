FROM openjdk:8-jdk-alpine

WORKDIR /app

EXPOSE 8080
COPY target/k8s-metrics-demo-0.1.5-SNAPSHOT.jar /app

RUN mkdir /app/resources
COPY src/main/resources/data1.txt /app/resources
COPY src/main/resources/data2.txt /app/resources

ENTRYPOINT exec java $JAVA_OPTS -jar ./k8s-metrics-demo-0.1.5-SNAPSHOT.jar
