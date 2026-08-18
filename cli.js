#!/usr/bin/env node

var os = require('os'),
    fs = require('fs'),
    util = require('util'),
    child = require('child_process'),
    exec = require('./lib/binaryPath');

if( !fs.existsSync( exec ) ){
  console.error( util.format(
    'pbf2json: no binary available for %s-%s, expected: %s',
    os.platform(), os.arch(), exec
  ));
  console.error( 'working from a git clone? run \'npm run compile\' to build it.' );
  process.exit( 1 );
}

var proc = child.spawn( exec, process.argv.slice( 2 ), { stdio: 'inherit' });

// propagate signals from parent to child
process.on( 'SIGINT',  function(){ proc.kill( 'SIGINT' ); });
process.on( 'SIGTERM', function(){ proc.kill( 'SIGTERM' ); });

proc.on( 'error', function( err ){
  console.error( util.format( 'pbf2json: failed to run %s:', exec ), err.message );
  process.exit( 1 );
});

proc.on( 'exit', function( code, signal ){
  // mimic the shell convention of reporting a signalled exit as 128+signum
  process.exit( signal ? 128 + os.constants.signals[signal] : code );
});
