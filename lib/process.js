
var os = require('os'),
    util = require('util'),
    path = require('path'),
    child = require('child_process');

function spawn( flags, isSync ){

  // select correct executable to use for this system
  var systemBin = util.format( 'pbf.%s-%s', os.platform(), os.arch() );
  var exec = path.join(__dirname, '/../build/', systemBin );

  // spawn child process
  var proc;

  if( isSync ){
    // synchronous (blocking) process
    proc = child.spawnSync( exec, flags );
  }
  else {
    // async (non-blocking) process
    proc = child.spawn( exec, flags );

    // propagate signals from parent to child
    var listener =  function(){ proc.kill(); };
    process.on('SIGINT',  listener);
    process.on('SIGTERM', listener);

    // clean up listener
    proc.on( 'exit', function(){
      process.removeListener('SIGINT',  listener);
      process.removeListener('SIGTERM', listener);
    });
  }

  return proc;
}

module.exports = spawn;
