var http = require('http');
var port = process.env.PORT || process.env.port || process.env.OPENSHIFT_NODEJS_PORT || 8080;
var ip = process.env.OPENSHIFT_NODEJS_IP || '0.0.0.0';
var server = http.createServer(function (req, res) {
  var body = '';
  req.on('data', function (data) {
    body += data;
  });
  req.on('end', function () {
    res.writeHead(200, {'Content-Type': 'text/plain'});

    res.write('Hello world from node.js!\n');
    res.end('\n');

  });
});
server.listen(port);
console.log('Server running on ' + ip + ':' + port);
