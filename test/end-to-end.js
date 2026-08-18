
/**
  End-to-end tests of a small pbf extract.

  The somes.osm.pbf extract will be automatically downloaded before testing.
  @see: ./pretest.sh for more details, or run manually to download file.
**/

var fs = require('fs'),
    os = require('os'),
    path = require('path'),
    deep = require('deep-diff'),
    through = require('through2'),
    pbf2json = require('../index');

var workdir = fs.mkdtempSync( path.join( os.tmpdir(), 'pbf2json-e2e-' ) );

function test( name, tags, cb ){

  var tmpfile = path.join( workdir, name + '.json' ),
      leveldbDir = fs.mkdtempSync( path.join( workdir, 'leveldb-' ) ),
      pbfPath = path.resolve(__dirname) + '/vancouver_canada.osm.pbf',
      expectedPath = path.resolve(__dirname) + '/fixtures/' + name + '.json',
      actual = {};

  // give each test its own leveldb directory, the default of '/tmp' is shared
  // between concurrent runs and leaves its files behind
  pbf2json.createReadStream({ file: pbfPath, tags: tags, leveldb: leveldbDir })
    .pipe( through.obj( function( obj, _, next ){
      obj.gid = obj.type + ':' + obj.id;
      actual[ obj.gid ] = obj;
      next();
    }))
    .on('finish', function assert(){

      // write actual to disk, so failures can be inspected by hand
      fs.writeFileSync( tmpfile, JSON.stringify( actual, null, 2 ) );
      fs.rmSync( leveldbDir, { recursive: true, force: true } );

      var expected = JSON.parse( fs.readFileSync( expectedPath, { encoding: 'utf8' } ) );

      // actual != expected
      if( !deepEqual( actual, expected ) ){
        console.error( 'end-to-end tests failed :(' );
        console.error( 'contents of', tmpfile, 'do not match expected:', expectedPath );
        process.exit(1);
      }

      cb();
    });

}

var tests = [
  [ 'single',     ['building'] ],
  [ 'multiple',   ['building','shop'] ],
  [ 'colon',      ['addr:housenumber'] ],
  [ 'group',      ['addr:housenumber+addr:street'] ],
  [ 'multigroup', ['highway+name','waterway+name'] ],
  [ 'value',      ['amenity~toilets'] ],
  [ 'multivalue', ['amenity~toilets','amenity~kindergarten'] ]
];

function next(){
  var t = tests.shift();
  if( t ){ return test( t[0], t[1], next ); }
  fs.rmSync( workdir, { recursive: true, force: true } ); // only on success, failures exit early
}

// deep equal comparison, optimised for fast fail
var deepEqual = function(a, b) {
  if(!a || !b){ return false; }
  if(Object.keys(a).length !== Object.keys(b).length){ return false; }
  for(var i in a) {
    if( !b.hasOwnProperty(i) ){ return false; }
    if( deep.diff(a[i], b[i]) ){ return false; }
  }
  return true;
};

// run each test synchronously
next();
