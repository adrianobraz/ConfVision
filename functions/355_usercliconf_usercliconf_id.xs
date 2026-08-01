function usercliconf_usercliconf_id {
  input {
    uuid usercliconf_id?
  }

  stack {
    db.get UserCliConf {
      field_name = "id"
      field_value = $input.usercliconf_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = $model
}