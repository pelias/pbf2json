
var util = require('util'),
    stream = require('stream'),
    StringDecoder = require('string_decoder').StringDecoder,
    child = require('child_process'),
    exec = require('./lib/binaryPath'),
    generateParams = require('./lib/generateParams');

// custom log levels can be detected for lines with the format:
// [level] message
// supported levels (listed from least verbose to most verbose):
// error, warn, info
function getLogLevel( line ){
  if( line.indexOf('[warn]') > -1 ){ return 1; }
  if( line.indexOf('[info]') > -1 ){ return 2; }
  return 0;
}

function errorHandler( name, level ){
  return function( data ){
    data.toString('utf8').trim().split('\n').forEach( function( line ){
      if( getLogLevel( line ) <= level ){
        console.error( util.format( '[%s]:', name ), line );
      }
    });
  };
}

function createReadStream( config ){

  const params = generateParams(config);

  var proc = child.spawn( exec, params );

  // propagate signals from parent to child
  process.on('SIGINT',  function(){ proc.kill(); });
  process.on('SIGTERM', function(){ proc.kill(); });

  const decoder = createJsonDecodeStream();

  // the stream must not end until the exit status is known, a child killed
  // part way through (eg. by the OOM killer) closes stdout just like a
  // successful run does and would otherwise look like a complete import
  proc.stdout.pipe( decoder, { end: false } );

  proc.on( 'error', function( err ){
    decoder.destroy( err );
  });

  function onClose( code, signal ){
    if( code === 0 ){ return decoder.end(); }
    decoder.destroy( new Error( signal ?
      `pbf2json was terminated by signal ${signal}` :
      `pbf2json exited with code ${code}`
    ));
  }
  proc.on( 'close', onClose );

  // errors are only logged here, it's the caller's responsibility to handle
  // them (eg. by exiting non-zero), attaching this listener prevents node
  // from treating them as uncaught
  decoder.on( 'error', errorHandler( 'decoder', config.loglevel || 0 ) );
  proc.stderr.on( 'data', errorHandler( 'pbf2json', config.loglevel || 0 ) );

  // terminate the process and pipeline
  decoder.kill = function(){
    proc.removeListener( 'close', onClose );
    proc.kill();
    decoder.end();
  };

  return decoder;
}

// splits the child process stdout on newlines, emitting one object per line.
// the StringDecoder ensures multi-byte characters spanning two chunks survive.
function createJsonDecodeStream(){

  var utf8 = new StringDecoder('utf8'),
      soFar = '';

  function lines( str ){
    var pieces = ( soFar + str ).split(/\r?\n/);
    soFar = pieces.pop(); // the trailing piece is an incomplete line
    return pieces;
  }

  function decode( line ){
    if( !line.length ){ return; } // remove empty lines
    try {
      var o = JSON.parse( line );
      if( o ){ this.push( o ); }
    }
    catch( e ){
      this.emit( 'error', e );
    }
  }

  return new stream.Transform({
    readableObjectMode: true,
    transform: function( chunk, enc, next ){
      lines( utf8.write( chunk ) ).forEach( decode, this );
      next();
    },
    flush: function( next ){
      lines( utf8.end() ).forEach( decode, this );
      decode.call( this, soFar );
      next();
    }
  });
}

module.exports.createReadStream = createReadStream;
