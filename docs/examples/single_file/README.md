# Single file

You can choose to define a single file with multiple apps. This can be achieved
using a separator(`---`) between two app definitions.

Checkout the snippet from the file [wordpress.yml](wordpress.yml)

```yaml
...
services:
- ports:
  - port: 3306
---
name: wordpress
containers:
- image: wordpress:4
  env:
...
```

The application definition of `database` and `wordpress` is delineated using the
separator `---`.
