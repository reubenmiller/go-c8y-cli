import React, { useEffect, useRef } from 'react';
import 'asciinema-player/dist/bundle/asciinema-player.css';

export default function AsciinemaPlayer({ src, ...opts }) {
  const ref = useRef(null);

  useEffect(() => {
    // Dynamic import keeps SSR happy (Docusaurus renders server-side)
    let player;
    import('asciinema-player').then((module) => {
      player = module.create(src, ref.current, opts);
    });
    return () => player?.dispose();
  }, [src]);

  return <div ref={ref} />;
}
