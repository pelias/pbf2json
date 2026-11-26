
const child = require('child_process');
const index = require('../index');

module.exports.tests = {};

module.exports.tests.interface = function(test) {
  test('valid interface', function(t) {
    t.equal(typeof index.createReadStream, 'function', 'valid function');
    t.end();
  });
};

function fakeBinary( script ){
  const spawn = child.spawn;
  child.spawn = function(){
    child.spawn = spawn;
    return spawn( process.execPath, [ '-e', script ] );
  };
}

function collect( stream, cb ){
  const records = [];
  stream.on( 'data', function( obj ){ records.push( obj ); } );
  stream.on( 'end', function(){ cb( null, records ); } );
  stream.on( 'error', function( err ){ cb( err, records ); } );
}

module.exports.tests.exit = function(test) {
  const lines = 'console.log(JSON.stringify({ id: 1 })); console.log(JSON.stringify({ id: 2 }));';

  test('clean exit ends the stream', function(t) {
    fakeBinary( lines );
    collect( index.createReadStream({ file: 'test.pbf' }), function( err, records ){
      t.error( err, 'no error' );
      t.deepEqual( records, [{ id: 1 }, { id: 2 }], 'all records emitted' );
      t.end();
    });
  });

  test('non-zero exit code errors the stream', function(t) {
    fakeBinary( lines + 'process.exitCode = 3;' );
    collect( index.createReadStream({ file: 'test.pbf' }), function( err ){
      t.ok( err, 'stream errored' );
      t.equal( err && err.message, 'pbf2json exited with code 3' );
      t.end();
    });
  });

  test('termination by signal errors the stream', function(t) {
    fakeBinary( lines + 'process.stdout.write("", function(){ process.kill(process.pid, "SIGKILL"); });' );
    collect( index.createReadStream({ file: 'test.pbf' }), function( err ){
      t.ok( err, 'stream errored' );
      t.equal( err && err.message, 'pbf2json was terminated by signal SIGKILL' );
      t.end();
    });
  });

  test('kill ends the stream without error', function(t) {
    fakeBinary( 'setInterval(function(){}, 1000);' );
    const stream = index.createReadStream({ file: 'test.pbf' });
    collect( stream, function( err ){
      t.error( err, 'no error' );
      t.end();
    });
    stream.kill();
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
