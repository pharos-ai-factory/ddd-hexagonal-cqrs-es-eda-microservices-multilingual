import protobuf from 'protobufjs';
import {writeFile, mkdir} from 'node:fs/promises';
import {execFileSync} from 'node:child_process';
import path from 'node:path';
const groups = JSON.parse(execFileSync('python3', ['scripts/contract_sources.py'], {encoding: 'utf8'}));
for (const [name, group] of Object.entries(groups)) {
  if (name === 'menu_private') continue;
  const schema = new protobuf.Root();
  const definitions = name === 'requests' ? Object.entries(group.sources).filter(([logical]) => !logical.includes('/contexts/') || group.owners.engagement.includes(logical.split('/contexts/')[1].split('/')[0])) : Object.entries(group.sources);
  const files = Object.fromEntries(definitions.map(([logical, physical]) => [logical, path.resolve(physical)]));
  const physical = new Set(Object.values(files));
  schema.resolvePath = (_origin, target) => physical.has(target) ? target : (files[target] ?? path.resolve("node_modules/protobufjs", target));
  // Synchronous imports keep descriptor definition order reproducible.
  schema.loadSync(name === 'requests' ? Object.values(files) : files[group.entrypoint]);
  schema.resolveAll();
  const root = schema.toJSON();
  const outputs = name.endsWith('_private')
    ? [`services/engagement/src/contexts/${name.replace('_private', '')}/adaptors/messaging/generated/private_messages.json`]
    : (name === 'realtime' ? ['engagement', 'web'] : ['engagement']).map(service => `services/${service}/src/adaptors/generated/${name}.json`);
  for (const output of outputs) {
    await mkdir(path.dirname(output), {recursive: true});
    await writeFile(output, `${JSON.stringify(root, null, 2)}\n`);
  }
}

// Wire interfaces make owner ACL mappings fail compilation when published fields change.
const requestSchema = protobuf.Root.fromJSON(JSON.parse(await (await import('node:fs/promises')).readFile('services/engagement/src/adaptors/generated/requests.json', 'utf8')));
requestSchema.resolveAll();
const interfaces = ['// Code generated from Protobuf request contracts. DO NOT EDIT.'];
const scalar = type => ['string','bytes'].includes(type) ? (type === 'bytes' ? 'Uint8Array' : 'string') : type === 'bool' ? 'boolean' : 'number';
for (const owner of ['shared', 'loyalty', 'communication']) {
  const namespace = requestSchema.lookup(owner === 'shared' ? 'cafe.requests.v1' : 'cafe.'+owner+'.requests.v1');
  interfaces.push(`export namespace ${owner} {`);
  for (const type of namespace.nestedArray) {
    if (!(type instanceof protobuf.Type) || ['Request','Reply','Command','Query'].includes(type.name)) continue;
    interfaces.push(`export interface ${type.name} {`);
    for (const field of type.fieldsArray) {
      const optional = field.options?.proto3_optional || (field.resolvedType instanceof protobuf.Type && !field.repeated);
      const value = field.resolvedType instanceof protobuf.Type
        ? (field.resolvedType.parent === namespace ? '' : 'shared.')+field.resolvedType.name : scalar(field.type);
      interfaces.push(`  ${field.name}${optional ? '?' : ''}: ${value}${field.repeated ? '[]' : ''};`);
    }
    interfaces.push('}');
  }
  interfaces.push('}');
}
await writeFile('services/engagement/src/adaptors/generated/request-types.ts', interfaces.join('\n')+'\n');
