query "usercliconf_equipamentos/{usercliconf_equipamentos_id}" verb=DELETE {
  api_group = "UserCliConf"

  input {
    int usercliconf_equipamentos_id? filters=min:1
  }

  stack {
    db.del UserCliConf_Equipamentos {
      field_name = "id"
      field_value = $input.usercliconf_equipamentos_id
    }
  }

  response = null
}