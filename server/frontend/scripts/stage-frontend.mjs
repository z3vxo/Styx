import { cp, mkdir, readFile, rm } from 'node:fs/promises';

const source = new URL('../dist/', import.meta.url);
const destination = new URL('../../backend/internal/web/dist/', import.meta.url);
// Confirm the build exists before replacing only the generated embed directory.
await readFile(new URL('index.html', source));
await rm(destination, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });
console.log('Staged frontend for Go embedding.');
