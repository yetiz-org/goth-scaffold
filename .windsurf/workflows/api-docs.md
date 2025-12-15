---
description: Update OpenAPI documentation based on Handler and Route definitions
auto_execution_mode: 3
---

# API Docs Workflow

Update OpenAPI documentation based on code facts, remove non-existent endpoints, add missing endpoints.

## GOLDEN RULES

1. **route.go is the source of truth**: OpenAPI must be consistent with endpoint definitions in `app/handlers/route.go`
2. **Remove non-existent endpoints**: Endpoints in OpenAPI that don't exist in code must be removed
3. **Add missing endpoints**: Endpoints in code that are missing from OpenAPI must be added
4. **No fabrication**: All schemas must be based on handler implementation
5. **operationId must be unique**: No duplicates across all yaml files
6. **Valid YAML format**: Ensure syntax is correct and parseable

---

## WORKFLOW

### Step 1: Read Route Definitions

// turbo

```bash
grep -n "SetEndpoint" app/handlers/route.go
```

### Step 2: List Endpoints in OpenAPI Files

// turbo

```bash
# List all paths defined in OpenAPI
grep -E "^  /api/|^  /mgmt/|^  /adm/" docs/openapi/*.yaml | sort
```

### Step 3: Compare and Identify Differences

#### 3.1 Find Extra Endpoints in OpenAPI

Compare route.go with OpenAPI files, find:
- Endpoints that exist in OpenAPI but not in route.go
- These endpoints must be removed from OpenAPI

#### 3.2 Find Missing Endpoints

Compare route.go with OpenAPI files, find:
- Endpoints that exist in route.go but not in OpenAPI
- These endpoints must be added to the corresponding OpenAPI file

### Step 4: Handler to OpenAPI Mapping Principles

| Handler Location | OpenAPI File | Description |
|-----------------|--------------|-------------|
| `endpoints/*.go` (root) | - | Page handlers, not APIs |
| `endpoints/v1/` | `openapi.yaml` | API endpoints |

**Note**: Can be split into multiple yaml files as needed (e.g., `auth.yaml`, `admin.yaml`, etc.)

### Step 5: Process Each Difference

#### 5.1 Remove Extra Endpoints

For each endpoint to remove:
1. Find the corresponding OpenAPI file
2. Remove the complete endpoint definition

#### 5.2 Add Missing Endpoints

For each endpoint to add:
1. Read the corresponding handler implementation
2. Confirm HTTP method, parameters, request/response schema
3. Add to the corresponding OpenAPI file

#### 5.3 Parameter Format Check

**Must compare each field type between handler struct and OpenAPI schema:**

| Handler Type | OpenAPI Format | Description |
|-------------|----------------|-------------|
| `*int64` (timestamp) | `type: integer, format: int64` | Unix timestamp |
| `*time.Time` | `type: string, format: date-time` | ISO 8601 format |
| `*string` | `type: string` + `nullable: true` | Nullable string |
| `string` | `type: string` | Required string |

### Step 6: Verification

// turbo

```bash
# 1. Check for duplicate operationId
grep -h 'operationId:' docs/openapi/*.yaml | sort | uniq -d

# 2. YAML syntax validation
for f in docs/openapi/*.yaml; do ruby -ryaml -e "YAML.safe_load(File.read('$f'))" && echo "$f OK"; done
```

---

## OpenAPI Endpoint Format

### Basic Structure

```yaml
  /api/v1/path/{param}:
    get:
      tags:
        - TagName
      summary: Brief description
      description: |-
        Detailed description
      operationId: uniqueOperationId  # Must be unique!
      security:
        - OAuth2: [ ]
      parameters:
        - name: param
          in: path
          required: true
          schema:
            type: string
          description: Parameter description
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  field:
                    type: string
        '401':
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

### operationId Naming Convention

| HTTP Method | Naming Pattern | Example |
|-------------|---------------|---------|
| GET (list) | `list{Resource}` | `listUsers` |
| GET (single) | `get{Resource}` | `getUser` |
| POST | `create{Resource}` | `createUser` |
| PATCH | `update{Resource}` | `updateUser` |
| DELETE | `delete{Resource}` | `deleteUser` |

---

## Parameter Format

### Path Parameter

```yaml
- name: id
  in: path
  required: true
  schema:
    type: string
  description: Resource ID
```

### Query Parameter

```yaml
- name: offset
  in: query
  schema:
    type: integer
    default: 0
  description: Pagination offset
```

### Enum Parameter

```yaml
- name: type
  in: query
  schema:
    type: string
    enum: [ type_a, type_b, type_c ]
  description: Resource type
```

---

## Request Body Format

### JSON Body

```yaml
requestBody:
  required: true
  content:
    application/json:
      schema:
        type: object
        required:
          - name
        properties:
          name:
            type: string
          type:
            type: string
            enum: [ admin, user ]
            default: user
```

### Multipart Form (File Upload)

```yaml
requestBody:
  required: true
  content:
    multipart/form-data:
      schema:
        type: object
        required:
          - file
        properties:
          file:
            type: string
            format: binary
            description: File to upload
```

---

## Response Format

### Success with Data

```yaml
'200':
  description: Success
  content:
    application/json:
      schema:
        type: object
        properties:
          id:
            type: string
          name:
            type: string
```

### Success without Data

```yaml
'200':
  description: Operation successful
```

### Error Response

```yaml
'400':
  description: Bad request
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/Error'
```

---

## CHECKLIST

### Synchronization Check

- [ ] Read all endpoints from `app/handlers/route.go`
- [ ] Listed all endpoints in OpenAPI
- [ ] **Removed extra endpoints**: Endpoints in OpenAPI that don't exist in code have been removed
- [ ] **Added missing endpoints**: Endpoints in code that are missing from OpenAPI have been added

### Format Check

- [ ] operationId is unique across all yaml files
- [ ] parameters match handler implementation
- [ ] request body schema matches implementation
- [ ] response schema matches implementation
- [ ] Appropriate error responses (400, 401, 403, 404)

### Verification

- [ ] `grep -h 'operationId:' docs/openapi/*.yaml | sort | uniq -d` shows no duplicates
- [ ] YAML syntax validation passes

---

## TROUBLESHOOTING

### Duplicate operationId

```bash
# Find duplicate operationIds
grep -h 'operationId:' docs/openapi/*.yaml | sort | uniq -d

# Find which file has the problem
grep -l 'operationId: duplicatedName' docs/openapi/*.yaml
```

### YAML Syntax Error

```bash
ruby -ryaml -e "YAML.safe_load(File.read('docs/openapi/file.yaml'))"
```

Common issues:
1. Inconsistent indentation (use spaces, not tabs)
2. Incorrect `$ref` path
3. Unquoted strings (especially values containing `:`)
