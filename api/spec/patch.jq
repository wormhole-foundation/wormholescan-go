# Fixes applied to the served Wormholescan Swagger 2.0 document before
# conversion and code generation. Each rule names the upstream defect it
# corrects; remove a rule when the served spec no longer needs it.

# Path parameter type by name. Everything not listed is a string.
def path_param_type:
  if . == "chain_id" or . == "chain" or . == "seq" or . == "sequence" then "integer" else "string" end;

# Timestamp fields that the server renders as RFC 3339. Guardian heartbeat
# timestamps are unix nanoseconds rendered as decimal strings and stay strings.
def is_rfc3339_field($def; $name):
  ($def | test("Heartbeat") | not)
  and ($name | test("At$") or $name == "timestamp" or $name == "releaseTime" or $name == "txTimestamp");

# 1. Fiber ":param" path syntax -> OpenAPI "{param}".
.paths |= with_entries(.key |= gsub(":(?<n>[A-Za-z_]+)"; "{\(.n)}"))

# 2. Path parameters the spec never declares (governor, observations, relays).
| .paths |= with_entries(
    . as $e
    | ($e.key | [match("\\{(?<n>[A-Za-z_]+)\\}"; "g").captures[0].string]) as $names
    | .value |= with_entries(
        .value.parameters |= ((. // []) + (($names - (. // [] | map(select(.in == "path") | .name))) | map({
          name: ., in: "path", required: true, type: (. | path_param_type)
        })))
      )
  )

# 3. Go []byte fields: swag renders them as arrays of integers; the server
#    sends base64 strings.
| .definitions |= with_entries(
    .value.properties |= (if . == null then . else with_entries(
      if .value.type == "array" and .value.items.type == "integer"
      then .value = {type: "string", format: "byte"} else . end
    ) end)
  )

# 4. RFC 3339 timestamps carry no format in the spec.
| .definitions |= with_entries(
    .key as $def
    | .value.properties |= (if . == null then . else with_entries(
      if .value.type == "string" and is_rfc3339_field($def; .key)
      then .value.format = "date-time" else . end
    ) end)
  )

# 5. Response shapes that differ from what the server sends.
#    /api/v1/operations returns {"operations": [...]}, not a bare array.
| .paths["/api/v1/operations"].get.responses["200"].schema = {
    type: "object",
    properties: {operations: {type: "array", items: {"$ref": "#/definitions/operations.OperationResponse"}}}
  }
#    /api/v1/governor/vaas returns a bare array, not the pagination envelope.
| .paths["/api/v1/governor/vaas"].get.responses["200"].schema = {
    type: "array", items: {"$ref": "#/definitions/governor.GovernorVaasResponse"}
  }
#    /api/v1/supply/{circulating,total} return a bare number.
| .paths["/api/v1/supply/circulating"].get.responses["200"].schema = {type: "number"}
| .paths["/api/v1/supply/total"].get.responses["200"].schema = {type: "number"}

# 6. No `produces` or `consumes` anywhere, so converters emit "*/*" for
#    responses and drop the one JSON request body.
| .produces = ["application/json"]
| .consumes = ["application/json"]

# 7. Fields the spec omits or mistypes (checked against live responses).
#    vaa.VaaDoc has no `sequence`; observations.ObservationDoc renders it as a
#    string but the server sends a number. (operations and governor responses
#    really do send it as a string.)
| .definitions["vaa.VaaDoc"].properties.sequence = {type: "integer", format: "int64"}
| .definitions["observations.ObservationDoc"].properties.sequence = {type: "integer", format: "int64"}

# 8. /api/v1/vaas/{chain_id}/{emitter}/{seq} returns {"data": VaaDoc}, not an
#    array envelope.
| .definitions["response.Response-vaa_VaaDoc"] = {
    type: "object", properties: {data: {"$ref": "#/definitions/vaa.VaaDoc"}}
  }
| .paths["/api/v1/vaas/{chain_id}/{emitter}/{seq}"].get.responses["200"].schema = {"$ref": "#/definitions/response.Response-vaa_VaaDoc"}

# 9. Error responses carry an undocumented body:
#    {"code": 5, "message": "NOT FOUND", "details": [{"request_id": "..."}]}.
| .definitions["response.Error"] = {
    type: "object",
    properties: {
      code: {type: "integer"},
      message: {type: "string"},
      details: {type: "array", items: {type: "object", additionalProperties: true}}
    }
  }
| .paths |= with_entries(.value |= with_entries(
    .value.responses |= with_entries(
      if (.key == "400" or .key == "404" or .key == "500") and .value.schema == null
      then .value.schema = {"$ref": "#/definitions/response.Error"} else . end
    )
  ))

# 10. /api/v1/global-tx/{chain_id}/{emitter}/{seq} returns
#     transactions.GlobalTransactionDoc ({id, originTx, destinationTx}),
#     not transactions.Tx (x-chain activity volume). Evidence: live
#     GET https://api.wormholescan.io/api/v1/global-tx/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/691371
#     returned {"id":"...","originTx":{...},"destinationTx":null}.
| .paths["/api/v1/global-tx/{chain_id}/{emitter}/{seq}"].get.responses["200"].schema = {
    "$ref": "#/definitions/transactions.GlobalTransactionDoc"
  }

# 20. /v1/signed_vaa and /v1/signed_batch_vaa declare vaaBytes as an integer
#     array; the live server sends a base64 string. Recorded
#     /v1/signed_vaa/1/19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe/755119
#     returns {"vaaBytes":"AQAAAAcN..."}. Rule 3 only rewrites .definitions.
| .paths["/v1/signed_vaa/{chain_id}/{emitter}/{seq}"].get.responses["200"].schema.properties.vaaBytes = {type: "string", format: "byte"}
| .paths["/v1/signed_batch_vaa/{chain_id}/{emitter}/sequence/{seq}"].get.responses["200"].schema.properties.vaaBytes = {type: "string", format: "byte"}

# 21. /v1/governor/is_vaa_enqueued returns {"isEnqueued": bool}, not
#     governor.EnqueuedVaaResponse. Recorded live: {"isEnqueued":false}.
| .definitions["guardian.IsVaaEnqueuedResponse"] = {
    type: "object",
    properties: {isEnqueued: {type: "boolean"}}
  }
| .paths["/v1/governor/is_vaa_enqueued/{chain_id}/{emitter}/{seq}"].get.responses["200"].schema = {"$ref": "#/definitions/guardian.IsVaaEnqueuedResponse"}

# 22. /v1/governor/token_list returns {"entries":[TokenList]}, not a bare array.
#     Recorded live response is an object with an entries array of
#     originChainId/originAddress/price.
| .definitions["guardian.TokenListResponse"] = {
    type: "object",
    properties: {
      entries: {type: "array", items: {"$ref": "#/definitions/governor.TokenList"}}
    }
  }
| .paths["/v1/governor/token_list"].get.responses["200"].schema = {"$ref": "#/definitions/guardian.TokenListResponse"}

# 23. heartbeats.HeartbeatNetworkResponse omits safeHeight and finalizedHeight.
#     Guardiand Heartbeat.Network sends them as decimal strings (same as height).
#     Wormholescan currently omits them (recorded /v1/heartbeats); keep the
#     fields so a guardian --publicWeb response still maps.
| .definitions["heartbeats.HeartbeatNetworkResponse"].properties += {
    safeHeight: {type: "string"},
    finalizedHeight: {type: "string"}
  }

# 24. heartbeats.RawHeartbeat omits p2pNodeId. Guardiand Heartbeat.p2pNodeId is
#     bytes (JSON base64). Wormholescan currently omits it (recorded
#     /v1/heartbeats); keep the field so a guardian --publicWeb response maps.
| .definitions["heartbeats.RawHeartbeat"].properties.p2pNodeId = {type: "string", format: "byte"}

# 40. /api/v1/vaas/{chain_id}/{emitter}/{seq}/duplicated returns
#     {"data":[DuplicateVaaDoc]} whose sequence is a string and which omits
#     isDuplicated, isSolanaShim, payload, and txHash.
#     Evidence: wormhole-explorer api/routes/wormscan/vaa/types.go
#     DuplicateVaaResponse and FindDuplicatedById in controller.go.
#     Live non-duplicate ids return 404; no 200 sample was available.
| .definitions["vaa.DuplicateVaaDoc"] = {
    type: "object",
    properties: {
      digest: {type: "string"},
      emitterAddr: {type: "string"},
      emitterChain: {"$ref": "#/definitions/vaa.ChainID"},
      emitterNativeAddr: {type: "string"},
      guardianSetIndex: {type: "integer"},
      id: {type: "string"},
      indexedAt: {type: "string", format: "date-time"},
      sequence: {type: "string"},
      timestamp: {type: "string", format: "date-time"},
      updatedAt: {type: "string", format: "date-time"},
      vaa: {type: "string", format: "byte"},
      version: {type: "integer"}
    }
  }
| .definitions["response.Response-array_vaa_DuplicateVaaDoc"] = {
    type: "object",
    properties: {
      data: {type: "array", items: {"$ref": "#/definitions/vaa.DuplicateVaaDoc"}},
      pagination: {"$ref": "#/definitions/response.ResponsePagination"}
    }
  }
| .paths["/api/v1/vaas/{chain_id}/{emitter}/{seq}/duplicated"].get.responses["200"].schema =
    {"$ref": "#/definitions/response.Response-array_vaa_DuplicateVaaDoc"}

# 41. find-observations-by-id returns one ObservationDoc, not an array.
#     Evidence: wormhole-explorer api/handlers/observations/service.go FindOne
#     returns *ObservationDoc; the controller JSON-encodes that single value.
| .paths["/api/v1/observations/{chain}/{emitter}/{sequence}/{signer}/{hash}"].get.responses["200"].schema =
    {"$ref": "#/definitions/observations.ObservationDoc"}
