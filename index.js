
var util = require('util'),
    path = require('path'),
    os = require('os'),
    split = require('split'),
    through = require('through2'),
    child = require('child_process'),
    exec = path.join(__dirname, 'build', util.format( 'pbf2json.%s-%s', os.platform(), os.arch() ) );

function errorHandler( name ){
  return function( data ){
    data.toString('utf8').trim().split('\n').forEach( function( line ){
      console.error( util.format( '[%s]:', name ), line );
    });
  };
}

function createReadStream( config ){

  var flags = [];
  flags.push( util.format( '-tags=%s', config.tags ) );
  if( config.hasOwnProperty( 'leveldb' ) ){
    flags.push( util.format( '-leveldb=%s', config.leveldb ) );
  }
  flags.push( config.file );

  var proc = child.spawn( exec, flags );

  // propagate signals from parent to child
  process.on('SIGINT',  function(){ proc.kill(); });
  process.on('SIGTERM', function(){ proc.kill(); });

  var decoder = createJsonDecodeStream();
  proc.stdout
    .pipe( split() )
    .pipe( through( function( chunk, enc, next ){
      var str = chunk.toString('utf8'); // convert buffers to strings
      // remove empty lines
      if( 'string' === typeof str && str.length ){
        this.push( str );
      }
      next();
    }))
    .pipe( decoder );

  // print error and exit on decoder pipeline error
  decoder.on( 'error', errorHandler( 'decoder' ) );

  // print error and exit on stderr
  proc.stderr.on( 'data', errorHandler( 'pbf2json' ) );

  // terminate the process and pipeline
  decoder.kill = function(){
    proc.kill();
    decoder.end();
  };

  return decoder;
}

function createJsonDecodeStream(){
  return through.obj( function( str, enc, next ){
    try {
      var o = JSON.parse( str );
      if( o ){ this.push( o ); }
    }
    catch( e ){
      this.emit( 'error', e );
    }
    finally {
      next();
    }
  });
}

module.exports.createReadStream = createReadStream;
