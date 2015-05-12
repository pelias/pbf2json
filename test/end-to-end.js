
/**
  End-to-end tests of a small pbf extract.

  The somes.osm.pbf extract will be automatically downloaded before testing.
  @see: ./pretest.sh for more details, or run manually to download file.
**/

var fs = require('fs'),
    path = require('path'),
    tmp = require('tmp'),
    deep = require('deep-diff'),
    naivedb = require('naivedb'),
    pbf2json = require('../index');

var tmpfile = tmp.fileSync({ postfix: '.json' }).name,
    pbfPath = path.resolve(__dirname) + '/vancouver_canada.osm.pbf',
    expectedPath = path.resolve(__dirname) + '/vancouver.extract.expected.json';

fs.writeFileSync( tmpfile, '{}' ); // init naivedb
var db = naivedb(tmpfile);

pbf2json.createReadStream({ file: pbfPath, tags: ['amenity'] })
  .pipe( db.createWriteStream('id') )
  .on('finish', function assert(){

    db.writeSync();

    var actual = JSON.parse( fs.readFileSync( tmpfile, { encoding: 'utf8' } ) ),
        expected = JSON.parse( fs.readFileSync( expectedPath, { encoding: 'utf8' } ) );

    var diff = deep.diff( actual, expected );

    if( diff ){
      console.log( 'end-to-end error:', tmpfile );
      console.log( 'actual !== expected' );
      console.log( diff );
      process.exit(1);
    }
  });