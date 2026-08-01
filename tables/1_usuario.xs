table Usuario {
  auth = true

  schema {
    uuid id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    email email filters=trim|lower
    password senha? {
      sensitive = true
      visibility = "internal"
    }
  
    date? Data?=now
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "id", op: "asc"}]}
    {type: "btree", field: [{name: "email", op: "asc"}]}
    {type: "btree|unique", field: [{name: "email", op: "asc"}]}
  ]
}