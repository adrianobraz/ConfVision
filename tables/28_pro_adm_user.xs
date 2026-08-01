table pro_adm_user {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text nome? filters=trim
    date? data?
    int pro_adm_id? {
      table = "Pro_adm"
    }
  
    password senha? {
      sensitive = true
    }
  
    email email? filters=trim|lower
    bool admin?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree|unique", field: [{name: "email", op: "asc"}]}
    {type: "btree", field: [{name: "nome", op: "asc"}]}
  ]
}