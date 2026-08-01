// Delete UserCliConf record
query "usercliconf/{usercliconf_id}" verb=DELETE {
  api_group = "UserCliConf"

  input {
    uuid usercliconf_id?
  }

  stack {
    db.transaction {
      stack {
        db.del UserCliConf {
          field_name = "id"
          field_value = $input.usercliconf_id
        }
      
        db.bulk.delete UserCliConf_Equipamentos {
          where = $db.UserCliConf_Equipamentos.usercliconf_id == $input.usercliconf_id
        } as $UserCliConf_Equipamentos1
      }
    }
  }

  response = ""
}