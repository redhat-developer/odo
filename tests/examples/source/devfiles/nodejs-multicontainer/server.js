require('appmetrics-dash').attach();

const appName = require('./../package').name;
const http = require('http');
const express = require('express');
const logger = require('pino')({
  name: appName,
  level: process.env.LOG_LEVEL || 'info',
});
const localConfig = require('./config/local.json');
const path = require('path');

const app = express();
const server = http.createServer(app);

app.use(require('express-pino-logger')({logger}))

require('./routers/index')(app, server);

// Add your code here

const port = process.env.PORT || localConfig.port;
server.listen(port, function(){
  logger.info(`node listening on http://localhost:${port}`);
});

app.use(function (req, res, next) {
  res.sendFile(path.join(__dirname, '../public', '404.html'));
});

app.use(function (err, req, res, next) {
	res.sendFile(path.join(__dirname, '../public', '500.html'));
});

module.exports = server;
