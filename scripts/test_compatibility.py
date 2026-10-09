import copy
import unittest
from compatibility import proto_changes
from http_compatibility import changes


def document():
    return {'paths': {'/orders': {'post': {'operationId': 'create',
        'requestBody': {'content': {'application/json': {'schema': {'type': 'object', 'properties': {'name': {'type': 'string'}}}}}},
        'responses': {'200': {'content': {'application/json': {'schema': {'type': 'object', 'required': ['id'], 'properties': {'id': {'type': 'string'}}}}}}}}}}}


class CompatibilityTests(unittest.TestCase):
    def test_required_request_addition_fails_but_optional_addition_passes(self):
        old, new = document(), document()
        schema = new['paths']['/orders']['post']['requestBody']['content']['application/json']['schema']
        schema['properties']['note'] = {'type': 'string'}
        self.assertEqual(changes(old, new), [])
        schema['required'] = ['note']
        self.assertTrue(changes(old, new))

    def test_removed_response_guarantee_and_operation_fail(self):
        old, new = document(), document()
        new['paths']['/orders']['post']['responses']['200']['content']['application/json']['schema']['required'] = []
        self.assertTrue(changes(old, new))
        self.assertTrue(changes(old, {'paths': {}}))

    def test_request_and_response_types_and_required_headers_fail(self):
        for category, entry in [('requestBody', None), ('responses', '200')]:
            old, new = document(), document()
            body = new['paths']['/orders']['post'][category]
            if entry: body = body[entry]
            body['content']['application/json']['schema']['type'] = 'array'
            self.assertTrue(changes(old, new))
        old, new = document(), document()
        new['paths']['/orders']['post']['parameters'] = [{'in': 'header', 'name': 'x-new', 'required': True, 'schema': {'type': 'string'}}]
        self.assertTrue(changes(old, new))

    def test_protobuf_field_reuse_removal_and_new_required_semantics_fail(self):
        old = {'cafe.Example': {'fields': {'1': {'name': 'id', 'options': ''}}}}
        for new in ({}, {'cafe.Example': {'fields': {}}}, {'cafe.Example': {'fields': {'1': {'name': 'other', 'options': ''}}}}):
            self.assertTrue(proto_changes(old, new))
        new = copy.deepcopy(old)
        new['cafe.Example']['fields']['2'] = {'name': 'optional', 'options': ''}
        self.assertEqual(proto_changes(old, new), [])
        new['cafe.Example']['fields']['2']['options'] = 'required'
        self.assertTrue(proto_changes(old, new))
