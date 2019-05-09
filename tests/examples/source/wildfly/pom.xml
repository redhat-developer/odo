<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>org.openshift</groupId>
    <packaging>war</packaging>
    <version>1.0</version>
    <name>InsultApp</name>
  
    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>1.8</maven.compiler.source>
    	<maven.compiler.target>1.8</maven.compiler.target>
    </properties>
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-war-plugin</artifactId>
            <version>2.6</version>
            <configuration>
                <failOnMissingWebXml>false</failOnMissingWebXml>
            </configuration>
        </plugin>
    </plugins>
</build>

<profiles>
    <profile>
     <!-- When built in OpenShift the 'openshift' profile will be used when invoking mvn. -->
     <!-- Use this profile for any OpenShift specific customization your app will need. -->
     <!-- By default that is to put the resulting archive into the 'deployments' folder. -->
     <!-- http://maven.apache.org/guides/mini/guide-building-for-different-environments.html -->
     <id>openshift</id>
     <build>
        <finalName>InsultApp</finalName>
        <plugins>
          <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-war-plugin</artifactId>
                <version>2.6</version>
                <configuration>
                    <failOnMissingWebXml>false</failOnMissingWebXml>
                    <outputDirectory>target</outputDirectory>
              		  <warName>ROOT</warName>
                </configuration>
            </plugin>
        </plugins>
      </build>
    </profile>
  </profiles>
	<artifactId>InsultApp</artifactId>
</project>
