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
