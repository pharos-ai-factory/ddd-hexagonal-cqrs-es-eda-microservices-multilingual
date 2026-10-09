"""Conservative directional compatibility for the repository's OpenAPI subset."""
METHODS = {'get', 'post', 'put', 'patch', 'delete', 'head', 'options'}
ANNOTATIONS = {'description', 'title', 'example', 'examples', 'deprecated', 'summary', 'externalDocs'}


def dereference(value, document):
    if isinstance(value, list):
        return [dereference(v, document) for v in value]
    if not isinstance(value, dict):
        return value
    if '$ref' in value:
        target = document
        for part in value['$ref'].removeprefix('#/').split('/'):
            target = target[part.replace('~1', '/').replace('~0', '~')]
        return dereference(target, document)
    return {k: dereference(v, document) for k, v in value.items() if k not in ANNOTATIONS}


def schema_changes(old, new, response=False, path='schema'):
    errors = []
    # Requests may widen accepted values; responses may narrow their possible values.
    narrow, wide = (new, old) if response else (old, new)
    for key in ('type', 'format', 'nullable', 'additionalProperties', 'oneOf', 'anyOf', 'allOf', 'not', 'discriminator', 'pattern'):
        if old.get(key) != new.get(key):
            errors.append(path+': changed '+key)
    if 'enum' in wide and ('enum' not in narrow or not set(narrow['enum']) <= set(wide['enum'])):
        errors.append(path+': incompatible enum')
    for key in ('minimum', 'minLength', 'minItems', 'minProperties'):
        if wide.get(key, float('-inf')) > narrow.get(key, float('-inf')):
            errors.append(path+': stricter '+key)
    for key in ('maximum', 'maxLength', 'maxItems', 'maxProperties'):
        if wide.get(key, float('inf')) < narrow.get(key, float('inf')):
            errors.append(path+': stricter '+key)
    if not set(wide.get('required', [])) <= set(narrow.get('required', [])):
        errors.append(path+': incompatible required properties')
    for name, child in old.get('properties', {}).items():
        if name not in new.get('properties', {}):
            errors.append(path+': removed property '+name)
        else:
            errors += schema_changes(child, new['properties'][name], response, path+'.'+name)
    if 'items' in old:
        errors += schema_changes(old['items'], new.get('items', {}), response, path+'[]')
    handled = ANNOTATIONS | {'type', 'format', 'nullable', 'additionalProperties', 'oneOf', 'anyOf', 'allOf', 'not',
        'discriminator', 'pattern', 'enum', 'required', 'properties', 'items', 'minimum', 'maximum', 'minLength',
        'maxLength', 'minItems', 'maxItems', 'minProperties', 'maxProperties'}
    for key in old.keys() | new.keys():
        if key not in handled and old.get(key) != new.get(key):
            errors.append(path+': review changed schema keyword '+key)
    return errors


def content_changes(old, new, response, path):
    errors = []
    for media, content in old.get('content', {}).items():
        if media not in new.get('content', {}):
            errors.append(path+': removed media '+media)
        else:
            errors += schema_changes(content.get('schema', {}), new['content'][media].get('schema', {}), response, path)
            if content.get('encoding') != new['content'][media].get('encoding'):
                errors.append(path+': changed media encoding')
    return errors


def changes(before, after):
    old, new = dereference(before, before), dereference(after, after)
    errors = []
    if old.get('components', {}).get('securitySchemes') != new.get('components', {}).get('securitySchemes'):
        errors.append('changed authentication schemes')
    for route, item in old['paths'].items():
        for method in METHODS & item.keys():
            path = method.upper()+' '+route
            a, b = item[method], new['paths'].get(route, {}).get(method)
            if b is None:
                errors.append(path+': removed operation')
                continue
            for key in ('operationId', 'security'):
                if a.get(key, old.get(key)) != b.get(key, new.get(key)):
                    errors.append(path+': changed '+key)
            parameters = lambda op, parent: {(p['in'], p['name']): p for p in parent.get('parameters', [])+op.get('parameters', [])}
            ap, bp = parameters(a, item), parameters(b, new['paths'][route])
            for key, p in ap.items():
                if key not in bp:
                    errors.append(path+': removed parameter '+str(key))
                else:
                    errors += schema_changes(p.get('schema', {}), bp[key].get('schema', {}), path=path+' '+str(key))
                    for option in ('style', 'explode', 'allowReserved', 'allowEmptyValue', 'content'):
                        if p.get(option) != bp[key].get(option):
                            errors.append(path+': changed parameter '+option)
                    if not p.get('required') and bp[key].get('required'):
                        errors.append(path+': parameter became required '+str(key))
            for key in bp.keys()-ap.keys():
                if bp[key].get('required'):
                    errors.append(path+': new required parameter '+str(key))
            ar, br = a.get('requestBody', {}), b.get('requestBody', {})
            if not ar.get('required') and br.get('required'):
                errors.append(path+': request body became required')
            errors += content_changes(ar, br, False, path+' request')
            for status, response in a.get('responses', {}).items():
                if status not in b.get('responses', {}):
                    errors.append(path+': removed response '+status)
                else:
                    errors += content_changes(response, b['responses'][status], True, path+' response '+status)
                    for header, schema in response.get('headers', {}).items():
                        updated = b['responses'][status].get('headers', {}).get(header)
                        if updated is None:
                            errors.append(path+': removed response header '+header)
                        else:
                            errors += schema_changes(schema.get('schema', {}), updated.get('schema', {}), True, path+' header '+header)
            for status in b.get('responses', {}).keys()-a.get('responses', {}).keys():
                if status.startswith('2'):
                    errors.append(path+': new success response '+status)
    return errors
