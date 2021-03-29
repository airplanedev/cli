const argv = require('yargs')
  .option('name', {
    type: 'string',
  })
  .argv

console.log(`Hello, ${argv.name}!`);
