// undici 8 (pulled in by jsdom) needs Node >= 22.19 and calls
// zlib.createZstdDecompress at load time; older Node 22 releases lack it, so
// every test worker crashes before running a test. Tests never receive
// zstd-encoded responses, so a stand-in is enough. Delete this file (and the
// execArgv entry in vitest.config.ts) once everyone runs Node >= 22.19.
import zlib from "node:zlib";

if (!zlib.createZstdDecompress) {
  zlib.createZstdDecompress = zlib.createGunzip;
}
