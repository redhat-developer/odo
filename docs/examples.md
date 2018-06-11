# Examples

Odo is compatible with any language listed within OpenShift's Catalog service.

This can be found by using `odo catalog list`.

Example:

```sh
The following components can be deployed:
- httpd
- nodejs
- perl
- php
- python
- ruby
- wildfly
```

### httpd

Build and serve static content via httpd on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/sclorg/httpd-container/blob/master/2.4/README.md.

```sh
  odo create httpd --git https://github.com/openshift/httpd-ex.git
```

### nodejs

Build and run Node.js applications on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/sclorg/s2i-nodejs-container/blob/master/4/README.md.

```sh
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git
```

### perl

Build and run Perl applications on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/sclorg/s2i-perl-container/blob/master/5.24/README.md.

```sh
  odo create perl --git https://github.com/openshift/dancer-ex.git
```

### php

Build and run PHP applications on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/sclorg/s2i-php-container/blob/master/7.0/README.md.

```sh
  odo create php --git https://github.com/openshift/cakephp-ex.git
```

### python

Build and run Python applications on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/sclorg/s2i-python-container/blob/master/3.5/README.md.

```sh
  odo create python --git https://github.com/openshift/django-ex.git
```

### ruby

Build and run Ruby applications on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/sclorg/s2i-ruby-container/blob/master/2.3/README.md.

```sh
  odo create ruby --git https://github.com/openshift/ruby-ex.git
```

### wildfly

Build and run WildFly applications on CentOS 7. For more information about using this builder image, including OpenShift considerations, see https://github.com/openshift-s2i/s2i-wildfly/blob/master/README.md.

```sh
  odo create wildfly --git https://github.com/openshift/openshift-jee-sample.git
```
