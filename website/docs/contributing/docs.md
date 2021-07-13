---
title: Documentation
sidebar_position: 3
---
Below steps should help you get started with contributing to this website. Note that angular brackets (`<`, `>`) indicate placeholder data; you are not supposed to use them, but replace the brackets and text inside it with relevant information:
* Fork the [odo repo](https://github.com/openshift/odo/) on GitHub under your namespace.
* Clone the forked repository on your system:
  ```shell
  $ git clone https://github.com/<your-namespace>/odo/
  ```
 
* Now `cd` into the directory where you cloned the repository: 
  ```shell
  $ cd odo
  ```
* Create a branch for the issue you are working on:
  ```shell
  $ git checkout -b <branch-name>
  ```
  
* Website documentation is under the `website/` directory in the root of the repo, and the markdown files rendered on this website are in `docs/` directory inside that. So to make changes to these, `cd` into it:
  ```shell
  $ cd website/docs
  ```
  
* Make the changes you want to propose to the documentation. 
  
* To see how your proposed change will look like on the website, you can run local instance of this website on your system. To do this, run below command from `website/docs` directory:
  ```shell
  # if you are doing it for the first time
  $ npm install # this command installs dependencies required to create the website
  
  $ npm run start
  ```
  
* When you are happy with the changes, push it to your fork:
  ```shell
  $ git add <your-changed-files>
  $ git commit --message "<brief-explanation-of-changes>"
  $ git push origin <branch-name> # use the <branch-name> from earlier step
  ```
* Open a pull request by visiting the [odo repository on GitHub](https://github.com/openshift/odo/).