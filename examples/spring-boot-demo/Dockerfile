FROM openjdk:8-jdk-alpine

WORKDIR /app
RUN apk add maven
COPY . /app


EXPOSE 8080
#COPY target/foremast-spring-boot-demo-0.1.3-SNAPSHOT.jar /app
COPY target/k8s-metrics-demo-0.1.3.jar /app

RUN mkdir /app/resources
COPY src/main/resources/data3.txt /app/resources
COPY src/main/resources/data2.txt /app/resources
COPY src/main/resources/data1.txt /app/resources

#ENTRYPOINT exec java $JAVA_OPTS -jar target/foremast-spring-boot-demo-0.1.3-SNAPSHOT.jar
ENTRYPOINT exec java $JAVA_OPTS -jar target/k8s-metrics-demo-0.1.3.jar
