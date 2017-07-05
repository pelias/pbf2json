
var through = require('through2');
var options = {
  objectMode: true,
  highWaterMark: 256 // number of objects to keep in memory
};

// json parsing stream
function streamFactory(){
  return through( options, function( str, _, next ){
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

module.exports = streamFactory;
