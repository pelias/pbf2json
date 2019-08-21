
var tape = require('tape');
var common = {};

var tests = [
  require('./index.js'),
  require('./lib/generateParams.js')
];

tests.map(function(t) {
  t.all(tape, common);
});
