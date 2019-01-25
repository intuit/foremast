FROM openjdk:8-jdk-alpine

WORKDIR /app
RUN apk add maven
COPY . /app
#RUN ./build.sh

EXPOSE 8080
COPY target/k8s-metrics-demo-0.1.3.jar /app

RUN mkdir /app/resources
COPY src/main/resources/data.txt /app/resources
COPY src/main/resources/data2.txt /app/resources

ENTRYPOINT exec java $JAVA_OPTS -jar target/k8s-metrics-demo-0.1.3.jar
