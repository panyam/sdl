# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: sdl/v1/models.proto
# Protobuf Python Version: 6.31.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    6,
    31,
    1,
    '',
    'sdl/v1/models.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x13sdl/v1/models.proto\x12\x06sdl.v1\x1a\x1fgoogle/protobuf/timestamp.proto\"e\n\nPagination\x12\x19\n\x08page_key\x18\x01 \x01(\tR\x07pageKey\x12\x1f\n\x0bpage_offset\x18\x02 \x01(\x05R\npageOffset\x12\x1b\n\tpage_size\x18\x03 \x01(\x05R\x08pageSize\"\xa2\x01\n\x12PaginationResponse\x12\"\n\rnext_page_key\x18\x02 \x01(\tR\x0bnextPageKey\x12(\n\x10next_page_offset\x18\x03 \x01(\x05R\x0enextPageOffset\x12\x19\n\x08has_more\x18\x04 \x01(\x08R\x07hasMore\x12#\n\rtotal_results\x18\x05 \x01(\x05R\x0ctotalResults\"\xb3\x02\n\x06\x43\x61nvas\x12\x39\n\ncreated_at\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x39\n\nupdated_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tupdatedAt\x12\x0e\n\x02id\x18\x03 \x01(\tR\x02id\x12#\n\ractive_system\x18\x04 \x01(\tR\x0c\x61\x63tiveSystem\x12!\n\x0cloaded_files\x18\x05 \x03(\tR\x0bloadedFiles\x12\x31\n\ngenerators\x18\x06 \x03(\x0b\x32\x11.sdl.v1.GeneratorR\ngenerators\x12(\n\x07metrics\x18\x07 \x03(\x0b\x32\x0e.sdl.v1.MetricR\x07metrics\"\xc2\x02\n\tGenerator\x12\x39\n\ncreated_at\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x39\n\nupdated_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tupdatedAt\x12\x0e\n\x02id\x18\x03 \x01(\tR\x02id\x12\x1b\n\tcanvas_id\x18\x04 \x01(\tR\x08\x63\x61nvasId\x12\x12\n\x04name\x18\x05 \x01(\tR\x04name\x12\x1c\n\tcomponent\x18\x06 \x01(\tR\tcomponent\x12\x16\n\x06method\x18\x07 \x01(\tR\x06method\x12\x12\n\x04rate\x18\x08 \x01(\x01R\x04rate\x12\x1a\n\x08\x64uration\x18\t \x01(\x01R\x08\x64uration\x12\x18\n\x07\x65nabled\x18\n \x01(\x08R\x07\x65nabled\"\xd0\x04\n\x06Metric\x12\x39\n\ncreated_at\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x39\n\nupdated_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tupdatedAt\x12\x0e\n\x02id\x18\x03 \x01(\tR\x02id\x12\x1b\n\tcanvas_id\x18\x04 \x01(\tR\x08\x63\x61nvasId\x12\x12\n\x04name\x18\x05 \x01(\tR\x04name\x12\x1c\n\tcomponent\x18\x06 \x01(\tR\tcomponent\x12\x18\n\x07methods\x18\x07 \x03(\tR\x07methods\x12\x18\n\x07\x65nabled\x18\x08 \x01(\x08R\x07\x65nabled\x12\x1f\n\x0bmetric_type\x18\t \x01(\tR\nmetricType\x12 \n\x0b\x61ggregation\x18\n \x01(\tR\x0b\x61ggregation\x12-\n\x12\x61ggregation_window\x18\x0b \x01(\x01R\x11\x61ggregationWindow\x12!\n\x0cmatch_result\x18\x0c \x01(\tR\x0bmatchResult\x12*\n\x11match_result_type\x18\r \x01(\tR\x0fmatchResultType\x12)\n\x10oldest_timestamp\x18\x0e \x01(\x01R\x0foldestTimestamp\x12)\n\x10newest_timestamp\x18\x0f \x01(\x01R\x0fnewestTimestamp\x12&\n\x0fnum_data_points\x18\x10 \x01(\x03R\rnumDataPointsBw\n\ncom.sdl.v1B\x0bModelsProtoP\x01Z#github.com/panyam/sdl/gen/go/sdl/v1\xa2\x02\x03SXX\xaa\x02\x06Sdl.V1\xca\x02\x06Sdl\\V1\xe2\x02\x12Sdl\\V1\\GPBMetadata\xea\x02\x07Sdl::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'sdl.v1.models_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'\n\ncom.sdl.v1B\013ModelsProtoP\001Z#github.com/panyam/sdl/gen/go/sdl/v1\242\002\003SXX\252\002\006Sdl.V1\312\002\006Sdl\\V1\342\002\022Sdl\\V1\\GPBMetadata\352\002\007Sdl::V1'
  _globals['_PAGINATION']._serialized_start=64
  _globals['_PAGINATION']._serialized_end=165
  _globals['_PAGINATIONRESPONSE']._serialized_start=168
  _globals['_PAGINATIONRESPONSE']._serialized_end=330
  _globals['_CANVAS']._serialized_start=333
  _globals['_CANVAS']._serialized_end=640
  _globals['_GENERATOR']._serialized_start=643
  _globals['_GENERATOR']._serialized_end=965
  _globals['_METRIC']._serialized_start=968
  _globals['_METRIC']._serialized_end=1560
# @@protoc_insertion_point(module_scope)
