import protobuf from 'protobufjs';
import 'protobufjs/ext/descriptor/index.js';
import { readFileSync, writeFileSync } from 'node:fs';
const root = protobuf.Root.fromDescriptor(readFileSync(process.argv[2]));
writeFileSync(process.argv[3], JSON.stringify(root.toJSON(), null, 2) + '\n');
