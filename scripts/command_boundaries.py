"""AST checks for Python aggregate capabilities and DI isolation."""
import ast
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
READ_METHODS = {'restore', 'snapshot', 'events'}


def aggregate_methods(root=ROOT):
    result = {}
    for path in (root/'services/operations/src/operations/contexts').glob('*/domain.py'):
        for node in ast.parse(path.read_text()).body:
            if isinstance(node, ast.ClassDef):
                methods = {item.name for item in node.body if isinstance(item, ast.FunctionDef)}
                if 'snapshot' in methods:
                    result[node.name] = methods - READ_METHODS - {'__init__'}
    return result


def violations(source, aggregates, path='contexts/example/application.py'):
    tree = ast.parse(source)
    parents = {child: parent for parent in ast.walk(tree) for child in ast.iter_child_nodes(parent)}
    symbols = {}
    ports = {'AggregateCommandPort', 'PostgresAggregateCommandStore'}
    for node in ast.walk(tree):
        if isinstance(node, ast.ClassDef) and node.name.endswith('CommandHandler'):
            symbols[node.name] = '@handler'
    for node in ast.walk(tree):
        if isinstance(node, ast.ImportFrom):
            for item in node.names:
                if item.name in aggregates and (node.module or '').endswith('.domain'):
                    symbols[item.asname or item.name] = item.name
                if item.name in {'AggregateCommandPort', 'PostgresAggregateCommandStore'}:
                    ports.add(item.asname or item.name)
                if item.name.endswith('CommandHandler'):
                    symbols[item.asname or item.name] = '@handler'

    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for item in node.names:
                if item.name.endswith('.domain'):
                    for name in aggregates:
                        symbols[(item.asname or item.name)+'.'+name] = name
        if isinstance(node, ast.ImportFrom):
            for item in node.names:
                if item.name == 'domain':
                    for name in aggregates:
                        symbols[(item.asname or item.name)+'.'+name] = name

    def resolve(node):
        if isinstance(node, ast.Subscript):
            return resolve(node.value)
        if isinstance(node, ast.Name) and node.id in ports:
            return '@command-store'
        if isinstance(node, ast.Constant) and isinstance(node.value, str):
            return symbols.get(node.value)
        if isinstance(node, ast.Call):
            return resolve(node.func)
        if isinstance(node, ast.Attribute) and node.attr in READ_METHODS | {'open'}:
            return resolve(node.value)
        return symbols.get(ast.unparse(node))

    # Follow assigned/annotated aggregate aliases, including self attributes.
    for _ in range(sum(1 for _ in ast.walk(tree))):
        previous = dict(symbols)
        for node in ast.walk(tree):
            if isinstance(node, ast.Assign):
                identity = resolve(node.value)
                if isinstance(node.value, ast.Tuple):
                    for target in node.targets:
                        if isinstance(target, ast.Tuple):
                            for left, right in zip(target.elts, node.value.elts):
                                if value := resolve(right):
                                    symbols[ast.unparse(left)] = value
                if identity:
                    for target in node.targets:
                        symbols[ast.unparse(target)] = identity
            if isinstance(node, (ast.arg, ast.AnnAssign)):
                annotation = node.annotation
                if annotation and (identity := resolve(annotation)):
                    symbols[node.arg if isinstance(node, ast.arg) else ast.unparse(node.target)] = identity
        if symbols == previous:
            break

    def allowed(node):
        method, owner = None, None
        parent = parents.get(node)
        while parent:
            if isinstance(parent, (ast.FunctionDef, ast.AsyncFunctionDef)):
                # Nested transaction closures remain inside the execute method.
                method = parent.name
            if isinstance(parent, ast.ClassDef):
                owner = parent.name
                break
            parent = parents.get(parent)
        return owner and owner.endswith('CommandHandler') and method == 'execute' and '/application' in '/'+path

    errors = []
    stores = {}
    for node in ast.walk(tree):
        mutation = False
        if isinstance(node, ast.Attribute):
            identity = resolve(node.value)
            if identity == '@handler' and '/application' in '/'+path:
                errors.append(f'{path}:{node.lineno}: direct command-handler capability in application')
            if identity == '@command-store' and node.attr == 'execute' and allowed(node):
                parent, repeated = parents.get(node), False
                if not isinstance(parent, ast.Call) or parent.func is not node:
                    repeated = True
                while parent:
                    if isinstance(parent, (ast.For, ast.AsyncFor, ast.While, ast.Lambda, ast.comprehension)):
                        repeated = True
                    if isinstance(parent, (ast.FunctionDef, ast.AsyncFunctionDef)):
                        if parent.name == 'execute' and isinstance(parents.get(parent), ast.ClassDef):
                            stores[parent] = stores.get(parent, 0) + 1
                            break
                        repeated = True
                    parent = parents.get(parent)
                if repeated:
                    errors.append(f'{path}:{node.lineno}: aggregate store execution must be one direct, non-repeated call')
            if identity in aggregates and node.attr.startswith('_'):
                errors.append(f'{path}:{node.lineno}: aggregate internals accessed outside its domain')
                continue
            mutation = (identity in aggregates and node.attr in aggregates[identity]) or (identity == '@command-store' and node.attr == 'execute')
        if isinstance(node, ast.Call):
            mutation |= symbols.get(ast.unparse(node.func)) in aggregates
        # Restrict the write-port capability to command handler classes. This also
        # rejects indirect calls through a stored execute alias in event handlers.
        if isinstance(node, ast.Name) and node.id in ports and '/application' in '/'+path:
            parent = parents.get(node)
            while parent and not isinstance(parent, ast.ClassDef):
                parent = parents.get(parent)
            mutation |= parent is None or not parent.name.endswith('CommandHandler')
            if mutation:
                errors.append(f'{path}:{node.lineno}: aggregate write port outside command handler')
                continue
        if mutation and not allowed(node):
            errors.append(f'{path}:{node.lineno}: aggregate mutation outside CommandHandler.execute')
    for method, count in stores.items():
        if count > 1:
            errors.append(f'{path}:{method.lineno}: a command handler may invoke the aggregate store once')
    return errors


def check(root=ROOT):
    errors = []
    aggregates = aggregate_methods(root)
    for path in (root/'services/operations/src/operations').rglob('*.py'):
        if 'generated' in path.parts or 'foundation' in path.parts or path.name == 'domain.py':
            continue
        errors.extend(violations(path.read_text(), aggregates, path.relative_to(root).as_posix()))
    return errors
