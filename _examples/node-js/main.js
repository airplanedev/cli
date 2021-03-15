
const c = require('cowsay2');
const argv = require('process').argv;

console.log(c.say(argv[2] || 'Hello, World!'));
