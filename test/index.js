
var index = require('../index');

module.exports.tests = {};

module.exports.tests.interface = function(test) {
  test('valid interface', function(t) {
    t.equal(typeof index.createReadStream, 'function', 'valid function');
    t.end();
  });
};

module.exports.all = function (tape, common) {

  function test(name, testFunction) {
    return tape('index: ' + name, testFunction);
  }

  for( var testCase in module.exports.tests ){
    module.exports.tests[testCase](test, common);
  }
};
