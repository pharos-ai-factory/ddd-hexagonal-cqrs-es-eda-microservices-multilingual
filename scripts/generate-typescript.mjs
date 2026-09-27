import protobuf from 'protobufjs';
import {readFile, writeFile, mkdir} from 'node:fs/promises';
for (const [name, source] of [
  ['events', 'contracts/events/proto/cafe/v1/events.proto'],
  ['realtime', 'contracts/realtime/proto/cafe/realtime/v1/realtime.proto'],
]) {
  const root = protobuf.parse(await readFile(source, 'utf8')).root.toJSON();
  for (const service of name === 'events' ? ['engagement'] : ['engagement', 'web']) {
    const directory = `services/${service}/src/adaptors/generated`;
    await mkdir(directory, {recursive: true});
    await writeFile(`${directory}/${name}.json`, `${JSON.stringify(root, null, 2)}\n`);
  }
}
await writeFile('services/engagement/src/adaptors/generated/catalogue.json',
  await readFile('contracts/events/catalogue.json'));
