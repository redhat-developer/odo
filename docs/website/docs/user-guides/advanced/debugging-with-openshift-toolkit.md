---
title: Running an Application with OpenShift Toolkit
sidebar_position: 7
---

Debugging is an unavoidable part of development, and it can prove even more difficult when developing an application that runs remotely.

However, this task is made absurdly simple with the help of OpenShift Toolkit.

## OpenShift Toolkit
[OpenShift Toolkit](https://github.com/redhat-developer/intellij-openshift-connector) is an IDE plugin that allows you to do all things that `odo` does, i.e. create, test, debug and deploy cloud-native applications on a cloud-native environment in simple steps.
`odo` enables this plugin to do what it does.

## Prerequisites
1. [You have logged in to your cluster](../quickstart/nodejs.md#step-1-connect-to-your-cluster-and-create-a-new-namespace-or-project).
2. [You have initialized a Node.js application with odo](../quickstart/nodejs.md#step-2-initializing-your-application--odo-init-).
3. Open the application in the IDE.
4. Install OpenShift Toolkit Plugin in your preferred VS Code or a Jet Brains IDE.

In the plugin window, you should be able to see the cluster you are logged into in "APPLICATION EXPLORER" section, and your component "my-nodejs-app" in "COMPONENTS" section.


## Step 1. Start the Dev session to run the application on cluster

1. Right click on "my-nodejs-app" and select "Start on Dev".

![Starting Dev session](../../assets/user-guides/advanced/Start%20Dev%20Session.png)

2. Wait until the application is running on the cluster.

![Wait until Dev session finishes](../../assets/user-guides/advanced/Wait%20until%20Dev%20Session%20finishes.png)

Our application is now available at 127.0.0.1:20001. The debug server is running at 127.0.0.1:20002.

## Step 2. Start the Debugging session

1. Right click on "my-nodejs-app" and select "Debug".

![Select Debug](../../assets/user-guides/advanced/Select%20Debug%20Session.png)

2. Debug session should have started successfully at the debug port, in this case, 3000. And you must be looking at the "DEBUG CONSOLE".

![Debug session starts](../../assets/user-guides/advanced/Debug%20Session%20Starts.png)

## Step 3. Set Breakpoints in the application

Now that the debug session is running, we can set breakpoints in the code.

1. Open 'server.js' file if you haven't opened it already. We will set a breakpoint on Line 55 by clicking the red dot that appears right next to line numbers.

![Add breakpoint](../../assets/user-guides/advanced/Add%20Breakpoint.png)

2. From a new terminal, or a browser window, ping the url at which the application is available, in this case, it is 127.0.0.1:20001.

![Ping Application](../../assets/user-guides/advanced/Ping%20Application.png)

3. The debug session should halt execution at the breakpoint, at which point you can start debugging the application.

![Application Debugged](../../assets/user-guides/advanced/Application%20Debugged.png)


To learn more about running and debugging an application on cluster with OpenShift Toolkit, see the links below.
1. [Using OpenShift Toolkit - project with existing devfile](https://www.youtube.com/watch?v=2jfV0QqG8Sg)
2. [Using OpenShift Toolkit with two microservices](https://www.youtube.com/watch?v=8SpV6UZ23_c)
3. [Using OpenShift Toolkit - project without devfile](https://www.youtube.com/watch?v=sqqznqoWNSg)
