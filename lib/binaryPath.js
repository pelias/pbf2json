
var util = require('util'),
    path = require('path'),
    os = require('os');

// the precompiled binary shipped in ./build for the current platform, note that
// it may not exist: the npm package ships every target but a git clone only
// contains whatever ./compile.sh has been run for.
module.exports = path.join(
  __dirname, '..', 'build',
  util.format( 'pbf2json.%s-%s', os.platform(), os.arch() )
);
