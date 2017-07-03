
var through = require('through2');

// json parsing stream
function streamFactory(){
  return through.obj( function( str, _, next ){
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
