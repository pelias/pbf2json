const generateParams = require('../../lib/generateParams');

module.exports.tests = {};

module.exports.tests.params = function(test) {
  test('PBF file', function(t) {
    const config = {
      file: '/some/path/to/osm.pbf'
    };

    const params = generateParams(config);

    t.equal(params[params.length - 1], '/some/path/to/osm.pbf', 'final parameter is path to PBF file');
    t.end();
  });
  test('PBF file', function(t) {
    const config = {
      tags: [
        'tag:one',
        'tag:two',
        'combination~tags'
      ]
    };

    const params = generateParams(config);

    const expected = '-tags=tag:one,tag:two,combination~tags';

    t.equal(params[0], expected, 'tag array is serialized into parameter');
    t.end();
  });

  test('waynodes', function(t) {
    const config = {
      waynodes: true
    };

    const params = generateParams(config);

    const expected = '--waynodes=true';

    t.equal(params[0], expected, 'waynodes is serialized into parameter');
    t.end();
  });

  test('metadata: true', function(t) {
    const options = {
      metadata: true
    };

    const params = generateParams({ options });

    const expected = '--metadata=true';

    t.equal(params[0], expected, 'metadata is set true');
    t.end();
  });

  test('metadata: false', function(t) {
    const options = {
      metadata: false
    };

    const params = generateParams({ options });

    t.equal(params[0], undefined, 'metadata is unset for false');
    t.end();
  });

  test('metadata: invalid', function(t) {
    const options = {
      metadata: 'foo'
    };

    const params = generateParams({ options });

    t.equal(params[0], undefined, 'metadata is invalid');
    t.end();
  });

  test('metadata: empty', function(t) {
    const params = generateParams({});

    t.equal(params[0], undefined, 'options is unset');
    t.end();
  });

  test('metadata: omit', function(t) {
    const params = generateParams({ options: {} });

    t.equal(params[0], undefined, 'metadata is unset');
    t.end();
  });
};

module.exports.all = function (tape, common) {

  function test(name, testFunction) {
    return tape('generateParams: ' + name, testFunction);
  }

  for( var testCase in module.exports.tests ){
    module.exports.tests[testCase](test, common);
  }
};
