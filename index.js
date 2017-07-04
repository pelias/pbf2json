
// spawn a 'pbf' process and manage it
// see: https://github.com/missinglink/pbf

var fs = require('fs'),
    tmp = require('tmp'),
    proc = require('./lib/process'),
    jsonStream = require('./lib/jsonStream');

// convert the legacy 'features' file format to a config file
// as required by the missinglink/pbf library
function writeConfig( features ){

  // create a new temp file
  var tmpFileName = tmp.fileSync().name;

  // convert pbf2json format to pbf format
  var tags = features.map( function( feat ){
    return feat.replace(/~/g, '=').split('+');
  });

  // write a config file as required by missinglink/pbf
  fs.writeFileSync(tmpFileName, JSON.stringify({
    node: tags,
    way: tags
  }), 'utf8');

  // return the tmp file path
  return tmpFileName;
}

// execute the 'genmask' command synchronously
function genmask( settings ){

  // command line arguments
  var args = ['genmask', '-i'];
  args.push( '-c', settings.config );
  args.push( settings.pbf );
  args.push( settings.mask );

  // debug command
  // console.error( args );

  // return the child process object
  return proc( args, true );
}

// execute the 'jsonflat' command asyncronously
function jsonflat( settings ){

  // command line arguments
  var args = ['json-flat', '-c'];
  args.push( '-m', settings.mask );
  args.push( '-l', settings.leveldb );
  args.push( settings.pbf );

  // debug command
  // console.error( args );

  // spawn child process
  var child = proc( args );

  // return stream of parsed json objects from process stdout
  return jsonStream( child );
}

// create a readable object stream from the output of the 'jsonflat' command.
// note: we run the 'genmask' command first to ensure that both the mask
// and the index file have been created (which are required by 'jsonflat').
function createReadStream( conf ){

  // a 'state' object, with all the settings we will be using
  var settings = {
    config: writeConfig( conf.tags ),
    mask: tmp.tmpNameSync(),
    leveldb: conf.leveldb || tmp.dirSync().name,
    pbf: conf.file
  };

  // launch the genmask command synchronously
  var genmaskProc = genmask( settings );
  if( genmaskProc.status > 0 ){
    throw new Error( 'pbf genmask: process exited with code: ' + genmaskProc.status );
  }

  // return stream of parsed json objects from process stdout
  return jsonflat( settings );
}

module.exports.writeConfig = writeConfig;
module.exports.genmask = genmask;
module.exports.jsonflat = jsonflat;
module.exports.createReadStream = createReadStream;
