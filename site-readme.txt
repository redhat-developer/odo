In order to view this website, please use a webserver such as SimpleHTTPServer or Nginx

For example:

python -m SimpleHTTPServer 8080

OR

docker run -it --rm -p 8080:80 --name web -v $PWD:/usr/share/nginx/html nginx

Then visit:

localhost:8080

Thanks!
