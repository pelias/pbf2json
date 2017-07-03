
var util = require('util'),
    split = require('split2'),
    through = require('through2'),
    parser = require('./parser');

function createJsonStream( proc, done ){

  // create a json parser stream
  var decoder = parser();

  // pipe process stdout
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

  var exitCode = 0;

  // detect process exit code
  proc.on('exit', function(code){
    exitCode = code;
  });

  // call callback when we are done;
  proc.stdout.on( 'close', function(){
    if( 'function' === typeof done ){ done( exitCode ); }
  });

  // print error and exit on stderr
  proc.stderr.on( 'data', errorHandler( 'pbf' ) );

  // terminate the process and pipeline
  decoder.kill = function(){
    proc.kill();
    decoder.end();
  };

  return decoder;
}

function errorHandler( name ){
  return function( data ){
    data.toString('utf8').trim().split('\n').forEach( function( line ){
      console.log( util.format( '[%s]:', name ), line );
    });
  };
}

module.exports = createJsonStream;
