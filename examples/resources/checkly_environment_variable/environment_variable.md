---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "checkly_environment_variable Resource - terraform-provider-checkly"
subcategory: ""
description: |-
  
---

# checkly_environment_variable (Resource)



## Example Usage

```terraform
# Simple Enviroment Variable example
resource "checkly_environment_variable" "variable_1" {
  key = "API_KEY"
  value = "loZd9hOGHDUrGvmW"
  locked = true
}

resource "checkly_environment_variable" "variable_2" {
  key = "API_URL"
  value = "http://localhost:3000"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String)
- `value` (String)

### Optional

- `locked` (Boolean)

### Read-Only

- `id` (String) The ID of this resource.

