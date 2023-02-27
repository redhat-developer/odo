const express = require('express');
const http = require('http');

/*
 * App endpoints are bound to 0.0.0.0:3000
 */
const app = express();
app.get('/', (req, res) => {
  console.log('GET /');
  res.send('Hello from Node.js Application!');
});
app.get('*', (req, res) => { res.status(404).send("Not Found"); });
http.createServer(app).listen(3000, '0.0.0.0', () => { console.log(`App started on 0.0.0.0:3000`); });

/*
 * Admin endpoints are bound to 127.0.0.1:3001
 */
const adminApp = express();
adminApp.get('/', (req, res) => {
  console.log('[admin] GET /');
  res.send('Hello from Node.js Admin Application!');
});
adminApp.get('*', (req, res) => { res.status(404).send("Admin endpoint not Found"); });
http.createServer(adminApp).listen(3001, '127.0.0.1', () => { console.log(`Admin App started on 127.0.0.1:3001`); });
