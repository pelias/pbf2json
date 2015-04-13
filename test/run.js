
var tape = require('tape');
var common = {};

var tests = [
  require('./index.js')
];

tests.map(function(t) {
  t.all(tape, common);
});