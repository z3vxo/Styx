import React from 'react';
import threshold from './styx-threshold.svg?raw';

export function Logo() {
  return <span className="brand-mark" aria-hidden="true" dangerouslySetInnerHTML={{ __html: threshold }} />;
}
