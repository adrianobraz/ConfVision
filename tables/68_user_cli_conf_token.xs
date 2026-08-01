table UserCliConfToken {
  auth = true

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    email email? filters=trim|lower
    password senha? {
      sensitive = true
      visibility = "internal"
    }
  
    text usuario? filters=trim
    text password? filters=trim
    int nivel?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "usuario", op: "asc"}]}
  ]
}