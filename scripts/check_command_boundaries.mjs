/** Compiler-resolved checks for aggregate behaviour and command-store access. */
import ts from 'typescript';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const readMethods = new Set(['snapshot', 'events']);
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
  }
  return errors;
}
export function productionProgram() {
  const config = ts.readConfigFile(path.join(root, 'services/engagement/tsconfig.json'), ts.sys.readFile);
  const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.join(root, 'services/engagement'));
  return ts.createProgram(parsed.fileNames, parsed.options);
}
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const errors = violations(productionProgram(), name => name.startsWith(path.join(root, 'services/engagement/src')) &&
    !/\/(generated|domain)\/|\.(test|integration)\.ts$/.test(name));
  if (errors.length) { console.error(errors.join('\n')); process.exitCode = 1; }
  else console.info('TypeScript command mutation boundaries verified');
}
