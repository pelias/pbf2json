
var fs = require('fs'),
    tmp = require('tmp'),
    proc = require('./lib/process'),
    jsonStream = require('./lib/jsonStream');

// args is an array if all command-line args which will
// be passed when spawning the process.
function api( args, done ){

  // spawn child process
  var child = proc( args );

  // create a json read stream
  return jsonStream( child, done );
}

function writeConfig( features ){
  var tmpFile = tmp.fileSync().name;

  // convert pbf2json format to pbf format
  var tags = features.map( function( feat ){
    return feat.replace(/~/g, '=').split('+');
  });

  fs.writeFileSync(tmpFile, JSON.stringify({
    node: tags,
    way: tags
  }), 'utf8');

  return tmpFile;
}

function genmask( settings, done ){

  var args = ['genmask', '-i'];
  args.push( '-c', settings.config );
  args.push( settings.pbf );
  args.push( settings.mask );

  // debug command
  // console.error( args );
  return api( args, done );
}

function jsonflat( settings, done ){

  var args = ['json-flat', '-c'];
  args.push( '-m', settings.mask );
  args.push( '-l', settings.leveldb );
  args.push( settings.pbf );

  // debug command
  // console.error( args );
  return api( args, done );
}

function createReadStream( conf, done ){

  var settings = {
    config: writeConfig( conf.tags ),
    mask: tmp.tmpNameSync(),
    leveldb: conf.leveldb || tmp.dirSync().name,
    pbf: conf.file
  };

  genmask( settings, function( code ){
    if( code > 0 ){
      return done( new Error( 'pbf genmask: process existed with code: ' + code ) );
    }
    return done( 0, jsonflat( settings ) );
  });
}

module.exports.api = api;
module.exports.writeConfig = writeConfig;
module.exports.genmask = genmask;
module.exports.jsonflat = jsonflat;
module.exports.createReadStream = createReadStream;
