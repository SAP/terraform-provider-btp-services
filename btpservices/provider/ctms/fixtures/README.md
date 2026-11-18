# CTMS acceptance recordings

No cassettes have been fabricated or recorded. Tests are located beside their implementations in nodes, routes, destinations,
and ui. The recorder lives in provider/tfutils, with a CTMS setup wrapper in testutils. The tests use the
same go-vcr and Terraform plugin-testing libraries as CI/CD. They skip when their
cassette is absent unless TEST_FORCE_REC=true explicitly enables recording.
Replay mode never falls back to a live HTTP transport.

Each data source has a matching test file and an independent cassette named after
the implementation file, such as datasource_node and datasource_nodes. Missing-object
cases stay in the singular test file and use a separate _missing cassette. The UI URL
and transport-request logs URL tests also have separate files and cassettes.

## Recording prerequisites

Use a disposable CTMS landscape and credentials allowed to create, update, and
delete nodes and routes. Tests create nodes/routes prefixed tf-ctms- and delete
them during cleanup. Do not attach unrelated routes, schedules, or transport
requests to these objects. Do not run overlapping recording jobs against the same
landscape because the test names are fixed for reproducible replay.

Set these environment variables:

- BTP_CTMS_ENDPOINT: service root without /v2.
- BTP_CTMS_TOKEN_URL: complete OAuth token endpoint.
- BTP_CTMS_CLIENT_ID and BTP_CTMS_CLIENT_SECRET.
- BTP_CTMS_TEST_DESTINATION: an existing MTA-to-Cloud-Foundry destination for
  target strategy and destination data-source tests.
- BTP_CTMS_TEST_TRANSPORT_REQUEST_ID and BTP_CTMS_TEST_NODE_ID: existing identifiers
  for UI log-link tests. These tests only retrieve links.

Use Terraform 1.14 or later. Run from the repository root when ready:

~~~sh
TEST_FORCE_REC=true go test ./btpservices/provider/ctms/... -run '^TestResourceCTMSNode$' -count=1
TEST_FORCE_REC=true go test ./btpservices/provider/ctms/... -count=1
~~~

The first command isolates basic node CRUD and imports. The second records all
CTMS acceptance cases. Do not use TEST_FORCE_REC when replaying saved cassettes:

~~~sh
go test ./btpservices/provider/ctms/... -count=1
~~~

## Coverage and acceptance gates

- Minimal node defaults, explicit boolean/text resets, numeric and structured imports.
- MTA strategy updates and clearing the entire target collection.
- Virtual nodes and automatic imports.
- External node drift, relinquished ownership, and external deletion/recreation.
- Route rename, description reset, and node replacement with route rewiring.
- Node/route lookups by ID and name and plural data sources.
- Destination summaries, expanded lookup, and tenant-ID requests.
- Tenant UI and transport-request log links, with and without node context.
- Missing named objects.
- Node/route discovery, filters, full results, scoped identities, and configuration generation.

These tests intentionally make assertions about reset and ordering behavior that
must be checked against the real API. A failing recording is evidence to review
the implementation/design, not a reason to edit a cassette into a passing result.
Additional service evidence is still required for multiple-target cardinality and
ordering, imported schedules, notification prerequisites, and server-side failure
scenarios. Existing client tests already cover malformed responses, HTTP error
classification, transport cancellation, and uncertain-create retry behavior.

## Recording integrity

CTMS configures the shared recorder to match method, URL, and request body.
Shared JSON processing preserves numeric IDs without converting them through float64.
CI/CD keeps its existing default matching and redaction rules. Tokens, authentication/session headers,
OAuth form values, and connection hosts are removed or normalized; destination
and UI URLs are replaced with safe placeholders. Test-input destination names
and UI query IDs are normalized structurally, preserving unrelated numeric IDs.
Review every cassette for tenant/business metadata before staging it. Recordings
include cleanup requests. Provider/client tests and lint pass after the CTMS configuration
simplification. CTMS acceptance cases skip while cassettes are absent; no live recordings
have been made.
