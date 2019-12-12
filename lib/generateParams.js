function generateParams(config) {
  const flags = [];

  if (config.tags) {
    const tags = config.tags.join(',');
    flags.push( `-tags=${tags}` );
  }
  if( config.hasOwnProperty( 'leveldb' ) ){
    flags.push( `-leveldb=${config.leveldb}` );
  }
  if( config.hasOwnProperty( 'waynodes' ) ){
    flags.push( `--waynodes=${config.waynodes}` );
  }

  flags.push( config.file );

  return flags;
}

module.exports = generateParams;
