/** Compiler-resolved checks for aggregate behaviour and command-store access. */
import ts from 'typescript';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const readMethods = new Set(['snapshot', 'events']);
export function commandLayout(file) {
  const names = file.statements.filter(node => ts.isClassDeclaration(node) || ts.isTypeAliasDeclaration(node) || ts.isInterfaceDeclaration(node))
    .map(node => node.name?.text ?? '');
  const commands = names.filter(name => name.endsWith('Command'));
  const handlers = names.filter(name => name.endsWith('CommandHandler'));
  if (!commands.length && !handlers.length) return [];
  if (path.dirname(file.fileName).replaceAll('\\', '/').endsWith('/application/commands') && commands.length === 1 &&
    handlers.length === 1 && handlers[0] === commands[0]+'Handler') return [];
  return [`${file.fileName}: keep one command and its handler together in application/commands`];
}
function enclosing(node, predicate) {
  for (let parent = node.parent; parent; parent = parent.parent) if (predicate(parent)) return parent;
}
function aggregateClass(node) {
  return ts.isClassDeclaration(node) && node.members.some(member => member.name?.getText() === 'snapshot') &&
    node.getSourceFile().fileName.replaceAll('\\', '/').includes('/contexts/') &&
    node.getSourceFile().fileName.replaceAll('\\', '/').includes('/domain/');
}
function commandExecution(node) {
  const method = enclosing(node, ts.isMethodDeclaration);
  const owner = enclosing(node, ts.isClassDeclaration);
  return owner?.name?.text.endsWith('CommandHandler') && method?.name?.getText() === 'execute' &&
    owner.getSourceFile().fileName.replaceAll('\\', '/').includes('/application/');
}
export function violations(program, selected) {
  const checker = program.getTypeChecker(), errors = [];
  function symbolAt(node) {
    let symbol = checker.getSymbolAtLocation(node);
    if (symbol && symbol.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
    return symbol;
  }
  for (const file of program.getSourceFiles()) {
    if (!selected(file.fileName)) continue;
    const stores = new Map();
    const report = (node, message) => errors.push(`${file.fileName}:${file.getLineAndCharacterOfPosition(node.getStart()).line+1}: ${message}`);
    function visit(node) {
      let forbidden = false;
      if (ts.isNewExpression(node)) {
        const declarations = symbolAt(node.expression)?.declarations ?? [];
        const restoring = file.fileName.replaceAll('\\', '/').endsWith('/adaptors/restore.ts');
        forbidden = declarations.some(aggregateClass) && !restoring;
      }
      if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
        const key = ts.isPropertyAccessExpression(node) ? node.name : node.argumentExpression;
        const name = ts.isStringLiteral(key) ? key.text : key.getText();
        let symbol = symbolAt(key);
        if (!symbol && ts.isElementAccessExpression(node) && ts.isStringLiteral(key))
          symbol = checker.getPropertyOfType(checker.getTypeAtLocation(node.expression), key.text);
        const declarations = symbol?.declarations ?? [];
        const core = file.fileName.replaceAll('\\', '/').includes('/application/');
        if (core && declarations.some(d => ts.isClassDeclaration(d.parent) && d.parent.name?.text.endsWith('CommandHandler')))
          report(node, 'application handlers must use durable commands; direct command-handler capabilities are forbidden');
        const store = declarations.some(d => (ts.isInterfaceDeclaration(d.parent) || ts.isClassDeclaration(d.parent)) &&
          ['AggregateCommandPort', 'PostgresAggregateCommandStore'].includes(d.parent.name?.text) && name === 'execute');
        if (store && commandExecution(node)) {
          const method = enclosing(node, ts.isMethodDeclaration);
          stores.set(method, (stores.get(method) ?? 0) + 1);
          let repeated = false;
          for (let p = node.parent; p && p !== method; p = p.parent)
            if (ts.isIterationStatement(p, false) || ts.isFunctionLike(p)) repeated = true;
          if (!ts.isCallExpression(node.parent) || node.parent.expression !== node || repeated)
            report(node, 'aggregate store execution must be one direct, non-repeated call');
        }
        forbidden ||= declarations.some(declaration => {
          const owner = declaration.parent;
          return (aggregateClass(owner) && !readMethods.has(name)) ||
            ((ts.isInterfaceDeclaration(owner) || ts.isClassDeclaration(owner)) &&
              ['AggregateCommandPort', 'PostgresAggregateCommandStore'].includes(owner.name?.text) && name === 'execute');
        });
      }
      if (forbidden && !commandExecution(node)) {
        const position = file.getLineAndCharacterOfPosition(node.getStart());
        errors.push(`${file.fileName}:${position.line+1}: aggregate mutation capability outside CommandHandler.execute`);
      }
      ts.forEachChild(node, visit);
    }
    visit(file);
    for (const [method, count] of stores) if (count > 1) report(method, 'a command handler may invoke the aggregate store once');
  }
  return errors;
}
export function productionProgram() {
  const config = ts.readConfigFile(path.join(root, 'services/engagement/tsconfig.json'), ts.sys.readFile);
  const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.join(root, 'services/engagement'));
  return ts.createProgram(parsed.fileNames, parsed.options);
}
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const program = productionProgram();
  const selected = name => name.startsWith(path.join(root, 'services/engagement/src')) &&
    !/\/(generated|domain)\/|\.(test|integration)\.ts$/.test(name);
  const errors = violations(program, selected);
  for (const file of program.getSourceFiles())
    if (selected(file.fileName) && file.fileName.includes('/application/')) errors.push(...commandLayout(file));
  if (errors.length) { console.error(errors.join('\n')); process.exitCode = 1; }
  else console.info('TypeScript command mutation boundaries verified');
}
