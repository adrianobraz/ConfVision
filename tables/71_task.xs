table task {
  auth = false

  schema {
    int id
    bool Running?
    timestamp? update?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
  ]
}