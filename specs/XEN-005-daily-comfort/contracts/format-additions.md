# Contract: workspace format additions (stage 5, still `xenon: 1`)

Fills the auth sub-objects reserved in XEN-001 format §6.2. Field names equal Postman's v2.1
attribute keys so import/export maps 1:1 (XEN-003 postman-mapping "Auth").

```yaml
auth:
  type: apikey
  apikey: { key: X-Api-Key, value: "{{api_key}}", in: header }      # header | query

auth:
  type: oauth2
  oauth2:
    grant_type: authorization_code_with_pkce   # authorization_code | authorization_code_with_pkce | implicit | password_credentials | client_credentials
    tokenName: "Acme SSO"
    redirect_uri: "http://127.0.0.1:53682/callback"
    authUrl: "https://sso.acme.test/oauth2/authorize"
    accessTokenUrl: "https://sso.acme.test/oauth2/token"
    refreshTokenUrl: ""                        # empty → accessTokenUrl
    clientId: "xenon-desktop"
    clientSecret: "{{sso_client_secret}}"
    username: ""                               # password grant
    password: ""
    scope: "openid profile api"
    state: ""                                  # empty → random per request
    challengeAlgorithm: S256                   # S256 | plain
    code_verifier: ""                          # empty → generated
    client_authentication: header              # header | body
    addTokenTo: header                         # header | queryParams
    headerPrefix: Bearer
    useToken: access_token                     # access_token | id_token  (Xenon field)
    autoRefresh: true                          # Xenon field
    authRequestParams: []                      # [{ key, value, enabled, send_as: request_url }]
    tokenRequestParams: []                     # send_as: request_body | request_url | request_header
    refreshRequestParams: []
    audience: ""
    resource: ""

auth:
  type: awsv4
  awsv4: { accessKey: "{{aws_key}}", secretKey: "{{aws_secret}}", sessionToken: "",
           region: us-east-1, service: execute-api, addAuthDataToQuery: false }

auth:
  type: digest
  digest: { username: "{{user}}", password: "{{pass}}", algorithm: MD5, qop: "",
            realm: "", nonce: "", opaque: "", nonceCount: "", clientNonce: "",
            disableRetryRequest: false }
```

Rules:
- **Never in files**: OAuth access/refresh/ID tokens (vault), Digest server state learned at
  runtime (`realm`/`nonce`/`opaque`/`nonceCount`/`clientNonce` are only written if the user typed
  them), cookies.
- Postman's `accessToken`/`tokenType` keys on import: a non-empty literal token is moved to the
  vault (not kept in the file); `{{var}}` references are kept as typed.
- Request settings gain `disableCookieJar: false`.
