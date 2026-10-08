"""Service ownership and Go interfaces over native context envelopes."""
from contract_sources import OWNERS

OWNED = {'storefront': ('menu', 'ordering'), 'operations': ('preparation', 'collection'),
         'engagement': ('loyalty', 'communication'), 'api': OWNERS}


def envelopes(owners, package_root):
    lines = ['// Code generated from context ownership. DO NOT EDIT.', 'package requestsv1', 'import (',
             '"google.golang.org/protobuf/proto"', '"fmt"', f'shared "{package_root}/cafe/requests/v1/shared"']
    lines += [f'{owner} "{package_root}/cafe/requests/v1/contexts/{owner}"' for owner in owners]
    lines += [')', 'type Request interface { proto.Message; GetContractVersion() uint32; GetRequestId() string; GetContext() string }',
              'type Reply interface { Request; GetOutcome() *shared.Outcome; GetError() *shared.RequestError }',
              'type CommandEnvelope interface { proto.Message; GetMetadata() *shared.CommandMetadata }']
    for name in ('PageRequest', 'CommandMetadata', 'Rejection', 'Outcome', 'RequestError'):
        lines.append(f'type {name} = shared.{name}')
    lines.append('var E_RequiredInput = shared.E_RequiredInput')
    for name in ('Request', 'Reply'):
        lines.append(f'func New{name}(owner,id string) ({name},error) {{ switch owner {{')
        lines += [f'case "{owner}": return &{owner}.{name}{{ContractVersion:1,RequestId:id,Context:owner}},nil' for owner in owners]
        lines.append('}; return nil,fmt.Errorf("unknown request owner: %s",owner) }')
    for kind in ('Command', 'Query'):
        result = 'CommandEnvelope' if kind == 'Command' else 'proto.Message'
        lines.append(f'func {kind}(request Request) {result} {{ switch value:=request.(type) {{')
        lines += [f'case *{owner}.Request: if value.Get{kind}()!=nil {{ return value.Get{kind}() }}' for owner in owners if kind != 'Command' or owner != 'communication']
        lines.append('}; return nil }')
    return '\n'.join(lines)+'\n'
