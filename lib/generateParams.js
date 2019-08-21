const util = require('util');

function generateParams(config) {
  const flags = [];

  flags.push( util.format( '-tags=%s', config.tags ) );
  if( config.hasOwnProperty( 'leveldb' ) ){
    flags.push( util.format( '-leveldb=%s', config.leveldb ) );
  }
  flags.push( config.file );

  return flags;
}

module.exports = generateParams;
