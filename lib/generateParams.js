const util = require('util');

function generateParams(config) {
  const flags = [];

  if (config.tags) {
    const tags = config.tags.join(',');
    flags.push( util.format( '-tags=%s', tags ) );
  }
  if( config.hasOwnProperty( 'leveldb' ) ){
    flags.push( util.format( '-leveldb=%s', config.leveldb ) );
  }
  flags.push( config.file );

  return flags;
}

module.exports = generateParams;
