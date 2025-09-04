
/**
  End-to-end tests of a small pbf extract.

  The somes.osm.pbf extract will be automatically downloaded before testing.
  @see: ./pretest.sh for more details, or run manually to download file.
**/

var fs = require('fs'),
    path = require('path'),
    tmp = require('tmp'),
    deep = require('deep-diff'),
    through = require('through2'),
    naivedb = require('naivedb'),
    pbf2json = require('../index');

function test( name, tags, cb ){

  var tmpfile = tmp.fileSync({ postfix: '.json' }).name,
      pbfPath = path.resolve(__dirname) + '/vancouver_canada.osm.pbf',
      expectedPath = path.resolve(__dirname) + '/fixtures/' + name + '.json';

  fs.writeFileSync( tmpfile, '{}' ); // init naivedb
  var db = naivedb(tmpfile);

  pbf2json.createReadStream({ file: pbfPath, tags: tags })
    .pipe( through.obj( function( obj, _, next ){
      obj.gid = obj.type + ':' + obj.id;
      next(null, obj);
    }))
    .pipe( db.createWriteStream('gid') )
    .on('finish', function assert(){

      // write actual to disk
      db.writeSync();

      // load files from disk
      var actual = JSON.parse( fs.readFileSync( tmpfile, { encoding: 'utf8' } ) ),
          expected = JSON.parse( fs.readFileSync( expectedPath, { encoding: 'utf8' } ) );

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
  if( t ){ test( t[0], t[1], next ); }
}

// deep equal comparison, optimised for fast fail
var deepEqual = function(a, b) {
  if(!a || !b){ return false; }
  if(Object.keys(a).length !== Object.keys(b).length){ return false; }
  for(var i in a) {
    if( !b.hasOwnProperty(i) ){ return false; }
    // centroid values vary slightly between CPU architecture
    let prefilter;
    if( a[i].hasOwnProperty('centroid') && b[i].hasOwnProperty('centroid') ) {
      if (!equal(a[i].centroid.lat, b[i].centroid.lat, 1e-6)) { return false; }
      if (!equal(a[i].centroid.lon, b[i].centroid.lon, 1e-6)) { return false; }
      prefilter = (_path, key) => key === 'centroid'; // skip centroid field for the diff
    }
    if( deep.diff(a[i], b[i], prefilter) ){ return false; }
  }
  return true;
};

// ensure two numbers are equal within a threshold
function equal(a, b, delta = 0.0) {
  return Math.abs(parseFloat(a) - parseFloat(b)) <= delta;
}

// run each test synchronously
next();
